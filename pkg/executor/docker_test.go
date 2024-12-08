package executor

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/iamlongalong/runshell/pkg/types"
)

// TestDockerExecutor 测试 Docker 执行器的基本功能。
// 测试场景包括：
// 1. 执行基本命令（echo）
// 2. 执行文件系统命令（ls）
// 3. 处理不存在的命令
//
// 注意：
// - 需要 Docker 环境
// - 可以通过 SKIP_DOCKER_TESTS 环境变量跳过测试
func TestDockerExecutor(t *testing.T) {
	// 检查是否跳过 Docker 测试
	if os.Getenv("SKIP_DOCKER_TESTS") != "" {
		t.Skip("Skipping Docker tests")
	}

	// 创建 Docker 执行器实例
	executor, err := NewDockerExecutor("ubuntu:latest")
	if err != nil {
		t.Fatal("Failed to create Docker executor:", err)
	}

	// 定义测试用例
	tests := []struct {
		name         string                            // 测试用例名称
		command      string                            // 要执行的命令
		args         []string                          // 命令参数
		want         string                            // 期望的输出
		wantErr      error                             // 期望的错误
		wantExitCode int                               // 期望的退出码
		verifyOutput func(t *testing.T, output string) // 自定义输出验证函数
	}{
		{
			name:         "echo command in container", // 测试 echo 命令
			command:      "echo",
			args:         []string{"hello world"},
			want:         "hello world\n",
			wantExitCode: 0,
		},
		{
			name:         "ls command in container", // 测试 ls 命令
			command:      "ls",
			args:         []string{"/"},
			wantExitCode: 0,
			verifyOutput: func(t *testing.T, output string) {
				// 验证输出包含标准 Linux 目录
				expectedDirs := []string{
					"bin",
					"dev",
					"etc",
					"home",
					"lib",
					"media",
					"mnt",
					"opt",
					"proc",
					"root",
					"run",
					"sbin",
					"srv",
					"sys",
					"tmp",
					"usr",
					"var",
				}
				for _, dir := range expectedDirs {
					if !strings.Contains(output, dir) {
						t.Errorf("Expected output to contain %q", dir)
					}
				}
			},
		},
		{
			name:         "non-existent command in container", // 测试不存在的命令
			command:      "nonexistentcommand",
			args:         []string{},
			wantErr:      types.ErrCommandNotFound,
			wantExitCode: 127,
		},
	}

	// 执行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 执行命令
			result, err := executor.Execute(context.Background(), tt.command, tt.args, nil)

			// 验证错误情况
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Execute() error = %v, want %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("Execute() unexpected error: %v", err)
			}

			// 验证退出码
			if result.ExitCode != tt.wantExitCode {
				t.Errorf("Execute() exit code = %d, want %d", result.ExitCode, tt.wantExitCode)
			}

			// 验证输出
			if tt.verifyOutput != nil {
				tt.verifyOutput(t, result.Output)
			} else if tt.wantErr == nil && result.Output != tt.want {
				t.Errorf("Execute() output = %q, want %q", result.Output, tt.want)
			}
		})
	}
}

// TestDockerExecutorWithRegisteredCommands 测试 Docker 执行器的命令注册功能。
// 测试场景包括：
// 1. 注册自定义命令
// 2. 获取命令信息
// 3. 获取命令帮助
// 4. 列出命令
// 5. 执行注册的命令
// 6. 注销命令
//
// 注意：
// - 需要 Docker 环境
// - 可以通过 SKIP_DOCKER_TESTS 环境变量跳过测试
func TestDockerExecutorWithRegisteredCommands(t *testing.T) {
	// 检查是否跳过 Docker 测试
	if os.Getenv("SKIP_DOCKER_TESTS") != "" {
		t.Skip("Skipping Docker tests")
	}

	// 创建 Docker 执行器实例
	executor, err := NewDockerExecutor("ubuntu:latest")
	if err != nil {
		t.Fatal("Failed to create Docker executor:", err)
	}

	// 创建并注册测试命令
	cmd := &types.Command{
		Name:        "test",
		Description: "Test command",
		Usage:       "test [args...]",
		Category:    "test",
		Handler:     &testCommandHandler{output: "test output\n"},
	}

	if err := executor.RegisterCommand(cmd); err != nil {
		t.Fatalf("Failed to register command: %v", err)
	}

	// 测试场景1：获取命令信息
	info, err := executor.GetCommandInfo("test")
	if err != nil {
		t.Errorf("GetCommandInfo() error = %v", err)
	}
	if info.Name != cmd.Name || info.Description != cmd.Description {
		t.Error("Command info mismatch")
	}

	// 测试场景2：获取命令帮助
	help, err := executor.GetCommandHelp("test")
	if err != nil {
		t.Errorf("GetCommandHelp() error = %v", err)
	}
	if help != cmd.Usage {
		t.Errorf("Command help = %q, want %q", help, cmd.Usage)
	}

	// 测试场景3：列出命令
	commands, err := executor.ListCommands(&types.CommandFilter{Category: "test"})
	if err != nil {
		t.Errorf("ListCommands() error = %v", err)
	}
	if len(commands) != 1 || commands[0].Name != cmd.Name {
		t.Error("ListCommands() returned wrong commands")
	}

	// 测试场景4：执行注册的命令
	result, err := executor.Execute(context.Background(), "test", []string{"arg1"}, nil)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("Execute() exit code = %d, want 0", result.ExitCode)
	}
	if result.Output != "test output\n" {
		t.Errorf("Execute() output = %q, want %q", result.Output, "test output\n")
	}

	// 测试场景5：注销命令
	if err := executor.UnregisterCommand("test"); err != nil {
		t.Errorf("UnregisterCommand() error = %v", err)
	}

	// 验证命令已被注销
	if _, err := executor.GetCommandInfo("test"); err == nil {
		t.Error("GetCommandInfo() should return error for removed command")
	}
}
