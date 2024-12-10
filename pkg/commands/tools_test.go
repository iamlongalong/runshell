package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestWgetCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "download file",
			args:    []string{"https://example.com/file.txt"},
			wantErr: false,
		},
		{
			name:    "no url",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "invalid url",
			args:    []string{"not-a-url"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &WgetCommand{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			mockExec := executor.NewMockExecutor()
			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "wget", Args: tt.args},
				Executor: mockExec,
				Options: &types.ExecuteOptions{
					Stdout: stdout,
					Stderr: stderr,
				},
			}

			result, err := cmd.Execute(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 0, result.ExitCode)
			}
		})
	}
}

func TestTarCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		files   map[string]string
		wantErr bool
	}{
		{
			name:    "create archive",
			args:    []string{"-czf", "archive.tar.gz", "file1.txt", "file2.txt"},
			files:   map[string]string{"file1.txt": "content1", "file2.txt": "content2"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			files:   map[string]string{},
			wantErr: true,
		},
		{
			name:    "invalid option",
			args:    []string{"-invalid"},
			files:   map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &TarCommand{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			mockExec := executor.NewMockExecutor()
			for path, content := range tt.files {
				mockExec.WriteFile(path, content)
			}

			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "tar", Args: tt.args},
				Executor: mockExec,
				Options: &types.ExecuteOptions{
					Stdout: stdout,
					Stderr: stderr,
				},
			}

			result, err := cmd.Execute(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 0, result.ExitCode)
			}
		})
	}
}

func TestZipCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		files   map[string]string
		wantErr bool
	}{
		{
			name:    "create archive",
			args:    []string{"archive.zip", "file1.txt", "file2.txt"},
			files:   map[string]string{"file1.txt": "content1", "file2.txt": "content2"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			files:   map[string]string{},
			wantErr: true,
		},
		{
			name:    "missing files",
			args:    []string{"archive.zip"},
			files:   map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &ZipCommand{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			mockExec := executor.NewMockExecutor()
			for path, content := range tt.files {
				mockExec.WriteFile(path, content)
			}

			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "zip", Args: tt.args},
				Executor: mockExec,
				Options: &types.ExecuteOptions{
					Stdout: stdout,
					Stderr: stderr,
				},
			}

			result, err := cmd.Execute(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 0, result.ExitCode)
			}
		})
	}
}

func TestPipCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "version check",
			args:    []string{"pip", "--version"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟执行器
			mockExec := &executor.MockExecutor{
				ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
					ctx.Options.Stdout.Write([]byte("pip 21.0.1"))
					return &types.ExecuteResult{
						CommandName: ctx.Command.Command,
						ExitCode:    0,
					}, nil
				},
			}

			buf := &bytes.Buffer{}

			cmd := &PipCommand{}
			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "pip", Args: tt.args},
				Executor: mockExec,
				Options: &types.ExecuteOptions{
					Stdout: buf,
				},
			}

			result, err := cmd.Execute(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
			assert.Contains(t, buf.String(), "pip")
		})
	}
}

func TestDockerCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "version check",
			args:    []string{"docker", "--version"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟执行器
			mockExec := &executor.MockExecutor{
				ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
					ctx.Options.Stdout.Write([]byte("Docker version 20.10.8"))
					return &types.ExecuteResult{
						CommandName: ctx.Command.Command,
						ExitCode:    0,
					}, nil
				},
			}

			buf := &bytes.Buffer{}

			cmd := &DockerCommand{}
			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "docker", Args: tt.args},
				Executor: mockExec,
				Options: &types.ExecuteOptions{
					Stdout: buf,
				},
			}

			result, err := cmd.Execute(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
			assert.Contains(t, buf.String(), "Docker version")
		})
	}
}

func TestNodeCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "version check",
			args:    []string{"node", "--version"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟执行器
			mockExec := &executor.MockExecutor{
				ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
					ctx.Options.Stdout.Write([]byte("v16.0.0"))
					return &types.ExecuteResult{
						CommandName: ctx.Command.Command,
						ExitCode:    0,
					}, nil
				},
			}

			buf := &bytes.Buffer{}

			cmd := &NodeCommand{}
			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "node", Args: tt.args},
				Executor: mockExec,
				Options: &types.ExecuteOptions{
					Stdout: buf,
				},
			}

			result, err := cmd.Execute(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
			assert.Contains(t, buf.String(), "v16")
		})
	}
}

func TestNPMCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "version check",
			args:    []string{"npm", "--version"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟执行器
			mockExec := &executor.MockExecutor{
				ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
					ctx.Options.Stdout.Write([]byte("8.0.0"))
					return &types.ExecuteResult{
						CommandName: ctx.Command.Command,
						ExitCode:    0,
					}, nil
				},
			}

			buf := &bytes.Buffer{}

			cmd := &NPMCommand{}
			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "npm", Args: tt.args},
				Executor: mockExec,
				Options: &types.ExecuteOptions{
					Stdout: buf,
				},
			}

			result, err := cmd.Execute(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
			assert.Contains(t, buf.String(), "8.0.0")
		})
	}
}
