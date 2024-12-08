package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CommandExecution 表示命令执行记录
type CommandExecution struct {
	Command   string    // 命令名称
	Args      []string  // 命令参数
	StartTime time.Time // 开始时间
	EndTime   time.Time // 结束时间
	ExitCode  int       // 退出码
	Error     error     // 错误信息
	Status    string    // 执行状态
}

// Auditor 审计器
type Auditor struct {
	logFile string
}

// NewAuditor 创建新的审计器
func NewAuditor(logFile string) (*Auditor, error) {
	// 确保日志目录存在
	if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %v", err)
	}

	return &Auditor{
		logFile: logFile,
	}, nil
}

// LogCommandExecution 记录命令执行
func (a *Auditor) LogCommandExecution(exec *CommandExecution) error {
	// 准备日志内容
	logEntry := fmt.Sprintf("[%s] Command: %s, Args: %v, Status: %s, ExitCode: %d",
		exec.StartTime.Format(time.RFC3339),
		exec.Command,
		exec.Args,
		exec.Status,
		exec.ExitCode,
	)

	if exec.Error != nil {
		logEntry += fmt.Sprintf(", Error: %v", exec.Error)
	}

	logEntry += "\n"

	// 写入日志文件
	f, err := os.OpenFile(a.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit log file: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(logEntry); err != nil {
		return fmt.Errorf("failed to write audit log: %v", err)
	}

	return nil
}
