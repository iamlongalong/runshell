// Package server 实现了 RunShell 的 HTTP API 服务。
//
// 本包提供了以下主要功能：
//
// 1. HTTP API：
//   - 命令执行 API
//   - 命令管理 API
//   - 健康检查 API
//
// 2. 服务器管理：
//   - 优雅启动和关闭
//   - 连接管理
//   - 请求超时控制
//
// 3. 安全特性：
//   - 请求认证
//   - 访问控制
//   - 请求限流
//
// 4. 监控和统计：
//   - 请求计数
//   - 响应时间统计
//   - 错误率统计
//
// API 端点：
//
//	GET  /health          - 健康检查
//	POST /exec            - 执行命令
//	GET  /commands        - 列出命令
//	GET  /commands/{name} - 获取命令信息
//
// 使用示例：
//
//	// 创建服务器
//	srv := server.NewServer(executor, ":8080")
//
//	// 启动服务器
//	go srv.Start()
//
//	// 等待服务器就绪
//	srv.WaitForReady()
//
//	// 关闭服务器
//	srv.Stop(ctx)
//
// 本包提供了 RESTful API，使得 RunShell 可以作为服务运行。
package server
