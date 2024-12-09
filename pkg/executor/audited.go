// Package executor 实现了命令执行器的核心功能。
// 本文件实现了带审计功能的执行器装饰器。
package executor

import (
	"github.com/iamlongalong/runshell/pkg/audit"
	"github.com/iamlongalong/runshell/pkg/types"
)

// AuditedExecutor 审计执行器
type AuditedExecutor struct {
	executor types.Executor
	auditor  *audit.Auditor
}

// NewAuditedExecutor 创建新的审计执行器
func NewAuditedExecutor(executor types.Executor, auditor *audit.Auditor) *AuditedExecutor {
	return &AuditedExecutor{
		executor: executor,
		auditor:  auditor,
	}
}

// SetOptions 设置执行选项
func (e *AuditedExecutor) SetOptions(options *types.ExecuteOptions) {
	e.executor.SetOptions(options)
}

// Execute 执行命令并记录审计日志
func (e *AuditedExecutor) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	// 记录开始执行
	e.auditor.LogCommandExecution(&audit.CommandExecution{
		Command:   ctx.Args[0],
		Args:      ctx.Args[1:],
		StartTime: types.GetTimeNow(),
		Status:    "started",
	})

	// 执行命令
	result, err := e.executor.Execute(ctx)

	// 记录执行结果
	e.auditor.LogCommandExecution(&audit.CommandExecution{
		Command:  ctx.Args[0],
		Args:     ctx.Args[1:],
		ExitCode: result.ExitCode,
		Error:    err,
		EndTime:  types.GetTimeNow(),
		Status:   "completed",
	})

	return result, err
}

// ListCommands 列出所有可用命令
func (e *AuditedExecutor) ListCommands() []types.CommandInfo {
	return e.executor.ListCommands()
}

// Close 关闭执行器，清理资源
func (e *AuditedExecutor) Close() error {
	// 记录关闭事件
	e.auditor.LogCommandExecution(&audit.CommandExecution{
		Command: "close",
		Status:  "completed",
		EndTime: types.GetTimeNow(),
	})

	// 关闭底层执行器
	return e.executor.Close()
}
