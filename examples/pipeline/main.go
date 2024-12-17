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
func runPipeline(ctx context.Context, pipeExec *executor.PipelineExecutor, cmdStr string, env map[string]string) error {
	fmt.Printf("\nExecuting: %s\n", cmdStr)

	result, err := pipeExec.Execute(&types.ExecuteContext{
		Context: ctx,
		Command: types.Command{
			Command: cmdStr,
		},
		Options: &types.ExecuteOptions{
			Env: env,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %v", err)
	}

	fmt.Printf("Pipeline Output:\n%s\n", result.Output)

	fmt.Printf("Pipeline exit code: %d\n", result.ExitCode)
	return nil
}

func main() {
	// 创建本地执行器
	localExec := executor.NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{}, nil)

	ctx := context.Background()

	// 创建管道执行器
	pipeExec := executor.NewPipelineExecutor(localExec)

	// 示例 1: 简单的管道命令
	fmt.Println("\n=== Example 1: Simple Pipeline ===")
	err := runPipeline(ctx, pipeExec, "ls -la | grep go", nil)
	if err != nil {
		log.Printf("Example 1 failed: %v\n", err)
	}

	// 示例 2: 多重管道命令
	fmt.Println("\n=== Example 2: Multiple Pipes ===")
	err = runPipeline(ctx, pipeExec, "ls -la | grep go | wc -l", nil)
	if err != nil {
		log.Printf("Example 2 failed: %v\n", err)
	}

	// 示例 3: 带环境变量的管道命令
	fmt.Println("\n=== Example 3: Environment Variables ===")
	err = runPipeline(ctx, pipeExec, "pwd | tr '[:lower:]' '[:upper:]'", map[string]string{
		"PWD": "/custom/path",
	})
	if err != nil {
		log.Printf("Example 3 failed: %v\n", err)
	}

	// 示例 4: 文本处理
	fmt.Println("\n=== Example 4: Text Processing ===")
	err = runPipeline(ctx, pipeExec, "echo hello world | tr '[:lower:]' '[:upper:]'", nil)
	if err != nil {
		log.Printf("Example 4 failed: %v\n", err)
	}

	// 示例 5: 多命令管道
	fmt.Println("\n=== Example 5: Multiple Commands ===")
	err = runPipeline(ctx, pipeExec, `echo 'hi\nhello\nworld' | grep o`, nil)
	if err != nil {
		log.Printf("Example 5 failed: %v\n", err)
	}
}
