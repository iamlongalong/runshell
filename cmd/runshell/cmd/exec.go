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
	execDockerImage string
	execWorkDir     string
	execEnvVars     []string
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
		var exec types.Executor
		if execDockerImage != "" {
			exec = executor.NewDockerExecutor(execDockerImage)
		} else {
			exec = executor.NewLocalExecutor()
		}

		// 创建管道执行器
		pipeExec := executor.NewPipelineExecutor(exec)

		// 检查是否包含管道符
		cmdStr := strings.Join(args, " ")
		if strings.Contains(cmdStr, "|") {
			// 解析管道命令
			pipeline, err := pipeExec.ParsePipeline(cmdStr)
			if err != nil {
				return fmt.Errorf("failed to parse pipeline: %w", err)
			}

			// 设置执行选项
			pipeline.Options = &types.ExecuteOptions{
				WorkDir: execWorkDir,
				Env:     parseEnvVars(execEnvVars),
				Stdin:   os.Stdin,
				Stdout:  os.Stdout,
				Stderr:  os.Stderr,
			}
			pipeline.Context = cmd.Context()

			// 执行管道命令
			result, err := pipeExec.ExecutePipeline(pipeline)
			if err != nil {
				return fmt.Errorf("failed to execute pipeline: %w", err)
			}

			if result.ExitCode != 0 {
				return fmt.Errorf("pipeline failed with exit code %d", result.ExitCode)
			}

			return nil
		}

		// 非管道命令的处理
		ctx := &types.ExecuteContext{
			Context: cmd.Context(),
			Args:    args,
			Options: &types.ExecuteOptions{
				WorkDir: execWorkDir,
				Env:     parseEnvVars(execEnvVars),
				Stdin:   os.Stdin,
				Stdout:  os.Stdout,
				Stderr:  os.Stderr,
			},
		}

		result, err := exec.Execute(ctx)
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
	execCmd.Flags().StringArrayVar(&execEnvVars, "env", nil, "Environment variables (KEY=VALUE)")
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
