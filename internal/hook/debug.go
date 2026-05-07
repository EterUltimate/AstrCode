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

package hook

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DebugHook Hook 调试处理器（Phase 6）
type DebugHook struct {
	mu          sync.Mutex
	executions  []HookExecution
	maxHistory  int
	enabled     bool
}

// HookExecution 记录单次 Hook 执行
type HookExecution struct {
	ID        string                 `json:"id"`
	HookType  HookType               `json:"hook_type"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration_ms"`
	Error     string                 `json:"error,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// NewDebugHook 创建调试 Hook
func NewDebugHook(maxHistory int) *DebugHook {
	return &DebugHook{
		maxHistory: maxHistory,
		executions: make([]HookExecution, 0, maxHistory),
		enabled:    true,
	}
}

// Handle 实现 HookHandler 接口
func (d *DebugHook) Handle(ctx context.Context, event *HookEvent) error {
	if !d.enabled {
		return nil
	}

	start := time.Now()
	executionID := fmt.Sprintf("%s_%d", event.Type, start.UnixNano())

	// 执行实际逻辑（这里只是记录，不干预）
	err := d.recordExecution(executionID, event, start)

	return err
}

// recordExecution 记录执行信息
func (d *DebugHook) recordExecution(id string, event *HookEvent, start time.Time) error {
	duration := time.Since(start)

	d.mu.Lock()
	defer d.mu.Unlock()

	// 类型断言确保 Data 是 map[string]interface{}
	var dataMap map[string]interface{}
	if event.Data != nil {
		if m, ok := event.Data.(map[string]interface{}); ok {
			dataMap = m
		}
	}

	execution := HookExecution{
		ID:        id,
		HookType:  event.Type,
		Timestamp: start,
		Duration:  duration,
		Data:      dataMap,
	}

	d.executions = append(d.executions, execution)

	// 保持历史记录不超过限制
	if len(d.executions) > d.maxHistory {
		d.executions = d.executions[1:]
	}

	return nil
}

// GetExecutions 获取执行历史
func (d *DebugHook) GetExecutions() []HookExecution {
	d.mu.Lock()
	defer d.mu.Unlock()

	result := make([]HookExecution, len(d.executions))
	copy(result, d.executions)
	return result
}

// GetExecutionsByType 按类型过滤执行历史
func (d *DebugHook) GetExecutionsByType(hookType HookType) []HookExecution {
	d.mu.Lock()
	defer d.mu.Unlock()

	var result []HookExecution
	for _, exec := range d.executions {
		if exec.HookType == hookType {
			result = append(result, exec)
		}
	}
	return result
}

// Clear 清除历史记录
func (d *DebugHook) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.executions = d.executions[:0]
}

// Enable 启用调试
func (d *DebugHook) Enable() {
	d.enabled = true
}

// Disable 禁用调试
func (d *DebugHook) Disable() {
	d.enabled = false
}

// GetStats 获取统计信息
func (d *DebugHook) GetStats() map[string]interface{} {
	d.mu.Lock()
	defer d.mu.Unlock()

	stats := map[string]interface{}{
		"total_executions": len(d.executions),
		"enabled":          d.enabled,
		"max_history":      d.maxHistory,
	}

	// 按类型统计
	typeCounts := make(map[HookType]int)
	var totalDuration time.Duration

	for _, exec := range d.executions {
		typeCounts[exec.HookType]++
		totalDuration += exec.Duration
	}

	stats["by_type"] = typeCounts

	if len(d.executions) > 0 {
		stats["avg_duration_ms"] = totalDuration.Milliseconds() / int64(len(d.executions))
	}

	return stats
}
