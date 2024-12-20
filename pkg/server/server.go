// Package server 实现了 RunShell 的 HTTP API 服务。
// 本文件实现了 HTTP 服务器的核心功能。
package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iamlongalong/runshell/cmd/runshell/docs"
	"github.com/iamlongalong/runshell/pkg/log"
	"github.com/iamlongalong/runshell/pkg/types"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {
	// programmatically set swagger info
	docs.SwaggerInfo.Title = "RunShell API"
	docs.SwaggerInfo.Description = "API for executing and managing shell commands"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http"}
}

// @title           RunShell API
// @version         1.0
// @description     API for executing and managing shell commands
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1
// @schemes   http

// @securityDefinitions.basic  BasicAuth

// ErrorResponse 表示错误响应
// swagger:model
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request parameter"` // 错误信息
}

// ExecRequest 表示执行命令的请求
// swagger:model
type ExecRequest struct {
	Command string   `json:"command" binding:"required" example:"ls"`  // 要执行的命令
	Args    []string `json:"args,omitempty" example:"[\"-l\",\"-a\"]"` // 命令参数

	WorkDir string            `json:"workdir,omitempty"` // 工作目录
	Env     map[string]string `json:"env,omitempty"`     // 环境变量
}

// ExecResponse 表示执行命令的响应
// swagger:model
type ExecResponse struct {
	ExitCode int    `json:"exit_code" example:"0"`      // 命令退出码
	Output   string `json:"output" example:"file1.txt"` // 命令输出
	Error    string `json:"error,omitempty"`            // 错误信息，如果有的话
}

// Server 表示 HTTP 服务器。
type Server struct {
	executorBuilder types.ExecutorBuilder
	sessionManager  types.SessionManager
	addr            string
	engine          *gin.Engine
	server          *http.Server
	listener        net.Listener
	mu              sync.Mutex
}

// NewServer 创建新的服务器
func NewServer(executorBuilder types.ExecutorBuilder, addr string) *Server {
	// 启用调试模式
	gin.SetMode(gin.DebugMode)
	engine := gin.Default()

	// 添加详细的请求响应日志中间件
	engine.Use(func(c *gin.Context) {
		// 生成请求ID
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())
		c.Set("RequestID", requestID)

		// 开始时间
		start := time.Now()

		// 打印请求信息
		log.Debug("\n=== [REQUEST-%s] %v ===", requestID, start.Format("2006-01-02 15:04:05.000"))
		log.Debug("Path: %s", c.Request.URL.Path)
		log.Debug("Method: %s", c.Request.Method)
		log.Debug("Client IP: %s", c.ClientIP())
		log.Debug("Headers:")
		for k, v := range c.Request.Header {
			log.Debug("  %s: %v", k, v)
		}
		log.Debug("Query Parameters:")
		for k, v := range c.Request.URL.Query() {
			log.Debug("  %s: %v", k, v)
		}

		// 如果是 JSON 请求，打印请求体
		if c.Request.Body != nil && c.Request.Header.Get("Content-Type") == "application/json" {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			// 重新设置 body 以供后续中间件使用
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			log.Debug("Request Body (JSON):\n%s", string(bodyBytes))
		}

		// 创建自定义的 ResponseWriter 来捕获响应
		blw := &bodyLogWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(start)

		// 打印响应信息
		log.Debug("\n=== [RESPONSE-%s] %v ===", requestID, time.Now().Format("2006-01-02 15:04:05.000"))
		log.Debug("Duration: %v", duration)
		log.Debug("Status: %d", c.Writer.Status())
		log.Debug("Response Headers:")
		for k, v := range c.Writer.Header() {
			log.Debug("  %s: %v", k, v)
		}

		// 打印响应体
		if blw.body.Len() > 0 {
			contentType := c.Writer.Header().Get("Content-Type")
			log.Debug("Response Body (%s):\n%s", contentType, blw.body.String())
		}

		// 如果有错误，打印错误信息
		if len(c.Errors) > 0 {
			log.Debug("\nErrors:")
			for i, err := range c.Errors {
				log.Debug("  %d. [%v] %v", i+1, err.Type, err.Err)
				if err.Meta != nil {
					log.Debug("     Meta: %+v", err.Meta)
				}
			}
		}

		log.Debug("\n=== [END-%s] === Total Time: %v ===", requestID, duration)
		log.Debug(strings.Repeat("=", 80))
	})

	s := &Server{
		executorBuilder: executorBuilder,
		sessionManager:  NewMemorySessionManager(),
		addr:            addr,
		engine:          engine,
	}

	s.setupRoutes()
	return s
}

