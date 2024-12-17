package commands

import (
	"github.com/iamlongalong/runshell/pkg/types"
)

// DefaultCommandProvider 提供默认的内置命令实现。
type DefaultCommandProvider struct{}

// NewDefaultCommandProvider 创建一个新的默认命令提供者。
func NewDefaultCommandProvider() *DefaultCommandProvider {
	return &DefaultCommandProvider{}
}

// GetCommands 实现 BuiltinCommandProvider 接口。
// 返回所有内置命令。
func (p *DefaultCommandProvider) GetCommands() []types.ICommand {
	return GetBuiltinCommands()
}
