// Package server 实现了 RunShell 的 HTTP API 服务。
// 本文件实现了 HTTP 服务器的核心功能。
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/iamlongalong/runshell/pkg/types"
)

// Server 表示 HTTP 服务器。
// 特性：
// - RESTful API 接口
// - 健康检查
// - 命令执行
// - 优雅关闭
// - 线程安全
type Server struct {
	executor types.Executor // 命令执行器
	addr     string         // 服务器监听地址
	srv      *http.Server   // HTTP 服务器实例
	ready    chan struct{}  // 服务器就绪信号
	mu       sync.Mutex     // 保护并发访问
	started  bool           // 服务器启动状态
	ln       net.Listener   // TCP 监听器
}

// NewServer 创建一个新的 HTTP 服务器实例。
// 参数：
//   - executor：命令执行器实例
//   - addr：服务器监听地址（如 ":8080"）
//
// 返回值：
//   - *Server：服务器实例
func NewServer(executor types.Executor, addr string) *Server {
	return &Server{
		executor: executor,
		addr:     addr,
		ready:    make(chan struct{}),
	}
}

// Start 启动 HTTP 服务器。
// 启动流程：
// 1. 检查服务器状态
// 2. 创建路由
// 3. 配置监听器
// 4. 启动服务
// 5. 等待服务就绪
//
// 返回值：
//   - error：启动过程中的错误
func (s *Server) Start() error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return fmt.Errorf("server already started")
	}
	s.started = true
	s.mu.Unlock()

	fmt.Printf("Creating server mux for %s\n", s.addr)
	mux := http.NewServeMux()

	// 注册 API 路由
	mux.HandleFunc("/health", s.handleHealth) // 健康检查
	mux.HandleFunc("/exec", s.handleExec)     // 命令执行

	s.srv = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	// 创建 TCP 监听器
	fmt.Printf("Creating listener for %s\n", s.addr)
	ln, err := net.Listen("tcp4", s.addr)
	if err != nil {
		s.mu.Lock()
		s.started = false
		s.mu.Unlock()
		return fmt.Errorf("failed to create listener: %v", err)
	}
	s.ln = ln
	fmt.Printf("Listener created successfully for %s\n", s.addr)

	// 在后台启动服务
	go func() {
		fmt.Printf("Starting server on %s\n", s.addr)
		if err := s.srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// 等待服务就绪
	go func() {
		for {
			fmt.Printf("Trying to connect to %s\n", s.addr)
			conn, err := net.DialTimeout("tcp4", s.addr, 100*time.Millisecond)
			if err != nil {
				fmt.Printf("Connection attempt failed: %v\n", err)
				time.Sleep(10 * time.Millisecond)
				continue
			}
			conn.Close()
			fmt.Printf("Server is ready on %s\n", s.addr)
			close(s.ready)
			return
		}
	}()

	return nil
}

// WaitForReady 等待服务器就绪。
// 阻塞直到服务器完全启动并可以接受请求。
func (s *Server) WaitForReady() {
	<-s.ready
}

// Stop 停止 HTTP 服务器。
// 停止流程：
// 1. 检查服务器状态
// 2. 关闭监听器
// 3. 优雅关闭服务器
//
// 参数：
//   - ctx：用于控制关闭超时的上下文
//
// 返回值：
//   - error：关闭过程中的错误
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return nil
	}
	s.started = false
	s.mu.Unlock()

	if s.ln != nil {
		s.ln.Close()
	}

	if s.srv != nil {
		return s.srv.Shutdown(ctx)
	}
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

// ExecuteRequest 表示命令执行请求的结构。
type ExecuteRequest struct {
	Command string   `json:"command"`           // 要执行的命令
	Args    []string `json:"args"`              // 命令参数
	WorkDir string   `json:"workdir,omitempty"` // 工作目录
	Env     []string `json:"env,omitempty"`     // 环境变量
}

// ExecuteResponse 表示命令执行响应的结构。
type ExecuteResponse struct {
	ExitCode int    `json:"exit_code"`        // 命令退出码
	Output   string `json:"output,omitempty"` // 命令输出
	Error    string `json:"error,omitempty"`  // 错误信息
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

	// 解析请求体
	var req ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// 验证请求参数
	if req.Command == "" {
		http.Error(w, "Command is required", http.StatusBadRequest)
		return
	}

	// 准备执行选项
	opts := &types.ExecuteOptions{
		WorkDir: req.WorkDir,
		Env:     make(map[string]string),
	}

	// 解析环境变量
	for _, env := range req.Env {
		key, value, found := strings.Cut(env, "=")
		if !found {
			http.Error(w, fmt.Sprintf("Invalid environment variable format: %s", env), http.StatusBadRequest)
			return
		}
		opts.Env[key] = value
	}

	// 执行命令
	result, err := s.executor.Execute(r.Context(), req.Command, req.Args, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 准备响应
	resp := ExecuteResponse{
		ExitCode: result.ExitCode,
		Output:   result.Output,
	}
	if result.Error != nil {
		resp.Error = result.Error.Error()
	}

	// 发送响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleCommands 处理命令列表请求。
// 路由：GET /commands
// 查询参数：
//   - category：按类别过滤
//   - pattern：按模式匹配
//
// 响应：
//   - 200 OK：返回命令列表
//   - 405 Method Not Allowed：非 GET 请求
//   - 500 Internal Server Error：获取失败
func (s *Server) handleCommands(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filter := &types.CommandFilter{
		Category: r.URL.Query().Get("category"),
		Pattern:  r.URL.Query().Get("pattern"),
	}

	commands, err := s.executor.ListCommands(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(commands)
}

// handleHelp 处理命令帮助请求。
// 路由：GET /api/v1/help/{command}
// 路径参数：
//   - command：命令名称
//
// 响应：
//   - 200 OK：返回帮助信息
//   - 400 Bad Request：命令名称为空
//   - 405 Method Not Allowed：非 GET 请求
//   - 500 Internal Server Error：获取失败
func (s *Server) handleHelp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cmdName := strings.TrimPrefix(r.URL.Path, "/api/v1/help/")
	if cmdName == "" {
		http.Error(w, "Command name is required", http.StatusBadRequest)
		return
	}

	help, err := s.executor.GetCommandHelp(cmdName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
