# Hook 系统使用指南

## 📖 概述

Hook 系统是 AstrCode 的插件化生命周期干预机制，允许在关键执行点插入自定义逻辑，实现安全检查、审计日志、性能监控等功能。

**核心特性**:
- ✅ 8 种钩子类型覆盖完整生命周期
- ✅ 3 种执行模式（Blocking/NonBlocking/Advisory）
- ✅ 超时控制和 panic 恢复
- ✅ 线程安全
- ✅ 灵活的注册/注销机制

---

## 🏗️ 架构设计

### Hook 类型

```go
// Session 生命周期
HookSessionStart    // 会话开始
HookSessionEnd      // 会话结束

// Turn 生命周期
HookTurnStart       // 轮次开始
HookTurnEnd         // 轮次结束

// Tool 执行
HookBeforeToolUse   // 工具使用前
HookAfterToolUse    // 工具使用后

// LLM 调用
HookBeforeLLMCall   // LLM 调用前
HookAfterLLMCall    // LLM 调用后

// Prompt 组装
HookPromptAssemble  // Prompt 组装时
```

### Hook 模式

| 模式 | 说明 | 使用场景 |
|------|------|----------|
| **Blocking** | 同步执行，可阻断操作 | 安全检查、权限验证 |
| **NonBlocking** | 异步执行，不阻塞主流程 | 审计日志、指标收集 |
| **Advisory** | 仅提供信息，不影响执行 | 日志记录、调试 |

---

## 🚀 快速开始

### 1. 创建 HookRegistry

```go
import "github.com/EterUltimate/astrcode/internal/hook"

registry := hook.NewHookRegistry()
```

### 2. 注册钩子

```go
registry.Register(hook.HookBeforeToolUse, hook.RegisteredHook{
    ID:      "security-check",
    Name:    "Security Policy Check",
    Mode:    hook.HookModeBlocking,
    Timeout: 3 * time.Second,
    Handler: func(ctx context.Context, event hook.HookEvent) hook.HookResult {
        // 你的逻辑
        return hook.HookResult{Allowed: true}
    },
})
```

### 3. 执行钩子

```go
event := hook.HookEvent{
    Type:      hook.HookBeforeToolUse,
    SessionID: "session-123",
    TurnID:    "turn-456",
    Data:      map[string]interface{}{"tool_name": "read"},
}

results := registry.Execute(context.Background(), hook.HookBeforeToolUse, event)

// 检查结果
for _, result := range results {
    if !result.Allowed {
        log.Printf("Hook denied: %v", result.Error)
        return
    }
}
```

---

## 📝 实用示例

### 示例 1: 安全检查（Blocking）

阻止未授权的工具执行：

```go
func SecurityCheckHook(allowedTools []string) hook.HookHandler {
    return func(ctx context.Context, event hook.HookEvent) hook.HookResult {
        data, ok := event.Data.(map[string]interface{})
        if !ok {
            return hook.HookResult{Allowed: false, Error: fmt.Errorf("invalid data")}
        }

        toolName := data["tool_name"].(string)
        
        for _, allowed := range allowedTools {
            if toolName == allowed {
                return hook.HookResult{Allowed: true}
            }
        }

        return hook.HookResult{
            Allowed: false,
            Error:   fmt.Errorf("tool '%s' is not allowed", toolName),
        }
    }
}

// 注册
registry.Register(hook.HookBeforeToolUse, hook.RegisteredHook{
    ID:      "security",
    Name:    "Security Check",
    Mode:    hook.HookModeBlocking,
    Handler: SecurityCheckHook([]string{"read", "search"}),
})
```

### 示例 2: 审计日志（NonBlocking）

异步记录所有操作：

```go
func AuditLogHook(auditLogger func(event hook.HookEvent)) hook.HookHandler {
    return func(ctx context.Context, event hook.HookEvent) hook.HookResult {
        go func() {
            auditLogger(event) // 异步记录
        }()
        
        return hook.HookResult{Allowed: true}
    }
}

// 注册
registry.Register(hook.HookBeforeToolUse, hook.RegisteredHook{
    ID:      "audit",
    Name:    "Audit Logger",
    Mode:    hook.HookModeNonBlocking,
    Handler: AuditLogHook(func(event hook.HookEvent) {
        log.Printf("[AUDIT] %s by session %s", event.Type, event.SessionID)
    }),
})
```

### 示例 3: 性能监控

测量钩子执行时间：

