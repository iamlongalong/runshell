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

	// IsPiped 是否是管道命令
	IsPiped bool

	// PipeInput 管道输入
	PipeInput io.Reader

	// PipeOutput 管道输出
	PipeOutput io.Writer

	// PipeContext 管道上下文
	PipeContext *PipelineContext

	// Executor 是执行器实例
	Executor Executor
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

// Command 表示一个可执行的命令
type Command struct {
	Name        string            // 命令名称
	Description string            // 命令描述
	Usage       string            // 命令用法
	Category    string            // 命令分类
	Metadata    map[string]string // 命令元数据
	Handler     CommandHandler    // 命令处理器
}

// Execute 执行命令
func (c *Command) Execute(ctx *ExecuteContext) (*ExecuteResult, error) {
	return c.Handler.Execute(ctx)
}

// CommandHandler 定义了命令处理器的接口
type CommandHandler interface {
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

// Executor 定义了命令执行器的接口
type Executor interface {
	// Execute 执行命令
	Execute(ctx *ExecuteContext) (*ExecuteResult, error)

	// ListCommands 列出所有可用命令
	ListCommands() []CommandInfo

	// Close 关闭执行器，清理资源
	Close() error

	// SetOptions 设置执行选项
	SetOptions(options *ExecuteOptions)
}

// CommandInfo 表示命令信息
type CommandInfo struct {
	Name        string            // 命令名称
	Description string            // 命令描述
	Usage       string            // 命令用法
	Category    string            // 命令分��
	Metadata    map[string]string // 命令元数据
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

// PipeCommand 表示管道命令
type PipeCommand struct {
	Command string   // 命令名称
	Args    []string // 命令参数
}

// PipelineContext 表示管道上下文
type PipelineContext struct {
	Commands []*PipeCommand  // 管道中的命令列表
	Options  *ExecuteOptions // 执行选项
	Context  context.Context // 上下文
}

// Session 表示一个执行会话
type Session struct {
	// ID 是会话的唯一标识符
	ID string `json:"id"`

	// Executor 是会话使用的执行器
	Executor Executor `json:"-"`

	// Options 是会话的执行选项
	Options *ExecuteOptions `json:"options,omitempty"`

	// Context 是会话的上下文
	Context context.Context `json:"-"`

	// Cancel 是用于取消会话的函数
	Cancel context.CancelFunc `json:"-"`

	// CreatedAt 是会话创建时间
	CreatedAt time.Time `json:"created_at"`

	// LastAccessedAt 是最后访问时间
	LastAccessedAt time.Time `json:"last_accessed_at"`

	// Metadata 存储会话相关的元数据
	Metadata map[string]string `json:"metadata,omitempty"`

	// Status 是会话状态
	Status string `json:"status"`
}

// SessionManager 定义了会话管理器的接口
type SessionManager interface {
	// CreateSession 创建新的会话
	CreateSession(executor Executor, options *ExecuteOptions) (*Session, error)

	// GetSession 获取会话
	GetSession(id string) (*Session, error)

	// ListSessions 列出所有会话
	ListSessions() ([]*Session, error)

	// DeleteSession 删除会话
	DeleteSession(id string) error

	// UpdateSession 更新会话
	UpdateSession(session *Session) error
}

// SessionRequest 表示创建会话的请求
type SessionRequest struct {
	// ExecutorType 是执行器类型（local/docker）
	ExecutorType string `json:"executor_type"`

	// DockerImage 是 Docker 执行器使用的镜像
	DockerImage string `json:"docker_image,omitempty"`

	// Options 是执行选项
	Options *ExecuteOptions `json:"options,omitempty"`

	// Metadata 是会话元数据
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SessionResponse 表示会话操作的响应
type SessionResponse struct {
	// Session 是会话信息
	Session *Session `json:"session"`

	// Error 是错误信息
	Error string `json:"error,omitempty"`
}

// ExecRequest 表示执行命令的请求
type ExecRequest struct {
	// Command 是要执行的命令
	Command string `json:"command"`

	// Args 是命令的参数
	Args []string `json:"args"`

	// WorkDir 是工作目录
	WorkDir string `json:"work_dir,omitempty"`

	// Env 是环境变量
	Env map[string]string `json:"env,omitempty"`
}
