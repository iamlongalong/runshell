// Package script 提供脚本管理和执行功能。
//
// 本包实现了以下主要功能：
//
// 1. 脚本管理：
//   - 脚本保存和加载
//   - 脚本版本控制
//   - 脚本元数据管理
//
// 2. 脚本执行：
//   - 支持多种脚本类型
//   - 参数传递和环境变量
//   - 执行状态跟踪
//
// 3. 脚本存储：
//   - 文件系统存储
//   - 命名空间管理
//   - 权限控制
//
// 4. 脚本模板：
//   - 模板定义和渲染
//   - 变量替换
//   - 条件判断
//
// 使用示例：
//
//	// 创建脚本管理器
//	manager, err := script.NewScriptManager("/scripts", executor)
//
//	// 保存脚本
//	scriptPath, err := manager.SaveScript("test.sh", content)
//
//	// 执行脚本
//	result, err := manager.ExecuteScript(scriptPath, workDir, args)
//
//	// 列出脚本
//	scripts, err := manager.ListScripts()
//
// 本包提供了完整的脚本生命周期管理功能。
package script
