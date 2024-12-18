package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/iamlongalong/runshell/pkg/commands/script"
	"github.com/iamlongalong/runshell/pkg/executor/docker"
	"github.com/iamlongalong/runshell/pkg/types"
)

const (
	defaultPythonImage = "python:latest"
	defaultShellImage  = "ubuntu:latest"
)

var scriptDir string

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// 检查命令行参数
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: %s <script-dir> <script-name> [args...]", os.Args[0])
	}

	// 初始化执行环境
	var err error
	scriptDir, err = filepath.Abs(os.Args[1])
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if err := validateScriptDir(scriptDir); err != nil {
		return err
	}

	// 创建执行器
	executors, err := createExecutors()
	if err != nil {
		return fmt.Errorf("failed to create executors: %w", err)
	}

	// 创建脚本管理器
	scriptMgr, err := createScriptManager(scriptDir, executors[script.PythonScript])
	if err != nil {
		return fmt.Errorf("failed to create script manager: %w", err)
	}
	defer scriptMgr.Close()

	// 显示可用脚本
	printAvailableScripts(scriptMgr)

	// 执行脚本
	return executeScript(scriptMgr, os.Args[2], os.Args[3:])
}

func validateScriptDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("invalid script directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}
	return nil
}

func createExecutors() (map[script.ScriptType]types.Executor, error) {
	executors := make(map[script.ScriptType]types.Executor)

	// 创建 Python 执行器
	pythonExec, err := docker.NewDockerExecutor(types.DockerConfig{
		Image:                     defaultPythonImage,
		AllowUnregisteredCommands: true,
		BindMount:                 scriptDir + ":/app",
		WorkDir:                   "/app",
	}, &types.ExecuteOptions{}, nil)
	if err != nil {
		return nil, err
	}
	executors[script.PythonScript] = pythonExec

	return executors, nil
}

func createScriptManager(rootDir string, defaultExecutor types.Executor) (*script.ScriptManager, error) {
	return script.NewScriptManager(&script.Config{
		RootDir:  rootDir,
		Executor: defaultExecutor,
	}, &types.ExecuteOptions{
		Env: map[string]string{
			"PYTHONPATH":              "/app",
			"PYTHONUNBUFFERED":        "1",
			"PYTHONDONTWRITEBYTECODE": "1",
		},
	})
}

func printAvailableScripts(scriptMgr *script.ScriptManager) {
	fmt.Println("\nAvailable scripts:")
	for _, cmd := range scriptMgr.ListCommands() {
		fmt.Printf("- %s: %s\n", cmd.Name, cmd.Description)
		fmt.Printf("  Usage: %s\n", cmd.Usage)
	}
	fmt.Println()
}

func executeScript(scriptMgr *script.ScriptManager, scriptName string, args []string) error {
	// 创建执行上下文
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Command: types.Command{
			Command: scriptName,
			Args:    args,
		},
		Options: &types.ExecuteOptions{
			Env: make(map[string]string),
		},
	}

	// 执行脚本
	result, err := scriptMgr.Execute(ctx)
	if err != nil {
		return fmt.Errorf("script execution failed: %w", err)
	}

	// 检查执行结果
	if result.ExitCode != 0 {
		return fmt.Errorf("script exited with code %d", result.ExitCode)
	}

	fmt.Println("Script executed successfully")
	return nil
}
