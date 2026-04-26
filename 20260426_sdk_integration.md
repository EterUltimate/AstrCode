# AstrCode SDK & Skill 集成

## 时间
2026-04-26 12:10

## 目标
基于 astrbot-sdk (Python) 仓库的实际架构，将 AstrCode 的 SDK 层和 Skill 系统与 AstrBot 对齐。

## 分析结论

### astrbot-sdk 关键架构
1. **JSON-RPC 2.0 协议** — 所有通信通过 JSON-RPC，支持 HTTP 和 WebSocket 传输
2. **Star/Plugin 系统** — plugin.yaml 发现，Handler 注册表，支持 STDIO 和 WebSocket 两种运行模式
3. **Agent SDK** — Tool/Function Calling，支持 OpenAI/Anthropic/Google 三种 schema 转换
4. **流式支持** — handler_stream_start/update/end 通知
5. **请求管理** — RPCRequestHelper 管理单次请求（Future）和流式请求（Queue）

### AstrBot-Skill
- 前端项目（Vue.js + node_modules），不包含 Skill 定义内容，跳过。

### everything-claude-code
- 仓库为空，跳过。

## 完成的变更

### 新增/重写文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/model/astrbot.go` | **新增** | 完整对齐 astrbot-sdk 的数据模型：JSONRPC、Event、Star、Tool、Stream |
| `internal/sdk/client.go` | **重写** | JSON-RPC 协议客户端，支持 handshake/call_handler/skill.call/tool.call/tool.list |
| `internal/sdk/transport.go` | **新增** | WebSocket 传输层（对齐 websocket.py），心跳+自动重连+请求匹配 |
| `internal/skill/star_manager.go` | **新增** | Star 发现器（对齐 star_manager.py），plugin.yaml 解析，Handler 注册 |
| `internal/agent/agent.go` | **重写** | 集成 StarManager，新增 ProcessEvent/ListTools 接口 |
| `internal/skill/executor.go` | **重写** | 新增 ExecuteHandler 直接调用 AstrBot Handler |

### 数据模型对齐清单

| AstrBot Python 类 | Go 对应 | 文件 |
|---------------------|---------|------|
| `JSONRPCRequest` | `JSONRPCRequest` | model/astrbot.go |
| `JSONRPCSuccessResponse` | `JSONRPCSuccessResponse` | model/astrbot.go |
| `JSONRPCErrorResponse` | `JSONRPCErrorResponse` | model/astrbot.go |
| `AstrMessageEvent` | `AstrMessageEvent` | model/astrbot.go |
| `StarMetadata` | `StarMetadata` | model/astrbot.go |
| `FunctionTool` | `FunctionTool` | model/astrbot.go |
| `ToolSet` + openai_schema() | `ToolSet` + `ToOpenAISchema()` | model/astrbot.go |
| `Message` / `ContentPart` / `ToolCall` | 同名结构体 | model/astrbot.go |
| `HandlerStreamStart/Update/End` | 同名通知结构体 | model/astrbot.go |
| `WebSocketClient` | `WebSocketTransport` | sdk/transport.go |
| `StarManager.discover_star` | `StarManager.DiscoverStar` | skill/star_manager.go |

### 依赖新增
- `github.com/gorilla/websocket v1.5.3` — WebSocket 传输
- `gopkg.in/yaml.v3 v3.0.1` — plugin.yaml 解析

## 编译状态
✅ `go build ./...` 通过，无错误
