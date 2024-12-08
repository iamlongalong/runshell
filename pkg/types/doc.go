// Package types 定义了 RunShell 框架的核心类型和接口。
//
// 本包包含以下主要组件：
//
// 1. 执行器接口 (Executor)：
//   - 定义了命令执行的核心接口
//   - 支持命令注册和管理
//   - 提供命令执行上下文和结果处理
//
// 2. 命令相关类型：
//   - Command：表示一个可执行的命令
//   - CommandHandler：命令处理器接口
//   - CommandFilter：命令过滤器
//
// 3. 执行选项和上下文：
//   - ExecuteOptions：命令执行选项
//   - ExecuteContext：执行上下文
//   - ExecuteResult：执行结果
//
// 4. 资源和用户：
//   - ResourceUsage：资源使用统计
//   - User：用户信息
//
// 5. 错误处理：
//   - ExecuteError：执行错误类型
//   - 预定义错误常量
//
// 使用示例：
//
//	exec := NewExecutor()
//	result, err := exec.Execute(ctx, "ls", []string{"-l"}, &ExecuteOptions{
//	    WorkDir: "/tmp",
//	    Env: map[string]string{"PATH": "/usr/bin"},
//	})
package types
