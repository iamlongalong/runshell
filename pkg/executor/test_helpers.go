// Package executor 实现了命令执行器的核心功能。
// 本文件提供了用于测试的辅助类型和函数。
package executor

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/iamlongalong/runshell/pkg/types"
)

// MockExecutor 模拟执行器
type MockExecutor struct {
	ExecuteFunc func(ctx *types.ExecuteContext) (*types.ExecuteResult, error)
	CloseFunc   func() error
	Options     *types.ExecuteOptions
	FileSystem  map[string]string // 模拟文件系统
}

func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		FileSystem: make(map[string]string),
	}
}

func (m *MockExecutor) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx)
	}

	// 默认行为：处理常见命令
	cmd := ctx.Args[0]
	args := ctx.Args[1:]

	// 处理无效命令
	if strings.Contains(strings.Join(ctx.Args, " "), "invalid") {
		return &types.ExecuteResult{
			CommandName: cmd,
			ExitCode:    1,
			Error:       fmt.Errorf("invalid command"),
		}, fmt.Errorf("invalid command")
	}

	switch cmd {
	case "ls":
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		// 列出指定路径下的文件
		var files []string
		prefix := path
		if prefix == "." {
			prefix = ""
		}

		// 如果路径不是 "." 且不存在任何文件以该路径为前缀，返回错误
		if prefix != "" {
			hasFiles := false
			for filePath := range m.FileSystem {
				if strings.HasPrefix(filePath, prefix) {
					hasFiles = true
					break
				}
			}
			if !hasFiles {
				return &types.ExecuteResult{
					CommandName: cmd,
					ExitCode:    1,
					Error:       fmt.Errorf("no such file or directory"),
				}, fmt.Errorf("no such file or directory")
			}
		}

		// 收集文件列表
		for filePath := range m.FileSystem {
			if prefix == "" || strings.HasPrefix(filePath, prefix) {
				base := filepath.Base(filePath)
				if !contains(files, base) {
					files = append(files, base)
				}
			}
		}

		output := strings.Join(files, "\n")
		if len(files) > 0 {
			output += "\n"
		}

		if ctx.Options.Stdout != nil {
			ctx.Options.Stdout.Write([]byte(output))
		}
		return &types.ExecuteResult{
			CommandName: cmd,
			ExitCode:    0,
			Output:      output,
		}, nil

	case "cat":
		if len(args) == 0 {
			return &types.ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("missing file operand"),
			}, fmt.Errorf("missing file operand")
		}
		content, exists := m.FileSystem[args[0]]
		if !exists {
			return &types.ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("no such file"),
			}, fmt.Errorf("no such file")
		}
		if ctx.Options.Stdout != nil {
			ctx.Options.Stdout.Write([]byte(content))
		}
		return &types.ExecuteResult{
			CommandName: cmd,
			ExitCode:    0,
			Output:      content,
		}, nil

	case "mkdir":
		if len(args) == 0 {
			return &types.ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("missing operand"),
			}, fmt.Errorf("missing operand")
		}
		m.FileSystem[args[0]] = ""
		return &types.ExecuteResult{
			CommandName: cmd,
			ExitCode:    0,
		}, nil

	case "rm":
		if len(args) == 0 {
			return &types.ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("missing operand"),
			}, fmt.Errorf("missing operand")
		}
		if !m.fileExists(args[0]) {
			return &types.ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("no such file"),
			}, fmt.Errorf("no such file")
		}
		delete(m.FileSystem, args[0])
		return &types.ExecuteResult{
			CommandName: cmd,
			ExitCode:    0,
		}, nil

	case "cp":
		if len(args) < 2 {
			return &types.ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("missing file operand"),
			}, fmt.Errorf("missing file operand")
		}
		content, exists := m.FileSystem[args[0]]
		if !exists {
			return &types.ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("no such file"),
			}, fmt.Errorf("no such file")
		}
		m.FileSystem[args[1]] = content
		return &types.ExecuteResult{
			CommandName: cmd,
			ExitCode:    0,
		}, nil

	case "pwd":
		output := ctx.Options.WorkDir + "\n"
		if ctx.Options.Stdout != nil {
			ctx.Options.Stdout.Write([]byte(output))
		}
		return &types.ExecuteResult{
			CommandName: cmd,
			ExitCode:    0,
			Output:      output,
		}, nil

	case "ps":
		output := "  PID  CPU%  MEM%  COMMAND\n  1    0.0   0.1   init\n"
		if ctx.Options.Stdout != nil {
			ctx.Options.Stdout.Write([]byte(output))
		}
		return &types.ExecuteResult{
			CommandName: cmd,
			ExitCode:    0,
			Output:      output,
		}, nil

	case "top":
		output := "System Overview\nCPU: 0.1%\nMEM: 50%\n"
		if ctx.Options.Stdout != nil {
			ctx.Options.Stdout.Write([]byte(output))
		}
		return &types.ExecuteResult{
			CommandName: cmd,
			ExitCode:    0,
			Output:      output,
		}, nil

	case "df":
		output := "Filesystem  Size  Used  Avail  Use%  Mounted on\n/dev/sda1   100G  50G   50G    50%   /\n"
		if ctx.Options.Stdout != nil {
			ctx.Options.Stdout.Write([]byte(output))
		}
		return &types.ExecuteResult{
			CommandName: cmd,
			ExitCode:    0,
			Output:      output,
		}, nil

	case "git":
		if len(args) > 0 && args[0] == "status" && ctx.Options.WorkDir == "/tmp" {
			return &types.ExecuteResult{
				CommandName: cmd,
				ExitCode:    128,
				Error:       fmt.Errorf("not a git repository"),
			}, fmt.Errorf("not a git repository")
		}

	case "go":
		if len(args) > 0 && args[0] == "mod" && ctx.Options.WorkDir == "/tmp" {
			return &types.ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("go.mod file not found"),
			}, fmt.Errorf("go.mod file not found")
		}

	case "wget":
		if len(args) == 0 {
			return &types.ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("missing URL"),
			}, fmt.Errorf("missing URL")
		}
		if !strings.HasPrefix(args[0], "http") {
			return &types.ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("invalid URL"),
			}, fmt.Errorf("invalid URL")
		}
		return &types.ExecuteResult{
			CommandName: cmd,
			ExitCode:    0,
		}, nil

	case "tar":
		if len(args) == 0 {
			return &types.ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("missing operand"),
			}, fmt.Errorf("missing operand")
		}
		return &types.ExecuteResult{
			CommandName: cmd,
			ExitCode:    0,
		}, nil

	case "zip":
		if len(args) < 2 {
			return &types.ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("missing file operand"),
			}, fmt.Errorf("missing file operand")
		}
		// 检查所有源文件是否存在
		for _, file := range args[1:] {
			if !m.fileExists(file) {
				return &types.ExecuteResult{
					CommandName: cmd,
					ExitCode:    1,
					Error:       fmt.Errorf("no such file: %s", file),
				}, fmt.Errorf("no such file: %s", file)
			}
		}
		return &types.ExecuteResult{
			CommandName: cmd,
			ExitCode:    0,
		}, nil

	case "python":
		if len(args) == 0 {
			return &types.ExecuteResult{
				CommandName: cmd,
				ExitCode:    1,
				Error:       fmt.Errorf("missing operand"),
			}, fmt.Errorf("missing operand")
		}
		if args[0] == "--version" {
			output := "Python 3.9.0\n"
			if ctx.Options.Stdout != nil {
				ctx.Options.Stdout.Write([]byte(output))
			}
			return &types.ExecuteResult{
				CommandName: cmd,
				ExitCode:    0,
				Output:      output,
			}, nil
		}
		return &types.ExecuteResult{
			CommandName: cmd,
			ExitCode:    0,
		}, nil
	}

	// 默认返回成功
	return &types.ExecuteResult{
		CommandName: cmd,
		ExitCode:    0,
		Output:      "mock output",
	}, nil
}

func (m *MockExecutor) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockExecutor) ListCommands() []types.CommandInfo {
	return []types.CommandInfo{
		{
			Name:        "test",
			Description: "Test command",
			Usage:       "test [args...]",
			Category:    "test",
		},
	}
}

func (m *MockExecutor) SetOptions(options *types.ExecuteOptions) {
	m.Options = options
}

func (m *MockExecutor) fileExists(path string) bool {
	_, exists := m.FileSystem[path]
	return exists
}

func (m *MockExecutor) WriteFile(path, content string) {
	m.FileSystem[path] = content
}

func (m *MockExecutor) ReadFile(path string) (string, bool) {
	content, exists := m.FileSystem[path]
	return content, exists
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
