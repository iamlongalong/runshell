package commands

import (
	"fmt"

	"github.com/iamlongalong/runshell/pkg/types"
)

// GitCommand 实现了 git 命令。
// 用于执行 Git 版本控制操作。
type GitCommand struct{}

func (c *GitCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "git",
		Description: "Git version control",
		Usage:       "git [options] <command> [args]",
		Category:    "vcs",
	}
}

// Execute 执行 git 命令。
// 参数：
//   - 所有 git 子命令和参数
func (c *GitCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	return ctx.Executor.ExecuteCommand(ctx)
}

// GoCommand 实现了 go 命令。
// 用于执行 Go 语言工具链操作。
type GoCommand struct{}

func (c *GoCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "go",
		Description: "Go language tools",
		Usage:       "go <command> [args]",
		Category:    "language",
	}
}

// Execute 执行 go 命令。
// 参数：
//   - 所有 go 子命令和参数
func (c *GoCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	return ctx.Executor.ExecuteCommand(ctx)
}

// PythonCommand 实现了 python 命令。
// 用于执行 Python 解释器和脚本。
type PythonCommand struct{}

func (c *PythonCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "python",
		Description: "Run Python interpreter",
		Usage:       "python [options] [script] [args]",
		Category:    "language",
	}
}

// Execute 执行 python 命令。
func (c *PythonCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Command.Args) == 0 {
		return nil, fmt.Errorf("python: missing operand")
	}

	return ctx.Executor.ExecuteCommand(ctx)
}

// PipCommand 实现了 pip 命令。
// 用于管理 Python 包。
type PipCommand struct{}

func (c *PipCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "pip",
		Description: "Python package installer",
		Usage:       "pip [options] <command> [args]",
		Category:    "package",
	}
}

// Execute 执行 pip 命令。
func (c *PipCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Command.Args) == 0 {
		return nil, fmt.Errorf("pip: missing operand")
	}

	return ctx.Executor.ExecuteCommand(ctx)
}

// 用于管理容器和镜像。
type DockerCommand struct{}

func (c *DockerCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "docker",
		Description: "Docker container operations",
		Usage:       "docker [options] <command> [args]",
		Category:    "container",
	}
}

// Execute 执行 docker 命令。
func (c *DockerCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Command.Args) == 0 {
		return nil, fmt.Errorf("docker: missing operand")
	}

	return ctx.Executor.ExecuteCommand(ctx)
}

// NodeCommand 实现了 node 命令。
// 用于执行 Node.js 解释器和脚本。
type NodeCommand struct{}

func (c *NodeCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "node",
		Description: "Run Node.js interpreter",
		Usage:       "node [options] [script] [args]",
		Category:    "language",
	}
}

// Execute 执行 node 命令。
func (c *NodeCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Command.Args) == 0 {
		return nil, fmt.Errorf("node: missing operand")
	}

	return ctx.Executor.ExecuteCommand(ctx)
}

// NPMCommand 实现了 npm 命令。
// 用于管理 Node.js 包。
type NPMCommand struct{}

func (c *NPMCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "npm",
		Description: "Node.js package manager",
		Usage:       "npm [options] <command> [args]",
		Category:    "package",
	}
}

// Execute 执行 npm 命令。
func (c *NPMCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Command.Args) == 0 {
		return nil, fmt.Errorf("npm: missing operand")
	}

	return ctx.Executor.ExecuteCommand(ctx)
}
