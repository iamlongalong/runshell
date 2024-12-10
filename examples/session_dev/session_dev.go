package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/server"
	"github.com/iamlongalong/runshell/pkg/types"
)

const (
	serverAddr  = "http://localhost:8081"
	dockerImage = "golang:1.20"
)

// 创建会话
func createSession(projectDir string) (string, error) {
	// 准备请求体
	reqBody := types.SessionRequest{
		ExecutorType: types.ExecutorTypeDocker,
		DockerConfig: &types.DockerConfig{
			Image:   dockerImage,
			WorkDir: executor.DefaultWorkDir,
		},
		Options: &types.ExecuteOptions{
			WorkDir: executor.DefaultWorkDir,
			Env: map[string]string{
				"GOPROXY": "https://goproxy.cn,direct",
			},
			Metadata: map[string]string{
				"bind_mount": fmt.Sprintf("%s:%s", projectDir, executor.DefaultWorkDir),
			},
		},
	}

	// 编码请求
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	log.Printf("Creating session with options: %+v\n", reqBody)

	// 发送请求
	resp, err := http.Post(serverAddr+"/sessions", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Error != "" {
			return "", fmt.Errorf("session creation failed: %s", errResp.Error)
		}
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

	log.Printf("Session created with ID: %s\n", result.Session.ID)
	return result.Session.ID, nil
}

// 执行命令
func execCommand(sessionID string, command string, args ...string) error {
	// 准备请求体
	reqBody := struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
		WorkDir string   `json:"workdir"`
	}{
		Command: command,
		Args:    args,
		WorkDir: executor.DefaultWorkDir,
	}

	log.Printf("Executing command: %s %v\n", command, args)

	// 编码请求
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// 发送请求
	resp, err := http.Post(fmt.Sprintf("%s/sessions/%s", serverAddr, sessionID), "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to execute command: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error  string `json:"error"`
			Output string `json:"output"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			if errResp.Output != "" {
				fmt.Println(errResp.Output)
			}
			return fmt.Errorf("command failed: %s", errResp.Error)
		}
		return fmt.Errorf("command failed with status: %s", resp.Status)
	}

	// 解析响应
	var result struct {
		types.ExecuteResult
		Output string `json:"output"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// 检查执行结果
	if result.ExitCode != 0 {
		if result.Output != "" {
			fmt.Println(result.Output)
		}
		return fmt.Errorf("command failed with exit code %d: %v", result.ExitCode, result.Error)
	}

	// 显示输出
	if result.Output != "" {
		fmt.Println(result.Output)
	}

	log.Printf("Command completed successfully\n")
	return nil
}

// 删除会话
func deleteSession(sessionID string) error {
	log.Printf("Deleting session: %s\n", sessionID)

	// 创建删除请求
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/sessions/%s", serverAddr, sessionID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete session: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete failed with status: %s", resp.Status)
	}

	log.Printf("Session deleted successfully\n")
	return nil
}

func main() {
	// 创建执行器和服务器
	exec := executor.NewLocalExecutor(types.LocalConfig{
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{})
	srv := server.NewServer(exec, ":8081")

	// 启动服务器
	go func() {
		log.Printf("Starting server on %s...\n", serverAddr)
		if err := srv.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 创建项目目录（在用户主目录下）
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}
	projectDir := filepath.Join(homeDir, "runshell-projects", "hello-world")

	// 确保目录不存在
	if err := os.RemoveAll(projectDir); err != nil {
		log.Fatalf("Failed to clean project directory: %v", err)
	}

	// 创建新的目录
	if err := os.MkdirAll(projectDir, 0777); err != nil {
		log.Fatalf("Failed to create project directory: %v", err)
	}

	// 确保目录权限正确
	if err := os.Chmod(projectDir, 0777); err != nil {
		log.Fatalf("Failed to set directory permissions: %v", err)
	}

	log.Printf("Project directory: %s\n", projectDir)

	// 创建会话
	log.Println("Creating session...")
	sessionID, err := createSession(projectDir)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer deleteSession(sessionID)

	// 等待一下确保容器完全启动
	time.Sleep(2 * time.Second)

	// 创建一个简单的程序来打印一些信息
	log.Println("Creating info.go...")
	code := `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("=== RunShell Session Example ===")
	fmt.Println("This is a simple program that prints some information.")
	fmt.Println("It demonstrates the session functionality of RunShell.")
	fmt.Println("\nFeatures demonstrated:")
	fmt.Println("1. Docker container environment")
	fmt.Println("2. Volume mounting")
	fmt.Println("3. Command execution")
	fmt.Println("4. File operations")
	fmt.Println("\nProgram arguments:", os.Args)
	fmt.Println("\nThank you for using RunShell!")
}`

	if err := os.WriteFile(filepath.Join(projectDir, "info.go"), []byte(code), 0666); err != nil {
		log.Fatalf("Failed to create info.go: %v", err)
	}

	// 初始化 Go 模块
	log.Println("Initializing Go module...")
	if err := execCommand(sessionID, "go", "mod", "init", "example"); err != nil {
		log.Fatalf("Failed to initialize module: %v", err)
	}

	// 运行程序
	log.Println("\nRunning the program...")
	if err := execCommand(sessionID, "go", "run", "info.go", "arg1", "arg2"); err != nil {
		log.Fatalf("Failed to run program: %v", err)
	}

	log.Printf("\nExample completed successfully!\n")

	// 优雅关闭服务器
	if err := srv.Stop(); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}
}
