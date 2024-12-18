package types

import (
	"context"
	"io"
	"time"
)

// ExecuteOptions 定义命令执行的选项。
// 包含工作目录、环境变量、超时设置、输入输出流等配置。
// swagger:model
type ExecuteOptions struct {
	// WorkDir 指定命令执行的工作目录，所有相对路径都相对于此目录
	WorkDir string `json:"workdir,omitempty"`

	// Env 指定命令执行时的环境变量
	Env map[string]string `json:"env,omitempty"`

	// Timeout 指定命令执行的超时时间（纳秒）
	// swagger:strfmt int64
	Timeout int64 `json:"timeout,omitempty" example:"30000000000"` // 30 seconds in nanoseconds

	// Stdin 指定命令的标准输入流
	Stdin io.Reader `json:"-"`

	// Stdout 指定命令的标准输出流
	Stdout io.Writer `json:"-"`

	// Stderr 指定命令的标准错误流
	Stderr io.Writer `json:"-"`

	// User 指定执行命令的用户信息
	User *User `json:"user,omitempty"`

	// Metadata 存储额外的元数据信息
	Metadata map[string]string `json:"metadata,omitempty"`

	// TTY 是否分配伪终端
	TTY bool `json:"tty,omitempty"`

	// Shell 指定执行命令的 shell, 默认使用 /bin/bash
	Shell string `json:"shell,omitempty"`
}

