package commands

import (
	"bytes"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/iamlongalong/runshell/pkg/types"
)

// TestPSCommand 测试 ps 命令的功能。
// 测试场景：
// 1. 执行 ps 命令获取进程列表
// 2. 验证输出包含必要的列标题
// 3. 验证命令执行成功
func TestPSCommand(t *testing.T) {
	cmd := &PSCommand{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	// 执行命令
	result, err := cmd.Execute(&types.ExecuteContext{
		Options: &types.ExecuteOptions{
			Stdout: stdout,
			Stderr: stderr,
		},
		StartTime: time.Now(),
	})

	// 验证执行结果
	if err != nil {
		t.Errorf("PSCommand.Execute() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("PSCommand.Execute() exit code = %v, want 0", result.ExitCode)
	}

	// 验证输出格式
	output := stdout.String()
	if !strings.Contains(output, "PID") || !strings.Contains(output, "CPU%") {
		t.Errorf("PSCommand.Execute() output does not contain expected headers")
	}
}

// TestTopCommand 测试 top 命令的功能。
// 测试场景：
// 1. 执行 top 命令获取系统概览
// 2. 验证输出包含系统信息
// 3. 验证命令执行成功
func TestTopCommand(t *testing.T) {
	cmd := &TopCommand{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	// 执行命令
	result, err := cmd.Execute(&types.ExecuteContext{
		Options: &types.ExecuteOptions{
			Stdout: stdout,
			Stderr: stderr,
		},
		StartTime: time.Now(),
	})

	// 验证执行结果
	if err != nil {
		t.Errorf("TopCommand.Execute() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("TopCommand.Execute() exit code = %v, want 0", result.ExitCode)
	}

	// 验证输出内容
	output := stdout.String()
	if !strings.Contains(output, "System Overview") {
		t.Errorf("TopCommand.Execute() output does not contain system overview")
	}
}

// TestDFCommand 测试 df 命令的功能。
// 测试场景：
// 1. 执行 df 命令获取磁盘使用情况
// 2. 验证输出包含必要的列标题
// 3. 验证命令执行成功
func TestDFCommand(t *testing.T) {
	cmd := &DFCommand{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	// 执行命令
	result, err := cmd.Execute(&types.ExecuteContext{
		Options: &types.ExecuteOptions{
			Stdout: stdout,
			Stderr: stderr,
		},
		StartTime: time.Now(),
	})

	// 验证执行结果
	if err != nil {
		t.Errorf("DFCommand.Execute() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("DFCommand.Execute() exit code = %v, want 0", result.ExitCode)
	}

	// 验证输出格式
	output := stdout.String()
	if !strings.Contains(output, "Filesystem") || !strings.Contains(output, "Size") {
		t.Errorf("DFCommand.Execute() output does not contain expected headers")
	}
}

// TestUNameCommand 测试 uname 命令的功能。
// 测试场景：
// 1. 测试无参数执行（只显示操作系统）
// 2. 测试 -a 参数执行（显示完整系统信息）
// 3. 验证输出内容的正确性
func TestUNameCommand(t *testing.T) {
	cmd := &UNameCommand{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	// 测试场景1：无参数执行
	result, err := cmd.Execute(&types.ExecuteContext{
		Options: &types.ExecuteOptions{
			Stdout: stdout,
			Stderr: stderr,
		},
		StartTime: time.Now(),
	})

	// 验证基本执行结果
	if err != nil {
		t.Errorf("UNameCommand.Execute() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("UNameCommand.Execute() exit code = %v, want 0", result.ExitCode)
	}

	// 验证输出不为空
	output := stdout.String()
	if len(strings.TrimSpace(output)) == 0 {
		t.Error("UNameCommand.Execute() output is empty")
	}

	// 测试场景2：使用 -a 参数
	stdout.Reset()
	stderr.Reset()

	result, err = cmd.Execute(&types.ExecuteContext{
		Args: []string{"-a"},
		Options: &types.ExecuteOptions{
			Stdout: stdout,
			Stderr: stderr,
		},
		StartTime: time.Now(),
	})

	// 验证带参数执行结果
	if err != nil {
		t.Errorf("UNameCommand.Execute(-a) error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("UNameCommand.Execute(-a) exit code = %v, want 0", result.ExitCode)
	}

	// 验证详细输出内容
	output = stdout.String()
	if !strings.Contains(output, runtime.GOARCH) {
		t.Error("UNameCommand.Execute(-a) output does not contain architecture")
	}
}

// TestEnvCommand 测试 env 命令的功能。
// 测试场景：
// 1. 测试无参数执行（显示所有环境变量）
// 2. 测试带过滤参数执行（只显示匹配的环境变量）
// 3. 验证环境变量的显示和过滤功能
func TestEnvCommand(t *testing.T) {
	// 设置测试环境变量
	testKey := "TEST_ENV_VAR"
	testValue := "test_value"
	os.Setenv(testKey, testValue)
	defer os.Unsetenv(testKey)

	cmd := &EnvCommand{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	// 测试场景1：无参数执行
	result, err := cmd.Execute(&types.ExecuteContext{
		Options: &types.ExecuteOptions{
			Stdout: stdout,
			Stderr: stderr,
		},
		StartTime: time.Now(),
	})

	// 验证基本执行结果
	if err != nil {
		t.Errorf("EnvCommand.Execute() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("EnvCommand.Execute() exit code = %v, want 0", result.ExitCode)
	}

	// 验证输出包含测试环境变量
	output := stdout.String()
	if !strings.Contains(output, testKey+"="+testValue) {
		t.Error("EnvCommand.Execute() output does not contain test environment variable")
	}

	// 测试场景2：带过滤参数执行
	stdout.Reset()
	stderr.Reset()

	result, err = cmd.Execute(&types.ExecuteContext{
		Args: []string{"TEST_"},
		Options: &types.ExecuteOptions{
			Stdout: stdout,
			Stderr: stderr,
		},
		StartTime: time.Now(),
	})

	// 验证过滤执行结果
	if err != nil {
		t.Errorf("EnvCommand.Execute(pattern) error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("EnvCommand.Execute(pattern) exit code = %v, want 0", result.ExitCode)
	}

	// 验证过滤结果的正确性
	output = stdout.String()
	if !strings.Contains(output, testKey+"="+testValue) {
		t.Error("EnvCommand.Execute(pattern) output does not contain matching environment variable")
	}
	if strings.Contains(output, "PATH=") {
		t.Error("EnvCommand.Execute(pattern) output contains non-matching environment variable")
	}
}

// TestKillCommand 测试 kill 命令的功能。
// 测试场景：
// 1. 测试无参数执行（应该返回错误）
// 2. 测试无效进程ID（应该返回错误）
// 3. 验证错误处理机制
func TestKillCommand(t *testing.T) {
	cmd := &KillCommand{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	// 测试场景1：无参数执行
	result, err := cmd.Execute(&types.ExecuteContext{
		Options: &types.ExecuteOptions{
			Stdout: stdout,
			Stderr: stderr,
		},
		StartTime: time.Now(),
	})

	// 验证无参数错误处理
	if err == nil {
		t.Error("KillCommand.Execute() without args should return error")
	}
	if result != nil && result.ExitCode == 0 {
		t.Error("KillCommand.Execute() without args should not return success")
	}

	// 测试场景2：无效进程ID
	stdout.Reset()
	stderr.Reset()

	result, err = cmd.Execute(&types.ExecuteContext{
		Args: []string{"invalid"},
		Options: &types.ExecuteOptions{
			Stdout: stdout,
			Stderr: stderr,
		},
		StartTime: time.Now(),
	})

	// 验证无效参数错误处理
	if err == nil {
		t.Error("KillCommand.Execute() with invalid PID should return error")
	}
	if result != nil && result.ExitCode == 0 {
		t.Error("KillCommand.Execute() with invalid PID should not return success")
	}
}
