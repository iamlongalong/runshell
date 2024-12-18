package commands

import (
	"fmt"

	"github.com/iamlongalong/runshell/pkg/types"
)

// WgetCommand 实现了 wget 命令。
// 用于从网络下载文件。
type WgetCommand struct{}

func (c *WgetCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "wget",
		Description: "Download files from the web",
		Usage:       "wget [options] url",
		Category:    "network",
	}
}

// Execute 执行 wget 命令。
func (c *WgetCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Command.Args) == 0 {
		return nil, fmt.Errorf("wget: missing URL")
	}

	return ctx.Executor.ExecuteCommand(ctx)
}

// TarCommand 实现了 tar 命令。
type TarCommand struct{}

func (c *TarCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "tar",
		Description: "Create or extract archives",
		Usage:       "tar [options] [archive] [file...]",
		Category:    "file",
	}
}

// Execute 执行 tar 命令。
func (c *TarCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Command.Args) == 0 {
		return nil, fmt.Errorf("tar: missing operand")
	}

	return ctx.Executor.ExecuteCommand(ctx)
}

// ZipCommand 实现了 zip 命令。
// 用于创建和解压缩 ZIP 格式的归档文件。
type ZipCommand struct{}

func (c *ZipCommand) Info() types.CommandInfo {
	return types.CommandInfo{
		Name:        "zip",
		Description: "Package and compress files",
		Usage:       "zip [options] [zipfile] [file...]",
		Category:    "file",
	}
}

// Execute 执行 zip 命令。
// 参数：
//   - 归档文件名
//   - 要添加到归档的文件列表
func (c *ZipCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Command.Args) == 0 {
		return nil, fmt.Errorf("zip: missing operand")
	}

	return ctx.Executor.ExecuteCommand(ctx)
}
