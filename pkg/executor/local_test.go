package executor

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/iamlongalong/runshell/pkg/types"
)

// TestLocalExecutor 测试本地执行器的基本功能。
// 测试场景包括：
// 1. 执行系统命令（echo）
// 2. 文件操作命令（cat）
// 3. 错误处理（不存在的命令）
func TestLocalExecutor(t *testing.T) {
	executor := NewLocalExecutor()

	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "local-test-*")
	if err != nil {
		t.Fatal(err)
	}
	// 测试完成后清理临时目录
	defer os.RemoveAll(tempDir)

	// 创建测试文件并写入内容
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	// 定义测试用例
	tests := []struct {
		name     string                                           // 测试用例名称
		cmd      string                                           // 要执行的命令
		args     []string                                         // 命令参数
		wantErr  bool                                             // 是否期望出错
		wantCode int                                              // 期望的退出码
		setup    func() *types.ExecuteOptions                     // 设置执行选项
		verify   func(t *testing.T, stdout, stderr *bytes.Buffer) // 验证执行结果
	}{
		{
			name: "echo command", // 测试 echo 命令
			cmd:  "echo",
			args: []string{"hello", "world"},
			setup: func() *types.ExecuteOptions {
				stdout := &bytes.Buffer{}
				return &types.ExecuteOptions{
					WorkDir: tempDir,
					Stdout:  stdout,
					Stderr:  &bytes.Buffer{},
				}
			},
			verify: func(t *testing.T, stdout, stderr *bytes.Buffer) {
				if got := stdout.String(); got != "hello world\n" {
					t.Errorf("echo command output = %q, want %q", got, "hello world\n")
				}
			},
		},
		{
			name: "cat command", // 测试 cat 命令
			cmd:  "cat",
			args: []string{testFile},
			setup: func() *types.ExecuteOptions {
				stdout := &bytes.Buffer{}
				return &types.ExecuteOptions{
					WorkDir: tempDir,
					Stdout:  stdout,
					Stderr:  &bytes.Buffer{},
				}
			},
			verify: func(t *testing.T, stdout, stderr *bytes.Buffer) {
				if got := stdout.String(); got != "test content" {
					t.Errorf("cat command output = %q, want %q", got, "test content")
				}
			},
		},
		{
			name:     "non-existent command", // 测试不存在的命令
			cmd:      "nonexistentcommand",
			wantErr:  true,
			wantCode: -1,
			setup: func() *types.ExecuteOptions {
				return &types.ExecuteOptions{
					WorkDir: tempDir,
					Stdout:  &bytes.Buffer{},
					Stderr:  &bytes.Buffer{},
				}
			},
			verify: func(t *testing.T, stdout, stderr *bytes.Buffer) {},
		},
	}

	// 执行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 准备执行环境
			opts := tt.setup()
			stdout := opts.Stdout.(*bytes.Buffer)
			stderr := opts.Stderr.(*bytes.Buffer)

			// 执行命令
			result, err := executor.Execute(context.Background(), tt.cmd, tt.args, opts)

			// 验证错误情况
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 验证退出码
			if tt.wantErr {
				if tt.wantCode != 0 && result != nil && result.ExitCode != tt.wantCode {
					t.Errorf("Execute() exit code = %v, want %v", result.ExitCode, tt.wantCode)
				}
				return
			}

			// 验证输出结果
			tt.verify(t, stdout, stderr)
		})
	}
}

// TestLocalExecutorWithRegisteredCommands 测试本地执行器的命令注册功能。
// 测试场景包括：
// 1. 注册自定义命令
// 2. 获取命令信息
// 3. 执行注册的命令
// 4. 列出命令
// 5. 注销命令
func TestLocalExecutorWithRegisteredCommands(t *testing.T) {
	executor := NewLocalExecutor()

	// 注册测试命令
	testCmd := &types.Command{
		Name:        "test",
		Description: "Test command",
		Usage:       "test [args...]",
		Category:    "test",
		Handler: &testCommandHandler{
			output: "test output",
		},
	}

	// 测试命令注册
	if err := executor.RegisterCommand(testCmd); err != nil {
		t.Fatal(err)
	}

	// 测试获取命令信息
	cmd, err := executor.GetCommandInfo("test")
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Name != "test" || cmd.Description != "Test command" {
		t.Errorf("GetCommandInfo() got = %v, want name=test, desc=Test command", cmd)
	}

	// 测试执行注册的命令
	stdout := &bytes.Buffer{}
	result, err := executor.Execute(context.Background(), "test", nil, &types.ExecuteOptions{
		Stdout: stdout,
		Stderr: &bytes.Buffer{},
	})

	// 验证执行结果
	if err != nil {
		t.Fatal(err)
	}
	if result.ExitCode != 0 {
		t.Errorf("Execute() exit code = %v, want 0", result.ExitCode)
	}
	if got := stdout.String(); got != "test output" {
		t.Errorf("Execute() output = %q, want %q", got, "test output")
	}

	// 测试列出命令
	commands, err := executor.ListCommands(&types.CommandFilter{Category: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if len(commands) != 1 || commands[0].Name != "test" {
		t.Errorf("ListCommands() got = %v, want 1 command named 'test'", commands)
	}

	// 测试注销命令
	if err := executor.UnregisterCommand("test"); err != nil {
		t.Fatal(err)
	}

	// 验证命令已被注销
	if _, err := executor.GetCommandInfo("test"); err != types.ErrCommandNotFound {
		t.Errorf("GetCommandInfo() after unregister error = %v, want ErrCommandNotFound", err)
	}
}
