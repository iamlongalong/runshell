# Pipeline Example

本示例展示了如何使用 RunShell 的管道功能执行命令链。

## 功能特点

1. 命令链支持
   - 支持标准 Unix 管道操作
   - 多级命令链执行
   - 环境变量传递
   - 错误处理和状态报告

2. 执行器特性
   - 标准 Unix 命令支持（ls, grep, wc 等）
   - 输入输出流管理
   - 执行状态跟踪
   - 优雅的错误处理

## 示例说明

### 1. 简单管道命令
```bash
ls -la | grep go
```
列出当前目录内容并过滤包含 "go" 的行。这个示例展示了基本的管道功能。

### 2. 多重管道命令
```bash
ls -la | grep go | wc -l
```
统计当前目录中包含 "go" 的文件数量。展示了多级管道的使用。

### 3. 带环境变量的管道命令
```bash
env | grep PATH
```
显示系统路径信息，并支持添加自定义环境变量。展示了环境变量的处理。

### 4. 复杂的管道命令
```bash
ps aux | grep go | sort -k2 | head -n 3
```
查找并排序 Go 相关进程，显示前三个。展示了复杂命令链的处理能力。

### 5. 文件处理管道命令
```bash
cat ../go.mod | grep require | sort | uniq -c
```
分析 go.mod 文件中的依赖项。展示了文件内容处理能力。

## 使用方法

1. 直接运行示例：
```bash
go run pipeline.go
```

2. 在代码中使用：
```go
// 创建执行器
localExec := executor.NewLocalExecutor()
pipeExec := executor.NewPipelineExecutor(localExec)

// 执行管道命令
err := runPipeline(pipeExec, "ls -la | grep go", nil)
if err != nil {
    log.Printf("Pipeline failed: %v\n", err)
}
```

3. 带环境变量的执行：
```go
env := map[string]string{
    "CUSTOM_PATH": "/custom/path",
}
err := runPipeline(pipeExec, "env | grep PATH", env)
```

## 注意事项

1. 命令执行
   - 确保所需命令在系统中可用
   - 注意命令的执行权限
   - 考虑命令的执行时间

2. 错误处理
   - 检查每个命令的返回值
   - 处理管道中断的情况
   - 注意错误信息的传递

3. 资源管理
   - 及时关闭不需要的管道
   - 注意内存使用（特别是大文件处理）
   - 避免管道死锁

## 扩展建议

1. 功能增强
   - 添加超时控制
   - 实现并行管道
   - 支持命令重试
   - 添加管道缓冲控制

2. 使用场景
   - 日志处理和分析
   - 文件批处理
   - 系统监控
   - 数据转换

3. 最佳实践
   - 使用适当的缓冲区大小
   - 实现优雅的错误处理
   - 添加详细的日志记录
   - 考虑资源限制

## 相关文档

- [RunShell 文档](../../README.md)
- [执行器文档](../../pkg/executor/doc.go)
- [类型定义](../../pkg/types/doc.go) 