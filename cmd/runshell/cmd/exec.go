package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/spf13/cobra"
)

var (
	execDockerImage   string
	execWorkDir       string
	execEnvVars       []string
	allowUnregistered bool
)

var execCmd = &cobra.Command{
	Use:   "exec [command] [args...]",
	Short: "Execute a command",
	Long:  `Execute a command with optional Docker container and environment variables.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires at least 1 arg(s), only received %d", len(args))
		}

		// 创建执行器
		exec, err := createExecutor(execDockerImage)
		if err != nil {
			return fmt.Errorf("failed to create executor: %w", err)
		}

		// 创建管道执行器
		pipeExec := executor.NewPipelineExecutor(exec)

		// 非管道命令的处理
		ctx := &types.ExecuteContext{
			Context: cmd.Context(),
			Command: types.Command{Command: args[0], Args: args[1:]},
			Options: &types.ExecuteOptions{
				WorkDir: execWorkDir,
				Env:     parseEnvVars(execEnvVars),
				Stdin:   os.Stdin,
				Stdout:  os.Stdout,
				Stderr:  os.Stderr,
			},
			Executor: exec,
		}

		result, err := pipeExec.Execute(ctx)
		if err != nil {
			return fmt.Errorf("failed to execute command: %w", err)
		}

		if result.ExitCode != 0 {
			return fmt.Errorf("command failed with exit code %d", result.ExitCode)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
	execCmd.Flags().StringVar(&execDockerImage, "docker-image", "", "Docker image to run command in")
	execCmd.Flags().StringVar(&execWorkDir, "workdir", "", "Working directory for command execution")
	execCmd.Flags().StringArrayVarP(&execEnvVars, "env", "e", nil, "Environment variables (KEY=VALUE)")
	execCmd.Flags().BoolVarP(&allowUnregistered, "allow-unregistered", "a", true, "Allow unregistered commands")
}

func parseEnvVars(vars []string) map[string]string {
	env := make(map[string]string)
	for _, v := range vars {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}
	return env
}

// createExecutor 创建执行器
func createExecutor(execType string) (types.Executor, error) {
	var builder types.ExecutorBuilder

	switch execType {
	case "docker":
		builder = executor.NewDockerExecutorBuilder(types.DockerConfig{
			Image:                     "ubuntu:latest",
			WorkDir:                   "/workspace",
			AllowUnregisteredCommands: true,
		}).WithOptions(nil)
	case "local":
		builder = executor.NewLocalExecutorBuilder(types.LocalConfig{
			AllowUnregisteredCommands: true,
			UseBuiltinCommands:        true,
		}).WithOptions(nil)
	default:
		return nil, fmt.Errorf("unsupported executor type: %s", execType)
	}

	return builder.Build()
}
