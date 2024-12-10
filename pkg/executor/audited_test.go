package executor

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/iamlongalong/runshell/pkg/audit"
	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestAuditedExecutor(t *testing.T) {
	// 创建临时文件用于审计日志
	tmpFile, err := os.CreateTemp("", "audit-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// 创建审计器
	auditor, err := audit.NewFileAuditor(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create auditor: %v", err)
	}

	// 创建模拟执行器
	mockExec := &MockExecutor{
		ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
			return &types.ExecuteResult{
				CommandName: ctx.Command.Command,
				ExitCode:    0,
				StartTime:   types.GetTimeNow(),
				EndTime:     types.GetTimeNow(),
			}, nil
		},
	}

	// 创建审计执行器
	exec := NewAuditedExecutor(mockExec, auditor)

	// 执行命令
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Command: types.Command{
			Command: "test",
			Args:    []string{"arg1", "arg2"},
		},
		Options: &types.ExecuteOptions{},
	}

	result, err := exec.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)

	// 读取审计日志
	content, err := os.ReadFile(tmpFile.Name())
	assert.NoError(t, err)

	// 检查审计日志内容
	logContent := string(content)
	assert.Contains(t, strings.ToLower(logContent), "started")
	assert.Contains(t, strings.ToLower(logContent), "completed")
}
