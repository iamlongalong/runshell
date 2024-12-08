package script

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/iamlongalong/runshell/pkg/types"
)

// MockExecutor 是一个用于测试的模拟执行器。
// 实现了 types.Executor 接口的最小子集，只提供测试所需的功能。
// 特性：
// - 可配置的 Execute 方法行为
// - 默认返回成功结果
// - 记录执行参数供验证
type MockExecutor struct {
	executeFunc func(ctx context.Context, cmdName string, args []string, opts *types.ExecuteOptions) (*types.ExecuteResult, error)
}

// Execute 执行模拟的命令。
// 如果设置了 executeFunc，则使用它；
// 否则返回默认的成功结果。
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

// GetCommandInfo 返回空的命令信息（测试用）。
func (m *MockExecutor) GetCommandInfo(cmdName string) (*types.Command, error) {
	return nil, nil
}

// GetCommandHelp 返回空的帮助信息（测试用）。
func (m *MockExecutor) GetCommandHelp(cmdName string) (string, error) {
	return "", nil
}

// ListCommands 返回空的命令列表（测试用）。
func (m *MockExecutor) ListCommands(filter *types.CommandFilter) ([]*types.Command, error) {
	return nil, nil
}

// RegisterCommand 返回成功（测试用）。
func (m *MockExecutor) RegisterCommand(cmd *types.Command) error {
	return nil
}

// UnregisterCommand 返回成功（测试用）。
func (m *MockExecutor) UnregisterCommand(cmdName string) error {
	return nil
}

// TestScriptManager 测试脚本管理器的所有功能。
// 测试场景包括：
// 1. 创建脚本管理器
// 2. 保存脚本
// 3. 列出脚本
// 4. 获取脚本内容
// 5. 执行脚本
// 6. 删除脚本
func TestScriptManager(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "script_test")
	if err != nil {
		t.Fatal(err)
	}
	// 测试完成后清理临时目录
	defer os.RemoveAll(tempDir)

	// 创建模拟执行器
	mockExec := &MockExecutor{}

	// 创建脚本管理器
	manager, err := NewScriptManager(tempDir, mockExec)
	if err != nil {
		t.Fatal(err)
	}

	// 测试场景1：保存脚本
	scriptContent := []byte("#!/bin/sh\necho 'test script'")
	scriptPath, err := manager.SaveScript("test.sh", scriptContent)
	if err != nil {
		t.Errorf("Failed to save script: %v", err)
	}

	// 验证脚本文件是否存在
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Error("Script file was not created")
	}

	// 测试场景2：列出脚本
	scripts, err := manager.ListScripts()
	if err != nil {
		t.Errorf("Failed to list scripts: %v", err)
	}
	if len(scripts) != 1 || scripts[0] != "test.sh" {
		t.Errorf("Expected one script named 'test.sh', got %v", scripts)
	}

	// 测试场景3：获取脚本内容
	content, err := manager.GetScriptContent("test.sh")
	if err != nil {
		t.Errorf("Failed to get script content: %v", err)
	}
	if string(content) != string(scriptContent) {
		t.Errorf("Script content mismatch, got %q, want %q", string(content), string(scriptContent))
	}

	// 测试场景4：执行脚本
	// 设置模拟执行器的行为
	mockExec.executeFunc = func(ctx context.Context, cmdName string, args []string, opts *types.ExecuteOptions) (*types.ExecuteResult, error) {
		// 验证执行参数
		if cmdName != "sh" {
			t.Errorf("Expected command 'sh', got %q", cmdName)
		}
		if len(args) != 1 || args[0] != scriptPath {
			t.Errorf("Expected args [%q], got %v", scriptPath, args)
		}
		return &types.ExecuteResult{
			CommandName: cmdName,
			ExitCode:    0,
			StartTime:   time.Now(),
			EndTime:     time.Now(),
			Output:      "test output",
		}, nil
	}

	// 创建工作目录
	workDir := filepath.Join(tempDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 执行脚本
	result, err := manager.ExecuteScript("test.sh", workDir, nil)
	if err != nil {
		t.Errorf("Failed to execute script: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("Script execution failed with exit code %d", result.ExitCode)
	}

	// 测试场景5：删除脚本
	if err := manager.DeleteScript("test.sh"); err != nil {
		t.Errorf("Failed to delete script: %v", err)
	}

	// 验证脚本是否已被删除
	if _, err := os.Stat(scriptPath); !os.IsNotExist(err) {
		t.Error("Script file was not deleted")
	}

	// 验证脚本列表是否为空
	scripts, err = manager.ListScripts()
	if err != nil {
		t.Errorf("Failed to list scripts after deletion: %v", err)
	}
	if len(scripts) != 0 {
		t.Errorf("Expected no scripts after deletion, got %v", scripts)
	}
}
