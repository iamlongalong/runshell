package script

import (
	"context"
	"testing"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestScriptManager(t *testing.T) {
	// 创建模拟执行器
	mockExec := &executor.MockExecutor{}

	// 创建脚本管理器
	sm := NewScriptManager(mockExec)

	// 测试保存脚本
	script := &Script{
		Name:    "test",
		Command: "echo",
		Args:    []string{"hello"},
		WorkDir: "/tmp",
		Env:     map[string]string{"FOO": "bar"},
	}

	err := sm.SaveScript(script)
	assert.NoError(t, err)

	// 测试获取脚本
	saved, err := sm.GetScript("test")
	assert.NoError(t, err)
	assert.Equal(t, script.Name, saved.Name)
	assert.Equal(t, script.Command, saved.Command)
	assert.Equal(t, script.Args, saved.Args)
	assert.Equal(t, script.WorkDir, saved.WorkDir)
	assert.Equal(t, script.Env, saved.Env)

	// 测试列出脚本
	scripts := sm.ListScripts()
	assert.Len(t, scripts, 1)
	assert.Equal(t, script.Name, scripts[0].Name)

	// 测试执行脚本
	executed := false
	mockExec.ExecuteFunc = func(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
		executed = true
		assert.Equal(t, script.Command, ctx.Args[0])
		assert.Equal(t, script.Args, ctx.Args[1:])
		assert.Equal(t, script.WorkDir, ctx.Options.WorkDir)
		assert.Equal(t, script.Env, ctx.Options.Env)
		return &types.ExecuteResult{}, nil
	}

	_, err = sm.Execute(context.Background(), script)
	assert.NoError(t, err)
	assert.True(t, executed)

	// 测试删除脚本
	err = sm.DeleteScript("test")
	assert.NoError(t, err)

	// 验证脚本已被删除
	scripts = sm.ListScripts()
	assert.Len(t, scripts, 0)
}
