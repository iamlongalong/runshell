// Package commands 实现了 RunShell 的内置命令。
// 本文件实现了脚本管理和执行的命令。
package commands

import (
	"fmt"

	"github.com/iamlongalong/runshell/pkg/script"
	"github.com/iamlongalong/runshell/pkg/types"
)

// ScriptCommand 脚本命令
type ScriptCommand struct {
	scriptManager *script.ScriptManager
}

// NewScriptCommand 创建新的脚本命令
func NewScriptCommand(executor types.Executor) *ScriptCommand {
	return &ScriptCommand{
		scriptManager: script.NewScriptManager(executor),
	}
}

// Execute 执行脚本命令
func (c *ScriptCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Args) == 0 {
		return nil, fmt.Errorf("script command requires at least 2 arguments: <action> <name>")
	}

	action := ctx.Args[0]

	switch action {
	case "list":
		return c.listScripts(ctx)
	case "run", "save", "get", "delete":
		if len(ctx.Args) < 2 {
			return nil, fmt.Errorf("script command requires at least 2 arguments: <action> <name>")
		}
		name := ctx.Args[1]
		switch action {
		case "run":
			return c.runScript(ctx, name)
		case "save":
			return c.saveScript(ctx, name)
		case "get":
			return c.getScript(ctx, name)
		case "delete":
			return c.deleteScript(ctx, name)
		}
	}

	return nil, fmt.Errorf("unknown action: %s", action)
}

// runScript 运行脚本
func (c *ScriptCommand) runScript(ctx *types.ExecuteContext, name string) (*types.ExecuteResult, error) {
	script, err := c.scriptManager.GetScript(name)
	if err != nil {
		return nil, err
	}

	return c.scriptManager.Execute(ctx.Context, script)
}

// saveScript 保存脚本
func (c *ScriptCommand) saveScript(ctx *types.ExecuteContext, name string) (*types.ExecuteResult, error) {
	if len(ctx.Args) < 3 {
		return nil, fmt.Errorf("save action requires command to execute")
	}

	script := &script.Script{
		Name:    name,
		Command: ctx.Args[2],
		Args:    ctx.Args[3:],
		WorkDir: ctx.Options.WorkDir,
		Env:     ctx.Options.Env,
	}

	if err := c.scriptManager.SaveScript(script); err != nil {
		return nil, err
	}

	return &types.ExecuteResult{
		CommandName: "script",
		ExitCode:    0,
		StartTime:   types.GetTimeNow(),
		EndTime:     types.GetTimeNow(),
	}, nil
}

// listScripts 列出所有脚本
func (c *ScriptCommand) listScripts(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	scripts := c.scriptManager.ListScripts()
	for _, s := range scripts {
		fmt.Fprintf(ctx.Options.Stdout, "%s: %s %v\n", s.Name, s.Command, s.Args)
	}

	return &types.ExecuteResult{
		CommandName: "script",
		ExitCode:    0,
		StartTime:   types.GetTimeNow(),
		EndTime:     types.GetTimeNow(),
	}, nil
}

// getScript 获取脚本内容
func (c *ScriptCommand) getScript(ctx *types.ExecuteContext, name string) (*types.ExecuteResult, error) {
	script, err := c.scriptManager.GetScript(name)
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(ctx.Options.Stdout, "Name: %s\n", script.Name)
	fmt.Fprintf(ctx.Options.Stdout, "Command: %s\n", script.Command)
	fmt.Fprintf(ctx.Options.Stdout, "Args: %v\n", script.Args)
	fmt.Fprintf(ctx.Options.Stdout, "WorkDir: %s\n", script.WorkDir)
	fmt.Fprintf(ctx.Options.Stdout, "Env:\n")
	for k, v := range script.Env {
		fmt.Fprintf(ctx.Options.Stdout, "  %s=%s\n", k, v)
	}

	return &types.ExecuteResult{
		CommandName: "script",
		ExitCode:    0,
		StartTime:   types.GetTimeNow(),
		EndTime:     types.GetTimeNow(),
	}, nil
}

// deleteScript 删除脚本
func (c *ScriptCommand) deleteScript(ctx *types.ExecuteContext, name string) (*types.ExecuteResult, error) {
	if err := c.scriptManager.DeleteScript(name); err != nil {
		return nil, err
	}

	return &types.ExecuteResult{
		CommandName: "script",
		ExitCode:    0,
		StartTime:   types.GetTimeNow(),
		EndTime:     types.GetTimeNow(),
	}, nil
}
