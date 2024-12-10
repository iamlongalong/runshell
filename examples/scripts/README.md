# Script Management System Examples

这个目录包含了脚本管理系统的示例代码。

## 目录结构

```
examples/scripts/
└── rootDir/                  # 脚本根目录
    ├── process.py           # CSV 转 JSON 示例脚本
    ├── process.meta.json    # 脚本元数据
    └── sample_data.csv      # 示例数据
```

## 可用示例

### CSV 转 JSON 工具
- **用途**：将 CSV 文件转换为 JSON 格式
- **功能**：
  - CSV 到 JSON 的转换
  - 空值过滤
- **使用方法**：参见 process.meta.json 中的配置

## 运行脚本

使用脚本管理系统执行脚本：

```bash
# 基本用法
runshell exec rootDir data_processor --input sample_data.csv --output result.json

# 带空值过滤
runshell exec rootDir data_processor --input sample_data.csv --output result.json --filter-empty
```

## 添加新脚本

添加新脚本需要：
1. 在 rootDir 下创建脚本文件
2. 创建对应的 .meta.json 元数据文件
3. 在元数据文件中定义脚本的类型、参数等信息