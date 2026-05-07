package hook

import (
	"context"
	"testing"
	"time"
)

func TestHookRegistry_RegisterAndExecute(t *testing.T) {
	registry := NewHookRegistry()

	// 注册一个简单的钩子
	executed := false
	registry.Register(HookBeforeToolUse, RegisteredHook{
		ID:   "test-hook",
		Name: "Test Hook",
		Mode: HookModeBlocking,
		Handler: func(ctx context.Context, event HookEvent) HookResult {
			executed = true
			return HookResult{Allowed: true}
		},
		Timeout: 1 * time.Second,
	})

	// 执行钩子
	event := HookEvent{
		Type:      HookBeforeToolUse,
		SessionID: "test-session",
		TurnID:    "test-turn",
		Data:      map[string]interface{}{"tool_name": "read"},
	}

	results := registry.Execute(context.Background(), HookBeforeToolUse, event)

	if !executed {
		t.Error("Hook was not executed")
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if !results[0].Allowed {
		t.Error("Hook should allow execution")
	}
}

func TestHookRegistry_BlockingHookDeny(t *testing.T) {
	registry := NewHookRegistry()

	// 注册一个拒绝的钩子
	registry.Register(HookBeforeToolUse, RegisteredHook{
		ID:   "deny-hook",
		Name: "Deny Hook",
		Mode: HookModeBlocking,
		Handler: func(ctx context.Context, event HookEvent) HookResult {
			return HookResult{Allowed: false}
		},
		Timeout: 1 * time.Second,
	})

	// 再注册一个应该不会被执行的钩子
	secondExecuted := false
	registry.Register(HookBeforeToolUse, RegisteredHook{
		ID:   "second-hook",
		Name: "Second Hook",
		Mode: HookModeBlocking,
		Handler: func(ctx context.Context, event HookEvent) HookResult {
			secondExecuted = true
			return HookResult{Allowed: true}
		},
		Timeout: 1 * time.Second,
	})

	event := HookEvent{
		Type:      HookBeforeToolUse,
		SessionID: "test-session",
	}

	results := registry.Execute(context.Background(), HookBeforeToolUse, event)

	if len(results) != 1 {
		t.Errorf("Expected 1 result (short-circuit), got %d", len(results))
	}

	if secondExecuted {
		t.Error("Second hook should not be executed after blocking deny")
	}
}

func TestHookRegistry_Timeout(t *testing.T) {
	registry := NewHookRegistry()

	// 注册一个超时的钩子
	registry.Register(HookBeforeToolUse, RegisteredHook{
		ID:      "slow-hook",
		Name:    "Slow Hook",
		Mode:    HookModeBlocking,
		Timeout: 100 * time.Millisecond,
		Handler: func(ctx context.Context, event HookEvent) HookResult {
			time.Sleep(200 * time.Millisecond) // 故意超时
			return HookResult{Allowed: true}
		},
	})

	event := HookEvent{
		Type:      HookBeforeToolUse,
		SessionID: "test-session",
	}

	start := time.Now()
	results := registry.Execute(context.Background(), HookBeforeToolUse, event)
	duration := time.Since(start)

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0].Error == nil {
		t.Error("Expected timeout error")
	}

	// 应该在 100ms 左右超时，而不是等待 200ms
	if duration > 150*time.Millisecond {
		t.Errorf("Hook should timeout quickly, but took %v", duration)
	}
}

