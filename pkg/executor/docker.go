// Package executor 实现了命令执行器的核心功能。
// 本文件实现了 Docker 容器中的命令执行器。
package executor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

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
	containerOptDevNull  = "/dev/null"
)

// DockerConfig 表示 Docker 执行器的配置
type DockerConfig struct {
	Image     string // Docker 镜像
	WorkDir   string // 工作目录
	BindMount string // 目录绑定
}

// DockerExecutor Docker 命令执行器
type DockerExecutor struct {
	commands    sync.Map              // 注册的命令
	config      DockerConfig          // Docker 配置
	options     *types.ExecuteOptions // 执行选项
	containerID string                // 当前容器ID
	mu          sync.Mutex
}

// NewDockerExecutor 创建新的 Docker 执行器
func NewDockerExecutor(config DockerConfig) *DockerExecutor {
	return &DockerExecutor{
		config:  config,
		options: &types.ExecuteOptions{},
	}
}

// SetOptions 设置执行选项
func (e *DockerExecutor) SetOptions(options *types.ExecuteOptions) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.options = options

	// 更新配置
	if options != nil && options.Metadata != nil {
		if bindMount, ok := options.Metadata[MetadataKeyBindMount]; ok {
			e.config.BindMount = bindMount
		}
		if workDir, ok := options.Metadata[MetadataKeyWorkDir]; ok {
			e.config.WorkDir = workDir
		}
	}
}

// ensureContainer 确保容器存在并运行
func (e *DockerExecutor) ensureContainer() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 如果容器已存在，检查其状态
	if e.containerID != "" {
		cmd := exec.Command("docker", dockerCmdInspect, dockerOptForce, "{{.State.Running}}", e.containerID)
		output, err := cmd.CombinedOutput()
		if err == nil && strings.TrimSpace(string(output)) == "true" {
			return nil
		}
		// 容器不存在或未运行，移除它
		exec.Command("docker", dockerCmdRm, dockerOptForce, e.containerID).Run()
		e.containerID = ""
	}

	// 创建新容器
	containerName := fmt.Sprintf("%s%d", containerNamePrefix, types.GetTimeNow().UnixNano())
	args := []string{
		dockerCmdRun,
		dockerOptDetach,
		dockerOptAutoRemove,
		dockerOptInteractive,
		dockerOptName, containerName,
	}

	// 添加目录绑定
	if e.config.BindMount != "" {
		// 分割源目录和目标目录
		parts := strings.Split(e.config.BindMount, ":")
		if len(parts) >= 2 {
			// 确保源目录存在
			srcDir := parts[0]
			if err := os.MkdirAll(srcDir, 0777); err != nil {
				return fmt.Errorf("failed to create source directory: %v", err)
			}
			// 添加绑定
			args = append(args, dockerOptVolume, e.config.BindMount)
		}
	}

	// 添加镜像
	args = append(args, e.config.Image)

	// 添加命令
	args = append(args, containerCmdTail, containerOptTailF, containerOptDevNull) // 保持容器运行

	// 打印调试信息
	fmt.Printf("Creating container with args: %v\n", args)

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create container: %v, output: %s", err, string(output))
	}

	e.containerID = strings.TrimSpace(string(output))

	// 确保工作目录存在并设置权限
	if e.config.WorkDir != "" {
		// 使用root用户创建目录
		mkdirCmd := exec.Command("docker", dockerCmdExec, e.containerID, containerCmdMkdir, containerOptMkdirP, e.config.WorkDir)
		if err := mkdirCmd.Run(); err != nil {
			return fmt.Errorf("failed to create work directory: %v", err)
		}

		// 设置目录权限
		chmodCmd := exec.Command("docker", dockerCmdExec, e.containerID, containerCmdChmod, containerOptChmod777, e.config.WorkDir)
		if err := chmodCmd.Run(); err != nil {
			return fmt.Errorf("failed to set work directory permissions: %v", err)
		}

		// 列出目录内容以验证
		lsCmd := exec.Command("docker", dockerCmdExec, e.containerID, containerCmdLs, containerOptLsLa, e.config.WorkDir)
		lsOutput, err := lsCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to list work directory: %v", err)
		}
		fmt.Printf("Work directory contents:\n%s\n", string(lsOutput))
	}

	return nil
}

