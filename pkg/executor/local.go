package executor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/iamlongalong/runshell/pkg/log"
	"github.com/iamlongalong/runshell/pkg/types"
)

// LocalExecutor 本地命令执行器
type LocalExecutor struct {
	commands sync.Map // 注册的命令
	config   types.LocalConfig
	options  *types.ExecuteOptions
}

// NewLocalExecutor 创建新的本地执行器
func NewLocalExecutor(config types.LocalConfig, options *types.ExecuteOptions) *LocalExecutor {
	log.Debug("Creating new local executor with options: %+v", options)
	if options == nil {
		options = &types.ExecuteOptions{}
	}
	return &LocalExecutor{
		config:  config,
		options: options,
	}
}

const (
	LocalExecutorName = "local"
)

// Name 返回执行器名称
func (e *LocalExecutor) Name() string {
	return LocalExecutorName
}

// SetOptions 设置执行选项
func (e *LocalExecutor) SetOptions(options *types.ExecuteOptions) {
	log.Debug("Setting local executor options: %+v", options)
	e.options = options
}

// executeCommand 执行单个命令
func (e *LocalExecutor) executeCommand(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Command.Command == "" {
		log.Error("No command specified for execution")
		return nil, fmt.Errorf("no command specified")
	}

	log.Debug("Executing local command: %v", ctx.Command.Args)

	// 检查是否是内置命令
	if cmd, ok := e.commands.Load(ctx.Command.Command); ok {
		log.Debug("Found built-in command: %s", ctx.Command)
		command := cmd.(types.ICommand)
		return command.Execute(ctx)
	}

	if !e.config.AllowUnregisteredCommands {
		log.Error("Unregistered command not allowed: %s", ctx.Command)
		return nil, fmt.Errorf("unregistered command not allowed: %s", ctx.Command)
	}

	// 准备命令
	cmdPath, err := exec.LookPath(ctx.Command.Command)
	if err != nil {
		log.Error("Command not found: %s", ctx.Command.Command)
		return nil, fmt.Errorf("command not found: %s", ctx.Command)
	}
	log.Debug("Found command path: %s", cmdPath)

	cmd := exec.CommandContext(ctx.Context, cmdPath, ctx.Command.Args...)

	// 设置工作目录
	if ctx.Options.WorkDir != "" {
		log.Debug("Setting working directory: %s", ctx.Options.WorkDir)
		cmd.Dir = ctx.Options.WorkDir
	}

	// 设置环境变量
	if len(ctx.Options.Env) > 0 {
		log.Debug("Setting environment variables: %v", ctx.Options.Env)
		cmd.Env = os.Environ()
		for k, v := range ctx.Options.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	var stdoutBuf bytes.Buffer
	// 设置输入输出
	if ctx.Options != nil {
		if ctx.Options.Stdin != nil {
			cmd.Stdin = ctx.Options.Stdin
		}
		if ctx.Options.Stdout != nil {
			cmd.Stdout = io.MultiWriter(&stdoutBuf, ctx.Options.Stdout)
		} else {
			cmd.Stdout = &stdoutBuf
		}
		if ctx.Options.Stderr != nil {
			cmd.Stderr = ctx.Options.Stderr
		}
	}

	// 执行命令
	startTime := types.GetTimeNow()
	err = cmd.Run()
	endTime := types.GetTimeNow()

	// 准备结果
	result := &types.ExecuteResult{
		CommandName: ctx.Command.Command,
		StartTime:   startTime,
		EndTime:     endTime,
		Output:      stdoutBuf.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		result.Error = err
		log.Error("Command execution failed: %v", err)
		return result, err
	}

	result.ExitCode = 0
	return result, nil
}

// Execute 执行命令
func (e *LocalExecutor) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	// 合并默认选项和用户自定义选项
	if ctx.Options == nil {
		ctx.Options = &types.ExecuteOptions{}
	}
	ctx.Options = ctx.Options.Merge(e.options)

	// 如果是管道命令，使用管道执行器
	if ctx.IsPiped {
		return e.executePipeline(ctx)
	}

	return e.executeCommand(ctx)
}

// ListCommands 列出所有可用命令
func (e *LocalExecutor) ListCommands() []types.CommandInfo {
	var commands []types.CommandInfo
	e.commands.Range(func(key, value interface{}) bool {
		if cmd, ok := value.(types.ICommand); ok {
			commands = append(commands, cmd.Info())
		}
		return true
	})
	return commands
}

// Close 关闭执行器
func (e *LocalExecutor) Close() error {
	return nil
}

// RegisterCommand 注册命令
func (e *LocalExecutor) RegisterCommand(cmd types.ICommand) error {
	if cmd == nil {
		return fmt.Errorf("command is nil")
	}

	info := cmd.Info()
	if info.Name == "" {
		return fmt.Errorf("command name is empty")
	}

	e.commands.Store(info.Name, cmd)
	return nil
}

// UnregisterCommand 注销命令
func (e *LocalExecutor) UnregisterCommand(name string) error {
	e.commands.Delete(name)
	return nil
}

// executePipeline 执行管道命令
func (e *LocalExecutor) executePipeline(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.PipeContext == nil || len(ctx.PipeContext.Commands) == 0 {
		return nil, fmt.Errorf("invalid pipeline context")
	}

	var lastResult *types.ExecuteResult
	var lastOutput io.Reader

	startTime := types.GetTimeNow()

	for i, pipeCmd := range ctx.PipeContext.Commands {
		cmdCtx := &types.ExecuteContext{
			Context: ctx.Context,
			Command: types.Command{
				Command: pipeCmd.Command,
				Args:    pipeCmd.Args,
			},
			Options:  ctx.Options,
			Executor: ctx.Executor,
		}

		if i > 0 && lastOutput != nil {
			cmdCtx.Options.Stdin = lastOutput
		}

		var outputBuf bytes.Buffer
		cmdCtx.Options.Stdout = &outputBuf

		result, err := e.executeCommand(cmdCtx)
		if err != nil {
			return result, err
		}

		lastResult = result
		lastOutput = strings.NewReader(outputBuf.String())
	}

	if lastResult != nil {
		lastResult.StartTime = startTime
		lastResult.EndTime = types.GetTimeNow()
	}

	return lastResult, nil
}
