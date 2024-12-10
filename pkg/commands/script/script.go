// Package script 提供脚本管理和执行功能。
package script

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/iamlongalong/runshell/pkg/log"
	"github.com/iamlongalong/runshell/pkg/types"
)

type ScriptType string

const (
	PythonScript     ScriptType = "python"
	JavaScriptScript ScriptType = "javascript"
	ShellScript      ScriptType = "shell"
	MetaFileSuffix   string     = ".meta.json" // 元数据文件后缀
	MaxNameLength               = 128          // 脚本名称最大长度
	MaxPathLength               = 256          // 路径最大长度
)

var (
	// 合法的脚本名称格式
	validNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_\-/]*[a-zA-Z0-9]$`)
	// 合法的参数标志格式
	validFlagPattern = regexp.MustCompile(`^--[a-zA-Z][a-zA-Z0-9\-]*$`)
)

// ScriptMeta 脚本描述
type ScriptMeta struct {
	Name        string     `json:"name"`        // 脚本名称
	Type        ScriptType `json:"type"`        // 脚本类型
	Command     string     `json:"command"`     // 主脚本文件
	Description string     `json:"description"` // 简短描述
	Args        []ArgMeta  `json:"args"`        // 参数列表
	Examples    []Example  `json:"examples"`    // 使用示例
}

// ArgMeta 参数定义
type ArgMeta struct {
	Name        string `json:"name"`        // 参数名
	Flag        string `json:"flag"`        // 参数标志，如 --input 或 -i
	Required    bool   `json:"required"`    // 是否必需
	Description string `json:"description"` // 参数说明
}

// Example 使用示例
type Example struct {
	Desc    string `json:"desc"`    // 示例说明
	Command string `json:"command"` // 完整的命令示例
}

// Script 表示一个可执行的脚本
type Script struct {
	Meta     *ScriptMeta
	Dir      string // 脚本所在目录
	FilePath string // 脚本文件完整路径
	MetaPath string // 描述文件路径
}

// ScriptManager 实现 types.Executor 接口
type ScriptManager struct {
	rootDir  string             // 脚本根目录
	scripts  map[string]*Script // 名称到脚本的映射
	options  *types.ExecuteOptions
	executor types.Executor // 执行器
}

type Config struct {
	RootDir  string
	Executor types.Executor
}

// NewScriptManager 创建新的脚本管理器
func NewScriptManager(cfg *Config, options *types.ExecuteOptions) (*ScriptManager, error) {
	sm := &ScriptManager{
		rootDir:  cfg.RootDir,
		scripts:  make(map[string]*Script),
		executor: cfg.Executor,
		options:  options,
	}

	// 初始化时扫描目录
	if err := sm.scanScripts(); err != nil {
		return nil, fmt.Errorf("scan scripts error: %w", err)
	}

	return sm, nil
}

// Name 实现 Executor 接口
func (sm *ScriptManager) Name() string {
	return "script"
}

// Execute 实现 Executor 接口
func (sm *ScriptManager) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	if ctx == nil {
		return nil, fmt.Errorf("execute context is nil")
	}

	if len(ctx.Command.Command) == 0 {
		return nil, fmt.Errorf("script name required")
	}

	if ctx.Options == nil {
		ctx.Options = &types.ExecuteOptions{}
	}

	ctx.Options = ctx.Options.Merge(sm.options)

	// 优先使用传入的执行器
	if ctx.Executor == nil {
		ctx.Executor = sm.executor
	}

	scriptName := ctx.Command.Command
	script, ok := sm.scripts[scriptName]
	if !ok {
		return nil, fmt.Errorf("script not found: %s", scriptName)
	}

	// 验证参数
	if err := script.ValidateArgs(ctx.Command.Args); err != nil {
		return nil, err
	}

	newctx := ctx.Copy()
	newctx.Executor = sm.executor

	// 执行脚本
	return script.Execute(newctx)
}

// ListCommands 实现 Executor 接口
func (sm *ScriptManager) ListCommands() []types.CommandInfo {
	commands := make([]types.CommandInfo, 0, len(sm.scripts))
	for _, script := range sm.scripts {
		commands = append(commands, types.CommandInfo{
			Name:        script.Meta.Name,
			Description: script.Meta.Description,
			Usage:       sm.generateUsage(script),
		})
	}
	return commands
}

// Close 实现 Executor 接口
func (sm *ScriptManager) Close() error {
	sm.scripts = nil
	return nil
}

// getScriptNameFromPath 从路径中获取脚本名称
func getScriptNameFromPath(rootDir, path string) string {
	// 获取相对于根目录的路径
	relPath, err := filepath.Rel(rootDir, path)
	if err != nil {
		return filepath.Base(path)
	}

	// 将路径分隔符替换为命名分隔符
	return strings.ReplaceAll(relPath, string(filepath.Separator), "/")
}

// getScriptPathFromName 从脚本名称获取实际路径
func getScriptPathFromName(rootDir, name string) string {
	// 将命名分隔符替换回路径分隔符
	relativePath := strings.ReplaceAll(name, "/", string(filepath.Separator))
	return filepath.Join(rootDir, relativePath)
}

// scanScripts 扫描脚本目录
func (sm *ScriptManager) scanScripts() error {
	return filepath.Walk(sm.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 查找元数据文件
		if strings.HasSuffix(path, MetaFileSuffix) {
			script, err := sm.loadScript(path)
			if err != nil {
				log.Error("Failed to load script at %s: %v", path, err)
				return nil // 继续扫描其他脚本
			}

			sm.scripts[script.Meta.Name] = script
		}

		return nil
	})
}

// loadScript 加载单个脚本
func (sm *ScriptManager) loadScript(metaPath string) (*Script, error) {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var meta ScriptMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	// 获取脚本文件路径
	scriptDir := filepath.Dir(metaPath)
	scriptPath := filepath.Join(scriptDir, meta.Command)
	if _, err := os.Stat(scriptPath); err != nil {
		return nil, fmt.Errorf("script file not found: %s", scriptPath)
	}

	// 如果meta中没有指定名称，从路径生成
	if meta.Name == "" {
		meta.Name = getScriptNameFromPath(sm.rootDir, scriptPath)
	}

	return &Script{
		Meta:     &meta,
		Dir:      scriptDir,
		FilePath: scriptPath,
		MetaPath: metaPath,
	}, nil
}

// generateUsage 生成命令使用说明
func (sm *ScriptManager) generateUsage(script *Script) string {
	var flags []string
	for _, arg := range script.Meta.Args {
		if arg.Required {
			flags = append(flags, fmt.Sprintf("%s <value>", arg.Flag))
		} else {
			flags = append(flags, fmt.Sprintf("[%s <value>]", arg.Flag))
		}
	}
	return fmt.Sprintf("%s %s", script.Meta.Name, strings.Join(flags, " "))
}

// generateExamples 生成示例说明
func (sm *ScriptManager) generateExamples(script *Script) []string {
	examples := make([]string, len(script.Meta.Examples))
	for i, example := range script.Meta.Examples {
		examples[i] = fmt.Sprintf("%s\n%s", example.Desc, example.Command)
	}
	return examples
}

// ValidateArgs 验证参数
func (s *Script) ValidateArgs(args []string) error {
	// 检查必需参数
	requiredFlags := make(map[string]bool)
	for _, arg := range s.Meta.Args {
		if arg.Required {
			requiredFlags[arg.Flag] = false
		}
	}

	// 检查提供的参数
	usedFlags := make(map[string]bool)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			// 检查是否是有效的标志
			validFlag := false
			for _, metaArg := range s.Meta.Args {
				if arg == metaArg.Flag {
					// 检查重复使用
					if usedFlags[arg] {
						return fmt.Errorf("duplicate argument flag: %s", arg)
					}
					usedFlags[arg] = true

					if metaArg.Required {
						requiredFlags[metaArg.Flag] = true
					}
					validFlag = true
					break
				}
			}
			if !validFlag {
				return fmt.Errorf("unknown argument flag: %s", arg)
			}

			// 跳过标志的值
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
			}
		}
	}

	// 检查是否所有必需参数都提供了
	for flag, provided := range requiredFlags {
		if !provided {
			return fmt.Errorf("required argument not provided: %s", flag)
		}
	}

	return nil
}

// Execute 执行脚本
func (s *Script) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	// 构建命令
	cmd, err := s.buildCommand(ctx.Command.Args)
	if err != nil {
		log.Error("Failed to build command: %v", err)
		return nil, err
	}

	log.Debug("Built command: %v", cmd)
	log.Debug("Script type: %s", s.Meta.Type)
	log.Debug("Script file: %s", s.FilePath)
	log.Debug("Script dir: %s", s.Dir)

	newctx := ctx.Copy()
	newctx.Command = *cmd

	// 执行命令
	log.Debug("Executing command with context: %+v", newctx)
	result, err := ctx.Executor.Execute(newctx)
	if err != nil {
		log.Error("Command execution failed: %v", err)
		return nil, err
	}

	return result, nil
}

// buildCommand 构建执行命令
func (s *Script) buildCommand(args []string) (*types.Command, error) {
	// 获取相对于工作目录的路径
	scriptName := filepath.Base(s.Meta.Command)

	var cmdArgs []string

	switch s.Meta.Type {
	case PythonScript:
		cmdArgs = append([]string{scriptName}, args...)
		return &types.Command{
			Command: "python",
			Args:    cmdArgs,
		}, nil
	case JavaScriptScript:
		cmdArgs = append([]string{scriptName}, args...)
		return &types.Command{
			Command: "node",
			Args:    cmdArgs,
		}, nil
	case ShellScript:
		cmdArgs = append([]string{scriptName}, args...)
		return &types.Command{
			Command: "bash",
			Args:    cmdArgs,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported script type: %s", s.Meta.Type)
	}
}

// CreateScript 创建新脚本
func (sm *ScriptManager) CreateScript(meta *ScriptMeta, scriptContent string) error {
	// 验证元数据
	if err := validateScriptMeta(meta); err != nil {
		return fmt.Errorf("invalid script meta: %w", err)
	}

	// 验证路径
	if err := validatePath(sm.rootDir, meta.Command); err != nil {
		return fmt.Errorf("invalid script path: %w", err)
	}

	// 检查脚本是否已存在
	if _, exists := sm.scripts[meta.Name]; exists {
		return fmt.Errorf("script already exists: %s", meta.Name)
	}

	// 获取完整的脚本路径
	scriptPath := filepath.Join(sm.rootDir, meta.Command)

	// 确保目录存在
	scriptDir := filepath.Dir(scriptPath)
	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		return fmt.Errorf("failed to create script directory: %w", err)
	}

	// 写入脚本文件
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return fmt.Errorf("failed to write script file: %w", err)
	}

	// 写入元数据文件
	metaPath := scriptPath + MetaFileSuffix
	metaData, err := json.MarshalIndent(meta, "", "    ")
	if err != nil {
		os.Remove(scriptPath)
		return fmt.Errorf("failed to marshal meta data: %w", err)
	}

	if err := os.WriteFile(metaPath, metaData, 0644); err != nil {
		os.Remove(scriptPath)
		return fmt.Errorf("failed to write meta file: %w", err)
	}

	// 加载新创建的脚本
	script, err := sm.loadScript(metaPath)
	if err != nil {
		os.Remove(scriptPath)
		os.Remove(metaPath)
		return fmt.Errorf("failed to load created script: %w", err)
	}

	sm.scripts[meta.Name] = script
	return nil
}

// UpdateScript 更新现有脚本
func (sm *ScriptManager) UpdateScript(name string, meta *ScriptMeta, scriptContent string) error {
	// 检查脚本是否存在
	script, exists := sm.scripts[name]
	if !exists {
		return fmt.Errorf("script not found: %s", name)
	}

	// 如果提供了新的元数据，验证并更新
	if meta != nil {
		if err := validateScriptMeta(meta); err != nil {
			return fmt.Errorf("invalid script meta: %w", err)
		}

		// 如果更改了名称，确保新名称未被使用
		if meta.Name != name {
			if _, exists := sm.scripts[meta.Name]; exists {
				return fmt.Errorf("script name already exists: %s", meta.Name)
			}
		}

		// 如果命令名发生变化，需要处理文件重命名
		if meta.Command != script.Meta.Command {
			newScriptPath := getScriptPathFromName(sm.rootDir, meta.Command)
			newMetaPath := newScriptPath + MetaFileSuffix

			// 确保目标目录存在
			newScriptDir := filepath.Dir(newScriptPath)
			if err := os.MkdirAll(newScriptDir, 0755); err != nil {
				return fmt.Errorf("failed to create script directory: %w", err)
			}

			// 重命名脚本文件
			if err := os.Rename(script.FilePath, newScriptPath); err != nil {
				return fmt.Errorf("failed to rename script file: %w", err)
			}

			// 重命名元数据文件
			if err := os.Rename(script.MetaPath, newMetaPath); err != nil {
				// 如果元数据重命名失败，回滚脚本文件重命名
				os.Rename(newScriptPath, script.FilePath)
				return fmt.Errorf("failed to rename meta file: %w", err)
			}

			script.FilePath = newScriptPath
			script.MetaPath = newMetaPath
			script.Dir = filepath.Dir(newScriptPath)
		}

		// 更新元数据文件
		metaData, err := json.MarshalIndent(meta, "", "    ")
		if err != nil {
			return fmt.Errorf("failed to marshal meta data: %w", err)
		}

		if err := os.WriteFile(script.MetaPath, metaData, 0644); err != nil {
			return fmt.Errorf("failed to write meta file: %w", err)
		}
	}

	// 如果提供了新脚本内容，更新脚本文件
	if scriptContent != "" {
		if err := os.WriteFile(script.FilePath, []byte(scriptContent), 0755); err != nil {
			return fmt.Errorf("failed to write script file: %w", err)
		}
	}

	// 重新加载脚本
	updatedScript, err := sm.loadScript(script.MetaPath)
	if err != nil {
		return fmt.Errorf("failed to reload updated script: %w", err)
	}

	// 更新映射
	if meta != nil && meta.Name != name {
		delete(sm.scripts, name)
		sm.scripts[meta.Name] = updatedScript
	} else {
		sm.scripts[name] = updatedScript
	}

	return nil
}

// DeleteScript 删除脚本
func (sm *ScriptManager) DeleteScript(name string) error {
	script, exists := sm.scripts[name]
	if !exists {
		return fmt.Errorf("script not found: %s", name)
	}

	// 删除脚本文件和元数据文件
	if err := os.Remove(script.FilePath); err != nil {
		return fmt.Errorf("failed to delete script file: %w", err)
	}
	if err := os.Remove(script.MetaPath); err != nil {
		return fmt.Errorf("failed to delete meta file: %w", err)
	}

	// 从映射中删除
	delete(sm.scripts, name)
	return nil
}

// GetScript 获取脚本详情
func (sm *ScriptManager) GetScript(name string) (*Script, error) {
	script, exists := sm.scripts[name]
	if !exists {
		return nil, fmt.Errorf("script not found: %s", name)
	}
	return script, nil
}

// ListScripts 列出所有脚本
func (sm *ScriptManager) ListScripts() []*Script {
	scripts := make([]*Script, 0, len(sm.scripts))
	for _, script := range sm.scripts {
		scripts = append(scripts, script)
	}
	return scripts
}

// validateScriptMeta 验证脚本元数据
func validateScriptMeta(meta *ScriptMeta) error {
	if meta == nil {
		return fmt.Errorf("meta is nil")
	}

	// 验证名称
	if meta.Name == "" {
		return fmt.Errorf("script name is required")
	}
	if len(meta.Name) > MaxNameLength {
		return fmt.Errorf("script name too long (max %d characters)", MaxNameLength)
	}
	if !validNamePattern.MatchString(meta.Name) {
		return fmt.Errorf("invalid script name format: %s", meta.Name)
	}

	// 验证命令
	if meta.Command == "" {
		return fmt.Errorf("script command is required")
	}
	if len(meta.Command) > MaxPathLength {
		return fmt.Errorf("script command path too long (max %d characters)", MaxPathLength)
	}

	// 验证脚本类型
	switch meta.Type {
	case PythonScript, JavaScriptScript, ShellScript:
		// 有效的脚本类型
	default:
		return fmt.Errorf("unsupported script type: %s", meta.Type)
	}

	// 验证参数
	argNames := make(map[string]bool)
	flagNames := make(map[string]bool)
	for _, arg := range meta.Args {
		// 验证参数名
		if arg.Name == "" {
			return fmt.Errorf("argument name is required")
		}
		if !validNamePattern.MatchString(arg.Name) {
			return fmt.Errorf("invalid argument name format: %s", arg.Name)
		}
		if argNames[arg.Name] {
			return fmt.Errorf("duplicate argument name: %s", arg.Name)
		}
		argNames[arg.Name] = true

		// 验证参数标志
		if arg.Flag == "" {
			return fmt.Errorf("argument flag is required")
		}
		if !validFlagPattern.MatchString(arg.Flag) {
			return fmt.Errorf("invalid argument flag format: %s", arg.Flag)
		}
		if flagNames[arg.Flag] {
			return fmt.Errorf("duplicate argument flag: %s", arg.Flag)
		}
		flagNames[arg.Flag] = true
	}

	return nil
}

// validatePath 验证路径安全性
func validatePath(rootDir, path string) error {
	// 检查路径长度
	if len(path) > MaxPathLength {
		return fmt.Errorf("path too long (max %d characters)", MaxPathLength)
	}

	// 检查路径中的特殊字符
	if strings.Contains(path, "./") || strings.Contains(path, ".\\") {
		return fmt.Errorf("relative path notation not allowed: %s", path)
	}

	// 规范化路径
	cleanPath := filepath.Clean(path)

	// 不允许绝对路径
	if filepath.IsAbs(cleanPath) {
		return fmt.Errorf("absolute path not allowed: %s", path)
	}

	// 不允许父目录引用
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("parent directory reference not allowed: %s", path)
	}

	// 验证最终路径在根目录下
	fullPath := filepath.Join(rootDir, cleanPath)
	if !strings.HasPrefix(fullPath, rootDir) {
		return fmt.Errorf("path escapes root directory: %s", path)
	}

	return nil
}
