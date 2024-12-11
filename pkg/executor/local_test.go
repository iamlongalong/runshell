package executor

import (
	"bytes"
	"context"
	"testing"

	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

// testCommand 是一个用于测试的命令实现
type testCommand struct {
	name        string
	description string
	usage       string
	output      string
	exitCode    int
}

func (c *testCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Options != nil && ctx.Options.Stdout != nil {
		ctx.Options.Stdout.Write([]byte(c.output))
	}
	return &types.ExecuteResult{
		CommandName: c.name,
		ExitCode:    c.exitCode,
		Output:      c.output,
	}, nil
}

func (c *testCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        c.name,
		Description: c.description,
		Usage:       c.usage,
	}
}

func TestLocalExecutor(t *testing.T) {
	exec := NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, nil)

	var output bytes.Buffer
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Command: types.Command{
			Command: "echo",
			Args:    []string{"hello"},
		},
		Options: &types.ExecuteOptions{
			Stdout: &output,
		},
	}

	result, err := exec.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, output.String(), "hello")
}

func TestLocalExecutorWithWorkDir(t *testing.T) {
	exec := NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, nil)

	// 创建临时目录
	tempDir := t.TempDir()

	var output bytes.Buffer
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Command: types.Command{
			Command: "pwd",
		},
		Options: &types.ExecuteOptions{
			WorkDir: tempDir,
			Stdout:  &output,
		},
	}

	result, err := exec.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, output.String(), tempDir)
}

func TestLocalExecutorWithEnv(t *testing.T) {
	exec := NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, nil)

	var output bytes.Buffer
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Command: types.Command{
			Command: "env",
		},
		Options: &types.ExecuteOptions{
			Env: map[string]string{
				"TEST_VAR": "test_value",
			},
			Stdout: &output,
		},
	}

	result, err := exec.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, output.String(), "TEST_VAR=test_value")
}

func TestLocalExecutorWithPipe(t *testing.T) {
	tests := []struct {
		name      string
		commands  []*types.Command
		wantErr   bool
		exitCode  int
		checkFunc func(t *testing.T, output string)
	}{
		{
			name: "Simple pipeline",
			commands: []*types.Command{
				{Command: "echo", Args: []string{"total 123"}},
				{Command: "grep", Args: []string{"total"}},
			},
			wantErr:  false,
			exitCode: 0,
			checkFunc: func(t *testing.T, output string) {
				assert.NotEmpty(t, output)
				assert.Contains(t, output, "total")
			},
		},
		{
			name: "Pipeline with no matches",
			commands: []*types.Command{
				{Command: "echo", Args: []string{"hello world"}},
				{Command: "grep", Args: []string{"nonexistentfile"}},
			},
			wantErr:  false,
			exitCode: 1,
			checkFunc: func(t *testing.T, output string) {
				assert.Empty(t, output)
			},
		},
		{
			name: "Multi-stage pipeline",
			commands: []*types.Command{
				{Command: "echo", Args: []string{"conf1\nconf2\nconf3"}},
				{Command: "grep", Args: []string{"conf"}},
				{Command: "sort"},
			},
			wantErr:  false,
			exitCode: 0,
			checkFunc: func(t *testing.T, output string) {
				assert.NotEmpty(t, output)
				assert.Contains(t, output, "conf")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec := NewLocalExecutor(types.LocalConfig{
				AllowUnregisteredCommands: true,
			}, nil)

			var output bytes.Buffer
			pipeline := &types.PipelineContext{
				Context:  context.Background(),
				Commands: tt.commands,
				Options: &types.ExecuteOptions{
					Stdout: &output,
				},
			}

			ctx := &types.ExecuteContext{
				Context:     context.Background(),
				PipeContext: pipeline,
				Executor:    exec,
			}

			result, err := exec.executePipeline(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			if result.ExitCode != 0 {
				assert.Equal(t, tt.exitCode, result.ExitCode)
				return
			}

			assert.NoError(t, err)
			if tt.checkFunc != nil {
				tt.checkFunc(t, output.String())
			}
		})
	}
}

func TestLocalExecutor_Execute(t *testing.T) {
	exec := NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, nil)

	cmd := &testCommand{
		name:        "test_command",
		description: "Test command",
		usage:       "test_command [args...]",
		output:      "hello",
		exitCode:    0,
	}

	err := exec.RegisterCommand(cmd)
	assert.NoError(t, err)

	var output bytes.Buffer
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Command: types.Command{
			Command: "test_command",
		},
		Options: &types.ExecuteOptions{
			Stdout: &output,
		},
	}

	result, err := exec.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "hello", output.String())
}

func TestLocalExecutor_RegisterCommand(t *testing.T) {
	exec := NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, nil)

	cmd := &testCommand{
		name:        "test_command",
		description: "Test command",
		usage:       "test_command [args...]",
	}

	err := exec.RegisterCommand(cmd)
	assert.NoError(t, err)

	// 测试注册空命令
	err = exec.RegisterCommand(nil)
	assert.Error(t, err)

	// 测试注册空名称命令
	cmd = &testCommand{}
	err = exec.RegisterCommand(cmd)
	assert.Error(t, err)
}

