// Package commands 实现了 RunShell 的内置命令。
// 本文件实现了一系列实用工具命令，包括文件操作、文本处理等功能。
package commands

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/iamlongalong/runshell/pkg/types"
)

// TouchCommand 实现了 'touch' 命令。
// 用于创建新文件或更新文件的访问和修改时间。
type TouchCommand struct{}

// Execute 执行 touch 命令。
// 功能：
//   - 如果文件不存在，创建空文件
//   - 如果文件存在，更新其访问和修改时间
//
// 参数：
//   - 一个或多个文件路径
//
// 权限：
//   - 新创建的文件权限为 0644
func (c *TouchCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "touch",
		StartTime:   ctx.StartTime,
	}

	if len(ctx.Args) == 0 {
		result.Error = fmt.Errorf("touch: missing file operand")
		return result, result.Error
	}

	for _, path := range ctx.Args {
		if !filepath.IsAbs(path) {
			path = filepath.Join(ctx.Options.WorkDir, path)
		}

		// 如果文件不存在，创建空文件
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			result.Error = err
			return result, err
		}
		f.Close()

		// 更新文件的访问和修改时间
		now := time.Now()
		if err := os.Chtimes(path, now, now); err != nil {
			result.Error = err
			return result, err
		}
	}

	return result, nil
}

// WriteCommand 实现了 'write' 命令。
// 用于将指定内容写入文件。
type WriteCommand struct{}

// Execute 执行 write 命令。
// 功能：
//   - 将指定内容写入文件
//   - 如果文件不存在则创建
//   - 如果文件存在则覆盖
//
// 参数：
//   - 第一个参数为文件路径
//   - 后续参数作为文件内容（以空格连接）
//
// 权限：
//   - 新创建的文件权限为 0644
func (c *WriteCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "write",
		StartTime:   ctx.StartTime,
	}

	if len(ctx.Args) < 2 {
		result.Error = fmt.Errorf("write: missing file operand or content")
		return result, result.Error
	}

	path := ctx.Args[0]
	if !filepath.IsAbs(path) {
		path = filepath.Join(ctx.Options.WorkDir, path)
	}

	content := strings.Join(ctx.Args[1:], " ")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		result.Error = err
		return result, err
	}

	return result, nil
}

// FindCommand 实现了 'find' 命令。
// 用于在目录层次结构中搜索文件。
type FindCommand struct{}

// Execute 执行 find 命令。
// 功能：
//   - 递归搜索指定目录
//   - 支持按文件名模式匹配
//   - 输出匹配文件的完整路径
//
// 参数：
//   - 第一个参数为起始搜索路径
//   - 第二个参数（可选）为搜索模式
func (c *FindCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "find",
		StartTime:   ctx.StartTime,
	}

	if len(ctx.Args) == 0 {
		result.Error = fmt.Errorf("find: missing path operand")
		return result, result.Error
	}

	// 构建find命令的参数
	args := []string{"find", ctx.Args[0]}
	if len(ctx.Args) > 1 {
		args = append(args, "-name", ctx.Args[1])
	}

	// 创建新的执行上下文
	newCtx := &types.ExecuteContext{
		Context:     ctx.Context,
		Args:        args,
		Options:     ctx.Options,
		StartTime:   ctx.StartTime,
		IsPiped:     ctx.IsPiped,
		PipeInput:   ctx.PipeInput,
		PipeOutput:  ctx.PipeOutput,
		PipeContext: ctx.PipeContext,
		Executor:    ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(newCtx)
}

// GrepCommand 实现了 'grep' 命令。
// 用于在文件中搜索匹配指定模式的行。
type GrepCommand struct{}

