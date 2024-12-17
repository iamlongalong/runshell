// Package commands 实现了 RunShell 的内置命令。
package commands

import (
	"fmt"

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
		Usage:       "ls [path]",
	}
}

// Execute 执行 ls 命令。
func (c *LSCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	// 直接使用执行器执行命令
	return ctx.Executor.Execute(ctx)
}

// CatCommand 实现了 cat 命令。
// 用于显示文件内容。
type CatCommand struct{}

func (c *CatCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "cat",
		Description: "Concatenate and print files",
		Usage:       "cat [file...]",
	}
}

// Execute 执行 cat 命令。
func (c *CatCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Command.Args) == 0 {
		return nil, fmt.Errorf("no file specified")
	}

	// 直接使用执行器执行命令
	return ctx.Executor.Execute(ctx)
}

// MkdirCommand 实现了 mkdir 命令。
// 用于创建新目录。
type MkdirCommand struct{}

func (c *MkdirCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "mkdir",
		Description: "Create directories",
		Usage:       "mkdir [directory...]",
	}
}

// Execute 执行 mkdir 命令。
func (c *MkdirCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Command.Args) == 0 {
		return nil, fmt.Errorf("no directory specified")
	}

	// 直接使用执行器执行命令
	return ctx.Executor.Execute(ctx)
}

// RmCommand 实现了 rm 命令。
// 用于删除文件或目录。
type RmCommand struct{}

func (c *RmCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "rm",
		Description: "Remove files or directories",
		Usage:       "rm [file...]",
	}
}

// Execute 执行 rm 命令。
func (c *RmCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Command.Args) == 0 {
		return nil, fmt.Errorf("no file specified")
	}

	// 直接使用执行器执行命令
	return ctx.Executor.Execute(ctx)
}

// CpCommand 实现了 cp 命令。
// 用于复制文件或目录。
type CpCommand struct{}

func (c *CpCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "cp",
		Description: "Copy files and directories",
		Usage:       "cp [source] [dest]",
	}
}

// Execute 执行 cp 命令。
func (c *CpCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Command.Args) < 2 {
		return nil, fmt.Errorf("cp requires source and destination")
	}

	// 直接使用执行器执行命令
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
	}
}

// Execute 执行 pwd 命令。
func (c *PWDCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	workDir := "/"
	if ctx.Options != nil && ctx.Options.WorkDir != "" {
		workDir = ctx.Options.WorkDir
	}

	if ctx.Options != nil && ctx.Options.Stdout != nil {
		fmt.Fprintln(ctx.Options.Stdout, workDir)
	}

	return &types.ExecuteResult{
		CommandName: "pwd",
		ExitCode:    0,
		Output:      workDir + "\n",
	}, nil
}

// ReadFileCommand 实现了 readfile 命令。
// 用于读取文件指定行范围的内容。
type ReadFileCommand struct{}

func (c *ReadFileCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "readfile",
		Description: "Read file contents",
		Usage:       "readfile [file] [start_line] [end_line]",
	}
}

// Execute 执行 readfile 命令。
func (c *ReadFileCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Command.Args) < 1 {
		return nil, fmt.Errorf("no file specified")
	}

	// 直接使用执行器执行命令
	return ctx.Executor.Execute(ctx)
}
