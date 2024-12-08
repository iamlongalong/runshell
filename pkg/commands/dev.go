package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/iamlongalong/runshell/pkg/types"
)

// GitCommand 实现 git 命令。
type GitCommand struct{}

// Execute 执行 git 命令。
func (c *GitCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	// 检查 git 是否安装
	if _, err := exec.LookPath("git"); err != nil {
		return nil, fmt.Errorf("git is not installed: %v", err)
	}

	// 准备命令
	cmd := exec.CommandContext(ctx.Context, "git", ctx.Args...)

	// 设置工作目录
	if ctx.Options.WorkDir != "" {
		// 验证工作目录是否是 git 仓库
		if _, err := os.Stat(filepath.Join(ctx.Options.WorkDir, ".git")); err != nil {
			return nil, fmt.Errorf("not a git repository: %v", err)
		}
		cmd.Dir = ctx.Options.WorkDir
	}

	// 设置环境变量
	if len(ctx.Options.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range ctx.Options.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// 设置输入输出
	if ctx.Options.Stdin != nil {
		cmd.Stdin = ctx.Options.Stdin
	}
	if ctx.Options.Stdout != nil {
		cmd.Stdout = ctx.Options.Stdout
	}
	if ctx.Options.Stderr != nil {
		cmd.Stderr = ctx.Options.Stderr
	}

	// 执行命令
	startTime := types.GetTimeNow()
	err := cmd.Run()
	endTime := types.GetTimeNow()

	// 准备结果
	result := &types.ExecuteResult{
		CommandName: "git",
		StartTime:   startTime,
		EndTime:     endTime,
	}

	// 处理错误
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		result.Error = err
		return result, fmt.Errorf("git command failed: %v", err)
	}

	result.ExitCode = 0
	return result, nil
}

// GoCommand 实现 go 命令。
type GoCommand struct{}

// Execute 执行 go 命令。
func (c *GoCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	// 检查 go 是否安装
	if _, err := exec.LookPath("go"); err != nil {
		return nil, fmt.Errorf("go is not installed: %v", err)
	}

	// 准备命令
	cmd := exec.CommandContext(ctx.Context, "go", ctx.Args...)

	// 设置工作目录
	if ctx.Options.WorkDir != "" {
		// 验证工作目录是否包含 go.mod
		if _, err := os.Stat(filepath.Join(ctx.Options.WorkDir, "go.mod")); err != nil {
			return nil, fmt.Errorf("not a Go module: %v", err)
		}
		cmd.Dir = ctx.Options.WorkDir
	}

	// 设置环境变量
	cmd.Env = os.Environ()
	// 添加 GOPROXY 环境变量（如果没有设置）
	if !hasEnv(cmd.Env, "GOPROXY") {
		cmd.Env = append(cmd.Env, "GOPROXY=https://goproxy.cn,direct")
	}
	// 添加用户指定的环境变量
	for k, v := range ctx.Options.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// 设置输入输出
	if ctx.Options.Stdin != nil {
		cmd.Stdin = ctx.Options.Stdin
	}
	if ctx.Options.Stdout != nil {
		cmd.Stdout = ctx.Options.Stdout
	}
	if ctx.Options.Stderr != nil {
		cmd.Stderr = ctx.Options.Stderr
	}

	// 执行命令
	startTime := types.GetTimeNow()
	err := cmd.Run()
	endTime := types.GetTimeNow()

	// 准备结果
	result := &types.ExecuteResult{
		CommandName: "go",
		StartTime:   startTime,
		EndTime:     endTime,
	}

	// 处理错误
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		result.Error = err
		return result, fmt.Errorf("go command failed: %v", err)
	}

	result.ExitCode = 0
	return result, nil
}

// hasEnv 检查环境变量列表中是否包含指定的变量。
func hasEnv(env []string, key string) bool {
	prefix := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return true
		}
	}
	return false
}
