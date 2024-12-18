package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/iamlongalong/runshell/pkg/executor/docker"
	"github.com/iamlongalong/runshell/pkg/server"
	"github.com/iamlongalong/runshell/pkg/types"
)

const (
	serverAddr = "http://localhost:8081"
	projectDir = "/tmp/runshell-projects/hello-world"
)

func createSession() (string, error) {
	log.Printf("Creating session...")

	// 创建项目目录
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create project directory: %v", err)
	}

	// 准备会话配置
	sessionConfig := struct {
		ExecutorType string                `json:"executor_type"`
		DockerConfig *types.DockerConfig   `json:"docker_config,omitempty"`
		LocalConfig  *types.LocalConfig    `json:"local_config,omitempty"`
		Options      *types.ExecuteOptions `json:"options,omitempty"`
		Metadata     map[string]string     `json:"metadata,omitempty"`
	}{
		ExecutorType: "docker",
		DockerConfig: &types.DockerConfig{
			Image:                     "golang:1.20",
			WorkDir:                   "/workspace",
			User:                      "",
			BindMount:                 fmt.Sprintf("%s:/workspace", projectDir),
			AllowUnregisteredCommands: true,
		},
		Options: &types.ExecuteOptions{
			WorkDir: "/workspace",
			Env: map[string]string{
				"GOPROXY": "https://goproxy.cn,direct",
			},
		},
	}

	log.Printf("Creating session with options: %+v", sessionConfig)

	// 序列化配置
	body, err := json.Marshal(sessionConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session config: %v", err)
	}

	// 发送创建会话请求
	resp, err := http.Post(serverAddr+"/api/v1/sessions", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to send session creation request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("session creation failed with status: %s", resp.Status)
	}

	// 解析响应
	var result struct {
		Session struct {
			ID string `json:"id"`
		} `json:"session"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return result.Session.ID, nil
}

func execCommand(sessionID string, command string, args []string) error {
	// 准备命令请求
	execRequest := struct {
		Command string                `json:"command"`
		Args    []string              `json:"args"`
		Options *types.ExecuteOptions `json:"options,omitempty"`
	}{
		Command: command,
		Args:    args,
		Options: &types.ExecuteOptions{
			WorkDir: "/workspace",
		},
	}

	// 序列化请求
	body, err := json.Marshal(execRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal exec request: %v", err)
	}

	// 发送执行命令请求
	resp, err := http.Post(
		fmt.Sprintf("%s/api/v1/sessions/%s/exec", serverAddr, sessionID),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return fmt.Errorf("failed to send exec request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("command execution failed with status: %s", resp.Status)
	}

	// 读取并打印输出
	output, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	fmt.Printf("Command output:\n%s\n", string(output))
	return nil
}

func deleteSession(sessionID string) error {
	// 创建删除请求
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/api/v1/sessions/%s", serverAddr, sessionID),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %v", err)
	}

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send delete request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("session deletion failed with status: %s", resp.Status)
	}

	return nil
}

func main() {
	log.Printf("Project directory: %s", projectDir)

	// 创建 Docker 执行器构建器
	execBuilder := docker.NewDockerExecutorBuilder(types.DockerConfig{
		Image:                     "golang:1.20",
		WorkDir:                   "/workspace",
		BindMount:                 fmt.Sprintf("%s:/workspace", projectDir),
		AllowUnregisteredCommands: true,
	}).WithOptions(&types.ExecuteOptions{
		WorkDir: "/workspace",
		Env: map[string]string{
			"GOPROXY": "https://goproxy.cn,direct",
		},
	})

	// 启动服务器
	srv := server.NewServer(execBuilder, ":8081")

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 创建会话
	sessionID, err := createSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	// 执行命令
	commands := []struct {
		cmd  string
		args []string
	}{
		{"go", []string{"version"}},
		{"pwd", nil},
		{"ls", []string{"-la"}},
	}

	for _, cmd := range commands {
		if err := execCommand(sessionID, cmd.cmd, cmd.args); err != nil {
			log.Printf("Failed to execute command %s: %v", cmd.cmd, err)
		}
	}

	// 删除会话
	if err := deleteSession(sessionID); err != nil {
		log.Printf("Failed to delete session: %v", err)
	}
}
