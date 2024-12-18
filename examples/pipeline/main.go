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
	"strings"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
)

// Example pipeline configurations
var examples = []struct {
	name        string
	pipeline    string
	env         map[string]string
	expectedErr string
}{
	{
		name:     "Simple Pipeline",
		pipeline: "ls -la | grep go",
	},
	{
		name:     "Multiple Pipes",
		pipeline: "ls -la | grep go | wc -l",
	},
	{
		name:     "Environment Variables",
		pipeline: "pwd | tr '[:lower:]' '[:upper:]'",
		env: map[string]string{
			"TEST_VAR": "test_value",
		},
	},
	{
		name:     "Text Processing",
		pipeline: "echo hello world | tr '[:lower:]' '[:upper:]'",
	},
	{
		name:     "Multiple Commands",
		pipeline: "echo 'hi\nhello\nworld' | grep o",
	},
}

func runPipeline(ctx context.Context, pipeline string, env map[string]string) error {
	fmt.Printf("\nExecuting: %s\n", pipeline)

	// Create local executor
	execBuilder := executor.NewLocalExecutorBuilder(types.LocalConfig{
		AllowUnregisteredCommands: true,
	})
	exec, err := execBuilder.Build(&types.ExecuteOptions{
		Env: env,
	})
	if err != nil {
		return fmt.Errorf("failed to create executor: %v", err)
	}
	defer exec.Close()

	// Parse pipeline into commands
	commands := make([]*types.Command, 0)
	parts := strings.Split(pipeline, "|")
	for _, part := range parts {
		cmdParts := strings.Fields(strings.TrimSpace(part))
		if len(cmdParts) == 0 {
			continue
		}
		cmd := &types.Command{
			Command: cmdParts[0],
		}
		if len(cmdParts) > 1 {
			cmd.Args = cmdParts[1:]
		}
		commands = append(commands, cmd)
	}

	if len(commands) == 0 {
		return fmt.Errorf("no valid commands in pipeline")
	}

	// Create pipeline context
	pipeCtx := &types.PipelineContext{
		Context:  ctx,
		Commands: commands,
	}

	// Execute pipeline
	execCtx := &types.ExecuteContext{
		Context:     ctx,
		IsPiped:     true,
		PipeContext: pipeCtx,
		Command:     *commands[0], // Set the first command as the main command
		Options: &types.ExecuteOptions{
			Env: env,
		},
	}

	result, err := exec.Execute(execCtx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %v", err)
	}

	fmt.Printf("Output:\n%s\n", result.Output)
	fmt.Printf("Exit Code: %d\n", result.ExitCode)

	return nil
}

func main() {
	ctx := context.Background()

	for _, example := range examples {
		fmt.Printf("\n=== Example: %s ===\n", example.name)
		err := runPipeline(ctx, example.pipeline, example.env)
		if err != nil {
			if example.expectedErr != "" && strings.Contains(err.Error(), example.expectedErr) {
				fmt.Printf("Got expected error: %v\n", err)
			} else {
				fmt.Printf("%s failed: %v\n", example.name, err)
			}
		}
	}
}
