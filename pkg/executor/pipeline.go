package executor

import (
	"fmt"
	"strings"

	"github.com/iamlongalong/runshell/pkg/log"
	"github.com/iamlongalong/runshell/pkg/types"
)

// PipelineExecutor 管道执行器
type PipelineExecutor struct {
	executor types.Executor
}

// NewPipelineExecutor 创建新的管道执行器
func NewPipelineExecutor(executor types.Executor) *PipelineExecutor {
	return &PipelineExecutor{
		executor: executor,
	}
}

const (
	PipelineExecutorName = "pipeline"
)

// Name 返回执行器名称
func (e *PipelineExecutor) Name() string {
	return PipelineExecutorName
}

// ParsePipeline 解析管道命令
func (e *PipelineExecutor) ParsePipeline(cmdStr string) (*types.PipelineContext, error) {
	log.Debug("Parsing pipeline command: %s", cmdStr)

	// 按管道符分割命令
	cmds := strings.Split(cmdStr, "|")
	if len(cmds) == 0 {
		log.Error("Empty pipeline command")
		return nil, fmt.Errorf("empty pipeline")
	}

	// 创建管道上下文
	pipeline := &types.PipelineContext{
		Commands: make([]*types.Command, 0, len(cmds)),
		Options:  &types.ExecuteOptions{},
	}

	// 解析每个命令
	for i, cmd := range cmds {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			log.Debug("Skipping empty command at position %d", i+1)
			continue
		}

		// 分割命令和参数
		parts := strings.Fields(cmd)
		if len(parts) == 0 {
			log.Debug("Skipping invalid command at position %d", i+1)
			continue
		}

		log.Debug("Adding command to pipeline: %v", parts)
		pipeline.Commands = append(pipeline.Commands, &types.Command{
			Command: parts[0],
			Args:    parts[1:],
		})
	}

	if len(pipeline.Commands) == 0 {
		log.Error("No valid commands found in pipeline")
		return nil, fmt.Errorf("no valid commands in pipeline")
	}

	log.Debug("Successfully parsed pipeline with %d commands", len(pipeline.Commands))
	return pipeline, nil
}

// ExecutePipeline 执行管道命令
func (p *PipelineExecutor) executePipeline(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx == nil || len(ctx.PipeContext.Commands) == 0 {
		log.Error("No valid commands found in pipeline")
		return nil, fmt.Errorf("no commands in pipeline")
	}

	if ctx.Options == nil {
		ctx.Options = &types.ExecuteOptions{}
	}

	// 执行每个命令
	for _, cmd := range ctx.PipeContext.Commands {
		if cmd == nil || cmd.Command == "" {
			log.Error("Command not found: %v", cmd)
			return nil, fmt.Errorf("command not found: %v", cmd)
		}
	}

	// Create a new context with the same options but without output duplication
	execCtx := &types.ExecuteContext{
		Context:     ctx.Context,
		PipeContext: ctx.PipeContext,
		IsPiped:     true,
		Options:     ctx.Options,
	}

	result, err := p.executor.Execute(execCtx)
	if err != nil {
		return result, err
	}

	return result, nil
}

// Execute 实现 Executor 接口
func (e *PipelineExecutor) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx == nil {
		log.Error("Execute context is nil")
		return nil, fmt.Errorf("execute context is nil")
	}

	if ctx.Command.Command == "" {
		log.Error("No command specified")
		return nil, fmt.Errorf("no command specified")
	}

	// 如果是管道命令
	if strings.Contains(ctx.Command.Command, "|") {
		// 解析管道命令
		pipeline, err := e.ParsePipeline(ctx.Command.Command)
		if err != nil {
			log.Error("Failed to parse pipeline: %v", err)
			return nil, fmt.Errorf("failed to parse pipeline: %w", err)
		}

		// 设置管道上下文
		pipeline.Context = ctx.Context
		pipeline.Options = ctx.Options
		ctx.PipeContext = pipeline
		ctx.IsPiped = true

		// 执行管道命令
		return e.executePipeline(ctx)
	}

	// 如果不是管道命令，直接执行
	return e.executor.Execute(ctx)
}

// ListCommands 实现 Executor 接口
func (e *PipelineExecutor) ListCommands() []types.CommandInfo {
	log.Debug("Listing commands from underlying executor")
	return e.executor.ListCommands()
}
