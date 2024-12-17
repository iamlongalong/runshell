#!/bin/bash

# 设置错误时退出
set -e

echo "Running all examples..."

# 运行 audited 示例
echo -e "\n=== Running audited example ==="
go run examples/audited/main.go

# 运行 pipeline 示例
echo -e "\n=== Running pipeline example ==="
go run examples/pipeline/main.go

# 运行 pipeline_docker 示例
echo -e "\n=== Running pipeline_docker example ==="
go run examples/pipeline_docker/main.go

# 运行 session_dev 示例
echo -e "\n=== Running session_dev example ==="
go run examples/session_dev/main.go

echo -e "\nAll examples completed successfully!" 