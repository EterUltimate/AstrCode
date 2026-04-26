package model

import (
	"sync"
	"time"
)

// TaskStore 任务状态存储
type TaskStore struct {
	tasks map[string]*Task
	plans map[string]*Plan
	mu    sync.RWMutex
}

// NewTaskStore 创建任务存储
func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks: make(map[string]*Task),
		plans: make(map[string]*Plan),
	}
}

// CreateTask 创建任务
func (s *TaskStore) CreateTask(task *Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = task
}

// GetTask 获取任务
func (s *TaskStore) GetTask(id string) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

// UpdateTask 更新任务
func (s *TaskStore) UpdateTask(task *Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = task
}

// SavePlan 保存计划
func (s *TaskStore) SavePlan(plan *Plan) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.plans[plan.ID] = plan
}

// GetPlan 获取计划
func (s *TaskStore) GetPlan(id string) (*Plan, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.plans[id]
	return p, ok
}

// ListTasks 列出所有任务
func (s *TaskStore) ListTasks(limit int) []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

// ListRunning 列出运行中的任务
func (s *TaskStore) ListRunning() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var tasks []*Task
	for _, t := range s.tasks {
		if t.Status == TaskStatusExecuting || t.Status == TaskStatusPlanning {
			tasks = append(tasks, t)
		}
	}
	return tasks
}

// ============================================================
// StepTimeline 步骤时间线（用于可视化）
// ============================================================

// StepTimelineItem 步骤时间线条目
type StepTimelineItem struct {
	StepID    string `json:"step_id"`
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message,omitempty"`
	Duration  int64  `json:"duration_ms,omitempty"`
}

// StepTimeline 步骤时间线
type StepTimeline struct {
	Items []StepTimelineItem `json:"items"`
}

// Add 添加事件
func (t *StepTimeline) Add(stepID, status, message string) {
	t.Items = append(t.Items, StepTimelineItem{
		StepID:    stepID,
		Status:    status,
		Timestamp: time.Now().UnixMilli(),
		Message:   message,
	})
}

// AddWithDuration 添加带耗时的结束事件
func (t *StepTimeline) AddWithDuration(stepID, status, message string, startNano int64) {
	t.Items = append(t.Items, StepTimelineItem{
		StepID:    stepID,
		Status:    status,
		Timestamp: time.Now().UnixMilli(),
		Message:   message,
		Duration:  (time.Now().UnixNano() - startNano) / 1e6,
	})
}

// ============================================================
// WebSocket 事件（实时推送）
// ============================================================

// WSEvent WebSocket 事件
type WSEvent struct {
	Type      string      `json:"type"` // task_created / task_updated / step_started / step_completed / step_failed / plan_ready
	TaskID    string      `json:"task_id"`
	PlanID    string      `json:"plan_id,omitempty"`
	StepID    string      `json:"step_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

// NewWSEvent 创建 WebSocket 事件
func NewWSEvent(eventType, taskID, stepID string, data interface{}) *WSEvent {
	return &WSEvent{
		Type:      eventType,
		TaskID:    taskID,
		StepID:    stepID,
		Timestamp: time.Now().UnixMilli(),
		Data:      data,
	}
}

// ============================================================
// Execution Snapshot（用于 API 查询当前执行状态）
// ============================================================

// ExecutionSnapshot 执行快照
type ExecutionSnapshot struct {
	TaskID      string         `json:"task_id"`
	PlanID      string         `json:"plan_id"`
	TaskStatus  string         `json:"task_status"`
	Steps       []StepSnapshot `json:"steps"`
	Timeline    *StepTimeline  `json:"timeline,omitempty"`
	StartedAt   int64          `json:"started_at"`
	CompletedAt int64          `json:"completed_at,omitempty"`
}

// StepSnapshot 步骤快照
type StepSnapshot struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Skill      string      `json:"skill,omitempty"`
	Status     string      `json:"status"`
	DependsOn  []string    `json:"depends_on,omitempty"`
	Result     interface{} `json:"result,omitempty"`
	Error      string      `json:"error,omitempty"`
	RetryCount int         `json:"retry_count"`
	Duration   int64       `json:"duration_ms,omitempty"`
}

// BuildSnapshot 从 Plan 构建执行快照
func BuildSnapshot(task *Task, plan *Plan, timeline *StepTimeline) *ExecutionSnapshot {
	if plan == nil {
		return &ExecutionSnapshot{
			TaskID:      task.ID,
			TaskStatus:  task.Status,
			StartedAt:   task.CreatedAt,
			CompletedAt: task.CompletedAt,
		}
	}

	steps := make([]StepSnapshot, len(plan.Steps))
	for i, s := range plan.Steps {
		steps[i] = StepSnapshot{
			ID:         s.ID,
			Type:       string(s.Type),
			Skill:      s.Skill,
			Status:     string(s.Status),
			DependsOn:  s.DependsOn,
			Result:     s.Result,
			Error:      s.Error,
			RetryCount: s.RetryCount,
		}
	}

	return &ExecutionSnapshot{
		TaskID:      task.ID,
		PlanID:      plan.ID,
		TaskStatus:  task.Status,
		Steps:       steps,
		Timeline:    timeline,
		StartedAt:   task.CreatedAt,
		CompletedAt: task.CompletedAt,
	}
}
