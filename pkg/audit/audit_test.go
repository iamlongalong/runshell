package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestAuditor(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "audit_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建审计器
	logFile := filepath.Join(tempDir, "audit.log")
	auditor, err := NewFileAuditor(logFile)
	if err != nil {
		t.Fatalf("Failed to create auditor: %v", err)
	}

	// 记录命令执行
	exec := &types.CommandExecution{
		Command:   types.Command{Command: "test", Args: []string{"arg1", "arg2"}},
		StartTime: time.Now(),
		EndTime:   time.Now(),
		ExitCode:  0,
		Status:    "completed",
	}

	err = auditor.LogCommandExecution(exec)
	assert.NoError(t, err)

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
	assert.Contains(t, logStr, "completed")
	assert.Contains(t, logStr, "ExitCode: 0")
}