func TestHookRegistry_NonBlocking(t *testing.T) {
	registry := NewHookRegistry()

	executed := false
	registry.Register(HookBeforeToolUse, RegisteredHook{
		ID:   "async-hook",
		Name: "Async Hook",
		Mode: HookModeNonBlocking,
		Handler: func(ctx context.Context, event HookEvent) HookResult {
			time.Sleep(50 * time.Millisecond)
			executed = true
			return HookResult{Allowed: true}
		},
		Timeout: 1 * time.Second,
	})

	event := HookEvent{
		Type: HookBeforeToolUse,
	}

	start := time.Now()
	results := registry.Execute(context.Background(), HookBeforeToolUse, event)
	duration := time.Since(start)

	// NonBlocking 应该立即返回
	if duration > 10*time.Millisecond {
		t.Errorf("NonBlocking hook should return immediately, but took %v", duration)
	}

	// 等待异步执行完成
	time.Sleep(100 * time.Millisecond)

	if !executed {
		t.Error("Async hook should have been executed")
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestHookRegistry_Unregister(t *testing.T) {
	registry := NewHookRegistry()

	registry.Register(HookBeforeToolUse, RegisteredHook{
		ID:   "hook-1",
		Name: "Hook 1",
		Mode: HookModeAdvisory,
		Handler: func(ctx context.Context, event HookEvent) HookResult {
			return HookResult{Allowed: true}
		},
	})

	registry.Register(HookBeforeToolUse, RegisteredHook{
		ID:   "hook-2",
		Name: "Hook 2",
		Mode: HookModeAdvisory,
		Handler: func(ctx context.Context, event HookEvent) HookResult {
			return HookResult{Allowed: true}
		},
	})

	if registry.GetHookCount(HookBeforeToolUse) != 2 {
		t.Error("Should have 2 hooks")
	}

	// 注销一个
	registry.Unregister(HookBeforeToolUse, "hook-1")

	if registry.GetHookCount(HookBeforeToolUse) != 1 {
		t.Error("Should have 1 hook after unregister")
	}
}

func TestHookRegistry_HasHooks(t *testing.T) {
	registry := NewHookRegistry()

	if registry.HasHooks(HookBeforeToolUse) {
		t.Error("Should not have hooks initially")
	}

	registry.Register(HookBeforeToolUse, RegisteredHook{
		ID:   "test",
		Name: "Test",
		Mode: HookModeAdvisory,
		Handler: func(ctx context.Context, event HookEvent) HookResult {
			return HookResult{Allowed: true}
		},
	})

	if !registry.HasHooks(HookBeforeToolUse) {
		t.Error("Should have hooks after registration")
	}
}

func TestHookRegistry_ListHooks(t *testing.T) {
	registry := NewHookRegistry()

	registry.Register(HookBeforeToolUse, RegisteredHook{
		ID:   "hook-1",
		Name: "Hook 1",
		Mode: HookModeBlocking,
		Handler: func(ctx context.Context, event HookEvent) HookResult {
			return HookResult{Allowed: true}
		},
	})

	registry.Register(HookAfterToolUse, RegisteredHook{
		ID:   "hook-2",
		Name: "Hook 2",
		Mode: HookModeAdvisory,
		Handler: func(ctx context.Context, event HookEvent) HookResult {
			return HookResult{Allowed: true}
		},
	})

	hooks := registry.ListHooks()

	if len(hooks) != 2 {
		t.Errorf("Expected 2 hook types, got %d", len(hooks))
	}

	if len(hooks[HookBeforeToolUse]) != 1 {
		t.Error("Should have 1 BeforeToolUse hook")
	}

	if len(hooks[HookAfterToolUse]) != 1 {
		t.Error("Should have 1 AfterToolUse hook")
	}
}

func TestHookRegistry_PanicRecovery(t *testing.T) {
	registry := NewHookRegistry()

	registry.Register(HookBeforeToolUse, RegisteredHook{
		ID:   "panic-hook",
		Name: "Panic Hook",
		Mode: HookModeBlocking,
		Handler: func(ctx context.Context, event HookEvent) HookResult {
			panic("intentional panic")
		},
		Timeout: 1 * time.Second,
	})

	event := HookEvent{
		Type: HookBeforeToolUse,
	}

	results := registry.Execute(context.Background(), HookBeforeToolUse, event)

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0].Error == nil {
		t.Error("Should catch panic as error")
	}

	if results[0].Allowed {
		t.Error("Panic should result in not allowed")
	}
}

func TestRegisterDefaultHooks(t *testing.T) {
	registry := NewHookRegistry()
	RegisterDefaultHooks(registry)

	// 检查是否注册了钩子
	if !registry.HasHooks(HookSessionStart) {
		t.Error("Should have SessionStart hooks")
	}

	if !registry.HasHooks(HookBeforeToolUse) {
		t.Error("Should have BeforeToolUse hooks")
	}

	// 应该有多个钩子类型
	hooks := registry.ListHooks()
	if len(hooks) < 5 {
		t.Errorf("Expected at least 5 hook types, got %d", len(hooks))
	}
}

func TestHookEvent_MarshalJSON(t *testing.T) {
	event := HookEvent{
		Type:      HookBeforeToolUse,
		SessionID: "test-session",
		TurnID:    "test-turn",
		Data:      map[string]interface{}{"key": "value"},
		Metadata:  map[string]string{"meta": "data"},
		Timestamp: time.Now(),
	}

	data, err := event.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Marshaled data should not be empty")
	}
}
