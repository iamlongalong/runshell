package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
)

func runPipeline(pipeExec *executor.PipelineExecutor, cmdStr string, env map[string]string) error {
	fmt.Printf("\nExecuting: %s\n", cmdStr)

	pipeline, err := pipeExec.ParsePipeline(cmdStr)
	if err != nil {
		return fmt.Errorf("failed to parse pipeline: %v", err)
	}

	pipeline.Context = context.Background()
	pipeline.Options = &types.ExecuteOptions{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Env:    env,
	}

	result, err := pipeExec.ExecutePipeline(pipeline)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %v", err)
	}

	fmt.Printf("Pipeline exit code: %d\n", result.ExitCode)
	return nil
}

func main() {
	// 创建本地执行器
	localExec := executor.NewLocalExecutor()

	// 创建管道执行器
	pipeExec := executor.NewPipelineExecutor(localExec)

	// 示例 1: 简单的管道命令
	fmt.Println("\n=== Example 1: Simple Pipeline ===")
	err := runPipeline(pipeExec, "echo hello world | grep world", nil)
	if err != nil {
		log.Printf("Example 1 failed: %v\n", err)
	}

	// 示例 2: 多重管道命令
	fmt.Println("\n=== Example 2: Multiple Pipes ===")
	err = runPipeline(pipeExec, "ls -la | grep go | wc -l", nil)
	if err != nil {
		log.Printf("Example 2 failed: %v\n", err)
	}

	// 示例 3: 带环境变量的管道命令
	fmt.Println("\n=== Example 3: Pipeline with Environment Variables ===")
	err = runPipeline(pipeExec, "env | grep PATH", map[string]string{
		"CUSTOM_PATH": "/custom/path",
	})
	if err != nil {
		log.Printf("Example 3 failed: %v\n", err)
	}

	// 示例 4: 复杂的管道命令
	fmt.Println("\n=== Example 4: Complex Pipeline ===")
	err = runPipeline(pipeExec, "ps aux | grep go | sort -k2 | head -n 3", nil)
	if err != nil {
		log.Printf("Example 4 failed: %v\n", err)
	}

	// 示例 5: 文件处理管道命令
	fmt.Println("\n=== Example 5: File Processing Pipeline ===")
	err = runPipeline(pipeExec, "cat ../go.mod | grep require | sort | uniq -c", nil)
	if err != nil {
		log.Printf("Example 5 failed: %v\n", err)
	}
}
