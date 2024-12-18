#!/bin/bash

# 设置错误时退出
set -e

# 显示执行的命令
set -x

# 版本信息
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse --short HEAD)
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

# 生成 Swagger 文档
which swag >/dev/null || go install github.com/swaggo/swag/cmd/swag@latest
swag init -g pkg/server/server.go -o cmd/runshell/docs --parseDependency

# 编译参数
LDFLAGS="-X main.Version=${VERSION} -X main.GitCommit=${COMMIT} -X main.BuildTime=${BUILD_TIME}"

# 创建输出目录
mkdir -p bin

# 编译 Linux 版本
GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o bin/runshell-linux-amd64 ./cmd/runshell

# 编译 macOS 版本
GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o bin/runshell-darwin-amd64 ./cmd/runshell
GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o bin/runshell-darwin-arm64 ./cmd/runshell

# 编译 Windows 版本
GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o bin/runshell-windows-amd64.exe ./cmd/runshell

# 显示编译结果
ls -lh bin/ 