// Execute 执行命令
func (e *DockerExecutor) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	// 更新选项
	if ctx.Options != nil {
		e.SetOptions(ctx.Options)
	}

	if len(ctx.Args) == 0 {
		return nil, fmt.Errorf("no command specified")
	}

	// 检查是否内置命令
	if cmd, ok := e.commands.Load(ctx.Args[0]); ok {
		command := cmd.(*types.Command)
		// 如果没有设置工作目录，使用配置中的工作目录
		if ctx.Options == nil {
			ctx.Options = &types.ExecuteOptions{}
		}
		if ctx.Options.WorkDir == "" && e.config.WorkDir != "" {
			ctx.Options.WorkDir = e.config.WorkDir
		}

		// 确保容器运行
		if err := e.ensureContainer(); err != nil {
			return nil, fmt.Errorf("failed to ensure container: %v", err)
		}

		// 执行内置命令
		return command.Execute(ctx)
	}

	// 确保容器运行
	if err := e.ensureContainer(); err != nil {
		return nil, fmt.Errorf("failed to ensure container: %v", err)
	}

	// 准备 Docker exec 命令
	args := []string{dockerCmdExec}

	// 添加工作目录
	if e.config.WorkDir != "" {
		args = append(args, dockerOptWorkDir, e.config.WorkDir)
	}

	// 添加环境变量
	if ctx.Options != nil {
		for k, v := range ctx.Options.Env {
			args = append(args, dockerOptEnv, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// 添加容器ID和命令
	args = append(args, e.containerID)
	args = append(args, ctx.Args...)

	// 创建命令
	cmd := exec.CommandContext(ctx.Context, "docker", args...)

	// 设置输入输出
	if ctx.Options != nil {
		if ctx.IsPiped {
			if ctx.PipeInput != nil {
				cmd.Stdin = ctx.PipeInput
			}
			if ctx.PipeOutput != nil {
				cmd.Stdout = ctx.PipeOutput
			}
			if ctx.Options.Stderr != nil {
				cmd.Stderr = ctx.Options.Stderr
			}
		} else {
			cmd.Stdin = ctx.Options.Stdin
			cmd.Stdout = ctx.Options.Stdout
			cmd.Stderr = ctx.Options.Stderr
		}
	}

	// 打印调试信息
	fmt.Printf("Executing docker command: %v\n", cmd.Args)
	fmt.Printf("Working directory: %s\n", e.config.WorkDir)
	fmt.Printf("Environment variables: %v\n", ctx.Options.Env)

	// 执行命令
	startTime := types.GetTimeNow()
	err := cmd.Run()
	endTime := types.GetTimeNow()

	result := &types.ExecuteResult{
		CommandName: ctx.Args[0],
		StartTime:   startTime,
		EndTime:     endTime,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			fmt.Printf("Command failed with exit code %d: %s\n", exitErr.ExitCode(), exitErr.Stderr)
		} else {
			result.ExitCode = 1
			fmt.Printf("Command failed with error: %v\n", err)
		}
		result.Error = err
		return result, err
	}

	return result, nil
}

// Close 关闭执行器，清理资源
func (e *DockerExecutor) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.containerID != "" {
		cmd := exec.Command("docker", "rm", "-f", e.containerID)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to remove container: %v", err)
		}
		e.containerID = ""
	}
	return nil
}

// ListCommands 列出所有可用命令
func (e *DockerExecutor) ListCommands() []types.CommandInfo {
	commands := make([]types.CommandInfo, 0)
	e.commands.Range(func(key, value interface{}) bool {
		cmd := value.(*types.Command)
		info := types.CommandInfo{
			Name:        key.(string),
			Description: cmd.Description,
			Usage:       cmd.Usage,
			Category:    cmd.Category,
			Metadata:    cmd.Metadata,
		}
		commands = append(commands, info)
		return true
	})
	return commands
}

// RegisterCommand 注册命令
func (e *DockerExecutor) RegisterCommand(cmd *types.Command) error {
	if cmd == nil {
		return fmt.Errorf("command is nil")
	}
	if cmd.Name == "" {
		return fmt.Errorf("command name is empty")
	}
	if cmd.Handler == nil {
		return fmt.Errorf("command handler is nil")
	}
	e.commands.Store(cmd.Name, cmd)
	return nil
}

// UnregisterCommand 注销命令
func (e *DockerExecutor) UnregisterCommand(cmdName string) error {
	if cmdName == "" {
		return fmt.Errorf("command name is empty")
	}
	e.commands.Delete(cmdName)
	return nil
}
