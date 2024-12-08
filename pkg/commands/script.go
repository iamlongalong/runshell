// Package commands 实现了 RunShell 的内置命令。
// 本文件实现了脚本管理和执行的命令。
package commands

import (
	"fmt"
	"time"

	"github.com/iamlongalong/runshell/pkg/script"
	"github.com/iamlongalong/runshell/pkg/types"
)

// ScriptCommand 实现了脚本管理和执行命令。
// 支持以下功能：
// - 运行脚本
// - 保存脚本
// - 列出脚本
// - 显示脚本内容
// - 删除脚本
type ScriptCommand struct {
	scriptManager *script.ScriptManager
}

// NewScriptCommand 创建一个新的脚本命令实例。
// 参数：
//   - scriptDir：脚本存储目录
//   - executor：命令执行器实例
//
// 返回值：
//   - *ScriptCommand：脚本命令实例
//   - error：创建过程中的错误
func NewScriptCommand(scriptDir string, executor types.Executor) (*ScriptCommand, error) {
	manager, err := script.NewScriptManager(scriptDir, executor)
	if err != nil {
		return nil, err
	}

	return &ScriptCommand{
		scriptManager: manager,
	}, nil
}

// Execute 执行脚本命令。
// 支持的子命令：
//   - run：运行脚本
//   - save：保存脚本
//   - list：列出脚本
//   - show：显示脚本内容
//   - delete：删除脚本
//
// 参数：
//   - ctx：执行上下文，包含命令参数和选项
//
// 返回值：
//   - *types.ExecuteResult：执行结果
//   - error：执行过程中的错误
func (c *ScriptCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Args) < 1 {
		return &types.ExecuteResult{
			CommandName: "script",
			ExitCode:    1,
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
			Error:       fmt.Errorf("script requires subcommand"),
			Output:      "Usage: script [run|save|list|show|delete] [args...]\n",
		}, nil
	}

	subcommand := ctx.Args[0]
	switch subcommand {
	case "run":
		return c.runScript(ctx)
	case "save":
		return c.saveScript(ctx)
	case "list":
		return c.listScripts(ctx)
	case "show":
		return c.showScript(ctx)
	case "delete":
		return c.deleteScript(ctx)
	default:
		return &types.ExecuteResult{
			CommandName: "script",
			ExitCode:    1,
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
			Error:       fmt.Errorf("unknown subcommand: %s", subcommand),
			Output:      "Usage: script [run|save|list|show|delete] [args...]\n",
		}, nil
	}
}

// runScript 运行指定的脚本。
// 用法：script run <script-name> <workdir> [args...]
//
// 参数：
//   - script-name：要运行的脚本名称
//   - workdir：脚本执行的工作目录
//   - args：传递给脚本的参数
func (c *ScriptCommand) runScript(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Args) < 3 {
		return &types.ExecuteResult{
			CommandName: "script",
			ExitCode:    1,
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
			Error:       fmt.Errorf("script run requires script name and workdir"),
			Output:      "Usage: script run <script-name> <workdir> [args...]\n",
		}, nil
	}

	scriptName := ctx.Args[1]
	workDir := ctx.Args[2]
	var scriptArgs []string
	if len(ctx.Args) > 3 {
		scriptArgs = ctx.Args[3:]
	}

	// 验证工作目录
	if workDir == "" {
		return &types.ExecuteResult{
			CommandName: "script",
			ExitCode:    1,
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
			Error:       fmt.Errorf("workdir must be specified"),
			Output:      "Usage: script run <script-name> <workdir> [args...]\n",
		}, nil
	}

	// 执行脚本
	result, err := c.scriptManager.ExecuteScript(scriptName, workDir, scriptArgs)
	if err != nil {
		return &types.ExecuteResult{
			CommandName: "script",
			ExitCode:    1,
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
			Error:       err,
		}, nil
	}
	return result, nil
}

// saveScript 保存新的脚本。
// 用法：script save <name> <content>
//
// 参数：
//   - name：脚本名称
//   - content：脚本内容
func (c *ScriptCommand) saveScript(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Args) < 3 {
		return &types.ExecuteResult{
			CommandName: "script",
			ExitCode:    1,
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
			Error:       fmt.Errorf("script save requires name and content"),
			Output:      "Usage: script save <n> <content>\n",
		}, nil
	}

	name := ctx.Args[1]
	content := []byte(ctx.Args[2])

	scriptPath, err := c.scriptManager.SaveScript(name, content)
	if err != nil {
		return &types.ExecuteResult{
			CommandName: "script",
			ExitCode:    1,
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
			Error:       err,
		}, nil
	}

	return &types.ExecuteResult{
		CommandName: "script",
		ExitCode:    0,
		StartTime:   ctx.StartTime,
		EndTime:     time.Now(),
		Output:      fmt.Sprintf("Script saved as: %s\n", scriptPath),
	}, nil
}

// listScripts 列出所有可用的脚本。
// 用法：script list
//
// 输出：
//   - 每行一个脚本名称
func (c *ScriptCommand) listScripts(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	scripts, err := c.scriptManager.ListScripts()
	if err != nil {
		return &types.ExecuteResult{
			CommandName: "script",
			ExitCode:    1,
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
			Error:       err,
		}, nil
	}

	var output string
	for _, script := range scripts {
		output += script + "\n"
	}

	return &types.ExecuteResult{
		CommandName: "script",
		ExitCode:    0,
		StartTime:   ctx.StartTime,
		EndTime:     time.Now(),
		Output:      output,
	}, nil
}

// showScript 显示指定脚本的内容。
// 用法：script show <script-name>
//
// 参数：
//   - script-name：要显示的脚本名称
func (c *ScriptCommand) showScript(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Args) < 2 {
		return &types.ExecuteResult{
			CommandName: "script",
			ExitCode:    1,
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
			Error:       fmt.Errorf("script show requires script name"),
			Output:      "Usage: script show <script-name>\n",
		}, nil
	}

	content, err := c.scriptManager.GetScriptContent(ctx.Args[1])
	if err != nil {
		return &types.ExecuteResult{
			CommandName: "script",
			ExitCode:    1,
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
			Error:       err,
		}, nil
	}

	return &types.ExecuteResult{
		CommandName: "script",
		ExitCode:    0,
		StartTime:   ctx.StartTime,
		EndTime:     time.Now(),
		Output:      string(content),
	}, nil
}

// deleteScript 删除指定的脚本。
// 用法：script delete <script-name>
//
// 参数：
//   - script-name：要删除的脚本名称
func (c *ScriptCommand) deleteScript(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if len(ctx.Args) < 2 {
		return &types.ExecuteResult{
			CommandName: "script",
			ExitCode:    1,
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
			Error:       fmt.Errorf("script delete requires script name"),
			Output:      "Usage: script delete <script-name>\n",
		}, nil
	}

	if err := c.scriptManager.DeleteScript(ctx.Args[1]); err != nil {
		return &types.ExecuteResult{
			CommandName: "script",
			ExitCode:    1,
			StartTime:   ctx.StartTime,
			EndTime:     time.Now(),
			Error:       err,
		}, nil
	}

	return &types.ExecuteResult{
		CommandName: "script",
		ExitCode:    0,
		StartTime:   ctx.StartTime,
		EndTime:     time.Now(),
		Output:      fmt.Sprintf("Script %s deleted\n", ctx.Args[1]),
	}, nil
}
