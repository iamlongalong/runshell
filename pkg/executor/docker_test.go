package executor

import (
	"bytes"
	"context"
	"os/exec"
	"testing"

	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestDockerExecutor_Execute(t *testing.T) {
	// 跳过如果没有 docker 命令
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not installed, skipping tests")
	}

	tests := []struct {
		name        string
		dockerImage string
		command     string
		args        []string
		env         map[string]string
		workDir     string
		wantErr     bool
		wantOutput  string
		wantCode    int
	}{
		{
			name:        "Simple command",
			dockerImage: "ubuntu:latest",
			command:     "ls",
			args:        []string{"-l"},
			wantErr:     false,
			wantOutput:  "total",
			wantCode:    0,
		},
		{
			name:        "Command with spaces",
			dockerImage: "busybox:latest",
			command:     "ls",
			args:        []string{"-la", "/etc"},
			wantErr:     false,
			wantOutput:  "total",
			wantCode:    0,
		},
		{
			name:        "Pipeline command",
			dockerImage: "busybox:latest",
			command:     "ls -la | grep total",
			wantErr:     false,
			wantOutput:  "total",
			wantCode:    0,
		},
		{
			name:        "Pipeline with grep no matches",
			dockerImage: "busybox:latest",
			command:     "ls -la | grep nonexistentfile",
			wantErr:     true,
			wantOutput:  "",
			wantCode:    1,
		},
		{
			name:        "Complex pipeline",
			dockerImage: "busybox:latest",
			command:     "ls -la /etc | grep conf | sort",
			wantErr:     false,
			wantOutput:  "conf",
			wantCode:    0,
		},
		{
			name:        "Invalid command",
			dockerImage: "busybox:latest",
			command:     "nonexistentcommand",
			wantErr:     true,
			wantOutput:  "",
			wantCode:    127,
		},
		{
			name:        "Invalid image",
			dockerImage: "nonexistentimage:latest",
			command:     "ls",
			wantErr:     true,
			wantOutput:  "",
			wantCode:    125, // Docker pull error code
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建执行器
			exec, err := NewDockerExecutor(types.DockerConfig{
				Image:                     tt.dockerImage,
				AllowUnregisteredCommands: true,
			}, &types.ExecuteOptions{})
			if err != nil {
				if tt.wantErr {
					return // 预期的错误情况
				}
				t.Fatalf("Failed to create Docker executor: %v", err)
			}
			if exec != nil {
				defer exec.Close()
			}

			// 如果创建执行器失败且期望错误，则测试通过
			if exec == nil && tt.wantErr {
				return
			}

			// 准备执行上下文
			var output bytes.Buffer
			ctx := &types.ExecuteContext{
				Context: context.Background(),
				Command: types.Command{
					Command: tt.command,
					Args:    tt.args,
				},
				Options: &types.ExecuteOptions{
					Env:     tt.env,
					WorkDir: tt.workDir,
					Stdout:  &output,
				},
			}

			// 执行命令
			result, err := exec.Execute(ctx)
			if tt.wantErr {
				if err == nil {
					assert.Equal(t, tt.wantCode, result.ExitCode)
				} else {
					assert.Error(t, err)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, result.ExitCode)
			if tt.wantOutput != "" {
				assert.Contains(t, output.String(), tt.wantOutput)
			}
		})
	}
}

