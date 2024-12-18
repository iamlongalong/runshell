// Package server 实现了 RunShell 的 HTTP API 服务。
// 本文件实现了交互式终端的 WebSocket 处理。
package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/iamlongalong/runshell/pkg/log"
	"github.com/iamlongalong/runshell/pkg/types"
)

var defaultTerminal = &Terminal{
	Type: "xterm-256color",
	Rows: 24,
	Cols: 80,
	Raw:  true,
}

type Terminal struct {
	Type string `json:"type"`
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
	Raw  bool   `json:"raw"`
}

// InteractiveRequest 表示交互式命令请求
type InteractiveRequest struct {
	Command  string            `json:"command" binding:"required"`
	Args     []string          `json:"args,omitempty"`
	WorkDir  string            `json:"workdir,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	Terminal *Terminal         `json:"terminal"`
}

// WSMessage 表示 WebSocket 消息
type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ResizeMessage 表示终端大小调整消息
type ResizeMessage struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// handleInteractiveExec 处理交互命令执行
func (s *Server) handleInteractiveExec(c *gin.Context) {
	// 升级到 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.handleError(c, http.StatusInternalServerError, err, "Failed to upgrade connection")
		return
	}
	defer conn.Close()

	// 读取初始请求
	var req InteractiveRequest // 直接运行 bash
	// if err := conn.ReadJSON(&req); err != nil {
	// 	s.handleWSError(conn, fmt.Errorf("failed to read request: %w", err))
	// 	return
	// }

	log.Info("Received interactive request: %+v", req)

	// 如果没有指定命令，默认使用 bash
	if req.Command == "" {
		req.Command = "bash"
	}

	// 如果没有指定终端，使用默认终端
	if req.Terminal == nil {
		req.Terminal = defaultTerminal
	}

	// 创建执行器
	executor, err := s.executorBuilder.Build(&types.ExecuteOptions{
		WorkDir: req.WorkDir,
		Env:     req.Env,
		TTY:     true,
	})
	if err != nil {
		s.handleWSError(conn, fmt.Errorf("failed to create executor: %w", err))
		return
	}

	// 创建管道用于 IO 通信
	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()

	// 创建错误通道
	errCh := make(chan error, 2)

	// 创建交互式上下文
	ctx := &types.ExecuteContext{
		Context:     c.Request.Context(),
		Interactive: true,
		Command: types.Command{
			Command: req.Command,
			Args:    req.Args,
		},
		InteractiveOpts: &types.InteractiveOptions{
			TerminalType: req.Terminal.Type,
			Rows:         req.Terminal.Rows,
			Cols:         req.Terminal.Cols,
			Raw:          req.Terminal.Raw,
		},
		Options: &types.ExecuteOptions{
			WorkDir: req.WorkDir,
			Env:     req.Env,
			TTY:     true,
			Stdin:   stdinR,
			Stdout:  stdoutW,
			Stderr:  stdoutW,
		},
		Executor: executor,
	}

	// 处理 WebSocket 输入
	go func() {
		defer stdinW.Close()

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Error("WebSocket read error: %v", err)
					errCh <- err
				}
				return
			}

			fmt.Printf("Input received: messageType=%d, data=%q\n", messageType, message)

			// 确保命令以换行符结束
			if len(message) > 0 && !bytes.HasSuffix(message, []byte("\n")) {
				message = append(message, '\n')
			}

			// 写入到标准输入
			if _, err := stdinW.Write(message); err != nil {
				log.Error("Failed to write to stdin: %v", err)
				errCh <- err
				return
			}

			fmt.Printf("Command sent to PTY: %q\n", message)
		}
	}()

	// 处理命令输出
	go func() {
		defer stdoutR.Close()

		buffer := make([]byte, 32*1024)
		for {
			n, err := stdoutR.Read(buffer)
			if err != nil {
				if err != io.EOF {
					log.Error("Failed to read from stdout: %v", err)
					errCh <- err
				}
				return
			}

			// 只有当确实读取到数据时才发送
			if n > 0 {
				output := buffer[:n]
				fmt.Printf("Command output: len=%d, data=%q\n", n, output)

				// 发送输出到客户端
				if err := conn.WriteMessage(websocket.BinaryMessage, output); err != nil {
					log.Error("Failed to write to websocket: %v", err)
					errCh <- err
					return
				}
			}
		}
	}()

	// 执行命令
	result, err := executor.Execute(ctx)
	if err != nil {
		s.handleWSError(conn, fmt.Errorf("command execution failed: %w", err))
		return
	}

	// 发送执行结果
	if err := conn.WriteJSON(ExecResponse{
		ExitCode: result.ExitCode,
		Output:   result.Output,
	}); err != nil {
		s.handleWSError(conn, fmt.Errorf("failed to send result: %w", err))
		return
	}
}

// handleWSError 处理 WebSocket 错误
func (s *Server) handleWSError(conn *websocket.Conn, err error) {
	log.Error("WebSocket error: %v", err)
	conn.WriteJSON(ErrorResponse{
		Error: err.Error(),
	})
}
