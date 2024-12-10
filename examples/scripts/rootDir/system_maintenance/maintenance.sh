#!/bin/bash

# 系统维护脚本：执行基本的系统检查任务

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查磁盘使用情况
check_disk_usage() {
    log_info "检查磁盘使用情况..."
    df -h | awk '{
        if (NR == 1) {
            print $0
        } else {
            use=$5
            gsub(/%/, "", use)
            if (use > 80) {
                printf "'${YELLOW}'%s'${NC}'\n", $0
            } else {
                print $0
            }
        }
    }'
}

# 检查内存使用情况
check_memory_usage() {
    log_info "检查内存使用情况..."
    free -h
}

# 列出正在运行的进程
list_processes() {
    log_info "列出正在运行的进程..."
    ps aux | head -n 11
}

# 主函数
main() {
    local CHECK_DISK=false
    local CHECK_MEMORY=false
    local LIST_PROCESSES=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --check-disk)
                CHECK_DISK=true
                shift
                ;;
            --check-memory)
                CHECK_MEMORY=true
                shift
                ;;
            --list-processes)
                LIST_PROCESSES=true
                shift
                ;;
            *)
                log_error "未知参数: $1"
                exit 1
                ;;
        esac
    done
    
    # 执行系统检查
    if [ "$CHECK_DISK" = true ]; then
        check_disk_usage
    fi
    
    if [ "$CHECK_MEMORY" = true ]; then
        check_memory_usage
    fi
    
    if [ "$LIST_PROCESSES" = true ]; then
        list_processes
    fi
    
    # 如果没有指定任何参数，显示所有信息
    if [ "$CHECK_DISK" = false ] && [ "$CHECK_MEMORY" = false ] && [ "$LIST_PROCESSES" = false ]; then
        check_disk_usage
        echo
        check_memory_usage
        echo
        list_processes
    fi
    
    log_info "系统检查完成"
}

# 执行主函数
main "$@" 