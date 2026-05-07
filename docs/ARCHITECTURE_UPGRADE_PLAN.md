# AstrCode 架构升级计划

## 📖 概述

本文档基于对 [astrcodey](https://github.com/whatevertogo/astrcodey)（Rust 实现的 AI 编码助手）的深入分析，提取其优秀架构设计并适配到当前 Go 版本的 AstrCode 项目中。

**核心目标**: 在保持 Go 语言实现的前提下，吸收 astrcodey 的架构精华，提升系统的可扩展性、可靠性和性能。

---

## 🎯 架构对比分析

### 当前 AstrCode (Go) vs AstrCodey (Rust)

| 维度 | AstrCode (当前) | AstrCodey (参考) | 差距分析 |
|------|----------------|------------------|----------|
| **状态管理** | 内存中 TaskStore | Session-first + JSONL event log | ❌ 缺少持久化事件溯源 |
| **Agent 模型** | 长期运行的 Agent 对象 | 临时处理器（从 session 重建） | ⚠️ 耦合度高，恢复困难 |
| **扩展系统** | 简单 Skill 加载 | Extension-first + Hook 系统 | ❌ 缺少生命周期干预能力 |
| **工具执行** | 顺序执行 | 并行批处理（最多5并发） | ⚠️ 性能瓶颈 |
| **上下文管理** | 无自动压缩 | Token budgeting + auto-compaction | ❌ 可能超出 LLM 限制 |
| **前后端协议** | REST + WebSocket | JSON-RPC 2.0 over stdio | ✅ 当前方案更通用 |
| **运行模式** | 单一模式 | Code/Plan mode 切换 | ❌ 缺少计划模式 |
| **Prompt 组装** | 简单模板 | Contributor 模式 + 四层缓存 | ⚠️ 灵活性不足 |

---

## 🚀 架构升级路线图

### Phase 1: 核心基础设施（高优先级）

#### 1.1 Hook 系统 🪝

**目标**: 提供插件化的生命周期干预能力

**设计**:
```go
// internal/hook/hook.go

type HookType string
const (
    // Session 生命周期
    HookSessionStart    HookType = "session_start"
    HookSessionEnd      HookType = "session_end"
    
    // Turn 生命周期
    HookTurnStart       HookType = "turn_start"
    HookTurnEnd         HookType = "turn_end"
    
    // Tool 执行
    HookBeforeToolUse   HookType = "before_tool_use"
    HookAfterToolUse    HookType = "after_tool_use"
    
    // LLM 调用
    HookBeforeLLMCall   HookType = "before_llm_call"
    HookAfterLLMCall    HookType = "after_llm_call"
    
    // Prompt 组装
    HookPromptAssemble  HookType = "prompt_assemble"
)

type HookMode string
const (
    HookModeBlocking    HookMode = "blocking"     // 可阻断操作
    HookModeNonBlocking HookMode = "non_blocking" // 异步执行，不阻塞
    HookModeAdvisory    HookMode = "advisory"     // 仅提供信息
)

type HookEvent struct {
    Type      HookType
    SessionID string
    TurnID    string
    Data      interface{}
}

type HookResult struct {
    Allowed bool        // Blocking mode 下是否允许继续
    Error   error       // 错误信息
    Metadata map[string]interface{} // 附加元数据
}

type HookHandler func(ctx context.Context, event HookEvent) HookResult

// HookRegistry 管理所有注册的 hook
type HookRegistry struct {
    hooks map[HookType][]RegisteredHook
}

type RegisteredHook struct {
    ID      string
    Mode    HookMode
    Handler HookHandler
    Timeout time.Duration
}
```

**集成点**:
- `internal/agent/agent.go`: Agent loop 中插入 hook 调用
- `internal/skill/executor.go`: Tool 执行前后触发 hook
- `internal/llm/client.go`: LLM 调用前后触发 hook

**实施步骤**:
1. 创建 `internal/hook` 包
2. 实现 HookRegistry 和事件分发机制
3. 在关键路径集成 hook 调用
4. 添加超时处理和错误恢复
5. 编写单元测试

**预期收益**:
- ✅ 支持安全策略检查（如禁止某些文件操作）
- ✅ 支持审计日志记录
- ✅ 支持自定义工具预处理/后处理
- ✅ 支持动态 prompt 注入

---

#### 1.2 Session 事件日志 📝

**目标**: 实现 append-only 的事件溯源，支持会话恢复和审计

**设计**:
```go
// internal/session/event_log.go

type EventType string
const (
    EventSessionStart     EventType = "session_start"
    EventSessionEnd       EventType = "session_end"
    EventUserMessage      EventType = "user_message"
    EventAssistantMessage EventType = "assistant_message"
    EventToolCall         EventType = "tool_call"
    EventToolResult       EventType = "tool_result"
    EventCompaction       EventType = "compaction"
    EventError            EventType = "error"
)

type Event struct {
    ID        string                 `json:"id"`
    Type      EventType              `json:"type"`
    Timestamp time.Time              `json:"timestamp"`
    SessionID string                 `json:"session_id"`
    TurnID    string                 `json:"turn_id,omitempty"`
    Data      json.RawMessage        `json:"data"`
    Metadata  map[string]string      `json:"metadata,omitempty"`
}

// EventLog append-only JSONL 日志
type EventLog struct {
    sessionID string
    filePath  string
    file      *os.File
    mu        sync.Mutex
}

func (l *EventLog) Append(event Event) error {
    l.mu.Lock()
    defer l.mu.Unlock()
    
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    _, err = l.file.Write(append(data, '\n'))
    if err != nil {
        return err
    }
    
    return l.file.Sync() // 确保落盘
}

func (l *EventLog) Replay(fromID string) ([]Event, error) {
    // 从指定事件 ID 开始回放
}
```

**存储结构**:
```
sessions/
├── <session_id>/
│   ├── events.jsonl      # Append-only event log
│   ├── snapshot.json     # 最新快照
│   └── metadata.json     # Session 元数据
```

**实施步骤**:
1. 创建 `internal/session` 包
2. 实现 EventLog（JSONL 写入）
3. 实现 EventReplayer（从日志重建状态）
4. 集成到 Agent loop
5. 添加文件锁防止并发写入冲突

**预期收益**:
- ✅ 支持会话中断后恢复
- ✅ 支持会话 fork/branch
- ✅ 完整的审计追踪
- ✅ 调试时可重放历史

---

#### 1.3 Context 自动压缩 🗜️

**目标**: 防止对话超出 LLM 上下文限制

**设计**:
```go
// internal/context/compactor.go

type Compactor struct {
    threshold float64  // 0.835 = 83.5%
    llmClient llm.Client
}

func (c *Compactor) ShouldCompact(currentTokens, maxTokens int) bool {
    ratio := float64(currentTokens) / float64(maxTokens)
    return ratio > c.threshold
}

func (c *Compactor) Compact(messages []Message) ([]Message, error) {
    // 策略 1: Provider-backed compaction (优先)
    summary, err := c.llmBackedCompact(messages)
    if err == nil {
        return summary, nil
    }
    
    // 策略 2: Deterministic compaction (fallback)
    return c.deterministicCompact(messages), nil
}

func (c *Compactor) llmBackedCompact(messages []Message) ([]Message, error) {
    // 使用 LLM 生成摘要
    prompt := "Summarize the following conversation..."
    // ... 调用 LLM
}

func (c *Compactor) deterministicCompact(messages []Message) []Message {
    // 规则基础压缩：
    // 1. 保留最近 N 条消息
    // 2. 删除中间的工具调用细节
    // 3. 保留关键决策点
}
```

**Token 估算**:
```go
// internal/context/token_counter.go

func EstimateTokens(text string) int {
    // 简化估算：1 token ≈ 4 characters (英文)
    // 或使用 tiktoken-go 进行精确计算
    return len(text) / 4
}
```

**实施步骤**:
1. 创建 `internal/context` 包
2. 实现 TokenCounter
3. 实现 Compactor（双层策略）
4. 在 Agent loop 中集成 threshold 检测
5. 保存 compact snapshot 用于调试

**预期收益**:
- ✅ 避免超出 LLM 上下文限制
- ✅ 减少 API 成本
- ✅ 保持对话连贯性

---

### Phase 2: 性能优化（中优先级）

#### 2.1 Tool Pipeline 并行化 ⚡

**目标**: 提升工具执行吞吐量

**当前问题**: 工具顺序执行，延迟累加

**改进方案**:
```go
// internal/agent/tool_pipeline.go

type ToolPipeline struct {
    maxConcurrency int  // 最大并发数，默认 5
}

func (p *ToolPipeline) ExecuteBatch(tools []ToolCall) []ToolResult {
    results := make([]ToolResult, len(tools))
    
    // 使用 semaphore 控制并发
    sem := make(chan struct{}, p.maxConcurrency)
    var wg sync.WaitGroup
    
    for i, tool := range tools {
        wg.Add(1)
        go func(idx int, t ToolCall) {
            defer wg.Done()
            
            sem <- struct{}{} // 获取令牌
            defer func() { <-sem }() // 释放令牌
            
            results[idx] = p.executeSingle(t)
        }(i, tool)
    }
    
    wg.Wait()
    return results
}
```

**实施步骤**:
1. 重构 `internal/agent/agent.go` 中的工具执行逻辑
2. 引入 `golang.org/x/sync/errgroup` 或自定义 semaphore
3. 添加结果排序保证（保持原始顺序）
4. 性能基准测试

**预期收益**:
- ✅ 工具执行速度提升 3-5x
- ✅ 更好的资源利用率

---

#### 2.2 Extension System 增强 🔌

**目标**: 支持更多类型的扩展注册

**当前局限**: 仅支持 Skill YAML 文件加载

**扩展能力**:
```go
// internal/extension/registry.go

type ExtensionCapability string
const (
    CapCustomTool      ExtensionCapability = "custom_tool"
    CapSlashCommand    ExtensionCapability = "slash_command"
    CapPromptContributor ExtensionCapability = "prompt_contributor"
    CapContextProvider ExtensionCapability = "context_provider"
    CapAgentProfile    ExtensionCapability = "agent_profile"
)

type Extension struct {
    ID           string
    Name         string
    Version      string
    Capabilities []ExtensionCapability
    Hooks        []HookRegistration
}

// 注册自定义工具
func (e *Extension) RegisterTool(tool ToolDefinition) error

// 注册 slash command
func (e *Extension) RegisterCommand(cmd SlashCommand) error
```

**目录结构**:
```
~/.astrcode/extensions/    # 全局扩展
  ├── github-tools/
  ├── docker-helper/
  └── ...

.astrcode/extensions/      # 项目级扩展（优先级更高）
  ├── project-specific-tool/
  └── ...
```

**实施步骤**:
1. 创建 `internal/extension` 包
2. 实现扩展加载器（扫描目录）
3. 实现能力注册表
4. 支持项目级扩展覆盖全局扩展
5. 添加扩展沙箱（可选，安全性）

**预期收益**:
- ✅ 更灵活的定制能力
- ✅ 社区生态潜力
- ✅ 团队级共享配置

---

### Phase 3: 高级特性（低优先级）

#### 3.1 运行模式切换 🔄

**目标**: 支持 Code Mode 和 Plan Mode

**设计**:
```go
// internal/agent/mode.go

type AgentMode string
const (
    ModeCode AgentMode = "code"  // 完整工具访问
    ModePlan AgentMode = "plan"  // 只读 + 计划管理
)

type ModeController struct {
    currentMode AgentMode
}

func (mc *ModeController) IsToolAllowed(toolName string) bool {
    if mc.currentMode == ModePlan {
        // Plan mode 只允许读取工具和计划管理工具
        allowedTools := []string{"read", "search", "plan_create", "plan_update"}
        return contains(allowedTools, toolName)
    }
    return true // Code mode 允许所有工具
}

// Exit Gate: Plan mode 退出前的自我审查
func (mc *ModeController) ValidatePlanExit(plan string) error {
    // 检查清单：
    // 1. 是否包含 ## Implementation Plan heading
    // 2. 是否有明确的步骤列表
    // 3. 是否考虑了边界情况
}
```

**UI 集成**:
- 前端添加模式切换按钮
- Plan mode 下显示计划编辑器
- Exit gate 验证失败时阻止切换

**实施步骤**:
1. 实现 ModeController
2. 在工具执行前检查模式权限
3. 实现 Exit Gate 验证逻辑
4. 前端 UI 集成
5. 计划 artifact 持久化 (`<session>/plan/plan.md`)

**预期收益**:
- ✅ 更安全的代码审查流程
- ✅ 强制规划思维
- ✅ 减少盲目执行

---

#### 3.2 Prompt Assembly System 📦

**目标**: 模块化的 prompt 组装

**设计**:
```go
// internal/prompt/composer.go

type PromptBlock struct {
    Content   string
    Priority  int           // 排序权重
    CacheLevel CacheLevel   // Stable/SemiStable/Inherited/Dynamic
    Condition  func(ctx) bool // 条件渲染
}

type PromptContributor interface {
    Contribute(ctx context.Context) ([]PromptBlock, error)
}

type PromptComposer struct {
    contributors []PromptContributor
    cache        PromptCache
}

func (pc *PromptComposer) Assemble(ctx context.Context) (string, error) {
    // 1. 收集所有 contributor 的 blocks
    // 2. 去重
    // 3. 按 priority 排序
    // 4. 解析模板变量
    // 5. 校验依赖
    // 6. 返回最终 prompt
}
```

**四层缓存**:
- **Stable**: 系统提示（极少变化）
- **SemiStable**: 项目配置（偶尔变化）
- **Inherited**: 父 session 继承
- **Dynamic**: 当前对话历史（每次变化）

**实施步骤**:
1. 重构 `internal/prompt/engine.go`
2. 实现 Contributor 接口
3. 实现四层缓存
4. 内置 contributor（system prompt, skills, context）
5. 支持扩展注册 custom contributor

**预期收益**:
- ✅ 更灵活的 prompt 定制
- ✅ 减少重复组装开销
- ✅ 清晰的依赖管理

---

## 📊 实施时间表

| Phase | 任务 | 预计工时 | 优先级 |
|-------|------|---------|--------|
| Phase 1 | Hook 系统 | 2-3 天 | 🔥 高 |
| Phase 1 | Session 事件日志 | 2-3 天 | 🔥 高 |
| Phase 1 | Context 自动压缩 | 1-2 天 | 🔥 高 |
| Phase 2 | Tool Pipeline 并行化 | 1-2 天 | 💡 中 |
| Phase 2 | Extension System 增强 | 3-4 天 | 💡 中 |
| Phase 3 | 运行模式切换 | 2-3 天 | 🌟 低 |
| Phase 3 | Prompt Assembly System | 2-3 天 | 🌟 低 |

**总计**: 13-20 天

---

## ⚠️ 风险与缓解

### 风险 1: 向后兼容性
**问题**: 新架构可能破坏现有 API
**缓解**:
- 保持现有 REST API 不变
- 新功能通过配置开关控制
- 提供迁移指南

### 风险 2: 性能开销
**问题**: Hook 系统和事件日志可能增加延迟
**缓解**:
- Hook 超时控制（默认 5s）
- 异步非阻塞 hook
- 批量写入事件日志
- 性能基准测试

### 风险 3: 复杂度增加
**问题**: 系统复杂度上升，维护成本增加
**缓解**:
- 清晰的模块边界
- 完善的文档和示例
- 渐进式采用（可选功能）

---

## 📚 参考资料

- [AstrCodey Repository](https://github.com/whatevertogo/astrcodey)
- [PROJECT_ARCHITECTURE.md](https://github.com/whatevertogo/astrcodey/blob/main/PROJECT_ARCHITECTURE.md)
- [Claude Code Architecture](https://docs.anthropic.com/en/docs/claude-code)
- [OpenCode Architecture](https://github.com/anomalyco/opencode)

---

## ✅ 下一步行动

1. **评审此文档**: 团队讨论确定优先级
2. **启动 Phase 1**: 从 Hook 系统开始实施
3. **每周进度同步**: 确保按计划推进
4. **性能监控**: 实施后对比关键指标

---

**最后更新**: 2026-05-04  
**作者**: AstrCode Team  
**版本**: v1.0
