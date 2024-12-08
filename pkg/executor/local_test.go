package executor

import (
	"context"
	"os"
	"testing"

	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestLocalExecutor(t *testing.T) {
	// 创建本地执行器
	exec := NewLocalExecutor()

	// 测试执行命令
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Args:    []string{"echo", "hello"},
		Options: &types.ExecuteOptions{},
	}

	result, err := exec.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)

	// 测试列出命令
	commands := exec.ListCommands()
	assert.NotNil(t, commands)
}

func TestLocalExecutorWithWorkDir(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "executor_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建本地执行器
	exec := NewLocalExecutor()

	// 测试执行命令
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Args:    []string{"pwd"},
		Options: &types.ExecuteOptions{
			WorkDir: tempDir,
		},
	}

	result, err := exec.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
}

func TestLocalExecutorWithEnv(t *testing.T) {
	// 创建本地执行器
	exec := NewLocalExecutor()

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
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
}

func TestLocalExecutorWithPipe(t *testing.T) {
	// 创建本地执行器
	exec := NewLocalExecutor()

	// 创建管道执行器
	pipeExec := NewPipelineExecutor(exec)

	// 测试管道命令
	pipeline, err := pipeExec.ParsePipeline("echo hello | grep hello")
	assert.NoError(t, err)

	pipeline.Context = context.Background()
	pipeline.Options = &types.ExecuteOptions{}

	result, err := pipeExec.ExecutePipeline(pipeline)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
}
