# RunShell

RunShell 是一个强大的命令执行框架，支持本地和 Docker 容器中执行命令。它提供了丰富的内置命令、审计日志、HTTP API 等功能。

## 功能特性

- **多种执行模式**
  - 本地命令执行
  - Docker 容器中执行
  - 交互式 Shell
  - HTTP API 服务

- **命令管理**
  - 内置常用命令
  - 自定义命令注册
  - 命令分类管理
  - 命令帮助系统

- **安全特性**
  - 命令执行审计
  - 用户权限控制
  - 资源使用统计
  - 超时控制

- **其他特性**
  - 环境变量管理
  - 工作目录设置
  - 输入输出流控制
  - 错误处理机制

## 项目结构

```
.
├── cmd/                    # 命令行工具
│   └── runshell/          # 主程序
├── pkg/                    # 核心包
│   ├── audit/             # 审计日志
│   ├── commands/          # 内置命令
│   ├── executor/          # 执行器
│   ├── script/            # 脚本管理
│   ├── server/            # HTTP 服务器
│   └── types/             # 公共类型
├── script/                # 构建和测试脚本
├── docker/                # Docker 相关文件
└── docs/                  # 文档
```

## 快速开始

### 安装

#### 从源码安装

```bash
# 克隆仓库
git clone https://github.com/iamlongalong/runshell.git
cd runshell

# 安装依赖
make deps

# 构建项目
make build-local
```

#### 使用 Docker

```bash
# 构建并运行 Docker 容器
make docker-build docker-run
```

### 使用说明

#### 命令行使用

```bash
# 执行简单命令
runshell exec ls -l

# 设置工作目录
runshell exec --workdir /tmp ls -l

# 设置环境变量
runshell exec --env KEY=VALUE env

# 启动 HTTP 服务器
runshell server --http :8080

# 启动交互式 Shell
runshell shell
```

## 开发指南

### Make 命令

项目提供了一系列 Make 命令来简化开发和部署流程：

#### 基础操作
- `make` - 执行默认操作（清理、测试、构建）
- `make clean` - 清理构建产物
- `make deps` - 更新依赖
- `make help` - 显示所有可用命令

#### 测试相关
- `make test` - 运行所有测试
- `make test-unit` - 只运行单元测试（跳过集成测试）
- `make coverage` - 生成代码覆盖率报告

#### 构建相关
- `make build` - 构建所有平台版本
- `make build-local` - 只构建当前平台版本

#### Docker 相关
- `make docker-build` - 构建 Docker 镜像
- `make docker-run` - 运行 Docker 容器
- `make docker-stop` - 停止 Docker 容器

#### 开发工具
- `make fmt` - 格式化代码
- `make lint` - 代码检查
- `make run` - 运行本地服务器
- `make tag` - 创建新的 Git 标签

### 开发流程

1. **准备开发环境**
   ```bash
   # 更新依赖
   make deps
   
   # 格式化代码
   make fmt
   ```

2. **运行测试**
   ```bash
   # 运行单元测试
   make test-unit
   
   # 运行所有测试
   make test
   ```

3. **本地调试**
   ```bash
   # 构建并运行服务器
   make run
   ```

4. **发布新版本**
   ```bash
   # 代码检查
   make lint
   
   # 运行测试
   make test
   
   # 创建新标签
   make tag
   ```

## API 文档

### RESTful API

- `GET /health` - 健康检查
- `POST /exec` - 执行命令
- `GET /commands` - 列出可用命令
- `GET /commands/{name}` - 获取命令信息

详细的 API 文档请参考 [API.md](docs/API.md)。

## 配置

RunShell 支持以下配置选项：

- `--audit-dir` - 审计日志目录
- `--docker-image` - 默认 Docker 镜像
- `--http` - HTTP 服务器地址

更多配置选项请参考 [CONFIG.md](docs/CONFIG.md)。

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

在提交代码前，请确保：

1. 通过所有测试 (`make test`)
2. 代码符合规范 (`make lint`)
3. 更新相关文档
4. 添加必要的测试用例

## 许可证

本项目采用 MIT 许可证，详见 [LICENSE](LICENSE) 文件。

## 作者

- iamlongalong

## 致谢

感谢以下开源项目：

- [cobra](https://github.com/spf13/cobra)
- [docker](https://github.com/docker/docker) 