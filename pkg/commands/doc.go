// Package commands 实现了 RunShell 的内置命令。
//
// 本包提供了以下类别的命令：
//
// 1. 文件系统命令：
//   - ls：列出目录内容
//   - cat：查看文件内容
//   - mkdir：创建目录
//   - rm：删除文件或目录
//   - cp：复制文件
//   - pwd：显示当前工作目录
//
// 2. 系统命令：
//   - ps：显示进程状态
//   - top：显示系统任务
//   - df：显示磁盘使用情况
//   - uname：显示系统信息
//   - env：显示环境变量
//   - kill：终止进程
//
// 3. 脚本命令：
//   - script：脚本管理和执行
//   - template：模板管理
//   - var：变量管理
//
// 4. 实用工具：
//   - echo：显示文本
//   - sleep：延时执行
//   - date：显示日期时间
//
// 使用示例：
//
//	// 注册所有内置命令
//	commands.RegisterAll(executor)
//
//	// 注册特定类别的命令
//	commands.RegisterFileSystemCommands(executor)
//	commands.RegisterSystemCommands(executor)
//
// 所有命令都实现了 types.CommandHandler 接口。
package commands
