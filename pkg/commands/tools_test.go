package commands

import (
	"bytes"
	"context"
	"os"
	"os/exec"
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
				Args:     tt.args,
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
				Args:     tt.args,
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
				Args:     tt.args,
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
	// 跳过如果 pip 未安装
	if _, err := exec.LookPath("pip3"); err != nil {
		if _, err := exec.LookPath("pip"); err != nil {
			t.Skip("Pip is not installed")
		}
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "pip-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试用例
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "version check",
			args:    []string{"--version"},
			wantErr: false,
		},
		{
			name:    "invalid argument",
			args:    []string{"--invalid-arg"},
			wantErr: true,
		},
	}

	// 执行测试
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &PipCommand{}
			ctx := &types.ExecuteContext{
				Context: context.Background(),
				Args:    tt.args,
				Options: &types.ExecuteOptions{
					WorkDir: tempDir,
				},
				Executor: executor.NewMockExecutor(),
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

func TestDockerCommand(t *testing.T) {
	// 跳过如果 docker 未安装
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker is not installed")
	}

	// 创建测试用例
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "version check",
			args:    []string{"--version"},
			wantErr: false,
		},
		{
			name:    "invalid argument",
			args:    []string{"--invalid-arg"},
			wantErr: true,
		},
	}

	// 执行测试
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &DockerCommand{}
			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Args:     tt.args,
				Executor: executor.NewMockExecutor(),
				Options:  &types.ExecuteOptions{},
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

func TestNodeCommand(t *testing.T) {
	// 跳过如果 node 未安装
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("Node.js is not installed")
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "node-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试用例
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "version check",
			args:    []string{"--version"},
			wantErr: false,
		},
		{
			name:    "invalid argument",
			args:    []string{"--invalid-arg"},
			wantErr: true,
		},
	}

	// 执行测试
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &NodeCommand{}
			ex := &executor.LocalExecutor{}
			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Args:     tt.args,
				Executor: ex,
				Options:  &types.ExecuteOptions{},
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

func TestNPMCommand(t *testing.T) {
	// 跳过如果 npm 未安装
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("NPM is not installed")
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "npm-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试用例
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "version check",
			args:    []string{"--version"},
			wantErr: false,
		},
		{
			name:    "invalid argument",
			args:    []string{"--invalid-arg"},
			wantErr: true,
		},
	}

	// 执行测试
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &NPMCommand{}
			ex := &executor.LocalExecutor{}
			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Args:     tt.args,
				Executor: ex,
				Options:  &types.ExecuteOptions{},
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
