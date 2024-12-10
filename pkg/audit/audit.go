package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/iamlongalong/runshell/pkg/types"
)

// FileAuditor 实现基于文件的审计器
type FileAuditor struct {
	logFile string
}

// NewFileAuditor 创建新的文件审计器
func NewFileAuditor(logFile string) (*FileAuditor, error) {
	// 确保日志目录存在
	if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %v", err)
	}

	return &FileAuditor{
		logFile: logFile,
	}, nil
}

// LogCommandExecution 实现基于文件的命令执行记录
func (a *FileAuditor) LogCommandExecution(exec *types.CommandExecution) error {
	// 准备日志内容
	logEntry := fmt.Sprintf("[%s] Command: %s, Args: %v, Status: %s, ExitCode: %d",
		exec.StartTime.Format(time.RFC3339),
		exec.Command.Command,
		exec.Command.Args,
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

// ConsoleAuditor 实现基于控制台的审计器
type ConsoleAuditor struct {
	timeFormat string
}

// NewConsoleAuditor 创建新的控制台审计器
func NewConsoleAuditor() *ConsoleAuditor {
	return &ConsoleAuditor{
		timeFormat: "2006-01-02 15:04:05.000",
	}
}

// LogCommandExecution 实现基于控制台的命令执行记录
func (a *ConsoleAuditor) LogCommandExecution(exec *types.CommandExecution) error {
	// 格式化时间
	var startTime, endTime string
	if !exec.StartTime.IsZero() {
		startTime = exec.StartTime.Format(a.timeFormat)
	}
	if !exec.EndTime.IsZero() {
		endTime = exec.EndTime.Format(a.timeFormat)
	}

	// 打印审计信息
	fmt.Printf("\n=== Audit Log ===\n")
	fmt.Printf("Command:    %s\n", exec.Command)
	fmt.Printf("Arguments:  %v\n", exec.Command.Args)
	fmt.Printf("Status:     %s\n", exec.Status)
	fmt.Printf("Start Time: %s\n", startTime)
	fmt.Printf("End Time:   %s\n", endTime)
	if exec.Error != nil {
		fmt.Printf("Error:      %v\n", exec.Error)
	}
	if exec.ExitCode != 0 {
		fmt.Printf("Exit Code:  %d\n", exec.ExitCode)
	}
	fmt.Println("===============")

	return nil
}

// MultiAuditor 实现多重审计器
type MultiAuditor struct {
	auditors []types.Auditor
}

// NewMultiAuditor 创建新的多重审计器
func NewMultiAuditor(auditors ...types.Auditor) *MultiAuditor {
	return &MultiAuditor{
		auditors: auditors,
	}
}

// LogCommandExecution 实现多重审计记录
func (a *MultiAuditor) LogCommandExecution(exec *types.CommandExecution) error {
	var lastErr error
	for _, auditor := range a.auditors {
		if err := auditor.LogCommandExecution(exec); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
