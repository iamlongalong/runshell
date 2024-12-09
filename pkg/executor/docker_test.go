package executor

import (
	"context"
	"testing"

	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestDockerExecutor(t *testing.T) {
	// 创建 Docker 执行器
	exec := NewDockerExecutor(DockerConfig{
		Image: "alpine:latest",
	})

	// 测试执行命令
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Args:    []string{"echo", "hello"},
		Options: &types.ExecuteOptions{},
	}

	result, err := exec.Execute(ctx)
	if err != nil {
		t.Skipf("Docker test skipped: %v", err)
	}

	assert.Equal(t, 0, result.ExitCode)

	// 测试列出命令
	commands := exec.ListCommands()
	assert.NotNil(t, commands)
}

func TestDockerExecutorWithWorkDir(t *testing.T) {
	// 创建 Docker 执行器
	exec := NewDockerExecutor(DockerConfig{
		Image:   "alpine:latest",
		WorkDir: "/tmp",
	})

	// 测试执行命令
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Args:    []string{"pwd"},
		Options: &types.ExecuteOptions{
			WorkDir: "/tmp",
		},
	}

	result, err := exec.Execute(ctx)
	if err != nil {
		t.Skipf("Docker test skipped: %v", err)
	}

	assert.Equal(t, 0, result.ExitCode)
}

func TestDockerExecutorWithEnv(t *testing.T) {
	// 创建 Docker 执行器
	exec := NewDockerExecutor(DockerConfig{
		Image:   "alpine:latest",
		WorkDir: "/tmp",
	})

	// 测试执行命令
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Args:    []string{"env"},
		Options: &types.ExecuteOptions{
			Env: map[string]string{
				"TEST_VAR": "test_value",
			},
		},
	}

	result, err := exec.Execute(ctx)
	if err != nil {
		t.Skipf("Docker test skipped: %v", err)
	}

	assert.Equal(t, 0, result.ExitCode)
}
