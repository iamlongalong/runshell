package types

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// MockFile represents a mock file in memory
type MockFile struct {
	Content string
	Exists  bool
	IsDir   bool
	Mode    os.FileMode
}

// MockExecutor 是一个用于测试的执行器实现。
type MockExecutor struct {
	ExecuteFunc func(ctx *ExecuteContext) (*ExecuteResult, error)
	commands    sync.Map
	files       sync.Map // stores MockFile objects
}

// NewMockExecutor 创建一个新的模拟执行器。
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{}
}

// Name 返回执行器名称。
func (m *MockExecutor) Name() string {
	return "mock"
}

// ExecuteCommand 直接执行命令
func (m *MockExecutor) ExecuteCommand(ctx *ExecuteContext) (*ExecuteResult, error) {
	return m.Execute(ctx)
}

// Execute 执行命令。
func (m *MockExecutor) Execute(ctx *ExecuteContext) (*ExecuteResult, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx)
	}

	// 默认的执行逻辑
	cmd := ctx.Command.Command
	args := ctx.Command.Args

	// 如果没有命令，返回错误
	if cmd == "" && len(args) == 0 {
		return &ExecuteResult{
			CommandName: "",
			ExitCode:    1,
			Error:       fmt.Errorf("no command specified"),
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
		}, fmt.Errorf("no command specified")
	}

	// 如果命令在参数中，则使用第一个参数作为命令
	if cmd == "" && len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}

	// 处理无效命令
	if !contains([]string{"ls", "cat", "mkdir", "rm", "cp", "pwd", "ps", "top", "df", "git", "go", "wget", "tar", "zip", "kill", "python", "node", "bash"}, cmd) {
		return &ExecuteResult{
			CommandName: cmd,
			ExitCode:    127,
			Error:       fmt.Errorf("command not found: %s", cmd),
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
		}, fmt.Errorf("command not found: %s", cmd)
	}

	// 处理特定命令
	switch cmd {
	case "ls":
		// 列出目录内容
		path := "/"
		if len(args) > 0 {
			path = args[0]
		}

		var output strings.Builder
		var found bool
		m.files.Range(func(key, value interface{}) bool {
			filePath := key.(string)
			if path == "/" || strings.HasPrefix(filePath, path) {
				found = true
				// 如果是根目录，只显示文件名，否则显示完整路径
				if path == "/" {
					output.WriteString(strings.TrimPrefix(filePath, "/") + "\n")
				} else {
					output.WriteString(filePath + "\n")
				}
			}
			return true
		})

		if !found && path != "/" {
			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("no such directory: %s", path),
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, fmt.Errorf("no such directory: %s", path)
		}

		if ctx.Options != nil && ctx.Options.Stdout != nil {
			ctx.Options.Stdout.Write([]byte(output.String()))
		}

		return &ExecuteResult{
			CommandName: cmd,
			ExitCode:    0,
			Output:      output.String(),
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
		}, nil

	case "cat":
		if len(args) == 0 {
			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("no file specified"),
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, fmt.Errorf("no file specified")
		}

		path := args[0]
		if value, ok := m.files.Load(path); ok {
			file := value.(*MockFile)
			if file.IsDir {
				return &ExecuteResult{
					CommandName: cmd,
					ExitCode:    1,
					Error:       fmt.Errorf("is a directory: %s", path),
					StartTime:   ctx.StartTime,
					EndTime:     time.Now(),
				}, fmt.Errorf("is a directory: %s", path)
			}

			if ctx.Options != nil && ctx.Options.Stdout != nil {
				ctx.Options.Stdout.Write([]byte(file.Content))
			}

			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    0,
				Output:      file.Content,
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, nil
		}

		return &ExecuteResult{
			CommandName: cmd,
			ExitCode:    1,
			Error:       fmt.Errorf("no such file: %s", path),
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
		}, fmt.Errorf("no such file: %s", path)

	case "mkdir":
		if len(args) == 0 {
			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("no directory specified"),
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, fmt.Errorf("no directory specified")
		}

		path := args[0]
		m.files.Store(path, &MockFile{
			Content: "",
			Exists:  true,
			IsDir:   true,
			Mode:    0755,
		})

		return &ExecuteResult{
			CommandName: cmd,
			ExitCode:    0,
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
		}, nil

	case "rm":
		if len(args) == 0 {
			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("no file specified"),
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, fmt.Errorf("no file specified")
		}

		path := args[0]
		if _, ok := m.files.Load(path); !ok {
			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("no such file: %s", path),
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, fmt.Errorf("no such file: %s", path)
		}

		m.files.Delete(path)
		return &ExecuteResult{
			CommandName: cmd,
			ExitCode:    0,
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
		}, nil

	case "cp":
		if len(args) < 2 {
			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("cp requires source and destination"),
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, fmt.Errorf("cp requires source and destination")
		}

		src := args[0]
		dst := args[1]

		value, ok := m.files.Load(src)
		if !ok {
			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("no such file: %s", src),
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, fmt.Errorf("no such file: %s", src)
		}

		srcFile := value.(*MockFile)
		m.files.Store(dst, &MockFile{
			Content: srcFile.Content,
			Exists:  true,
			IsDir:   srcFile.IsDir,
			Mode:    srcFile.Mode,
		})

		return &ExecuteResult{
			CommandName: cmd,
			ExitCode:    0,
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
		}, nil

	case "pwd", "ps", "top", "df":
		// 这些命令已经在各自的Command实现中处理
		if command, ok := m.commands.Load(cmd); ok {
			if cmd, ok := command.(ICommand); ok {
				ctx.Executor = m
				return cmd.Execute(ctx)
			}
		}
		return &ExecuteResult{
			CommandName: cmd,
			ExitCode:    0,
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
		}, nil

	case "git":
		if len(args) > 0 && args[0] == "status" && ctx.Options != nil && ctx.Options.WorkDir == "/tmp" {
			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    128,
				Error:       fmt.Errorf("not a git repository"),
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, fmt.Errorf("not a git repository")
		}

	case "go":
		if len(args) > 0 && args[0] == "mod" && ctx.Options != nil && ctx.Options.WorkDir == "/tmp" {
			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("go.mod file not found"),
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, fmt.Errorf("go.mod file not found")
		}

	case "wget":
		if len(args) == 0 {
			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("missing URL"),
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, fmt.Errorf("missing URL")
		}
		if !strings.HasPrefix(args[0], "http") {
			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("invalid URL"),
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, fmt.Errorf("invalid URL")
		}

	case "tar":
		if len(args) == 0 {
			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("missing operand"),
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, fmt.Errorf("missing operand")
		}
		if args[0] == "-invalid" {
			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("invalid option"),
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, fmt.Errorf("invalid option")
		}

	case "zip":
		if len(args) < 2 {
			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("missing file operand"),
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, fmt.Errorf("missing file operand")
		}

	case "kill":
		if len(args) == 0 {
			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("missing operand"),
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, fmt.Errorf("missing operand")
		}
		if args[0] == "12345" {
			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("no such process"),
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, fmt.Errorf("no such process")
		}

	case "python":
		if len(args) == 0 {
			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("missing operand"),
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, fmt.Errorf("missing operand")
		}
		if args[0] == "--version" {
			output := "Python 3.9.0\n"
			if ctx.Options != nil && ctx.Options.Stdout != nil {
				ctx.Options.Stdout.Write([]byte(output))
			}
			return &ExecuteResult{
				CommandName: cmd,
				ExitCode:    0,
				Output:      output,
				StartTime:   ctx.StartTime,
				EndTime:     time.Now(),
			}, nil
		}
	}

	// 默认返回成功
	return &ExecuteResult{
		CommandName: cmd,
		ExitCode:    0,
		StartTime:   ctx.StartTime,
		EndTime:     time.Now(),
	}, nil
}

