package script

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScript 测试脚本结构体
type TestScript struct {
	Meta    *ScriptMeta
	Content string
}

// setupTestEnv 设置测试环境
func setupTestEnv(t *testing.T) (string, func()) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "script_test_*")
	require.NoError(t, err)

	// 返回清理函数
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// createTestManager 创建测试用的脚本管理器
func createTestManager(t *testing.T, rootDir string) *ScriptManager {
	manager, err := NewScriptManager(&Config{
		RootDir:  rootDir,
		Executor: types.NewMockExecutor(),
	}, nil)
	require.NoError(t, err)
	return manager
}

// mockExecutor 模拟执行器
type mockExecutor struct{}

func (m *mockExecutor) Name() string { return "mock" }
func (m *mockExecutor) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	return &types.ExecuteResult{}, nil
}
func (m *mockExecutor) ListCommands() []types.CommandInfo        { return nil }
func (m *mockExecutor) Close() error                             { return nil }
func (m *mockExecutor) SetOptions(options *types.ExecuteOptions) {}

// TestCreateScript 测试创建脚本
func TestCreateScript(t *testing.T) {
	rootDir, cleanup := setupTestEnv(t)
	defer cleanup()

	manager := createTestManager(t, rootDir)

	tests := []struct {
		name    string
		script  TestScript
		wantErr bool
	}{
		{
			name: "create simple script",
			script: TestScript{
				Meta: &ScriptMeta{
					Name:        "test1",
					Type:        PythonScript,
					Command:     "test1.py",
					Description: "Test script 1",
					Args: []ArgMeta{
						{Name: "input", Flag: "--input", Required: true},
					},
				},
				Content: "print('hello')",
			},
			wantErr: false,
		},
		{
			name: "create script in subdirectory",
			script: TestScript{
				Meta: &ScriptMeta{
					Name:        "group1/test2",
					Type:        PythonScript,
					Command:     "group1/test2.py",
					Description: "Test script 2",
				},
				Content: "print('world')",
			},
			wantErr: false,
		},
		{
			name: "create script with invalid type",
			script: TestScript{
				Meta: &ScriptMeta{
					Name:    "test3",
					Type:    "invalid",
					Command: "test3.py",
				},
				Content: "",
			},
			wantErr: true,
		},
		{
			name: "create script with duplicate name",
			script: TestScript{
				Meta: &ScriptMeta{
					Name:    "test1",
					Type:    PythonScript,
					Command: "test1_dup.py",
				},
				Content: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.CreateScript(tt.script.Meta, tt.script.Content)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// 验证文件是否创建
			scriptPath := filepath.Join(rootDir, tt.script.Meta.Command)
			metaPath := scriptPath + MetaFileSuffix

			// 检查脚本文件
			content, err := os.ReadFile(scriptPath)
			assert.NoError(t, err)
			assert.Equal(t, tt.script.Content, string(content))

			// 检查元数据文件
			metaContent, err := os.ReadFile(metaPath)
			assert.NoError(t, err)
			var meta ScriptMeta
			err = json.Unmarshal(metaContent, &meta)
			assert.NoError(t, err)
			assert.Equal(t, tt.script.Meta.Name, meta.Name)
		})
	}
}

// TestUpdateScript 测试更新脚本
func TestUpdateScript(t *testing.T) {
	rootDir, cleanup := setupTestEnv(t)
	defer cleanup()

	manager := createTestManager(t, rootDir)

	// 创建初始脚本
	initialScript := TestScript{
		Meta: &ScriptMeta{
			Name:        "test1",
			Type:        PythonScript,
			Command:     "test1.py",
			Description: "Initial description",
		},
		Content: "print('initial')",
	}
	err := manager.CreateScript(initialScript.Meta, initialScript.Content)
	require.NoError(t, err)

	tests := []struct {
		name       string
		scriptName string
		newMeta    *ScriptMeta
		newContent string
		wantErr    bool
	}{
		{
			name:       "update description only",
			scriptName: "test1",
			newMeta: &ScriptMeta{
				Name:        "test1",
				Type:        PythonScript,
				Command:     "test1.py",
				Description: "Updated description",
			},
			wantErr: false,
		},
		{
			name:       "update content only",
			scriptName: "test1",
			newContent: "print('updated')",
			wantErr:    false,
		},
		{
			name:       "update name and move to subdirectory",
			scriptName: "test1",
			newMeta: &ScriptMeta{
				Name:        "group1/test1",
				Type:        PythonScript,
				Command:     "group1/test1.py",
				Description: "Moved script",
			},
			wantErr: false,
		},
		{
			name:       "update non-existent script",
			scriptName: "nonexistent",
			newMeta: &ScriptMeta{
				Name: "nonexistent",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.UpdateScript(tt.scriptName, tt.newMeta, tt.newContent)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// 如果更新成功，验证更新结果
			if tt.newMeta != nil {
				script, err := manager.GetScript(tt.newMeta.Name)
				assert.NoError(t, err)
				assert.Equal(t, tt.newMeta.Description, script.Meta.Description)
			}

			if tt.newContent != "" {
				script, err := manager.GetScript(tt.scriptName)
				if err == nil { // 如果没有改名
					content, err := os.ReadFile(script.FilePath)
					assert.NoError(t, err)
					assert.Equal(t, tt.newContent, string(content))
				}
			}
		})
	}
}

