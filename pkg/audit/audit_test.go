package audit

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/iamlongalong/runshell/pkg/types"
)

// TestAuditor 测试审计器的核心功能。
// 测试内容包括：
// 1. 创建审计器
// 2. 记录命令执行
// 3. 读取审计日志
// 4. 搜索审计日志
// 5. 验证日志文件
func TestAuditor(t *testing.T) {
	// 创建临时目录用于存储测试日志文件
	tempDir, err := os.MkdirTemp("", "audit_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	// 测试完成后清理临时目录
	defer os.RemoveAll(tempDir)

	// 创建审计器实例
	auditor, err := NewAuditor(tempDir)
	if err != nil {
		t.Fatalf("Failed to create auditor: %v", err)
	}
	// 确保测试结束时关闭审计器
	defer auditor.Close()

	// 准备测试数据：创建执行上下文
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Command: "test",
		Args:    []string{"arg1", "arg2"},
		Options: &types.ExecuteOptions{
			WorkDir: "/test/dir",
			Env:     map[string]string{"KEY": "VALUE"},
		},
		StartTime: time.Now(),
	}

	// 准备测试数据：创建执行结果
	result := &types.ExecuteResult{
		CommandName: "test",
		ExitCode:    0,
		StartTime:   ctx.StartTime,
		EndTime:     ctx.StartTime.Add(time.Second),
		Output:      "test output",
	}

	// 测试场景1：记录命令执行
	if err := auditor.LogCommandExecution(ctx, result); err != nil {
		t.Errorf("Failed to log command execution: %v", err)
	}

	// 测试场景2：读取时间范围内的日志
	events, err := auditor.GetAuditLogs(ctx.StartTime.Add(-time.Hour), ctx.StartTime.Add(time.Hour))
	if err != nil {
		t.Errorf("Failed to get audit logs: %v", err)
	}

	// 验证读取到的日志数量
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	// 验证日志内容正确性
	if events[0].CommandName != "test" {
		t.Errorf("Expected command name 'test', got '%s'", events[0].CommandName)
	}

	// 测试场景3：按条件搜索日志
	events, err = auditor.SearchAuditLogs(map[string]string{
		"command": "test",
	})
	if err != nil {
		t.Errorf("Failed to search audit logs: %v", err)
	}

	// 验证搜索结果
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	// 测试场景4：验证日志文件创建
	logFiles, err := filepath.Glob(filepath.Join(tempDir, "*.log"))
	if err != nil {
		t.Errorf("Failed to list log files: %v", err)
	}

	// 验证日志文件数量
	if len(logFiles) != 1 {
		t.Errorf("Expected 1 log file, got %d", len(logFiles))
	}
}
