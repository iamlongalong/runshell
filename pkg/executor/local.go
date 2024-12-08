package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/iamlongalong/runshell/pkg/types"
)

// LocalExecutor 实现了本地命令执行器。
// 特性：
// - 支持内置命令和系统命令
// - 线程安全的命令注册表
// - 环境变量和工作目录配置
// - 用户权限控制（TODO）
type LocalExecutor struct {
	commands sync.Map // 存储注册的命令，key 为命令名，value 为 *types.Command
}

// NewLocalExecutor 创建一个新的本地执行器实例。
// 返回值：
//   - *LocalExecutor：初始化好的本地执行器
func NewLocalExecutor() *LocalExecutor {
	return &LocalExecutor{}
}

// Execute 执行指定的命令。
// 执行过程：
// 1. 首先检查是否为注册的内置命令
// 2. 如果不是内置命令，则尝试在系统 PATH 中查找并执行
//
// 参数：
//   - ctx：上下文，用于取消和超时控制
//   - cmdName：要执行的命令名称
//   - args：命令参数列表
//   - opts：执行选项（工作目录、环境变量等）
//
// 返回值：
//   - *types.ExecuteResult：执行结果
//   - error：执行过程中的错误
func (e *LocalExecutor) Execute(ctx context.Context, cmdName string, args []string, opts *types.ExecuteOptions) (*types.ExecuteResult, error) {
	if opts == nil {
		opts = &types.ExecuteOptions{
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}
	}

	// 检查命令是否已注册
	if cmd, ok := e.commands.Load(cmdName); ok {
		command := cmd.(*types.Command)
		execCtx := &types.ExecuteContext{
			Context:   ctx,
			Args:      args,
			Options:   opts,
			StartTime: time.Now(),
		}
		return command.Handler.Execute(execCtx)
	}

	// 如果命令未注册，尝试在系统PATH中查找
	return e.executeSystemCommand(ctx, cmdName, args, opts)
}

// executeSystemCommand 执行系统命令。
// 内部函数，用于执行未注册的系统命令。
//
// 功能：
// - 在系统 PATH 中查找命令
// - 设置工作目录和环境变量
// - 处理输入输出流
// - 收集执行结果和资源使用情况
//
// 参数：
//   - ctx：上下文
//   - cmdName：命令名称
//   - args：命令参数
//   - opts：执行选项
func (e *LocalExecutor) executeSystemCommand(ctx context.Context, cmdName string, args []string, opts *types.ExecuteOptions) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: cmdName,
		StartTime:   time.Now(),
		ExitCode:    -1, // 默认为错误状态
	}

	// 检查命令是否存在
	_, err := exec.LookPath(cmdName)
	if err != nil {
		result.Error = fmt.Errorf("command not found: %s", cmdName)
		return result, result.Error
	}

	cmd := exec.CommandContext(ctx, cmdName, args...)

	// 设置工作目录
	if opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	// 设置环境变量
	if opts.Env != nil {
		env := os.Environ()
		for k, v := range opts.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = env
	}

	// 设置输入输出流
	cmd.Stdin = opts.Stdin
	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr

	// 设置用户权限
	if opts.User != nil {
		// TODO: implement user switching
	}

	// 启动命令
	err = cmd.Start()
	if err != nil {
		result.Error = err
		result.EndTime = time.Now()
		return result, err
	}

	// 等待命令完成
	err = cmd.Wait()
	result.EndTime = time.Now()

	// 处理执行结果
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		result.Error = err
		return result, err
	}

	result.ExitCode = 0 // 成功时设置为0

	// TODO: implement resource usage collection
	result.ResourceUsage = types.ResourceUsage{}

	return result, nil
}

// GetCommandInfo 获取命令信息。
// 参数：
//   - cmdName：命令名称
//
// 返回值：
//   - *types.Command：命令信息
//   - error：如果命令不存在则返回 ErrCommandNotFound
func (e *LocalExecutor) GetCommandInfo(cmdName string) (*types.Command, error) {
	if cmd, ok := e.commands.Load(cmdName); ok {
		return cmd.(*types.Command), nil
	}
	return nil, types.ErrCommandNotFound
}

// GetCommandHelp 获取命令的帮助信息。
// 参数：
//   - cmdName：命令名称
//
// 返回值：
//   - string：命令的使用说明
//   - error：如果命令不存在则返回错误
func (e *LocalExecutor) GetCommandHelp(cmdName string) (string, error) {
	cmd, err := e.GetCommandInfo(cmdName)
	if err != nil {
		return "", err
	}
	return cmd.Usage, nil
}

// ListCommands 列出符合过滤条件的命令。
// 参数：
//   - filter：命令过滤器，可按类别过滤
//
// 返回值：
//   - []*types.Command：符合条件的命令列表
//   - error：列出过程中的错误
func (e *LocalExecutor) ListCommands(filter *types.CommandFilter) ([]*types.Command, error) {
	var commands []*types.Command
	e.commands.Range(func(key, value interface{}) bool {
		cmd := value.(*types.Command)
		if filter == nil || (filter.Category == "" || filter.Category == cmd.Category) {
			commands = append(commands, cmd)
		}
		return true
	})
	return commands, nil
}

// RegisterCommand 注册新命令。
// 参数：
//   - cmd：要注册的命令
//
// 返回值：
//   - error：如果命令无效则返回错误
func (e *LocalExecutor) RegisterCommand(cmd *types.Command) error {
	if cmd == nil || cmd.Name == "" {
		return fmt.Errorf("invalid command")
	}
	e.commands.Store(cmd.Name, cmd)
	return nil
}

// UnregisterCommand 注销已注册的命令。
// 参数：
//   - cmdName：要注销的命令名称
//
// 返回值：
//   - error：注销过程中的错误
func (e *LocalExecutor) UnregisterCommand(cmdName string) error {
	e.commands.Delete(cmdName)
	return nil
}
