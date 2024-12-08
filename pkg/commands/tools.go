package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/iamlongalong/runshell/pkg/types"
)

// WgetCommand 实现 wget 命令。
type WgetCommand struct{}

// Execute 执行 wget 命令。
func (c *WgetCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Args) < 1 {
		return nil, fmt.Errorf("wget requires a URL")
	}

	url := ctx.Args[0]
	output := ""
	if len(ctx.Args) > 1 {
		output = ctx.Args[1]
	} else {
		output = filepath.Base(url)
	}

	if !filepath.IsAbs(output) {
		output = filepath.Join(ctx.Options.WorkDir, output)
	}

	// 创建 HTTP 客户端
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx.Context, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 创建输出文件
	file, err := os.Create(output)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 复制内容
	startTime := types.GetTimeNow()
	_, err = io.Copy(file, resp.Body)
	endTime := types.GetTimeNow()

	if err != nil {
		return &types.ExecuteResult{
			CommandName: "wget",
			ExitCode:    1,
			StartTime:   startTime,
			EndTime:     endTime,
			Error:       err,
		}, err
	}

	return &types.ExecuteResult{
		CommandName: "wget",
		ExitCode:    0,
		StartTime:   startTime,
		EndTime:     endTime,
	}, nil
}

// TarCommand 实现 tar 命令。
type TarCommand struct{}

// Execute 执行 tar 命令。
func (c *TarCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Args) < 2 {
		return nil, fmt.Errorf("tar requires operation and file arguments")
	}

	// 准备命令
	cmd := exec.CommandContext(ctx.Context, "tar", ctx.Args...)
	cmd.Dir = ctx.Options.WorkDir

	// 设置环境变量
	if len(ctx.Options.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range ctx.Options.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// 设置输入输出
	cmd.Stdin = ctx.Options.Stdin
	cmd.Stdout = ctx.Options.Stdout
	cmd.Stderr = ctx.Options.Stderr

	// 执行命令
	startTime := types.GetTimeNow()
	err := cmd.Run()
	endTime := types.GetTimeNow()

	result := &types.ExecuteResult{
		CommandName: "tar",
		StartTime:   startTime,
		EndTime:     endTime,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		result.Error = err
		return result, err
	}

	return result, nil
}

// ZipCommand 实现 zip 命令。
type ZipCommand struct{}

// Execute 执行 zip 命令。
func (c *ZipCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Args) < 2 {
		return nil, fmt.Errorf("zip requires archive and file arguments")
	}

	// 准备命令
	cmd := exec.CommandContext(ctx.Context, "zip", ctx.Args...)
	cmd.Dir = ctx.Options.WorkDir

	// 设置环境变量
	if len(ctx.Options.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range ctx.Options.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// 设置输入输出
	cmd.Stdin = ctx.Options.Stdin
	cmd.Stdout = ctx.Options.Stdout
	cmd.Stderr = ctx.Options.Stderr

	// 执行命令
	startTime := types.GetTimeNow()
	err := cmd.Run()
	endTime := types.GetTimeNow()

	result := &types.ExecuteResult{
		CommandName: "zip",
		StartTime:   startTime,
		EndTime:     endTime,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		result.Error = err
		return result, err
	}

	return result, nil
}

// PythonCommand 实现 python 命令。
type PythonCommand struct{}

// Execute 执行 python 命令。
func (c *PythonCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	// 检查 python 是否安装
	if _, err := exec.LookPath("python3"); err != nil {
		if _, err := exec.LookPath("python"); err != nil {
			return nil, fmt.Errorf("python is not installed")
		}
	}

	// 准备命令（优先使用 python3）
	pythonCmd := "python3"
	if _, err := exec.LookPath("python3"); err != nil {
		pythonCmd = "python"
	}

	cmd := exec.CommandContext(ctx.Context, pythonCmd, ctx.Args...)
	cmd.Dir = ctx.Options.WorkDir

	// 设置环境变量
	cmd.Env = os.Environ()
	// 添加用户指定的环境变量
	for k, v := range ctx.Options.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// 设置输入输出
	cmd.Stdin = ctx.Options.Stdin
	cmd.Stdout = ctx.Options.Stdout
	cmd.Stderr = ctx.Options.Stderr

	// 执行命令
	startTime := types.GetTimeNow()
	err := cmd.Run()
	endTime := types.GetTimeNow()

	result := &types.ExecuteResult{
		CommandName: pythonCmd,
		StartTime:   startTime,
		EndTime:     endTime,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		result.Error = err
		return result, err
	}

	return result, nil
}

