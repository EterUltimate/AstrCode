# AstrCode

🚀 **AstrCode = AstrBot 的 Agent 编排引擎 + Skill 开发系统 + 执行优化层**

## 系统架构

```
User Task
    ↓
Agent（决策 + 编排）
    ↓
Skill Retriever（选技能）←── Embedding + 关键词
    ↓
Skill Planner（生成 Plan） ←── LLM + 依赖图验证
    ↓
Executor（执行 Plan） ←── JSON-RPC → AstrBot Runtime
    ↓
WebSocket 推送 → Dashboard 实时可视化
```

## 核心模块

| 模块 | 文件 | 说明 |
|------|------|------|
| SDK Client | `internal/sdk/client.go` | JSON-RPC 2.0，自动重试+指数退避 |
| WS Transport | `internal/sdk/transport.go` | WebSocket 心跳+自动重连 |
| Star Manager | `internal/skill/star_manager.go` | plugin.yaml 发现 |
| Skill Retriever | `internal/skill/retriever.go` | 向量检索 + 关键词回退 |
| Skill Planner | `internal/skill/planner.go` | LLM 生成 Plan + 循环依赖检测 |
| Adaptive Planner | `internal/skill/adaptive.go` | 失败重规划 + 降级策略 |
| Executor | `internal/skill/executor.go` | 依赖图执行 + 并行 + 重试 + 事件推送 |
| Agent Core | `internal/agent/agent.go` | 决策→检索→规划→执行 + 异步支持 |
| AstrBot Models | `internal/model/astrbot.go` | 完整数据模型（JSONRPC/Event/Star/Tool/Stream） |
| Plan Models | `internal/model/plan.go` | 增强 Plan（依赖图/条件/并行） |
| Task Store | `internal/model/taskstore.go` | 任务状态存储 + 时间线 + WebSocket 事件 |
| Prompt Engine | `internal/prompt/engine.go` | 三阶段 Prompt |
| Embedding | `internal/rag/embedding.go` | OpenAI 兼容 Embedding |
| Vector Index | `internal/rag/index.go` | Skill 向量索引 |
| Cache | `internal/cache/` | 三级缓存 + 持久化 + Redis |
| WS Hub | `internal/api/hub.go` | WebSocket 广播中心 |
| Dashboard | `web/index.html` | 实时执行可视化面板 |

## API Endpoints

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/task` | 提交任务（`{task, async}`） |
| GET | `/api/task/{id}` | 查询任务状态 |
| GET | `/api/tasks` | 列出所有任务 |
| GET | `/api/snapshot/{id}` | 执行快照（可视化数据） |
| GET | `/api/skills` | 获取技能列表 |
| POST | `/api/plan` | 仅生成计划 |
| POST | `/api/execute` | 直接执行 Handler |
| GET | `/ws` | WebSocket 实时事件流 |
| GET | `/health` | 健康检查 |

## WebSocket 事件类型

| 事件 | 说明 |
|------|------|
| `task_created` | 任务创建 |
| `task_completed` | 任务完成 |
| `task_failed` | 任务失败 |
| `step_started` | 步骤开始 |
| `step_completed` | 步骤完成 |
| `step_failed` | 步骤失败 |
| `step_retry` | 步骤重试 |

## 快速开始

```bash
go build -o bin/astrcode.exe cmd/server/main.go

./bin/astrcode.exe \
  -addr :8080 \
  -astrbot-url http://localhost:6185 \
  -llm-url http://localhost:11434 \
  -llm-model qwen2.5 \
  -static-dir ./web
```

打开 `http://localhost:8080` 查看 Dashboard。

## 开发阶段

- [x] **Phase 1** — SDK 接入（JSON-RPC）、Skill 调用、基础 Executor
- [x] **Phase 1.5** — WebSocket 传输、Star 发现、AstrBot 数据模型对齐
- [x] **Phase 2** — Skill Retriever（embedding）、Prompt Engine（三阶段）、Plan 优化
- [x] **Phase 3** — Cache 持久化、Adaptive 执行、错误恢复
- [x] **Phase 4** — 任务状态管理、WebSocket 实时推送、执行可视化、Dashboard UI、异步任务

## 许可证

MIT
