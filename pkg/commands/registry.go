// Package commands 实现了 RunShell 的内置命令。
// 本文件实现了命令注册机制，用于将所有内置命令注册到执行器中。
package commands

import (
	"github.com/iamlongalong/runshell/pkg/types"
)

// GetBuiltinCommands 注册所有内置命令
func GetBuiltinCommands() []types.ICommand {
	// 创建命令处理器
	cmds := []types.ICommand{
		&LSCommand{},
		&CatCommand{},
		&MkdirCommand{},
		&RmCommand{},
		&CpCommand{},
		&PWDCommand{},
		&PSCommand{},
		&TopCommand{},
		&DFCommand{},
		&UNameCommand{},
		&EnvCommand{},
		&KillCommand{},
		&WgetCommand{},
		&TarCommand{},
		&ZipCommand{},
		&PythonCommand{},
		&PipCommand{},
		&DockerCommand{},
		&NodeCommand{},
		&NPMCommand{},
		&GitCommand{},
		&GoCommand{},
		&ReadFileCommand{},
	}
	return cmds
}
