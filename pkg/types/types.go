package types

import (
	"context"
	"io"
	"time"
)

// ExecuteOptions 定义命令执行的选项。
// 包含工作目录、环境变量、超时设置、输入输出流等配置。
type ExecuteOptions struct {
	// WorkDir 指定命令执行的工作目录，所有相对路径都相对于此目录
	WorkDir string

	// Env 指定命令执行时的环境变量
	Env map[string]string

	// Timeout 指定命令执行的超时时间
	Timeout time.Duration

	// Stdin 指定命令的标准输入流
	Stdin io.Reader

	// Stdout 指定命令的标准输出流
	Stdout io.Writer

	// Stderr 指定命令的标准错误流
	Stderr io.Writer

	// User 指定执行命令的用户信息
	User *User

	// Metadata 存储额外的元数据信息
	Metadata map[string]string
}

// ExecuteContext 包含命令执行的上下文信息。
// 提供了命令执行时需要的所有上下文数据。
type ExecuteContext struct {
	// Context 是标准的 context.Context 实例
	Context context.Context

	// Args 是命令的参数列表
	Args []string

	// Options 是执行选项
	Options *ExecuteOptions

	// StartTime 是命令开始执行的时间
	StartTime time.Time

	// Command 是要执行的命令名称
	Command string

	// Input 是命令的输入数据
	Input io.Reader
}

// ExecuteResult 表示命令执行的结果。
// 包含执行状态、输出、错误信息等。
type ExecuteResult struct {
	// CommandName 是执行的命令名称
	CommandName string

	// ExitCode 是命令的退出码
	ExitCode int

	// StartTime 是命令开始执行的时间
	StartTime time.Time

	// EndTime 是命令结束执行的时间
	EndTime time.Time

	// Error 是执行过程中的错误信息
	Error error

	// ResourceUsage 记录资源使用情况
	ResourceUsage ResourceUsage

	// Output 是命令的输出内容
	Output string
}

// ResourceUsage 记录命令执行过程中的资源使用情况。
type ResourceUsage struct {
	// CPUTime 是 CPU 使用时间
	CPUTime time.Duration

	// MemoryUsage 是内存使用量（字节）
	MemoryUsage int64

	// IORead 是 IO 读取量（字节）
	IORead int64

	// IOWrite 是 IO 写入量（字节）
	IOWrite int64
}

// User 表示执行命令的用户信息。
type User struct {
	// Username 是用户名
	Username string

	// UID 是用户 ID
	UID int

	// GID 是用户组 ID
	GID int

	// Groups 是用户所属的附加组 ID 列表
	Groups []int
}

// Command 表示一个已注册的命令。
type Command struct {
	// Name 是命令的名称
	Name string

	// Description 是命令的描述信息
	Description string

	// Usage 是命令的使用说明
	Usage string

	// Category 是命令的分类
	Category string

	// Handler 是命令的处理器
	Handler CommandHandler

	// Metadata 是命令的元数据
	Metadata map[string]string
}

// CommandHandler 定义命令处理器的接口。
// 所有自定义命令都需要实现此接口。
type CommandHandler interface {
	// Execute 执行命令并返回结果
	Execute(ctx *ExecuteContext) (*ExecuteResult, error)
}

// CommandFilter 定义命令过滤器。
// 用于在列出命令时进行过滤。
type CommandFilter struct {
	// Category 按分类过滤
	Category string

	// Pattern 按模式匹配过滤
	Pattern string
}

// Executor 定义命令执行器的接口。
// 提供命令执行、管理等核心功能。
type Executor interface {
	// Execute 执行指定的命令
	Execute(ctx context.Context, cmdName string, args []string, opts *ExecuteOptions) (*ExecuteResult, error)

	// GetCommandInfo 获取命令的信息
	GetCommandInfo(cmdName string) (*Command, error)

	// GetCommandHelp 获取命令的帮助信息
	GetCommandHelp(cmdName string) (string, error)

	// ListCommands 列出符合过滤条件的命令
	ListCommands(filter *CommandFilter) ([]*Command, error)

	// RegisterCommand 注册新命令
	RegisterCommand(cmd *Command) error

	// UnregisterCommand 注销已注册的命令
	UnregisterCommand(cmdName string) error
}

// ErrCommandNotFound 表示命令未找到
var ErrCommandNotFound = NewExecuteError("command not found", "COMMAND_NOT_FOUND")

// ErrCommandExecutionFailed 表示命令执行失败
var ErrCommandExecutionFailed = NewExecuteError("command execution failed", "EXECUTION_FAILED")

// ExecuteError 定义执行错误的类型。
// 包含错误消息和错误代码。
type ExecuteError struct {
	// Message 是错误消息
	Message string

	// Code 是错误代码
	Code string
}

// Error 实现 error 接口
func (e *ExecuteError) Error() string {
	return e.Message
}

// NewExecuteError 创建新的执行错误
func NewExecuteError(message, code string) *ExecuteError {
	return &ExecuteError{
		Message: message,
		Code:    code,
	}
}

// GetTimeNow 返回当前时间
// 主要用于测试时的时间模拟
func GetTimeNow() time.Time {
	return time.Now()
}
