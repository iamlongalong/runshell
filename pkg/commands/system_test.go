package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestPSCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "ps without args",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "ps with args",
			args:    []string{"-ef"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &PSCommand{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			mockExec := &executor.MockExecutor{
				ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
					ctx.Options.Stdout.Write([]byte("  PID  CPU%  MEM%  COMMAND\n  1    0.0   0.1   init\n"))
					return &types.ExecuteResult{
						CommandName: "ps",
						ExitCode:    0,
					}, nil
				},
			}
			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "ps", Args: tt.args},
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
				assert.Contains(t, stdout.String(), "PID")
				assert.Contains(t, stdout.String(), "CPU%")
			}
		})
	}
}

func TestTopCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "top without args",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "top with args",
			args:    []string{"-n", "1"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &TopCommand{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			mockExec := &executor.MockExecutor{
				ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
					ctx.Options.Stdout.Write([]byte("System Overview\nCPU: 0.1%\nMEM: 50%\n"))
					return &types.ExecuteResult{
						CommandName: "top",
						ExitCode:    0,
					}, nil
				},
			}
			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "top", Args: tt.args},
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
				assert.Contains(t, stdout.String(), "System Overview")
			}
		})
	}
}

func TestDFCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "df without args",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "df with args",
			args:    []string{"-h"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &DFCommand{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			mockExec := &executor.MockExecutor{}
			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "df", Args: tt.args},
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
				assert.Contains(t, stdout.String(), "Filesystem")
				assert.Contains(t, stdout.String(), "Size")
			}
		})
	}
}

func TestUNameCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "uname without args",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "uname with -a",
			args:    []string{"-a"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &UNameCommand{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			mockExec := &executor.MockExecutor{
				ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
					ctx.Options.Stdout.Write([]byte("Darwin 22.6.0"))
					return &types.ExecuteResult{
						CommandName: "uname",
						ExitCode:    0,
					}, nil
				},
			}
			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "uname", Args: tt.args},
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

func TestEnvCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "env without args",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "env with pattern",
			args:    []string{"PATH"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &EnvCommand{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			mockExec := &executor.MockExecutor{
				ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
					ctx.Options.Stdout.Write([]byte("PATH=/usr/local/bin:/usr/bin\nGOPATH=/Users/user/go\n"))
					return &types.ExecuteResult{
						CommandName: "env",
						ExitCode:    0,
					}, nil
				},
			}
			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "env", Args: tt.args},
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

func TestKillCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "kill without args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "kill with pid",
			args:    []string{"1234"},
			wantErr: false,
		},
		{
			name:    "kill with pid not found",
			args:    []string{"12345"},
			wantErr: true,
		},
		{
			name:    "kill with signal",
			args:    []string{"-9", "1234"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &KillCommand{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			mockExec := &executor.MockExecutor{}
			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "kill", Args: tt.args},
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
