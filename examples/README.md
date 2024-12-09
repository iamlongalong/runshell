# RunShell Examples

本目录包含了 RunShell 的使用示例。

## 目录结构

```
examples/
├── pipeline/       # 管道执行示例
│   └── pipeline.go # 演示如何使用管道执行命令
│
├── session_dev/    # 会话开发示例
│   ├── session_dev.go  # 使用会话进行开发的示例代码
│   └── session_dev.md  # 示例说明文档
│
└── README.md       # 本文件
```

## 示例说明

### 1. Pipeline Example

`pipeline/pipeline.go` 展示了如何使用 RunShell 的管道功能：
- 简单管道命令
- 多重管道命令
- 带环境变量的管道
- 复杂的管道处理
- 文件处理管道

运行方式：
```bash
cd pipeline
go run pipeline.go
```

### 2. Session Development Example

`session_dev/session_dev.go` 展示了如何使用 RunShell 的会话功能进行开发：
- Docker 容器环境
- 会话生命周期管理
- 完整开发工作流程
- 自动化构建和测试

运行方式：
```bash
# 终端 1：启动服务器
go run ../cmd/runshell/main.go server --addr :8080

# 终端 2：运行示例
cd session_dev
go run session_dev.go
```

## 注意事项

1. 运行示例前请确保：
   - Docker daemon 正在运行
   - 相关端口未被占用
   - 已安装所需依赖

2. 每个示例都有自己的 README 或文档，请参考具体说明。

3. 示例代码主要用于演示目的，生产环境使用时请注意安全性和错误处理。