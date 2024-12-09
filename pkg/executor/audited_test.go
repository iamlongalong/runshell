package executor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/iamlongalong/runshell/pkg/audit"
	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestAuditedExecutor(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "audit_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建审计器
	logFile := filepath.Join(tempDir, "audit.log")
	auditor, err := audit.NewAuditor(logFile)
	if err != nil {
		t.Fatalf("Failed to create auditor: %v", err)
	}

	// 创建模拟执行器
	mockExec := &MockExecutor{
		ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
			return &types.ExecuteResult{
				CommandName: ctx.Args[0],
				ExitCode:    0,
				StartTime:   types.GetTimeNow(),
				EndTime:     types.GetTimeNow(),
			}, nil
		},
	}

	// 创建审计执行器
	auditedExec := NewAuditedExecutor(mockExec, auditor)

	// 测试执行命令
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Args:    []string{"test", "arg1", "arg2"},
		Options: &types.ExecuteOptions{
			WorkDir: "/tmp",
			Env:     map[string]string{"FOO": "bar"},
		},
	}

	result, err := auditedExec.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)

	// 验证日志文件存在
	_, err = os.Stat(logFile)
	assert.NoError(t, err)

	// 读取日志内容
	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)

	// 验证日志内容
	logStr := string(content)
	assert.Contains(t, logStr, "test")
	assert.Contains(t, logStr, "arg1")
	assert.Contains(t, logStr, "arg2")
	assert.Contains(t, logStr, "started")
	assert.Contains(t, logStr, "completed")
	assert.Contains(t, logStr, "ExitCode: 0")

	// 测试列出命令
	commands := auditedExec.ListCommands()
	assert.Len(t, commands, 1)
	assert.Equal(t, "test", commands[0].Name)
}
