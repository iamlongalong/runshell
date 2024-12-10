package server

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestHealthEndpoint(t *testing.T) {
	// 创建服务器
	server := NewServer(&executor.MockExecutor{}, ":8080")

	// 创建测试请求
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// 处理请求
	server.handleHealth(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestExecEndpoint(t *testing.T) {
	t.Run("valid command", func(t *testing.T) {
		// 创建模拟执行器
		mockExec := &executor.MockExecutor{
			ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
				assert.Equal(t, "test", ctx.Command.Command)
				assert.Equal(t, []string{"arg1", "arg2"}, ctx.Command.Args)
				assert.Equal(t, "/tmp", ctx.Options.WorkDir)
				assert.Equal(t, map[string]string{"FOO": "bar"}, ctx.Options.Env)
				return &types.ExecuteResult{
					CommandName: "test",
					ExitCode:    0,
					StartTime:   time.Now(),
					EndTime:     time.Now(),
				}, nil
			},
		}

		// 创建服务器
		server := NewServer(mockExec, ":8080")

		// 创建请求体
		reqBody := ExecRequest{
			Command: "test",
			Args:    []string{"arg1", "arg2"},
			WorkDir: "/tmp",
			Env:     map[string]string{"FOO": "bar"},
		}
		body, _ := json.Marshal(reqBody)

		// 创建测试请求
		req := httptest.NewRequest("POST", "/exec", bytes.NewReader(body))
		w := httptest.NewRecorder()

		// 处理请求
		server.handleExec(w, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)

		var result types.ExecuteResult
		err := json.NewDecoder(w.Body).Decode(&result)
		assert.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
	})

	t.Run("invalid request", func(t *testing.T) {
		// 创建服务器
		server := NewServer(&executor.MockExecutor{}, ":8080")

		// 创建无效的请求体
		reqBody := []byte(`{invalid json}`)

		// 创建测试请求
		req := httptest.NewRequest("POST", "/exec", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		// 处理请求
		server.handleExec(w, req)

		// 验证响应
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestListCommandsEndpoint(t *testing.T) {
	// 创建服务器
	server := NewServer(&executor.MockExecutor{}, ":8080")

	// 创建测试请求
	req := httptest.NewRequest("GET", "/commands", nil)
	w := httptest.NewRecorder()

	// 处理请求
	server.handleListCommands(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var commands []types.CommandInfo
	err := json.NewDecoder(w.Body).Decode(&commands)
	assert.NoError(t, err)
	assert.Len(t, commands, 1)
	assert.Equal(t, "test", commands[0].Name)
}

func TestCommandHelpEndpoint(t *testing.T) {
	// 创建服务器
	server := NewServer(&executor.MockExecutor{}, ":8080")

	// 创建测试请求
	req := httptest.NewRequest("GET", "/help?command=test", nil)
	w := httptest.NewRecorder()

	// 处理请求
	server.handleCommandHelp(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test [args...]", w.Body.String())
}

func TestServerStartStop(t *testing.T) {
	// 创建服务器
	server := NewServer(&executor.MockExecutor{}, ":0")

	// 启动服务器
	err := server.Start()
	assert.NoError(t, err)

	// 等待服务器就绪
	time.Sleep(100 * time.Millisecond)

	// 停止服务器
	err = server.Stop()
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		assert.NoError(t, err)
	}

	// 等待服务器完全关闭
	time.Sleep(100 * time.Millisecond)

	// 验证服务器已关闭
	conn, err := net.DialTimeout("tcp", server.addr, time.Second)
	if err == nil {
		conn.Close()
		t.Fatal("Server is still running")
	}
}
