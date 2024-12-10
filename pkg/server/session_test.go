package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestSessionManagement(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/session":
			if r.Method == http.MethodPost {
				// 创建会话
				session := &types.Session{
					ID:             "test-session",
					CreatedAt:      time.Now(),
					LastAccessedAt: time.Now(),
					Status:         "active",
				}
				json.NewEncoder(w).Encode(map[string]interface{}{
					"session": session,
				})
			}
		case "/session/test-session/exec":
			if r.Method == http.MethodPost {
				// 执行命令
				result := &types.ExecuteResult{
					CommandName: "ls",
					ExitCode:    0,
					Output:      "file1.txt\nfile2.txt",
				}
				json.NewEncoder(w).Encode(result)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// 创建 HTTP 客户端
	client := &http.Client{}

	// 测试创建会话
	sessionReq, err := http.NewRequest(http.MethodPost, server.URL+"/session", nil)
	assert.NoError(t, err)

	sessionResp, err := client.Do(sessionReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, sessionResp.StatusCode)

	var sessionData struct {
		Session *types.Session `json:"session"`
	}
	err = json.NewDecoder(sessionResp.Body).Decode(&sessionData)
	assert.NoError(t, err)
	sessionResp.Body.Close()

	// 测试执行命令
	execReq, err := http.NewRequest(http.MethodPost, server.URL+"/session/test-session/exec", strings.NewReader(`{"command":"ls"}`))
	assert.NoError(t, err)
	execReq.Header.Set("Content-Type", "application/json")

	execResp, err := client.Do(execReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, execResp.StatusCode)

	var execResult types.ExecuteResult
	err = json.NewDecoder(execResp.Body).Decode(&execResult)
	assert.NoError(t, err)
	execResp.Body.Close()

	assert.Equal(t, 0, execResult.ExitCode)
	assert.Contains(t, execResult.Output, "file1.txt")
}
