#!/bin/bash

# 设置错误时退出
set -e

# 显示执行的命令
set -x

# 获取版本信息
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse --short HEAD)

# 生成 Swagger 文档
which swag >/dev/null || go install github.com/swaggo/swag/cmd/swag@latest
swag init -g pkg/server/server.go -o cmd/runshell/docs --parseDependency

# 构建 Docker 镜像
docker build \
  --build-arg VERSION=${VERSION} \
  --build-arg COMMIT=${COMMIT} \
  -t runshell:latest \
  -t runshell:${VERSION} \
  -f docker/Dockerfile \
  .

# 显示构建的镜像
docker images | grep runshell 