package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iamlongalong/runshell/pkg/types"
)

// TestLSCommand 测试 ls 命令的功能。
// 测试场景：
// 1. 创建临时目录和测试文件
// 2. 执行 ls 命令列出目录内容
// 3. 验证输出中包含所有测试文件
func TestLSCommand(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "ls-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFiles := []string{"file1.txt", "file2.txt"}
	for _, file := range testFiles {
		if err := os.WriteFile(filepath.Join(tempDir, file), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// 执行命令
	cmd := &LSCommand{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	result, err := cmd.Execute(&types.ExecuteContext{
		Args: []string{tempDir},
		Options: &types.ExecuteOptions{
			WorkDir: ".",
			Stdout:  stdout,
			Stderr:  stderr,
		},
		StartTime: time.Now(),
	})

	// 验证执行结果
	if err != nil {
		t.Errorf("LSCommand.Execute() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("LSCommand.Execute() exit code = %v, want 0", result.ExitCode)
	}

	// 验证输出内容
	output := stdout.String()
	for _, file := range testFiles {
		if !strings.Contains(output, file) {
			t.Errorf("LSCommand.Execute() output does not contain %q", file)
		}
	}
}

// TestCatCommand 测试 cat 命令的功能。
// 测试场景：
// 1. 创建临时文件并写入测试内容
// 2. 执行 cat 命令读取文件内容
// 3. 验证输出内容是否正确
func TestCatCommand(t *testing.T) {
	// 创建临时测试文件
	tempFile, err := os.CreateTemp("", "cat-test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	// 写入测试内容
	content := "test content"
	if _, err := tempFile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tempFile.Close()

	// 执行命令
	cmd := &CatCommand{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	result, err := cmd.Execute(&types.ExecuteContext{
		Args: []string{tempFile.Name()},
		Options: &types.ExecuteOptions{
			WorkDir: ".",
			Stdout:  stdout,
			Stderr:  stderr,
		},
		StartTime: time.Now(),
	})

	// 验证执行结果
	if err != nil {
		t.Errorf("CatCommand.Execute() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("CatCommand.Execute() exit code = %v, want 0", result.ExitCode)
	}
	if got := stdout.String(); got != content {
		t.Errorf("CatCommand.Execute() output = %q, want %q", got, content)
	}
}

// TestMkdirCommand 测试 mkdir 命令的功能。
// 测试场景：
// 1. 在临时目录中创建新目录
// 2. 验证目录是否成功创建
// 3. 检查目录权限和存在性
func TestMkdirCommand(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "mkdir-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// 设置新目录路径
	newDir := filepath.Join(tempDir, "newdir")
	cmd := &MkdirCommand{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	// 执行命令
	result, err := cmd.Execute(&types.ExecuteContext{
		Args: []string{newDir},
		Options: &types.ExecuteOptions{
			WorkDir: ".",
			Stdout:  stdout,
			Stderr:  stderr,
		},
		StartTime: time.Now(),
	})

	// 验证执行结果
	if err != nil {
		t.Errorf("MkdirCommand.Execute() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("MkdirCommand.Execute() exit code = %v, want 0", result.ExitCode)
	}

	// 验证目录是否创建成功
	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		t.Error("MkdirCommand.Execute() directory was not created")
	}
}

// TestRmCommand 测试 rm 命令的功能。
// 测试场景：
// 1. 创建测试文件
// 2. 执行 rm 命令删除文件
// 3. 验证文件是否成功删除
func TestRmCommand(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "rm-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// 执行命令
	cmd := &RmCommand{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	result, err := cmd.Execute(&types.ExecuteContext{
		Args: []string{testFile},
		Options: &types.ExecuteOptions{
			WorkDir: ".",
			Stdout:  stdout,
			Stderr:  stderr,
		},
		StartTime: time.Now(),
	})

	// 验证执行结果
	if err != nil {
		t.Errorf("RmCommand.Execute() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("RmCommand.Execute() exit code = %v, want 0", result.ExitCode)
	}

	// 验证文件是否已删除
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("RmCommand.Execute() file was not removed")
	}
}

// TestCpCommand 测试 cp 命令的功能。
// 测试场景：
// 1. 创建源文件并写入测试内容
// 2. 执行 cp 命令复制文件
// 3. 验证目标文件内容是否与源文件一致
func TestCpCommand(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "cp-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// 创建并写入源文件
	srcFile := filepath.Join(tempDir, "src.txt")
	content := "test content"
	if err := os.WriteFile(srcFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// 设置目标文件路径
	dstFile := filepath.Join(tempDir, "dst.txt")

	// 执行命令
	cmd := &CpCommand{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	result, err := cmd.Execute(&types.ExecuteContext{
		Args: []string{srcFile, dstFile},
		Options: &types.ExecuteOptions{
			WorkDir: ".",
			Stdout:  stdout,
			Stderr:  stderr,
		},
		StartTime: time.Now(),
	})

	// 验证执行结果
	if err != nil {
		t.Errorf("CpCommand.Execute() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("CpCommand.Execute() exit code = %v, want 0", result.ExitCode)
	}

	// 验证文件内容
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(dstContent) != content {
		t.Errorf("CpCommand.Execute() copied content = %q, want %q", string(dstContent), content)
	}
}

// TestPWDCommand 测试 pwd 命令的功能。
// 测试场景：
// 1. 设置工作目录
// 2. 执行 pwd 命令
// 3. 验证输出的工作目录路径是否正确
func TestPWDCommand(t *testing.T) {
	// 执行命令
	cmd := &PWDCommand{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	result, err := cmd.Execute(&types.ExecuteContext{
		Options: &types.ExecuteOptions{
			WorkDir: "/test/dir",
			Stdout:  stdout,
			Stderr:  stderr,
		},
		StartTime: time.Now(),
	})

	// 验证执行结果
	if err != nil {
		t.Errorf("PWDCommand.Execute() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("PWDCommand.Execute() exit code = %v, want 0", result.ExitCode)
	}

	// 验证输出路径
	got := strings.TrimSpace(stdout.String())
	if got != "/test/dir" {
		t.Errorf("PWDCommand.Execute() output = %q, want %q", got, "/test/dir")
	}
}
