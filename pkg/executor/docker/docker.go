// Package executor 实现了命令执行器的核心功能。
// 本文件实现了 Docker 容器中的命令执行器。
package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/iamlongalong/runshell/pkg/commands"
	"github.com/iamlongalong/runshell/pkg/log"
	runshellTypes "github.com/iamlongalong/runshell/pkg/types"
)

const (
	// DefaultWorkDir 默认工作目录，可以被外部使用
	DefaultWorkDir = "/workspace"

	// MetadataKeyBindMount 绑定挂载的元数据键
	MetadataKeyBindMount = "bind_mount"
	// MetadataKeyWorkDir 工作目录的元数据键
	MetadataKeyWorkDir = "work_dir"
)

// 内部使用的常量，不需要导出
const (
	containerNamePrefix = "runshell-"
)

// DockerExecutor Docker 命令执行器
type DockerExecutor struct {
	commands    sync.Map                      // 注册的命令
	config      runshellTypes.DockerConfig    // Docker 配置
	options     *runshellTypes.ExecuteOptions // 执行选项
	containerID string                        // 当前容器ID
	mu          sync.Mutex
}

// NewDockerExecutor 创建新的 Docker 执行器
func NewDockerExecutor(config runshellTypes.DockerConfig, options *runshellTypes.ExecuteOptions, provider runshellTypes.BuiltinCommandProvider) (*DockerExecutor, error) {
	log.Debug("Creating new Docker executor with config: %+v, options: %+v", config, options)
	if config.Image == "" {
		log.Error("Docker image not specified in config")
		return nil, fmt.Errorf("docker image not specified")
	}
	if options == nil {
		options = &runshellTypes.ExecuteOptions{}
	}
	// TODO: FIXME: if options.WorkDir is not empty, set it to config.WorkDir
	if options.WorkDir != "" {
		config.WorkDir = options.WorkDir
	}

	executor := &DockerExecutor{
		config:  config,
		options: options,
	}

	if provider != nil {
		for _, cmd := range provider.GetCommands() {
			executor.RegisterCommand(cmd)
		}
	}

	if err := executor.ensureContainer(); err != nil {
		return nil, fmt.Errorf("failed to ensure container: %v", err)
	}

	return executor, nil
}

const (
	DockerExecutorName = "docker"
)

// Name 返回执行器名
func (e *DockerExecutor) Name() string {
	return DockerExecutorName
}

