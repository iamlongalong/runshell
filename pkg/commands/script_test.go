package commands

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iamlongalong/runshell/pkg/types"
)

// TestScriptCommand 测试脚本命令的所有功能。
// 测试场景包括：
// 1. 创建和初始化
// 2. 保存脚本
// 3. 列出脚本
// 4. 显示脚本内容
// 5. 运行脚本
// 6. 删除脚本
// 7. 错误处理
func TestScriptCommand(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "script_cmd_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	// 测试完成后清理临时目录
	defer os.RemoveAll(tempDir)

	// 创建模拟执行器用于测试
	mockExec := &MockExecutor{}

	// 创建脚本命令实例
	scriptCmd, err := NewScriptCommand(tempDir, mockExec)
	if err != nil {
		t.Fatalf("Failed to create script command: %v", err)
	}

	// 测试场景1：保存脚本
	saveCtx := &types.ExecuteContext{
		Context:   context.Background(),
		Command:   "script",
		Args:      []string{"save", "test.sh", "echo 'test'"},
		StartTime: time.Now(),
		Options: &types.ExecuteOptions{
			WorkDir: tempDir,
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
		},
	}

	// 执行保存命令
	result, err := scriptCmd.Execute(saveCtx)
	if err != nil {
		t.Errorf("Failed to execute save command: %v", err)
	}

	// 验证保存结果
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d: %v", result.ExitCode, result.Error)
	}

	// 等待文件系统更新，确保文件已写入
	time.Sleep(100 * time.Millisecond)

	// 测试场景2：列出脚本
	listCtx := &types.ExecuteContext{
		Context:   context.Background(),
		Command:   "script",
		Args:      []string{"list"},
		StartTime: time.Now(),
		Options: &types.ExecuteOptions{
			WorkDir: tempDir,
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
		},
	}

	// 执行列表命令
	result, err = scriptCmd.Execute(listCtx)
	if err != nil {
		t.Errorf("Failed to execute list command: %v", err)
	}

	// 验证列表结果
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d: %v", result.ExitCode, result.Error)
	}

	// 验证输出包含已保存的脚本
	if !strings.Contains(result.Output, "test.sh") {
		t.Errorf("Expected output to contain 'test.sh', got '%s'", result.Output)
	}

	// 测试场景3：显示脚本内容
	showCtx := &types.ExecuteContext{
		Context:   context.Background(),
		Command:   "script",
		Args:      []string{"show", "test.sh"},
		StartTime: time.Now(),
		Options: &types.ExecuteOptions{
			WorkDir: tempDir,
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
		},
	}

	// 执行显示命令
	result, err = scriptCmd.Execute(showCtx)
	if err != nil {
		t.Errorf("Failed to execute show command: %v", err)
	}

	// 验证显示结果
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d: %v", result.ExitCode, result.Error)
	}

	// 验证脚本内容正确
	if !strings.Contains(result.Output, "echo 'test'") {
		t.Errorf("Expected script content to contain 'echo 'test'', got '%s'", result.Output)
	}

	// 测试场景4：运行脚本
	runCtx := &types.ExecuteContext{
		Context:   context.Background(),
		Command:   "script",
		Args:      []string{"run", "test.sh", tempDir},
		StartTime: time.Now(),
		Options: &types.ExecuteOptions{
			WorkDir: tempDir,
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
		},
	}

	// 设置模拟执行器的行为
	mockExec.executeFunc = func(ctx context.Context, cmdName string, args []string, opts *types.ExecuteOptions) (*types.ExecuteResult, error) {
		return &types.ExecuteResult{
			CommandName: cmdName,
			ExitCode:    0,
			StartTime:   time.Now(),
			EndTime:     time.Now(),
			Output:      "test output",
		}, nil
	}

	// 执行运行命令
	result, err = scriptCmd.Execute(runCtx)
	if err != nil {
		t.Errorf("Failed to execute run command: %v", err)
	}

	// 验证运行结果
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d: %v", result.ExitCode, result.Error)
	}

	// 测试场景5：删除脚本
	deleteCtx := &types.ExecuteContext{
		Context:   context.Background(),
		Command:   "script",
		Args:      []string{"delete", "test.sh"},
		StartTime: time.Now(),
		Options: &types.ExecuteOptions{
			WorkDir: tempDir,
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
		},
	}

	// 执行删除命令
	result, err = scriptCmd.Execute(deleteCtx)
	if err != nil {
		t.Errorf("Failed to execute delete command: %v", err)
	}

	// 验证删除结果
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d: %v", result.ExitCode, result.Error)
	}

	// 验证脚本文件已被物理删除
	files, err := filepath.Glob(filepath.Join(tempDir, "*.sh"))
	if err != nil {
		t.Errorf("Failed to list script files: %v", err)
	}

	if len(files) != 0 {
		t.Error("Script file still exists after deletion")
	}

	// 测试场景6：错误处理 - 无子命令
	noSubcmdCtx := &types.ExecuteContext{
		Context:   context.Background(),
		Command:   "script",
		Args:      []string{},
		StartTime: time.Now(),
		Options: &types.ExecuteOptions{
			WorkDir: tempDir,
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
		},
	}

	// 执行无子命令的情况
	result, err = scriptCmd.Execute(noSubcmdCtx)
	if err != nil {
		t.Errorf("Failed to execute command without subcommand: %v", err)
	}

	// 验证错误处理
	if result.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Output, "Usage:") {
		t.Error("Expected usage information in output")
	}

	// 测试场景7：错误处理 - 无效子命令
	invalidSubcmdCtx := &types.ExecuteContext{
		Context:   context.Background(),
		Command:   "script",
		Args:      []string{"invalid"},
		StartTime: time.Now(),
		Options: &types.ExecuteOptions{
			WorkDir: tempDir,
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
		},
	}

	// 执行无效子命令的情况
	result, err = scriptCmd.Execute(invalidSubcmdCtx)
	if err != nil {
		t.Errorf("Failed to execute command with invalid subcommand: %v", err)
	}

	// 验证错误处理
	if result.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Output, "Usage:") {
		t.Error("Expected usage information in output")
	}
}

// MockExecutor 是一个用于测试的模拟执行器。
// 实现了 types.Executor 接口，但只提供测试所需的最小功能。
type MockExecutor struct {
	// executeFunc 是一个可配置的函数，用于模拟命令执行
	executeFunc func(ctx context.Context, cmdName string, args []string, opts *types.ExecuteOptions) (*types.ExecuteResult, error)
}

// Execute 执行模拟的命令。
// 如果设置了 executeFunc，则使用它；否则返回默认的成功结果。
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

// GetCommandInfo 返回命令信息（测试用）。
func (m *MockExecutor) GetCommandInfo(cmdName string) (*types.Command, error) {
	return nil, nil
}

// GetCommandHelp 返回命令帮助信息（测试用）。
func (m *MockExecutor) GetCommandHelp(cmdName string) (string, error) {
	return "", nil
}

// ListCommands 返回空的命令列表（测试用）
func (m *MockExecutor) ListCommands(filter *types.CommandFilter) ([]*types.Command, error) {
	return nil, nil
}

func (m *MockExecutor) RegisterCommand(cmd *types.Command) error {
	return nil
}

func (m *MockExecutor) UnregisterCommand(cmdName string) error {
	return nil
}
