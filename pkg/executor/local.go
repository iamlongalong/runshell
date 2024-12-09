package executor

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/iamlongalong/runshell/pkg/types"
)

// LocalExecutor 本地命令执行器
type LocalExecutor struct {
	commands sync.Map // 注册的命令
	options  *types.ExecuteOptions
}

// NewLocalExecutor 创建新的本地执行器
func NewLocalExecutor() *LocalExecutor {
	return &LocalExecutor{
		options: &types.ExecuteOptions{},
	}
}

// SetOptions 设置执行选项
func (e *LocalExecutor) SetOptions(options *types.ExecuteOptions) {
	e.options = options
}

// Execute 执行命令
func (e *LocalExecutor) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Args) == 0 {
		return nil, fmt.Errorf("no command specified")
	}

	// 检查是否是内置命令
	if cmd, ok := e.commands.Load(ctx.Args[0]); ok {
		command := cmd.(*types.Command)
		return command.Execute(ctx)
	}

	// 准备命令
	cmdPath, err := exec.LookPath(ctx.Args[0])
	if err != nil {
		return nil, fmt.Errorf("command not found: %s", ctx.Args[0])
	}

	cmd := exec.CommandContext(ctx.Context, cmdPath, ctx.Args[1:]...)

	// 设置工作目录
	if ctx.Options.WorkDir != "" {
		cmd.Dir = ctx.Options.WorkDir
	}

	// 设置环境变量
	if len(ctx.Options.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range ctx.Options.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

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

	startTime := types.GetTimeNow()
	err = cmd.Run()
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
func (e *LocalExecutor) ListCommands() []types.CommandInfo {
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
func (e *LocalExecutor) RegisterCommand(cmd *types.Command) error {
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
func (e *LocalExecutor) UnregisterCommand(cmdName string) error {
	if cmdName == "" {
		return fmt.Errorf("command name is empty")
	}
	e.commands.Delete(cmdName)
	return nil
}

// Close 关闭执行器，清理资源
func (e *LocalExecutor) Close() error {
	// 本地执行器不需要特别的清理工作
	return nil
}