// Execute 执行 grep 命令。
// 功能：
//   - 支持正则表达式搜索
//   - 显示匹配行的文件名和行号
//   - 可同时搜索多个文件
//
// 参数：
//   - 第一个参数为搜索模式（正则表达式）
//   - 后续参数为要搜索的文件路径
func (c *GrepCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "grep",
		StartTime:   ctx.StartTime,
	}

	if len(ctx.Args) < 2 {
		result.Error = fmt.Errorf("grep: missing pattern or file operand")
		return result, result.Error
	}

	// 构建grep命令的参数
	args := []string{"grep", "-n", ctx.Args[0]}
	args = append(args, ctx.Args[1:]...)

	// 创建新的执行上下文
	newCtx := &types.ExecuteContext{
		Context:     ctx.Context,
		Args:        args,
		Options:     ctx.Options,
		StartTime:   ctx.StartTime,
		IsPiped:     ctx.IsPiped,
		PipeInput:   ctx.PipeInput,
		PipeOutput:  ctx.PipeOutput,
		PipeContext: ctx.PipeContext,
		Executor:    ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(newCtx)
}

// TailCommand 实现了 'tail' 命令。
// 用于显示文件的最后几行。
type TailCommand struct{}

// Execute 执行 tail 命令。
// 功能：
//   - 默认显示文件最后10行
//   - 支持通过 -n 参数指定行数
//   - 按行读取和显示文件内容
//
// 参数：
//   - -n<num>：显示的行数（可选，默认10）
//   - 文件路径
func (c *TailCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "tail",
		StartTime:   ctx.StartTime,
	}

	if len(ctx.Args) < 1 {
		result.Error = fmt.Errorf("tail: missing file operand")
		return result, result.Error
	}

	// 构建tail命令的参数
	args := []string{"tail"}
	if len(ctx.Args) > 1 && strings.HasPrefix(ctx.Args[0], "-n") {
		args = append(args, ctx.Args[0])
		args = append(args, ctx.Args[1])
	} else {
		args = append(args, "-n", "10", ctx.Args[0])
	}

	// 创建新的执行上下文
	newCtx := &types.ExecuteContext{
		Context:     ctx.Context,
		Args:        args,
		Options:     ctx.Options,
		StartTime:   ctx.StartTime,
		IsPiped:     ctx.IsPiped,
		PipeInput:   ctx.PipeInput,
		PipeOutput:  ctx.PipeOutput,
		PipeContext: ctx.PipeContext,
		Executor:    ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(newCtx)
}

// XargsCommand 实现了 'xargs' 命令。
// 用于从标准输入构建和执行命令。
type XargsCommand struct{}

// Execute 执行 xargs 命令。
// 功能：
//   - 从标准输入读取参数
//   - 将参数传递给指定命令执行
//   - 支持多行输入
//
// 参数：
//   - 第一个参数为要执行的命令
//   - 后续参数作为命令的固定参数
func (c *XargsCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "xargs",
		StartTime:   ctx.StartTime,
	}

	if len(ctx.Args) == 0 {
		result.Error = fmt.Errorf("xargs: missing command")
		return result, result.Error
	}

	cmdName := ctx.Args[0]
	cmdArgs := ctx.Args[1:]

	scanner := bufio.NewScanner(ctx.Input)
	for scanner.Scan() {
		args := append(cmdArgs, scanner.Text())
		cmd := exec.Command(cmdName, args...)
		cmd.Stdout = ctx.Options.Stdout
		cmd.Stderr = ctx.Options.Stderr

		if err := cmd.Run(); err != nil {
			result.Error = err
			return result, err
		}
	}

	if err := scanner.Err(); err != nil {
		result.Error = err
		return result, err
	}

	return result, nil
}

// SeedCommand implements the 'seed' command for random number generation
type SeedCommand struct{}

func (c *SeedCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "seed",
		StartTime:   ctx.StartTime,
	}

	if len(ctx.Args) < 2 {
		result.Error = fmt.Errorf("seed: usage: seed min max [count]")
		return result, result.Error
	}

	var min, max, count int
	fmt.Sscanf(ctx.Args[0], "%d", &min)
	fmt.Sscanf(ctx.Args[1], "%d", &max)

	if len(ctx.Args) > 2 {
		fmt.Sscanf(ctx.Args[2], "%d", &count)
	} else {
		count = 1
	}

	if max <= min {
		result.Error = fmt.Errorf("seed: max must be greater than min")
		return result, result.Error
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < count; i++ {
		num := r.Intn(max-min+1) + min
		fmt.Fprintln(ctx.Options.Stdout, num)
	}

	return result, nil
}

