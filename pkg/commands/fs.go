// Package commands 实现了 RunShell 的内置命令。
package commands

import (
	"fmt"
	"strconv"

	"github.com/iamlongalong/runshell/pkg/types"
)

// LSCommand 实现了 ls 命令。
// 用于列出目录内容。
type LSCommand struct {
}

func (c *LSCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "ls",
		Description: "List directory contents",
		Usage:       "ls [options] [directory]",
		Category:    "file",
	}
}

// Execute 执行 ls 命令。
func (c *LSCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	return ctx.Executor.Execute(ctx)
}

// 用于显示文件内容。
type CatCommand struct{}

func (c *CatCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "cat",
		Description: "Concatenate and print files",
		Usage:       "cat [options] [file...]",
		Category:    "file",
	}
}

// Execute 执行 cat 命令。
// 参数：
//   - file：要显示内容的文件路径
func (c *CatCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Command.Args) == 0 {
		return nil, fmt.Errorf("cat: missing file operand")
	}

	return ctx.Executor.Execute(ctx)
}

// 用于创建新目录。
type MkdirCommand struct{}

func (c *MkdirCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "mkdir",
		Description: "Make directories",
		Usage:       "mkdir [options] directory...",
		Category:    "file",
	}
}

// Execute 执行 mkdir 命令。
// 参数：
//   - path：要创建的目录路径
//   - [-p]：可选，如果父目录不存在则创建
func (c *MkdirCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Command.Args) == 0 {
		return nil, fmt.Errorf("mkdir: missing operand")
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(ctx)
}

// RmCommand 实现了 rm 命令。
// 用于删除文件或目录。
type RmCommand struct{}

func (c *RmCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "rm",
		Description: "Remove files or directories",
		Usage:       "rm [options] file...",
		Category:    "file",
	}
}

// Execute 执行 rm 命令。
func (c *RmCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Command.Args) == 0 {
		return nil, fmt.Errorf("rm: missing operand")
	}

	return ctx.Executor.Execute(ctx)
}

// CpCommand 实现了 cp 命令。
// 用于复制文件或目录。
type CpCommand struct{}

func (c *CpCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "cp",
		Description: "Copy files and directories",
		Usage:       "cp [options] source... destination",
		Category:    "file",
	}
}

// Execute 执行 cp 命令。
func (c *CpCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Command.Args) < 2 {
		return nil, fmt.Errorf("cp: missing file operand")
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(ctx)
}

// PWDCommand 实现了 pwd 命令。
// 用于显示当前工作目录。
type PWDCommand struct{}

func (c *PWDCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "pwd",
		Description: "Print working directory",
		Usage:       "pwd",
		Category:    "file",
	}
}

// Execute 执行 pwd 命令。
// 无参数。
func (c *PWDCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	return ctx.Executor.Execute(ctx)
}

// ReadFileCommand 实现了 readfile 命令。
// 用于读取文件指定行范围的内容。
type ReadFileCommand struct{}

func (c *ReadFileCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "readfile",
		Description: "Read file content by line range",
		Usage:       "readfile [options] file start_line end_line",
		Category:    "file",
	}
}

// Execute 执行 readfile 命令。
// 参数：
//   - file：要读取的文件路径
//   - start_line：起始行号（从1开始）
//   - end_line：结束行号（包含）
func (c *ReadFileCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Command.Args) != 3 {
		return nil, fmt.Errorf("readfile command requires exactly 3 arguments: <file> <start_line> <end_line>")
	}

	// 解析行号
	startLine, err := strconv.Atoi(ctx.Command.Args[1])
	if err != nil || startLine < 1 {
		return nil, fmt.Errorf("invalid start line")
	}

	endLine, err := strconv.Atoi(ctx.Command.Args[2])
	if err != nil || endLine < 1 {
		return nil, fmt.Errorf("invalid end line")
	}

	if startLine > endLine {
		return nil, fmt.Errorf("end line must be greater than or equal to start line")
	}

	// 准备命令
	execCtx := ctx.Copy()
	execCtx.Command.Command = "cat"
	execCtx.Command.Args = ctx.Command.Args
	execCtx.Options = ctx.Options
	execCtx.Executor = ctx.Executor

	return ctx.Executor.Execute(execCtx)
}
