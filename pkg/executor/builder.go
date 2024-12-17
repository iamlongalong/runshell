package executor

import (
	"github.com/iamlongalong/runshell/pkg/commands"
	"github.com/iamlongalong/runshell/pkg/types"
)

// LocalExecutorBuilder 是本地执行器的构建器。
type LocalExecutorBuilder struct {
	config   types.LocalConfig
	options  *types.ExecuteOptions
	provider types.BuiltinCommandProvider
}

// NewLocalExecutorBuilder 创建一个新的本地执行器构建器。
func NewLocalExecutorBuilder(config types.LocalConfig) *LocalExecutorBuilder {
	return &LocalExecutorBuilder{
		config: config,
	}
}

// WithOptions 设置执行选项。
func (b *LocalExecutorBuilder) WithOptions(options *types.ExecuteOptions) *LocalExecutorBuilder {
	b.options = options
	return b
}

// WithProvider 设置内置命令提供者。
func (b *LocalExecutorBuilder) WithProvider(provider types.BuiltinCommandProvider) *LocalExecutorBuilder {
	b.provider = provider
	return b
}

// Build 构建并返回一个新的本地执行器实例。
func (b *LocalExecutorBuilder) Build() (types.Executor, error) {
	if b.provider == nil && b.config.UseBuiltinCommands {
		b.provider = commands.NewDefaultCommandProvider()
	}
	return NewLocalExecutor(b.config, b.options, b.provider), nil
}