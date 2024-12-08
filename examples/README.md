# RunShell Pipeline Examples

这个目录包含了 RunShell 的使用示例。

## Pipeline 示例

`pipeline.go` 展示了如何使用 RunShell 的管道功能。这个示例包含了以下几种管道使用场景：

1. 简单管道：`echo "hello world" | grep "world"`
2. 多重管道：`ls -l | grep go | wc -l`
3. 带环境变量的管道：`env | grep PATH`
4. 复杂管道：`ps aux | grep go | sort -k2 | head -n 3`
5. 文件处理管道：`cat go.mod | grep require | sort | uniq -c`

### 运行示例

```bash
# 编译并运行示例
go run pipeline.go

# 或者先编译后运行
go build pipeline.go
./pipeline
```

### 示例输出说明

1. 简单管道示例
   - 输出 "hello world" 并通过 grep 过滤出包含 "world" 的行
   - 展示了基本的管道功能

2. 多重管道示例
   - 列出当前目录文件，过滤出包含 "go" 的行，并计算行数
   - 展示了多个命令的串联

3. 环境变量管道示例
   - 显示环境变量并过滤出包含 PATH 的行
   - 展示了如何在管道中使用环境变量

4. 复杂管道示例
   - 显示进程信息，过滤包含 "go" ��行，排序后显示前三行
   - 展示了复杂的管道命令组合

5. 文件处理管道示例
   - 读取 go.mod 文件，过滤出 require 行，排序并统计重复行
   - 展示了文件处理相关的管道操作

### 注意事项

1. 确保系统中安装了相关命令（grep, sort, wc 等）
2. 示例中的某些命令可能需要根据你的系统环境进行调整
3. 管道命令的执行结果可能因系统环境不同而有所不同