// MvCommand implements mv command
type MvCommand struct{}

// Execute 执行 mv 命令。
func (c *MvCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Args) != 2 {
		return nil, fmt.Errorf("mv requires source and destination arguments")
	}

	// 构建mv命令的参数
	args := []string{"mv"}
	args = append(args, ctx.Args...)

	// 创建新的执行上下文
	newCtx := &types.ExecuteContext{
		Context:     ctx.Context,
		Args:        args,
		Options:     ctx.Options,
		StartTime:   ctx.StartTime,
		IsPiped:     ctx.IsPiped,
		PipeInput:   ctx.PipeInput,
		PipeOutput:  ctx.PipeOutput,
		PipeContext: ctx.PipeContext,
		Executor:    ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(newCtx)
}

// HeadCommand implements head command
type HeadCommand struct{}

func (c *HeadCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Args) < 1 {
		return nil, fmt.Errorf("head requires a file argument")
	}

	// 构建head命令的参数
	args := []string{"head"}
	if len(ctx.Args) > 2 && ctx.Args[0] == "-n" {
		args = append(args, "-n", ctx.Args[1], ctx.Args[2])
	} else {
		args = append(args, "-n", "10", ctx.Args[0])
	}

	// 创建新的执行上下文
	newCtx := &types.ExecuteContext{
		Context:     ctx.Context,
		Args:        args,
		Options:     ctx.Options,
		StartTime:   ctx.StartTime,
		IsPiped:     ctx.IsPiped,
		PipeInput:   ctx.PipeInput,
		PipeOutput:  ctx.PipeOutput,
		PipeContext: ctx.PipeContext,
		Executor:    ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(newCtx)
}

// SortCommand implements sort command
type SortCommand struct{}

func (c *SortCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Args) < 1 {
		return nil, fmt.Errorf("sort requires a file argument")
	}

	// 构建sort命令的参数
	args := []string{"sort"}
	args = append(args, ctx.Args...)

	// 创建新的执行上下文
	newCtx := &types.ExecuteContext{
		Context:     ctx.Context,
		Args:        args,
		Options:     ctx.Options,
		StartTime:   ctx.StartTime,
		IsPiped:     ctx.IsPiped,
		PipeInput:   ctx.PipeInput,
		PipeOutput:  ctx.PipeOutput,
		PipeContext: ctx.PipeContext,
		Executor:    ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(newCtx)
}

// UniqCommand implements uniq command
type UniqCommand struct{}

func (c *UniqCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Args) < 1 {
		return nil, fmt.Errorf("uniq requires a file argument")
	}

	file, err := os.Open(ctx.Args[0])
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var output strings.Builder
	scanner := bufio.NewScanner(file)
	var lastLine string
	first := true

	for scanner.Scan() {
		currentLine := scanner.Text()
		if first || currentLine != lastLine {
			output.WriteString(currentLine + "\n")
			lastLine = currentLine
			first = false
		}
	}

	return &types.ExecuteResult{
		CommandName: "uniq",
		ExitCode:    0,
		StartTime:   ctx.StartTime,
		EndTime:     time.Now(),
		Output:      output.String(),
	}, nil
}

// NetstatCommand implements netstat command
type NetstatCommand struct{}

func (c *NetstatCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	conns, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var output strings.Builder
	output.WriteString("Interface\tStatus\tAddresses\n")

	for _, conn := range conns {
		addrs, err := conn.Addrs()
		if err != nil {
			continue
		}

		addrStrings := make([]string, len(addrs))
		for i, addr := range addrs {
			addrStrings[i] = addr.String()
		}

		output.WriteString(fmt.Sprintf("%s\t%v\t%s\n",
			conn.Name,
			conn.Flags,
			strings.Join(addrStrings, ", ")))
	}

	return &types.ExecuteResult{
		CommandName: "netstat",
		ExitCode:    0,
		StartTime:   ctx.StartTime,
		EndTime:     time.Now(),
		Output:      output.String(),
	}, nil
}