// ensureContainer 确保容器存在并运行
func (e *DockerExecutor) ensureContainer() error {
	log.Debug("Ensuring container exists and is running")
	e.mu.Lock()
	defer e.mu.Unlock()

	// 创建 Docker 客户端
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.43"), // 显式指定 API 版本
	)
	if err != nil {
		log.Error("Failed to create Docker client: %v", err)
		return fmt.Errorf("failed to create Docker client: %v", err)
	}
	defer cli.Close()

	ctx := context.Background()

	// 如果容器已存在，检查其状态
	if e.containerID != "" {
		log.Debug("Container exists with ID: %s, checking status", e.containerID)
		inspect, err := cli.ContainerInspect(ctx, e.containerID)
		if err != nil {
			log.Error("Failed to inspect container %s: %v", e.containerID, err)
		} else if inspect.State.Running {
			log.Debug("Container is running")
			return nil
		}
		// if is not running, start it
		if err := cli.ContainerStart(ctx, e.containerID, container.StartOptions{}); err == nil {
			log.Debug("Container is not running, starting it")
			return nil
		}

		// TODO: 请再想想，容器是不是要保持无状态特性
		log.Debug("Container start failed, removing it")
		if err := cli.ContainerRemove(ctx, e.containerID, container.RemoveOptions{Force: true}); err != nil {
			log.Error("Failed to remove container %s: %v", e.containerID, err)
		}
		e.containerID = ""
	}

	// 检查镜像是否存在
	if e.config.Image == "" {
		log.Error("Docker image not specified")
		return fmt.Errorf("docker image not specified")
	}

	// 创建新容器
	containerName := fmt.Sprintf("%s%d", containerNamePrefix, runshellTypes.GetTimeNow().UnixNano())
	log.Debug("Creating new container with name: %s", containerName)

	// 准备容器配置
	config := &container.Config{
		Image:       e.config.Image,
		Tty:         true,
		OpenStdin:   true,
		AttachStdin: true,
		Cmd:         []string{"tail", "-f", "/dev/null"},
	}

	// 准备主机配置
	hostConfig := &container.HostConfig{
		AutoRemove: true,
	}

	// 添加目录绑定
	if e.config.BindMount != "" {
		log.Debug("Adding bind mount: %s", e.config.BindMount)
		parts := strings.Split(e.config.BindMount, ":")
		if len(parts) < 2 {
			log.Error("Invalid bind mount format: %s (expected src:dest)", e.config.BindMount)
			return fmt.Errorf("invalid bind mount format: %s", e.config.BindMount)
		}
		srcDir := parts[0]
		if err := os.MkdirAll(srcDir, 0777); err != nil {
			log.Error("Failed to create source directory %s: %v", srcDir, err)
			return fmt.Errorf("failed to create source directory: %v", err)
		}
		hostConfig.Binds = []string{e.config.BindMount}
	}

	// 设置用户
	if e.config.User != "" {
		config.User = e.config.User
	}

	// 创建容器
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		log.Error("Failed to create container with image %s: %v", e.config.Image, err)
		return fmt.Errorf("failed to create container: %v", err)
	}

	e.containerID = resp.ID
	log.Debug("Container created with ID: %s", e.containerID)

	// 启动容器
	if err := cli.ContainerStart(ctx, e.containerID, container.StartOptions{}); err != nil {
		log.Error("Failed to start container %s: %v", e.containerID, err)
		return fmt.Errorf("failed to start container: %v", err)
	}

	// 确保工作目录存在并设置权限
	if e.config.WorkDir != "" {
		log.Debug("Setting up work directory: %s", e.config.WorkDir)

		// 创建工作目录
		execConfig := container.ExecOptions{
			Cmd:          []string{"mkdir", "-p", e.config.WorkDir},
			AttachStdout: true,
			AttachStderr: true,
		}
		execResp, err := cli.ContainerExecCreate(ctx, e.containerID, execConfig)
		if err != nil {
			log.Error("Failed to create exec for mkdir: %v", err)
			return fmt.Errorf("failed to create work directory: %v", err)
		}
		if err := cli.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{}); err != nil {
			log.Error("Failed to execute mkdir: %v", err)
			return fmt.Errorf("failed to create work directory: %v", err)
		}

		// 设置权限
		execConfig = container.ExecOptions{
			Cmd:          []string{"chmod", "777", e.config.WorkDir},
			AttachStdout: true,
			AttachStderr: true,
		}
		execResp, err = cli.ContainerExecCreate(ctx, e.containerID, execConfig)
		if err != nil {
			log.Error("Failed to create exec for chmod: %v", err)
			return fmt.Errorf("failed to set work directory permissions: %v", err)
		}
		if err := cli.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{}); err != nil {
			log.Error("Failed to execute chmod: %v", err)
			return fmt.Errorf("failed to set work directory permissions: %v", err)
		}

		// 验证目录设置
		execConfig = container.ExecOptions{
			Cmd:          []string{"ls", "-la", e.config.WorkDir},
			AttachStdout: true,
			AttachStderr: true,
		}
		execResp, err = cli.ContainerExecCreate(ctx, e.containerID, execConfig)
		if err != nil {
			log.Error("Failed to create exec for ls: %v", err)
			return fmt.Errorf("failed to list work directory: %v", err)
		}
		if err := cli.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{}); err != nil {
			log.Error("Failed to execute ls: %v", err)
			return fmt.Errorf("failed to list work directory: %v", err)
		}
	}

	return nil
}

