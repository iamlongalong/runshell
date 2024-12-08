// Package script 提供脚本管理和执行功能。
// 本文件实现了脚本管理器的核心功能。
package script

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/iamlongalong/runshell/pkg/types"
)

// ScriptManager 管理脚本的存储和执行。
// 特性：
// - 脚本文件管理（保存、列出、获取内容、删除）
// - 脚本执行
// - 自动文件命名和版本控制
// - 工作目录验证
type ScriptManager struct {
	scriptDir string         // 脚本存储目录
	executor  types.Executor // 用于执行脚本的执行器
}

// NewScriptManager 创建一个新的脚本管理器实例。
// 参数：
//   - scriptDir：脚本存储目录
//   - executor：用于执行脚本的执行器
//
// 返回值：
//   - *ScriptManager：脚本管理器实例
//   - error：创建过程中的错误
func NewScriptManager(scriptDir string, executor types.Executor) (*ScriptManager, error) {
	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create script directory: %v", err)
	}

	return &ScriptManager{
		scriptDir: scriptDir,
		executor:  executor,
	}, nil
}

// SaveScript 保存脚本到脚本目录。
// 功能：
// - 自动添加 .sh 扩展名
// - 处理文件名冲突（添加时间戳）
// - 设置适当的文件权限
//
// 参数：
//   - name：脚本名称
//   - content：脚本内容
//
// 返回值：
//   - string：保存的脚本文件路径
//   - error：保存过程中的错误
func (sm *ScriptManager) SaveScript(name string, content []byte) (string, error) {
	// 生成唯一的脚本文件名
	timestamp := time.Now().Format("20060102150405")
	filename := name
	if !strings.HasSuffix(filename, ".sh") {
		filename += ".sh"
	}
	scriptPath := filepath.Join(sm.scriptDir, filename)

	// 如果文件已存在，添加时间戳
	if _, err := os.Stat(scriptPath); err == nil {
		filename = fmt.Sprintf("%s_%s.sh", strings.TrimSuffix(name, ".sh"), timestamp)
		scriptPath = filepath.Join(sm.scriptDir, filename)
	}

	// 保存脚本内容
	if err := os.WriteFile(scriptPath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to save script: %v", err)
	}

	return scriptPath, nil
}

// ExecuteScript 执行已保存的脚本。
// 功能：
// - 验证工作目录
// - 支持相对路径和绝对路径
// - 传递命令行参数
// - 配置执行环境
//
// 参数：
//   - scriptPath：脚本文件路径
//   - workDir：工作目录
//   - args：传递给脚本的参数
//
// 返回值：
//   - *types.ExecuteResult：执行结果
//   - error：执行过程中的错误
func (sm *ScriptManager) ExecuteScript(scriptPath string, workDir string, args []string) (*types.ExecuteResult, error) {
	// 验证工作目录
	if workDir == "" {
		return nil, fmt.Errorf("workdir must be specified for script execution")
	}

	// 验证工作目录存在
	if _, err := os.Stat(workDir); err != nil {
		return nil, fmt.Errorf("workdir does not exist: %v", err)
	}

	// 如果传入的是相对路径，转换为完整路径
	if !filepath.IsAbs(scriptPath) {
		scriptPath = filepath.Join(sm.scriptDir, scriptPath)
	}

	// 验证脚本文件存在
	if _, err := os.Stat(scriptPath); err != nil {
		return nil, fmt.Errorf("script file does not exist: %v", err)
	}

	// 准备执行选项
	opts := &types.ExecuteOptions{
		WorkDir: workDir,
		Env:     make(map[string]string),
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}

	// 执行脚本
	return sm.executor.Execute(context.Background(), "sh", append([]string{scriptPath}, args...), opts)
}

// ListScripts 列出所有已保存的脚本。
// 功能：
// - 只列出 .sh 文件
// - 忽略子目录
// - 返回文件名列表
//
// 返回值：
//   - []string：脚本文件名列表
//   - error：列出过程中的错误
func (sm *ScriptManager) ListScripts() ([]string, error) {
	var scripts []string

	entries, err := os.ReadDir(sm.scriptDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read script directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sh" {
			scripts = append(scripts, entry.Name())
		}
	}

	return scripts, nil
}

// GetScriptContent 获取已保存脚本的内容。
// 参数：
//   - scriptName：脚本文件名
//
// 返回值：
//   - []byte：脚本内容
//   - error：读取过程中的错误
func (sm *ScriptManager) GetScriptContent(scriptName string) ([]byte, error) {
	scriptPath := filepath.Join(sm.scriptDir, scriptName)
	return os.ReadFile(scriptPath)
}

// DeleteScript 删除已保存的脚本。
// 参数：
//   - scriptName：要删除的脚本文件名
//
// 返回值：
//   - error：删除过程中的错误
func (sm *ScriptManager) DeleteScript(scriptName string) error {
	scriptPath := filepath.Join(sm.scriptDir, scriptName)
	return os.Remove(scriptPath)
}
