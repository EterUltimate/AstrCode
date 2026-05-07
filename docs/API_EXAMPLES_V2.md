# AstrCode v2.0 API 使用示例

本文档提供 AstrCode v2.0 新特性的完整 API 使用示例。

---

## 📋 目录

- [Hook 系统](#hook-系统)
- [Session 事件日志](#session-事件日志)
- [Context 自动压缩](#context-自动压缩)
- [运行模式切换](#运行模式切换)
- [Prompt 模块化组装](#prompt-模块化组装)
- [并行执行](#并行执行)

---

## 🎣 Hook 系统

### 1. 注册 Hook Handler

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/EterUltimate/astrcode/internal/hook"
)

// LoggingHook 记录所有工具调用
type LoggingHook struct{}

func (h *LoggingHook) Handle(ctx context.Context, event *hook.HookEvent) error {
    log.Printf("[HOOK] %s: %+v", event.Type, event.Data)
    return nil
}

// MetricsHook 收集性能指标
type MetricsHook struct{}

func (h *MetricsHook) Handle(ctx context.Context, event *hook.HookEvent) error {
    if event.Type == hook.AfterToolUse {
        duration := event.Data["duration"].(time.Duration)
        log.Printf("[METRICS] Tool execution took %v", duration)
    }
    return nil
}

func main() {
    // 创建 Hook 注册表
    registry := hook.NewHookRegistry()
    
    // 注册多个 handler 到同一 Hook 点
    registry.Register(hook.BeforeToolUse, &LoggingHook{})
    registry.Register(hook.AfterToolUse, &LoggingHook{})
    registry.Register(hook.AfterToolUse, &MetricsHook{})
    
    // 触发 Hook
    ctx := context.Background()
    event := &hook.HookEvent{
        Type: hook.BeforeToolUse,
        Data: map[string]interface{}{
            "tool":   "file_read",
            "params": map[string]string{"path": "/tmp/test.txt"},
        },
    }
    
    err := registry.Trigger(ctx, event)
    if err != nil {
        log.Printf("Hook execution failed: %v", err)
    }
}
```

### 2. 条件性 Hook 执行

```go
// ConditionalHook 根据条件决定是否执行
type ConditionalHook struct {
    enabled bool
}

func (h *ConditionalHook) Handle(ctx context.Context, event *hook.HookEvent) error {
    if !h.enabled {
        return nil // 跳过执行
    }
    
    // 只在调试模式下记录详细信息
    if debugMode {
        log.Printf("[DEBUG] Full event data: %+v", event)
    }
    
    return nil
}

// 运行时启用/禁用
hook := &ConditionalHook{enabled: false}
registry.Register(hook.BeforeLLMCall, hook)

// 稍后启用
hook.enabled = true
```

### 3. 错误处理

```go
// ValidationHook 验证工具参数
type ValidationHook struct{}

func (h *ValidationHook) Handle(ctx context.Context, event *hook.HookEvent) error {
    if event.Type == hook.BeforeToolUse {
        toolName := event.Data["tool"].(string)
        
        // 检查工具是否在白名单中
        if !isAllowed(toolName) {
            return fmt.Errorf("tool '%s' is not allowed", toolName)
        }
    }
    return nil
}

// Trigger 会返回第一个错误
err := registry.Trigger(ctx, event)
if err != nil {
    // Hook 阻止了执行
    log.Printf("Execution blocked: %v", err)
    return err
}
```

---

## 📊 Session 事件日志

### 1. 创建 Session Logger

```go
package main

import (
    "context"
    "log"
    
    "github.com/EterUltimate/astrcode/internal/session"
)

func main() {
    // 创建 logger，指定日志文件路径
    config := session.LoggerConfig{
        LogDir:       "./sessions",
        SnapshotInterval: 100,  // 每 100 事件创建快照
        MaxLogSize:   5 * 1024 * 1024, // 5MB
    }
    
    logger, err := session.NewLogger(config)
    if err != nil {
        log.Fatalf("Failed to create logger: %v", err)
    }
    defer logger.Close()
    
    // 记录会话开始
    ctx := context.Background()
    logger.LogEvent(ctx, "session_start", map[string]interface{}{
        "task":      "Create a weather plugin",
        "timestamp": time.Now(),
    })
    
    // 记录工具调用
    logger.LogEvent(ctx, "before_tool_use", map[string]interface{}{
        "tool": "file_write",
        "path": "/plugins/weather/main.py",
    })
    
    logger.LogEvent(ctx, "after_tool_use", map[string]interface{}{
        "status": "success",
        "duration_ms": 150,
    })
    
    // 记录 LLM 调用
    logger.LogEvent(ctx, "before_llm_call", map[string]interface{}{
        "prompt_tokens": 1234,
    })
    
    logger.LogEvent(ctx, "after_llm_call", map[string]interface{}{
        "response_tokens": 567,
        "model": "gpt-4o",
    })
    
    // 记录会话结束
    logger.LogEvent(ctx, "session_end", map[string]interface{}{
        "status": "completed",
        "total_events": logger.EventCount(),
    })
}
```

### 2. 从快照恢复

```go
// 加载最近的快照
snapshot, err := logger.LoadLatestSnapshot()
if err != nil {
    log.Printf("No snapshot found, starting fresh")
    return
}

log.Printf("Restored from snapshot: %d events", snapshot.EventCount)

// 重放后续事件
events, err := logger.ReadEventsAfter(snapshot.Timestamp)
if err != nil {
    log.Printf("Failed to read events: %v", err)
    return
}

for _, event := range events {
    log.Printf("Replaying: %s", event.Type)
    // 处理事件...
}
```

### 3. 查询事件日志

```go
// 查询特定类型的事件
events, err := logger.QueryEvents(session.Query{
    EventType: "tool_use",
    StartTime: time.Now().Add(-1 * time.Hour),
    EndTime:   time.Now(),
    Limit:     100,
})

if err != nil {
    log.Printf("Query failed: %v", err)
    return
}

for _, event := range events {
    log.Printf("[%s] %s: %+v", event.Timestamp, event.Type, event.Data)
}
```

---

## 🗜️ Context 自动压缩

### 1. 基本使用

```go
package main

import (
    "context"
    "log"
    
    "github.com/EterUltimate/astrcode/internal/context"
    "github.com/EterUltimate/astrcode/internal/model"
)

func main() {
    // 创建压缩器
    config := context.CompressorConfig{
        MaxTokens:     8000,   // GPT-4 上下文窗口
        KeepRecent:    10,     // 保留最近 10 条消息
        SummaryLength: 200,    // 每条摘要最大长度
    }
    
    compressor := context.NewCompressor(config)
    
    // 添加消息
    messages := []model.Message{
        {Role: "user", Content: "Hello"},
        {Role: "assistant", Content: "Hi! How can I help?"},
        // ... 更多消息
    }
    
    for _, msg := range messages {
        compressor.AddMessage(msg)
    }
    
    // 检查是否需要压缩
    if compressor.NeedsCompression() {
        log.Printf("Token count: %d, compressing...", compressor.TokenCount())
        
        // 执行压缩
        compressed, err := compressor.Compress(context.Background())
        if err != nil {
            log.Printf("Compression failed: %v", err)
            return
        }
        
        log.Printf("Compressed from %d to %d tokens", 
            compressor.TokenCount(), 
            compressed.TokenCount())
        
        // 使用压缩后的消息
        messages = compressed.Messages
    }
}
```

### 2. 自定义摘要生成器

```go
// CustomSummarizer 使用 LLM 生成更智能的摘要
type CustomSummarizer struct {
    llmClient *llm.Client
}

func (s *CustomSummarizer) Summarize(ctx context.Context, messages []model.Message) (string, error) {
    // 构建摘要请求
    prompt := fmt.Sprintf("Summarize the following conversation in %d words:\n\n%s",
        s.config.SummaryLength/5, // 约 5 tokens/word
        formatMessages(messages),
    )
    
    // 调用 LLM
    response, err := s.llmClient.Generate(ctx, prompt)
    if err != nil {
        return "", err
    }
    
    return response.Text, nil
}

// 使用自定义摘要生成器
compressor := context.NewCompressor(config)
compressor.SetSummarizer(&CustomSummarizer{llmClient: llmClient})
```

### 3. 监控和指标

```go
// 获取压缩统计
stats := compressor.Stats()
log.Printf("Compression stats:")
log.Printf("  Total messages: %d", stats.TotalMessages)
log.Printf("  Compressed messages: %d", stats.CompressedMessages)
log.Printf("  Original tokens: %d", stats.OriginalTokens)
log.Printf("  Compressed tokens: %d", stats.CompressedTokens)
log.Printf("  Savings: %.2f%%", stats.SavingsPercent())
```

---

## 🎮 运行模式切换

### 1. 创建 ModeController

```go
package main

import (
    "context"
    "log"
    
    "github.com/EterUltimate/astrcode/internal/mode"
    "github.com/EterUltimate/astrcode/internal/agent"
)

func main() {
    // 创建模式管理器
    manager := mode.NewManager(mode.DefaultConfig())
    
    // 创建 Agent 和 ModeController
    agent := agent.NewAgent(llmClient, sdkClient)
    modeCtrl := agent.NewModeController(manager)
    
    // 初始模式：Code Mode（全功能）
    modeCtrl.SwitchMode(mode.CodeMode)
    log.Printf("Current mode: %s", modeCtrl.CurrentMode())
}
```

### 2. 检查工具权限

```go
// 在执行工具前检查权限
func executeTool(ctx context.Context, toolName string) error {
    // 检查是否允许
    if !modeCtrl.IsToolAllowed(toolName) {
        return fmt.Errorf("tool '%s' is not allowed in %s mode", 
            toolName, modeCtrl.CurrentMode())
    }
    
    // 执行工具...
    return nil
}

// 示例：Plan Mode 下尝试写入文件
modeCtrl.SwitchMode(mode.PlanMode)

err := executeTool(ctx, "file_write")
if err != nil {
    log.Printf("Blocked: %v", err)
    // Output: Blocked: tool 'file_write' is not allowed in plan mode
}

// Code Mode 下允许
modeCtrl.SwitchMode(mode.CodeMode)
err = executeTool(ctx, "file_write")
if err != nil {
    log.Printf("Error: %v", err)
} else {
    log.Println("File write succeeded")
}
```

### 3. 自定义模式配置

```go
// 自定义模式配置
customConfig := mode.Config{
    AllowedTools: map[mode.Mode][]string{
        mode.CodeMode: {
            "file_read", "file_write", "code_execute", 
            "search", "web_scrape", "git_commit",
        },
        mode.PlanMode: {
            "file_read", "search", "code_analyze",
        },
        // 添加新模式
        "review": {
            "file_read", "code_analyze", "security_scan",
        },
    },
    MaxSteps: map[mode.Mode]int{
        mode.CodeMode: 50,
        mode.PlanMode: 20,
        "review":      30,
    },
}

manager := mode.NewManager(customConfig)
modeCtrl := agent.NewModeController(manager)

// 切换到自定义模式
modeCtrl.SwitchMode("review")
```

### 4. 模式切换事件

```go
// 监听模式切换
modeCtrl.OnModeChange(func(oldMode, newMode mode.Mode) {
    log.Printf("Mode changed: %s → %s", oldMode, newMode)
    
    // 发送通知到 Dashboard
    websocket.Broadcast(map[string]interface{}{
        "type": "mode_changed",
        "old_mode": oldMode,
        "new_mode": newMode,
        "timestamp": time.Now(),
    })
})

// 切换模式（会触发回调）
modeCtrl.SwitchMode(mode.PlanMode)
```

---

## 🧩 Prompt 模块化组装

### 1. 基本使用

```go
package main

import (
    "context"
    "log"
    
    "github.com/EterUltimate/astrcode/internal/prompt"
    "github.com/EterUltimate/astrcode/internal/model"
)

func main() {
    // 创建 Composer（带缓存）
    cache := prompt.NewPromptCache(5 * time.Minute)
    composer := prompt.NewComposer(cache)
    
    // 注册贡献者
    systemContributor := prompt.NewSystemPromptContributor(
        "You are an intelligent coding assistant.",
    )
    composer.RegisterContributor(systemContributor)
    
    taskContributor := prompt.NewTaskContributor(
        "Create a Python weather plugin for AstrBot",
    )
    composer.RegisterContributor(taskContributor)
    
    skills := []model.Skill{
        {
            Name:        "astrbot_plugin_template",
            Description: "Template for AstrBot plugins",
            Code:        "# Plugin template code...",
        },
        {
            Name:        "python_best_practices",
            Description: "Python coding best practices",
            Code:        "# Best practices...",
        },
    }
    skillsContributor := prompt.NewSkillsContributor(skills)
    composer.RegisterContributor(skillsContributor)
    
    // 组装 prompt
    ctx := context.Background()
    finalPrompt, err := composer.Assemble(ctx)
    if err != nil {
        log.Fatalf("Failed to assemble prompt: %v", err)
    }
    
    log.Printf("Final prompt length: %d characters", len(finalPrompt))
    log.Printf("Prompt:\n%s", finalPrompt)
}
```

### 2. 自定义贡献者

```go
// CustomContributor 自定义贡献者
type CustomContributor struct {
    name     string
    priority int
    content  string
}

func (c *CustomContributor) Name() string {
    return c.name
}

func (c *CustomContributor) Contribute(ctx context.Context) ([]prompt.PromptBlock, error) {
    return []prompt.PromptBlock{{
        ID:       "custom_block",
        Content:  c.content,
        Priority: c.priority,
    }}, nil
}

// 使用自定义贡献者
custom := &CustomContributor{
    name:     "custom_rules",
    priority: 500,
    content:  "Always use type hints in Python code.",
}
composer.RegisterContributor(custom)
```

### 3. 动态注册/注销

```go
// 注册贡献者
contributor := prompt.NewTaskContributor("New task")
composer.RegisterContributor(contributor)

// 组装 prompt
prompt1, _ := composer.Assemble(ctx)

// 注销贡献者
composer.UnregisterContributor(contributor.Name())

// 再次组装（不再包含该贡献者）
prompt2, _ := composer.Assemble(ctx)

log.Printf("Prompt 1 length: %d", len(prompt1))
log.Printf("Prompt 2 length: %d", len(prompt2))
// prompt2 会比 prompt1 短
```

### 4. 缓存管理

```go
// 手动清除缓存
cache.Clear()

// 清除特定 key 的缓存
cache.Delete(cacheKey)

// 获取缓存统计
stats := cache.Stats()
log.Printf("Cache hits: %d", stats.Hits)
log.Printf("Cache misses: %d", stats.Misses)
log.Printf("Hit rate: %.2f%%", stats.HitRate())
```

---

## ⚡ 并行执行

### 1. 基本使用

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/EterUltimate/astrcode/internal/pipeline"
)

func main() {
    // 创建执行器（最多 5 并发）
    executor := pipeline.NewExecutor(5)
    
    // 定义任务
    tasks := []pipeline.Task{
        {Name: "task1", Duration: 2 * time.Second},
        {Name: "task2", Duration: 2 * time.Second},
        {Name: "task3", Duration: 2 * time.Second},
        {Name: "task4", Duration: 2 * time.Second},
        {Name: "task5", Duration: 2 * time.Second},
        {Name: "task6", Duration: 2 * time.Second},
    }
    
    // 并行执行
    ctx := context.Background()
    start := time.Now()
    
    results, err := executor.ExecuteBatch(ctx, tasks)
    elapsed := time.Since(start)
    
    if err != nil {
        log.Printf("Some tasks failed: %v", err)
    }
    
    log.Printf("Executed %d tasks in %v", len(tasks), elapsed)
    // Output: Executed 6 tasks in ~4s (instead of 12s serial)
    
    for i, result := range results {
        log.Printf("Task %d: %v", i+1, result)
    }
}
```

### 2. 错误处理

```go
// 任务可能失败
tasks := []pipeline.Task{
    {Name: "success_task", Duration: 1 * time.Second},
    {Name: "failing_task", Duration: 1 * time.Second, ShouldFail: true},
    {Name: "another_success", Duration: 1 * time.Second},
}

results, err := executor.ExecuteBatch(ctx, tasks)

// err 包含所有失败的错误
if err != nil {
    log.Printf("Batch execution errors: %v", err)
}

// 检查结果
for i, result := range results {
    if result.Error != nil {
        log.Printf("Task %d failed: %v", i, result.Error)
    } else {
        log.Printf("Task %d succeeded", i)
    }
}
```

### 3. 取消支持

```go
// 创建可取消的 context
ctx, cancel := context.WithCancel(context.Background())

// 启动执行
go func() {
    results, err := executor.ExecuteBatch(ctx, tasks)
    log.Printf("Execution completed: %v, error: %v", results, err)
}()

// 5 秒后取消
time.Sleep(5 * time.Second)
cancel()

// 正在执行的任务会收到取消信号并停止
```

### 4. 进度监控

```go
// 带进度回调的执行
executor.WithProgress(func(progress pipeline.Progress) {
    log.Printf("Progress: %d/%d tasks completed (%.2f%%)",
        progress.Completed,
        progress.Total,
        progress.Percent(),
    )
})

results, err := executor.ExecuteBatch(ctx, tasks)
```

---

## 🔗 综合示例：完整工作流

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/EterUltimate/astrcode/internal/agent"
    "github.com/EterUltimate/astrcode/internal/hook"
    "github.com/EterUltimate/astrcode/internal/session"
    "github.com/EterUltimate/astrcode/internal/mode"
)

func main() {
    // 1. 初始化组件
    llmClient := llm.NewClient(config)
    sdkClient := sdk.NewClient(astrbotURL)
    
    // 2. 创建 Agent（自动初始化所有组件）
    agent := agent.NewAgent(llmClient, sdkClient)
    
    // 3. 注册 Hook
    agent.HookRegistry().Register(hook.BeforeToolUse, &LoggingHook{})
    agent.HookRegistry().Register(hook.AfterToolUse, &MetricsHook{})
    
    // 4. 创建 Session Logger
    logger, _ := session.NewLogger(session.LoggerConfig{
        LogDir: "./sessions",
    })
    defer logger.Close()
    
    // 5. 设置模式
    agent.ModeController().SwitchMode(mode.CodeMode)
    
    // 6. 执行任务
    ctx := context.Background()
    task := "Create a weather query plugin for AstrBot"
    
    // 记录会话开始
    logger.LogEvent(ctx, "session_start", map[string]interface{}{
        "task": task,
        "mode": agent.ModeController().CurrentMode(),
    })
    
    // 执行
    start := time.Now()
    err := agent.ExecuteTask(ctx, task)
    elapsed := time.Since(start)
    
    // 记录会话结束
    logger.LogEvent(ctx, "session_end", map[string]interface{}{
        "status":   "completed",
        "duration": elapsed,
    })
    
    if err != nil {
        log.Fatalf("Task failed: %v", err)
    }
    
    log.Printf("Task completed in %v", elapsed)
}
```

---

## 📚 相关文档

- [架构升级报告](./ARCHITECTURE_UPGRADE.md)
- [Hook 系统指南](./HOOK_SYSTEM_GUIDE.md)
- [运行模式切换指南](./MODE_SWITCHING_GUIDE.md)
- [Prompt 组装系统指南](./PROMPT_COMPOSER_GUIDE.md)
- [性能基准测试报告](./PERFORMANCE_BENCHMARK.md)

---

**文档版本**: v1.0  
**最后更新**: 2026-05-04  
**维护者**: AstrCode Team
