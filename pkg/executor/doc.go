// Package executor 实现了 RunShell 框架的命令执行器。
//
// 本包提供了以下主要功能：
//
// 1. 本地命令执行器 (LocalExecutor)：
//   - 在本地系统中执行命令
//   - 支持工作目录和环境变量设置
//   - 提供命令注册和管理功能
//
// 2. Docker 命令执行器 (DockerExecutor)：
//   - 在 Docker 容器中执行命令
//   - 支持镜像选择和容器配置
//   - 提供容器生命周期管理
//
// 3. 审计执行器 (AuditedExecutor)：
//   - 包装其他执行器，提供审计功能
//   - 记录命令执行的详细信息
//   - 支持审计日志的持久化
//
// 使用示例：
//
//	// 创建本地执行器
//	localExec := executor.NewLocalExecutor()
//
//	// 创建 Docker 执行器
//	dockerExec, err := executor.NewDockerExecutor("ubuntu:latest")
//
//	// 创建审计执行器
//	auditedExec := executor.NewAuditedExecutor(localExec, auditor)
//
// 本包实现了 types.Executor 接口，提供了统一的命令执行接口。
package executor