```go
registry.Register(hook.HookBeforeLLMCall, hook.RegisteredHook{
    ID:      "perf-monitor",
    Name:    "Performance Monitor",
    Mode:    hook.HookModeAdvisory,
    Handler: func(ctx context.Context, event hook.HookEvent) hook.HookResult {
        startTime := time.Now()
        
        // 在 AfterLLMCall 中计算耗时
        return hook.HookResult{
            Allowed: true,
            Metadata: map[string]interface{}{
                "start_time": startTime,
            },
        }
    },
})
```

### 示例 4: 动态 Prompt 注入

在 Prompt 组装时添加自定义内容：

```go
registry.Register(hook.HookPromptAssemble, hook.RegisteredHook{
    ID:      "custom-instruction",
    Name:    "Custom Instruction Injector",
    Mode:    hook.HookModeBlocking,
    Handler: func(ctx context.Context, event hook.HookEvent) hook.HookResult {
        // 从配置读取自定义指令
        customInstruction := loadCustomInstruction()
        
        return hook.HookResult{
            Allowed: true,
            Metadata: map[string]interface{}{
                "additional_prompt": customInstruction,
            },
        }
    },
})
```

---

## 🔧 高级用法

### 批量注册默认钩子

```go
hook.RegisterDefaultHooks(registry)
```

这会注册：
- 日志钩子（所有类型，Advisory）
- 安全检查钩子（BeforeToolUse，Blocking）
- 审计日志钩子（所有类型，NonBlocking）

### 注销钩子

```go
registry.Unregister(hook.HookBeforeToolUse, "security-check")
```

### 查询钩子状态

```go
// 检查是否有钩子
if registry.HasHooks(hook.HookBeforeToolUse) {
    log.Println("有注册的 BeforeToolUse 钩子")
}

// 获取钩子数量
count := registry.GetHookCount(hook.HookBeforeToolUse)

// 列出所有钩子
allHooks := registry.ListHooks()
for hookType, names := range allHooks {
    log.Printf("%s: %v", hookType, names)
}
```

---

## ⚠️ 注意事项

### 1. 超时控制

所有钩子都有超时保护（默认 5 秒），避免阻塞主流程：

```go
RegisteredHook{
    Timeout: 3 * time.Second, // 自定义超时
    // ...
}
```

### 2. Panic 恢复

钩子执行中的 panic 会被捕获并转换为错误：

```go
// 即使 handler panic，也不会导致整个系统崩溃
results := registry.Execute(ctx, hookType, event)
// results[0].Error 会包含 panic 信息
```

### 3. 线程安全

HookRegistry 是线程安全的，可以在多个 goroutine 中并发使用：

```go
// 安全的并发执行
go func() {
    registry.Execute(ctx, hook.HookBeforeToolUse, event1)
}()

go func() {
    registry.Execute(ctx, hook.HookBeforeToolUse, event2)
}()
```

### 4. Blocking 模式的短路行为

当 Blocking 钩子拒绝时，后续钩子不会执行：

```go
// 如果第一个 Blocking 钩子返回 Allowed: false
// 第二个钩子不会被执行
results := registry.Execute(ctx, hookType, event)
// len(results) == 1
```

---

## 🧪 测试

运行单元测试：

```bash
go test -v ./internal/hook/...
```

测试覆盖：
- ✅ 基本注册和执行
- ✅ Blocking 模式拒绝
- ✅ 超时控制
- ✅ NonBlocking 异步执行
- ✅ 注销钩子
- ✅ Panic 恢复
- ✅ JSON 序列化

---

## 📊 性能考虑

| 操作 | 耗时 | 说明 |
|------|------|------|
| 注册钩子 | ~1μs | O(1) 追加到切片 |
| 执行 Advisory 钩子 | 取决于 handler | 同步执行 |
| 执行 Blocking 钩子 | 取决于 handler | 同步执行，可短路 |
| 执行 NonBlocking 钩子 | <1μs | 立即返回，异步执行 |
| 超时检测 | 精确到 ms | 基于 time.After |

**建议**:
- 对于日志、监控等非关键逻辑，使用 `NonBlocking` 模式
- 对于安全检查等关键逻辑，使用 `Blocking` 模式
- 保持钩子 handler 简洁，避免复杂计算

---

## 🔮 未来扩展

计划支持的钩子类型：
- `HookToolStreamStart` / `HookToolStreamEnd` - 流式工具执行
- `HookMessageDelta` - 消息增量输出
- `HookContextCompact` - 上下文压缩前后
- `HookExtensionLoad` / `HookExtensionUnload` - 扩展加载生命周期

---

## 📚 相关文档

- [架构升级计划](./ARCHITECTURE_UPGRADE_PLAN.md)
- [Hook 系统 API 参考](../internal/hook/hook.go)
- [示例代码](../internal/hook/examples.go)

---

**最后更新**: 2026-05-04  
**版本**: v1.0
