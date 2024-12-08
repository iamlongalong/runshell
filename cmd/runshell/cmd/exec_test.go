package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecCommand(t *testing.T) {
	// 创建带取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 保存原始命令状态
	origArgs := rootCmd.Args
	defer func() {
		rootCmd.Args = origArgs
	}()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{"exec"},
			wantErr: true,
		},
		{
			name:    "valid command",
			args:    []string{"exec", "ls", "-l"},
			wantErr: false,
		},
		{
			name:    "invalid command",
			args:    []string{"exec", "invalidcommand123"},
			wantErr: true,
		},
		{
			name:    "command with workdir",
			args:    []string{"exec", "--workdir", "/tmp", "ls", "-l"},
			wantErr: false,
		},
		{
			name:    "command with env",
			args:    []string{"exec", "--env", "TEST=value", "env"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置命令状态
			rootCmd.ResetFlags()
			rootCmd.SetArgs(tt.args)

			// 设置上下文
			rootCmd.SetContext(ctx)

			// 执行命令
			err := rootCmd.Execute()

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err, "Expected error for args: %v", tt.args)
			} else {
				assert.NoError(t, err, "Unexpected error for args: %v", tt.args)
			}
		})
	}
}
