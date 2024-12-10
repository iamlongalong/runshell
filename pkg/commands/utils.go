// Package commands 实现了 RunShell 的内置命令。
// 本文件实现了一系列实用工具命令，包括文件操作、文本处理等功能。
package commands

import (
	"fmt"
	"time"

	"github.com/iamlongalong/runshell/pkg/types"
)

// TouchCommand 实现了 'touch' 命令。
// 用于创建新文件或更新文件的访问和修改时间。
type TouchCommand struct{}

// Execute 执行 touch 命令。
func (c *TouchCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	return ctx.Executor.Execute(ctx)
}

// WriteCommand 实现了 'write' 命令。
// 用于将指定内容写入文件。
type WriteCommand struct{}

// Execute 执行 write 命令。
func (c *WriteCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	return ctx.Executor.Execute(ctx)
}

// FindCommand 实现了 'find' 命令。
type FindCommand struct{}

// Execute 执行 find 命令。
func (c *FindCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	return ctx.Executor.Execute(ctx)
}

// GrepCommand 实现了 'grep' 命令。
type GrepCommand struct{}

// Execute 执行 grep 命令。
func (c *GrepCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	return ctx.Executor.Execute(ctx)
}

// TailCommand 实现了 'tail' 命令。
type TailCommand struct{}

// Execute 执行 tail 命令。
// 参数：
//   - -n<num>：显示的行数（可选，默认10）
//   - 文件路径
func (c *TailCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	return ctx.Executor.Execute(ctx)
}

type XargsCommand struct{}

func (c *XargsCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	return ctx.Executor.Execute(ctx)
}

// MvCommand implements mv command
type MvCommand struct{}

// Execute 执行 mv 命令。
func (c *MvCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Command.Args) != 2 {
		return nil, fmt.Errorf("mv requires source and destination arguments")
	}

	return ctx.Executor.Execute(ctx)
}

// HeadCommand implements head command
type HeadCommand struct{}

func (c *HeadCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Command.Args) < 1 {
		return nil, fmt.Errorf("head requires a file argument")
	}

	return ctx.Executor.Execute(ctx)
}

// SortCommand implements sort command
type SortCommand struct{}

func (c *SortCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Command.Args) < 1 {
		return nil, fmt.Errorf("sort requires a file argument")
	}

	// 构建sort命令的参数
	args := []string{"sort"}
	args = append(args, ctx.Command.Args...)

	// 创建的执行上下文
	newCtx := &types.ExecuteContext{
		Context:     ctx.Context,
		Command:     types.Command{Command: "sort", Args: args},
		Options:     ctx.Options,
		StartTime:   ctx.StartTime,
		IsPiped:     ctx.IsPiped,
		PipeContext: ctx.PipeContext,
		Executor:    ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(newCtx)
}

// UniqCommand implements uniq command
type UniqCommand struct{}

func (c *UniqCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Command.Args) < 1 {
		return nil, fmt.Errorf("uniq requires a file argument")
	}

	return ctx.Executor.Execute(ctx)
}

// NetstatCommand implements netstat command
type NetstatCommand struct{}

func (c *NetstatCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	return ctx.Executor.Execute(ctx)
}

// IfconfigCommand implements ifconfig command
type IfconfigCommand struct{}

func (c *IfconfigCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	return ctx.Executor.Execute(ctx)
}

// CurlCommand implements curl command
type CurlCommand struct{}

// Execute 执行 curl 命令。
func (c *CurlCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Command.Args) < 1 {
		return nil, fmt.Errorf("curl requires a URL argument")
	}

	// 构建curl命令的参数
	args := []string{"curl"}
	args = append(args, ctx.Command.Args...)

	// 创建新的执行上下文
	newCtx := &types.ExecuteContext{
		Context:     ctx.Context,
		Command:     types.Command{Command: "curl", Args: args},
		Options:     ctx.Options,
		StartTime:   ctx.StartTime,
		IsPiped:     ctx.IsPiped,
		PipeContext: ctx.PipeContext,
		Executor:    ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(newCtx)
}

// SedCommand implements sed command
type SedCommand struct{}

func (c *SedCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "sed",
		StartTime:   ctx.StartTime,
	}

	if ctx.Executor == nil {
		result.Error = fmt.Errorf("executor is required")
		return result, result.Error
	}

	if len(ctx.Command.Args) < 2 {
		result.Error = fmt.Errorf("sed requires pattern and file arguments")
		return result, result.Error
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:     ctx.Context,
		Command:     types.Command{Command: "sed", Args: ctx.Command.Args},
		Options:     ctx.Options,
		StartTime:   ctx.StartTime,
		IsPiped:     ctx.IsPiped,
		PipeContext: ctx.PipeContext,
		Executor:    ctx.Executor,
	}

	// 通过executor执行命令
	execResult, err := ctx.Executor.Execute(execCtx)
	if err != nil {
		result.Error = err
		result.ExitCode = 1
		return result, err
	}

	result.ExitCode = execResult.ExitCode
	result.EndTime = time.Now()

	return result, nil
}
