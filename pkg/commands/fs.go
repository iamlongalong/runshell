// Package commands 实现了 RunShell 的内置命令。
package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/iamlongalong/runshell/pkg/types"
)

// LSCommand 实现了 'ls' 命令。
// 用于列出目录内容，显示文件和目录的详细信息。
type LSCommand struct{}

// Execute 执行 ls 命令。
// 参数：
//   - 如果没有参数，列出当前目录内容
//   - 如果有参数，列出指定目录内容
//
// 输出格式：权限模式、文件大小、文件名
func (c *LSCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "ls",
		StartTime:   ctx.StartTime,
	}

	path := "."
	if len(ctx.Args) > 0 {
		path = ctx.Args[0]
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(ctx.Options.WorkDir, path)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		result.Error = err
		return result, err
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fmt.Fprintf(ctx.Options.Stdout, "%s\t%d\t%s\n", info.Mode(), info.Size(), entry.Name())
	}

	return result, nil
}

// CatCommand 实现了 'cat' 命令。
// 用于查看文件内容，支持同时查看多个文件。
type CatCommand struct{}

// Execute 执行 cat 命令。
// 参数：
//   - 至少需要一个文件路径参数
//   - 支持多个文件路径，按顺序输出内容
//
// 错误处理：
//   - 如果文件不存在或无法访问，返回错误
func (c *CatCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "cat",
		StartTime:   ctx.StartTime,
	}

	if len(ctx.Args) == 0 {
		result.Error = fmt.Errorf("no file specified")
		return result, result.Error
	}

	for _, path := range ctx.Args {
		if !filepath.IsAbs(path) {
			path = filepath.Join(ctx.Options.WorkDir, path)
		}

		file, err := os.Open(path)
		if err != nil {
			result.Error = err
			return result, err
		}
		defer file.Close()

		_, err = io.Copy(ctx.Options.Stdout, file)
		if err != nil {
			result.Error = err
			return result, err
		}
	}

	return result, nil
}

// MkdirCommand 实现了 'mkdir' 命令。
// 用于创建目录，支持创建多级目录。
type MkdirCommand struct{}

// Execute 执行 mkdir 命令。
// 参数：
//   - 至少需要一个目录路径参数
//   - 支持创建多个目录
//   - 自动创建所需的父目录
//
// 权限：
//   - 新创建的目录权限为 0755
func (c *MkdirCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "mkdir",
		StartTime:   ctx.StartTime,
	}

	if len(ctx.Args) == 0 {
		result.Error = fmt.Errorf("no directory specified")
		return result, result.Error
	}

	for _, path := range ctx.Args {
		if !filepath.IsAbs(path) {
			path = filepath.Join(ctx.Options.WorkDir, path)
		}

		err := os.MkdirAll(path, 0755)
		if err != nil {
			result.Error = err
			return result, err
		}
	}

	return result, nil
}

// RmCommand 实现了 'rm' 命令。
// 用于删除文件或目录。
type RmCommand struct{}

// Execute 执行 rm 命令。
// 参数：
//   - 至少需要一个路径参数
//   - -r/-R：递归删除目录及其内容
//   - -f：强制删除，忽略不存在的文件和错误
//
// 安全性：
//   - 删除目录时必须使用 -r 选项
//   - 非强制模式下会返回错误信息
func (c *RmCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "rm",
		StartTime:   ctx.StartTime,
	}

	if len(ctx.Args) == 0 {
		result.Error = fmt.Errorf("no path specified")
		return result, result.Error
	}

	recursive := false
	force := false
	args := ctx.Args

	// 处理选项
	for i := 0; i < len(args); i++ {
		if args[i] == "-r" || args[i] == "-R" {
			recursive = true
			args = append(args[:i], args[i+1:]...)
			i--
		} else if args[i] == "-f" {
			force = true
			args = append(args[:i], args[i+1:]...)
			i--
		}
	}

	for _, path := range args {
		if !filepath.IsAbs(path) {
			path = filepath.Join(ctx.Options.WorkDir, path)
		}

		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) && force {
				continue
			}
			result.Error = err
			return result, err
		}

		if info.IsDir() && !recursive {
			result.Error = fmt.Errorf("cannot remove '%s': Is a directory", path)
			return result, result.Error
		}

		var removeErr error
		if recursive {
			removeErr = os.RemoveAll(path)
		} else {
			removeErr = os.Remove(path)
		}

		if removeErr != nil && !force {
			result.Error = removeErr
			return result, removeErr
		}
	}

	return result, nil
}

// CpCommand 实现了 'cp' 命令。
// 用于复制文件。
type CpCommand struct{}

// Execute 执行 cp 命令。
// 参数：
//   - 需要源文件和目标文件路径
//   - 不支持目录复制
//
// 限制：
//   - 只支持单个文件复制
//   - 目标文件已存在时会被覆盖
func (c *CpCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "cp",
		StartTime:   ctx.StartTime,
	}

	if len(ctx.Args) < 2 {
		result.Error = fmt.Errorf("cp requires source and destination arguments")
		return result, result.Error
	}

	src := ctx.Args[0]
	dst := ctx.Args[1]

	if !filepath.IsAbs(src) {
		src = filepath.Join(ctx.Options.WorkDir, src)
	}
	if !filepath.IsAbs(dst) {
		dst = filepath.Join(ctx.Options.WorkDir, dst)
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		result.Error = err
		return result, err
	}

	if srcInfo.IsDir() {
		result.Error = fmt.Errorf("cp: omitting directory '%s'", src)
		return result, result.Error
	}

	srcFile, err := os.Open(src)
	if err != nil {
		result.Error = err
		return result, err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		result.Error = err
		return result, err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		result.Error = err
		return result, err
	}

	return result, nil
}

// PWDCommand 实现了 'pwd' 命令。
// 用于显示当前工作目录。
type PWDCommand struct{}

// Execute 执行 pwd 命令。
// 输出：
//   - 如果设置了工作目录，显示设置的目录
//   - 否则显示进程当前工作目录
//
// 错误处理：
//   - 如果无法获取工作目录，返回错误
func (c *PWDCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "pwd",
		StartTime:   ctx.StartTime,
	}

	wd := ctx.Options.WorkDir
	if wd == "" {
		var err error
		wd, err = os.Getwd()
		if err != nil {
			result.Error = err
			return result, err
		}
	}

	fmt.Fprintln(ctx.Options.Stdout, wd)
	return result, nil
}
