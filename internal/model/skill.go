package model

// Skill 表示一个可执行的技能
type Skill struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Embedding   []float32 `json:"embedding,omitempty"`
	Path        string    `json:"path"`
	Summary     string    `json:"summary,omitempty"`
	Usage       string    `json:"usage,omitempty"`
	Full        string    `json:"full,omitempty"`
}

// Task 表示用户提交的任务
type Task struct {
	ID          string `json:"id"`
	Content     string `json:"content"`
	Status      string `json:"status"`
	Result      string `json:"result,omitempty"`
	Error       string `json:"error,omitempty"`
	CreatedAt   int64  `json:"created_at"`
	CompletedAt int64  `json:"completed_at,omitempty"`
}

// TaskStatus 定义任务状态
const (
	TaskStatusPending   = "pending"
	TaskStatusPlanning  = "planning"
	TaskStatusExecuting = "executing"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
)
