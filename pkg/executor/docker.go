// Package executor 实现了命令执行器的核心功能。
// 本文件实现了 Docker 容器中的命令执行器。
package executor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/iamlongalong/runshell/pkg/types"
)

// DockerExecutor 实现了在 Docker 容器中执行命令的执行器。
// 特性：
// - 支持在指定镜像中执行命令
// - 自动管理容器的生命周期
// - 支持环境变量和工作目录配置
// - 命令注册和管理
type DockerExecutor struct {
	client       *client.Client // Docker 客户端实例
	commands     sync.Map       // 注册的命令映射
	defaultImage string         // 默认使用的 Docker 镜像
}

// NewDockerExecutor 创建一个新的 Docker 执行器实例。
// 参数：
//   - defaultImage：默认使用的 Docker 镜像名称
//
// 返回值：
//   - *DockerExecutor：执行器实例
//   - error：创建过程中的错误
func NewDockerExecutor(defaultImage string) (*DockerExecutor, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	return &DockerExecutor{
		client:       cli,
		defaultImage: defaultImage,
	}, nil
}

// Execute 执行指定的命令。
// 执行过程：
// 1. 首先检查是否为注册的内置命令
// 2. 如果不是内置命令，则在 Docker 容器中执行
//
// 参数：
//   - ctx：上下文，用于取消和超时控制
//   - cmdName：要执行的命令名称
//   - args：命令参数列表
//   - opts：执行选项（工作目录、环境变量等）
//
// 返回值：
//   - *types.ExecuteResult：执行结果
//   - error：执行过程中的错误
func (e *DockerExecutor) Execute(ctx context.Context, cmdName string, args []string, opts *types.ExecuteOptions) (*types.ExecuteResult, error) {
	if opts == nil {
		opts = &types.ExecuteOptions{
			Env:     make(map[string]string),
			WorkDir: "",
		}
	}

	// 检查命令是否已注册
	if cmd, ok := e.commands.Load(cmdName); ok {
		command := cmd.(*types.Command)
		execCtx := &types.ExecuteContext{
			Context:   ctx,
			Args:      args,
			Options:   opts,
			StartTime: time.Now(),
		}
		return command.Handler.Execute(execCtx)
	}

	// 如果命令未注册，使用默认Docker执行方式
	return e.executeInDocker(ctx, cmdName, args, opts)
}

// executeInDocker 在 Docker 容器中执行命令。
// 执行流程：
// 1. 创建容器配置
// 2. 创建并启动容器
// 3. 等待命令执行完成
// 4. 收集执行结果
// 5. 清理容器
//
// 参数：
//   - ctx：上下文
//   - cmdName：命令名称
//   - args：命令参数
//   - opts：执行选项
func (e *DockerExecutor) executeInDocker(ctx context.Context, cmdName string, args []string, opts *types.ExecuteOptions) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: cmdName,
		StartTime:   time.Now(),
	}

	// 准备容器配置
	// 首先检查命令是否存在，不存在则返回错误
	shellCmd := fmt.Sprintf("command -v %s >/dev/null 2>&1 || { echo '%s: command not found' >&2; exit 127; }; %s %s", cmdName, cmdName, cmdName, shellEscapeArgs(args))
	config := &container.Config{
		Image:      e.defaultImage,
		Cmd:        []string{"/bin/sh", "-c", shellCmd},
		Tty:        false,
		OpenStdin:  false,
		StdinOnce:  true,
		Env:        makeEnvSlice(opts.Env),
		WorkingDir: opts.WorkDir,
	}

	// 创建容器
	resp, err := e.client.ContainerCreate(ctx, config, &container.HostConfig{}, nil, nil, "")
	if err != nil {
		result.Error = err
		result.EndTime = time.Now()
		return result, err
	}

	// 确保容器被清理
	defer e.client.ContainerRemove(context.Background(), resp.ID, dockertypes.ContainerRemoveOptions{Force: true})

	// 启动容器
	if err := e.client.ContainerStart(ctx, resp.ID, dockertypes.ContainerStartOptions{}); err != nil {
		result.Error = err
		result.EndTime = time.Now()
		return result, err
	}

	// 等待容器结束
	statusCh, errCh := e.client.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		result.Error = err
		result.EndTime = time.Now()
		return result, err
	case status := <-statusCh:
		result.ExitCode = int(status.StatusCode)
		if result.ExitCode == 127 {
			result.Error = types.ErrCommandNotFound
		}
	}

	// 获取容器日志
	out, err := e.client.ContainerLogs(ctx, resp.ID, dockertypes.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
	})
	if err != nil {
		result.Error = err
		result.EndTime = time.Now()
		return result, err
	}
	defer out.Close()

	// 创建输出缓冲区
	var stdout, stderr strings.Builder

	// 复制输出
	_, err = stdcopy.StdCopy(&stdout, &stderr, out)
	if err != nil {
		result.Error = err
		result.EndTime = time.Now()
		return result, err
	}

	// 设置输出
	result.Output = stdout.String()
	if stderr.String() != "" {
		if result.Error == nil {
			result.Error = fmt.Errorf(stderr.String())
		}
	}

	result.EndTime = time.Now()
	return result, result.Error
}

