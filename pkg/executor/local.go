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

// LocalExecutorBuilder 构建本地执行器的构建器
type LocalExecutorBuilder struct {
	config  types.LocalConfig
	options *types.ExecuteOptions
}

// NewLocalExecutorBuilder 创建新的本地执行器构建器
func NewLocalExecutorBuilder(config types.LocalConfig, options *types.ExecuteOptions) *LocalExecutorBuilder {
	if options == nil {
		options = &types.ExecuteOptions{}
	}
	return &LocalExecutorBuilder{
		config:  config,
		options: options,
	}
}

// Build 实现 ExecutorBuilder 接口
func (b *LocalExecutorBuilder) Build() (types.Executor, error) {
	return NewLocalExecutor(b.config, b.options), nil
}
