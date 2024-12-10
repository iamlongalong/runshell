// Package main demonstrates how to use Docker executor with pipeline functionality.
//
// This example shows:
// 1. How to execute single commands in a Docker container
// 2. How to execute pipeline commands (commands connected with pipes)
// 3. How to capture and handle command output
// 4. How to handle errors and exit codes
//
// The example uses busybox:latest as the base image because it's small and contains
// common Unix utilities needed for the demonstration.
package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
)

func main() {
	// 创建 Docker 执行器
	// 使用 busybox:latest 作为基础镜像，它体积小但包含常用的 Unix 工具
	exec, err := executor.NewDockerExecutor(types.DockerConfig{
		Image:                     "ubuntu:latest",
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{})
	if err != nil {
		fmt.Printf("Failed to create Docker executor: %v\n", err)
		os.Exit(1)
	}
	// 确保在程序退出时清理 Docker 资源
	defer exec.Close()

	// 创建管道执行器
	// PipelineExecutor 封装了管道命令的解析和执行逻辑
	pipeExec := executor.NewPipelineExecutor(exec)

	// 示例1：执行单个命令
	// 演示如何在 Docker 容器中执行单个命令并捕获输出
	fmt.Println("=== Testing single command ===")
	var output1 bytes.Buffer
	ctx1 := &types.ExecuteContext{
		Context: context.Background(),                                // 使用默认上下文，实际应用中可能需要可取消的上下文
		Command: types.Command{Command: "ls", Args: []string{"-al"}}, // 列出当前目录的详细信息
		Options: &types.ExecuteOptions{
			Stdout: &output1,  // 捕获标准输出到 buffer
			Stderr: os.Stderr, // 错误输出直接显示到终端
		},
	}
	_, err = exec.Execute(ctx1)
	if err != nil {
		fmt.Printf("Single command failed: %v\n", err)
		os.Exit(1)
	}

	// 显示命令执行结果
	fmt.Printf("ls -al output size: %d bytes\n", output1.Len())
	fmt.Println("ls -al output:")
	fmt.Println(output1.String())

	// 示例2：执行管道命令
	// 演示如何执行包含管道的复杂命令
	fmt.Println("\n=== Testing pipeline command ===")

	// 解析管道命令
	// ParsePipeline 会将命令字符串解析成一系列要按顺序执行的命令
	pipeline, err := pipeExec.ParsePipeline("ls -al | grep s")
	if err != nil {
		fmt.Printf("Failed to parse pipeline: %v\n", err)
		os.Exit(1)
	}

	// 设置管道命令的执行选项
	var pipeOutput bytes.Buffer
	pipeline.Context = context.Background()
	pipeline.Options = &types.ExecuteOptions{
		Stdout: &pipeOutput, // 捕获整个管道的最终输出
		Stderr: os.Stderr,   // 错误输出到终端
	}

	// 执行管道命令
	// ExecutePipeline 会确保管道中的命令按顺序执行，并正确处理它们之间的数据流
	result, err := pipeExec.ExecutePipeline(pipeline)
	if err != nil {
		fmt.Printf("Failed to execute pipeline: %v\n", err)
		os.Exit(1)
	}

	// 显示管道命令的执行结果
	fmt.Printf("Pipeline output size: %d bytes\n", pipeOutput.Len())
	fmt.Println("Pipeline output:")
	fmt.Println(pipeOutput.String())
	fmt.Printf("Pipeline completed with exit code: %d\n", result.ExitCode)
}

// 使用示例：
//
// 1. 确保已安装 Docker 并且 Docker daemon 正在运行
// 2. 确保有权限访问 Docker daemon
// 3. 运行示例：
//    go run main.go
//
// 预期输出：
// - 首先显示单个 ls -al 命令的完整输出
// - 然后显示 ls -al | grep s 管道命令的过滤后输出
//
// 注意事项：
// - 示例使用 busybox:latest 镜像，首次运行时会自动��载
// - 所有命令在容器内执行，不会影响主机系统
// - 程序结束时会自动清理创建的容器
