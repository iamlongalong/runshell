package commands

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestWgetCommand(t *testing.T) {
	// 创建测试服务器
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test content"))
	}))
	defer ts.Close()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "wget-test")
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
			name:    "download file",
			args:    []string{ts.URL, "test.txt"},
			wantErr: false,
		},
		{
			name:    "missing url",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "invalid url",
			args:    []string{"http://invalid.url"},
			wantErr: true,
		},
	}

	// 执行测试
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &WgetCommand{}
			ctx := &types.ExecuteContext{
				Context: context.Background(),
				Args:    tt.args,
				Options: &types.ExecuteOptions{
					WorkDir: tempDir,
				},
			}

			result, err := cmd.Execute(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 0, result.ExitCode)

				// 验证文件内容
				if len(tt.args) > 1 {
					content, err := os.ReadFile(filepath.Join(tempDir, tt.args[1]))
					assert.NoError(t, err)
					assert.Equal(t, "test content", string(content))
				}
			}
		})
	}
}

func TestTarCommand(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "tar-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 创建测试用例
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "create archive",
			args:    []string{"-czf", "test.tar.gz", "test.txt"},
			wantErr: false,
		},
		{
			name:    "missing arguments",
			args:    []string{"-czf"},
			wantErr: true,
		},
	}

	// 执行测试
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &TarCommand{}
			ctx := &types.ExecuteContext{
				Context: context.Background(),
				Args:    tt.args,
				Options: &types.ExecuteOptions{
					WorkDir: tempDir,
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
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "zip-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 创建测试用例
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "create archive",
			args:    []string{"test.zip", "test.txt"},
			wantErr: false,
		},
		{
			name:    "missing arguments",
			args:    []string{"test.zip"},
			wantErr: true,
		},
	}

	// 执行测试
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &ZipCommand{}
			ctx := &types.ExecuteContext{
				Context: context.Background(),
				Args:    tt.args,
				Options: &types.ExecuteOptions{
					WorkDir: tempDir,
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

func TestPythonCommand(t *testing.T) {
	// 跳过如果 python 未安装
	if _, err := exec.LookPath("python3"); err != nil {
		if _, err := exec.LookPath("python"); err != nil {
			t.Skip("Python is not installed")
		}
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "python-test")
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
			cmd := &PythonCommand{}
			ctx := &types.ExecuteContext{
				Context: context.Background(),
				Args:    tt.args,
				Options: &types.ExecuteOptions{
					WorkDir: tempDir,
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
				Context: context.Background(),
				Args:    tt.args,
				Options: &types.ExecuteOptions{},
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
			ctx := &types.ExecuteContext{
				Context: context.Background(),
				Args:    tt.args,
				Options: &types.ExecuteOptions{
					WorkDir: tempDir,
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
			ctx := &types.ExecuteContext{
				Context: context.Background(),
				Args:    tt.args,
				Options: &types.ExecuteOptions{
					WorkDir: tempDir,
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
