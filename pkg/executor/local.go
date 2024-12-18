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
	"github.com/creack/pty"
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

// Execute 执行命令
func (e *LocalExecutor) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

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

	if ctx.Options == nil {
		ctx.Options = &types.ExecuteOptions{}
	}
	if ctx.Options.WorkDir == "" && e.config.WorkDir != "" {
		ctx.Options.WorkDir = e.config.WorkDir
	}

	// 如果是交互式命令
	if ctx.Interactive {
		return e.ExecuteInteractive(ctx)
	}

	log.Debug("Executing command: %s %v", ctx.Command.Command, ctx.Command.Args)
	// 检查是否是内置命令
	if cmd, ok := e.commands.Load(ctx.Command.Command); ok {
		ctx.Executor = e
		return cmd.(types.ICommand).Execute(ctx)
	}

	if !e.config.AllowUnregisteredCommands {
		log.Error("Unregistered command not allowed: %s", ctx.Command.Command)
		return nil, fmt.Errorf("unregistered command not allowed: %s", ctx.Command.Command)
	}

	// 非交互式命令的处理
	return e.ExecuteCommand(ctx)
}

// ExecuteCommand 执行单命令
func (e *LocalExecutor) ExecuteCommand(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx == nil || ctx.Command.Command == "" {
		log.Error("No command specified for execution")
		return nil, fmt.Errorf("no command specified")
	}

	if ctx.Options == nil {
		ctx.Options = &types.ExecuteOptions{}
	}
	if ctx.Options.WorkDir == "" && e.config.WorkDir != "" {
		ctx.Options.WorkDir = e.config.WorkDir
	}

	log.Debug("Executing local command: %v", ctx.Command.Args)

	if ctx.IsPiped {
		return e.executePipeline(ctx)
	}

	return e.execute(ctx)
}

func (e *LocalExecutor) execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {

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

// ExecuteInteractive 执行交互式命令
func (e *LocalExecutor) ExecuteInteractive(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Options == nil {
		ctx.Options = &types.ExecuteOptions{}
	}

	// 创建命令
	cmd := exec.CommandContext(ctx.Context, ctx.Command.Command, ctx.Command.Args...)

	// 设置工作目录
	if ctx.Options.WorkDir != "" {
		cmd.Dir = ctx.Options.WorkDir
	}

	if ctx.Options.User == nil {
		ctx.Options.User = &types.User{
			Username: "",
		}
	}

	if ctx.Options.Shell == "" {
		ctx.Options.Shell = "bash"
	}

	// 设置基本的终端环境变量
	defaultEnv := map[string]string{
		"TERM":      ctx.InteractiveOpts.TerminalType,
		"COLORTERM": "truecolor",
		"LANG":      "en_US.UTF-8",
		"LC_ALL":    "en_US.UTF-8",
		"PS1":       "\\w\\$ ", // 设置简单的提示符
	}

	if ctx.Options.WorkDir != "" {
		defaultEnv["HOME"] = ctx.Options.WorkDir
	}
	if ctx.Options.User.Username != "" {
		defaultEnv["USER"] = ctx.Options.User.Username
	}
	if ctx.Options.Shell != "" {
		defaultEnv["SHELL"] = ctx.Options.Shell
	}

	envMap := make(map[string]string)
	envMap = mergeEnv(envMap, defaultEnv)
	envMap = mergeEnv(envMap, ctx.Options.Env)

	// 合并环境变量
	for k, v := range envMap {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// 创建伪终端
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start pty: %w", err)
	}
	defer ptmx.Close()

	// 设置终端大小
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
	startTime := types.GetTimeNow()
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

	endTime := types.GetTimeNow()

	// 等待 IO 完成
	<-doneCh

	result := &types.ExecuteResult{
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

func mergeEnv(env1, env2 map[string]string) map[string]string {
	for k, v := range env2 {
		env1[k] = v
	}
	return env1
}