func TestLocalExecutor_UnregisterCommand(t *testing.T) {
	exec := NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, nil)

	cmd := &testCommand{
		name:        "test_command",
		description: "Test command",
		usage:       "test_command [args...]",
	}

	err := exec.RegisterCommand(cmd)
	assert.NoError(t, err)

	err = exec.UnregisterCommand("test_command")
	assert.NoError(t, err)

	// 测试注销不存在的命令
	err = exec.UnregisterCommand("nonexistent")
	assert.NoError(t, err)
}

func TestLocalExecutor_ListCommands(t *testing.T) {
	exec := NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, nil)

	cmd := &testCommand{
		name:        "test_command",
		description: "Test command",
		usage:       "test_command [args...]",
	}

	err := exec.RegisterCommand(cmd)
	assert.NoError(t, err)

	commands := exec.ListCommands()
	assert.Len(t, commands, 1)
	assert.Equal(t, "test_command", commands[0].Name)
	assert.Equal(t, "Test command", commands[0].Description)
	assert.Equal(t, "test_command [args...]", commands[0].Usage)
}

func TestLocalExecutorPipeline(t *testing.T) {
	tests := []struct {
		name      string
		pipeline  *types.PipelineContext
		wantErr   bool
		errMsg    string
		checkFunc func(t *testing.T, result *types.ExecuteResult)
	}{
		{
			name: "basic pipeline - ls and grep",
			pipeline: &types.PipelineContext{
				Context: context.Background(),
				Commands: []*types.Command{
					{Command: "echo", Args: []string{"total 123"}},
					{Command: "grep", Args: []string{"total"}},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *types.ExecuteResult) {
				assert.NotEmpty(t, result.Output)
				assert.Contains(t, result.Output, "total")
			},
		},
		{
			name: "pipeline with built-in command",
			pipeline: &types.PipelineContext{
				Context: context.Background(),
				Commands: []*types.Command{
					{Command: "echo", Args: []string{"hello world"}},
					{Command: "grep", Args: []string{"hello"}},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *types.ExecuteResult) {
				assert.Contains(t, result.Output, "hello world")
			},
		},
		{
			name: "pipeline with unregistered commands not allowed",
			pipeline: &types.PipelineContext{
				Context: context.Background(),
				Commands: []*types.Command{
					{Command: "ls"},
					{Command: "grep"},
				},
			},
			wantErr: true,
			errMsg:  "unregistered command not allowed",
		},
		{
			name: "pipeline with invalid command",
			pipeline: &types.PipelineContext{
				Context: context.Background(),
				Commands: []*types.Command{
					{Command: "invalidcmd123"},
				},
			},
			wantErr: true,
			errMsg:  "unregistered command not allowed",
		},
		{
			name: "pipeline with multiple built-in commands",
			pipeline: &types.PipelineContext{
				Context: context.Background(),
				Commands: []*types.Command{
					{Command: "echo", Args: []string{"hello"}},
					{Command: "grep", Args: []string{"hello"}},
					{Command: "wc", Args: []string{"-l"}},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result *types.ExecuteResult) {
				assert.Contains(t, result.Output, "1")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec := NewLocalExecutor(types.LocalConfig{
				AllowUnregisteredCommands: !tt.wantErr,
			}, nil)

			cmd := &testCommand{
				name:        "test_command",
				description: "Test command",
				usage:       "test_command [args...]",
				output:      "hello world",
			}

			err := exec.RegisterCommand(cmd)
			assert.NoError(t, err)

			ctx := &types.ExecuteContext{
				Context:     context.Background(),
				PipeContext: tt.pipeline,
				Executor:    exec,
			}

			result, err := exec.executePipeline(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestLocalExecutorPipelineEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		pipeline  *types.PipelineContext
		wantErr   bool
		errMsg    string
		checkFunc func(t *testing.T, result *types.ExecuteResult)
	}{
		{
			name:     "nil pipe context",
			pipeline: nil,
			wantErr:  true,
			errMsg:   "no commands in pipeline",
		},
		{
			name: "empty commands list",
			pipeline: &types.PipelineContext{
				Context:  context.Background(),
				Commands: []*types.Command{},
			},
			wantErr: true,
			errMsg:  "no commands in pipeline",
		},
		{
			name: "command with empty name",
			pipeline: &types.PipelineContext{
				Context: context.Background(),
				Commands: []*types.Command{
					{Command: ""},
				},
			},
			wantErr: true,
			errMsg:  "command not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec := NewLocalExecutor(types.LocalConfig{
				AllowUnregisteredCommands: true,
			}, nil)

			ctx := &types.ExecuteContext{
				Context:     context.Background(),
				PipeContext: tt.pipeline,
				Executor:    exec,
			}

			result, err := exec.executePipeline(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestLocalExecutorPipelineCancel(t *testing.T) {
	exec := NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	pipeline := &types.PipelineContext{
		Context: ctx,
		Commands: []*types.Command{
			{Command: "sleep", Args: []string{"10"}},
		},
	}

	// 立即取消上下文
	cancel()

	ectx := &types.ExecuteContext{
		Context:     ctx,
		PipeContext: pipeline,
		Executor:    exec,
	}

	result, err := exec.executePipeline(ectx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	assert.Nil(t, result)
}
