// AstrCode - Agent orchestration engine for AstrBot
// Copyright (C) 2026 EterUltimate
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
