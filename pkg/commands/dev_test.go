package commands

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestGitCommand(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 初始化 Git 仓库
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Failed to initialize git repository: %v", err)
	}

	// 创建测试用例
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "git version",
			args:    []string{"--version"},
			wantErr: false,
		},
		{
			name:    "git status in repo",
			args:    []string{"status"},
			wantErr: false,
		},
		{
			name:    "invalid git command",
			args:    []string{"invalid-command"},
			wantErr: true,
		},
		{
			name:    "git in non-repo",
			args:    []string{"status"},
			wantErr: true,
		},
	}

	// 执行测试
	mockExec := &executor.LocalExecutor{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &GitCommand{}
			workDir := tempDir
			if tt.name == "git in non-repo" {
				workDir = os.TempDir()
			}
			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Args:     tt.args,
				Executor: mockExec,
				Options: &types.ExecuteOptions{
					WorkDir: workDir,
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

func TestGoCommand(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "go-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建 go.mod 文件
	goModContent := []byte("module test\n\ngo 1.20\n")
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), goModContent, 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// 创建测试用例
	tests := []struct {
		name    string
		args    []string
		workDir string
		env     map[string]string
		wantErr bool
	}{
		{
			name:    "go version",
			args:    []string{"version"},
			workDir: "",
			wantErr: false,
		},
		{
			name:    "go mod verify",
			args:    []string{"mod", "verify"},
			workDir: tempDir,
			wantErr: false,
		},
		{
			name:    "invalid go command",
			args:    []string{"invalid-command"},
			workDir: tempDir,
			wantErr: true,
		},
		{
			name:    "go in non-module",
			args:    []string{"mod", "verify"},
			workDir: os.TempDir(),
			wantErr: true,
		},
	}

	// 执行测试
	ex := &executor.LocalExecutor{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &GoCommand{}
			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Args:     tt.args,
				Executor: ex,
				Options: &types.ExecuteOptions{
					WorkDir: tt.workDir,
					Env:     tt.env,
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
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "python-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test Python script
	scriptContent := []byte("print('Hello, World!')")
	scriptPath := filepath.Join(tempDir, "script.py")
	if err := os.WriteFile(scriptPath, scriptContent, 0644); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "python version",
			args:    []string{"--version"},
			wantErr: false,
		},
		{
			name:    "run script",
			args:    []string{scriptPath},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &PythonCommand{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			ex := &executor.LocalExecutor{}
			ctx := &types.ExecuteContext{
				Context:  context.Background(),
				Args:     tt.args,
				Executor: ex,
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
				if tt.args[0] == "--version" {
					assert.Contains(t, stdout.String(), "Python")
				}
			}
		})
	}
}
