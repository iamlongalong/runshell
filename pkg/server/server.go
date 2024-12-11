// Package server 实现了 RunShell 的 HTTP API 服务。
// 本文件实现了 HTTP 服务器的核心功能。
package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/iamlongalong/runshell/pkg/types"
)

// ExecRequest 表示执行命令的请求
type ExecRequest struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	WorkDir string            `json:"workdir"`
	Env     map[string]string `json:"env"`
}

// ExecResponse 表示执行命令的响应
type ExecResponse struct {
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
}

// Server 表示 HTTP 服务器。
// 特性：
// - RESTful API 接口
// - 健康检查
// - 命令执行
// - 优雅关闭
// - 线程安全
type Server struct {
	executorBuilder types.ExecutorBuilder
	sessionManager  types.SessionManager
	addr            string
	server          *http.Server
	listener        net.Listener
	mu              sync.Mutex
}

// NewServer 创建新的服务器
func NewServer(executorBuilder types.ExecutorBuilder, addr string) *Server {
	return &Server{
		executorBuilder: executorBuilder,
		sessionManager:  NewMemorySessionManager(),
		addr:            addr,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		return fmt.Errorf("server is already running")
	}

	// 创建路由
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/exec", s.handleExec)
	mux.HandleFunc("/commands", s.handleListCommands)
	mux.HandleFunc("/help", s.handleCommandHelp)

	// 会话管理路由
	mux.HandleFunc("/sessions", s.handleSessions)
	mux.HandleFunc("/sessions/", s.handleSessionOperations)

	// 创建监听器
	fmt.Printf("Creating listener for %s\n", s.addr)
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	fmt.Printf("Listener created successfully for %s\n", s.addr)

	s.listener = listener
	s.server = &http.Server{
		Handler: mux,
	}

	// 启动服务器
	fmt.Printf("Starting server on %s\n", s.addr)
	go func() {
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	return nil
}

// Stop 停止服务器
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server == nil {
		return nil
	}

	// 关闭所有会话
	sessions, _ := s.sessionManager.ListSessions()
	for _, session := range sessions {
		s.sessionManager.DeleteSession(session.ID)
	}

	// 关闭服务器
	if err := s.server.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	s.server = nil
	return nil
}

// handleHealth 处理健康检查请求。
// 路由：GET /health
// 响应：
//   - 200 OK：服务器正常运行
//   - 405 Method Not Allowed：非 GET 请求
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleExec 处理命令执行请求。
// 路由：POST /exec
// 请求体：ExecuteRequest JSON
// 响应：
//   - 200 OK：命令执行成功，返回 ExecuteResponse
//   - 400 Bad Request：请求格式错误
//   - 405 Method Not Allowed：非 POST 请求
//   - 500 Internal Server Error：执行错误
func (s *Server) handleExec(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求
	var req ExecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// 创建执行器
	executor, err := s.executorBuilder.Build()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create executor: %v", err), http.StatusInternalServerError)
		return
	}

	// 准备执行选项
	opts := &types.ExecuteOptions{
		WorkDir: req.WorkDir,
		Env:     req.Env,
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}

	// 执行命令
	execCtx := &types.ExecuteContext{
		Context: r.Context(),
		Command: types.Command{
			Command: req.Command,
			Args:    req.Args,
		},
		Options:  opts,
		Executor: executor,
	}

	result, err := executor.Execute(execCtx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Command execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回结果
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleListCommands 处理列出命令请求
func (s *Server) handleListCommands(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 创建执行器
	executor, err := s.executorBuilder.Build()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create executor: %v", err), http.StatusInternalServerError)
		return
	}

	commands := executor.ListCommands()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(commands)
}

// handleCommandHelp 处理获取命令帮助请求
func (s *Server) handleCommandHelp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cmdName := r.URL.Query().Get("command")
	if cmdName == "" {
		http.Error(w, "Command name is required", http.StatusBadRequest)
		return
	}

	// 创建执行器
	executor, err := s.executorBuilder.Build()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create executor: %v", err), http.StatusInternalServerError)
		return
	}

	help, err := s.getCommandHelp(executor, cmdName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get command help: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(help))
}

