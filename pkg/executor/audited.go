// Package executor 实现了命令执行器的核心功能。
// 本文件实现了带审计功能的执行器装饰器。
package executor

import (
	"context"

	"github.com/iamlongalong/runshell/pkg/audit"
	"github.com/iamlongalong/runshell/pkg/types"
)

// AuditedExecutor 是一个执行器装饰器，为基础执行器添加审计日志功能。
// 特性：
// - 记录所有命令执行的审计日志
// - 包装任意类型的基础执行器
// - 透明地转发所有执行器接口方法
// - 自动处理执行失败的情况
type AuditedExecutor struct {
	executor types.Executor // 被装饰的基础执行器
	auditor  *audit.Auditor // 审计日志记录器
}

// NewAuditedExecutor 创建一个新的带审计功能的执行器。
// 参数：
//   - executor：基础执行器实例
//   - auditor：审计日志记录器实例
//
// 返回值：
//   - *AuditedExecutor：带审计功能的执行器实例
func NewAuditedExecutor(executor types.Executor, auditor *audit.Auditor) *AuditedExecutor {
	return &AuditedExecutor{
		executor: executor,
		auditor:  auditor,
	}
}

// Execute 执行命令并记录审计日志。
// 执行流程：
// 1. 创建执行上下文
// 2. 执行命令
// 3. 记录审计日志
// 4. 返回执行结果
//
// 参数：
//   - ctx：上下文，用于取消和超时控制
//   - cmdName：要执行的命令名称
//   - args：命令参数列表
//   - opts：执行选项
//
// 返回值：
//   - *types.ExecuteResult：执行结果
//   - error：执行过程中的错误
func (e *AuditedExecutor) Execute(ctx context.Context, cmdName string, args []string, opts *types.ExecuteOptions) (*types.ExecuteResult, error) {
	// 创建执行上下文
	execCtx := &types.ExecuteContext{
		Context:   ctx,
		Command:   cmdName,
		Args:      args,
		Options:   opts,
		StartTime: types.GetTimeNow(),
	}

	// 执行命令
	result, err := e.executor.Execute(ctx, cmdName, args, opts)
	if err != nil {
		// 处理执行失败的情况
		// 如果没有结果对象，创建一个包含错误信息的结果
		if result == nil {
			result = &types.ExecuteResult{
				CommandName: cmdName,
				ExitCode:    -1,
				StartTime:   execCtx.StartTime,
				EndTime:     types.GetTimeNow(),
				Error:       err,
			}
		}
		// 记录失败的审计日志
		e.auditor.LogCommandExecution(execCtx, result)
		return result, err
	}

	// 记录成功的审计日志
	e.auditor.LogCommandExecution(execCtx, result)
	return result, nil
}

// GetCommandInfo 获取命令信息。
// 直接转发到基础执行器。
//
// 参数：
//   - cmdName：命令名称
//
// 返回值：
//   - *types.Command：命令信息
//   - error：获取过程中的错误
func (e *AuditedExecutor) GetCommandInfo(cmdName string) (*types.Command, error) {
	return e.executor.GetCommandInfo(cmdName)
}

// GetCommandHelp 获取命令的帮助信息。
// 直接转发到基础执行器。
//
// 参数：
//   - cmdName：命令名称
//
// 返回值：
//   - string：命令的使用说明
//   - error：获取过程中的错误
func (e *AuditedExecutor) GetCommandHelp(cmdName string) (string, error) {
	return e.executor.GetCommandHelp(cmdName)
}

// ListCommands 列出符合过滤条件的命令。
// 直接转发到基础执行器。
//
// 参数：
//   - filter：命令过滤器
//
// 返回值：
//   - []*types.Command：符合条件的命令列表
//   - error：列出过程中的错误
func (e *AuditedExecutor) ListCommands(filter *types.CommandFilter) ([]*types.Command, error) {
	return e.executor.ListCommands(filter)
}

// RegisterCommand 注册新命令。
// 直接转发到基础执行器。
//
// 参数：
//   - cmd：要注册的命令
//
// 返回值：
//   - error：注册过程中的错误
func (e *AuditedExecutor) RegisterCommand(cmd *types.Command) error {
	return e.executor.RegisterCommand(cmd)
}

// UnregisterCommand 注销已注册的命令。
// 直接转发到基础执行器。
//
// 参数：
//   - cmdName：要注销的命令名称
//
// 返回值：
//   - error：注销过程中的错误
func (e *AuditedExecutor) UnregisterCommand(cmdName string) error {
	return e.executor.UnregisterCommand(cmdName)
}
