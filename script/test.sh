#!/bin/bash

# 设置错误时退出
set -e

# 显示执行的命令
set -x

# 生成 Swagger 文档
which swag >/dev/null || go install github.com/swaggo/swag/cmd/swag@latest
swag init -g pkg/server/server.go -o cmd/runshell/docs --parseDependency

# 验证 Swagger 文档是否生成成功
if [ ! -f "cmd/runshell/docs/swagger.json" ]; then
    echo "Error: Swagger documentation not generated"
    exit 1
fi

# 运行所有测试，包括集成测试
go test -v -race -coverprofile=log/coverage.out ./...

# 显示测试覆盖率报告
go tool cover -func=log/coverage.out

# 如果需要在浏览器中查看详细覆盖率报告，取消下面的注释
# go tool cover -html=log/coverage.out 