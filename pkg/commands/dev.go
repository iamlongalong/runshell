package commands

import (
	"fmt"

	"github.com/iamlongalong/runshell/pkg/types"
)

// GitCommand 实现了 git 命令。
// 用于执行 Git 版本控制操作。
type GitCommand struct{}

// Execute 执行 git 命令。
// 参数：
//   - 所有 git 子命令和参数
func (c *GitCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"git"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// GoCommand 实现了 go 命令。
// 用于执行 Go 语言工具链操作。
type GoCommand struct{}

// Execute 执行 go 命令。
// 参数：
//   - 所有 go 子命令和参数
func (c *GoCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"go"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// PythonCommand 实现了 python 命令。
// 用于执行 Python 解释器和脚本。
type PythonCommand struct{}

// Execute 执行 python 命令。
// 参数：
//   - [script]：要执行的 Python 脚本文件
//   - [args...]：传递给脚本的参数
func (c *PythonCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Args) == 0 {
		return nil, fmt.Errorf("python: missing operand")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"python3"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// PipCommand 实现了 pip 命令。
// 用于管理 Python 包。
type PipCommand struct{}

// Execute 执行 pip 命令。
// 参数：
//   - install/uninstall：安装或卸载包
//   - [package]：包名
//   - [-r requirements.txt]：从文件安装依赖
func (c *PipCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Args) == 0 {
		return nil, fmt.Errorf("pip: missing operand")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"pip"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// DockerCommand 实现了 docker 命令。
// 用于管理容器和镜像。
type DockerCommand struct{}

// Execute 执行 docker 命令。
// 参数：
//   - 所有 docker 子命令和参数
func (c *DockerCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Args) == 0 {
		return nil, fmt.Errorf("docker: missing operand")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"docker"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// NodeCommand 实现了 node 命令。
// 用于执行 Node.js 解释器和脚本。
type NodeCommand struct{}

// Execute 执行 node 命令。
// 参数：
//   - [script]：要执行的 JavaScript 文件
//   - [args...]：传递给脚本的参数
func (c *NodeCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Args) == 0 {
		return nil, fmt.Errorf("node: missing operand")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"node"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// NPMCommand 实现了 npm 命令。
// 用于管理 Node.js 包。
type NPMCommand struct{}

// Execute 执行 npm 命令。
// 参数：
//   - install/uninstall：安装或卸载包
//   - [package]：包名
//   - [-g]：全局安装
func (c *NPMCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Args) == 0 {
		return nil, fmt.Errorf("npm: missing operand")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"npm"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}
