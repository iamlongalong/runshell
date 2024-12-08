// Package script 提供脚本管理和执行功能。
// 本文件实现了脚本管理器的核心功能。
package script

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/iamlongalong/runshell/pkg/types"
)

// Script 表示一个脚本
type Script struct {
	Name    string            // 脚本名称
	Command string            // 要执行的命令
	Args    []string          // 命令参数
	WorkDir string            // 工作目录
	Env     map[string]string // 环境变量
}

// ScriptManager 脚本管理器
type ScriptManager struct {
	executor types.Executor
	scripts  sync.Map
}

// NewScriptManager 创建新的脚本管理器
func NewScriptManager(executor types.Executor) *ScriptManager {
	return &ScriptManager{
		executor: executor,
	}
}

// Execute 执行脚本
func (sm *ScriptManager) Execute(ctx context.Context, script *Script) (*types.ExecuteResult, error) {
	// 验证脚本
	if err := sm.validateScript(script); err != nil {
		return nil, err
	}

	// 准备执行环境
	opts := &types.ExecuteOptions{
		WorkDir: script.WorkDir,
		Env:     script.Env,
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}

	// 执行命令
	execCtx := &types.ExecuteContext{
		Context: ctx,
		Args:    append([]string{script.Command}, script.Args...),
		Options: opts,
	}

	return sm.executor.Execute(execCtx)
}

// ListScripts 列出所有可用脚本
func (sm *ScriptManager) ListScripts() []Script {
	scripts := make([]Script, 0)
	sm.scripts.Range(func(key, value interface{}) bool {
		scripts = append(scripts, value.(Script))
		return true
	})
	return scripts
}

// GetScript 获取脚本
func (sm *ScriptManager) GetScript(name string) (*Script, error) {
	if script, ok := sm.scripts.Load(name); ok {
		s := script.(Script)
		return &s, nil
	}
	return nil, fmt.Errorf("script not found: %s", name)
}

// SaveScript 保存脚本
func (sm *ScriptManager) SaveScript(script *Script) error {
	if err := sm.validateScript(script); err != nil {
		return err
	}
	sm.scripts.Store(script.Name, *script)
	return nil
}

// DeleteScript 删除脚本
func (sm *ScriptManager) DeleteScript(name string) error {
	if _, ok := sm.scripts.Load(name); !ok {
		return fmt.Errorf("script not found: %s", name)
	}
	sm.scripts.Delete(name)
	return nil
}

// validateScript 验证脚本
func (sm *ScriptManager) validateScript(script *Script) error {
	if script == nil {
		return fmt.Errorf("script is nil")
	}
	if script.Name == "" {
		return fmt.Errorf("script name is empty")
	}
	if script.Command == "" {
		return fmt.Errorf("script command is empty")
	}
	return nil
}
