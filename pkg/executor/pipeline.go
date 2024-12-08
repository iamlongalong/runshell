package executor

import (
	"bytes"
	"fmt"
	"strings"

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

// ParsePipeline 解析管道命令
func (e *PipelineExecutor) ParsePipeline(cmdStr string) (*types.PipelineContext, error) {
	// 按管道符分割命令
	cmds := strings.Split(cmdStr, "|")
	if len(cmds) == 0 {
		return nil, fmt.Errorf("empty pipeline")
	}

	// 创建管道上下文
	pipeline := &types.PipelineContext{
		Commands: make([]*types.PipeCommand, 0, len(cmds)),
		Options:  &types.ExecuteOptions{},
	}

	// 解析每个命令
	for _, cmd := range cmds {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}

		// 分割命令和参数
		parts := strings.Fields(cmd)
		if len(parts) == 0 {
			continue
		}

		pipeline.Commands = append(pipeline.Commands, &types.PipeCommand{
			Command: parts[0],
			Args:    parts[1:],
		})
	}

	if len(pipeline.Commands) == 0 {
		return nil, fmt.Errorf("no valid commands in pipeline")
	}

	return pipeline, nil
}

// ExecutePipeline 执行管道命令
func (e *PipelineExecutor) ExecutePipeline(pipeline *types.PipelineContext) (*types.ExecuteResult, error) {
	if len(pipeline.Commands) == 0 {
		return nil, fmt.Errorf("empty pipeline")
	}

	var lastOutput *bytes.Buffer
	var firstResult *types.ExecuteResult
	var lastResult *types.ExecuteResult

	// 执行每个命令
	for i, cmd := range pipeline.Commands {
		// 创建命令上下文
		ctx := &types.ExecuteContext{
			Context: pipeline.Context,
			Args:    append([]string{cmd.Command}, cmd.Args...),
			Options: pipeline.Options,
			IsPiped: true,
		}

		// 设置管道
		if i > 0 && lastOutput != nil {
			ctx.PipeInput = lastOutput
		}

		if i < len(pipeline.Commands)-1 {
			output := &bytes.Buffer{}
			ctx.PipeOutput = output
			lastOutput = output
		} else {
			ctx.PipeOutput = pipeline.Options.Stdout
		}

		// 执行命令
		result, err := e.executor.Execute(ctx)
		if err != nil {
			if i == 0 {
				return result, fmt.Errorf("pipeline command %d failed: %w", i+1, err)
			}
			return result, fmt.Errorf("pipeline command %d failed: %w", i+1, err)
		}

		// 保存第一个和最后一个结果
		if i == 0 {
			firstResult = result
		}
		lastResult = result
	}

	// 返回组合结果
	return &types.ExecuteResult{
		CommandName: "pipeline",
		ExitCode:    lastResult.ExitCode,
		StartTime:   firstResult.StartTime,
		EndTime:     lastResult.EndTime,
	}, nil
}

// Execute 实现 Executor 接口
func (e *PipelineExecutor) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.PipeContext == nil {
		return e.executor.Execute(ctx)
	}
	return e.ExecutePipeline(ctx.PipeContext)
}

// ListCommands 实现 Executor 接口
func (e *PipelineExecutor) ListCommands() []types.CommandInfo {
	return e.executor.ListCommands()
}
