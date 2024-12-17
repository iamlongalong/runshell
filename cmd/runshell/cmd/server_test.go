package cmd

import (
	"testing"
)

func TestServerCommand(t *testing.T) {
	// 设置测试参数
	serverAddr = ":0"
	dockerImage = "ubuntu:latest"

	// 执行命令
	err := serverCmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestServerStartStop(t *testing.T) {
	// 设置测试参数
	serverAddr = ":0"
	dockerImage = "ubuntu:latest"

	// 执行命令
	err := serverCmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
