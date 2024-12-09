package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestSessionManagement(t *testing.T) {
	// 创建服务器
	srv := NewServer(executor.NewLocalExecutor(), "localhost:0")
	assert.NoError(t, srv.Start())
	defer srv.Stop()

	// 测试创建会话
	t.Run("Create Session", func(t *testing.T) {
		req := types.SessionRequest{
			ExecutorType: "local",
			Options: &types.ExecuteOptions{
				WorkDir: "/tmp",
				Env:     map[string]string{"TEST": "true"},
			},
		}
		body, err := json.Marshal(req)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/sessions", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		srv.handleSessions(w, r)

		t.Logf("Response Status: %d", w.Code)
		t.Logf("Response Body: %s", w.Body.String())

		assert.Equal(t, http.StatusOK, w.Code)

		var resp types.SessionResponse
		err = json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Session.ID)
		assert.Equal(t, "active", resp.Session.Status)
		assert.NotZero(t, resp.Session.CreatedAt)
		assert.NotZero(t, resp.Session.LastAccessedAt)

		// 保存会话ID用于后续测试
		sessionID := resp.Session.ID

		// 测试列出会话
		t.Run("List Sessions", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/sessions", nil)
			srv.handleSessions(w, r)

			t.Logf("List Response Status: %d", w.Code)
			t.Logf("List Response Body: %s", w.Body.String())

			assert.Equal(t, http.StatusOK, w.Code)

			var sessions []*types.Session
			err := json.NewDecoder(w.Body).Decode(&sessions)
			assert.NoError(t, err)
			assert.NotEmpty(t, sessions)
			assert.Equal(t, sessionID, sessions[0].ID)
		})

		// 测试在会话中执行命令
		t.Run("Execute Command", func(t *testing.T) {
			execReq := types.ExecRequest{
				Command: "echo",
				Args:    []string{"hello"},
			}
			body, err := json.Marshal(execReq)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/sessions/%s", sessionID), bytes.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
			srv.handleSessionOperations(w, r)

			t.Logf("Exec Response Status: %d", w.Code)
			t.Logf("Exec Response Body: %s", w.Body.String())

			assert.Equal(t, http.StatusOK, w.Code)

			var result types.ExecuteResult
			err = json.NewDecoder(w.Body).Decode(&result)
			assert.NoError(t, err)
			assert.Equal(t, 0, result.ExitCode)
		})

		// 测试删除会话
		t.Run("Delete Session", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/sessions/%s", sessionID), nil)
			srv.handleSessionOperations(w, r)

			t.Logf("Delete Response Status: %d", w.Code)
			if w.Body.Len() > 0 {
				t.Logf("Delete Response Body: %s", w.Body.String())
			}

			assert.Equal(t, http.StatusNoContent, w.Code)

			// 验证会话已被删除
			w = httptest.NewRecorder()
			r = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/sessions/%s", sessionID), nil)
			srv.handleSessionOperations(w, r)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})

	// 测试Docker执行器会话
	t.Run("Docker Session", func(t *testing.T) {
		req := types.SessionRequest{
			ExecutorType: "docker",
			DockerImage:  "alpine:latest",
		}
		body, err := json.Marshal(req)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/sessions", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		srv.handleSessions(w, r)

		t.Logf("Docker Session Response Status: %d", w.Code)
		t.Logf("Docker Session Response Body: %s", w.Body.String())

		assert.Equal(t, http.StatusOK, w.Code)

		var resp types.SessionResponse
		err = json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Session.ID)
		assert.Equal(t, "active", resp.Session.Status)
		assert.NotZero(t, resp.Session.CreatedAt)
		assert.NotZero(t, resp.Session.LastAccessedAt)

		// 在Docker会话中执行命令
		execReq := types.ExecRequest{
			Command: "echo",
			Args:    []string{"hello from docker"},
		}
		body, err = json.Marshal(execReq)
		assert.NoError(t, err)

		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/sessions/%s", resp.Session.ID), bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		srv.handleSessionOperations(w, r)

		t.Logf("Docker Exec Response Status: %d", w.Code)
		t.Logf("Docker Exec Response Body: %s", w.Body.String())

		assert.Equal(t, http.StatusOK, w.Code)

		var result types.ExecuteResult
		err = json.NewDecoder(w.Body).Decode(&result)
		assert.NoError(t, err)
		assert.Equal(t, 0, result.ExitCode)
	})
}
