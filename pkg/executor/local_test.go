package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestLocalExecutor(t *testing.T) {
	// 创建本地执行器
	exec := NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{})

	// 测试执行命令
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Command: types.Command{
			Command: "echo",
			Args:    []string{"hello"},
		},
		Options: &types.ExecuteOptions{},
	}

	result, err := exec.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Output, "hello")
	// 测试列出命令
	exec.RegisterCommand(&testCommand{
		info: types.CommandInfo{
			Name:        "test_command",
			Description: "Test command",
			Usage:       "test_command [args...]",
		},
	})
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
	exec := NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{})

	// 测试执行命令
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Command: types.Command{
			Command: "pwd",
		},
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
	exec := NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{})

	// 测试执行命令
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Command: types.Command{
			Command: "env",
		},
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
	localExec := NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{})

	tests := []struct {
		name      string
		commands  [][]string
		wantErr   bool
		exitCode  int
		checkFunc func(t *testing.T, result *types.ExecuteResult)
	}{
		{
			name: "Simple pipeline",
			commands: [][]string{
				{"ls", "-la"},
				{"grep", "total"},
			},
			wantErr:  false,
			exitCode: 0,
			checkFunc: func(t *testing.T, result *types.ExecuteResult) {
				assert.NotEmpty(t, result.Output)
			},
		},
		{
			name: "Pipeline with no matches",
			commands: [][]string{
				{"ls", "-la"},
				{"grep", "nonexistentfile"},
			},
			wantErr:  false,
			exitCode: 0,
			checkFunc: func(t *testing.T, result *types.ExecuteResult) {
				assert.Empty(t, result.Output)
			},
		},
		{
			name: "Multi-stage pipeline",
			commands: [][]string{
				{"ls", "-la", "/etc"},
				{"grep", "conf"},
				{"sort"},
			},
			wantErr:  false,
			exitCode: 0,
			checkFunc: func(t *testing.T, result *types.ExecuteResult) {
				assert.NotEmpty(t, result.Output)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建管道执行器
			pipeExec := NewPipelineExecutor(localExec)

			// 设置执行选项
			options := &types.ExecuteOptions{}

			// 创建管道上下文
			pipeline := &types.PipelineContext{
				Commands: make([]*types.Command, len(tt.commands)),
				Options:  options,
				Context:  context.Background(),
			}

			// 设置每个命令
			for i, cmd := range tt.commands {
				pipeline.Commands[i] = &types.Command{
					Command: cmd[0],
					Args:    cmd[1:],
				}
			}

			// 执行管道命令
			result, err := pipeExec.ExecutePipeline(pipeline)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if tt.name == "Pipeline with no matches" {
				// grep 命令没有匹配项时返回 0
				assert.Equal(t, tt.exitCode, result.ExitCode)
				assert.Empty(t, result.Output)
			} else {
				assert.Equal(t, tt.exitCode, result.ExitCode)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

func TestLocalExecutor_Execute(t *testing.T) {
	// 创建执行器
	exec := NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{})

	// 测试执行命令
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Command: types.Command{
			Command: "echo",
			Args:    []string{"hello"},
		},
		Options: &types.ExecuteOptions{},
	}

	result, err := exec.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)

	// 测试列出命令
	commands := exec.ListCommands()
	assert.NotNil(t, commands)
}

func TestLocalExecutor_RegisterCommand(t *testing.T) {
	// 创建执行器
	exec := NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{})

	// 测试注册命令
	command := &testCommand{
		info: types.CommandInfo{
			Name:        "test_command",
			Description: "Test command",
			Usage:       "test_command [args...]",
		},
		fn: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
			return &types.ExecuteResult{
				CommandName: ctx.Command.Command,
				Output:      "test",
				ExitCode:    0,
			}, nil
		},
	}

	err := exec.RegisterCommand(command)
	assert.NoError(t, err)

	// 测试列出命令
	commands := exec.ListCommands()
	assert.NotNil(t, commands)
}

