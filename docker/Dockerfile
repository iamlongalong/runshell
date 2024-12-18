# 构建阶段
FROM golang:1.22-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git make

# 复制源代码
COPY . .

# 构建应用
RUN go build -o runshell ./cmd/runshell

# 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 创建非 root 用户
RUN adduser -D -h /app runshell

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/runshell /app/

# 设置用户
USER runshell

# 暴露端口
EXPOSE 8080

# 设置入口点
ENTRYPOINT ["/app/runshell"]

# 默认命令
CMD ["server", "--http", ":8080"] 