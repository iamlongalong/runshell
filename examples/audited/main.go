// Package main demonstrates how to use the audited executor functionality.
//
// This example shows:
// 1. How to set up an audited executor with different underlying executors
// 2. How to track command execution with audit logs
// 3. How to handle audit events and execution results
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
)

// CustomAuditor 自定义审计器
type CustomAuditor struct{}

// LogCommandExecution 实现审计记录
func (a *CustomAuditor) LogCommandExecution(exec *types.CommandExecution) error {
	fmt.Printf("\n=== Custom Audit Log ===\n")
	fmt.Printf("Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("Command: %s\n", exec.Command)
	fmt.Printf("Status: %s\n", exec.Status)
	if exec.Error != nil {
		fmt.Printf("Error: %v\n", exec.Error)
	}
	fmt.Printf("=====================\n\n")
	return nil
}

func main() {
	// 创建基础执行器
	baseExec := executor.NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{})

	// 创建自定义审计器
	auditor := &CustomAuditor{}

	// 创建审计执行器
	auditedExec := executor.NewAuditedExecutor(baseExec, auditor)
	defer auditedExec.Close()

	// 示例1：执行成功的命令
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Command: types.Command{Command: "echo", Args: []string{"Hello, World!"}},
	}
	result, err := auditedExec.Execute(ctx)
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		return
	}
	fmt.Printf("Command output: %s\n", result.Output)

	// 示例2：执行失败的命令
	ctx = &types.ExecuteContext{
		Context: context.Background(),
		Command: types.Command{Command: "nonexistent", Args: []string{"command"}},
	}
	result, err = auditedExec.Execute(ctx)
	if err != nil {
		fmt.Printf("Expected error executing nonexistent command: %v\n", err)
	}
}

// 使用示例：
//
// 运行示例：
//   go run main.go
//
// 预期输出：
// - 成功命令的审计日志
// - 失败命令的审计日志和错误信息
// - 带环境变量命令的审计日志
// - 超时命令的审计日志
//
// 注意事项：
// - 审计日志包含命令执行的完整生命周期
// - 可以根据需要自定义审计器的行为
// - 支持各种执行上下文和选项