func TestLocalExecutor_UnregisterCommand(t *testing.T) {
	// 创建执行器
	exec := NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{})

	// 先注册一个命令
	command := &testCommand{
		info: types.CommandInfo{
			Name:        "test_command",
			Description: "Test command",
			Usage:       "test_command [args...]",
		},
		fn: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
			return &types.ExecuteResult{
				CommandName: ctx.Command.Command,
				Output:      "test",
				ExitCode:    0,
			}, nil
		},
	}
	err := exec.RegisterCommand(command)
	assert.NoError(t, err)

	// 测试注销命令
	err = exec.UnregisterCommand("test_command")
	assert.NoError(t, err)

	// 测试列出命令
	commands := exec.ListCommands()
	assert.NotNil(t, commands)
}

func TestLocalExecutor_ListCommands(t *testing.T) {
	// 创建执行器
	exec := NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{})

	// 测试列出命令
	commands := exec.ListCommands()
	assert.NotNil(t, commands)
}

// TestLocalExecutorPipeline 测试本地执行器的管道功能
func TestLocalExecutorPipeline(t *testing.T) {
	tests := []struct {
		name                      string
		commands                  []*types.Command
		allowUnregisteredCommands bool
		registerCommands          []types.ICommand
		input                     string
		expectedOutput            string
		expectedError             string
	}{
		{
			name: "basic pipeline - ls and grep",
			commands: []*types.Command{
				{Command: "ls", Args: []string{"-l"}},
				{Command: "grep", Args: []string{"test"}},
			},
			allowUnregisteredCommands: true,
			expectedError:             "",
		},
		{
			name: "pipeline with built-in command",
			commands: []*types.Command{
				{Command: "echo-test", Args: []string{"hello"}},
				{Command: "grep", Args: []string{"hello"}},
			},
			registerCommands: []types.ICommand{
				&testCommand{
					info: types.CommandInfo{
						Name:        "echo-test",
						Description: "Test command 1",
						Usage:       "echo-test [args...]",
					},
					fn: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
						return &types.ExecuteResult{
							CommandName: ctx.Command.Command,
							Output:      "hello world\ntest line\n",
							ExitCode:    0,
						}, nil
					},
				},
			},
			allowUnregisteredCommands: true,
			expectedOutput:            "hello world",
			expectedError:             "",
		},
		{
			name: "pipeline with unregistered commands not allowed",
			commands: []*types.Command{
				{Command: "ls", Args: []string{"-l"}},
				{Command: "grep", Args: []string{"test"}},
			},
			allowUnregisteredCommands: false,
			expectedError:             "unregistered command not allowed",
		},
		{
			name: "pipeline with invalid command",
			commands: []*types.Command{
				{Command: "ls", Args: []string{"-l"}},
				{Command: "invalid-command", Args: []string{}},
			},
			allowUnregisteredCommands: true,
			expectedError:             "command not found",
		},
		{
			name: "pipeline with multiple built-in commands",
			commands: []*types.Command{
				{Command: "cmd1", Args: []string{}},
				{Command: "cmd2", Args: []string{}},
				{Command: "grep", Args: []string{"result"}},
			},
			registerCommands: []types.ICommand{
				&testCommand{
					info: types.CommandInfo{
						Name:        "cmd1",
						Description: "Test command 1",
						Usage:       "cmd1 [args...]",
					},
					fn: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
						return &types.ExecuteResult{
							CommandName: ctx.Command.Command,
							Output:      "first result\nsecond line\n",
							ExitCode:    0,
						}, nil
					},
				},
				&testCommand{
					info: types.CommandInfo{
						Name:        "cmd2",
						Description: "Test command 2",
						Usage:       "cmd2 [args...]",
					},
					fn: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
						return &types.ExecuteResult{
							CommandName: ctx.Command.Command,
							Output:      "processed result\nother line\n",
							ExitCode:    0,
						}, nil
					},
				},
			},
			allowUnregisteredCommands: true,
			expectedOutput:            "processed result",
			expectedError:             "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建执行器
			executor := NewLocalExecutor(types.LocalConfig{
				AllowUnregisteredCommands: tt.allowUnregisteredCommands,
			}, nil)

			// 注册命令
			if tt.registerCommands != nil {
				for _, cmd := range tt.registerCommands {
					err := executor.RegisterCommand(cmd)
					assert.NoError(t, err)
				}
			}

			// 准备上下文
			var stdout, stderr bytes.Buffer
			ctx := &types.ExecuteContext{
				Context: context.Background(),
				IsPiped: true,
				PipeContext: &types.PipelineContext{
					Commands: tt.commands,
				},
				Options: &types.ExecuteOptions{
					Stdout: &stdout,
					Stderr: &stderr,
				},
			}

			// 执行管道
			result, err := executor.Execute(ctx)

			// 验证结果
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.expectedOutput != "" {
					assert.Contains(t, stdout.String(), tt.expectedOutput)
				}
			}
		})
	}
}

