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
	log.Debug("Creating new pipeline executor")
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
func (e *PipelineExecutor) ExecutePipeline(pipeline *types.PipelineContext) (*types.ExecuteResult, error) {
	if len(pipeline.Commands) == 0 {
		log.Error("Attempting to execute empty pipeline")
		return nil, fmt.Errorf("empty pipeline")
	}

	log.Info("Starting pipeline execution with %d commands", len(pipeline.Commands))

	// 创建执行上下文
	ctx := &types.ExecuteContext{
		Context: pipeline.Context,
		Options: pipeline.Options,
		IsPiped: true,
		PipeContext: &types.PipelineContext{
			Context:  pipeline.Context,
			Commands: pipeline.Commands,
			Options:  pipeline.Options,
		},
	}

	// 执行完整的管道命令
	startTime := types.GetTimeNow()
	result, err := e.executor.Execute(ctx)
	endTime := types.GetTimeNow()

	if err != nil {
		log.Error("Pipeline execution failed: %v", err)
		return result, fmt.Errorf("pipeline execution failed: %w", err)
	}

	log.Info("Pipeline execution completed successfully")
	return &types.ExecuteResult{
		CommandName: "pipeline",
		ExitCode:    result.ExitCode,
		StartTime:   startTime,
		EndTime:     endTime,
		Output:      result.Output,
	}, nil
}

// Execute 实现 Executor 接口
func (e *PipelineExecutor) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx.PipeContext == nil {
		log.Debug("Executing single command through pipeline executor")
		return e.executor.Execute(ctx)
	}
	log.Debug("Executing pipeline command")
	return e.ExecutePipeline(ctx.PipeContext)
}

// ListCommands 实现 Executor 接口
func (e *PipelineExecutor) ListCommands() []types.CommandInfo {
	log.Debug("Listing commands from underlying executor")
	return e.executor.ListCommands()
}
