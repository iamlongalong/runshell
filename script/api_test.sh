#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志文件
LOG_FILE="log/api_test.log"
SERVER_LOG="log/server.log"

# 服务器进程 ID
SERVER_PID=""

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0

# 清理函数
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    if [ ! -z "$SERVER_PID" ]; then
        echo "Stopping server (PID: $SERVER_PID)"
        kill $SERVER_PID 2>/dev/null
        wait $SERVER_PID 2>/dev/null
    fi
}

# 设置退出钩子
trap cleanup EXIT

# 日志函数
log() {
    echo -e "$1" | tee -a $LOG_FILE
}

# 测试函数
run_test() {
    local name=$1
    local command=$2
    local expected_status=$3
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "\n${YELLOW}Running test: ${name}${NC}"
    echo "Command: $command"
    
    # 运行命令并捕获输出和状态码
    local output
    local status
    output=$(eval $command 2>&1)
    status=$?
    
    # 检查状态码
    if [ $status -eq $expected_status ]; then
        echo -e "${GREEN}✓ Test passed${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗ Test failed${NC}"
        echo "Expected status: $expected_status"
        echo "Got status: $status"
    fi
    
    # 记录输出
    echo "Output:"
    echo "$output"
    
    # 记录到日志文件
    {
        echo "=== Test: $name ==="
        echo "Command: $command"
        echo "Expected status: $expected_status"
        echo "Got status: $status"
        echo "Output:"
        echo "$output"
        echo "==================="
    } >> $LOG_FILE
}

# 等待服务器启动
wait_for_server() {
    local max_attempts=30
    local attempt=1
    
    echo -n "Waiting for server to start"
    while [ $attempt -le $max_attempts ]; do
        if curl -s http://localhost:7070/api/v1/health >/dev/null; then
            echo -e "\n${GREEN}Server is up!${NC}"
            return 0
        fi
        echo -n "."
        sleep 1
        attempt=$((attempt + 1))
    done
    
    echo -e "\n${RED}Server failed to start within $max_attempts seconds${NC}"
    return 1
}

# 清理旧的日志文件
rm -f $LOG_FILE $SERVER_LOG

# 启动服务器
echo -e "${YELLOW}Starting server...${NC}"
go run ./cmd/runshell/main.go server --addr :7070 --executor-type local > $SERVER_LOG 2>&1 &
SERVER_PID=$!

# 等待服务器启动
if ! wait_for_server; then
    echo -e "${RED}Failed to start server${NC}"
    exit 1
fi

# 运行测试用例
echo -e "\n${YELLOW}Running API tests...${NC}"

# 1. 测试 Swagger 文档
run_test "Swagger UI HTML" \
    "curl -s -w '%{http_code}' http://localhost:7070/swagger/index.html" \
    0

run_test "Swagger JSON Doc" \
    "curl -s -w '%{http_code}' http://localhost:7070/swagger/doc.json -o swagger_test.json && [ -s swagger_test.json ] && cat swagger_test.json | grep -q 'RunShell API'" \
    0

# 验证 Swagger JSON 文档的内容
echo -e "\n${YELLOW}Verifying Swagger JSON content...${NC}"
if [ -f swagger_test.json ]; then
    # 检查必要的 API 端点是否存在
    endpoints=("/health" "/exec" "/commands" "/help" "/sessions")
    missing_endpoints=()
    
    for endpoint in "${endpoints[@]}"; do
        if ! grep -q "\"$endpoint\"" swagger_test.json; then
            missing_endpoints+=("$endpoint")
        fi
    done
    
    if [ ${#missing_endpoints[@]} -eq 0 ]; then
        echo -e "${GREEN}✓ All required endpoints found in Swagger documentation${NC}"
    else
        echo -e "${RED}✗ Missing endpoints in Swagger documentation: ${missing_endpoints[*]}${NC}"
    fi
    
    # 清理测试文件
    rm swagger_test.json
else
    echo -e "${RED}✗ Failed to download Swagger documentation${NC}"
fi

# 2. 健康检查接口
run_test "Health Check" \
    "curl -s -w '%{http_code}' http://localhost:7070/api/v1/health" \
    0

# 3. 执行命令接口
run_test "Execute Command (ls)" \
    "curl -s -w '%{http_code}' -X POST http://localhost:7070/api/v1/exec -H 'Content-Type: application/json' -d '{\"command\": \"ls\", \"args\": [\"-l\"]}'" \
    0

# 4. 列出命令接口
run_test "List Commands" \
    "curl -s -w '%{http_code}' http://localhost:7070/api/v1/commands" \
    0

# 5. 命令帮助接口
run_test "Command Help" \
    "curl -s -w '%{http_code}' 'http://localhost:7070/api/v1/help?command=ls'" \
    0

# 6. 创建会话
run_test "Create Session" \
    "curl -s -w '%{http_code}' -X POST http://localhost:7070/api/v1/sessions -H 'Content-Type: application/json' -d '{\"executor_type\":\"local\"}'" \
    0

# 7. 无效命令测试
run_test "Invalid Command" \
    "curl -s -w '%{http_code}' -X POST http://localhost:7070/api/v1/exec -H 'Content-Type: application/json' -d '{\"command\": \"invalid_command\"}'" \
    0

# 8. 管道命令测试
run_test "Pipeline Command" \
    "curl -s -w '%{http_code}' -X POST http://localhost:7070/api/v1/exec -H 'Content-Type: application/json' -d '{\"command\": \"sh\", \"args\": [\"-c\", \"ls -l | grep go\"]}'" \
    0

# 打印测试结果摘要
echo -e "\n${YELLOW}Test Summary:${NC}"
echo "Total tests: $TOTAL_TESTS"
echo "Passed tests: $PASSED_TESTS"
echo "Failed tests: $((TOTAL_TESTS - PASSED_TESTS))"

# 检查服务器日志中的错误
echo -e "\n${YELLOW}Checking server logs for errors...${NC}"
if grep -i "error" $SERVER_LOG; then
    echo -e "${RED}Found errors in server log${NC}"
else
    echo -e "${GREEN}No errors found in server log${NC}"
fi

# 退出码
if [ $PASSED_TESTS -eq $TOTAL_TESTS ]; then
    echo -e "\n${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}Some tests failed!${NC}"
    exit 1
fi 