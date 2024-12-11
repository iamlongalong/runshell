// Package executor 实现了命令执行器的核心功能。
// 本文件实现了 Docker 容器中的命令执行器。
package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/iamlongalong/runshell/pkg/log"
	"github.com/iamlongalong/runshell/pkg/types"
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

	dockerCmdRun     = "run"
	dockerCmdExec    = "exec"
	dockerCmdRm      = "rm"
	dockerCmdInspect = "inspect"

	dockerOptDetach      = "-d"
	dockerOptAutoRemove  = "--rm"
	dockerOptInteractive = "-i"
	dockerOptName        = "--name"
	dockerOptVolume      = "-v"
	dockerOptWorkDir     = "-w"
	dockerOptEnv         = "-e"
	dockerOptForce       = "-f"

	containerCmdMkdir = "mkdir"
	containerCmdChmod = "chmod"
	containerCmdLs    = "ls"
	containerCmdTail  = "tail"

	containerOptMkdirP   = "-p"
	containerOptChmod777 = "777"
	containerOptLsLa     = "-la"
	containerOptTailF    = "-f"
	containerOptUser     = "--user"
	containerOptDevNull  = "/dev/null"
)

// DockerExecutor Docker 命令执行器
type DockerExecutor struct {
	commands    sync.Map              // 注册的命令
	config      types.DockerConfig    // Docker 配置
	options     *types.ExecuteOptions // 执行选项
	containerID string                // 当前容器ID
	mu          sync.Mutex
}

// NewDockerExecutor 创建新的 Docker 执行器
func NewDockerExecutor(config types.DockerConfig, options *types.ExecuteOptions) (*DockerExecutor, error) {
	log.Debug("Creating new Docker executor with config: %+v, options: %+v", config, options)
	if config.Image == "" {
		log.Error("Docker image not specified in config")
	}
	if options == nil {
		options = &types.ExecuteOptions{}
	}
	executor := &DockerExecutor{
		config:  config,
		options: options,
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

	// 检查 Docker 是否可用
	if _, err := exec.LookPath("docker"); err != nil {
		log.Error("Docker command not found in PATH: %v", err)
		return fmt.Errorf("docker command not found: %v", err)
	}

	// 如果容器已存在，检查其状态
	if e.containerID != "" {
		log.Debug("Container exists with ID: %s, checking status", e.containerID)
		cmd := exec.Command("docker", dockerCmdInspect, dockerOptForce, "{{.State.Running}}", e.containerID)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error("Failed to inspect container %s: %v", e.containerID, err)
		} else if strings.TrimSpace(string(output)) == "true" {
			log.Debug("Container is running")
			return nil
		}
		log.Debug("Container is not running or doesn't exist, removing it")
		if rmErr := exec.Command("docker", dockerCmdRm, dockerOptForce, e.containerID).Run(); rmErr != nil {
			log.Error("Failed to remove container %s: %v", e.containerID, rmErr)
		}
		e.containerID = ""
	}

	// 检查镜像是否存在
	if e.config.Image == "" {
		log.Error("Docker image not specified")
		return fmt.Errorf("docker image not specified")
	}

	// 创建新容器
	containerName := fmt.Sprintf("%s%d", containerNamePrefix, types.GetTimeNow().UnixNano())
	log.Debug("Creating new container with name: %s", containerName)
	args := []string{
		dockerCmdRun,
		dockerOptDetach,
		dockerOptAutoRemove,
		dockerOptInteractive,
		dockerOptName, containerName,
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
		args = append(args, dockerOptVolume, e.config.BindMount)
	}

	if e.config.User != "" {
		args = append(args, containerOptUser, e.config.User)
	}

	args = append(args, e.config.Image)
	args = append(args, containerCmdTail, containerOptTailF, containerOptDevNull)

	log.Debug("Running docker command with args: %v", args)
	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("Failed to create container with image %s: %v\nOutput: %s", e.config.Image, err, string(output))
		return fmt.Errorf("failed to create container: %v, output: %s", err, string(output))
	}

	e.containerID = strings.TrimSpace(string(output))
	if e.containerID == "" {
		log.Error("Container creation succeeded but got empty container ID")
		return fmt.Errorf("container creation succeeded but got empty container ID")
	}
	log.Debug("Container created with ID: %s", e.containerID)

	// 确保工作目录存在并设置权限
	if e.config.WorkDir != "" {
		log.Debug("Setting up work directory: %s", e.config.WorkDir)
		mkdirCmd := exec.Command("docker", dockerCmdExec, e.containerID, containerCmdMkdir, containerOptMkdirP, e.config.WorkDir)
		if err := mkdirCmd.Run(); err != nil {
			log.Error("Failed to create work directory %s in container %s: %v", e.config.WorkDir, e.containerID, err)
			return fmt.Errorf("failed to create work directory: %v", err)
		}

		log.Debug("Setting work directory permissions")
		chmodCmd := exec.Command("docker", dockerCmdExec, e.containerID, containerCmdChmod, containerOptChmod777, e.config.WorkDir)
		if err := chmodCmd.Run(); err != nil {
			log.Error("Failed to set permissions for directory %s in container %s: %v", e.config.WorkDir, e.containerID, err)
			return fmt.Errorf("failed to set work directory permissions: %v", err)
		}

		log.Debug("Verifying work directory setup")
		lsCmd := exec.Command("docker", dockerCmdExec, e.containerID, containerCmdLs, containerOptLsLa, e.config.WorkDir)
		lsOutput, err := lsCmd.CombinedOutput()
		if err != nil {
			log.Error("Failed to list work directory %s in container %s: %v", e.config.WorkDir, e.containerID, err)
			return fmt.Errorf("failed to list work directory: %v", err)
		}
		log.Debug("Work directory contents:\n%s", string(lsOutput))
	}

	return nil
}

