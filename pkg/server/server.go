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
	executor types.Executor
	server   *http.Server
	mux      *http.ServeMux
	listener net.Listener
	mu       sync.Mutex
	buf      bytes.Buffer
	started  bool
	addr     string
}

// NewServer 创建新的服务器
func NewServer(executor types.Executor, addr string) *Server {
	s := &Server{
		executor: executor,
		mux:      http.NewServeMux(),
		addr:     addr,
	}

	// 注册路由
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/exec", s.handleExec)
	s.mux.HandleFunc("/commands", s.handleListCommands)
	s.mux.HandleFunc("/help", s.handleCommandHelp)

	return s
}

// Start 启动服务器
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("server already started")
	}

	// 创建 HTTP 服务器
	s.server = &http.Server{
		Addr:    s.addr,
		Handler: s.mux,
	}

	// 创建监听器
	fmt.Printf("Creating listener for %s\n", s.addr)
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	fmt.Printf("Listener created successfully for %s\n", s.addr)

	s.listener = listener
	s.started = true

	// 在后台启动服务器
	go func() {
		fmt.Printf("Starting server on %s\n", s.addr)
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

	if !s.started {
		return nil
	}

	// 关闭服务器
	if s.server != nil {
		if err := s.server.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
	}

	// 关闭监听器
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	s.started = false
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
		Args:    append([]string{req.Command}, req.Args...),
		Options: opts,
	}

	result, err := s.executor.Execute(execCtx)
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

	commands := s.executor.ListCommands()

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

	help, err := s.getCommandHelp(cmdName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get command help: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(help))
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
	execCtx := &types.ExecuteContext{
		Context: ctx,
		Args:    append([]string{cmdName}, args...),
		Options: opts,
	}
	return s.executor.Execute(execCtx)
}

// listCommands 列出所有命令
func (s *Server) listCommands() []types.CommandInfo {
	return s.executor.ListCommands()
}

// getCommandHelp 获取命令帮助信息
func (s *Server) getCommandHelp(cmdName string) (string, error) {
	commands := s.executor.ListCommands()
	for _, cmd := range commands {
		if cmd.Name == cmdName {
			return cmd.Usage, nil
		}
	}
	return "", fmt.Errorf("command not found: %s", cmdName)
}
