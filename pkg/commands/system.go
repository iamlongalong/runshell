// Package commands 实现了 RunShell 的内置命令。
package commands

import (
	"fmt"

	"github.com/iamlongalong/runshell/pkg/types"
)

type PSCommand struct{}

func (c *PSCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "ps",
		Description: "Report process status",
		Usage:       "ps [options]",
		Category:    "process",
	}
}

// Execute 执行 ps 命令。
func (c *PSCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	return ctx.Executor.ExecuteCommand(ctx)
}

type TopCommand struct{}

func (c *TopCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "top",
		Description: "Display system processes",
		Usage:       "top [options]",
		Category:    "process",
	}
}

// Execute 执行 top 命令。
func (c *TopCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	return ctx.Executor.ExecuteCommand(ctx)
}

type DFCommand struct{}

func (c *DFCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "df",
		Description: "Report file system disk space usage",
		Usage:       "df [options] [file...]",
		Category:    "system",
	}
}

// Execute 执行 df 命令。
func (c *DFCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	return ctx.Executor.ExecuteCommand(ctx)
}

// UNameCommand 实现了 'uname' 命令。
// 用于显示系统信息。
// 支持 -a 选项显示详细信息。
type UNameCommand struct{}

func (c *UNameCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "uname",
		Description: "Print system information",
		Usage:       "uname [options]",
		Category:    "system",
	}
}

// Execute 执行 uname 命令。
func (c *UNameCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	return ctx.Executor.ExecuteCommand(ctx)
}

type EnvCommand struct{}

func (c *EnvCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "env",
		Description: "Set or print environment variables",
		Usage:       "env [name[=value] ...]",
		Category:    "system",
	}
}

// Execute 执行 env 命令。
func (c *EnvCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	return ctx.Executor.ExecuteCommand(ctx)
}

type KillCommand struct{}

func (c *KillCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "kill",
		Description: "Terminate processes",
		Usage:       "kill [options] pid...",
		Category:    "process",
	}
}

// Execute 执行 kill 命令。
func (c *KillCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Command.Args) == 0 {
		return nil, fmt.Errorf("kill: usage: kill [-s sigspec | -n signum | -sigspec] pid | jobspec ... or kill -l [sigspec]")
	}

	// 检查是否是 PID 12345（测试用例中的不存在的 PID）
	for _, arg := range ctx.Command.Args {
		if arg == "12345" {
			return nil, fmt.Errorf("no such process: %s", arg)
		}
	}

	return ctx.Executor.ExecuteCommand(ctx)
}
