package commands

import (
	"fmt"

	"github.com/iamlongalong/runshell/pkg/types"
)

// WgetCommand 实现了 wget 命令。
// 用于从网络下载文件。
type WgetCommand struct{}

// Execute 执行 wget 命令。
// 参数：
//   - URL：要下载的文件的URL
//   - [output]：可选的输出文件名
func (c *WgetCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Args) == 0 {
		return nil, fmt.Errorf("wget: missing URL")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"wget"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// TarCommand 实现了 tar 命令。
// 用于创建和解压缩归档文件。
type TarCommand struct{}

// Execute 执行 tar 命令。
// 参数：
//   - -c：创建新的归档文件
//   - -x：从归档文件中提取文件
//   - -f：指定归档文件名
//   - -z：使用gzip压缩
//   - -v：显示详细信息
func (c *TarCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Args) == 0 {
		return nil, fmt.Errorf("tar: missing operand")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"tar"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}

// ZipCommand 实现了 zip 命令。
// 用于创建和解压缩 ZIP 格式的归档文件。
type ZipCommand struct{}

// Execute 执行 zip 命令。
// 参数：
//   - 归档文件名
//   - 要添加到归档的文件列表
func (c *ZipCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.Executor == nil {
		return nil, fmt.Errorf("executor is required")
	}

	if len(ctx.Args) == 0 {
		return nil, fmt.Errorf("zip: missing operand")
	}

	// 准备命令
	execCtx := &types.ExecuteContext{
		Context:  ctx.Context,
		Args:     append([]string{"zip"}, ctx.Args...),
		Options:  ctx.Options,
		Executor: ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(execCtx)
}