// Merge 合并两个执行选项, 用于处理默认选项和用自定义选项
func (opts *ExecuteOptions) Merge(other *ExecuteOptions) *ExecuteOptions {
	if other == nil {
		if opts == nil {
			return &ExecuteOptions{
				Env:      make(map[string]string),
				Metadata: make(map[string]string),
			}
		}
		return opts
	}

	// 如果当前选项为空，创建一个新的选项
	if opts == nil {
		result := &ExecuteOptions{
			WorkDir:  other.WorkDir,
			Timeout:  other.Timeout,
			Stdin:    other.Stdin,
			Stdout:   other.Stdout,
			Stderr:   other.Stderr,
			User:     other.User,
			Env:      make(map[string]string),
			Metadata: make(map[string]string),
		}

		// 复制环境变量
		if other.Env != nil {
			for k, v := range other.Env {
				result.Env[k] = v
			}
		}

		// 复制元数据
		if other.Metadata != nil {
			for k, v := range other.Metadata {
				result.Metadata[k] = v
			}
		}

		return result
	}

	// 创建新的选项实例
	result := &ExecuteOptions{
		WorkDir:  opts.WorkDir,
		Timeout:  opts.Timeout,
		Stdin:    opts.Stdin,
		Stdout:   opts.Stdout,
		Stderr:   opts.Stderr,
		User:     opts.User,
		Env:      make(map[string]string),
		Metadata: make(map[string]string),
	}

	// 复制当前选项��环境变量
	if opts.Env != nil {
		for k, v := range opts.Env {
			result.Env[k] = v
		}
	}

	// 复制当前选项的元数据
	if opts.Metadata != nil {
		for k, v := range opts.Metadata {
			result.Metadata[k] = v
		}
	}

	// 合并其他选项的值
	if other.WorkDir != "" {
		result.WorkDir = other.WorkDir
	}

	// 合并环境变量
	if other.Env != nil {
		for k, v := range other.Env {
			if _, ok := result.Env[k]; !ok {
				result.Env[k] = v
			}
		}
	}

	if other.Timeout != 0 {
		result.Timeout = other.Timeout
	}

	if other.Stdin != nil {
		result.Stdin = other.Stdin
	}

	if other.Stdout != nil {
		result.Stdout = other.Stdout
	}

	if other.Stderr != nil {
		result.Stderr = other.Stderr
	}

	if other.User != nil {
		result.User = other.User
	}

	// 合并元数据
	if other.Metadata != nil {
		for k, v := range other.Metadata {
			if _, ok := result.Metadata[k]; !ok {
				result.Metadata[k] = v
			}
		}
	}

	return result
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

	// Interactive 是否是交互式命令
	Interactive bool

	// InteractiveOpts 交互式选项
	InteractiveOpts *InteractiveOptions
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
// swagger:model
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
// swagger:model
type ResourceUsage struct {
	// CPUTime 是 CPU 使用时间（纳秒）
	// swagger:strfmt int64
	CPUTime int64 `json:"cpu_time" example:"1000000000"` // 1 second in nanoseconds

	// MemoryUsage 是内存使用量（字节）
	MemoryUsage int64 `json:"memory_usage" example:"1048576"` // 1MB in bytes

	// IORead 是 IO 读取量（字节）
	IORead int64 `json:"io_read" example:"4096"` // 4KB in bytes

	// IOWrite 是 IO 写入量（字节）
	IOWrite int64 `json:"io_write" example:"4096"` // 4KB in bytes
}

// User 表示执行命令的用户信息。
// swagger:model
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
// swagger:model
type CommandInfo struct {
	Name        string            `json:"name" example:"ls"`                    // 命令名称
	Description string            `json:"description" example:"List directory"` // 命令描述
	Usage       string            `json:"usage" example:"ls [options] [path]"`  // 命令用法
	Category    string            `json:"category,omitempty"`                   // 命令类
	Metadata    map[string]string `json:"metadata,omitempty"`                   // 命令元数据
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

	// Pattern 按模式匹配滤
	Pattern string
}

// Executor 定义了命令执行器的接口
type Executor interface {
	// Name 返回执行器名称
	Name() string

	// Execute 执行命令 (包含内置命令代理)
	Execute(ctx *ExecuteContext) (*ExecuteResult, error)

	// ExecuteCommand 直接执行命令
	ExecuteCommand(ctx *ExecuteContext) (*ExecuteResult, error)

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

// Session 表示一个执行会话
// swagger:model
type Session struct {
	ID             string            `json:"id" example:"sess_123"` // 会话的唯一标识符
	Options        *ExecuteOptions   `json:"options,omitempty"`     // 会话的执行选项
	CreatedAt      time.Time         `json:"created_at"`            // 会话创建时间
	LastAccessedAt time.Time         `json:"last_accessed_at"`      // 最后访问时间
	Metadata       map[string]string `json:"metadata,omitempty"`    // 会话相关的元数据
	Status         string            `json:"status"`                // 会话状态

	// 以下字不会在 JSON 中序列化
	Executor Executor           `json:"-"` // 会话使用的执行器
	Context  context.Context    `json:"-"` // 会话的上下文
	Cancel   context.CancelFunc `json:"-"` // 用于取消会话的函数
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

// SessionRequest 表示创建会话的请求
// swagger:model
type SessionRequest struct {
	ExecutorType string            `json:"executor_type,omitempty"` // 执行器类型（local/docker）
	DockerConfig *DockerConfig     `json:"docker_config,omitempty"` // Docker 执行器配置
	LocalConfig  *LocalConfig      `json:"local_config,omitempty"`  // 本地执行器配置
	Options      *ExecuteOptions   `json:"options,omitempty"`       // 执行选项
	Metadata     map[string]string `json:"metadata,omitempty"`      // 会话元数据
}

// SessionResponse 表示会话操作的响应
// swagger:model
type SessionResponse struct {
	Session *Session `json:"session"`         // 会话信息
	Error   string   `json:"error,omitempty"` // 错误信息
}

// CommandExecution 表示命令执行记录
type CommandExecution struct {
	ID        string    // 命令执行的唯一标识符
	Command   Command   // 命令
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
	UseBuiltinCommands        bool   // 是否使用内置命令
}

// LocalConfig 本地执行器配置
type LocalConfig struct {
	AllowUnregisteredCommands bool   // 是否允许执行未注册的命令
	UseBuiltinCommands        bool   // 是否使用内置命令
	WorkDir                   string // 工作目录
}

// ExecutorBuilder 定义了执行器构建器的接口。
// 用于创建新的执行器实例。
type ExecutorBuilder interface {
	// Build 创建并返回一个新的执行器实例。
	// 每次调用都应该返回一个独立的执行器实例。
	Build(options *ExecuteOptions) (Executor, error)
}

// BuiltinCommandProvider 定义了内置命令提供者的接口。
// 用于提供内置命令的实现。
type BuiltinCommandProvider interface {
	// GetCommands 回所有内置命令。
	GetCommands() []ICommand
}

// ExecutorBuilderFunc 是一个便捷的函数类型，实现了 ExecutorBuilder 接口。
type ExecutorBuilderFunc func(options *ExecuteOptions) (Executor, error)

// Build 实现 ExecutorBuilder 接口。
func (f ExecutorBuilderFunc) Build(options *ExecuteOptions) (Executor, error) {
	return f(options)
}

// InteractiveOptions 定义交互式命令的选项
type InteractiveOptions struct {
	// TerminalType 是终端类型 (e.g., "xterm")
	TerminalType string `json:"terminal_type,omitempty"`

	// Rows 是终端行数
	Rows uint16 `json:"rows,omitempty"`

	// Cols 是终端列数
	Cols uint16 `json:"cols,omitempty"`

	// Raw 是否使用原始模式
	Raw bool `json:"raw,omitempty"`
}
