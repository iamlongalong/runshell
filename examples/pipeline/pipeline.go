// Package main demonstrates the pipeline functionality of RunShell.
//
// This example shows how to use the pipeline executor to run command chains,
// similar to Unix-style pipes. It demonstrates various pipeline patterns:
//   - Simple command piping
//   - Multi-stage pipelines
//   - Environment variable handling
//   - Process filtering and sorting
//   - File content processing
//
// The pipeline executor supports:
//   - Standard Unix commands (ls, grep, wc, etc.)
//   - Environment variable propagation
//   - Error handling and status reporting
//   - Input/Output stream management
//
// Example usage:
//
//	pipeExec := executor.NewPipelineExecutor(localExec)
//	err := runPipeline(pipeExec, "ls -la | grep go | wc -l", nil)
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
)

// runPipeline executes a pipeline of commands and handles the results.
//
// Parameters:
//   - pipeExec: The pipeline executor instance
//   - cmdStr: The command string containing one or more piped commands
//   - env: Optional environment variables for the pipeline
//
// The function:
//  1. Parses the command string into a pipeline
//  2. Sets up the execution context and options
//  3. Executes the pipeline and captures the result
//  4. Reports any errors that occur
//
// Example:
//
//	runPipeline(pipeExec, "ls -la | grep go", nil)
//	runPipeline(pipeExec, "cat file.txt | grep pattern | wc -l", env)
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
	localExec := executor.NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{})

	// 创建管道执行器
	pipeExec := executor.NewPipelineExecutor(localExec)

	// 示例 1: 简单的管道命令
	// 列出当前目录内容并过滤包含 "go" 的行
	fmt.Println("\n=== Example 1: Simple Pipeline ===")
	err := runPipeline(pipeExec, "ls -la | grep go", nil)
	if err != nil {
		log.Printf("Example 1 failed: %v\n", err)
	}

	// 示例 2: 多重管道命令
	// 统计当前目录中包含 "go" 的文件数量
	fmt.Println("\n=== Example 2: Multiple Pipes ===")
	err = runPipeline(pipeExec, "ls -la | grep go | wc -l", nil)
	if err != nil {
		log.Printf("Example 2 failed: %v\n", err)
	}

	// 示例 3: 带环境变量的管道命令
	// 显示系统路径信息，并添加自定义路径
	fmt.Println("\n=== Example 3: Pipeline with Environment Variables ===")
	err = runPipeline(pipeExec, "env | grep PATH", map[string]string{
		"CUSTOM_PATH": "/custom/path",
	})
	if err != nil {
		log.Printf("Example 3 failed: %v\n", err)
	}

	// 示例 4: 复杂的管道命令
	// 查找并排序 Go 相关进程，显示前三个
	fmt.Println("\n=== Example 4: Complex Pipeline ===")
	err = runPipeline(pipeExec, "ps aux | grep go | sort -k2 | head -n 3", nil)
	if err != nil {
		log.Printf("Example 4 failed: %v\n", err)
	}

	// 示例 5: 文件处理管道命令
	// 分析 go.mod 文件中的依赖项
	fmt.Println("\n=== Example 5: File Processing Pipeline ===")
	err = runPipeline(pipeExec, "cat ../go.mod | grep require | sort | uniq -c", nil)
	if err != nil {
		log.Printf("Example 5 failed: %v\n", err)
	}
}
