// Package commands 实现了 RunShell 的内置命令。
package commands

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// PSCommand 实现了 'ps' 命令。
// 用于显示系统中运行的进程信息。
// 输出格式：PID、CPU使用率、内存使用率、进程状态、进程名称。
type PSCommand struct{}

// Execute 执行 ps 命令。
// 输出格式：
//   - PID：进程ID
//   - CPU%：CPU使用率
//   - MEM%：内存使用率
//   - STATE：进程状态
//   - NAME：进程名称
func (c *PSCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "ps",
		StartTime:   ctx.StartTime,
	}

	processes, err := process.Processes()
	if err != nil {
		result.Error = err
		return result, err
	}

	fmt.Fprintf(ctx.Options.Stdout, "%-10s %-10s %-10s %-10s %s\n", "PID", "CPU%", "MEM%", "STATE", "NAME")

	for _, p := range processes {
		pid := p.Pid
		name, _ := p.Name()
		cpuPercent, _ := p.CPUPercent()
		memPercent, _ := p.MemoryPercent()
		status, _ := p.Status()

		fmt.Fprintf(ctx.Options.Stdout, "%-10d %-10.1f %-10.1f %-10s %s\n",
			pid, cpuPercent, memPercent, status, name)
	}

	return result, nil
}

// TopCommand 实现了 'top' 命令。
// 用于实时显示系统资源使用情况和进程信息。
// 显示内容包括：系统概览（主机名、操作系统、运行时间、CPU、内存）和进程列表。
type TopCommand struct{}

// Execute 执行 top 命令。
// 输出内容：
//   - 系统概览：主机名、操作系统、运行时间、CPU数量、内存使用情况
//   - 进程列表：PID、CPU使用率、内存使用率、状态、名称
func (c *TopCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "top",
		StartTime:   ctx.StartTime,
	}

	// 获取系统信息
	hostInfo, err := host.Info()
	if err != nil {
		result.Error = err
		return result, err
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		result.Error = err
		return result, err
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		result.Error = err
		return result, err
	}

	// 打印系统概览
	fmt.Fprintf(ctx.Options.Stdout, "System Overview:\n")
	fmt.Fprintf(ctx.Options.Stdout, "Hostname: %s\n", hostInfo.Hostname)
	fmt.Fprintf(ctx.Options.Stdout, "OS: %s %s\n", hostInfo.Platform, hostInfo.PlatformVersion)
	fmt.Fprintf(ctx.Options.Stdout, "Uptime: %v\n", time.Duration(hostInfo.Uptime)*time.Second)
	fmt.Fprintf(ctx.Options.Stdout, "CPU(s): %d\n", len(cpuInfo))
	fmt.Fprintf(ctx.Options.Stdout, "Memory: %v/%v (%.1f%%)\n",
		memInfo.Used, memInfo.Total, memInfo.UsedPercent)

	// 打印进程信息
	processes, err := process.Processes()
	if err != nil {
		result.Error = err
		return result, err
	}

	fmt.Fprintf(ctx.Options.Stdout, "\n%-10s %-10s %-10s %-10s %s\n",
		"PID", "CPU%", "MEM%", "STATE", "NAME")

	for _, p := range processes {
		pid := p.Pid
		name, _ := p.Name()
		cpuPercent, _ := p.CPUPercent()
		memPercent, _ := p.MemoryPercent()
		status, _ := p.Status()

		fmt.Fprintf(ctx.Options.Stdout, "%-10d %-10.1f %-10.1f %-10s %s\n",
			pid, cpuPercent, memPercent, status, name)
	}

	return result, nil
}

// DFCommand 实现了 'df' 命令。
// 用于显示文件系统的磁盘空间使用情况。
// 显示内容包括：文件系统、总大小、已用空间、可用空间、使用率、挂载点。
type DFCommand struct{}