// Execute 执行命令
func (e *DockerExecutor) Execute(ctx *runshellTypes.ExecuteContext) (*runshellTypes.ExecuteResult, error) {
	log.Debug("Executing command with context: %+v", ctx)

	// 合并默认选项和用户自定义选项
	if ctx.Options == nil {
		ctx.Options = &runshellTypes.ExecuteOptions{}
	}
	if ctx.Options.WorkDir == "" && e.config.WorkDir != "" {
		ctx.Options.WorkDir = e.config.WorkDir
	}

	ctx.Options = ctx.Options.Merge(e.options)

	if ctx.IsPiped {
		if ctx.PipeContext == nil || len(ctx.PipeContext.Commands) == 0 {
			log.Error("No commands in pipeline")
			return nil, fmt.Errorf("no commands in pipeline")
		}
	} else {
		if ctx.Command.Command == "" {
			log.Error("No command specified for execution")
			return nil, fmt.Errorf("no command specified")
		}
	}

	// 如果是交互式命令
	if ctx.Interactive {
		return e.ExecuteInteractive(ctx)
	}

	// 检查是否内置命令
	if cmd, ok := e.commands.Load(ctx.Command.Command); ok {
		log.Debug("Executing built-in command: %s", ctx.Command)
		command := cmd.(runshellTypes.ICommand)

		ctx.Executor = e
		return command.Execute(ctx)
	}

	// 执行普通命令
	return e.ExecuteCommand(ctx)
}

// ExecuteCommand 执行具体的命令
func (e *DockerExecutor) ExecuteCommand(ctx *runshellTypes.ExecuteContext) (*runshellTypes.ExecuteResult, error) {
	// 确保容器存在并运行
	if err := e.ensureContainer(); err != nil {
		log.Error("Failed to ensure container for command %v: %v", ctx.Command, err)
		return nil, fmt.Errorf("failed to ensure container: %v", err)
	}

	// 创建 Docker 客户端
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion("1.43"))
	if err != nil {
		log.Error("Failed to create Docker client: %v", err)
		return nil, fmt.Errorf("failed to create Docker client: %v", err)
	}
	defer cli.Close()

	// 设置工作目录和环境变量
	if ctx.Options == nil {
		ctx.Options = &runshellTypes.ExecuteOptions{}
	}
	if ctx.Options.WorkDir == "" && e.config.WorkDir != "" {
		ctx.Options.WorkDir = e.config.WorkDir
	}

	// 构建完整的命令字符串
	var execCtx context.Context
	var cmds []string
	if ctx.IsPiped && ctx.PipeContext != nil {
		execCtx = ctx.PipeContext.Context
		var cmdList []string
		for _, cmd := range ctx.PipeContext.Commands {
			cmdList = append(cmdList, fmt.Sprintf("%s %s", cmd.Command, strings.Join(cmd.Args, " ")))
		}
		// 使用 shell 来执行管道命令
		cmds = []string{"/bin/sh", "-c", strings.Join(cmdList, " | ")}
	} else {
		execCtx = ctx.Context
		cmds = []string{"/bin/sh", "-c", ctx.Command.Command + " " + strings.Join(ctx.Command.Args, " ")}
	}

	// 准备环境变量
	var env []string
	envMap := make(map[string]string)

	// 首先添加执行器级别的环境变量
	for k, v := range e.options.Env {
		envMap[k] = v
	}

	// 然后添加上下文级别的环境变量（可能会覆盖执行器级别的）
	if ctx.Options != nil {
		for k, v := range ctx.Options.Env {
			envMap[k] = v
		}
	}

	// 将合并后的环境变量转换为切片格式
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// 创建执行配置
	execConfig := container.ExecOptions{
		User:         e.config.User,
		WorkingDir:   ctx.Options.WorkDir,
		Cmd:          cmds,
		AttachStdin:  ctx.Options.Stdin != nil,
		AttachStdout: true,
		AttachStderr: true,
		Env:          env,
		Tty:          false,
	}

	// 创建执行实例
	log.Debug("Creating exec instance for command: %v", cmds)
	execResp, err := cli.ContainerExecCreate(execCtx, e.containerID, execConfig)
	if err != nil {
		log.Error("Failed to create exec instance: %v", err)
		return nil, fmt.Errorf("failed to create exec instance: %v", err)
	}

	// 附加到执���实例
	log.Debug("Attaching to exec instance: %s", execResp.ID)
	resp, err := cli.ContainerExecAttach(execCtx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		log.Error("Failed to attach to exec instance: %v", err)
		return nil, fmt.Errorf("failed to attach to exec instance: %v", err)
	}
	defer resp.Close()

	// 设置开始时间
	startTime := runshellTypes.GetTimeNow()

	// 创建输出缓冲区
	var outputBuf bytes.Buffer
	var writers []io.Writer
	writers = append(writers, &outputBuf)
	if ctx.Options.Stdout != nil {
		writers = append(writers, ctx.Options.Stdout)
	}
	outputWriter := io.MultiWriter(writers...)

	// 创建完成通道
	done := make(chan struct{})
	var copyErr error

	// 单个 goroutine 处理所有 I/O
	go func() {
		defer close(done)

		// 如果有输入，先处理输入
		if ctx.Options.Stdin != nil {
			go func() {
				io.Copy(resp.Conn, ctx.Options.Stdin)
				resp.CloseWrite()
			}()
		}

		// 处理输出
		_, err := io.Copy(outputWriter, resp.Reader)
		if err != nil && err != io.EOF {
			copyErr = err
		}
	}()

	// 等待命令完成
	var result *runshellTypes.ExecuteResult
	for {
		select {
		case <-done:
			if copyErr != nil {
				return nil, fmt.Errorf("error copying data: %v", copyErr)
			}
		default:
		}

		inspectResp, err := cli.ContainerExecInspect(execCtx, execResp.ID)
		if err != nil {
			log.Error("Failed to inspect exec instance: %v", err)
			break
		}

		if !inspectResp.Running {
			endTime := runshellTypes.GetTimeNow()

			result = &runshellTypes.ExecuteResult{
				CommandName: ctx.Command.Command,
				StartTime:   startTime,
				EndTime:     endTime,
				Output:      outputBuf.String(),
				ExitCode:    inspectResp.ExitCode,
			}

			if inspectResp.ExitCode != 0 {
				err = fmt.Errorf("command exited with code %d", inspectResp.ExitCode)
				result.Error = err
				log.Error("Command %v failed with exit code %d: %s", cmds, inspectResp.ExitCode, result.Output)
				return result, err
			}

			log.Info("Command completed successfully: %s", cmds)
			return result, nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil, fmt.Errorf("failed to execute command")
}

// Close 关闭执行器，清理资源
func (e *DockerExecutor) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.containerID != "" {
		log.Debug("Removing container: %s", e.containerID)
		cmd := exec.Command("docker", "rm", "-f", e.containerID)
		if err := cmd.Run(); err != nil {
			log.Error("Failed to remove container %s: %v", e.containerID, err)
			return fmt.Errorf("failed to remove container: %v", err)
		}
		log.Info("Container removed successfully: %s", e.containerID)
		e.containerID = ""
	}
	return nil
}