// PipCommand 实现 pip 命令。
type PipCommand struct{}

// Execute 执行 pip 命令。
func (c *PipCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	// 检查 pip 是否安装
	if _, err := exec.LookPath("pip3"); err != nil {
		if _, err := exec.LookPath("pip"); err != nil {
			return nil, fmt.Errorf("pip is not installed")
		}
	}

	// 准备命令（优先使用 pip3）
	pipCmd := "pip3"
	if _, err := exec.LookPath("pip3"); err != nil {
		pipCmd = "pip"
	}

	cmd := exec.CommandContext(ctx.Context, pipCmd, ctx.Args...)
	cmd.Dir = ctx.Options.WorkDir

	// 设置环境变量
	cmd.Env = os.Environ()
	// 添加用户指定的环境变量
	for k, v := range ctx.Options.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// 设置输入输出
	cmd.Stdin = ctx.Options.Stdin
	cmd.Stdout = ctx.Options.Stdout
	cmd.Stderr = ctx.Options.Stderr

	// 执行命令
	startTime := types.GetTimeNow()
	err := cmd.Run()
	endTime := types.GetTimeNow()

	result := &types.ExecuteResult{
		CommandName: pipCmd,
		StartTime:   startTime,
		EndTime:     endTime,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		result.Error = err
		return result, err
	}

	return result, nil
}

// DockerCommand 实现 docker 命令。
type DockerCommand struct{}

// Execute 执行 docker 命令。
func (c *DockerCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	// 检查 docker 是否安装
	if _, err := exec.LookPath("docker"); err != nil {
		return nil, fmt.Errorf("docker is not installed")
	}

	// 准备命令
	cmd := exec.CommandContext(ctx.Context, "docker", ctx.Args...)
	cmd.Dir = ctx.Options.WorkDir

	// 设置环境变量
	cmd.Env = os.Environ()
	// 添加用户指定的环境变量
	for k, v := range ctx.Options.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// 设置输入输出
	cmd.Stdin = ctx.Options.Stdin
	cmd.Stdout = ctx.Options.Stdout
	cmd.Stderr = ctx.Options.Stderr

	// 执行命令
	startTime := types.GetTimeNow()
	err := cmd.Run()
	endTime := types.GetTimeNow()

	result := &types.ExecuteResult{
		CommandName: "docker",
		StartTime:   startTime,
		EndTime:     endTime,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		result.Error = err
		return result, err
	}

	return result, nil
}

// NodeCommand 实现 node 命令。
type NodeCommand struct{}

// Execute 执行 node 命令。
func (c *NodeCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	// 检查 node 是否安装
	if _, err := exec.LookPath("node"); err != nil {
		return nil, fmt.Errorf("node is not installed")
	}

	// 准备命令
	cmd := exec.CommandContext(ctx.Context, "node", ctx.Args...)
	cmd.Dir = ctx.Options.WorkDir

	// 设置环境变量
	cmd.Env = os.Environ()
	// 添加用户指定的环境变量
	for k, v := range ctx.Options.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// 设置输入输出
	cmd.Stdin = ctx.Options.Stdin
	cmd.Stdout = ctx.Options.Stdout
	cmd.Stderr = ctx.Options.Stderr

	// 执行命令
	startTime := types.GetTimeNow()
	err := cmd.Run()
	endTime := types.GetTimeNow()

	result := &types.ExecuteResult{
		CommandName: "node",
		StartTime:   startTime,
		EndTime:     endTime,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		result.Error = err
		return result, err
	}

	return result, nil
}

// NPMCommand 实现 npm 命令。
type NPMCommand struct{}

// Execute 执行 npm 命令。
func (c *NPMCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	// 检查 npm 是否安装
	if _, err := exec.LookPath("npm"); err != nil {
		return nil, fmt.Errorf("npm is not installed")
	}

	// 准备命令
	cmd := exec.CommandContext(ctx.Context, "npm", ctx.Args...)
	cmd.Dir = ctx.Options.WorkDir

	// 设置环境变量
	cmd.Env = os.Environ()
	// 添加用户指定的环境变量
	for k, v := range ctx.Options.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// 设置输入输出
	cmd.Stdin = ctx.Options.Stdin
	cmd.Stdout = ctx.Options.Stdout
	cmd.Stderr = ctx.Options.Stderr

	// 执行命令
	startTime := types.GetTimeNow()
	err := cmd.Run()
	endTime := types.GetTimeNow()

	result := &types.ExecuteResult{
		CommandName: "npm",
		StartTime:   startTime,
		EndTime:     endTime,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		result.Error = err
		return result, err
	}

	return result, nil
}
