package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/iamlongalong/runshell/pkg/audit"
	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/server"
	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/spf13/cobra"
)

var (
	serverAddr  string
	auditDir    string
	dockerImage string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP server",
	Long:  `Start the HTTP server to handle command execution requests.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		// 创建执行器
		var exec types.Executor
		if dockerImage != "" {
			exec, err = executor.NewDockerExecutor(types.DockerConfig{
				Image: dockerImage,
			}, nil)
			if err != nil {
				return fmt.Errorf("failed to create Docker executor: %w", err)
			}
		} else {
			exec = executor.NewLocalExecutor(types.LocalConfig{
				AllowUnregisteredCommands: true,
			}, nil)
		}

		// 如果指定了审计目录，创建审计执行器
		if auditDir != "" {
			// 创建审计器
			logFile := filepath.Join(auditDir, "audit.log")
			auditor, err := audit.NewFileAuditor(logFile)
			if err != nil {
				return fmt.Errorf("failed to create auditor: %w", err)
			}

			// 创建审计执行器
			exec = executor.NewAuditedExecutor(exec, auditor)
		}

		// 创建服务器
		srv := server.NewServer(exec, serverAddr)

		// 启动服务器
		if err := srv.Start(); err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}

		// 等待中断信号
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		// 停止服务器
		if err := srv.Stop(); err != nil {
			return fmt.Errorf("failed to stop server: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&serverAddr, "addr", ":8080", "Server address")
	serverCmd.Flags().StringVar(&auditDir, "audit-dir", "", "Directory for audit logs")
	serverCmd.Flags().StringVar(&dockerImage, "docker-image", "", "Docker image to use")
}
