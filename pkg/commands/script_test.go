package commands

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestScriptCommand(t *testing.T) {
	// 创建模拟执行器
	mockExec := &executor.MockExecutor{
		ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
			return &types.ExecuteResult{
				CommandName: "script",
				ExitCode:    0,
				StartTime:   types.GetTimeNow(),
				EndTime:     types.GetTimeNow(),
			}, nil
		},
	}

	// 创建脚本命令
	cmd := NewScriptCommand(mockExec)

	// 创建输出缓冲区
	var stdout, stderr bytes.Buffer

	// 创建基本选项
	opts := &types.ExecuteOptions{
		WorkDir: os.TempDir(),
		Env:     map[string]string{},
		Stdout:  &stdout,
		Stderr:  &stderr,
		Stdin:   os.Stdin,
	}

	// 首先保存一个脚本
	saveCtx := &types.ExecuteContext{
		Context: context.Background(),
		Args:    []string{"save", "test", "echo", "hello"},
		Options: opts,
	}

	result, err := cmd.Execute(saveCtx)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)

	// 清空缓冲区
	stdout.Reset()
	stderr.Reset()

	// 然后运行其他测试
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
		check   func(t *testing.T, stdout string)
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
			errMsg:  "requires at least 2 arguments",
		},
		{
			name:    "invalid action",
			args:    []string{"invalid", "test"},
			wantErr: true,
			errMsg:  "unknown action",
		},
		{
			name:    "list scripts",
			args:    []string{"list"},
			wantErr: false,
			check: func(t *testing.T, stdout string) {
				assert.Contains(t, stdout, "test")
			},
		},
		{
			name:    "get script",
			args:    []string{"get", "test"},
			wantErr: false,
			check: func(t *testing.T, stdout string) {
				assert.Contains(t, stdout, "echo")
				assert.Contains(t, stdout, "hello")
			},
		},
		{
			name:    "run script",
			args:    []string{"run", "test"},
			wantErr: false,
		},
		{
			name:    "delete script",
			args:    []string{"delete", "test"},
			wantErr: false,
		},
		{
			name:    "verify deletion",
			args:    []string{"list"},
			wantErr: false,
			check: func(t *testing.T, stdout string) {
				assert.NotContains(t, stdout, "test")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清空缓冲区
			stdout.Reset()
			stderr.Reset()

			// 创建执行上下���
			ctx := &types.ExecuteContext{
				Context: context.Background(),
				Args:    tt.args,
				Options: opts,
			}

			// 执行命令
			result, err := cmd.Execute(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 0, result.ExitCode)
				if tt.check != nil {
					tt.check(t, stdout.String())
				}
			}
		})
	}
}
