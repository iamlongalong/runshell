package commands

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestGitCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "git version",
			args:    []string{"git", "version"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟执行器
			mockExec := &types.MockExecutor{
				ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
					ctx.Options.Stdout.Write([]byte("git version 2.30.1"))
					return &types.ExecuteResult{
						CommandName: ctx.Command.Command,
						ExitCode:    0,
					}, nil
				},
			}
			buf := &bytes.Buffer{}
			cmd := &GitCommand{}
			ctx := &types.ExecuteContext{
				Context: context.Background(),
				Command: types.Command{Command: tt.args[0]},
				Options: &types.ExecuteOptions{
					Stdout: buf,
				},
				Executor: mockExec,
			}

			result, err := cmd.Execute(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
			assert.Contains(t, buf.String(), "git version")
		})
	}
}

func TestGoCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "go version",
			args:    []string{"go", "version"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟执行器
			mockExec := &types.MockExecutor{
				ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
					ctx.Options.Stdout.Write([]byte("go version go1.21.0 darwin/amd64"))
					return &types.ExecuteResult{
						CommandName: ctx.Command.Command,
						ExitCode:    0,
					}, nil
				},
			}
			buf := &bytes.Buffer{}
			cmd := &GoCommand{}
			ctx := &types.ExecuteContext{
				Context: context.Background(),
				Command: types.Command{Command: tt.args[0]},
				Options: &types.ExecuteOptions{
					Stdout: buf,
				},
				Executor: mockExec,
			}

			result, err := cmd.Execute(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
			assert.Contains(t, buf.String(), "go version")
		})
	}
}

func TestHasEnv(t *testing.T) {
	tests := []struct {
		name     string
		env      []string
		key      string
		expected bool
	}{
		{
			name:     "existing env",
			env:      []string{"PATH=/usr/bin", "GOPROXY=https://proxy.golang.org"},
			key:      "GOPROXY",
			expected: true,
		},
		{
			name:     "non-existing env",
			env:      []string{"PATH=/usr/bin"},
			key:      "GOPROXY",
			expected: false,
		},
		{
			name:     "empty env",
			env:      []string{},
			key:      "GOPROXY",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := false
			for _, e := range tt.env {
				if strings.HasPrefix(e, tt.key+"=") {
					result = true
					break
				}
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPythonCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "python version",
			args:    []string{"python", "--version"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟执行器
			mockExec := &types.MockExecutor{
				ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
					output := "Python 3.9.0"
					ctx.Options.Stdout.Write([]byte(output))
					return &types.ExecuteResult{
						CommandName: ctx.Command.Command,
						ExitCode:    0,
						Output:      output,
					}, nil
				},
			}
			buf := &bytes.Buffer{}
			cmd := &PythonCommand{}
			ctx := &types.ExecuteContext{
				Context: context.Background(),
				Command: types.Command{Args: tt.args},
				Options: &types.ExecuteOptions{
					Stdout: buf,
				},
				Executor: mockExec,
			}

			result, err := cmd.Execute(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
			assert.Contains(t, result.Output, "Python")
		})
	}
}
