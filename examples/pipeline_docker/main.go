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
	"context"
	"fmt"
	"os"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
)

func main() {
	// 创建 Docker 执行器
	// 使用 busybox:latest 作为基础镜像，它体积小但包含常用的 Unix 工具
	exec, err := executor.NewDockerExecutorBuilder(types.DockerConfig{
		Image:                     "ubuntu:latest",
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{}).Build()
	if err != nil {
		fmt.Printf("Failed to create Docker executor: %v\n", err)
		os.Exit(1)
	}
	// 确保在程序退出时清理 Docker 资源
	defer exec.Close()

	// 创建管道执行器
	// PipelineExecutor 封装了管道命令的解析和执行逻辑
	pipeExec := executor.NewPipelineExecutor(exec)
	result, err := pipeExec.Execute(&types.ExecuteContext{
		Context: context.Background(),
		Command: types.Command{
			Command: "ls -al | grep s",
		},
	})

	// 处理执行错误
	if err != nil {
		fmt.Printf("Failed to execute pipeline command: %v\n", err)
		os.Exit(1)
	}

	// 打印命令执行结果
	fmt.Printf("Command Output:\n%s\n", result.Output)
	fmt.Printf("Exit Code: %d\n", result.ExitCode)

	// 演示更多管道命令示例
	examples := []string{
		"echo 'Hello World' | grep Hello",
		"cat /etc/passwd | grep root | cut -d: -f1",
		"ps aux | grep bash | wc -l",
	}

	for _, cmd := range examples {
		fmt.Printf("\nExecuting: %s\n", cmd)
		result, err := pipeExec.Execute(&types.ExecuteContext{
			Context: context.Background(),
			Command: types.Command{
				Command: cmd,
			},
		})

		if err != nil {
			fmt.Printf("Failed to execute command: %v\n", err)
			continue
		}

		fmt.Printf("Output:\n%s\n", result.Output)
		fmt.Printf("Exit Code: %d\n", result.ExitCode)
	}
}
