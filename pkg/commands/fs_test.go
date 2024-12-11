package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestLSCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		files   map[string]string
		wantErr bool
	}{
		{
			name:    "ls current dir",
			args:    []string{},
			files:   map[string]string{"file1.txt": "content1", "file2.txt": "content2"},
			wantErr: false,
		},
		{
			name:    "ls with path",
			args:    []string{"/test"},
			files:   map[string]string{"/test/file1.txt": "content1", "/test/file2.txt": "content2"},
			wantErr: false,
		},
		{
			name:    "ls non-existent dir",
			args:    []string{"/nonexistent"},
			files:   map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &LSCommand{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			mockExec := executor.NewMockExecutor()
			for path, content := range tt.files {
				mockExec.WriteFile(path, content)
			}

			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "ls", Args: tt.args},
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
				assert.NotEmpty(t, stdout.String())
			}
		})
	}
}

func TestCatCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		files    map[string]string
		wantErr  bool
		wantText string
	}{
		{
			name:     "cat existing file",
			args:     []string{"test.txt"},
			files:    map[string]string{"test.txt": "test content"},
			wantErr:  false,
			wantText: "test content",
		},
		{
			name:    "cat non-existent file",
			args:    []string{"nonexistent.txt"},
			files:   map[string]string{},
			wantErr: true,
		},
		{
			name:    "cat without args",
			args:    []string{},
			files:   map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &CatCommand{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			mockExec := executor.NewMockExecutor()
			for path, content := range tt.files {
				mockExec.WriteFile(path, content)
			}

			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "cat", Args: tt.args},
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
				assert.Equal(t, tt.wantText, stdout.String())
			}
		})
	}
}

func TestMkdirCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		files   map[string]string
		wantErr bool
	}{
		{
			name:    "mkdir new directory",
			args:    []string{"newdir"},
			files:   map[string]string{},
			wantErr: false,
		},
		{
			name:    "mkdir without args",
			args:    []string{},
			files:   map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &MkdirCommand{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			mockExec := executor.NewMockExecutor()
			for path, content := range tt.files {
				mockExec.WriteFile(path, content)
			}

			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "mkdir", Args: tt.args},
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
				if len(tt.args) > 0 {
					_, exists := mockExec.ReadFile(tt.args[0])
					assert.True(t, exists, "Directory was not created")
				}
			}
		})
	}
}

func TestRmCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		files   map[string]string
		wantErr bool
	}{
		{
			name:    "rm existing file",
			args:    []string{"test.txt"},
			files:   map[string]string{"test.txt": "content"},
			wantErr: false,
		},
		{
			name:    "rm non-existent file",
			args:    []string{"nonexistent.txt"},
			files:   map[string]string{},
			wantErr: true,
		},
		{
			name:    "rm without args",
			args:    []string{},
			files:   map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &RmCommand{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			mockExec := executor.NewMockExecutor()
			for path, content := range tt.files {
				mockExec.WriteFile(path, content)
			}

			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "rm", Args: tt.args},
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
				if len(tt.args) > 0 {
					_, exists := mockExec.ReadFile(tt.args[0])
					assert.False(t, exists, "File was not removed")
				}
			}
		})
	}
}

func TestCpCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		files   map[string]string
		wantErr bool
	}{
		{
			name:    "cp existing file",
			args:    []string{"src.txt", "dst.txt"},
			files:   map[string]string{"src.txt": "test content"},
			wantErr: false,
		},
		{
			name:    "cp non-existent file",
			args:    []string{"nonexistent.txt", "dst.txt"},
			files:   map[string]string{},
			wantErr: true,
		},
		{
			name:    "cp without args",
			args:    []string{},
			files:   map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &CpCommand{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			mockExec := executor.NewMockExecutor()
			for path, content := range tt.files {
				mockExec.WriteFile(path, content)
			}

			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "cp", Args: tt.args},
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
				if len(tt.args) >= 2 {
					srcContent, _ := mockExec.ReadFile(tt.args[0])
					dstContent, exists := mockExec.ReadFile(tt.args[1])
					assert.True(t, exists, "Destination file was not created")
					assert.Equal(t, srcContent, dstContent, "File content mismatch")
				}
			}
		})
	}
}

func TestPWDCommand(t *testing.T) {
	tests := []struct {
		name    string
		workDir string
		wantErr bool
	}{
		{
			name:    "pwd in existing dir",
			workDir: "/test/dir",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &PWDCommand{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			mockExec := executor.NewMockExecutor()
			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Command: "pwd"},
				Executor: mockExec,
				Options: &types.ExecuteOptions{
					Stdout:  stdout,
					Stderr:  stderr,
					WorkDir: tt.workDir,
				},
			}

			result, err := cmd.Execute(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 0, result.ExitCode)
				assert.Equal(t, tt.workDir+"\n", stdout.String())
			}
		})
	}
}

func TestReadFileCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		files   map[string]string
		wantErr bool
	}{
		{
			name:    "no args",
			args:    []string{},
			files:   map[string]string{},
			wantErr: true,
		},
		{
			name:    "invalid start line",
			args:    []string{"test.txt", "-1", "1"},
			files:   map[string]string{"test.txt": "line1\nline2\nline3"},
			wantErr: true,
		},
		{
			name:    "invalid end line",
			args:    []string{"test.txt", "1", "-1"},
			files:   map[string]string{"test.txt": "line1\nline2\nline3"},
			wantErr: true,
		},
		{
			name:    "start line > end line",
			args:    []string{"test.txt", "2", "1"},
			files:   map[string]string{"test.txt": "line1\nline2\nline3"},
			wantErr: true,
		},
		{
			name:    "file not found",
			args:    []string{"nonexistent.txt", "1", "1"},
			files:   map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &ReadFileCommand{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			mockExec := executor.NewMockExecutor()
			for path, content := range tt.files {
				mockExec.WriteFile(path, content)
			}

			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Command:  types.Command{Args: tt.args},
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