// GetCommandInfo 获取命令信息。
// 参数：
//   - cmdName：命令名称
//
// 返回值：
//   - *types.Command：命令信息
//   - error：如果命令不存在则返回 ErrCommandNotFound
func (e *DockerExecutor) GetCommandInfo(cmdName string) (*types.Command, error) {
	if cmd, ok := e.commands.Load(cmdName); ok {
		return cmd.(*types.Command), nil
	}
	return nil, types.ErrCommandNotFound
}

// GetCommandHelp 获取命令的帮助信息。
// 参数：
//   - cmdName：命令名称
//
// 返回值：
//   - string：命令的使用说明
//   - error：如果命令不存在则返回错误
func (e *DockerExecutor) GetCommandHelp(cmdName string) (string, error) {
	cmd, err := e.GetCommandInfo(cmdName)
	if err != nil {
		return "", err
	}
	return cmd.Usage, nil
}

// ListCommands 列出符合过滤条件的命令。
// 参数：
//   - filter：命令过滤器，可按类别过滤
//
// 返回值：
//   - []*types.Command：符合条件的命令列表
//   - error：列出过程中的错误
func (e *DockerExecutor) ListCommands(filter *types.CommandFilter) ([]*types.Command, error) {
	var commands []*types.Command
	e.commands.Range(func(key, value interface{}) bool {
		cmd := value.(*types.Command)
		if filter == nil || (filter.Category == "" || filter.Category == cmd.Category) {
			commands = append(commands, cmd)
		}
		return true
	})
	return commands, nil
}

// RegisterCommand 注册新命令。
// 参数：
//   - cmd：要注册的命令
//
// 返回值：
//   - error：如果命令无效则返回错误
func (e *DockerExecutor) RegisterCommand(cmd *types.Command) error {
	if cmd == nil || cmd.Name == "" {
		return fmt.Errorf("invalid command")
	}
	e.commands.Store(cmd.Name, cmd)
	return nil
}

// UnregisterCommand 注销已注册的命令。
// 参数：
//   - cmdName：要注销的命令名称
//
// 返回值：
//   - error：注销过程中的错误
func (e *DockerExecutor) UnregisterCommand(cmdName string) error {
	e.commands.Delete(cmdName)
	return nil
}

// makeEnvSlice 将环境变量映射转换为字符串切片。
// 参数：
//   - env：环境变量映射
//
// 返回值：
//   - []string：格式化后的环境变量列表
func makeEnvSlice(env map[string]string) []string {
	if env == nil {
		return nil
	}
	envSlice := make([]string, 0, len(env))
	for k, v := range env {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}
	return envSlice
}

// shellEscapeArgs 对命令参数进行 shell 转义。
// 参数：
//   - args：要转义的参数列表
//
// 返回值：
//   - string：转义后的参数字符串
func shellEscapeArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	var escaped []string
	for _, arg := range args {
		escaped = append(escaped, fmt.Sprintf("'%s'", strings.ReplaceAll(arg, "'", "'\\''")))
	}
	return strings.Join(escaped, " ")
}
