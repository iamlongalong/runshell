// Package commands 实现了 RunShell 的内置命令。
// 本文件实现了命令注册机制，用于将所有内置命令注册到执行器中。
package commands

import (
	"fmt"

	"github.com/iamlongalong/runshell/pkg/types"
)

// RegisterCommands 注册所有内置命令
func RegisterCommands(executor types.Executor) error {
	// 创建命令处理器
	commands := []types.Command{
		{
			Name:        "ls",
			Description: "List directory contents",
			Usage:       "ls [options] [directory]",
			Category:    "file",
			Handler:     &LSCommand{},
		},
		{
			Name:        "cat",
			Description: "Concatenate and print files",
			Usage:       "cat [options] [file...]",
			Category:    "file",
			Handler:     &CatCommand{},
		},
		{
			Name:        "mkdir",
			Description: "Make directories",
			Usage:       "mkdir [options] directory...",
			Category:    "file",
			Handler:     &MkdirCommand{},
		},
		{
			Name:        "rm",
			Description: "Remove files or directories",
			Usage:       "rm [options] file...",
			Category:    "file",
			Handler:     &RmCommand{},
		},
		{
			Name:        "cp",
			Description: "Copy files and directories",
			Usage:       "cp [options] source... destination",
			Category:    "file",
			Handler:     &CpCommand{},
		},
		{
			Name:        "pwd",
			Description: "Print working directory",
			Usage:       "pwd",
			Category:    "file",
			Handler:     &PWDCommand{},
		},
		{
			Name:        "script",
			Description: "Manage and execute scripts",
			Usage:       "script [run|save|list|get|delete] [args...]",
			Category:    "script",
			Handler:     NewScriptCommand(executor),
		},
		{
			Name:        "ps",
			Description: "Report process status",
			Usage:       "ps [options]",
			Category:    "process",
			Handler:     &PSCommand{},
		},
		{
			Name:        "top",
			Description: "Display system processes",
			Usage:       "top [options]",
			Category:    "process",
			Handler:     &TopCommand{},
		},
		{
			Name:        "df",
			Description: "Report file system disk space usage",
			Usage:       "df [options] [file...]",
			Category:    "system",
			Handler:     &DFCommand{},
		},
		{
			Name:        "uname",
			Description: "Print system information",
			Usage:       "uname [options]",
			Category:    "system",
			Handler:     &UNameCommand{},
		},
		{
			Name:        "env",
			Description: "Set or print environment variables",
			Usage:       "env [name[=value] ...]",
			Category:    "system",
			Handler:     &EnvCommand{},
		},
		{
			Name:        "kill",
			Description: "Terminate processes",
			Usage:       "kill [options] pid...",
			Category:    "process",
			Handler:     &KillCommand{},
		},
		{
			Name:        "wget",
			Description: "Download files from the web",
			Usage:       "wget [options] url",
			Category:    "network",
			Handler:     &WgetCommand{},
		},
		{
			Name:        "tar",
			Description: "Create or extract archives",
			Usage:       "tar [options] [archive] [file...]",
			Category:    "file",
			Handler:     &TarCommand{},
		},
		{
			Name:        "zip",
			Description: "Package and compress files",
			Usage:       "zip [options] zipfile file...",
			Category:    "file",
			Handler:     &ZipCommand{},
		},
		{
			Name:        "python",
			Description: "Run Python interpreter",
			Usage:       "python [options] [script] [args]",
			Category:    "language",
			Handler:     &PythonCommand{},
		},
		{
			Name:        "pip",
			Description: "Python package installer",
			Usage:       "pip [options] <command> [args]",
			Category:    "package",
			Handler:     &PipCommand{},
		},
		{
			Name:        "docker",
			Description: "Docker container operations",
			Usage:       "docker [options] <command> [args]",
			Category:    "container",
			Handler:     &DockerCommand{},
		},
		{
			Name:        "node",
			Description: "Run Node.js interpreter",
			Usage:       "node [options] [script] [args]",
			Category:    "language",
			Handler:     &NodeCommand{},
		},
		{
			Name:        "npm",
			Description: "Node.js package manager",
			Usage:       "npm [options] <command> [args]",
			Category:    "package",
			Handler:     &NPMCommand{},
		},
		{
			Name:        "git",
			Description: "Git version control",
			Usage:       "git [options] <command> [args]",
			Category:    "vcs",
			Handler:     &GitCommand{},
		},
		{
			Name:        "go",
			Description: "Go language tools",
			Usage:       "go <command> [args]",
			Category:    "language",
			Handler:     &GoCommand{},
		},
	}

	// 注册所有命令
	for _, cmd := range commands {
		if err := RegisterCommand(executor, &cmd); err != nil {
			return fmt.Errorf("failed to register command %s: %w", cmd.Name, err)
		}
	}

	return nil
}

// RegisterCommand 注册单个命令
func RegisterCommand(executor types.Executor, cmd *types.Command) error {
	if cmd == nil {
		return fmt.Errorf("command is nil")
	}
	if cmd.Name == "" {
		return fmt.Errorf("command name is empty")
	}
	if cmd.Handler == nil {
		return fmt.Errorf("command handler is nil")
	}

	// 注册命令
	if e, ok := executor.(interface {
		RegisterCommand(*types.Command) error
	}); ok {
		return e.RegisterCommand(cmd)
	}

	return fmt.Errorf("executor does not support command registration")
}