// Execute 执行命令
func (e *DockerExecutor) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	log.Debug("Executing command with context: %+v", ctx)

	// 合并默认选项和用户自定义选项
	if ctx.Options == nil {
		ctx.Options = &types.ExecuteOptions{}
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

	// 检查是否内置命令
	if cmd, ok := e.commands.Load(ctx.Command.Command); ok {
		log.Debug("Executing built-in command: %s", ctx.Command)
		command := cmd.(types.ICommand)
		if ctx.Options == nil {
			ctx.Options = &types.ExecuteOptions{}
		}
		if ctx.Options.WorkDir == "" && e.config.WorkDir != "" {
			log.Debug("Using default work directory: %s", e.config.WorkDir)
			ctx.Options.WorkDir = e.config.WorkDir
		}

		if err := e.ensureContainer(); err != nil {
			log.Error("Failed to ensure container for built-in command %s: %v", ctx.Command, err)
			return nil, fmt.Errorf("failed to ensure container: %v", err)
		}

		return command.Execute(ctx)
	}

	// 检查是否允许执行未注册的命令
	if !e.config.AllowUnregisteredCommands {
		// 检查是否是管道命令
		if ctx.IsPiped && ctx.PipeContext != nil {
			// 检查是否所有命令都支持
			for _, cmd := range ctx.PipeContext.Commands {
				if _, ok := e.commands.Load(cmd.Command); !ok {
					log.Error("Command not supported: %s", cmd.Command)
					return nil, fmt.Errorf("command not supported: %s", cmd.Command)
				}
			}
		} else {
			log.Error("Unregistered command not allowed: %s", ctx.Command)
			return nil, fmt.Errorf("unregistered command not allowed: %s", ctx.Command)
		}
	}

	// 确保容器存在并运行
	if err := e.ensureContainer(); err != nil {
		log.Error("Failed to ensure container for command %v: %v", ctx.Command, err)
		return nil, fmt.Errorf("failed to ensure container: %v", err)
	}

	log.Debug("Executing regular command: %v", ctx.Command)
	args := []string{dockerCmdExec}

	if e.config.WorkDir != "" {
		log.Debug("Setting work directory for command: %s", e.config.WorkDir)
		args = append(args, dockerOptWorkDir, e.config.WorkDir)
	}

	if ctx.Options != nil {
		for k, v := range ctx.Options.Env {
			log.Debug("Adding environment variable: %s=%s", k, v)
			args = append(args, dockerOptEnv, fmt.Sprintf("%s=%s", k, v))
		}
	}

	args = append(args, e.containerID)

	// 构建完整的命令字符串
	var cctx context.Context
	var fullCmd string
	if ctx.IsPiped && ctx.PipeContext != nil {
		cctx = ctx.PipeContext.Context
		var cmdList []string
		for _, cmd := range ctx.PipeContext.Commands {
			cmdList = append(cmdList, fmt.Sprintf("%s %s", cmd.Command, strings.Join(cmd.Args, " ")))
		}
		fullCmd = strings.Join(cmdList, " | ")
	} else {
		cctx = ctx.Context
		fullCmd = ctx.Command.Command + " " + strings.Join(ctx.Command.Args, " ")
	}

	log.Debug("Full shell command: %s", fullCmd)
	args = append(args, "sh", "-c", fullCmd)

	log.Debug("Final docker command args: %v", args)
	cmd := exec.CommandContext(cctx, "docker", args...)

	// 设置输入输出
	var stdoutBuf, stderrBuf bytes.Buffer
	if ctx.Options != nil {
		log.Debug("Setting up standard IO for command")
		if ctx.Options.Stdin != nil {
			cmd.Stdin = ctx.Options.Stdin
		}
		if ctx.Options.Stdout != nil {
			cmd.Stdout = io.MultiWriter(&stdoutBuf, ctx.Options.Stdout)
		} else {
			cmd.Stdout = &stdoutBuf
		}
		if ctx.Options.Stderr != nil {
			cmd.Stderr = io.MultiWriter(&stderrBuf, ctx.Options.Stderr)
		} else {
			cmd.Stderr = &stderrBuf
		}
	} else {
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
	}

	// 执行命令
	startTime := types.GetTimeNow()
	log.Info("Starting command execution: %s", fullCmd)
	err := cmd.Run()
	endTime := types.GetTimeNow()

	// 合并输出
	output := stdoutBuf.String()
	if stderrBuf.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += stderrBuf.String()
	}

	result := &types.ExecuteResult{
		CommandName: ctx.Command.Command,
		StartTime:   startTime,
		EndTime:     endTime,
		Output:      output,
		Error:       err,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			log.Error("Command %s failed with exit code %d: %s", fullCmd, exitErr.ExitCode(), output)
		} else {
			result.ExitCode = 1
			log.Error("Command %s failed with error: %v", fullCmd, err)
		}
		return result, err
	}

	log.Info("Command completed successfully: %s", fullCmd)
	return result, nil
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
func (e *DockerExecutor) ListCommands() []types.CommandInfo {
	log.Debug("Listing available commands")
	commands := make([]types.CommandInfo, 0)
	e.commands.Range(func(key, value interface{}) bool {
		cmd := value.(types.ICommand)
		info := cmd.Info()
		commands = append(commands, info)
		return true
	})
	log.Debug("Found %d commands", len(commands))
	return commands
}

// RegisterCommand 注册命令
func (e *DockerExecutor) RegisterCommand(cmd types.ICommand) error {
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

// DockerExecutorBuilder 构建Docker执行器的构建器
type DockerExecutorBuilder struct {
	config  types.DockerConfig
	options *types.ExecuteOptions
}

// NewDockerExecutorBuilder 创建新的Docker执行器构建器
func NewDockerExecutorBuilder(config types.DockerConfig, options *types.ExecuteOptions) *DockerExecutorBuilder {
	if options == nil {
		options = &types.ExecuteOptions{}
	}
	return &DockerExecutorBuilder{
		config:  config,
		options: options,
	}
}

// Build 实现 ExecutorBuilder 接口
func (b *DockerExecutorBuilder) Build() (types.Executor, error) {
	return NewDockerExecutor(b.config, b.options)
}
