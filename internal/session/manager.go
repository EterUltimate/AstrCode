package session

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// Session 表示一个会话
type Session struct {
	ID        string                 `json:"id"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	EventLog  *EventLog              `json:"-"`
}

// Manager 管理多个会话
type Manager struct {
	baseDir  string
	sessions map[string]*Session
}

// NewManager 创建会话管理器
func NewManager(baseDir string) (*Manager, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sessions directory: %w", err)
	}

	return &Manager{
		baseDir:  baseDir,
		sessions: make(map[string]*Session),
	}, nil
}

// CreateSession 创建新会话
func (m *Manager) CreateSession() (*Session, error) {
	sessionID := uuid.New().String()
	sessionDir := filepath.Join(m.baseDir, sessionID)

	eventLog, err := NewEventLog(sessionDir, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to create event log: %w", err)
	}

	session := &Session{
		ID:        sessionID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
		EventLog:  eventLog,
	}

	m.sessions[sessionID] = session

	// 记录 session_start 事件
	startEvent := &Event{
		ID:        uuid.New().String(),
		Type:      EventSessionStart,
		Timestamp: time.Now(),
		SessionID: sessionID,
		Data: map[string]interface{}{
			"created_at": session.CreatedAt,
		},
	}

	if err := eventLog.Append(startEvent); err != nil {
		return nil, fmt.Errorf("failed to log session start: %w", err)
	}

	return session, nil
}

// GetSession 获取会话
func (m *Manager) GetSession(sessionID string) (*Session, error) {
	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return session, nil
}

// EndSession 结束会话
func (m *Manager) EndSession(sessionID string) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}

	// 记录 session_end 事件
	endEvent := &Event{
		ID:        uuid.New().String(),
		Type:      EventSessionEnd,
		Timestamp: time.Now(),
		SessionID: sessionID,
		Data: map[string]interface{}{
			"ended_at": time.Now(),
		},
	}

	if err := session.EventLog.Append(endEvent); err != nil {
		return fmt.Errorf("failed to log session end: %w", err)
	}

	// 关闭事件日志
	if err := session.EventLog.Close(); err != nil {
		return fmt.Errorf("failed to close event log: %w", err)
	}

	// 从内存中移除
	delete(m.sessions, sessionID)

	return nil
}

// ListSessions 列出所有会话
func (m *Manager) ListSessions() []*Session {
	var sessions []*Session
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// RestoreSession 从磁盘恢复会话
func (m *Manager) RestoreSession(sessionID string) (*Session, error) {
	sessionDir := filepath.Join(m.baseDir, sessionID)

	// 检查目录是否存在
	if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("session directory not found: %s", sessionID)
	}

	// 重新打开事件日志
	eventLog, err := NewEventLog(sessionDir, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to restore event log: %w", err)
	}

	session := &Session{
		ID:        sessionID,
		CreatedAt: time.Now(), // TODO: 从事件中恢复实际时间
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
		EventLog:  eventLog,
	}

	m.sessions[sessionID] = session

	// 记录 session_restore 事件
	restoreEvent := &Event{
		ID:        uuid.New().String(),
		Type:      EventSessionRestore,
		Timestamp: time.Now(),
		SessionID: sessionID,
		Data: map[string]interface{}{
			"restored_at": time.Now(),
		},
	}

	if err := eventLog.Append(restoreEvent); err != nil {
		return nil, fmt.Errorf("failed to log session restore: %w", err)
	}

	return session, nil
}
