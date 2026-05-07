package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

func setupTestSession(t *testing.T) (*Manager, string) {
	tmpDir := filepath.Join(os.TempDir(), "astrcode-session-test")
	os.RemoveAll(tmpDir) // 清理旧数据

	manager, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	return manager, tmpDir
}

func cleanupTestSession(tmpDir string) {
	os.RemoveAll(tmpDir)
}

func TestEventLog_BasicAppend(t *testing.T) {
	manager, tmpDir := setupTestSession(t)
	defer cleanupTestSession(tmpDir)

	session, err := manager.CreateSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// 追加事件
	event := &Event{
		ID:        uuid.New().String(),
		Type:      EventUserMessage,
		Timestamp: time.Now(),
		SessionID: session.ID,
		Data: map[string]interface{}{
			"message": "Hello, world!",
		},
	}

	err = session.EventLog.Append(event)
	if err != nil {
		t.Fatalf("Failed to append event: %v", err)
	}

	// 验证事件数量
	count, err := session.EventLog.Count()
	if err != nil {
		t.Fatalf("Failed to count events: %v", err)
	}

	if count != 2 { // session_start + user_message
		t.Errorf("Expected 2 events, got %d", count)
	}
}

func TestEventLog_ReadAll(t *testing.T) {
	manager, tmpDir := setupTestSession(t)
	defer cleanupTestSession(tmpDir)

	session, err := manager.CreateSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// 追加多个事件
	for i := 0; i < 5; i++ {
		event := &Event{
			ID:        uuid.New().String(),
			Type:      EventUserMessage,
			Timestamp: time.Now(),
			SessionID: session.ID,
			Data: map[string]interface{}{
				"index": i,
			},
		}
		if err := session.EventLog.Append(event); err != nil {
			t.Fatalf("Failed to append event %d: %v", i, err)
		}
	}

	// 读取所有事件
	events, err := session.EventLog.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read all events: %v", err)
	}

	if len(events) != 6 { // session_start + 5 user_messages
		t.Errorf("Expected 6 events, got %d", len(events))
	}
}

func TestEventLog_FilterByType(t *testing.T) {
	manager, tmpDir := setupTestSession(t)
	defer cleanupTestSession(tmpDir)

	session, err := manager.CreateSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// 追加不同类型的事件
	userMsg := &Event{
		ID:        uuid.New().String(),
		Type:      EventUserMessage,
		Timestamp: time.Now(),
		SessionID: session.ID,
	}
	toolCall := &Event{
		ID:        uuid.New().String(),
		Type:      EventToolCall,
		Timestamp: time.Now(),
		SessionID: session.ID,
	}

	session.EventLog.Append(userMsg)
	session.EventLog.Append(toolCall)

	// 过滤特定类型
	userEvents, err := session.EventLog.FilterByType(EventUserMessage)
	if err != nil {
		t.Fatalf("Failed to filter events: %v", err)
	}

	if len(userEvents) != 1 {
		t.Errorf("Expected 1 user message event, got %d", len(userEvents))
	}
}

func TestEventLog_LastN(t *testing.T) {
	manager, tmpDir := setupTestSession(t)
	defer cleanupTestSession(tmpDir)

	session, err := manager.CreateSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// 追加 10 个事件
	for i := 0; i < 10; i++ {
		event := &Event{
			ID:        uuid.New().String(),
			Type:      EventUserMessage,
			Timestamp: time.Now(),
			SessionID: session.ID,
			Data: map[string]interface{}{
				"index": i,
			},
		}
		session.EventLog.Append(event)
	}

	// 获取最后 3 个
	lastEvents, err := session.EventLog.LastN(3)
	if err != nil {
		t.Fatalf("Failed to get last N events: %v", err)
	}

	if len(lastEvents) != 3 {
		t.Errorf("Expected 3 events, got %d", len(lastEvents))
	}

	// 验证是最后 3 个（索引 7, 8, 9）
	if lastEvents[0].Data["index"] != 7.0 {
		t.Errorf("Expected first of last 3 to be index 7, got %v", lastEvents[0].Data["index"])
	}
}

func TestEventLog_ReplayFrom(t *testing.T) {
	manager, tmpDir := setupTestSession(t)
	defer cleanupTestSession(tmpDir)

	session, err := manager.CreateSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// 记录起始事件 ID
	var targetEventID string

	// 追加事件
	for i := 0; i < 5; i++ {
		event := &Event{
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			SessionID: session.ID,
			Type:      EventUserMessage,
			Data: map[string]interface{}{
				"index": i,
			},
		}
		if i == 2 {
			targetEventID = event.ID
		}
		session.EventLog.Append(event)
	}

	// 从指定事件回放
	replayedEvents, err := session.EventLog.ReplayFrom(targetEventID)
	if err != nil {
		t.Fatalf("Failed to replay from event: %v", err)
	}

	if len(replayedEvents) != 3 { // 索引 2, 3, 4
		t.Errorf("Expected 3 events in replay, got %d", len(replayedEvents))
	}
}

func TestSession_CreateAndEnd(t *testing.T) {
	manager, tmpDir := setupTestSession(t)
	defer cleanupTestSession(tmpDir)

	// 创建会话
	session, err := manager.CreateSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.ID == "" {
		t.Error("Session ID should not be empty")
	}

	// 结束会话
	err = manager.EndSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to end session: %v", err)
	}

	// 验证会话已关闭
	_, err = manager.GetSession(session.ID)
	if err == nil {
		t.Error("Session should not exist after ending")
	}
}

func TestSession_Restore(t *testing.T) {
	manager, tmpDir := setupTestSession(t)
	defer cleanupTestSession(tmpDir)

	// 创建并结束会话
	session, err := manager.CreateSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// 追加一些事件
	event := &Event{
		ID:        uuid.New().String(),
		Type:      EventUserMessage,
		Timestamp: time.Now(),
		SessionID: sessionID,
	}
	session.EventLog.Append(event)

	// 结束会话
	manager.EndSession(sessionID)

	// 恢复会话
	restoredSession, err := manager.RestoreSession(sessionID)
	if err != nil {
		t.Fatalf("Failed to restore session: %v", err)
	}

	if restoredSession.ID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, restoredSession.ID)
	}

	// 验证事件日志可用
	count, err := restoredSession.EventLog.Count()
	if err != nil {
		t.Fatalf("Failed to count events in restored session: %v", err)
	}

	if count < 2 { // session_start + user_message (+ session_restore)
		t.Errorf("Expected at least 2 events in restored session, got %d", count)
	}
}

func TestEventLog_ThreadSafety(t *testing.T) {
	manager, tmpDir := setupTestSession(t)
	defer cleanupTestSession(tmpDir)

	session, err := manager.CreateSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// 并发写入
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			event := &Event{
				ID:        uuid.New().String(),
				Type:      EventUserMessage,
				Timestamp: time.Now(),
				SessionID: session.ID,
				Data: map[string]interface{}{
					"goroutine": index,
				},
			}
			session.EventLog.Append(event)
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证所有事件都已写入
	count, err := session.EventLog.Count()
	if err != nil {
		t.Fatalf("Failed to count events: %v", err)
	}

	expectedCount := 11 // session_start + 10 concurrent writes
	if count != expectedCount {
		t.Errorf("Expected %d events, got %d", expectedCount, count)
	}
}