// TestDeleteScript 测试删除脚本
func TestDeleteScript(t *testing.T) {
	rootDir, cleanup := setupTestEnv(t)
	defer cleanup()

	manager := createTestManager(t, rootDir)

	// 创建测试脚本
	script := TestScript{
		Meta: &ScriptMeta{
			Name:    "test1",
			Type:    PythonScript,
			Command: "test1.py",
		},
		Content: "print('test')",
	}
	err := manager.CreateScript(script.Meta, script.Content)
	require.NoError(t, err)

	// 保存文件路径以供后续检查
	scriptPath := filepath.Join(rootDir, script.Meta.Command)
	metaPath := scriptPath + MetaFileSuffix

	tests := []struct {
		name       string
		scriptName string
		wantErr    bool
	}{
		{
			name:       "delete existing script",
			scriptName: "test1",
			wantErr:    false,
		},
		{
			name:       "delete non-existent script",
			scriptName: "nonexistent",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.DeleteScript(tt.scriptName)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// 验证文件是否被删除
			if !tt.wantErr {
				// 检查脚本是否从管理器中删除
				_, err = manager.GetScript(tt.scriptName)
				assert.Error(t, err)

				// 检查文件是否从文件系统中删除
				_, err = os.Stat(scriptPath)
				assert.True(t, os.IsNotExist(err))
				_, err = os.Stat(metaPath)
				assert.True(t, os.IsNotExist(err))
			}
		})
	}
}

