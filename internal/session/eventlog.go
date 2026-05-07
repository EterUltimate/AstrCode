package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// EventType 定义事件类型
type EventType string

const (
	// Session 生命周期
	EventSessionStart   EventType = "session_start"
	EventSessionEnd     EventType = "session_end"
	EventSessionRestore EventType = "session_restore"

	// Turn 相关
	EventTurnStart EventType = "turn_start"
	EventTurnEnd   EventType = "turn_end"

	// 消息
	EventUserMessage      EventType = "user_message"
	EventAssistantMessage EventType = "assistant_message"

	// Tool 执行
	EventToolCall   EventType = "tool_call"
	EventToolResult EventType = "tool_result"

	// LLM 调用
	EventLLMCallRequest  EventType = "llm_call_request"
	EventLLMCallResponse EventType = "llm_call_response"

	// Context 管理
	EventContextCompaction EventType = "context_compaction"
	EventSnapshotCreated   EventType = "snapshot_created"

	// 错误
	EventError EventType = "error"
)

// Event 表示一个会话事件
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	SessionID string                 `json:"session_id"`
	TurnID    string                 `json:"turn_id,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
}

// EventLog 管理 JSONL 格式的事件日志
type EventLog struct {
	mu       sync.Mutex
	filePath string
	file     *os.File
}

// NewEventLog 创建新的事件日志
func NewEventLog(sessionDir string, sessionID string) (*EventLog, error) {
	// 确保目录存在
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	filePath := filepath.Join(sessionDir, fmt.Sprintf("%s_events.jsonl", sessionID))

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open event log file: %w", err)
	}

	return &EventLog{
		filePath: filePath,
		file:     file,
	}, nil
}

// Append 追加事件到日志（线程安全）
func (el *EventLog) Append(event *Event) error {
	el.mu.Lock()
	defer el.mu.Unlock()

	// 序列化事件为 JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// 写入 JSONL 格式（每行一个 JSON 对象）
	if _, err := el.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	// 立即刷盘，确保数据持久化
	if err := el.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	return nil
}

// ReadAll 读取所有事件
func (el *EventLog) ReadAll() ([]*Event, error) {
	el.mu.Lock()
	defer el.mu.Unlock()

	// 关闭当前文件以便读取
	if err := el.file.Close(); err != nil {
		return nil, fmt.Errorf("failed to close file for reading: %w", err)
	}

	// 重新打开文件进行读取
	file, err := os.Open(el.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for reading: %w", err)
	}
	defer file.Close()

	var events []*Event
	decoder := json.NewDecoder(file)

	for decoder.More() {
		var event Event
		if err := decoder.Decode(&event); err != nil {
			return nil, fmt.Errorf("failed to decode event: %w", err)
		}
		events = append(events, &event)
	}

	// 重新打开文件用于追加写入
	newFile, err := os.OpenFile(el.filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to reopen file for writing: %w", err)
	}
	el.file = newFile

	return events, nil
}

// ReplayFrom 从指定事件 ID 开始回放事件
func (el *EventLog) ReplayFrom(fromEventID string) ([]*Event, error) {
	allEvents, err := el.ReadAll()
	if err != nil {
		return nil, err
	}

	// 找到起始事件索引
	startIndex := 0
	for i, event := range allEvents {
		if event.ID == fromEventID {
			startIndex = i
			break
		}
	}

	return allEvents[startIndex:], nil
}

// Close 关闭事件日志
func (el *EventLog) Close() error {
	el.mu.Lock()
	defer el.mu.Unlock()

	if el.file != nil {
		return el.file.Close()
	}
	return nil
}

// GetFilePath 获取日志文件路径
func (el *EventLog) GetFilePath() string {
	return el.filePath
}

// Count 获取事件数量
func (el *EventLog) Count() (int, error) {
	events, err := el.ReadAll()
	if err != nil {
		return 0, err
	}
	return len(events), nil
}

// LastN 获取最后 N 个事件
func (el *EventLog) LastN(n int) ([]*Event, error) {
	allEvents, err := el.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(allEvents) <= n {
		return allEvents, nil
	}

	return allEvents[len(allEvents)-n:], nil
}

// FilterByType 按类型过滤事件
func (el *EventLog) FilterByType(eventType EventType) ([]*Event, error) {
	allEvents, err := el.ReadAll()
	if err != nil {
		return nil, err
	}

	var filtered []*Event
	for _, event := range allEvents {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}

	return filtered, nil
}
