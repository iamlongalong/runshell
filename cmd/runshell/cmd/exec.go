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
	workDir string
	envVars []string
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec [command] [args...]",
	Short: "Execute a command",
	Long: `Execute a command with optional environment variables and working directory.

Example:
  runshell exec ls -l
  runshell exec --env KEY=VALUE --workdir /tmp ls -l`,
	Args: cobra.MinimumNArgs(1),
	RunE: runExec,
}

func init() {
	rootCmd.AddCommand(execCmd)

	execCmd.Flags().StringVar(&workDir, "workdir", "", "Working directory for command execution")
	execCmd.Flags().StringArrayVar(&envVars, "env", nil, "Environment variables (KEY=VALUE)")

	// 禁用标志解析，这样可以正确处理命令参数中的标志
	execCmd.Flags().SetInterspersed(false)
}

func runExec(cmd *cobra.Command, args []string) error {
	// 创建本地执行器
	localExec := executor.NewLocalExecutor()

	// 如果指定了 Docker 镜像，创建 Docker 执行器
	var exec interface{} = localExec
	if dockerImage != "" {
		dockerExec, err := executor.NewDockerExecutor(dockerImage)
		if err != nil {
			return fmt.Errorf("failed to create Docker executor: %v", err)
		}
		exec = dockerExec
	}

	// 准备执行选项
	opts := &types.ExecuteOptions{
		WorkDir: workDir,
		Env:     make(map[string]string),
	}

	// 解析环境变量
	for _, env := range envVars {
		key, value, found := strings.Cut(env, "=")
		if !found {
			return fmt.Errorf("invalid environment variable format: %s", env)
		}
		opts.Env[key] = value
	}

	// 执行命令
	result, err := exec.(types.Executor).Execute(cmd.Context(), args[0], args[1:], opts)
	if err != nil {
		return fmt.Errorf("failed to execute command: %v", err)
	}

	// 输出结果
	if result.Output != "" {
		fmt.Print(result.Output)
	}

	// 如果有错误，输出到标准错误
	if result.Error != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", result.Error)
	}

	// 如果是测试模式，返回错误而不是退出
	if cmd.Context() != nil {
		if result.ExitCode != 0 {
			return fmt.Errorf("command failed with exit code %d", result.ExitCode)
		}
		return nil
	}

	// 使用命令的退出码退出
	os.Exit(result.ExitCode)
	return nil
}
