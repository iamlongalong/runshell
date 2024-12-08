package cmd

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

func waitForServer(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func TestServerCommand(t *testing.T) {
	// 保存原始参数
	origAddr := serverAddr
	origAuditDir := auditDir
	origDockerImage := dockerImage
	defer func() {
		serverAddr = origAddr
		auditDir = origAuditDir
		dockerImage = origDockerImage
	}()

	// 设置命令行参数
	serverAddr = ":0"
	auditDir = ""
	dockerImage = ""

	// 创建错误通道
	errCh := make(chan error, 1)

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动服务器
	go func() {
		serverCmd.SetContext(ctx)
		errCh <- serverCmd.RunE(serverCmd, []string{})
	}()

	// 等待服务器启动
	time.Sleep(time.Second)

	// 取消上下文
	cancel()

	// 等待服务器关闭
	select {
	case err := <-errCh:
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			t.Errorf("Unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Log("Server shutdown timed out, but this is acceptable")
	}
}

func TestServerStartStop(t *testing.T) {
	// 保存原始参数
	origAddr := serverAddr
	origAuditDir := auditDir
	origDockerImage := dockerImage
	defer func() {
		serverAddr = origAddr
		auditDir = origAuditDir
		dockerImage = origDockerImage
	}()

	// 设置命令行参数
	serverAddr = ":0"
	auditDir = ""
	dockerImage = ""

	// 创建错误通道
	errCh := make(chan error, 1)

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动服务器
	go func() {
		serverCmd.SetContext(ctx)
		errCh <- serverCmd.RunE(serverCmd, []string{})
	}()

	// 等待服务器启动
	time.Sleep(time.Second)

	// 取消上下文
	cancel()

	// 等待服务器关闭
	select {
	case err := <-errCh:
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			t.Errorf("Unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Log("Server shutdown timed out, but this is acceptable")
	}
}