// TestScriptExecution 测试脚本执行
func TestScriptExecution(t *testing.T) {
	rootDir, cleanup := setupTestEnv(t)
	defer cleanup()

	manager := createTestManager(t, rootDir)

	// 创建测试脚本
	script := TestScript{
		Meta: &ScriptMeta{
			Name:    "test1",
			Type:    PythonScript,
			Command: "test1.py",
			Args: []ArgMeta{
				{Name: "input", Flag: "--input", Required: true},
			},
		},
		Content: "print('test')",
	}
	err := manager.CreateScript(script.Meta, script.Content)
	require.NoError(t, err)

	tests := []struct {
		name    string
		command types.Command
		wantErr bool
	}{
		{
			name: "execute with required args",
			command: types.Command{
				Command: "test1",
				Args:    []string{"--input", "test.txt"},
			},
			wantErr: false,
		},
		{
			name: "execute without required args",
			command: types.Command{
				Command: "test1",
				Args:    []string{},
			},
			wantErr: true,
		},
		{
			name: "execute non-existent script",
			command: types.Command{
				Command: "nonexistent",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &types.ExecuteContext{
				Command: tt.command,
			}
			result, err := manager.Execute(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

// TestScriptValidation 测试脚本验证
func TestScriptValidation(t *testing.T) {
	tests := []struct {
		name    string
		meta    *ScriptMeta
		wantErr bool
	}{
		{
			name: "valid script meta",
			meta: &ScriptMeta{
				Name:    "test1",
				Type:    PythonScript,
				Command: "test1.py",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			meta: &ScriptMeta{
				Type:    PythonScript,
				Command: "test.py",
			},
			wantErr: true,
		},
		{
			name: "missing command",
			meta: &ScriptMeta{
				Name: "test",
				Type: PythonScript,
			},
			wantErr: true,
		},
		{
			name: "invalid script type",
			meta: &ScriptMeta{
				Name:    "test",
				Type:    "invalid",
				Command: "test.py",
			},
			wantErr: true,
		},
		{
			name: "duplicate argument flags",
			meta: &ScriptMeta{
				Name:    "test",
				Type:    PythonScript,
				Command: "test.py",
				Args: []ArgMeta{
					{Name: "input1", Flag: "--input"},
					{Name: "input2", Flag: "--input"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateScriptMeta(tt.meta)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestScriptDiscovery 测试脚本发现功能
func TestScriptDiscovery(t *testing.T) {
	rootDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// 在不同目录创建测试脚本
	scripts := []TestScript{
		{
			Meta: &ScriptMeta{
				Name:    "test1",
				Type:    PythonScript,
				Command: "test1.py",
			},
			Content: "print('test1')",
		},
		{
			Meta: &ScriptMeta{
				Name:    "group1/test2",
				Type:    PythonScript,
				Command: "group1/test2.py",
			},
			Content: "print('test2')",
		},
		{
			Meta: &ScriptMeta{
				Name:    "group1/subgroup/test3",
				Type:    PythonScript,
				Command: "group1/subgroup/test3.py",
			},
			Content: "print('test3')",
		},
	}

	// 创建脚本文件
	manager := createTestManager(t, rootDir)
	for _, script := range scripts {
		err := manager.CreateScript(script.Meta, script.Content)
		require.NoError(t, err)
	}

	// 创建新的管理器来测试发现功能
	newManager := createTestManager(t, rootDir)

	// 验证所有脚本都被发现
	for _, script := range scripts {
		discovered, err := newManager.GetScript(script.Meta.Name)
		assert.NoError(t, err)
		assert.NotNil(t, discovered)
		assert.Equal(t, script.Meta.Name, discovered.Meta.Name)
	}

	// 测试 ListCommands
	commands := newManager.ListCommands()
	assert.Equal(t, len(scripts), len(commands))
}

// TestCreateScriptEdgeCases 测试建脚本的边缘情况
func TestCreateScriptEdgeCases(t *testing.T) {
	rootDir, cleanup := setupTestEnv(t)
	defer cleanup()

	manager := createTestManager(t, rootDir)

	tests := []struct {
		name    string
		setup   func() error
		script  TestScript
		wantErr bool
	}{
		{
			name: "create script with empty content",
			script: TestScript{
				Meta: &ScriptMeta{
					Name:    "empty",
					Type:    PythonScript,
					Command: "empty.py",
				},
				Content: "",
			},
			wantErr: false,
		},
		{
			name: "create script with special characters in name",
			script: TestScript{
				Meta: &ScriptMeta{
					Name:    "special@#$%",
					Type:    PythonScript,
					Command: "special.py",
				},
				Content: "print('test')",
			},
			wantErr: true,
		},
		{
			name: "create script with very long name",
			script: TestScript{
				Meta: &ScriptMeta{
					Name:    strings.Repeat("a", 256),
					Type:    PythonScript,
					Command: "long.py",
				},
				Content: "print('test')",
			},
			wantErr: true,
		},
		{
			name: "create script with windows line endings",
			script: TestScript{
				Meta: &ScriptMeta{
					Name:    "windows",
					Type:    PythonScript,
					Command: "windows.py",
				},
				Content: "print('line1')\r\nprint('line2')\r\n",
			},
			wantErr: false,
		},
		{
			name: "create script in directory without write permission",
			setup: func() error {
				noWriteDir := filepath.Join(rootDir, "no_write")
				if err := os.MkdirAll(noWriteDir, 0555); err != nil {
					return err
				}
				return nil
			},
			script: TestScript{
				Meta: &ScriptMeta{
					Name:    "no_write/test",
					Type:    PythonScript,
					Command: "no_write/test.py",
				},
				Content: "print('test')",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				err := tt.setup()
				require.NoError(t, err)
			}

			err := manager.CreateScript(tt.script.Meta, tt.script.Content)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// 验证文件内容
			if !tt.wantErr {
				content, err := os.ReadFile(filepath.Join(rootDir, tt.script.Meta.Command))
				assert.NoError(t, err)
				assert.Equal(t, tt.script.Content, string(content))
			}
		})
	}
}

// TestMetadataValidation 测试元数据验证的边缘情况
func TestMetadataValidation(t *testing.T) {
	tests := []struct {
		name    string
		meta    *ScriptMeta
		wantErr bool
	}{
		{
			name:    "nil metadata",
			meta:    nil,
			wantErr: true,
		},
		{
			name: "empty fields",
			meta: &ScriptMeta{
				Name:    "",
				Type:    "",
				Command: "",
			},
			wantErr: true,
		},
		{
			name: "invalid argument name",
			meta: &ScriptMeta{
				Name:    "test",
				Type:    PythonScript,
				Command: "test.py",
				Args: []ArgMeta{
					{Name: "", Flag: "--flag"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid flag format",
			meta: &ScriptMeta{
				Name:    "test",
				Type:    PythonScript,
				Command: "test.py",
				Args: []ArgMeta{
					{Name: "arg", Flag: "invalid"},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate argument names",
			meta: &ScriptMeta{
				Name:    "test",
				Type:    PythonScript,
				Command: "test.py",
				Args: []ArgMeta{
					{Name: "arg", Flag: "--flag1"},
					{Name: "arg", Flag: "--flag2"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateScriptMeta(tt.meta)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestScriptPathHandling 测试路径处理的边缘情况
func TestScriptPathHandling(t *testing.T) {
	rootDir, cleanup := setupTestEnv(t)
	defer cleanup()

	manager := createTestManager(t, rootDir)

	tests := []struct {
		name    string
		script  TestScript
		wantErr bool
	}{
		{
			name: "create script with relative path",
			script: TestScript{
				Meta: &ScriptMeta{
					Name:    "relative",
					Type:    PythonScript,
					Command: "./subdir/test.py",
				},
				Content: "print('test')",
			},
			wantErr: true,
		},
		{
			name: "create script with absolute path",
			script: TestScript{
				Meta: &ScriptMeta{
					Name:    "absolute",
					Type:    PythonScript,
					Command: "/absolute/path/test.py",
				},
				Content: "print('test')",
			},
			wantErr: true,
		},
		{
			name: "create script with path traversal",
			script: TestScript{
				Meta: &ScriptMeta{
					Name:    "traversal",
					Type:    PythonScript,
					Command: "../test.py",
				},
				Content: "print('test')",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.CreateScript(tt.script.Meta, tt.script.Content)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

// TestConcurrentAccess 测试并发访问
func TestConcurrentAccess(t *testing.T) {
	rootDir, cleanup := setupTestEnv(t)
	defer cleanup()

	manager := createTestManager(t, rootDir)

	// 创建初始脚本
	script := TestScript{
		Meta: &ScriptMeta{
			Name:    "concurrent",
			Type:    PythonScript,
			Command: "concurrent.py",
		},
		Content: "print('test')",
	}
	err := manager.CreateScript(script.Meta, script.Content)
	require.NoError(t, err)

	// 并发访问测试
	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	// 并发读取
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := manager.GetScript("concurrent")
			if err != nil {
				errChan <- err
			}
		}()
	}

	// 并发更新
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			newContent := fmt.Sprintf("print('test%d')", i)
			err := manager.UpdateScript("concurrent", nil, newContent)
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// 检查否有错误发生
	for err := range errChan {
		assert.NoError(t, err)
	}
}

// TestScriptExecutionEdgeCases 测试脚本执行的边缘情况
func TestScriptExecutionEdgeCases(t *testing.T) {
	rootDir, cleanup := setupTestEnv(t)
	defer cleanup()

	manager := createTestManager(t, rootDir)

	// 创建测试脚本
	script := TestScript{
		Meta: &ScriptMeta{
			Name:    "test",
			Type:    PythonScript,
			Command: "test.py",
			Args: []ArgMeta{
				{Name: "input", Flag: "--input", Required: true},
				{Name: "optional", Flag: "--opt"},
			},
		},
		Content: "print('test')",
	}
	err := manager.CreateScript(script.Meta, script.Content)
	require.NoError(t, err)

	tests := []struct {
		name    string
		ctx     *types.ExecuteContext
		wantErr bool
	}{
		{
			name:    "execute with nil context",
			ctx:     nil,
			wantErr: true,
		},
		{
			name: "execute with nil command",
			ctx: &types.ExecuteContext{
				Command: types.Command{},
			},
			wantErr: true,
		},
		{
			name: "execute with invalid argument format",
			ctx: &types.ExecuteContext{
				Command: types.Command{
					Command: "test",
					Args:    []string{"--invalid-format"},
				},
			},
			wantErr: true,
		},
		{
			name: "execute with duplicate arguments",
			ctx: &types.ExecuteContext{
				Command: types.Command{
					Command: "test",
					Args:    []string{"--input", "file1", "--input", "file2"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := manager.Execute(tt.ctx)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// TestMultipleScriptTypes 测试不同类型脚本的创建和执行
func TestMultipleScriptTypes(t *testing.T) {
	rootDir, cleanup := setupTestEnv(t)
	defer cleanup()

	manager := createTestManager(t, rootDir)

	scripts := []TestScript{
		{
			Meta: &ScriptMeta{
				Name:        "python_script",
				Type:        PythonScript,
				Command:     "script.py",
				Description: "Python test script",
				Args: []ArgMeta{
					{Name: "input", Flag: "--input", Required: true},
				},
			},
			Content: `#!/usr/bin/env python3
import sys
print("Python script executed with args:", sys.argv[1:])`,
		},
		{
			Meta: &ScriptMeta{
				Name:        "js_script",
				Type:        JavaScriptScript,
				Command:     "script.js",
				Description: "JavaScript test script",
				Args: []ArgMeta{
					{Name: "input", Flag: "--input", Required: true},
				},
			},
			Content: `#!/usr/bin/env node
console.log("JavaScript script executed with args:", process.argv.slice(2));`,
		},
		{
			Meta: &ScriptMeta{
				Name:        "shell_script",
				Type:        ShellScript,
				Command:     "script.sh",
				Description: "Shell test script",
				Args: []ArgMeta{
					{Name: "input", Flag: "--input", Required: true},
				},
			},
			Content: `#!/bin/bash
echo "Shell script executed with args: $@"`,
		},
	}

	// 创建并测试每种类型的脚本
	for _, script := range scripts {
		t.Run(fmt.Sprintf("Create and execute %s", script.Meta.Type), func(t *testing.T) {
			// 创建脚本
			err := manager.CreateScript(script.Meta, script.Content)
			require.NoError(t, err)

			// 验证脚本创建
			createdScript, err := manager.GetScript(script.Meta.Name)
			require.NoError(t, err)
			assert.Equal(t, script.Meta.Type, createdScript.Meta.Type)

			// 验证文件权限
			info, err := os.Stat(filepath.Join(rootDir, script.Meta.Command))
			require.NoError(t, err)
			assert.True(t, info.Mode()&0111 != 0, "Script should be executable")

			// 执行脚本
			ctx := &types.ExecuteContext{
				Command: types.Command{
					Command: script.Meta.Name,
					Args:    []string{"--input", "test.txt"},
				},
			}
			result, err := manager.Execute(ctx)
			require.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
		})
	}
}

// TestScriptTypeSpecificFeatures 测试脚本类型特定的功能
func TestScriptTypeSpecificFeatures(t *testing.T) {
	rootDir, cleanup := setupTestEnv(t)
	defer cleanup()

	manager := createTestManager(t, rootDir)

	tests := []struct {
		name    string
		script  TestScript
		setup   func() error
		verify  func(t *testing.T, script *Script) error
		wantErr bool
	}{
		{
			name: "python script with requirements",
			script: TestScript{
				Meta: &ScriptMeta{
					Name:    "python_with_deps",
					Type:    PythonScript,
					Command: "deps.py",
				},
				Content: `#!/usr/bin/env python3
import requests  # 这会在没有安装依赖时失败
print("Test with dependencies")`,
			},
			wantErr: false, // 取决于环境是否安装了依赖
		},
		{
			name: "javascript script with npm packages",
			script: TestScript{
				Meta: &ScriptMeta{
					Name:    "js_with_deps",
					Type:    JavaScriptScript,
					Command: "deps.js",
				},
				Content: `#!/usr/bin/env node
try {
    require('lodash');  // 这会在没有安装依赖时失败
    console.log("Test with dependencies");
} catch (e) {
    process.exit(1);
}`,
			},
			wantErr: false, // 取决于环境是否安装了依赖
		},
		{
			name: "shell script with environment variables",
			script: TestScript{
				Meta: &ScriptMeta{
					Name:    "shell_with_env",
					Type:    ShellScript,
					Command: "env.sh",
				},
				Content: `#!/bin/bash
if [ -z "$TEST_ENV" ]; then
    echo "Environment variable TEST_ENV is required"
    exit 1
fi
echo "Environment variable value: $TEST_ENV"`,
			},
			setup: func() error {
				os.Setenv("TEST_ENV", "test_value")
				return nil
			},
			verify: func(t *testing.T, script *Script) error {
				ctx := &types.ExecuteContext{
					Command: types.Command{
						Command: script.Meta.Name,
					},
					Options: &types.ExecuteOptions{
						Env: map[string]string{
							"TEST_ENV": "test_value",
						},
					},
				}
				result, err := manager.Execute(ctx)
				if err != nil {
					return err
				}
				assert.Equal(t, 0, result.ExitCode)
				return nil
			},
			wantErr: false,
		},
		{
			name: "python script with unicode content",
			script: TestScript{
				Meta: &ScriptMeta{
					Name:    "python_unicode",
					Type:    PythonScript,
					Command: "unicode.py",
				},
				Content: `#!/usr/bin/env python3
# -*- coding: utf-8 -*-
print("Unicode test: 你好，世界！")`,
			},
			wantErr: false,
		},
		{
			name: "javascript script with ES6 features",
			script: TestScript{
				Meta: &ScriptMeta{
					Name:    "js_es6",
					Type:    JavaScriptScript,
					Command: "es6.js",
				},
				Content: `#!/usr/bin/env node
const test = async () => {
    await Promise.resolve();
    console.log("ES6 features test");
};
test();`,
			},
			wantErr: false,
		},
		{
			name: "shell script with complex command",
			script: TestScript{
				Meta: &ScriptMeta{
					Name:    "shell_complex",
					Type:    ShellScript,
					Command: "complex.sh",
					Args: []ArgMeta{
						{Name: "test", Flag: "--test", Required: false},
					},
				},
				Content: `#!/bin/bash
set -e
trap 'echo "Error on line $LINENO"' ERR
for i in {1..3}; do
    echo "Loop $i"
done
if [ "$1" = "--test" ]; then
    echo "Test argument received"
fi`,
			},
			verify: func(t *testing.T, script *Script) error {
				ctx := &types.ExecuteContext{
					Command: types.Command{
						Command: script.Meta.Name,
						Args:    []string{"--test"},
					},
				}
				result, err := manager.Execute(ctx)
				if err != nil {
					return err
				}
				assert.Equal(t, 0, result.ExitCode)
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 运行设置函数
			if tt.setup != nil {
				err := tt.setup()
				require.NoError(t, err)
			}

			// 创建脚本
			err := manager.CreateScript(tt.script.Meta, tt.script.Content)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// 获取创建的脚本
			script, err := manager.GetScript(tt.script.Meta.Name)
			require.NoError(t, err)

			// 运行验证函数
			if tt.verify != nil {
				err = tt.verify(t, script)
				assert.NoError(t, err)
			}
		})
	}
}

// TestScriptInteraction 测试脚本间的交互
func TestScriptInteraction(t *testing.T) {
	rootDir, cleanup := setupTestEnv(t)
	defer cleanup()

	manager := createTestManager(t, rootDir)

	// 创建一个生成数据的 Python 脚本
	pythonScript := TestScript{
		Meta: &ScriptMeta{
			Name:    "data_generator",
			Type:    PythonScript,
			Command: "generator.py",
		},
		Content: `#!/usr/bin/env python3
import json
data = {"message": "Hello from Python"}
with open("output.json", "w") as f:
    json.dump(data, f)`,
	}

	// 创建一个处理数据的 JavaScript 脚本
	jsScript := TestScript{
		Meta: &ScriptMeta{
			Name:    "data_processor",
			Type:    JavaScriptScript,
			Command: "processor.js",
		},
		Content: `#!/usr/bin/env node
const fs = require('fs');
const data = JSON.parse(fs.readFileSync('output.json'));
console.log("Processed:", data.message);`,
	}

	// 创建一个清理的 Shell 脚本
	shellScript := TestScript{
		Meta: &ScriptMeta{
			Name:    "cleanup",
			Type:    ShellScript,
			Command: "cleanup.sh",
		},
		Content: `#!/bin/bash
rm -f output.json
echo "Cleanup completed"`,
	}

	// 创建所有脚本
	scripts := []TestScript{pythonScript, jsScript, shellScript}
	for _, script := range scripts {
		err := manager.CreateScript(script.Meta, script.Content)
		require.NoError(t, err)
	}

	// 按顺序执行脚本
	for _, script := range scripts {
		ctx := &types.ExecuteContext{
			Command: types.Command{
				Command: script.Meta.Name,
			},
		}
		result, err := manager.Execute(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
	}
}