// Execute 执行 df 命令。
// 输出格式：
//   - Filesystem：文件系统设备
//   - Size：总大小（GB）
//   - Used：已用空间（GB）
//   - Avail：可用空间（GB）
//   - Use%：使用率
//   - Mounted on：挂载点
func (c *DFCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "df",
		StartTime:   ctx.StartTime,
	}

	partitions, err := disk.Partitions(false)
	if err != nil {
		result.Error = err
		return result, err
	}

	fmt.Fprintf(ctx.Options.Stdout, "%-20s %-10s %-10s %-10s %-5s %s\n",
		"Filesystem", "Size", "Used", "Avail", "Use%", "Mounted on")

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		total := float64(usage.Total) / 1024 / 1024 / 1024
		used := float64(usage.Used) / 1024 / 1024 / 1024
		free := float64(usage.Free) / 1024 / 1024 / 1024

		fmt.Fprintf(ctx.Options.Stdout, "%-20s %-10.1fG %-10.1fG %-10.1fG %-5.1f%% %s\n",
			partition.Device, total, used, free, usage.UsedPercent, partition.Mountpoint)
	}

	return result, nil
}

// UNameCommand 实现了 'uname' 命令。
// 用于显示系统信息。
// 支持 -a 选项显示详细信息。
type UNameCommand struct{}

// Execute 执行 uname 命令。
// 参数：
//   - 无参数：只显示操作系统名称
//   - -a：显示完整系统信息（操作系统、主机名、内核版本、平台、架构、版本）
func (c *UNameCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "uname",
		StartTime:   ctx.StartTime,
	}

	hostInfo, err := host.Info()
	if err != nil {
		result.Error = err
		return result, err
	}

	all := false
	if len(ctx.Args) > 0 && ctx.Args[0] == "-a" {
		all = true
	}

	if all {
		fmt.Fprintf(ctx.Options.Stdout, "%s %s %s %s %s %s\n",
			hostInfo.OS,
			hostInfo.Hostname,
			hostInfo.KernelVersion,
			hostInfo.Platform,
			runtime.GOARCH,
			hostInfo.PlatformVersion)
	} else {
		fmt.Fprintln(ctx.Options.Stdout, hostInfo.OS)
	}

	return result, nil
}

// EnvCommand 实现了 'env' 命令。
// 用于显示系统环境变量。
// 支持按模式过滤环境变量。
type EnvCommand struct{}

// Execute 执行 env 命令。
// 参数：
//   - 无参数：显示所有环境变量
//   - 有参数：按参数指定的前缀过滤环境变量（不区分大小写）
func (c *EnvCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "env",
		StartTime:   ctx.StartTime,
	}

	env := os.Environ()
	if len(ctx.Args) > 0 {
		pattern := strings.ToUpper(ctx.Args[0])
		for _, e := range env {
			if strings.HasPrefix(strings.ToUpper(e), pattern) {
				fmt.Fprintln(ctx.Options.Stdout, e)
			}
		}
	} else {
		for _, e := range env {
			fmt.Fprintln(ctx.Options.Stdout, e)
		}
	}

	return result, nil
}

// KillCommand 实现了 'kill' 命令。
// 用于终止指定的进程。
// 需要提供进程ID作为参数。
type KillCommand struct{}

// Execute 执行 kill 命令。
// 参数：
//   - 需要至少一个进程ID
//   - 支持同时终止多个进程
//
// 错误处理：
//   - 无效的进程ID
//   - 进程不存在
//   - 无权限终止进程
func (c *KillCommand) Execute(ctx *types.ExecuteContext) (*types.ExecuteResult, error) {
	result := &types.ExecuteResult{
		CommandName: "kill",
		StartTime:   ctx.StartTime,
		ExitCode:    1, // 默认为错误状态
	}

	if len(ctx.Args) == 0 {
		result.Error = fmt.Errorf("kill: usage: kill [-s sigspec | -n signum | -sigspec] pid | jobspec ... or kill -l [sigspec]")
		return result, result.Error
	}

	for _, arg := range ctx.Args {
		pid := 0
		_, err := fmt.Sscanf(arg, "%d", &pid)
		if err != nil {
			result.Error = fmt.Errorf("kill: invalid process id: %s", arg)
			return result, result.Error
		}

		proc, err := os.FindProcess(pid)
		if err != nil {
			result.Error = fmt.Errorf("kill: process not found: %d", pid)
			return result, result.Error
		}

		err = proc.Kill()
		if err != nil {
			result.Error = fmt.Errorf("kill: failed to kill process %d: %v", pid, err)
			return result, result.Error
		}
	}

	result.ExitCode = 0 // 成功时设置为0
	return result, nil
}
