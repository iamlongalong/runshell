// Package commands 实现了 RunShell 的内置命令。
package commands

import (
	"fmt"

	"github.com/iamlongalong/runshell/pkg/types"
)

// PSCommand 实现了 'ps' 命令。
// 用于显示系统中运行的进程信息。
// 输出格式：PID、CPU使用率、内存使用率、进程状态、进程名称。
type PSCommand struct{}

// Execute 执行 ps 命令。
// 输出格式：
//   - PID：进程ID
//   - CPU%：CPU使用率
//   - MEM%：内存使用率
//   - STATE：进程状态
//   - NAME：进程名称
func (c *PSCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"ps"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// TopCommand 实现了 'top' 命令。
// 用于实时显示系统资源使用情况和进程信息。
// 显示内容包括：系统概览（主机名、操作系统、运行时间、CPU、内存）和进程列表。
type TopCommand struct{}

// Execute 执行 top 命令。
// 输出内容：
//   - 系统概览：主机名、操作系统、运行时间、CPU数量、内存使用情况
//   - 进程列表：PID、CPU使用率、内存使用率、状态、名称
func (c *TopCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"top"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// DFCommand 实现了 'df' 命令。
// 用于显示文件系统的磁盘空间使用情况。
// 显示内容包括：文件系统、总大小、已用空间、可用空间、使用率、挂载点。
type DFCommand struct{}

// Execute 执行 df 命令。
// 输出格式：
//   - Filesystem：文件系统设备
//   - Size：总大小（GB）
//   - Used：已用空间（GB）
//   - Avail：可用空间（GB）
//   - Use%：使用率
//   - Mounted on：挂载点
func (c *DFCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"df"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// UNameCommand 实现了 'uname' 命令。
// 用于显示系统信息。
// 支持 -a 选项显示详细信息。
type UNameCommand struct{}

// Execute 执行 uname 命令。
// 参数：
//   - 无参数：只显示操作系统名称
//   - -a：显示完整系统信息（操作系统、主机名、内核版本、平台、架构、版本）
func (c *UNameCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"uname"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// EnvCommand 实现了 'env' 命令。
// 用于显示系统环境变量。
// 支持按模式过滤环境变量。
type EnvCommand struct{}

// Execute 执行 env 命令。
// 参数：
//   - 无参数：显示所有环境变量
//   - 有参数：按参数指定的前缀过滤环境变量（不区分大小写）
func (c *EnvCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"env"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// KillCommand 实现了 'kill' 命令。
// 用于终止指定的进程。
// 需要提供进程ID作为参数。
type KillCommand struct{}

// Execute 执行 kill 命令。
// 参数：
//   - 需要至少一个进程ID
//   - 支持同时终止多个进程
//
// 错误处理：
//   - 无效的进程ID
//   - 进程不存在
//   - 无权限终止进程
func (c *KillCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Args) == 0 {
		return nil, fmt.Errorf("kill: usage: kill [-s sigspec | -n signum | -sigspec] pid | jobspec ... or kill -l [sigspec]")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"kill"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}
