package executor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"al.essio.dev/pkg/shellescape"
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
func NewLocalExecutor(config types.LocalConfig, options *types.ExecuteOptions, provider types.BuiltinCommandProvider) *LocalExecutor {
	log.Debug("Creating new local executor with options: %+v", options)
	if options == nil {
		options = &types.ExecuteOptions{}
	}
	executor := &LocalExecutor{
		config:  config,
		options: options,
	}

	// 如果提供了内置命令提供者，注册所有内置命令
	if provider != nil {
		for _, cmd := range provider.GetCommands() {
			executor.RegisterCommand(cmd)
		}
	}

	return executor
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

// executeCommand 执行单命令
func (e *LocalExecutor) executeCommand(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx == nil || ctx.Command.Command == "" {
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

	// 创建命令
	cmd := exec.CommandContext(ctx.Context, cmdPath, ctx.Command.Args...)

	// 设置工作目录
	if ctx.Options != nil && ctx.Options.WorkDir != "" {
		log.Debug("Setting working directory: %s", ctx.Options.WorkDir)
		cmd.Dir = ctx.Options.WorkDir
	}

	// 设置环境变量
	if ctx.Options != nil && len(ctx.Options.Env) > 0 {
		log.Debug("Setting environment variables: %v", ctx.Options.Env)
		cmd.Env = os.Environ()
		for k, v := range ctx.Options.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// 设置输入输出
	var stdoutBuf, stderrBuf bytes.Buffer
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
	err = cmd.Run()
	endTime := types.GetTimeNow()

	// 准备结果
	result := &types.ExecuteResult{
		CommandName: ctx.Command.Command,
		StartTime:   startTime,
		EndTime:     endTime,
	}

	// 合并输出
	result.Output = stdoutBuf.String()
	if stderrBuf.Len() > 0 {
		if result.Output != "" {
			result.Output += "\n"
		}
		result.Output += stderrBuf.String()
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
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

	// 检查命令是否为空
	if ctx.Command.Command == "" {
		return nil, fmt.Errorf("no command specified")
	}

	log.Debug("Executing command: %s %v", ctx.Command.Command, ctx.Command.Args)

	// 检查是否是内置命令
	if cmd, ok := e.commands.Load(ctx.Command.Command); ok {
		log.Debug("Found built-in command: %s", ctx.Command.Command)
		command := cmd.(types.ICommand)
		return command.Execute(ctx)
	}

	if !e.config.AllowUnregisteredCommands {
		log.Error("Unregistered command not allowed: %s", ctx.Command.Command)
		return nil, fmt.Errorf("unregistered command not allowed: %s", ctx.Command.Command)
	}

	// 准备命令
	cmdPath, err := exec.LookPath(ctx.Command.Command)
	if err != nil {
		log.Error("Command not found: %s", ctx.Command.Command)
		return nil, fmt.Errorf("command not found: %s", ctx.Command.Command)
	}
	log.Debug("Found command path: %s", cmdPath)

	// 创建命令
	cmd := exec.CommandContext(ctx.Context, cmdPath, ctx.Command.Args...)

	// 合并选项，避免递归
	finalOptions := &types.ExecuteOptions{
		Env:      make(map[string]string),
		Metadata: make(map[string]string),
	}

	// 首先应用执行器的默认选项
	if e.options != nil {
		if e.options.WorkDir != "" {
			finalOptions.WorkDir = e.options.WorkDir
		}
		if e.options.Env != nil {
			for k, v := range e.options.Env {
				finalOptions.Env[k] = v
			}
		}
		finalOptions.Timeout = e.options.Timeout
		finalOptions.Stdin = e.options.Stdin
		finalOptions.Stdout = e.options.Stdout
		finalOptions.Stderr = e.options.Stderr
		finalOptions.User = e.options.User
	}

	// 然后应用上下文中的选项，覆盖默认选项
	if ctx.Options != nil {
		if ctx.Options.WorkDir != "" {
			finalOptions.WorkDir = ctx.Options.WorkDir
		}
		if ctx.Options.Env != nil {
			for k, v := range ctx.Options.Env {
				finalOptions.Env[k] = v
			}
		}
		if ctx.Options.Timeout != 0 {
			finalOptions.Timeout = ctx.Options.Timeout
		}
		if ctx.Options.Stdin != nil {
			finalOptions.Stdin = ctx.Options.Stdin
		}
		if ctx.Options.Stdout != nil {
			finalOptions.Stdout = ctx.Options.Stdout
		}
		if ctx.Options.Stderr != nil {
			finalOptions.Stderr = ctx.Options.Stderr
		}
		if ctx.Options.User != nil {
			finalOptions.User = ctx.Options.User
		}
	}

	// 设置工作目录
	if finalOptions.WorkDir != "" {
		log.Debug("Setting working directory: %s", finalOptions.WorkDir)
		cmd.Dir = finalOptions.WorkDir
	}

	// 设置环境变量
	if len(finalOptions.Env) > 0 {
		log.Debug("Setting environment variables: %v", finalOptions.Env)
		cmd.Env = os.Environ()
		for k, v := range finalOptions.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// 设置输入输出
	var stdoutBuf, stderrBuf bytes.Buffer
	if finalOptions.Stdin != nil {
		cmd.Stdin = finalOptions.Stdin
	}
	if finalOptions.Stdout != nil {
		cmd.Stdout = io.MultiWriter(&stdoutBuf, finalOptions.Stdout)
	} else {
		cmd.Stdout = &stdoutBuf
	}
	if finalOptions.Stderr != nil {
		cmd.Stderr = io.MultiWriter(&stderrBuf, finalOptions.Stderr)
	} else {
		cmd.Stderr = &stderrBuf
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
	}

	// 合并输出
	result.Output = stdoutBuf.String()
	if stderrBuf.Len() > 0 {
		if result.Output != "" {
			result.Output += "\n"
		}
		result.Output += stderrBuf.String()
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		result.Error = err
		log.Error("Command execution failed: %v", err)
		return result, err
	}

	result.ExitCode = 0
	return result, nil
}

// ListCommands 列出所有可用命令
func (e *LocalExecutor) ListCommands() []types.CommandInfo {
	var commands []types.CommandInfo

	// 首先添加已注册的命令
	e.commands.Range(func(key, value interface{}) bool {
		if cmd, ok := value.(types.ICommand); ok {
			commands = append(commands, cmd.Info())
		}
		return true
	})

	// 如果没有注册的命令且允许未注册的命令，返回一些常用命令
	if len(commands) == 0 && e.config.AllowUnregisteredCommands {
		return []types.CommandInfo{
			{
				Name:        "ls",
				Description: "List directory contents",
				Usage:       "ls [OPTION]... [FILE]...",
				Category:    "file",
			},
			{
				Name:        "pwd",
				Description: "Print working directory",
				Usage:       "pwd",
				Category:    "file",
			},
			{
				Name:        "cd",
				Description: "Change directory",
				Usage:       "cd [dir]",
				Category:    "file",
			},
			{
				Name:        "echo",
				Description: "Display a line of text",
				Usage:       "echo [STRING]...",
				Category:    "shell",
			},
		}
	}

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
		return nil, fmt.Errorf("no commands in pipeline")
	}

	// 检查命令是否为空
	for _, cmd := range ctx.PipeContext.Commands {
		if cmd == nil || cmd.Command == "" {
			return nil, fmt.Errorf("command not found")
		}
	}

	// 检查是否允许执行未注册的命令
	if !e.config.AllowUnregisteredCommands {
		for _, cmd := range ctx.PipeContext.Commands {
			if _, ok := e.commands.Load(cmd.Command); !ok {
				return nil, fmt.Errorf("unregistered command not allowed: %s", cmd.Command)
			}
		}
	}

	startTime := types.GetTimeNow()

	// 使用 bash -c 来实现
	cmds := []string{}
	for _, cmd := range ctx.PipeContext.Commands {
		if len(cmd.Args) > 0 {
			// 只对包含空格的参数进行简单的引号包裹
			quotedArgs := make([]string, len(cmd.Args))
			for i, arg := range cmd.Args {
				quotedArgs[i] = shellescape.Quote(arg)
			}
			cmds = append(cmds, fmt.Sprintf("%s %s", cmd.Command, strings.Join(quotedArgs, " ")))
		} else {
			cmds = append(cmds, cmd.Command)
		}
	}
	cmdStr := strings.Join(cmds, " | ")

	// Create command with explicit shell
	cmd := exec.CommandContext(ctx.Context, "bash", "-c", cmdStr)
	var stdoutBuf, stderrBuf bytes.Buffer

	// Set up output redirection
	if ctx.PipeContext != nil && ctx.PipeContext.Options != nil {
		if ctx.PipeContext.Options.Stdin != nil {
			cmd.Stdin = ctx.PipeContext.Options.Stdin
		}
		if ctx.PipeContext.Options.Stdout != nil {
			cmd.Stdout = ctx.PipeContext.Options.Stdout
		} else {
			cmd.Stdout = &stdoutBuf
		}
		if ctx.PipeContext.Options.Stderr != nil {
			cmd.Stderr = ctx.PipeContext.Options.Stderr
		} else {
			cmd.Stderr = &stderrBuf
		}
	} else {
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
	}

	// Execute command
	err := cmd.Run()
	endTime := types.GetTimeNow()

	// Check context cancellation
	if ctx.Context != nil && ctx.Context.Err() != nil {
		return nil, ctx.Context.Err()
	}

	// Prepare result
	result := &types.ExecuteResult{
		CommandName: cmdStr,
		StartTime:   startTime,
		EndTime:     endTime,
	}

	// Get output from buffer if not using direct output
	if ctx.PipeContext == nil || ctx.PipeContext.Options == nil || ctx.PipeContext.Options.Stdout == nil {
		result.Output = stdoutBuf.String()
		if stderrBuf.Len() > 0 {
			if result.Output != "" {
				result.Output += "\n"
			}
			result.Output += stderrBuf.String()
		}
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

	result.ExitCode = 0
	return result, nil
}