func TestDockerExecutor_ExecutePipeline(t *testing.T) {
	// 跳过如果没有 docker 命令
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not installed, skipping tests")
	}

	// 创建 Docker 执行器
	dockerExec, err := NewDockerExecutor(types.DockerConfig{
		Image:                     "busybox:latest",
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{})
	if err != nil {
		t.Fatalf("Failed to create Docker executor: %v", err)
	}
	defer dockerExec.Close()

	tests := []struct {
		name      string
		commands  []types.Command
		wantErr   bool
		exitCode  int
		checkFunc func(t *testing.T, output string)
	}{
		{
			name: "Simple pipeline",
			commands: []types.Command{
				{Command: "echo", Args: []string{"total 123"}},
				{Command: "grep", Args: []string{"total"}},
			},
			wantErr:  false,
			exitCode: 0,
			checkFunc: func(t *testing.T, output string) {
				assert.NotEmpty(t, output)
				assert.Contains(t, output, "total")
			},
		},
		{
			name: "Pipeline with no matches",
			commands: []types.Command{
				{Command: "echo", Args: []string{"hello world"}},
				{Command: "grep", Args: []string{"nonexistentfile"}},
			},
			wantErr:  true,
			exitCode: 1,
			checkFunc: func(t *testing.T, output string) {
				assert.Empty(t, output)
			},
		},
		{
			name: "Multi-stage pipeline",
			commands: []types.Command{
				{Command: "printf", Args: []string{"hello\\nworld\\ntest\\n"}},
				{Command: "grep", Args: []string{"world"}},
				{Command: "tr", Args: []string{"a-z", "A-Z"}},
			},
			wantErr:  false,
			exitCode: 0,
			checkFunc: func(t *testing.T, output string) {
				assert.NotEmpty(t, output)
				assert.Contains(t, output, "WORLD")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建管道执行器
			pipeExec := NewPipelineExecutor(dockerExec)

			// 设置执行选项
			var output bytes.Buffer
			pipeline := &types.PipelineContext{
				Commands: make([]*types.Command, len(tt.commands)),
				Options: &types.ExecuteOptions{
					Stdout: &output,
				},
				Context: context.Background(),
			}

			// 设置命令
			for i := range tt.commands {
				cmd := tt.commands[i]
				pipeline.Commands[i] = &cmd
			}

			// 执行管道命令
			ctx := &types.ExecuteContext{
				Context:     context.Background(),
				IsPiped:     true,
				PipeContext: pipeline,
			}
			result, err := pipeExec.executePipeline(ctx)

			if tt.wantErr {
				assert.Error(t, err)
				if result != nil {
					assert.Equal(t, tt.exitCode, result.ExitCode)
				}
				if tt.checkFunc != nil {
					tt.checkFunc(t, result.Output)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.exitCode, result.ExitCode)
			if tt.checkFunc != nil {
				tt.checkFunc(t, result.Output)
			}
		})
	}
}

func TestDockerExecutor_WorkDir(t *testing.T) {
	tests := []struct {
		name        string
		dockerImage string
		workDir     string
		command     string
		args        []string
		wantOutput  string
		wantErr     bool
	}{
		{
			name:        "Custom work directory",
			dockerImage: "busybox:latest",
			workDir:     "/tmp",
			command:     "pwd",
			args:        []string{},
			wantOutput:  "/tmp",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建执行器
			exec, err := NewDockerExecutor(types.DockerConfig{
				Image:                     tt.dockerImage,
				WorkDir:                   tt.workDir,
				AllowUnregisteredCommands: true,
			}, &types.ExecuteOptions{})
			if err != nil {
				t.Fatalf("Failed to create Docker executor: %v", err)
			}
			defer exec.Close()

			// 准备执行上下文
			var output bytes.Buffer
			ctx := &types.ExecuteContext{
				Context: context.Background(),
				Command: types.Command{
					Command: tt.command,
					Args:    tt.args,
				},
				Options: &types.ExecuteOptions{
					Stdout: &output,
				},
			}

			// 执行命令
			result, err := exec.Execute(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
			assert.Contains(t, output.String(), tt.wantOutput)
		})
	}
}

func TestDockerExecutor_Environment(t *testing.T) {
	tests := []struct {
		name        string
		dockerImage string
		env         map[string]string
		command     string
		args        []string
		wantOutput  string
		wantErr     bool
	}{
		{
			name:        "Custom environment variable",
			dockerImage: "busybox:latest",
			env: map[string]string{
				"TEST_VAR": "test_value",
			},
			command:    "env",
			args:       []string{},
			wantOutput: "TEST_VAR=test_value",
			wantErr:    false,
		},
		{
			name:        "Multiple environment variables",
			dockerImage: "busybox:latest",
			env: map[string]string{
				"VAR1": "value1",
				"VAR2": "value2",
			},
			command:    "env",
			args:       []string{},
			wantOutput: "VAR1=value1",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建执行器
			exec, err := NewDockerExecutor(types.DockerConfig{
				Image:                     tt.dockerImage,
				AllowUnregisteredCommands: true,
			}, &types.ExecuteOptions{})
			if err != nil {
				t.Fatalf("Failed to create Docker executor: %v", err)
			}
			defer exec.Close()

			// 准备执行上下文
			var output bytes.Buffer
			ctx := &types.ExecuteContext{
				Context: context.Background(),
				Command: types.Command{
					Command: tt.command,
					Args:    tt.args,
				},
				Options: &types.ExecuteOptions{
					Env:    tt.env,
					Stdout: &output,
				},
			}

			// 执行命令
			result, err := exec.Execute(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
			assert.Contains(t, output.String(), tt.wantOutput)
		})
	}
}

func TestDockerExecutor_ExecuteWithBindMount(t *testing.T) {
	exec, err := NewDockerExecutor(types.DockerConfig{
		Image:                     "ubuntu:latest",
		BindMount:                 "/tmp:/workspace",
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{})
	if err != nil {
		t.Fatalf("Failed to create Docker executor: %v", err)
	}

	// 测试执行命令
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Command: types.Command{
			Command: "ls",
			Args:    []string{"/workspace"},
		},
		Options: &types.ExecuteOptions{},
	}

	result, err := exec.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
}

func TestDockerExecutor_ExecuteWithWorkDir(t *testing.T) {
	exec, err := NewDockerExecutor(types.DockerConfig{
		Image:                     "ubuntu:latest",
		WorkDir:                   "/app",
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{})
	if err != nil {
		t.Fatalf("Failed to create Docker executor: %v", err)
	}

	// 测试执行命令
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Command: types.Command{
			Command: "pwd",
		},
		Options: &types.ExecuteOptions{},
	}

	result, err := exec.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
}

func TestDockerExecutor_ExecuteWithEnv(t *testing.T) {
	exec, err := NewDockerExecutor(types.DockerConfig{
		Image:                     "ubuntu:latest",
		AllowUnregisteredCommands: true,
	}, &types.ExecuteOptions{})
	if err != nil {
		t.Fatalf("Failed to create Docker executor: %v", err)
	}

	// 测试执行命令
	ctx := &types.ExecuteContext{
		Context: context.Background(),
		Command: types.Command{
			Command: "env",
		},
		Options: &types.ExecuteOptions{
			Env: map[string]string{
				"TEST_VAR": "test_value",
			},
		},
	}

	result, err := exec.Execute(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Output, "TEST_VAR=test_value")
}
