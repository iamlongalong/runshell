package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/server"
	"github.com/stretchr/testify/assert"
)

func waitForServer(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		// 尝试连接服务器
		conn, err := net.DialTimeout("tcp4", addr, time.Second)
		if err != nil {
			fmt.Printf("Failed to connect to %s: %v\n", addr, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		conn.Close()

		// 尝试发送健康检查请求
		client := &http.Client{
			Timeout: time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   time.Second,
					KeepAlive: time.Second,
				}).DialContext,
			},
		}
		resp, err := client.Get(fmt.Sprintf("http://%s/health", addr))
		if err != nil {
			fmt.Printf("Failed to send health check to %s: %v\n", addr, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return nil
		}
		fmt.Printf("Unexpected status code from %s: %d\n", addr, resp.StatusCode)
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("server did not start within %s", timeout)
}

func TestServerCommand(t *testing.T) {
	// ���置测试端口
	httpAddr = ":8081"

	// 创建本地执行器
	exec := executor.NewLocalExecutor()

	// 创建 HTTP 服务器
	srv := server.NewServer(exec, httpAddr)

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

	// 等待服务器启动
	fmt.Printf("Waiting for server to be ready on %s\n", httpAddr)
	err := waitForServer("127.0.0.1:8081", 10*time.Second)
	if err != nil {
		// 尝试获取错误通道中的错误
		select {
		case err := <-errChan:
			t.Fatalf("Server failed to start with error: %v", err)
		default:
			t.Fatalf("Server failed to start: %v", err)
		}
	}

	fmt.Println("Server is ready")

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Second,
			}).DialContext,
		},
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:8081/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// 发送请求并重试
	var resp *http.Response
	for i := 0; i < 5; i++ {
		fmt.Printf("Sending health check request (attempt %d)\n", i+1)
		resp, err = client.Do(req)
		if err == nil {
			break
		}
		fmt.Printf("Health check failed: %v\n", err)
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("Failed to send request after retries: %v", err)
	}
	defer resp.Body.Close()

	// 验证响应
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	fmt.Println("Health check successful")

	// 清理：关闭服务器
	if err := srv.Stop(ctx); err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}
	fmt.Println("Server stopped")
}
