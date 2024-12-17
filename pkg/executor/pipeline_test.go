package executor

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestPipelineExecutor(t *testing.T) {
	// 创建模拟执行器
	mockExec := &types.MockExecutor{
		ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
			if !ctx.IsPiped {
				return nil, fmt.Errorf("expected piped execution")
			}

			// 构建完整的命令字符串
			var cmds []string
			for _, cmd := range ctx.PipeContext.Commands {
				cmds = append(cmds, fmt.Sprintf("%s %s", cmd.Command, strings.Join(cmd.Args, " ")))
			}
			cmdStr := strings.Join(cmds, " | ")

			// 模拟执行 "echo hello | grep hello"
			if cmdStr == "echo hello | grep hello" {
				if ctx.Options != nil && ctx.Options.Stdout != nil {
					ctx.Options.Stdout.Write([]byte("hello\n"))
				}
				return &types.ExecuteResult{
					CommandName: cmdStr,
					ExitCode:    0,
					Output:      "hello\n",
				}, nil
			}

			return &types.ExecuteResult{
				CommandName: cmdStr,
				ExitCode:    0,
			}, nil
		},
	}

	// 创建管道执行器
	pipeExec := NewPipelineExecutor(mockExec)

	// 测试管道命令
	pipeline, err := pipeExec.ParsePipeline("echo hello | grep hello")
	assert.NoError(t, err)

	var output bytes.Buffer
	pipeline.Context = context.Background()
	pipeline.Options = &types.ExecuteOptions{
		Stdout: &output,
	}

	ctx := &types.ExecuteContext{
		Context:     context.Background(),
		PipeContext: pipeline,
		IsPiped:     true,
		Options:     pipeline.Options,
	}

	result, err := pipeExec.executePipeline(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "hello\n", output.String())
}

func TestParsePipeline(t *testing.T) {
	exec := NewPipelineExecutor(NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{}, nil))

	tests := []struct {
		name      string
		cmdStr    string
		wantErr   bool
		wantCount int
	}{
		{
			name:      "simple pipe",
			cmdStr:    "echo hello | grep hello",
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:      "multiple pipes",
			cmdStr:    "echo hello | grep hello | wc -l",
			wantErr:   false,
			wantCount: 3,
		},
		{
			name:      "empty command",
			cmdStr:    "",
			wantErr:   true,
			wantCount: 0,
		},
		{
			name:      "invalid pipe",
			cmdStr:    "|",
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline, err := exec.ParsePipeline(tt.cmdStr)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(pipeline.Commands))
		})
	}
}

func TestPipelineExecutor_ExecutePipeline(t *testing.T) {
	tests := []struct {
		name           string
		commands       string
		expectedError  bool
		expectedOutput string
		expectedCode   int
		env            map[string]string
	}{
		{
			name:           "simple pipe",
			commands:       "echo hello | grep hello",
			expectedOutput: "hello\n",
			expectedCode:   0,
		},
		{
			name:           "pipe with no output",
			commands:       "echo hello | grep world",
			expectedOutput: "",
			expectedCode:   1,    // grep 命令没有找到匹配时返回 1
			expectedError:  true, // 这是预期的错误
		},
		{
			name:           "pipe with environment variables",
			commands:       "echo $TEST_VAR | grep test",
			expectedOutput: "",
			expectedCode:   1,    // grep 命令没有找到匹配时返回 1
			expectedError:  true, // 这是预期的错误
			env: map[string]string{
				"TEST_VAR": "test value",
			},
		},
		{
			name:          "invalid pipe syntax",
			commands:      "echo hello |",
			expectedError: true,
		},
		{
			name:          "empty command in pipe",
			commands:      "echo hello | | grep world",
			expectedError: true,
		},
		{
			name:          "pipe starts with |",
			commands:      "| echo hello",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &types.MockExecutor{
				ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
					if tt.expectedError {
						return &types.ExecuteResult{
							ExitCode: tt.expectedCode,
							Output:   tt.expectedOutput,
						}, fmt.Errorf("exit status %d", tt.expectedCode)
					}
					return &types.ExecuteResult{
						ExitCode: tt.expectedCode,
						Output:   tt.expectedOutput,
					}, nil
				},
			}

			executor := NewPipelineExecutor(mockExecutor)
			ctx := &types.ExecuteContext{
				Command: types.Command{
					Command: tt.commands,
				},
				Options: &types.ExecuteOptions{
					Env: tt.env,
				},
			}

			result, err := executor.Execute(ctx)

			if tt.expectedError {
				assert.Error(t, err)
				if result != nil {
					assert.Equal(t, tt.expectedCode, result.ExitCode)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedCode, result.ExitCode)
				assert.Equal(t, tt.expectedOutput, result.Output)
			}
		})
	}
}

func TestPipelineExecutor_ParsePipeline(t *testing.T) {
	exec := NewPipelineExecutor(NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{}, nil))

	tests := []struct {
		name      string
		cmdStr    string
		wantErr   bool
		wantCount int
		wantCmds  []string
	}{
		{
			name:      "simple pipe",
			cmdStr:    "echo hello | grep hello",
			wantErr:   false,
			wantCount: 2,
			wantCmds:  []string{"echo", "grep"},
		},
		{
			name:      "multiple pipes",
			cmdStr:    "echo hello | grep hello | wc -l",
			wantErr:   false,
			wantCount: 3,
			wantCmds:  []string{"echo", "grep", "wc"},
		},
		{
			name:      "empty command",
			cmdStr:    "",
			wantErr:   true,
			wantCount: 0,
		},
		{
			name:      "invalid pipe",
			cmdStr:    "|",
			wantErr:   true,
			wantCount: 0,
		},
		{
			name:      "pipe with spaces",
			cmdStr:    "  echo   hello   |   grep   hello  ",
			wantErr:   false,
			wantCount: 2,
			wantCmds:  []string{"echo", "grep"},
		},
		{
			name:      "pipe with empty segments",
			cmdStr:    "echo hello | | grep hello",
			wantErr:   true,
			wantCount: 0,
		},
		{
			name:      "pipe ending with pipe",
			cmdStr:    "echo hello | grep hello |",
			wantErr:   true,
			wantCount: 0,
		},
		{
			name:      "pipe starting with pipe",
			cmdStr:    "| echo hello",
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline, err := exec.ParsePipeline(tt.cmdStr)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(pipeline.Commands))

			if tt.wantCmds != nil {
				for i, cmd := range tt.wantCmds {
					assert.Equal(t, cmd, pipeline.Commands[i].Command)
				}
			}
		})
	}
}