// ListCommands 列出所有可用命令
func (e *DockerExecutor) ListCommands() []runshellTypes.CommandInfo {
	log.Debug("Listing available commands")
	commands := make([]runshellTypes.CommandInfo, 0)
	e.commands.Range(func(key, value interface{}) bool {
		cmd := value.(runshellTypes.ICommand)
		info := cmd.Info()
		commands = append(commands, info)
		return true
	})
	log.Debug("Found %d commands", len(commands))
	return commands
}

// RegisterCommand 注册命令
func (e *DockerExecutor) RegisterCommand(cmd runshellTypes.ICommand) error {
	if cmd == nil {
		log.Error("Attempted to register nil command")
		return fmt.Errorf("command is nil")
	}
	if cmd.Info().Name == "" {
		log.Error("Attempted to register command with empty name")
		return fmt.Errorf("command name is empty")
	}

	log.Info("Registering command: %s", cmd.Info().Name)
	e.commands.Store(cmd.Info().Name, cmd)
	return nil
}

// UnregisterCommand 注销命令
func (e *DockerExecutor) UnregisterCommand(cmdName string) error {
	if cmdName == "" {
		log.Error("Attempted to unregister command with empty name")
		return fmt.Errorf("command name is empty")
	}
	log.Info("Unregistering command: %s", cmdName)
	e.commands.Delete(cmdName)
	return nil
}

// DockerExecutorBuilder 是 Docker 执行器的构建器。
type DockerExecutorBuilder struct {
	config  runshellTypes.DockerConfig
	options *runshellTypes.ExecuteOptions
}

