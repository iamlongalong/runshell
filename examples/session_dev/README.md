# Session Development Example

这个示例展示了如何使用 RunShell 的会话功能来在 Docker 容器中开发一个简单的 Go 项目。

## 功能特点

1. 使用 Docker 容器作为开发环境
   - 基于 golang:1.20 镜像
   - 主机目录绑定，代码持久化
   - 开发环境隔离

2. 自动管理会话生命周期
   - 自动创建和清理会话
   - 环境变量配置
   - 工作目录管理

3. 完整开发工作流程：
   - 初始化 Go 模块
   - 创建源代码文件
   - 格式化代码
   - 运行测试
   - 构建项目

## 项目结构

```
~/runshell-projects/hello-world/  # 项目根目录
├── go.mod                        # Go 模块文件
├── main.go                       # 主程序
└── app                          # 构建产物
```

## 使用方法

1. 确保已安装 Docker 并启动

2. 启动 RunShell 服务器：
   ```bash
   go run ../../cmd/runshell/main.go server --addr :8080
   ```

3. 在另一个终端运行示例：
   ```bash
   go run session_dev.go
   ```

## 工作流程说明

1. 创建会话
   - 使用 golang:1.20 镜像
   - 将主机目录 ~/runshell-projects/hello-world 绑定到容器的 /workspace
   - 设置 GOPROXY 等环境变量

2. 初始化项目
   - 在主机目录创建项目文件夹
   - 初始化 Go 模块
   - 创建 main.go 文件

3. 开发流程
   - 代码格式化
   - 运行测试
   - 构建项目

4. 继续开发
   - 项目文件保存在主机目录
   - 可以使用本地 IDE 编辑代码
   - 在容器中执行构建和测试

## 注意事项

1. 确保：
   - Docker daemon 正在运行
   - 8080 端口未被占用
   - 有主机目录的写入权限

2. 项目文件：
   - 保存在 ~/runshell-projects/hello-world
   - 不会随容器删除而丢失
   - 可以用任何编辑器修改

3. 容器环境：
   - 提供隔离的构建环境
   - 避免污染主机环境
   - 确保构建环境一致性

## 扩展建议

1. 开发工具增强：
   - 添加热重载功能
   - 集成调试工具
   - 添加依赖管理

2. 环境定制：
   - 自定义 Docker 镜像
   - 配置开发工具
   - 添加常用命令别名

3. 工作流优化：
   - 自动化测试
   - 持续集成
   - 开发环境共享 