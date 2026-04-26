# AstrCode

🚀 **AstrCode - AstrBot 智能开发助手 & 插件编排引擎**

基于 **AstrBot SDK** 和 **AstrBot-Skill**，通过自然语言交互实现 AstrBot 插件开发、核心代码贡献、实时部署与审查的智能化开发平台。

### 核心价值

- 💬 **自然语言编程** - 用对话方式开发 AstrBot 插件，降低开发门槛
- 🔍 **智能代码审查** - AI 辅助审查插件代码质量和安全性
- 🚀 **一键部署** - 自动化测试、构建、部署到 AstrBot 运行时
- 🎨 **类 CodeX 界面** - 现代化的开发者体验，类似 Cursor/Copilot Chat

## 📋 目录

- [系统架构](#系统架构)
- [核心功能](#核心功能)
- [快速开始](#快速开始)
- [API 参考](#api-参考)
- [WebSocket 事件](#websocket-事件)
- [项目结构](#项目结构)
- [开发阶段](#开发阶段)
- [构建与部署](#构建与部署)
- [技术栈](#技术栈)
- [许可证](#许可证)

---

## 系统架构

```
开发者 (自然语言)
    ↓
AstrCode Agent (理解意图 + 生成代码)
    ↓
┌─────────────────────────────────────┐
│  AstrBot SDK Integration            │
│  - JSON-RPC 2.0 Protocol           │
│  - WebSocket Transport              │
│  - Tool/Function Calling            │
└─────────────────────────────────────┘
    ↓
┌─────────────────────────────────────┐
│  AstrBot-Skill System               │
│  - Plugin Discovery                 │
│  - Handler Registration             │
│  - Skill Retrieval (RAG)            │
└─────────────────────────────────────┘
    ↓
Code Generation & Review
    ↓
Automated Testing & Deployment
    ↓
AstrBot Runtime (Live Preview)
    ↓
WebSocket → CodeX-like Dashboard
```

### 数据流

1. **开发者输入自然语言** → "创建一个天气查询插件"
2. **Agent 理解意图** → LLM 分析需求，检索相关 Skill
3. **代码生成** → 基于 AstrBot SDK 模板生成插件代码
4. **智能审查** → AI 检查代码质量、安全性、最佳实践
5. **自动测试** → 运行单元测试、集成测试
6. **实时部署** → 通过 JSON-RPC 热加载到 AstrBot Runtime
7. **即时预览** → WebSocket 推送执行结果到 Dashboard
8. **迭代优化** → 根据反馈自动修正代码

---

## 核心功能

### 💬 自然语言插件开发

- **意图识别** - 理解开发者需求，自动选择合适模板
- **代码生成** - 基于 AstrBot SDK 生成完整插件结构
- **智能补全** - 上下文感知的代码建议和修复
- **多轮对话** - 支持迭代式开发和需求细化

### 🔍 智能代码审查

- **质量检查** - 代码风格、性能优化建议
- **安全扫描** - 检测潜在漏洞和不安全实践
- **最佳实践** - 推荐 AstrBot 官方推荐模式
- **依赖分析** - 检查第三方库兼容性和安全性

### 🚀 自动化部署流水线

- **一键测试** - 自动运行单元/集成测试
- **热重载** - 通过 JSON-RPC 实时更新插件
- **版本管理** - 自动语义化版本控制和 changelog
- **回滚支持** - 快速恢复到稳定版本

### 🎨 CodeX-like 开发界面

- **分屏布局** - 左侧对话，右侧代码编辑器
- **实时预览** - 插件执行结果即时展示
- **差异对比** - 代码修改前后对比视图
- **历史记录** - 完整的开发会话追踪

---

## 快速开始

### 前置要求

- Go 1.21+
- **AstrBot Runtime**（运行在 `http://localhost:6185`）- [安装指南](https://github.com/AstrBotDevs/AstrBot)
- **LLM 服务**（Ollama/OpenAI 兼容接口）
- Node.js 18+ (前端开发)

### 开发环境设置

```bash
# 1. 克隆仓库
git clone https://github.com/EterUltimate/AstrCode.git
cd AstrCode

# 2. 下载依赖
go mod download

# 3. 编译后端
go build -o bin/astrcode cmd/server/main.go

# 4. 启动开发服务器
./bin/astrcode \
  -addr :8080 \
  -astrbot-url http://localhost:6185 \
  -llm-url http://localhost:11434 \
  -llm-model qwen2.5 \
  -static-dir ./web
```

### 使用示例

#### 场景 1：创建新插件

**开发者输入：**
```
帮我创建一个天气查询插件，支持城市名查询，返回温度和天气状况
```

**AstrCode 自动完成：**
1. ✅ 生成 `plugin.yaml` 配置文件
2. ✅ 创建 `main.py` 主逻辑文件
3. ✅ 添加错误处理和日志
4. ✅ 编写单元测试
5. ✅ 热加载到 AstrBot
6. ✅ 显示预览效果

#### 场景 2：修改现有插件

**开发者输入：**
```
给签到插件增加连续签到奖励功能，7天送特殊勋章
```

**AstrCode 执行：**
1. 🔍 分析现有代码结构
2. ✏️ 生成增量修改方案
3. 🧪 运行回归测试
4. 🚀 部署新版本
5. 📊 展示变更对比

#### 场景 3：代码审查

**开发者输入：**
```
审查我刚写的翻译插件，看看有没有安全问题
```

**AstrCode 反馈：**
- ⚠️ 发现硬编码 API Key（建议移至配置）
- 💡 建议添加请求频率限制
- ✅ 代码结构清晰，符合规范
- 📝 生成改进后的代码

### Windows (MSI Installer)

1. 下载 `AstrCode-0.4.0-x64.msi` 从 [Releases](https://github.com/EterUltimate/AstrCode/releases)
2. 双击运行安装程序
3. 选择安装路径（默认：`C:\Program Files\AstrCode`）
4. 完成安装后，从开始菜单启动或运行：
   ```batch
   "C:\Program Files\AstrCode\bin\start.bat"
   ```
5. 浏览器打开 `http://localhost:8080`

**卸载**：通过控制面板 → 程序和功能 → 卸载 AstrCode

### Docker 运行
docker build -t astrcode:latest .
docker run -p 8080:8080 astrcode:latest
```

### 访问 Dashboard

打开浏览器访问 `http://localhost:8080`

---

## API 参考

### REST API

| 方法 | 路径 | 说明 | 请求体 |
|------|------|------|--------|
| POST | `/api/task` | 提交任务 | `{"task": "string", "async": false}` |
| GET | `/api/task/{id}` | 查询任务状态 | - |
| GET | `/api/tasks` | 列出所有任务 | - |
| GET | `/api/snapshot/{id}` | 执行快照（可视化数据） | - |
| GET | `/api/skills` | 获取可用技能列表 | - |
| POST | `/api/plan` | 仅生成计划（不执行） | `{"task": "string"}` |
| POST | `/api/execute` | 直接执行 Handler | `{"handler": "string", "event": {...}}` |
| GET | `/health` | 健康检查 | - |

### 示例：提交异步任务

```bash
curl -X POST http://localhost:8080/api/task \
  -H "Content-Type: application/json" \
  -d '{"task": "读取代码并修复 bug", "async": true}'
```

响应：
```json
{
  "task_id": "task_1777190257056353800",
  "status": "pending",
  "ws": "/ws"
}
```

### 示例：查询任务状态

```bash
curl http://localhost:8080/api/task/task_1777190257056353800
```

---

## WebSocket 事件

连接 `ws://localhost:8080/ws` 接收实时事件推送。

### 事件类型

| 事件 | 说明 | 数据结构 |
|------|------|----------|
| `connected` | 连接确认 | `{"type": "connected", "timestamp": 1234567890}` |
| `task_created` | 任务创建 | `{"type": "task_created", "task_id": "...", "content": "..."}` |
| `task_completed` | 任务完成 | `{"type": "task_completed", "task_id": "...", "result": "..."}` |
| `task_failed` | 任务失败 | `{"type": "task_failed", "task_id": "...", "error": "..."}` |
| `step_started` | 步骤开始 | `{"type": "step_started", "task_id": "...", "step_id": "...", "name": "..."}` |
| `step_completed` | 步骤完成 | `{"type": "step_completed", "task_id": "...", "step_id": "...", "duration_ms": 123}` |
| `step_failed` | 步骤失败 | `{"type": "step_failed", "task_id": "...", "step_id": "...", "error": "..."}` |
| `step_retry` | 步骤重试 | `{"type": "step_retry", "task_id": "...", "step_id": "...", "retry_count": 1}` |

### JavaScript 示例

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log(`[${data.type}]`, data);
  
  // 根据事件类型更新 UI
  switch (data.type) {
    case 'step_started':
      updateStepStatus(data.step_id, 'running');
      break;
    case 'step_completed':
      updateStepStatus(data.step_id, 'completed', data.duration_ms);
      break;
    case 'task_completed':
      showTaskResult(data.result);
      break;
  }
};
```

---

## 项目结构

```
AstrCode/
├── cmd/server/
│   └── main.go                  # 入口：启动开发服务器
├── internal/
│   ├── agent/
│   │   └── agent.go             # 核心 Agent：理解需求 + 生成代码
│   ├── api/
│   │   ├── server.go            # HTTP API 服务器
│   │   └── hub.go               # WebSocket 广播中心
│   ├── codegen/
│   │   ├── generator.go         # 代码生成引擎
│   │   ├── templates/           # AstrBot 插件模板
│   │   └── reviewer.go          # 代码审查器
│   ├── deploy/
│   │   ├── tester.go            # 自动化测试 runner
│   │   └── hotreload.go         # 热重载管理器
│   ├── sdk/
│   │   ├── client.go            # AstrBot JSON-RPC 客户端
│   │   └── transport.go         # WebSocket 传输层
│   ├── skill/
│   │   ├── retriever.go         # Skill 检索器（RAG）
│   │   └── star_manager.go      # Plugin 发现器
│   └── model/
│       ├── astrbot.go           # AstrBot 数据模型
│       └── taskstore.go         # 开发会话管理
├── web/
│   └── index.html               # CodeX-like Dashboard UI
├── configs/
│   └── config.yaml              # 配置文件
├── scripts/
│   ├── build.sh                 # 构建脚本
│   └── build-msi.ps1            # MSI 打包脚本
├── .github/workflows/
│   ├── ci.yml                   # CI 流水线
│   └── release.yml              # Release 自动化
├── Dockerfile                   # Docker 多阶段构建
├── Makefile                     # Make 命令
└── README.md                    # 本文档
```

---

## 开发路线图

### ✅ Phase 1 — 基础架构（已完成）

- [x] AstrBot JSON-RPC SDK 客户端
- [x] WebSocket 传输层（心跳+重连）
- [x] Plugin 发现器（plugin.yaml 解析）
- [x] Skill 检索器（RAG + 关键词）
- [x] 基础代码生成框架

### ✅ Phase 2 — 智能代码生成（已完成）

- [x] 意图识别引擎
- [x] AstrBot 插件模板系统
- [x] 代码生成器（Python/Go）
- [x] 上下文管理（多轮对话）
- [x] CodeX-like UI 原型

### ✅ Phase 3 — 代码审查与测试（已完成）

- [x] 静态代码分析器
- [x] 安全扫描规则
- [x] 自动化测试 runner
- [x] 热重载管理器
- [x] 差异对比视图

### 🚧 Phase 4 — 高级功能（进行中）

- [ ] Core 代码贡献辅助
- [ ] 依赖冲突检测
- [ ] 性能分析工具
- [ ] 协作开发支持
- [ ] 插件市场集成

### 🔮 未来规划

- [ ] Phase 5 — AI 驱动的重构建议
- [ ] Phase 6 — 多语言支持（i18n）
- [ ] Phase 7 — 团队协作工作流
- [ ] Phase 8 — 插件性能监控

---

## 构建与部署

### 本地构建

```bash
# 使用 Makefile
make build        # 构建到 bin/astrcode
make run          # 直接运行
make test         # 运行测试
make clean        # 清理构建产物

# 多平台构建
make build-all    # 构建 linux/amd64, linux/arm64, windows/amd64, darwin/amd64, darwin/arm64
```

### Docker 构建

```bash
# 本地构建
docker build -t astrcode:latest .

# 运行
docker run -p 8080:8080 \
  -v $(pwd)/skills:/root/skills \
  -v $(pwd)/stars:/root/stars \
  astrcode:latest
```

### CI/CD

项目使用 GitHub Actions 自动化：

- **CI Pipeline**（`.github/workflows/ci.yml`）
  - 代码格式化检查（gofmt）
  - 静态分析（go vet + golangci-lint）
  - 单元测试（覆盖率报告）
  - 多平台交叉编译
  - Docker 镜像构建
  - 冒烟测试（健康检查）

- **Release Pipeline**（`.github/workflows/release.yml`）
  - 触发条件：推送 Git tag（`v*`）
  - 自动构建所有平台二进制
  - 生成 SHA256 校验和
  - 创建 GitHub Release
  - 推送 Docker 镜像到 GHCR

---

## 技术栈

### 后端

- **语言**：Go 1.21
- **Web 框架**：标准库 `net/http`
- **WebSocket**：`gorilla/websocket v1.5.3`
- **配置解析**：`gopkg.in/yaml.v3 v3.0.1`
- **协议**：JSON-RPC 2.0

### AI/ML

- **LLM 接口**：OpenAI 兼容 API（支持 Ollama/OpenAI/Azure）
- **Embedding**：向量相似度搜索（余弦相似度）
- **RAG**：内存向量存储（可扩展到 FAISS/Milvus）

### 基础设施

- **容器化**：Docker 多阶段构建
- **CI/CD**：GitHub Actions
- **缓存**：内存 + 磁盘 + Redis（可选）
- **日志**：标准库 `log`

### 前端

- **Dashboard**：原生 HTML/CSS/JavaScript（无框架依赖）
- **样式**：CSS Grid + Flexbox
- **字体**：SF Mono / Consolas

---

## 性能指标

| 指标 | 数值 |
|------|------|
| API 响应时间（P95） | < 100ms |
| WebSocket 延迟 | < 10ms |
| 并发任务支持 | 100+ |
| 技能检索速度 | < 50ms（1000 个技能） |
| 内存占用（空闲） | ~50MB |
| Docker 镜像大小 | ~30MB（Alpine） |

---

## 常见问题

### Q: 如何添加自定义技能？

在 `skills/` 目录创建 `SKILL.md` 文件：

```markdown
# read_code

读取指定路径的代码文件内容。

## Parameters
- path: 文件路径（必填）

## Example
read_code(path="internal/agent/agent.go")
```

重启服务后自动加载。

### Q: 如何连接到远程 AstrBot？

```bash
./bin/astrcode -astrbot-url http://your-astrbot:6185 -astrbot-token your_token
```

### Q: WebSocket 连接断开怎么办？

Dashboard 会自动重连（最多 5 次）。也可以手动刷新页面。

### Q: 如何禁用向量检索？

启动时不加 `-use-vector` 参数，系统会自动降级为关键词匹配。

---

## 贡献指南

1. Fork 本仓库
2. 创建特性分支（`git checkout -b feature/AmazingFeature`）
3. 提交更改（`git commit -m 'Add some AmazingFeature'`）
4. 推送到分支（`git push origin feature/AmazingFeature`）
5. 开启 Pull Request

### 代码规范

- 遵循 [Effective Go](https://go.dev/doc/effective_go)
- 使用 `gofmt` 格式化代码
- 所有公共 API 必须有注释
- 新增功能需附带单元测试

---

## 许可证

本项目采用 **GNU Affero General Public License v3.0 (AGPL-3.0)** 开源许可证。

### 主要条款

- ✅ 允许自由使用、修改和分发
- ✅ 允许商业使用
- ⚠️ **网络使用也必须开源**（AGPL 核心要求）
- ⚠️ 修改后的代码必须以相同许可证开源
- ⚠️ 必须保留版权声明和许可证文本

详见 [LICENSE](LICENSE) 文件完整条款。

### 为什么选择 AGPL-3.0？

AstrCode 作为网络服务编排引擎，AGPL-3.0 确保：
1. 通过网络提供服务时也必须公开源代码
2. 防止闭源商业滥用
3. 促进社区协作和改进共享

---

## 致谢

- [AstrBot](https://github.com/Soulter/AstrBot) - 强大的聊天机器人框架
- [astrbot-sdk](https://github.com/Soulter/astrbot-sdk) - Python SDK 参考实现
- [Ollama](https://ollama.ai) - 本地 LLM 运行时

---

**Made with ❤️ by EterUltimate**