// NewDockerExecutorBuilder 创建一个新的 Docker 执行器构建器。
func NewDockerExecutorBuilder(config runshellTypes.DockerConfig) *DockerExecutorBuilder {
	return &DockerExecutorBuilder{
		config: config,
	}
}

// Build 构建并返回一个新的 Docker 执行器实例。
func (b *DockerExecutorBuilder) Build(options *runshellTypes.ExecuteOptions) (runshellTypes.Executor, error) {
	if options == nil {
		options = b.options
	}
	if b.config.UseBuiltinCommands {
		provider := commands.NewDefaultCommandProvider()
		return NewDockerExecutor(b.config, options, provider)
	}
	return NewDockerExecutor(b.config, options, nil)
}

// WithOptions 设置执行选项。
func (b *DockerExecutorBuilder) WithOptions(options *runshellTypes.ExecuteOptions) *DockerExecutorBuilder {
	b.options = options
	return b
}

// ExecuteInteractive 在 Docker 容器中执行交互式命令
func (e *DockerExecutor) ExecuteInteractive(ctx *runshellTypes.ExecuteContext) (*runshellTypes.ExecuteResult, error) {
	if err := e.ensureContainer(); err != nil {
		return nil, fmt.Errorf("failed to ensure container: %v", err)
	}

	if ctx.Options == nil {
		ctx.Options = &runshellTypes.ExecuteOptions{}
	}

	// 构建 docker exec 命令
	args := []string{
		"exec",
		"-it", // 交互式和伪终端
	}

	// 设置工作目录
	if e.config.WorkDir != "" {
		args = append(args, "-w", e.config.WorkDir)
	}

	// 设置用户
	if e.config.User != "" {
		args = append(args, "--user", e.config.User)
	}

	// 添加环境变量
	if ctx.Options.Env != nil {
		for k, v := range ctx.Options.Env {
			args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
		}
	}

	// 添加容器ID
	args = append(args, e.containerID)

	// 添加要执行的命令
	args = append(args, ctx.Command.Command)
	args = append(args, ctx.Command.Args...)

	// 创建命令
	cmd := exec.CommandContext(ctx.Context, "docker", args...)

	// 创建伪终端
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start pty: %w", err)
	}
	defer ptmx.Close()

	// 设置端大小
	if ctx.InteractiveOpts != nil && ctx.InteractiveOpts.Rows > 0 && ctx.InteractiveOpts.Cols > 0 {
		if err := pty.Setsize(ptmx, &pty.Winsize{
			Rows: ctx.InteractiveOpts.Rows,
			Cols: ctx.InteractiveOpts.Cols,
		}); err != nil {
			log.Error("Failed to resize pty: %v", err)
		}
	}

	// 创建等待组和错误通道
	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	// 处理输入
	if ctx.Options.Stdin != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := io.Copy(ptmx, ctx.Options.Stdin)
			if err != nil {
				log.Error("Failed to copy stdin: %v", err)
				errCh <- err
			}
		}()
	}

	// 处理输出
	if ctx.Options.Stdout != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := io.Copy(ctx.Options.Stdout, ptmx)
			if err != nil {
				log.Error("Failed to copy stdout: %v", err)
				errCh <- err
			}
		}()
	}

	// 创建完成通道
	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	// 等待命令完成或出错
	startTime := runshellTypes.GetTimeNow()
	cmdDone := make(chan error, 1)
	go func() {
		cmdDone <- cmd.Wait()
	}()

	var cmdErr error
	select {
	case err := <-cmdDone:
		cmdErr = err
	case err := <-errCh:
		cmdErr = err
		cmd.Process.Kill()
	case <-ctx.Context.Done():
		cmdErr = ctx.Context.Err()
		cmd.Process.Kill()
	}

	endTime := runshellTypes.GetTimeNow()

	// 等待 IO 完成
	<-doneCh

	result := &runshellTypes.ExecuteResult{
		CommandName: ctx.Command.Command,
		StartTime:   startTime,
		EndTime:     endTime,
	}

	if cmdErr != nil {
		if exitErr, ok := cmdErr.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		result.Error = cmdErr
		return result, cmdErr
	}

	result.ExitCode = 0
	return result, nil
}
