package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// SnapshotType 定义快照类型
type SnapshotType string

const (
	SnapshotTypeFull     SnapshotType = "full"     // 完整快照
	SnapshotTypeDelta    SnapshotType = "delta"    // 增量快照
	SnapshotTypeDecision SnapshotType = "decision" // 决策点快照
)

// Snapshot 表示一个会话快照
type Snapshot struct {
	ID            string                 `json:"id"`
	Type          SnapshotType           `json:"type"`
	Timestamp     time.Time              `json:"timestamp"`
	SessionID     string                 `json:"session_id"`
	EventLogStart string                 `json:"event_log_start"` // 快照起始事件 ID
	EventLogEnd   string                 `json:"event_log_end"`   // 快照结束事件 ID
	State         map[string]interface{} `json:"state"`           // 会话状态
	Metadata      map[string]string      `json:"metadata,omitempty"`
}

// SnapshotManager 管理快照
type SnapshotManager struct {
	baseDir          string
	snapshots        map[string]*Snapshot
	lastFullSnapshot *Snapshot // 最后一个完整快照
}

// NewSnapshotManager 创建快照管理器
func NewSnapshotManager(baseDir string) (*SnapshotManager, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create snapshots directory: %w", err)
	}

	return &SnapshotManager{
		baseDir:   baseDir,
		snapshots: make(map[string]*Snapshot),
	}, nil
}

// CreateSnapshot 创建快照
func (sm *SnapshotManager) CreateSnapshot(session *Session, snapshotType SnapshotType, state map[string]interface{}) (*Snapshot, error) {
	snapshotID := uuid.New().String()

	// 获取事件日志范围
	events, err := session.EventLog.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read events: %w", err)
	}

	var eventLogStart, eventLogEnd string
	if len(events) > 0 {
		eventLogStart = events[0].ID
		eventLogEnd = events[len(events)-1].ID
	}

	snapshot := &Snapshot{
		ID:            snapshotID,
		Type:          snapshotType,
		Timestamp:     time.Now(),
		SessionID:     session.ID,
		EventLogStart: eventLogStart,
		EventLogEnd:   eventLogEnd,
		State:         state,
		Metadata:      make(map[string]string),
	}

	// 保存快照到磁盘
	if err := sm.saveSnapshot(snapshot); err != nil {
		return nil, fmt.Errorf("failed to save snapshot: %w", err)
	}

	sm.snapshots[snapshotID] = snapshot

	// 更新最后完整快照引用
	if snapshotType == SnapshotTypeFull {
		sm.lastFullSnapshot = snapshot
	}

	// 记录 snapshot_created 事件
	snapshotEvent := &Event{
		ID:        uuid.New().String(),
		Type:      EventSnapshotCreated,
		Timestamp: time.Now(),
		SessionID: session.ID,
		Data: map[string]interface{}{
			"snapshot_id":   snapshotID,
			"snapshot_type": string(snapshotType),
		},
	}

	if err := session.EventLog.Append(snapshotEvent); err != nil {
		return nil, fmt.Errorf("failed to log snapshot creation: %w", err)
	}

	return snapshot, nil
}

// GetSnapshot 获取快照
func (sm *SnapshotManager) GetSnapshot(snapshotID string) (*Snapshot, error) {
	snapshot, exists := sm.snapshots[snapshotID]
	if !exists {
		// 尝试从磁盘加载
		if err := sm.loadSnapshot(snapshotID); err != nil {
			return nil, fmt.Errorf("snapshot not found: %s", snapshotID)
		}
		snapshot = sm.snapshots[snapshotID]
	}
	return snapshot, nil
}

// RestoreFromSnapshot 从快照恢复会话状态
func (sm *SnapshotManager) RestoreFromSnapshot(session *Session, snapshotID string) (map[string]interface{}, error) {
	snapshot, err := sm.GetSnapshot(snapshotID)
	if err != nil {
		return nil, err
	}

	// 如果需要，回放快照之后的事件
	if snapshot.EventLogEnd != "" {
		eventsAfterSnapshot, err := session.EventLog.ReplayFrom(snapshot.EventLogEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to replay events after snapshot: %w", err)
		}

		// TODO: 根据事件类型应用状态变更
		_ = eventsAfterSnapshot
	}

	return snapshot.State, nil
}

// GetLastFullSnapshot 获取最后一个完整快照
func (sm *SnapshotManager) GetLastFullSnapshot() *Snapshot {
	return sm.lastFullSnapshot
}

// ListSnapshots 列出所有快照
func (sm *SnapshotManager) ListSnapshots() []*Snapshot {
	var snapshots []*Snapshot
	for _, snapshot := range sm.snapshots {
		snapshots = append(snapshots, snapshot)
	}
	return snapshots
}

// DeleteSnapshot 删除快照
func (sm *SnapshotManager) DeleteSnapshot(snapshotID string) error {
	delete(sm.snapshots, snapshotID)

	// 删除磁盘文件
	filePath := filepath.Join(sm.baseDir, fmt.Sprintf("%s.json", snapshotID))
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete snapshot file: %w", err)
	}

	return nil
}

// saveSnapshot 保存快照到磁盘
func (sm *SnapshotManager) saveSnapshot(snapshot *Snapshot) error {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	filePath := filepath.Join(sm.baseDir, fmt.Sprintf("%s.json", snapshot.ID))
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot file: %w", err)
	}

	return nil
}

// loadSnapshot 从磁盘加载快照
func (sm *SnapshotManager) loadSnapshot(snapshotID string) error {
	filePath := filepath.Join(sm.baseDir, fmt.Sprintf("%s.json", snapshotID))

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read snapshot file: %w", err)
	}

	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}

	sm.snapshots[snapshotID] = &snapshot
	return nil
}
