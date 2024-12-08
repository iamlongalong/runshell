package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/iamlongalong/runshell/pkg/types"
)

// AuditEvent 表示一个审计事件。
// 包含命令执行的完整信息，包括时间戳、命令名称、参数等。
type AuditEvent struct {
	// Timestamp 是事件发生的时间
	Timestamp time.Time `json:"timestamp"`
	// CommandName 是执行的命令名称
	CommandName string `json:"command_name"`
	// Args 是命令的参数列表
	Args []string `json:"args"`
	// WorkDir 是命令执行时的工作目录
	WorkDir string `json:"work_dir"`
	// User 是执行命令的用户名
	User string `json:"user"`
	// ExitCode 是命令的退出码
	ExitCode int `json:"exit_code"`
	// Error 是执行过程中的错误信息（如果有）
	Error string `json:"error,omitempty"`
	// Duration 是命令执行的持续时间
	Duration time.Duration `json:"duration"`
	// Metadata 是额外的元数据信息
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Auditor 处理审计日志记录。
// 提供了日志写入、查询和搜索功能。
type Auditor struct {
	// auditDir 是审计日志文件存储的目录
	auditDir string
	// logFile 是当前打开的日志文件
	logFile *os.File
	// mu 用于保护并发访问
	mu sync.Mutex
}

// NewAuditor 创建一个新的审计器实例。
// auditDir 参数指定审计日志文件的存储目录。
// 如果目录不存在，会自动创建。
func NewAuditor(auditDir string) (*Auditor, error) {
	// 确保审计目录存在
	if err := os.MkdirAll(auditDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit directory: %v", err)
	}

	// 创建当天的日志文件
	logPath := filepath.Join(auditDir, time.Now().Format("2006-01-02")+".log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %v", err)
	}

	return &Auditor{
		auditDir: auditDir,
		logFile:  logFile,
	}, nil
}

// LogCommandExecution 记录命令执行的详细信息。
// ctx 包含命令执行的上下文信息。
// result 包含命令执行的结果。
func (a *Auditor) LogCommandExecution(ctx *types.ExecuteContext, result *types.ExecuteResult) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	event := AuditEvent{
		Timestamp:   result.StartTime,
		CommandName: result.CommandName,
		Args:        ctx.Args,
		WorkDir:     ctx.Options.WorkDir,
		ExitCode:    result.ExitCode,
		Duration:    result.EndTime.Sub(result.StartTime),
		Metadata:    ctx.Options.Metadata,
	}

	if result.Error != nil {
		event.Error = result.Error.Error()
	}

	if ctx.Options.User != nil {
		event.User = ctx.Options.User.Username
	} else {
		currentUser, err := os.Getwd()
		if err == nil {
			event.User = currentUser
		}
	}

	// 序列化事件
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %v", err)
	}

	// 写入日志文件
	if _, err := a.logFile.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write audit log: %v", err)
	}

	return nil
}

// Close 关闭审计日志文件。
// 在程序退出前应该调用此方法以确保日志正确写入。
func (a *Auditor) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.logFile.Close()
}

// GetAuditLogs 获取指定时间范围内的审计日志。
// start 和 end 参数指定查询的时间范围。
// 返回符合时间范围的审计事件列表。
func (a *Auditor) GetAuditLogs(start, end time.Time) ([]AuditEvent, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var events []AuditEvent

	// 遍历日志文件
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		logPath := filepath.Join(a.auditDir, d.Format("2006-01-02")+".log")

		data, err := os.ReadFile(logPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("failed to read audit log %s: %v", logPath, err)
		}

		// 解析每一行
		for _, line := range strings.Split(string(data), "\n") {
			if line == "" {
				continue
			}

			var event AuditEvent
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				return nil, fmt.Errorf("failed to parse audit log entry: %v", err)
			}

			if event.Timestamp.After(start) && event.Timestamp.Before(end) {
				events = append(events, event)
			}
		}
	}

	return events, nil
}

// SearchAuditLogs 根据过滤条件搜索审计日志。
// filters 参数是一个键值对映射，支持以下过滤条件：
// - command: 命令名称
// - user: 用户名
// - exit_code: 退出码
// - error: 是否有错误（"true"表示只返回有错误的记录）
func (a *Auditor) SearchAuditLogs(filters map[string]string) ([]AuditEvent, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var events []AuditEvent

	// 获取所有日志文件
	files, err := filepath.Glob(filepath.Join(a.auditDir, "*.log"))
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %v", err)
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read audit log %s: %v", file, err)
		}

		// 解析每一行
		for _, line := range strings.Split(string(data), "\n") {
			if line == "" {
				continue
			}

			var event AuditEvent
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				return nil, fmt.Errorf("failed to parse audit log entry: %v", err)
			}

			// 应用过滤器
			match := true
			for k, v := range filters {
				switch k {
				case "command":
					if event.CommandName != v {
						match = false
					}
				case "user":
					if event.User != v {
						match = false
					}
				case "exit_code":
					if fmt.Sprintf("%d", event.ExitCode) != v {
						match = false
					}
				case "error":
					if v == "true" && event.Error == "" {
						match = false
					}
				}
			}

			if match {
				events = append(events, event)
			}
		}
	}

	return events, nil
}