// contains 检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ListCommands 返回可用的命令列表。
func (m *MockExecutor) ListCommands() []CommandInfo {
	return []CommandInfo{
		{
			Name:        "test",
			Description: "Test command for testing",
			Usage:       "test [args...]",
		},
	}
}

// Close 关闭执行器。
func (m *MockExecutor) Close() error {
	return nil
}

// RegisterCommand 注册命令。
func (m *MockExecutor) RegisterCommand(cmd ICommand) error {
	if cmd == nil {
		return nil
	}
	m.commands.Store(cmd.Info().Name, cmd)
	return nil
}

// UnregisterCommand 注销命令。
func (m *MockExecutor) UnregisterCommand(name string) error {
	m.commands.Delete(name)
	return nil
}

// WriteFile 写入文件（用于测试）。
func (m *MockExecutor) WriteFile(path string, content string) {
	m.files.Store(path, &MockFile{
		Content: content,
		Exists:  true,
		Mode:    0666,
	})
}

// ReadFile 读取文件（用于测试）。
func (m *MockExecutor) ReadFile(path string) (string, bool) {
	if value, ok := m.files.Load(path); ok {
		if file, ok := value.(*MockFile); ok && file.Exists {
			return file.Content, true
		}
	}
	return "", false
}

// DeleteFile 删除文件（用于测试）。
func (m *MockExecutor) DeleteFile(path string) {
	m.files.Store(path, &MockFile{
		Exists: false,
	})
}