// handleSessions 处理会话管理请求
func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
		// 创建新会话
		var req types.SessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
			return
		}

		// 创建执行器
		executor, err := s.executorBuilder.Build()
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		// 创建会话
		session, err := s.sessionManager.CreateSession(executor, req.Options)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		// 返回会话信息
		resp := types.SessionResponse{Session: session}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

	case http.MethodGet:
		// 列出所有会话
		sessions, err := s.sessionManager.ListSessions()
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(sessions); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// handleSessionOperations 处理会话操作请求
func (s *Server) handleSessionOperations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 获取会话ID
	sessionID := strings.TrimPrefix(r.URL.Path, "/sessions/")
	if sessionID == "" {
		http.Error(w, `{"error": "session ID is required"}`, http.StatusBadRequest)
		return
	}

	// 获取会话
	session, err := s.sessionManager.GetSession(sessionID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodPost:
		// 解析请求
		var req struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
			WorkDir string   `json:"workdir"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
			return
		}

		// 准备执行选项
		opts := &types.ExecuteOptions{
			WorkDir: req.WorkDir,
		}

		if session.Options != nil && session.Options.Env != nil {
			opts.Env = session.Options.Env
		}

		// 创建输出缓冲区
		var stdout, stderr bytes.Buffer
		opts.Stdout = &stdout
		opts.Stderr = &stderr

		// 执行命令
		execCtx := &types.ExecuteContext{
			Context: r.Context(),
			Command: types.Command{
				Command: req.Command,
				Args:    req.Args,
			},
			Options:  opts,
			Executor: session.Executor,
		}

		result, err := session.Executor.Execute(execCtx)
		if err != nil {
			// 准备错误响应
			response := struct {
				Error  string `json:"error"`
				Output string `json:"output"`
			}{
				Error:  err.Error(),
				Output: stdout.String() + stderr.String(),
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 准备成功响应
		response := struct {
			types.ExecuteResult
			Output string `json:"output"`
		}{
			ExecuteResult: *result,
			Output:        stdout.String() + stderr.String(),
		}

		// 返回结果
		json.NewEncoder(w).Encode(response)

	case http.MethodDelete:
		// 删除会话
		if err := s.sessionManager.DeleteSession(sessionID); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// StreamScanner helps to scan streaming output
type StreamScanner struct {
	reader *io.PipeReader
	buf    []byte
}

func NewStreamScanner(reader *io.PipeReader) *StreamScanner {
	return &StreamScanner{
		reader: reader,
		buf:    make([]byte, 4096),
	}
}

func (s *StreamScanner) Scan() bool {
	n, err := s.reader.Read(s.buf)
	if err != nil {
		return false
	}
	s.buf = s.buf[:n]
	return true
}

func (s *StreamScanner) Text() string {
	return string(s.buf)
}

// executeCommand 执行命令
func (s *Server) executeCommand(ctx context.Context, cmdName string, args []string, opts *types.ExecuteOptions) (*types.ExecuteResult, error) {
	// 创建执行器
	executor, err := s.executorBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	execCtx := &types.ExecuteContext{
		Context: ctx,
		Command: types.Command{
			Command: cmdName,
			Args:    args,
		},
		Options:  opts,
		Executor: executor,
	}
	return executor.Execute(execCtx)
}

// listCommands 列出所有命令
func (s *Server) listCommands() []types.CommandInfo {
	// 创建执行器
	executor, err := s.executorBuilder.Build()
	if err != nil {
		return nil
	}
	return executor.ListCommands()
}

// getCommandHelp 获取命令帮助信息
func (s *Server) getCommandHelp(executor types.Executor, cmdName string) (string, error) {
	commands := executor.ListCommands()
	for _, cmd := range commands {
		if cmd.Name == cmdName {
			return cmd.Usage, nil
		}
	}
	return "", fmt.Errorf("command not found: %s", cmdName)
}
