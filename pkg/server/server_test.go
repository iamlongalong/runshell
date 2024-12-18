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

	"github.com/gin-gonic/gin"
	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

// MockExecutor 是一个用于测试的执行器实现
type MockExecutor struct {
	ExecuteFunc      func(ctx *types.ExecuteContext) (*types.ExecuteResult, error)
	ListCommandsFunc func() []types.CommandInfo
	NameFunc         func() string
	CloseFunc        func() error
}

func (m *MockExecutor) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx)
	}
	return &types.ExecuteResult{}, nil
}

func (m *MockExecutor) ExecuteCommand(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	return m.Execute(ctx)
}

func (m *MockExecutor) ListCommands() []types.CommandInfo {
	if m.ListCommandsFunc != nil {
		return m.ListCommandsFunc()
	}
	return nil
}

func (m *MockExecutor) Name() string {
	if m.NameFunc != nil {
		return m.NameFunc()
	}
	return "mock"
}

func (m *MockExecutor) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func TestHandleHealth(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建服务器实例
	s := NewServer(nil, ":8080")

	// 创建测试上下文
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 调用健康检查处理函数
	s.handleHealth(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestHandleExec(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建模拟执行器
	mockExecutor := &MockExecutor{
		ExecuteFunc: func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
			return &types.ExecuteResult{
				ExitCode: 0,
				Output:   "test output",
			}, nil
		},
	}

	// 创建服务器实例
	s := NewServer(types.ExecutorBuilderFunc(func(options *types.ExecuteOptions) (types.Executor, error) {
		return mockExecutor, nil
	}), ":8080")

	// 创建测试请求
	reqBody := ExecRequest{
		Command: "test",
		Args:    []string{"-a"},
	}
	body, _ := json.Marshal(reqBody)

	// 创建测试上下文
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/exec", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// 调用执行命令处理函数
	s.handleExec(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var resp types.ExecuteResult
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.ExitCode)
	assert.Equal(t, "test output", resp.Output)
}

func TestHandleListCommands(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建模拟执行器
	mockExecutor := &MockExecutor{
		ListCommandsFunc: func() []types.CommandInfo {
			return []types.CommandInfo{
				{
					Name:        "test",
					Description: "Test command",
					Usage:       "test [options]",
				},
			}
		},
	}

	// 创建服务器实例
	s := NewServer(types.ExecutorBuilderFunc(func(options *types.ExecuteOptions) (types.Executor, error) {
		return mockExecutor, nil
	}), ":8080")

	// 创建测试上下文
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/commands", nil)

	// 调用列出命令处理函数
	s.handleListCommands(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var commands []types.CommandInfo
	err := json.NewDecoder(w.Body).Decode(&commands)
	assert.NoError(t, err)
	assert.Len(t, commands, 1)
	assert.Equal(t, "test", commands[0].Name)
}

func TestHandleCommandHelp(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建模拟执行器
	mockExecutor := &MockExecutor{
		ListCommandsFunc: func() []types.CommandInfo {
			return []types.CommandInfo{
				{
					Name:  "test",
					Usage: "test command usage",
				},
			}
		},
	}

	// 创建服务器实例
	s := NewServer(types.ExecutorBuilderFunc(func(options *types.ExecuteOptions) (types.Executor, error) {
		return mockExecutor, nil
	}), ":8080")

	t.Run("valid command", func(t *testing.T) {
		// 创建测试上下文
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/help?command=test", nil)
		c.Request.URL.RawQuery = "command=test"

		// 调用命令帮助处理函数
		s.handleCommandHelp(c)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "test command usage", w.Body.String())
	})

	t.Run("missing command", func(t *testing.T) {
		// 创建测试上下文
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/help", nil)

		// 调用命令帮助处理函数
		s.handleCommandHelp(c)

		// 验证响应
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unknown command", func(t *testing.T) {
		// 创建测试上下文
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/help?command=unknown", nil)
		c.Request.URL.RawQuery = "command=unknown"

		// 调用命令帮助处理函数
		s.handleCommandHelp(c)

		// 验证响应
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestServerStartStop(t *testing.T) {
	// 创建服务器
	server := NewServer(types.NewMockExecutorBuilder(&types.MockExecutor{}), ":0")

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
