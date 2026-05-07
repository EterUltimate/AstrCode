package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// HookType 定义钩子类型
type HookType string

const (
	// Session 生命周期
	HookSessionStart HookType = "session_start"
	HookSessionEnd   HookType = "session_end"

	// Turn 生命周期
	HookTurnStart HookType = "turn_start"
	HookTurnEnd   HookType = "turn_end"

	// Tool 执行
	HookBeforeToolUse HookType = "before_tool_use"
	HookAfterToolUse  HookType = "after_tool_use"

	// LLM 调用
	HookBeforeLLMCall HookType = "before_llm_call"
	HookAfterLLMCall  HookType = "after_llm_call"

	// Prompt 组装
	HookPromptAssemble HookType = "prompt_assemble"
)

// HookMode 定义钩子执行模式
type HookMode string

const (
	// Blocking: 可阻断操作，必须等待完成
	HookModeBlocking HookMode = "blocking"

	// NonBlocking: 异步执行，不阻塞主流程
	HookModeNonBlocking HookMode = "non_blocking"

	// Advisory: 仅提供信息，不影响执行流
	HookModeAdvisory HookMode = "advisory"
)

// HookEvent 钩子事件数据结构
type HookEvent struct {
	Type      HookType          `json:"type"`
	SessionID string            `json:"session_id"`
	TurnID    string            `json:"turn_id,omitempty"`
	Data      interface{}       `json:"data,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// HookResult 钩子执行结果
type HookResult struct {
	Allowed  bool                   `json:"allowed"` // Blocking mode 下是否允许继续
	Error    error                  `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// HookHandler 钩子处理器函数签名
type HookHandler func(ctx context.Context, event HookEvent) HookResult

// RegisteredHook 已注册的钩子
type RegisteredHook struct {
	ID      string
	Name    string
	Mode    HookMode
	Handler HookHandler
	Timeout time.Duration
}

// HookRegistry 钩子注册表
type HookRegistry struct {
	mu    sync.RWMutex
	hooks map[HookType][]RegisteredHook
}

// NewHookRegistry 创建新的钩子注册表
func NewHookRegistry() *HookRegistry {
	return &HookRegistry{
		hooks: make(map[HookType][]RegisteredHook),
	}
}

// Register 注册钩子
func (r *HookRegistry) Register(hookType HookType, hook RegisteredHook) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if hook.Timeout == 0 {
		hook.Timeout = 5 * time.Second // 默认超时 5 秒
	}

	r.hooks[hookType] = append(r.hooks[hookType], hook)
}

// Unregister 注销钩子
func (r *HookRegistry) Unregister(hookType HookType, hookID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	hooks := r.hooks[hookType]
	for i, h := range hooks {
		if h.ID == hookID {
			r.hooks[hookType] = append(hooks[:i], hooks[i+1:]...)
			break
		}
	}
}

// Execute 执行指定类型的所有钩子
func (r *HookRegistry) Execute(ctx context.Context, hookType HookType, event HookEvent) []HookResult {
	r.mu.RLock()
	hooks := r.hooks[hookType]
	r.mu.RUnlock()

	if len(hooks) == 0 {
		return nil
	}

	results := make([]HookResult, 0, len(hooks))

	// 设置事件时间戳
	event.Timestamp = time.Now()

	for _, hook := range hooks {
		// NonBlocking 模式异步执行，不等待结果
		if hook.Mode == HookModeNonBlocking {
			go func(h RegisteredHook, e HookEvent) {
				r.executeHook(ctx, h, e)
			}(hook, event)
			// 添加一个占位结果
			results = append(results, HookResult{
				Allowed: true,
				Metadata: map[string]interface{}{
					"async": true,
				},
			})
			continue
		}

		// Blocking 和 Advisory 模式同步执行
		result := r.executeHook(ctx, hook, event)
		results = append(results, result)

		// Blocking mode 下，如果钩子拒绝，立即返回
		if hook.Mode == HookModeBlocking && !result.Allowed {
			return results
		}
	}

	return results
}

// executeHook 执行单个钩子（带超时控制）
func (r *HookRegistry) executeHook(ctx context.Context, hook RegisteredHook, event HookEvent) HookResult {
	resultChan := make(chan HookResult, 1)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				resultChan <- HookResult{
					Allowed: false,
					Error:   fmt.Errorf("hook panic: %v", err),
				}
			}
		}()

		result := hook.Handler(ctx, event)
		resultChan <- result
	}()

	select {
	case result := <-resultChan:
		return result
	case <-time.After(hook.Timeout):
		return HookResult{
			Allowed: false,
			Error:   fmt.Errorf("hook timeout after %v", hook.Timeout),
		}
	case <-ctx.Done():
		return HookResult{
			Allowed: false,
			Error:   ctx.Err(),
		}
	}
}

// HasHooks 检查是否有注册的钩子
func (r *HookRegistry) HasHooks(hookType HookType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.hooks[hookType]) > 0
}

// GetHookCount 获取指定类型的钩子数量
func (r *HookRegistry) GetHookCount(hookType HookType) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.hooks[hookType])
}

// ListHooks 列出所有已注册的钩子
func (r *HookRegistry) ListHooks() map[HookType][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[HookType][]string)
	for hookType, hooks := range r.hooks {
		names := make([]string, len(hooks))
		for i, h := range hooks {
			names[i] = fmt.Sprintf("%s (%s)", h.Name, h.Mode)
		}
		result[hookType] = names
	}

	return result
}

// MarshalJSON 实现 JSON 序列化（用于调试）
func (e HookEvent) MarshalJSON() ([]byte, error) {
	type Alias HookEvent
	dataBytes, err := json.Marshal(e.Data)
	if err != nil {
		dataBytes = []byte("{}")
	}

	return json.Marshal(&struct {
		Type      string            `json:"type"`
		SessionID string            `json:"session_id"`
		TurnID    string            `json:"turn_id,omitempty"`
		Data      json.RawMessage   `json:"data"`
		Metadata  map[string]string `json:"metadata,omitempty"`
		Timestamp time.Time         `json:"timestamp"`
		*Alias
	}{
		Type:      string(e.Type),
		SessionID: e.SessionID,
		TurnID:    e.TurnID,
		Data:      dataBytes,
		Metadata:  e.Metadata,
		Timestamp: e.Timestamp,
		Alias:     (*Alias)(&e),
	})
}
