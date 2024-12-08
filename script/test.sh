#!/bin/bash

# 设置错误时退出
set -e

# 显示执行的命令
set -x

# 运行所有测试，包括集成测试
go test -v -race -coverprofile=coverage.out ./...

# 显示测试覆盖率报告
go tool cover -func=coverage.out

# 如果需要在浏览器中查看详细覆盖率报告，取消下面的注释
# go tool cover -html=coverage.out 