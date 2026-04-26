# AstrCode Phase 2 完成 — Skill Retriever + Prompt Engine + Plan 优化

## 时间
2026-04-26 12:33

## 目标
完成 Phase 2 开发：Skill Retriever（Embedding + RAG）、Prompt Engine（三阶段）、Plan 结构优化（依赖图+并行+重试）。

## 新增/重写文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/rag/embedding.go` | **新增** | OpenAI 兼容 Embedding API 客户端，支持单条/批量 |
| `internal/rag/index.go` | **新增** | Skill 向量索引 — IndexSkill/IndexSkills/Search |
| `internal/model/plan.go` | **新增** | 增强 Plan 模型 — StepType、StepStatus、Condition、PlanGraph |
| `internal/skill/retriever.go` | **重写** | 向量检索 + 关键词回退，RegisterAndIndex 接口 |
| `internal/skill/planner.go` | **重写** | 结构化 Plan 生成 + 循环依赖检测 + 计划验证 |
| `internal/skill/executor.go` | **重写** | 依赖图执行 + 并行控制 + 重试机制 |
| `internal/prompt/engine.go` | **重写** | 三阶段 Prompt：Decision/Plan/Call |
| `internal/agent/agent.go` | **重写** | 集成三阶段流程：Decision → Retrieval → Plan → Execute |
| `internal/model/skill.go` | **简化** | 移除 Plan/Step 定义（迁移到 plan.go） |
| `README.md` | **更新** | 完整 Phase 2 架构文档 |

## 核心能力

### 1. Skill Retriever（Embedding + RAG）
- **EmbeddingClient** — 调用 OpenAI 兼容 `/embeddings` API
- **SkillIndex** — 自动索引 Skill（name + description + summary）
- **双模式检索** — 优先向量检索，失败回退到关键词匹配
- **批量索引** — `IndexSkills` 批量处理

### 2. Prompt Engine（三阶段）
- **Phase 1: Decision** — 判断 action（plan/direct/clarify），输出置信度
- **Phase 2: Plan** — 基于决策生成结构化执行计划
- **Phase 3: Call/Parse** — Skill 调用 + 结果解析

### 3. Plan 结构优化
- **依赖图** — `PlanGraph` 支持 `GetReadySteps()` / `IsComplete()` / `HasFailed()`
- **并行执行** — `parallel` + `max_parallel` 控制并发
- **重试机制** — `max_retries` + 指数退避
- **条件分支** — `Condition` 支持 then/else 分支
- **验证** — 循环依赖检测（DFS）、重复 ID 检测、依赖存在性检查

### 4. Executor 增强
- **依赖图驱动** — 自动调度就绪步骤
- **并行控制** — semaphore 限制并发数
- **重试** — 每个步骤独立重试，带超时
- **状态追踪** — StepStatus 实时更新

## 编译状态
✅ `go build ./...` 通过
✅ `go test ./...` 通过

## 下一步（Phase 3）
- Cache 持久化（磁盘/Redis）
- Adaptive 执行（动态调整 plan）
- 错误恢复（失败步骤替代方案）
