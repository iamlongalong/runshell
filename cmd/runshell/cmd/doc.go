// Package cmd 实现了 RunShell 的命令行命令。
//
// 本包使用 cobra 框架实现了以下命令：
//
// 1. root - 根命令：
//   - 全局标志定义
//   - 版本信息
//   - 通用配置
//
// 2. exec - 执行命令：
//   - 命令执行
//   - 参数解析
//   - 选项处理
//
// 3. server - 服务器命令：
//   - HTTP 服务器
//   - 信号处理
//   - 优雅关闭
//
// 4. shell - Shell 命令：
//   - 交互式界面
//   - 命令历史
//   - 帮助系统
//
// 命令结构：
//
//	runshell
//	├── exec
//	│   ├── --workdir
//	│   └── --env
//	├── server
//	│   └── --http
//	└── shell
//
// 使用示例：
//
//	// 执行命令
//	rootCmd.AddCommand(execCmd)
//
//	// 启动服务器
//	rootCmd.AddCommand(serverCmd)
//
//	// 启动 Shell
//	rootCmd.AddCommand(shellCmd)
//
// 本包中的所有命令都支持上下文取消和优雅关闭。
package cmd
