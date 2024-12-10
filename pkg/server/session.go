package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/iamlongalong/runshell/pkg/executor"
	"github.com/iamlongalong/runshell/pkg/types"
)

// MemorySessionManager 内存会话管理器
type MemorySessionManager struct {
	sessions sync.Map
}

// NewMemorySessionManager 创建新的内存会话管理器
func NewMemorySessionManager() *MemorySessionManager {
	return &MemorySessionManager{}
}

// CreateSession 创建新的会话
func (m *MemorySessionManager) CreateSession(executor types.Executor, options *types.ExecuteOptions) (*types.Session, error) {
	// 创建会话ID
	id := uuid.New().String()

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 创建会话
	session := &types.Session{
		ID:             id,
		Executor:       executor,
		Options:        options,
		Context:        ctx,
		Cancel:         cancel,
		CreatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
		Metadata:       make(map[string]string),
		Status:         "active",
	}

	// 存储会话
	m.sessions.Store(id, session)

	return session, nil
}

// GetSession 获取会话
func (m *MemorySessionManager) GetSession(id string) (*types.Session, error) {
	if session, ok := m.sessions.Load(id); ok {
		s := session.(*types.Session)
		s.LastAccessedAt = time.Now()
		return s, nil
	}
	return nil, fmt.Errorf("session not found: %s", id)
}

// ListSessions 列出所有会话
func (m *MemorySessionManager) ListSessions() ([]*types.Session, error) {
	var sessions []*types.Session
	m.sessions.Range(func(key, value interface{}) bool {
		sessions = append(sessions, value.(*types.Session))
		return true
	})
	return sessions, nil
}

// DeleteSession 删除会话
func (m *MemorySessionManager) DeleteSession(id string) error {
	if session, ok := m.sessions.Load(id); ok {
		s := session.(*types.Session)
		s.Cancel() // 取消会话上下文

		// 关闭执行器
		if err := s.Executor.Close(); err != nil {
			return fmt.Errorf("failed to close executor: %v", err)
		}

		m.sessions.Delete(id)
		return nil
	}
	return fmt.Errorf("session not found: %s", id)
}

// UpdateSession 更新会话
func (m *MemorySessionManager) UpdateSession(session *types.Session) error {
	if _, ok := m.sessions.Load(session.ID); !ok {
		return fmt.Errorf("session not found: %s", session.ID)
	}
	session.LastAccessedAt = time.Now()
	m.sessions.Store(session.ID, session)
	return nil
}

// CreateExecutor 创建执行器
func CreateExecutor(req *types.SessionRequest) (types.Executor, error) {
	switch req.ExecutorType {
	case types.ExecutorTypeLocal:
		return executor.NewLocalExecutor(*req.LocalConfig, req.Options), nil
	case types.ExecutorTypeDocker:
		exec, err := executor.NewDockerExecutor(*req.DockerConfig, req.Options)
		if err != nil {
			return nil, fmt.Errorf("failed to create Docker executor: %w", err)
		}
		return exec, nil
	default:
		return nil, fmt.Errorf("unsupported executor type: %s", req.ExecutorType)
	}
}
