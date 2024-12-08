// Package audit 提供命令执行的审计功能。
//
// 本包实现了以下主要功能：
//
// 1. 审计日志记录：
//   - 记录命令执行的详细信息
//   - 包括执行时间、参数、结果等
//   - 支持自定义元数据
//
// 2. 日志持久化：
//   - 支持将审计日志写入文件
//   - 按日期自动分割日志文件
//   - 提供日志文件轮转功能
//
// 3. 日志查询：
//   - 支持按时间范围查询
//   - 支持按命令名称查询
//   - 支持按执行结果查询
//
// 4. 安全特性：
//   - 日志文件权限控制
//   - 防篡改机制
//   - 日志完整性校验
//
// 使用示例：
//
//	// 创建审计器
//	auditor, err := audit.NewAuditor("/var/log/runshell")
//
//	// 记录命令执行
//	auditor.LogCommandExecution(ctx, result)
//
//	// 查询审计日志
//	logs, err := auditor.GetAuditLogs(startTime, endTime)
//
// 本包通常与 executor 包配合使用，为命令执行提供审计功能。
package audit
