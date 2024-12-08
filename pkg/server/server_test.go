package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

// TestHealthEndpoint 测试健康检查端点。
// 测试场景：
// 1. 发送 GET 请求到 /health
// 2. 验证返回状态码为 200 OK
func TestHealthEndpoint(t *testing.T) {
	// 创建本地执行器
	exec := executor.NewLocalExecutor()

	// 创建服务器实例
	server := NewServer(exec, ":8082")

	// 创建测试 HTTP 请求
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// 执行请求处理
	server.handleHealth(w, req)

	// 验证响应
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestExecEndpoint 测试命令执行端点。
// 测试场景：
// 1. 测试有效命令执行
//   - 发送正确的命令请求
//   - 验证成功执行和输出
//
// 2. 测试无效请求处理
//   - 发送缺少必要参数的请求
//   - 验证错误处理
func TestExecEndpoint(t *testing.T) {
	// 创建模拟执行器
	mockExec := &MockExecutor{
		executeFunc: func(ctx context.Context, cmdName string, args []string, opts *types.ExecuteOptions) (*types.ExecuteResult, error) {
			return &types.ExecuteResult{
				ExitCode: 0,
				Output:   "hello",
			}, nil
		},
	}

	// 创建服务器实例
	server := NewServer(mockExec, ":8082")

	// 定义测试用例
	tests := []struct {
		name           string         // 测试用例名称
		request        ExecuteRequest // 请求内容
		expectedStatus int            // 期望的 HTTP 状态码
		expectedOutput string         // 期望的输出内容
	}{
		{
			name: "valid command", // 测试有效命令
			request: ExecuteRequest{
				Command: "echo",
				Args:    []string{"hello"},
				WorkDir: "/tmp",
				Env:     []string{"TEST=value"},
			},
			expectedStatus: http.StatusOK,
			expectedOutput: "hello",
		},
		{
			name: "invalid request", // 测试无效请求
			request: ExecuteRequest{
				Command: "",
				Args:    []string{},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	// 执行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 准备请求数据
			bodyBytes, _ := json.Marshal(tt.request)

			// 创建 HTTP 请求
			req := httptest.NewRequest("POST", "/exec", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// 执行请求处理
			server.handleExec(w, req)

			// 验证响应状态码
			resp := w.Result()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// 对于成功的请求，验证响应内容
			if tt.expectedStatus == http.StatusOK {
				var execResp ExecuteResponse
				err := json.NewDecoder(resp.Body).Decode(&execResp)
				assert.NoError(t, err)
				assert.Equal(t, 0, execResp.ExitCode)
				assert.Contains(t, execResp.Output, tt.expectedOutput)
			}
		})
	}
}

// MockExecutor 是一个用于测试的模拟执行器。
// 实现了 types.Executor 接口，提供可配置的行为。
// 特性：
// - 可配置的命令执行行为
// - 默认返回成功结果
// - 简化的接口实现
type MockExecutor struct {
	executeFunc func(ctx context.Context, cmdName string, args []string, opts *types.ExecuteOptions) (*types.ExecuteResult, error)
}

// Execute 执行模拟的命令。
// 如果设置了 executeFunc，则使用它；
// 否则返回默认的成功结果。
func (m *MockExecutor) Execute(ctx context.Context, cmdName string, args []string, opts *types.ExecuteOptions) (*types.ExecuteResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, cmdName, args, opts)
	}
	return &types.ExecuteResult{
		ExitCode: 0,
		Output:   "mock output",
	}, nil
}

// GetCommandInfo 返回空的命令信息（测试用）。
func (m *MockExecutor) GetCommandInfo(cmdName string) (*types.Command, error) {
	return nil, nil
}

// GetCommandHelp 返回空的帮助信息（测试用）。
func (m *MockExecutor) GetCommandHelp(cmdName string) (string, error) {
	return "", nil
}

// ListCommands 返回空的命令列表（测试用）。
func (m *MockExecutor) ListCommands(filter *types.CommandFilter) ([]*types.Command, error) {
	return nil, nil
}

// RegisterCommand 返回成功（测试用）。
func (m *MockExecutor) RegisterCommand(cmd *types.Command) error {
	return nil
}

// UnregisterCommand 返回成功（测试用）。
func (m *MockExecutor) UnregisterCommand(cmdName string) error {
	return nil
}