// CreateDir 创建目录（用于测试）。
func (m *MockExecutor) CreateDir(path string) {
	m.files.Store(path, &MockFile{
		Exists: true,
		IsDir:  true,
		Mode:   0777,
	})
}

// FileExists 检查文件是否存在（用于测试）。
func (m *MockExecutor) FileExists(path string) bool {
	if value, ok := m.files.Load(path); ok {
		if file, ok := value.(*MockFile); ok {
			return file.Exists
		}
	}
	return false
}

// IsDir 检查路径是否是目录（用于测试）。
func (m *MockExecutor) IsDir(path string) bool {
	if value, ok := m.files.Load(path); ok {
		if file, ok := value.(*MockFile); ok {
			return file.IsDir
		}
	}
	return false
}

// GetFileMode 获取文件模式（用于测试）。
func (m *MockExecutor) GetFileMode(path string) os.FileMode {
	if value, ok := m.files.Load(path); ok {
		if file, ok := value.(*MockFile); ok {
			return file.Mode
		}
	}
	return 0
}

// SetFileMode 设置文件模式（用于测试）。
func (m *MockExecutor) SetFileMode(path string, mode os.FileMode) {
	if value, ok := m.files.Load(path); ok {
		if file, ok := value.(*MockFile); ok {
			file.Mode = mode
			m.files.Store(path, file)
		}
	}
}

// ListDir 列出目录内容（用于测试）。
func (m *MockExecutor) ListDir(path string) []string {
	var files []string
	m.files.Range(func(key, value interface{}) bool {
		if file, ok := value.(*MockFile); ok && file.Exists {
			if strings.HasPrefix(key.(string), path) {
				files = append(files, key.(string))
			}
		}
		return true
	})
	return files
}

// CopyFile 复制文件（用于测试）。
func (m *MockExecutor) CopyFile(src, dst string) error {
	content, exists := m.ReadFile(src)
	if !exists {
		return fmt.Errorf("source file not found: %s", src)
	}
	m.WriteFile(dst, content)
	return nil
}

// NewMockExecutorBuilder creates a new mock executor builder for testing
func NewMockExecutorBuilder(mockExec *MockExecutor) ExecutorBuilder {
	return &mockExecutorBuilder{
		mockExec: mockExec,
	}
}

type mockExecutorBuilder struct {
	mockExec *MockExecutor
}

func (b *mockExecutorBuilder) Build(options *ExecuteOptions) (Executor, error) {
	return b.mockExec, nil
}

func (b *mockExecutorBuilder) WithWorkDir(workDir string) ExecutorBuilder {
	return b
}

func (b *mockExecutorBuilder) WithEnv(env map[string]string) ExecutorBuilder {
	return b
}

func (b *mockExecutorBuilder) WithStdout(stdout io.Writer) ExecutorBuilder {
	return b
}

func (b *mockExecutorBuilder) WithStderr(stderr io.Writer) ExecutorBuilder {
	return b
}

func (b *mockExecutorBuilder) WithStdin(stdin io.Reader) ExecutorBuilder {
	return b
}

func (b *mockExecutorBuilder) WithCommandProvider(provider BuiltinCommandProvider) ExecutorBuilder {
	return b
}
