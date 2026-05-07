package hook

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ExampleHookHandler 示例钩子处理器：记录日志
func LoggingHookHandler(hookName string) HookHandler {
	return func(ctx context.Context, event HookEvent) HookResult {
		log.Printf("[HOOK:%s] %s - Session: %s, Turn: %s",
			hookName, event.Type, event.SessionID, event.TurnID)

		return HookResult{
			Allowed: true,
			Metadata: map[string]interface{}{
				"logged": true,
			},
		}
	}
}

// ExampleHookHandler 示例钩子处理器：安全检查（Blocking）
func SecurityCheckHook(allowedTools []string) HookHandler {
	return func(ctx context.Context, event HookEvent) HookResult {
		if event.Type != HookBeforeToolUse {
			return HookResult{Allowed: true}
		}

		// 从事件数据中提取工具名称
		data, ok := event.Data.(map[string]interface{})
		if !ok {
			return HookResult{
				Allowed: false,
				Error:   fmt.Errorf("invalid event data"),
			}
		}

		toolName, ok := data["tool_name"].(string)
		if !ok {
			return HookResult{
				Allowed: false,
				Error:   fmt.Errorf("missing tool_name in event data"),
			}
		}

		// 检查工具是否在白名单中
		for _, allowed := range allowedTools {
			if toolName == allowed {
				return HookResult{
					Allowed: true,
					Metadata: map[string]interface{}{
						"security_check": "passed",
					},
				}
			}
		}

		return HookResult{
			Allowed: false,
			Error:   fmt.Errorf("tool '%s' is not allowed by security policy", toolName),
		}
	}
}

// ExampleHookHandler 示例钩子处理器：审计日志
func AuditLogHook(auditLogger func(event HookEvent)) HookHandler {
	return func(ctx context.Context, event HookEvent) HookResult {
		// 异步记录审计日志
		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("[AUDIT] panic while logging: %v", err)
				}
			}()

			if auditLogger != nil {
				auditLogger(event)
			}
		}()

		return HookResult{
			Allowed: true,
			Metadata: map[string]interface{}{
				"audited": true,
			},
		}
	}
}

// ExampleHookHandler 示例钩子处理器：性能监控
func PerformanceMonitorHook(monitor func(hookType HookType, duration int64)) HookHandler {
	return func(ctx context.Context, event HookEvent) HookResult {
		startTime := event.Timestamp.UnixNano()

		// 在钩子执行完成后记录耗时
		return HookResult{
			Allowed: true,
			Metadata: map[string]interface{}{
				"start_time": startTime,
				"monitor":    true,
			},
		}
	}
}

// RegisterDefaultHooks 注册默认钩子（用于测试和演示）
func RegisterDefaultHooks(registry *HookRegistry) {
	// 1. 日志钩子（所有类型，Advisory）
	hookTypes := []HookType{
		HookSessionStart,
		HookSessionEnd,
		HookTurnStart,
		HookTurnEnd,
		HookBeforeToolUse,
		HookAfterToolUse,
		HookBeforeLLMCall,
		HookAfterLLMCall,
	}

	for _, ht := range hookTypes {
		registry.Register(ht, RegisteredHook{
			ID:      "logger",
			Name:    "Default Logger",
			Mode:    HookModeAdvisory,
			Handler: LoggingHookHandler("logger"),
			Timeout: 2 * time.Second,
		})
	}

	// 2. 安全策略钩子（工具使用前，Blocking）
	allowedTools := []string{"read", "search", "write", "edit"}
	registry.Register(HookBeforeToolUse, RegisteredHook{
		ID:      "security-check",
		Name:    "Security Policy Check",
		Mode:    HookModeBlocking,
		Handler: SecurityCheckHook(allowedTools),
		Timeout: 3 * time.Second,
	})

	// 3. 审计日志钩子（所有类型，NonBlocking）
	auditLogger := func(event HookEvent) {
		log.Printf("[AUDIT] %s - Session: %s", event.Type, event.SessionID)
	}

	for _, ht := range hookTypes {
		registry.Register(ht, RegisteredHook{
			ID:      "audit",
			Name:    "Audit Logger",
			Mode:    HookModeNonBlocking,
			Handler: AuditLogHook(auditLogger),
			Timeout: 5 * time.Second,
		})
	}
}
