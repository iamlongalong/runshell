// Package executor 实现了命令执行器的核心功能。
// 本文件实现了带审计功能的执行器装饰器。
package executor

import (
	"time"

	"github.com/google/uuid"
	"github.com/iamlongalong/runshell/pkg/log"
	"github.com/iamlongalong/runshell/pkg/types"
)

// AuditedExecutor 是一个带审计功能的执行器装饰器
type AuditedExecutor struct {
	executor types.Executor
	auditor  types.Auditor
}

// NewAuditedExecutor 创建一个新的审计执行器
func NewAuditedExecutor(executor types.Executor, auditor types.Auditor) *AuditedExecutor {
	return &AuditedExecutor{
		executor: executor,
		auditor:  auditor,
	}
}

const (
	AuditedExecutorName = "audited"
)

// Name 返回执行器名称
func (e *AuditedExecutor) Name() string {
	return AuditedExecutorName
}

// ExecuteCommand 执行命令并记录审计日志
func (e *AuditedExecutor) ExecuteCommand(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	return e.Execute(ctx)
}

// Execute 执行命令并记录审计日志
func (e *AuditedExecutor) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	// 创建审计记录
	execution := &types.CommandExecution{
		ID:        uuid.New().String(),
		Command:   ctx.Command,
		StartTime: time.Now(),
		Status:    "STARTED",
	}

	log.Debug("Recording command start in audit log")
	e.auditor.LogCommandExecution(execution)

	// 执行命令
	log.Info("Executing command: %s", ctx.Command)
	result, err := e.executor.Execute(ctx)

	// 更新审计记录
	execution.EndTime = time.Now()
	execution.Error = err
	if result != nil {
		execution.ExitCode = result.ExitCode
	}
	if err != nil {
		execution.Status = "FAILED"
	} else {
		execution.Status = "COMPLETED"
	}

	log.Debug("Recording command completion in audit log")
	e.auditor.LogCommandExecution(execution)

	return result, err
}

// ListCommands 列出所有可用命令
func (e *AuditedExecutor) ListCommands() []types.CommandInfo {
	return e.executor.ListCommands()
}

// Close 关闭执行器
func (e *AuditedExecutor) Close() error {
	// 记录关闭事件
	execution := &types.CommandExecution{
		Command:   types.Command{Command: "close"},
		StartTime: time.Now(),
		Status:    "EXECUTOR_CLOSE",
	}

	log.Debug("Recording executor closure in audit log")
	e.auditor.LogCommandExecution(execution)

	// 关闭底层执行器
	if err := e.executor.Close(); err != nil {
		execution.Error = err
		execution.Status = "EXECUTOR_CLOSE_FAILED"
		e.auditor.LogCommandExecution(execution)
		return err
	}

	return nil
}
