// Package commands 实现了 RunShell 的内置命令。
package commands

import (
	"fmt"
	"strconv"

	"github.com/iamlongalong/runshell/pkg/types"
)

// LSCommand 实现了 ls 命令。
// 用于列出目录内容。
type LSCommand struct{}

// Execute 执行 ls 命令。
// 参数：
//   - [path]：要列出内容的目录路径，默认为当前目录
func (c *LSCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"ls"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// CatCommand 实现了 cat 命令。
// 用于显示文件内容。
type CatCommand struct{}

// Execute 执行 cat 命令。
// 参数：
//   - file：要显示内容的文件路径
func (c *CatCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Args) == 0 {
		return nil, fmt.Errorf("cat: missing file operand")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"cat"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// MkdirCommand 实现了 mkdir 命令。
// 用于创建新目录。
type MkdirCommand struct{}

// Execute 执行 mkdir 命令。
// 参数：
//   - path：要创建的目录路径
//   - [-p]：可选，如果父目录不存在则创建
func (c *MkdirCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Args) == 0 {
		return nil, fmt.Errorf("mkdir: missing operand")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"mkdir"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// RmCommand 实现了 rm 命令。
// 用于删除文件或目录。
type RmCommand struct{}

// Execute 执行 rm 命令。
// 参数：
//   - path：要删除的文件或目录路径
//   - [-r]：可选，递归删除目录及其内容
//   - [-f]：可选，强制删除，不提示
func (c *RmCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Args) == 0 {
		return nil, fmt.Errorf("rm: missing operand")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"rm"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// CpCommand 实现了 cp 命令。
// 用于复制文件或目录。
type CpCommand struct{}

// Execute 执行 cp 命令。
// 参数：
//   - src：源文件或目录路径
//   - dst：目标文件或目录路径
//   - [-r]：可选，递归复制目录
func (c *CpCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Args) < 2 {
		return nil, fmt.Errorf("cp: missing file operand")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"cp"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// PWDCommand 实现了 pwd 命令。
// 用于显示当前工作目录。
type PWDCommand struct{}

// Execute 执行 pwd 命令。
// 无参数。
func (c *PWDCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     []string{"pwd"},
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// ReadFileCommand 实现了 readfile 命令。
// 用于读取文件指定行范围的内容。
type ReadFileCommand struct{}

// Execute 执行 readfile 命令。
// 参数：
//   - file：要读取的文件路径
//   - start_line：起始行号（从1开始）
//   - end_line：结束行号（包含）
func (c *ReadFileCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Args) != 3 {
		return nil, fmt.Errorf("readfile command requires exactly 3 arguments: <file> <start_line> <end_line>")
	}

	// 解析行号
	startLine, err := strconv.Atoi(ctx.Args[1])
	if err != nil || startLine < 1 {
		return nil, fmt.Errorf("invalid start line")
	}

	endLine, err := strconv.Atoi(ctx.Args[2])
	if err != nil || endLine < 1 {
		return nil, fmt.Errorf("invalid end line")
	}

	if startLine > endLine {
		return nil, fmt.Errorf("end line must be greater than or equal to start line")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     []string{"cat", ctx.Args[0]},
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 先检查文件是否存在
	result, err := ctx.Executor.Execute(execCtx)
	if err != nil {
		return nil, fmt.Errorf("no such file or directory")
	}

	// 通过executor执行命令
	return result, nil
}