// IfconfigCommand implements ifconfig command
type IfconfigCommand struct{}

func (c *IfconfigCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var output strings.Builder
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		output.WriteString(fmt.Sprintf("%s: flags=%d<", iface.Name, iface.Flags))
		if iface.Flags&net.FlagUp != 0 {
			output.WriteString("UP,")
		}
		if iface.Flags&net.FlagBroadcast != 0 {
			output.WriteString("BROADCAST,")
		}
		if iface.Flags&net.FlagLoopback != 0 {
			output.WriteString("LOOPBACK,")
		}
		if iface.Flags&net.FlagPointToPoint != 0 {
			output.WriteString("POINTTOPOINT,")
		}
		if iface.Flags&net.FlagMulticast != 0 {
			output.WriteString("MULTICAST,")
		}
		output.WriteString(">\n")

		for _, addr := range addrs {
			output.WriteString(fmt.Sprintf("\tinet %s\n", addr.String()))
		}
		output.WriteString("\n")
	}

	return &types.ExecuteResult{
		CommandName: "ifconfig",
		ExitCode:    0,
		StartTime:   ctx.StartTime,
		EndTime:     time.Now(),
		Output:      output.String(),
	}, nil
}

// CurlCommand implements curl command
type CurlCommand struct{}

// Execute 执行 curl 命令。
func (c *CurlCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Args) < 1 {
		return nil, fmt.Errorf("curl requires a URL argument")
	}

	// 构建curl命令的参数
	args := []string{"curl"}
	args = append(args, ctx.Args...)

	// 创建新的执行上下文
	newCtx := &types.ExecuteContext{
		Context:     ctx.Context,
		Args:        args,
		Options:     ctx.Options,
		StartTime:   ctx.StartTime,
		IsPiped:     ctx.IsPiped,
		PipeInput:   ctx.PipeInput,
		PipeOutput:  ctx.PipeOutput,
		PipeContext: ctx.PipeContext,
		Executor:    ctx.Executor,
	}

	// 通过executor执行命令
	return ctx.Executor.Execute(newCtx)
}

// SedCommand implements sed command
type SedCommand struct{}

func (c *SedCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Args) < 2 {
		return nil, fmt.Errorf("sed requires pattern and file arguments")
	}

	pattern := ctx.Args[0]
	filename := ctx.Args[1]

	// 使用系统的sed命令
	cmd := exec.Command("sed", pattern, filename)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("sed error: %v", err)
	}

	return &types.ExecuteResult{
		CommandName: "sed",
		ExitCode:    0,
		StartTime:   ctx.StartTime,
		EndTime:     time.Now(),
		Output:      string(output),
	}, nil
}

// PipeCommand implements command piping
type PipeCommand struct {
	Left  types.CommandHandler
	Right types.CommandHandler
}

func (c *PipeCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	// 执行左侧命令
	leftResult, err := c.Left.Execute(ctx)
	if err != nil {
		return nil, err
	}

	// 创建新的上下文，将左侧命令的输出作为右侧命令的输入
	rightCtx := &types.ExecuteContext{
		Context:   ctx.Context,
		Args:      ctx.Args,
		Options:   ctx.Options,
		StartTime: time.Now(),
		Input:     strings.NewReader(leftResult.Output),
		Executor:  ctx.Executor,
	}

	// 执行右侧命令
	rightResult, err := c.Right.Execute(rightCtx)
	if err != nil {
		return nil, err
	}

	return rightResult, nil
}

// RedirectCommand implements output redirection
type RedirectCommand struct {
	Command    types.CommandHandler
	OutputFile string
	Append     bool
}

func (c *RedirectCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	// 执行命令
	result, err := c.Command.Execute(ctx)
	if err != nil {
		return nil, err
	}

	// 打开输出文件
	flag := os.O_WRONLY | os.O_CREATE
	if c.Append {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}

	file, err := os.OpenFile(c.OutputFile, flag, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 写入输出
	if _, err := file.WriteString(result.Output); err != nil {
		return nil, err
	}

	return result, nil
}
