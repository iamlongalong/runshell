package cmd

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecCommand(t *testing.T) {
	// 创建带取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 保存原始命令状态
	origArgs := rootCmd.Args
	origDockerImage := dockerImage
	defer func() {
		rootCmd.Args = origArgs
		dockerImage = origDockerImage
	}()

	// 在测试环境中禁用 Docker
	dockerImage = ""

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		skipCI  bool // 在 CI 环境中跳过的测试
	}{
		{
			name:    "no args",
			args:    []string{"exec"},
			wantErr: true,
		},
		{
			name:    "valid command",
			args:    []string{"exec", "echo", "test"},
			wantErr: false,
		},
		{
			name:    "invalid command",
			args:    []string{"exec", "invalidcommand123"},
			wantErr: true,
		},
		{
			name:    "command with workdir",
			args:    []string{"exec", "--workdir", "/tmp", "echo", "test"},
			wantErr: false,
		},
		{
			name:    "command with env",
			args:    []string{"exec", "--env", "TEST=value", "env"},
			wantErr: false,
		},
		{
			name:    "docker command",
			args:    []string{"exec", "--docker-image", "ubuntu:latest", "ls"},
			wantErr: false,
			skipCI:  true, // 在 CI 环境中跳过 Docker 测试
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 检查是否在 CI 环境中且需要跳过
			if tt.skipCI && os.Getenv("CI") == "true" {
				t.Skip("Skipping in CI environment")
			}

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