// bodyLogWriter 是一个自定义的 ResponseWriter，用于捕获响应体和状态码
type bodyLogWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// setupRoutes 设置所有路由
func (s *Server) setupRoutes() {
	// API 文档
	s.engine.GET("/swagger/*any", func(c *gin.Context) {
		log.Debug("Swagger request received: %s", c.Request.URL.Path)
		log.Debug("Swagger request headers: %v", c.Request.Header)

		if c.Param("any") == "/doc.json" {
			doc := docs.SwaggerInfo.ReadDoc()
			c.Header("Content-Type", "application/json")
			c.String(http.StatusOK, doc)
			return
		}

		handler := ginSwagger.WrapHandler(
			swaggerFiles.Handler,
			ginSwagger.URL("/swagger/doc.json"),
			ginSwagger.DefaultModelsExpandDepth(-1),
			ginSwagger.DocExpansion("none"),
			ginSwagger.InstanceName("runshell"),
		)

		handler(c)
	})

	// API v1 路由组
	v1 := s.engine.Group("/api/v1")
	{
		// 健康检查
		v1.GET("/health", s.handleHealth)

		// 命令执行
		v1.POST("/exec", s.handleExec)
		v1.GET("/exec/interactive", s.handleInteractiveExec)
		v1.GET("/commands", s.handleListCommands)
		v1.GET("/help", s.handleCommandHelp)

		// 会话管理
		v1.GET("/sessions", s.handleListSessions)
		v1.POST("/sessions", s.handleCreateSession)
		v1.DELETE("/sessions/:id", s.handleDeleteSession)
		v1.POST("/sessions/:id/exec", s.handleSessionExec)
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		return fmt.Errorf("server is already running")
	}

	// 创建监听器
	log.Info("Creating listener for %s", s.addr)
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	log.Info("Listener created successfully for %s", s.addr)

	s.listener = listener
	s.server = &http.Server{
		Handler: s.engine,
	}

	// 启动服务器
	log.Info("Starting server on %s", s.addr)
	go func() {
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Error("Server error: %v", err)
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

// @Summary     Health Check
// @Description Check if the server is running
// @Tags        health
// @Accept      json
// @Produce     json
// @Success     200 {string} string "OK"
// @Router      /health [get]
func (s *Server) handleHealth(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

// @Summary     Execute Command
// @Description Execute a shell command
// @Tags        commands
// @Accept      json
// @Produce     json
// @Param       request body ExecRequest true "Command execution request"
// @Success     200 {object} ExecResponse
// @Failure     400 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /exec [post]
func (s *Server) handleExec(c *gin.Context) {
	var req ExecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.handleError(c, http.StatusBadRequest, err, "Invalid request format")
		return
	}

	log.Debug("Received exec request: %+v", req)

	executor, err := s.executorBuilder.Build(&types.ExecuteOptions{
		WorkDir: req.WorkDir,
		Env:     req.Env,
	})
	if err != nil {
		s.handleError(c, http.StatusInternalServerError, err, fmt.Sprintf("Failed to create executor: %v", err))
		return
	}

	log.Debug("Created executor: %s", executor.Name())

	// 准备执行选项
	var outputBuf bytes.Buffer
	opts := &types.ExecuteOptions{
		WorkDir: req.WorkDir,
		Env:     req.Env,
		Stdout:  &outputBuf,
		Stderr:  &outputBuf,
	}

	log.Debug("Prepared execution options: %+v", opts)

	// 执行命令
	execCtx := &types.ExecuteContext{
		Context: c.Request.Context(),
		Command: types.Command{
			Command: req.Command,
			Args:    req.Args,
		},
		Options:  opts,
		Executor: executor,
	}

	log.Debug("Executing command: %s %v", req.Command, req.Args)

	result, err := executor.Execute(execCtx)
	if err != nil {
		log.Error("Command execution failed: %v", err)
		s.handleError(c, http.StatusInternalServerError, err, fmt.Sprintf("Command execution failed: %v", err))
		return
	}

	log.Info("Command execution succeeded: %+v", result)

	// 构造响应
	response := ExecResponse{
		ExitCode: result.ExitCode,
		Output:   result.Output,
	}
	if result.Error != nil {
		response.Error = result.Error.Error()
	}

	c.JSON(http.StatusOK, response)
}

// @Summary     List Commands
// @Description List all available commands
// @Tags        commands
// @Accept      json
// @Produce     json
// @Success     200 {array} types.CommandInfo
// @Failure     500 {object} ErrorResponse
// @Router      /commands [get]
func (s *Server) handleListCommands(c *gin.Context) {
	executor, err := s.executorBuilder.Build(nil)
	if err != nil {
		s.handleError(c, http.StatusInternalServerError, err, fmt.Sprintf("Failed to create executor: %v", err))
		return
	}

	commands := executor.ListCommands()
	c.JSON(http.StatusOK, commands)
}

// @Summary     Get Command Help
// @Description Get help information for a specific command
// @Tags        commands
// @Accept      json
// @Produce     json
// @Param       command query string true "Command name"
// @Success     200 {string} string
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /help [get]
func (s *Server) handleCommandHelp(c *gin.Context) {
	cmdName := c.Query("command")
	if cmdName == "" {
		s.handleError(c, http.StatusBadRequest, fmt.Errorf("command name is required"), "")
		return
	}

	executor, err := s.executorBuilder.Build(nil)
	if err != nil {
		s.handleError(c, http.StatusInternalServerError, err, fmt.Sprintf("Failed to create executor: %v", err))
		return
	}

	help, err := s.getCommandHelp(executor, cmdName)
	if err != nil {
		s.handleError(c, http.StatusNotFound, err, fmt.Sprintf("Failed to get command help: %v", err))
		return
	}

	c.String(http.StatusOK, help)
}

// @Summary     List Sessions
// @Description List all active sessions
// @Tags        sessions
// @Accept      json
// @Produce     json
// @Success     200 {array} types.Session
// @Failure     500 {object} ErrorResponse
// @Router      /sessions [get]
func (s *Server) handleListSessions(c *gin.Context) {
	sessions, err := s.sessionManager.ListSessions()
	if err != nil {
		s.handleError(c, http.StatusInternalServerError, err, "")
		return
	}
	c.JSON(http.StatusOK, sessions)
}

// @Summary     Create Session
// @Description Create a new session
// @Tags        sessions
// @Accept      json
// @Produce     json
// @Param       request body types.SessionRequest true "Session creation request"
// @Success     200 {object} types.SessionResponse
// @Failure     400 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /sessions [post]
func (s *Server) handleCreateSession(c *gin.Context) {
	var req types.SessionRequest
	// TODO: 支持 builder 的配置，可使用 config 来配置
	if err := c.ShouldBindJSON(&req); err != nil {
		s.handleError(c, http.StatusBadRequest, err, "Invalid request format")
		return
	}

	executor, err := s.executorBuilder.Build(&types.ExecuteOptions{
		WorkDir: req.Options.WorkDir,
		Env:     req.Options.Env,
	})
	if err != nil {
		s.handleError(c, http.StatusInternalServerError, err, "")
		return
	}

	session, err := s.sessionManager.CreateSession(executor, req.Options)
	if err != nil {
		s.handleError(c, http.StatusInternalServerError, err, "")
		return
	}

	c.JSON(http.StatusOK, types.SessionResponse{Session: session})
}

// @Summary     Delete Session
// @Description Delete a session by ID
// @Tags        sessions
// @Accept      json
// @Produce     json
// @Param       id path string true "Session ID"
// @Success     204 "No Content"
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /sessions/{id} [delete]
func (s *Server) handleDeleteSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		s.handleError(c, http.StatusBadRequest, fmt.Errorf("session id is required"), "")
		return
	}
	if err := s.sessionManager.DeleteSession(sessionID); err != nil {
		s.handleError(c, http.StatusNotFound, err, "")
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Execute Command in Session
// @Description Execute a command in a specific session
// @Tags        sessions
// @Accept      json
// @Produce     json
// @Param       id path string true "Session ID"
// @Param       request body ExecRequest true "Command execution request"
// @Success     200 {object} ExecResponse
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /sessions/{id}/exec [post]
func (s *Server) handleSessionExec(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		s.handleError(c, http.StatusBadRequest, fmt.Errorf("session id is required"), "")
		return
	}

	session, err := s.sessionManager.GetSession(sessionID)
	if err != nil {
		s.handleError(c, http.StatusNotFound, err, "")
		return
	}

	var req ExecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.handleError(c, http.StatusBadRequest, err, "Invalid request format")
		return
	}

	opts := &types.ExecuteOptions{
		WorkDir: req.WorkDir,
	}

	if session.Options != nil && session.Options.Env != nil {
		opts.Env = session.Options.Env
	}

	execCtx := &types.ExecuteContext{
		Context: c.Request.Context(),
		Command: types.Command{
			Command: req.Command,
			Args:    req.Args,
		},
		Options:  opts,
		Executor: session.Executor,
	}

	result, err := session.Executor.Execute(execCtx)
	if err != nil {
		s.handleError(c, http.StatusInternalServerError, err, "")
		return
	}

	c.JSON(http.StatusOK, result)
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

// handleError 统一的错误处理函数
func (s *Server) handleError(c *gin.Context, status int, err error, msg string) {
	// 添加错误到 gin context
	_ = c.Error(err)

	if status == http.StatusInternalServerError {
		log.Error("Internal Server Error Details:")
		log.Error("Error: %v", err)
		log.Error("Path: %s", c.Request.URL.Path)
		log.Error("Method: %s", c.Request.Method)
		log.Error("Message: %s", msg)
		log.Error("Headers: %v", c.Request.Header)
		log.Error("Query: %v", c.Request.URL.Query())
		if c.Request.Body != nil {
			body, _ := c.GetRawData()
			log.Error("Body: %s", string(body))
		}
		log.Error("----------------------------------------")
	}

	if msg == "" {
		msg = err.Error()
	}
	c.JSON(status, ErrorResponse{Error: msg})
}
