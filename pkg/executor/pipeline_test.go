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
	mockExec := &MockExecutor{
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
	}, &types.ExecuteOptions{}))

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
	// 创建基础执行器
	exec := NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{})

	// 创建管道执行器
	pipeExec := NewPipelineExecutor(exec)

	tests := []struct {
		name    string
		cmdStr  string
		wantErr bool
		want    string
	}{
		{
			name:    "simple pipe",
			cmdStr:  "echo hello | grep hello",
			wantErr: false,
			want:    "hello\n",
		},
		{
			name:    "multiple pipes",
			cmdStr:  "echo hello world | grep hello | wc -l",
			wantErr: false,
			want:    "1\n",
		},
		{
			name:    "empty pipeline",
			cmdStr:  "",
			wantErr: true,
		},
		{
			name:    "invalid command",
			cmdStr:  "invalidcmd123 | grep test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 解析管道命令
			pipeline, err := pipeExec.ParsePipeline(tt.cmdStr)
			if tt.wantErr && err != nil {
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			// 准备输出缓冲区
			var output bytes.Buffer
			pipeline.Options = &types.ExecuteOptions{
				Stdout: &output,
				Stderr: &output,
			}
			pipeline.Context = context.Background()

			// 执行管道命令
			ctx := &types.ExecuteContext{
				Context:     context.Background(),
				PipeContext: pipeline,
				IsPiped:     true,
				Options:     pipeline.Options,
			}

			result, err := pipeExec.executePipeline(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
			if tt.want != "" {
				assert.Equal(t, strings.TrimSpace(tt.want), strings.TrimSpace(output.String()))
			}
		})
	}
}

func TestPipelineExecutor_ExecutePipelineWithError(t *testing.T) {
	// 创建管道执行器
	exec := NewPipelineExecutor(NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{}))

	tests := []struct {
		name    string
		cmdStr  string
		wantErr bool
		want    string
	}{
		{
			name:    "simple pipe",
			cmdStr:  "echo hello | grep hello",
			wantErr: false,
			want:    "hello\n",
		},
		{
			name:    "multiple pipes",
			cmdStr:  "echo hello world | grep hello | wc -l",
			wantErr: false,
			want:    "1\n",
		},
		{
			name:    "empty pipeline",
			cmdStr:  "",
			wantErr: true,
		},
		{
			name:    "invalid command",
			cmdStr:  "invalidcmd123 | grep test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 解析管道命令
			pipeline, err := exec.ParsePipeline(tt.cmdStr)
			if tt.wantErr && err != nil {
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			// 准备输出缓冲区
			var output bytes.Buffer
			pipeline.Options = &types.ExecuteOptions{
				Stdout: &output,
				Stderr: &output,
			}
			pipeline.Context = context.Background()

			// 执行管道命令
			ctx := &types.ExecuteContext{
				Context:     context.Background(),
				PipeContext: pipeline,
				IsPiped:     true,
				Options:     pipeline.Options,
			}

			result, err := exec.executePipeline(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
			if tt.want != "" {
				assert.Equal(t, strings.TrimSpace(tt.want), strings.TrimSpace(output.String()))
			}
		})
	}
}
