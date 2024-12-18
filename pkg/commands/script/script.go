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
	executor types.Executor
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
	scriptPath := filepath.Join(scriptDir, filepath.Base(meta.Command))
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
		executor: sm.executor,
	}, nil
}

// generateUsage ��成命令使用说明
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
	result, err := ctx.Executor.ExecuteCommand(newctx)
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

// validatePath 验证脚本路径
func validatePath(rootDir, path string) error {
	// 检查是否是绝对路径
	if filepath.IsAbs(path) {
		return fmt.Errorf("absolute path not allowed")
	}

	// 检查路径中是否包含 ".." 或 "."
	if strings.Contains(path, "..") || strings.Contains(path, "./") {
		return fmt.Errorf("path traversal not allowed")
	}

	// 检查路径是否在根目录下
	fullPath := filepath.Join(rootDir, path)
	rel, err := filepath.Rel(rootDir, fullPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("path must be under root directory")
	}

	return nil
}

// getScriptPermissions 根据脚本类型获取文件权限
func getScriptPermissions(scriptType ScriptType) os.FileMode {
	switch scriptType {
	case PythonScript, JavaScriptScript, ShellScript:
		return 0755 // 可执行文件
	default:
		return 0644 // 普通文件
	}
}

// CreateScript 创建新脚本
func (sm *ScriptManager) CreateScript(meta *ScriptMeta, content string) error {
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

	// 构建脚本路径
	scriptPath := filepath.Join(sm.rootDir, meta.Command)
	metaPath := scriptPath + MetaFileSuffix

	// 确保目录存在
	scriptDir := filepath.Dir(scriptPath)
	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		return fmt.Errorf("failed to create script directory: %w", err)
	}

	// 写入脚本文件
	if err := os.WriteFile(scriptPath, []byte(content), getScriptPermissions(meta.Type)); err != nil {
		return fmt.Errorf("failed to write script file: %w", err)
	}

	// 写入元数据文件
	metaData, err := json.MarshalIndent(meta, "", "    ")
	if err != nil {
		os.Remove(scriptPath) // 清理脚本文件
		return fmt.Errorf("failed to marshal meta data: %w", err)
	}

	if err := os.WriteFile(metaPath, metaData, 0644); err != nil {
		os.Remove(scriptPath) // 清理脚本文件
		return fmt.Errorf("failed to write meta file: %w", err)
	}

	// 加载脚本
	script, err := sm.loadScript(metaPath)
	if err != nil {
		os.Remove(scriptPath) // 清理脚本文件
		os.Remove(metaPath)   // 清理元数据文件
		return fmt.Errorf("failed to load script: %w", err)
	}

	// 添加到映射
	sm.scripts[meta.Name] = script

	return nil
}

// UpdateScript 更新脚本
func (sm *ScriptManager) UpdateScript(name string, newMeta *ScriptMeta, newContent string) error {
	script, ok := sm.scripts[name]
	if !ok {
		return fmt.Errorf("script not found: %s", name)
	}

	// 如果提供了新的元数据
	if newMeta != nil {
		// 如果要移动到新目录，确保目录存在
		if newMeta.Command != script.Meta.Command {
			newDir := filepath.Dir(filepath.Join(sm.rootDir, newMeta.Command))
			if err := os.MkdirAll(newDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		}

		// 更新元数据文件
		data, err := json.MarshalIndent(newMeta, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal meta: %w", err)
		}

		// 如果需要移动文件
		if newMeta.Command != script.Meta.Command {
			// 新的文件路径
			newScriptPath := filepath.Join(sm.rootDir, newMeta.Command)
			newMetaPath := newScriptPath + MetaFileSuffix

			// 移动脚本文件
			if err := os.Rename(script.FilePath, newScriptPath); err != nil {
				return fmt.Errorf("failed to move script file: %w", err)
			}

			// 写入新的元数据文件
			if err := os.WriteFile(newMetaPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write meta file: %w", err)
			}

			// 删除旧的元数据文件
			if err := os.Remove(script.MetaPath); err != nil {
				log.Error("Failed to remove old meta file: %v", err)
			}

			// 更新脚本对象
			script.Meta = newMeta
			script.FilePath = newScriptPath
			script.MetaPath = newMetaPath
			script.Dir = filepath.Dir(newScriptPath)

			// 更新映射
			delete(sm.scripts, name)
			sm.scripts[newMeta.Name] = script
		} else {
			// 仅更新元数��
			if err := os.WriteFile(script.MetaPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write meta file: %w", err)
			}
			script.Meta = newMeta
		}
	}

	// 如果提供了新的内容
	if newContent != "" {
		if err := os.WriteFile(script.FilePath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to write script file: %w", err)
		}
	}

	return nil
}

// DeleteScript 删除脚本
func (sm *ScriptManager) DeleteScript(name string) error {
	script, ok := sm.scripts[name]
	if !ok {
		return fmt.Errorf("script not found: %s", name)
	}

	// 删除脚本文件
	if err := os.Remove(script.FilePath); err != nil {
		return fmt.Errorf("failed to delete script file: %w", err)
	}

	// 删除元数据文件
	if err := os.Remove(script.MetaPath); err != nil {
		log.Error("Failed to delete meta file: %v", err)
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
