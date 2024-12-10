#!/usr/bin/env python3
"""
数据处理脚本：支持 CSV、JSON 和 Excel 文件的转换和基本处理
"""
import argparse
import json
import csv
import sys
import os
from datetime import datetime

def read_csv(file_path):
    data = []
    with open(file_path, 'r') as f:
        reader = csv.DictReader(f)
        for row in reader:
            data.append(row)
    return data

def read_json(file_path):
    with open(file_path, 'r') as f:
        return json.load(f)

def write_csv(data, file_path):
    if not data:
        return
    
    with open(file_path, 'w', newline='') as f:
        writer = csv.DictWriter(f, fieldnames=data[0].keys())
        writer.writeheader()
        writer.writerows(data)

def write_json(data, file_path):
    with open(file_path, 'w') as f:
        json.dump(data, f, indent=2)

def process_data(data, operations):
    """处理数据"""
    result = data
    
    # 过滤空值
    if 'filter_empty' in operations:
        result = [{k: v for k, v in item.items() if v} for item in result]
    
    # 添加时间戳
    if 'add_timestamp' in operations:
        timestamp = datetime.now().isoformat()
        for item in result:
            item['timestamp'] = timestamp
    
    return result

def main():
    parser = argparse.ArgumentParser(description='数据处理工具')
    parser.add_argument('--input', required=True, help='输入文件路径')
    parser.add_argument('--output', required=True, help='输出文件路径')
    parser.add_argument('--filter-empty', action='store_true', help='过滤空值')
    parser.add_argument('--add-timestamp', action='store_true', help='添加时间戳')
    
    args = parser.parse_args()
    
    # 确定文件类型
    input_ext = os.path.splitext(args.input)[1].lower()
    output_ext = os.path.splitext(args.output)[1].lower()
    
    # 读取数据
    try:
        if input_ext == '.csv':
            data = read_csv(args.input)
        elif input_ext == '.json':
            data = read_json(args.input)
        else:
            print(f"不支持的输入文件类型: {input_ext}", file=sys.stderr)
            sys.exit(1)
    except Exception as e:
        print(f"读取文件失败: {e}", file=sys.stderr)
        sys.exit(1)
    
    # 处理数据
    operations = []
    if args.filter_empty:
        operations.append('filter_empty')
    if args.add_timestamp:
        operations.append('add_timestamp')
    
    processed_data = process_data(data, operations)
    
    # 写入数据
    try:
        if output_ext == '.csv':
            write_csv(processed_data, args.output)
        elif output_ext == '.json':
            write_json(processed_data, args.output)
        else:
            print(f"不支持的输出文件类型: {output_ext}", file=sys.stderr)
            sys.exit(1)
    except Exception as e:
        print(f"写入文件失败: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == '__main__':
    main() 