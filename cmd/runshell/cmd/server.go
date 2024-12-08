package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iamlongalong/runshell/pkg/audit"
	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/server"
	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP server",
	Long: `Start the HTTP server that provides a RESTful API for command execution.
The server supports streaming output and command management.

Example:
  runshell server --http :8080
  runshell server --http :8080 --audit-dir /var/log/runshell`,
	RunE: runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&httpAddr, "http", ":8080", "HTTP server address")
}

func runServer(cmd *cobra.Command, args []string) error {
	fmt.Printf("Starting server with address: %s\n", httpAddr)

	// 创建本地执行器
	localExec := executor.NewLocalExecutor()

	// 如果指定了 Docker 镜像，创建 Docker 执行��
	var exec interface{} = localExec
	if dockerImage != "" {
		dockerExec, err := executor.NewDockerExecutor(dockerImage)
		if err != nil {
			return fmt.Errorf("failed to create Docker executor: %v", err)
		}
		exec = dockerExec
	}

	// 如果指定了审计目录，创建审计包装器
	if auditDir != "" {
		auditor, err := audit.NewAuditor(auditDir)
		if err != nil {
			return fmt.Errorf("failed to create auditor: %v", err)
		}
		defer auditor.Close()

		exec = executor.NewAuditedExecutor(exec.(types.Executor), auditor)
	}

	// 创建 HTTP 服务器
	srv := server.NewServer(exec.(types.Executor), httpAddr)

	// 创建错误通道
	errChan := make(chan error, 1)

	// 在后台启动服务器
	go func() {
		fmt.Printf("Starting HTTP server on %s\n", httpAddr)
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
			errChan <- err
		}
	}()

	// 等待服务器就绪
	fmt.Println("Waiting for server to be ready")
	srv.WaitForReady()
	fmt.Println("Server is ready")

	// 如果是测试模式，等待上下文取消
	if cmd.Context() != nil {
		<-cmd.Context().Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Stop(ctx)
	}

	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号或错误
	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %v", err)
	case sig := <-sigChan:
		fmt.Printf("Received signal %v, shutting down...\n", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Stop(ctx)
	}
}