// testCommandHandler 用于测试的命令处理器
type testCommandHandler struct {
	output string
	error  error
}

func (h *testCommandHandler) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if h.error != nil {
		return nil, h.error
	}

	if ctx.Options != nil && ctx.Options.Stdout != nil {
		fmt.Fprint(ctx.Options.Stdout, h.output)
	}

	return &types.ExecuteResult{
		CommandName: ctx.Command.Command,
		Output:      h.output,
		ExitCode:    0,
	}, nil
}

// TestLocalExecutorPipelineEdgeCases 测试边缘情况
func TestLocalExecutorPipelineEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() *types.ExecuteContext
		expectedError string
	}{
		{
			name: "nil pipe context",
			setupContext: func() *types.ExecuteContext {
				return &types.ExecuteContext{
					Context: context.Background(),
					IsPiped: true,
				}
			},
			expectedError: "no commands in pipeline",
		},
		{
			name: "empty commands list",
			setupContext: func() *types.ExecuteContext {
				return &types.ExecuteContext{
					Context: context.Background(),
					IsPiped: true,
					PipeContext: &types.PipelineContext{
						Commands: []*types.Command{},
					},
				}
			},
			expectedError: "no commands in pipeline",
		},
		{
			name: "command with empty name",
			setupContext: func() *types.ExecuteContext {
				return &types.ExecuteContext{
					Context: context.Background(),
					IsPiped: true,
					PipeContext: &types.PipelineContext{
						Commands: []*types.Command{
							{Command: "", Args: []string{}},
						},
					},
				}
			},
			expectedError: "command not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewLocalExecutor(types.LocalConfig{
				AllowUnregisteredCommands: true,
			}, nil)

			ctx := tt.setupContext()
			_, err := executor.Execute(ctx)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// TestLocalExecutorPipelineCancel 测试管道执行的取消功能
func TestLocalExecutorPipelineCancel(t *testing.T) {
	executor := NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, nil)

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 准备一个长时间运行的管道
	execCtx := &types.ExecuteContext{
		Context: ctx,
		IsPiped: true,
		PipeContext: &types.PipelineContext{
			Commands: []*types.Command{
				{Command: "sleep", Args: []string{"1"}},
				{Command: "echo", Args: []string{"done"}},
			},
		},
	}

	// 在另一个goroutine中取消上下文
	go func() {
		cancel()
	}()

	// 执行管道
	_, err := executor.Execute(execCtx)

	// 验证是否因为取消而中断
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

type testCommand struct {
	info types.CommandInfo
	fn   func(ctx *types.ExecuteContext) (*types.ExecuteResult, error)
}

func (c *testCommand) Info() types.CommandInfo {
	return c.info
}

func (c *testCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	return c.fn(ctx)
}
