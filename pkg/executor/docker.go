// Package executor 实现了命令执行器的核心功能。
// 本文件实现了 Docker 容器中的命令执行器。
package executor

import (
	"fmt"
	"os/exec"
	"sync"

	"github.com/iamlongalong/runshell/pkg/types"
)

// DockerExecutor Docker 命令执行器
type DockerExecutor struct {
	commands     sync.Map // 注册的命令
	defaultImage string   // 默认 Docker 镜像
}

// NewDockerExecutor 创建新的 Docker 执行器
func NewDockerExecutor(image string) *DockerExecutor {
	return &DockerExecutor{
		defaultImage: image,
	}
}

// Execute 执行命令
func (e *DockerExecutor) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Args) == 0 {
		return nil, fmt.Errorf("no command specified")
	}

	// 检查是否是内置命令
	if cmd, ok := e.commands.Load(ctx.Args[0]); ok {
		command := cmd.(*types.Command)
		return command.Execute(ctx)
	}

	// 准备 Docker 命令
	args := []string{"run", "--rm"}

	// 添加工作目录
	if ctx.Options.WorkDir != "" {
		args = append(args, "-w", ctx.Options.WorkDir)
	}

	// 添加环境变量
	for k, v := range ctx.Options.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// 添加镜像名称
	args = append(args, e.defaultImage)

	// 添加要执行的命令和参数
	args = append(args, ctx.Args...)

	// 创建命令
	cmd := exec.CommandContext(ctx.Context, "docker", args...)

	// 设置输入输出
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
		} else {
			result.ExitCode = 1
		}
		result.Error = err
		return result, err
	}

	return result, nil
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
