package executor

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestPipelineExecutor(t *testing.T) {
	// 创建本地执行器
	exec := NewLocalExecutor()

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
			result, err := pipeExec.ExecutePipeline(pipeline)
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

func TestParsePipeline(t *testing.T) {
	exec := NewPipelineExecutor(NewLocalExecutor())

	tests := []struct {
		name      string
		cmdStr    string
		wantCmds  int
		wantFirst string
		wantErr   bool
	}{
		{
			name:      "simple pipe",
			cmdStr:    "ls -l | grep test",
			wantCmds:  2,
			wantFirst: "ls",
			wantErr:   false,
		},
		{
			name:    "empty command",
			cmdStr:  "  |  ",
			wantErr: true,
		},
		{
			name:      "multiple pipes",
			cmdStr:    "echo hello | grep l | wc -l",
			wantCmds:  3,
			wantFirst: "echo",
			wantErr:   false,
		},
		{
			name:    "empty string",
			cmdStr:  "",
			wantErr: true,
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
			assert.Equal(t, tt.wantCmds, len(pipeline.Commands))
			if tt.wantFirst != "" {
				assert.Equal(t, tt.wantFirst, pipeline.Commands[0].Command)
			}
		})
	}
}
