# RunShell Examples

本目录包含了 RunShell 的使用示例。

## 目录结构

```
examples/
├── audited/          # 审计执行器示例
├── pipeline/         # 基础管道执行示例
├── pipeline_docker/  # Docker 环境下的管道执行示例
└── session_dev/     # 开发会话示例
```

## 示例说明

### Audited
演示命令执行审计功能：
```bash
cd audited && go run main.go
```

### Pipeline
演示基本的管道命令执行功能：
```bash
cd pipeline && go run pipeline.go
```

### Pipeline Docker
演示在 Docker 容器中执行管道命令：
```bash
cd pipeline_docker && go run main.go
```

### Session Dev
演示开发会话功能：
```bash
cd session_dev && go run session_dev.go
```

## 运行要求
- Go 1.16+
- Docker（对于 Docker 相关示例）
- 适当的系统权限