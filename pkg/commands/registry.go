// Package commands 实现了 RunShell 的内置命令。
// 本文件实现了命令注册机制，用于将所有内置命令注册到执行器中。
package commands

import "github.com/iamlongalong/runshell/pkg/types"

// RegisterBuiltinCommands 注册所有内置命令到执行器。
// 内置命令按以下类别组织：
// 1. 文件系统命令：用于文件和目录操作
// 2. 系统命令：用于系统状态和进程管理
// 3. 文本处理命令：用于文本文件处理
// 4. 网络命令：用于网络操作
//
// 每个命令都包含：
//   - Name：命令名称
//   - Description：命令描述
//   - Usage：使用说明
//   - Category：所属类别
//   - Handler：命令处理器
//
// 参数：
//   - executor：命令执行器实例
//
// 返回值：
//   - error：注册过程中的错误，如果全部成功则返回 nil
func RegisterBuiltinCommands(executor types.Executor) error {
	// 定义所有内置命令
	commands := []*types.Command{
		// 文件系统命令组
		// 用于文件和目录的基本操作
		{
			Name:        "ls",
			Description: "List directory contents",
			Usage:       "ls [path]",
			Category:    "filesystem",
			Handler:     &LSCommand{},
		},
		{
			Name:        "cat",
			Description: "Concatenate and print files",
			Usage:       "cat [file...]",
			Category:    "filesystem",
			Handler:     &CatCommand{},
		},
		{
			Name:        "mkdir",
			Description: "Make directories",
			Usage:       "mkdir [directory...]",
			Category:    "filesystem",
			Handler:     &MkdirCommand{},
		},
		{
			Name:        "rm",
			Description: "Remove files or directories",
			Usage:       "rm [-r] [-f] [file...]",
			Category:    "filesystem",
			Handler:     &RmCommand{},
		},
		{
			Name:        "cp",
			Description: "Copy files",
			Usage:       "cp source dest",
			Category:    "filesystem",
			Handler:     &CpCommand{},
		},
		{
			Name:        "pwd",
			Description: "Print working directory",
			Usage:       "pwd",
			Category:    "filesystem",
			Handler:     &PWDCommand{},
		},
		{
			Name:        "mv",
			Description: "Move (rename) files",
			Usage:       "mv source dest",
			Category:    "filesystem",
			Handler:     &MvCommand{},
		},
		{
			Name:        "touch",
			Description: "Create empty files or update timestamps",
			Usage:       "touch [file...]",
			Category:    "filesystem",
			Handler:     &TouchCommand{},
		},
		{
			Name:        "write",
			Description: "Write content to file",
			Usage:       "write file content...",
			Category:    "filesystem",
			Handler:     &WriteCommand{},
		},
		{
			Name:        "find",
			Description: "Search for files in a directory hierarchy",
			Usage:       "find path [pattern]",
			Category:    "filesystem",
			Handler:     &FindCommand{},
		},

		// 系统命令组
		// 用于系统监控和进程管理
		{
			Name:        "ps",
			Description: "Report process status",
			Usage:       "ps",
			Category:    "system",
			Handler:     &PSCommand{},
		},
		{
			Name:        "top",
			Description: "Display system tasks",
			Usage:       "top",
			Category:    "system",
			Handler:     &TopCommand{},
		},
		{
			Name:        "df",
			Description: "Report file system disk space usage",
			Usage:       "df",
			Category:    "system",
			Handler:     &DFCommand{},
		},
		{
			Name:        "uname",
			Description: "Print system information",
			Usage:       "uname [-a]",
			Category:    "system",
			Handler:     &UNameCommand{},
		},
		{
			Name:        "env",
			Description: "Display environment variables",
			Usage:       "env [pattern]",
			Category:    "system",
			Handler:     &EnvCommand{},
		},
		{
			Name:        "kill",
			Description: "Terminate processes",
			Usage:       "kill pid...",
			Category:    "system",
			Handler:     &KillCommand{},
		},
		{
			Name:        "netstat",
			Description: "Show network status",
			Usage:       "netstat",
			Category:    "system",
			Handler:     &NetstatCommand{},
		},
		{
			Name:        "ifconfig",
			Description: "Configure network interface",
			Usage:       "ifconfig",
			Category:    "system",
			Handler:     &IfconfigCommand{},
		},

		// 文本处理命令组
		// 用于文本文件的处理和分析
		{
			Name:        "grep",
			Description: "Search for patterns in files",
			Usage:       "grep pattern [file...]",
			Category:    "text",
			Handler:     &GrepCommand{},
		},
		{
			Name:        "tail",
			Description: "Output the last part of files",
			Usage:       "tail [-n lines] file",
			Category:    "text",
			Handler:     &TailCommand{},
		},
		{
			Name:        "head",
			Description: "Output the first part of files",
			Usage:       "head [-n lines] file",
			Category:    "text",
			Handler:     &HeadCommand{},
		},
		{
			Name:        "sort",
			Description: "Sort lines of text files",
			Usage:       "sort file",
			Category:    "text",
			Handler:     &SortCommand{},
		},
		{
			Name:        "uniq",
			Description: "Report or omit repeated lines",
			Usage:       "uniq file",
			Category:    "text",
			Handler:     &UniqCommand{},
		},
		{
			Name:        "sed",
			Description: "Stream editor for filtering and transforming text",
			Usage:       "sed pattern file",
			Category:    "text",
			Handler:     &SedCommand{},
		},
		{
			Name:        "xargs",
			Description: "Build and execute command lines from standard input",
			Usage:       "xargs command [args...]",
			Category:    "text",
			Handler:     &XargsCommand{},
		},

		// 网络命令组
		// 用于网络操作和数据传输
		{
			Name:        "curl",
			Description: "Transfer data from or to a server",
			Usage:       "curl url",
			Category:    "network",
			Handler:     &CurlCommand{},
		},
	}

	// 注册所有命令
	// 如果任何命令注册失败，立即返回错误
	for _, cmd := range commands {
		if err := executor.RegisterCommand(cmd); err != nil {
			return err
		}
	}

	return nil
}
