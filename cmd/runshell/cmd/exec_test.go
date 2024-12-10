package cmd

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestExecCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "echo command",
			args:    []string{"echo", "hello"},
			wantErr: false,
		},
		{
			name:    "invalid command",
			args:    []string{"invalidcmd123"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟执行器
			var mockExec types.Executor
			if tt.wantErr {
				mockExec = &executor.MockExecutor{
					ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
						return &types.ExecuteResult{
							CommandName: ctx.Command.Command,
							ExitCode:    127,
						}, fmt.Errorf("command not found: %s", ctx.Command.Command)
					},
				}
			} else {
				mockExec = &executor.MockExecutor{
					ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
						return &types.ExecuteResult{
							CommandName: ctx.Command.Command,
							ExitCode:    0,
							Output:      "hello",
						}, nil
					},
				}
			}

			// 设置命令行参数
			cmd := &cobra.Command{
				Use:   "exec",
				Short: "Execute a command",
				RunE: func(cmd *cobra.Command, args []string) error {
					if len(args) < 1 {
						return fmt.Errorf("requires at least 1 arg(s), only received %d", len(args))
					}

					// 检查是否包含管道符
					cmdStr := strings.Join(args, " ")
					if strings.Contains(cmdStr, "|") {
						// 创建管道执行器
						pipeExec := executor.NewPipelineExecutor(mockExec)

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
						Command: types.Command{Command: args[0]},
						Options: &types.ExecuteOptions{
							WorkDir: execWorkDir,
							Env:     parseEnvVars(execEnvVars),
							Stdin:   os.Stdin,
							Stdout:  os.Stdout,
							Stderr:  os.Stderr,
						},
						Executor: mockExec,
					}

					result, err := mockExec.Execute(ctx)
					if err != nil {
						return err
					}

					if result != nil && result.ExitCode != 0 {
						return fmt.Errorf("command failed with exit code %d: %s", result.ExitCode, result.CommandName)
					}

					return nil
				},
			}

			cmd.SetArgs(tt.args)

			// 执行命令
			err := cmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
