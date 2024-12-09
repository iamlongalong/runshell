// Package executor 实现了命令执行器的核心功能。
// 本文件提供了用于测试的辅助类型和函数。
package executor

import (
	"github.com/iamlongalong/runshell/pkg/types"
)

// MockExecutor 模拟执行器
type MockExecutor struct {
	ExecuteFunc func(ctx *types.ExecuteContext) (*types.ExecuteResult, error)
	CloseFunc   func() error
	Options     *types.ExecuteOptions
}

func (m *MockExecutor) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx)
	}
	return &types.ExecuteResult{}, nil
}

func (m *MockExecutor) ListCommands() []types.CommandInfo {
	return []types.CommandInfo{
		{
			Name:        "test",
			Description: "Test command",
			Usage:       "test [args...]",
			Category:    "test",
		},
	}
}

func (m *MockExecutor) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockExecutor) SetOptions(options *types.ExecuteOptions) {
	m.Options = options
}
