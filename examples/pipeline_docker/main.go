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
	"log"
	"strings"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
)

func main() {
	// 创建 Docker 执行器
	dockerExec, err := executor.NewDockerExecutorBuilder(types.DockerConfig{
		Image:                     "ubuntu:latest",
		WorkDir:                   "/workspace",
		AllowUnregisteredCommands: true,
	}).WithOptions(&types.ExecuteOptions{}).Build()
	if err != nil {
		log.Fatalf("Failed to create Docker executor: %v", err)
	}
	// 确保在程序退出时清理 Docker 资源
	defer dockerExec.Close()

	// 创建管道执行器
	// PipelineExecutor 封装了管道命令的解析和执行逻辑
	pipeExec := executor.NewPipelineExecutor(dockerExec)

	// 演示管道命令示例
	examples := []string{
		"ls -al | grep s",
		"echo 'Hello World' | grep Hello",
		"cat /etc/passwd | grep root | cut -d: -f1",
		"ps aux | grep bash | wc -l",
		"echo 'test' | grep nonexistent", // 这个命令预期会返回非零退出码
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
			// 检查是否是 grep 命令没有找到匹配项
			if result != nil && result.ExitCode == 1 && strings.Contains(cmd, "grep") {
				fmt.Printf("No matches found\n")
				continue
			}
			fmt.Printf("Error executing command: %v\n", err)
			continue
		}

		fmt.Printf("Output:\n%s\n", result.Output)
		fmt.Printf("Exit Code: %d\n", result.ExitCode)
	}
}
