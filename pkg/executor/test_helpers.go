// Package executor 实现了命令执行器的核心功能。
// 本文件提供了用于测试的辅助类型和函数。
package executor

import (
	"context"
	"time"

	"github.com/iamlongalong/runshell/pkg/types"
)

// testCommandHandler 是一个用于测试的命令处理器。
// 实现了 types.CommandHandler 接口，用于模拟命令执行。
// 特性：
// - 可配置固定的输出内容
// - 始终返回成功的执行结果
// - 支持输出重定向
type testCommandHandler struct {
	output string // 预设的输出内容
}

// Execute 执行测试命令。
// 功能：
// - 将预设的输出内容写入标准输出
// - 返回成功的执行结果
//
// 参数：
//   - ctx：执行上下文
//
// 返回值：
//   - *types.ExecuteResult：执行结果
//   - error：执行过程中的错误
func (h *testCommandHandler) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Options.Stdout != nil {
		_, err := ctx.Options.Stdout.Write([]byte(h.output))
		if err != nil {
			return nil, err
		}
	}
	return &types.ExecuteResult{
		CommandName: ctx.Command,
		ExitCode:    0,
		StartTime:   ctx.StartTime,
		EndTime:     time.Now(),
		Output:      h.output,
	}, nil
}

// MockExecutor 是一个用于测试的模拟执行器。
// 实现了 types.Executor 接口，提供可配置的行为。
// 特性：
// - 所有方法行为都可以通过函数字段自定义
// - 提供默认的成功返回值
// - 适用于单元测试和集成测试
type MockExecutor struct {
	// executeFunc 定义 Execute 方法的行为
	executeFunc func(ctx context.Context, cmdName string, args []string, opts *types.ExecuteOptions) (*types.ExecuteResult, error)
	// getCommandInfo 定义 GetCommandInfo 方法的行为
	getCommandInfo func(cmdName string) (*types.Command, error)
	// getCommandHelp 定义 GetCommandHelp 方法的行为
	getCommandHelp func(cmdName string) (string, error)
	// listCommands 定义 ListCommands 方法的行为
	listCommands func(filter *types.CommandFilter) ([]*types.Command, error)
	// registerCommand 定义 RegisterCommand 方法的行为
	registerCommand func(cmd *types.Command) error
	// unregisterCommand 定义 UnregisterCommand 方法的行为
	unregisterCommand func(cmdName string) error
}

// Execute 执行模拟的命令。
// 如果设置了 executeFunc，则使用它；
// 否则返回默认的成功结果。
//
// 参数：
//   - ctx：上下文
//   - cmdName：命令名称
//   - args：命令参数
//   - opts：执行选项
//
// 返回值：
//   - *types.ExecuteResult：执行结果
//   - error：执行错误
func (m *MockExecutor) Execute(ctx context.Context, cmdName string, args []string, opts *types.ExecuteOptions) (*types.ExecuteResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, cmdName, args, opts)
	}
	return &types.ExecuteResult{
		CommandName: cmdName,
		ExitCode:    0,
		StartTime:   time.Now(),
		EndTime:     time.Now(),
		Output:      "mock output",
	}, nil
}

// GetCommandInfo 获取模拟的命令信息。
// 如果设置了 getCommandInfo，则使用它；
// 否则返回空值。
//
// 参数：
//   - cmdName：命令名称
//
// 返回值：
//   - *types.Command：命令信息
//   - error：获取错误
func (m *MockExecutor) GetCommandInfo(cmdName string) (*types.Command, error) {
	if m.getCommandInfo != nil {
		return m.getCommandInfo(cmdName)
	}
	return nil, nil
}

// GetCommandHelp 获取模拟的命令帮助信息。
// 如果设置了 getCommandHelp，则使用它；
// 否则返回空字符串。
//
// 参数：
//   - cmdName：命令名称
//
// 返回值：
//   - string：帮助信息
//   - error：获取错误
func (m *MockExecutor) GetCommandHelp(cmdName string) (string, error) {
	if m.getCommandHelp != nil {
		return m.getCommandHelp(cmdName)
	}
	return "", nil
}

// ListCommands 列出模拟的命令列表。
// 如果设置了 listCommands，则使用它；
// 否则返回默认的测试命令列表。
//
// 参数：
//   - filter：命令过滤器
//
// 返回值：
//   - []*types.Command：命令列表
//   - error：列出错误
func (m *MockExecutor) ListCommands(filter *types.CommandFilter) ([]*types.Command, error) {
	if m.listCommands != nil {
		return m.listCommands(filter)
	}
	return []*types.Command{
		{
			Name:        "test",
			Description: "Test command",
			Usage:       "test [args...]",
			Category:    "test",
		},
	}, nil
}

// RegisterCommand 注册模拟的命令。
// 如果设置了 registerCommand，则使用它；
// 否则返回成功。
//
// 参数：
//   - cmd：要注册的命令
//
// 返回值：
//   - error：注册错误
func (m *MockExecutor) RegisterCommand(cmd *types.Command) error {
	if m.registerCommand != nil {
		return m.registerCommand(cmd)
	}
	return nil
}

// UnregisterCommand 注销模拟的命令。
// 如果设置了 unregisterCommand，则使用它；
// 否则返回成功。
//
// 参数：
//   - cmdName：要注销的命令名称
//
// 返回值：
//   - error：注销错误
func (m *MockExecutor) UnregisterCommand(cmdName string) error {
	if m.unregisterCommand != nil {
		return m.unregisterCommand(cmdName)
	}
	return nil
}
