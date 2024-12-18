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
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is nil")
	}

	// 创建一个新的上下文，避免递归调用
	newCtx := ctx.Copy() // 使用 Copy 方法复制上下文
	newCtx.Command = types.Command{Command: "ls", Args: ctx.Command.Args}

	// 直接使用执行器执行系统命令
	return newCtx.Executor.ExecuteCommand(newCtx)
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
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is nil")
	}

	if len(ctx.Command.Args) == 0 {
		return nil, fmt.Errorf("no file specified")
	}

	// 创建一个新的上下文，避免递归调用
	newCtx := ctx.Copy()
	newCtx.Command = types.Command{Command: "cat", Args: ctx.Command.Args}

	// 直接使用执行器执行系统命令
	return newCtx.Executor.ExecuteCommand(newCtx)
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
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is nil")
	}

	if len(ctx.Command.Args) == 0 {
		return nil, fmt.Errorf("no directory specified")
	}

	// 创建一个新的上下文，避免递归调用
	newCtx := ctx.Copy()
	newCtx.Command = types.Command{Command: "mkdir", Args: ctx.Command.Args}

	// 直接使用执行器执行系统命令
	return newCtx.Executor.ExecuteCommand(newCtx)
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
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is nil")
	}

	if len(ctx.Command.Args) == 0 {
		return nil, fmt.Errorf("no file specified")
	}

	// 创建一个新的上下文，避免递归调用
	newCtx := ctx.Copy()
	newCtx.Command = types.Command{Command: "rm", Args: ctx.Command.Args}

	// 直接使用执行器执行系统命令
	return newCtx.Executor.ExecuteCommand(newCtx)
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
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is nil")
	}

	if len(ctx.Command.Args) < 2 {
		return nil, fmt.Errorf("cp requires source and destination")
	}

	// 创建一个新的上下文，避免递归调用
	newCtx := ctx.Copy()
	newCtx.Command = types.Command{Command: "cp", Args: ctx.Command.Args}

	// 直接使用执行器执行系统命令
	return newCtx.Executor.ExecuteCommand(newCtx)
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
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is nil")
	}

	if len(ctx.Command.Args) < 1 {
		return nil, fmt.Errorf("no file specified")
	}

	// 如果提供了行号参数，验证它们
	if len(ctx.Command.Args) >= 3 {
		startLine := ctx.Command.Args[1]
		endLine := ctx.Command.Args[2]
		if startLine < "1" {
			return nil, fmt.Errorf("invalid start line: %s", startLine)
		}
		if endLine < "1" {
			return nil, fmt.Errorf("invalid end line: %s", endLine)
		}
		if startLine > endLine {
			return nil, fmt.Errorf("start line (%s) is greater than end line (%s)", startLine, endLine)
		}
	}

	// 创建一个新的上下文，避免递归调用
	newCtx := ctx.Copy()
	newCtx.Command = types.Command{Command: "cat", Args: []string{ctx.Command.Args[0]}}

	// 直接使用执行器执行系统命令
	return newCtx.Executor.ExecuteCommand(newCtx)
}
