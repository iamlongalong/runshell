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

// Merge 合并两个执行选项, 用于处理默认选项和用户自定义选项
func (opts *ExecuteOptions) Merge(other *ExecuteOptions) *ExecuteOptions {
	if other == nil {
		return opts
	}

	if opts == nil {
		opts = &ExecuteOptions{}
	}

	if opts.Env == nil {
		opts.Env = make(map[string]string)
	}

	if opts.Metadata == nil {
		opts.Metadata = make(map[string]string)
	}

	if opts.WorkDir == "" {
		opts.WorkDir = other.WorkDir
	}

	for k, v := range other.Env {
		if _, ok := opts.Env[k]; !ok {
			opts.Env[k] = v
		}
	}

	if opts.Timeout == 0 {
		opts.Timeout = other.Timeout
	}

	if opts.Stdin == nil {
		opts.Stdin = other.Stdin
	}

	if opts.Stdout == nil {
		opts.Stdout = other.Stdout
	}

	if opts.Stderr == nil {
		opts.Stderr = other.Stderr
	}

	if opts.User == nil {
		opts.User = other.User
	}

	for k, v := range other.Metadata {
		if _, ok := opts.Metadata[k]; !ok {
			opts.Metadata[k] = v
		}
	}

	return opts
}

// ExecuteContext 包含命令执行的上下文信息。
// 提供了命令执行时需要的所有上下文数据。
type ExecuteContext struct {
	// Context 是标准的 context.Context 实例
	Context context.Context

	// Command 是要执行的命令名称
	Command Command

	// Options 是执行选项
	Options *ExecuteOptions

	// StartTime 是命令开始执行的时间
	StartTime time.Time

	// Input 是命令的输入数据
	Input io.Reader

	// IsPiped 是否是管道命令
	IsPiped bool
	// PipeContext 管道上下文
	PipeContext *PipelineContext

	// Executor 是执行器实例
	Executor Executor
}

func (ctx *ExecuteContext) Copy() *ExecuteContext {
	return &ExecuteContext{
		Context:     ctx.Context,
		Command:     ctx.Command,
		Options:     ctx.Options,
		StartTime:   ctx.StartTime,
		Input:       ctx.Input,
		IsPiped:     ctx.IsPiped,
		PipeContext: ctx.PipeContext,
		Executor:    ctx.Executor,
	}
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

	// Output 是命令的输出
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

// CommandHandler 是 ICommand 的别名，用于保持向后兼容性
type CommandHandler = ICommand

// CommandInfo 表示一个可执行的命令
type CommandInfo struct {
	Name        string            // 命令名称
	Description string            // 命令描述
	Usage       string            // 命令用法
	Category    string            // 命令分类
	Metadata    map[string]string // 命令元数据
}

// ICommand 定义了命令处理器的接口
type ICommand interface {
	Info() CommandInfo
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
	// Name 返回执行器名称
	Name() string

	// Execute 执行命令
	Execute(ctx *ExecuteContext) (*ExecuteResult, error)

	// ListCommands 出所有可用命令
	ListCommands() []CommandInfo

	// Close 关闭执行器，清理资源
	Close() error
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
func GetTimeNow() time.Time {
	return time.Now()
}

// Command 表示命令
type Command struct {
	Command string   // 命令名称
	Args    []string // 命令参数
}

// PipelineContext 表示管道上下文
type PipelineContext struct {
	Context context.Context // 上下文

	Commands []*Command      // 管道中的命令列表
	Options  *ExecuteOptions // 执行选项
}

// Session 表示一个执行会
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

const (
	// ExecutorTypeLocal 表示本地执行器
	ExecutorTypeLocal = "local"
	// ExecutorTypeDocker 表示 Docker 执行器
	ExecutorTypeDocker = "docker"
)

// SessionRequest 表示创会话的请求
type SessionRequest struct {
	// ExecutorType 是执行器类型（local/docker）
	ExecutorType string `json:"executor_type"`

	// DockerConfig 是 Docker 执行器配置
	DockerConfig *DockerConfig `json:"docker_config,omitempty"`

	// LocalConfig 是本地执行器配置
	LocalConfig *LocalConfig `json:"local_config,omitempty"`

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

// CommandExecution 表示命令执行记录
type CommandExecution struct {
	Command Command // 命令

	StartTime time.Time // 开始时间
	EndTime   time.Time // 结束时间
	ExitCode  int       // 退出码
	Error     error     // 错误信息
	Status    string    // 执行状态
}

// Auditor 定义审计器接口
type Auditor interface {
	// LogCommandExecution 记录命令执行
	LogCommandExecution(exec *CommandExecution) error
}

// DockerConfig 表示 Docker 执行器的配置
type DockerConfig struct {
	Image                     string // Docker 镜像
	WorkDir                   string // 工作目录
	User                      string // 用户
	BindMount                 string // 目录绑定
	AllowUnregisteredCommands bool   // 是否允许执行未注册的命令
}

// LocalConfig 本地执行器配置
type LocalConfig struct {
	AllowUnregisteredCommands bool // 是否允许执行未注册的命令
	UseBuiltinCommands        bool // 是否使用内置命令
}

// ExecutorBuilder 定义了执行器构建器的接口。
// 用于创建新的执行器实例。
type ExecutorBuilder interface {
	// Build 创建并返回一个新的执行器实例。
	// 每次调用都应该返回一个独立的执行器实例。
	Build() (Executor, error)
}

// BuiltinCommandProvider 定义了内置命令提供者的接口。
// 用于提供内置命令的实现。
type BuiltinCommandProvider interface {
	// GetCommands 返回所有内置命令。
	GetCommands() []ICommand
}

// ExecutorBuilderFunc 是一个便捷的函数类型，实现了 ExecutorBuilder 接口。
type ExecutorBuilderFunc func() (Executor, error)

// Build 实现 ExecutorBuilder 接口。
func (f ExecutorBuilderFunc) Build() (Executor, error) {
	return f()
}
