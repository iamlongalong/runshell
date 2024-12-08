package executor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/iamlongalong/runshell/pkg/audit"
	"github.com/iamlongalong/runshell/pkg/types"
)

// TestAuditedExecutor 测试带审计功能的执行器。
// 测试场景包括：
// 1. 成功执行命令并记录审计日志
// 2. 处理执行失败的情况
// 3. 命令管理功能（注册、获取信息、列出、注销）
// 4. 审计日志文件的创建和内容验证
func TestAuditedExecutor(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "audited_executor_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	// 测试完成后清理临时目录
	defer os.RemoveAll(tempDir)

	// 创建审计器实例
	auditor, err := audit.NewAuditor(tempDir)
	if err != nil {
		t.Fatalf("Failed to create auditor: %v", err)
	}
	// 确保审计器正确关闭
	defer auditor.Close()

	// 创建模拟执行器，模拟成功执行的情况
	mockExec := &MockExecutor{
		executeFunc: func(ctx context.Context, cmdName string, args []string, opts *types.ExecuteOptions) (*types.ExecuteResult, error) {
			return &types.ExecuteResult{
				CommandName: cmdName,
				ExitCode:    0,
				StartTime:   time.Now(),
				EndTime:     time.Now(),
				Output:      "test output",
			}, nil
		},
	}

	// 创建被测试的审计执行器
	auditedExec := NewAuditedExecutor(mockExec, auditor)

	// 测试场景1：成功执行命令
	result, err := auditedExec.Execute(context.Background(), "test", []string{"arg1", "arg2"}, &types.ExecuteOptions{
		WorkDir: "/test",
		Env:     map[string]string{"KEY": "VALUE"},
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	})

	// 验证执行结果
	if err != nil {
		t.Errorf("Failed to execute command: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	// 验证审计日志文件
	files, err := filepath.Glob(filepath.Join(tempDir, "*.log"))
	if err != nil {
		t.Errorf("Failed to list log files: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 log file, got %d", len(files))
	}

	// 测试场景2：执行失败的情况
	mockExec.executeFunc = func(ctx context.Context, cmdName string, args []string, opts *types.ExecuteOptions) (*types.ExecuteResult, error) {
		return nil, types.ErrCommandNotFound
	}

	// 执行不存在的命令
	result, err = auditedExec.Execute(context.Background(), "invalid", []string{}, &types.ExecuteOptions{})

	// 验证错误处理
	if err != types.ErrCommandNotFound {
		t.Errorf("Expected ErrCommandNotFound, got %v", err)
	}

	if result.ExitCode != -1 {
		t.Errorf("Expected exit code -1, got %d", result.ExitCode)
	}

	// 测试场景3：命令管理功能
	// 创建测试命令
	cmd := &types.Command{
		Name:        "test",
		Description: "Test command",
		Usage:       "test [args...]",
		Category:    "test",
	}

	// 设置模拟执行器的命令信息获取行为
	mockExec.getCommandInfo = func(cmdName string) (*types.Command, error) {
		if cmdName == "test" {
			return cmd, nil
		}
		return nil, types.ErrCommandNotFound
	}

	// 测试命令注册
	if err := auditedExec.RegisterCommand(cmd); err != nil {
		t.Errorf("Failed to register command: %v", err)
	}

	// 测试获取命令信息
	info, err := auditedExec.GetCommandInfo("test")
	if err != nil {
		t.Errorf("Failed to get command info: %v", err)
	}
	if info.Name != cmd.Name || info.Description != cmd.Description || info.Usage != cmd.Usage || info.Category != cmd.Category {
		t.Error("Command info mismatch")
	}

	// 测试获取命令帮助
	help, err := auditedExec.GetCommandHelp("test")
	if err != nil {
		t.Errorf("Failed to get command help: %v", err)
	}
	if help != "" {
		t.Errorf("Expected empty help text, got '%s'", help)
	}

	// 测试列出命令
	commands, err := auditedExec.ListCommands(&types.CommandFilter{Category: "test"})
	if err != nil {
		t.Errorf("Failed to list commands: %v", err)
	}
	if len(commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(commands))
	}

	// 测试注销命令
	if err := auditedExec.UnregisterCommand("test"); err != nil {
		t.Errorf("Failed to unregister command: %v", err)
	}
}
