package model

// ============================================================
// JSON-RPC 协议模型（对齐 astrbot-sdk/runtime/rpc/jsonrpc.py）
// ============================================================

// JSONRPCBaseMessage JSON-RPC 基础消息
type JSONRPCBaseMessage struct {
	JSONRPC string `json:"jsonrpc"`
}

// JSONRPCRequest JSON-RPC 请求
type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      string                 `json:"id,omitempty"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// JSONRPCSuccessResponse JSON-RPC 成功响应
type JSONRPCSuccessResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      string                 `json:"id,omitempty"`
	Result  map[string]interface{} `json:"result"`
}

// JSONRPCErrorData JSON-RPC 错误数据
type JSONRPCErrorData struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSONRPCErrorResponse JSON-RPC 错误响应
type JSONRPCErrorResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      string           `json:"id,omitempty"`
	Error   JSONRPCErrorData `json:"error"`
}

// ============================================================
// AstrBot 事件模型（对齐 astrbot-sdk/api/event/）
// ============================================================

// AstrMessageEvent AstrBot 消息事件
type AstrMessageEvent struct {
	MessageStr         string                 `json:"message_str"`
	MessageObj         AstrBotMessage         `json:"message_obj"`
	PlatformMeta       PlatformMetadata       `json:"platform_meta"`
	SessionID          string                 `json:"session_id"`
	Role               string                 `json:"role"` // "admin" | "member"
	IsWake             bool                   `json:"is_wake"`
	IsAtOrWakeCommand  bool                   `json:"is_at_or_wake_command"`
	Extras             map[string]interface{} `json:"extras,omitempty"`
	Result             *MessageEventResult    `json:"result,omitempty"`
	HasSendOper        bool                   `json:"has_send_oper"`
	CallLLM            bool                   `json:"call_llm"`
	PluginsName        []string               `json:"plugins_name,omitempty"`
}

// AstrBotMessage 消息对象
type AstrBotMessage struct {
	Type      string `json:"type"` // "friend" | "group"
	Message   []MessageComponent `json:"message"`
	GroupID   string `json:"group_id"`
	SelfID    string `json:"self_id"`
	Sender    Sender `json:"sender"`
}

// Sender 消息发送者
type Sender struct {
	UserID    string `json:"user_id"`
	Nickname  string `json:"nickname"`
}

// MessageComponent 消息组件（简化版）
type MessageComponent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	File string `json:"file,omitempty"`
}

// PlatformMetadata 平台元信息
type PlatformMetadata struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// MessageEventResult 事件结果
type MessageEventResult struct {
	ResultType string `json:"result_type"` // "continue" | "stop"
	Chain      []MessageComponent `json:"chain,omitempty"`
	Message    string `json:"message,omitempty"`
}

// ============================================================
// Star/Plugin 元数据（对齐 astrbot-sdk/api/star/star.py）
// ============================================================

// StarMetadata Star/插件元数据
type StarMetadata struct {
	Name                 string   `json:"name"`
	Author               string   `json:"author,omitempty"`
	Desc                 string   `json:"desc,omitempty"`
	Version              string   `json:"version,omitempty"`
	Repo                 string   `json:"repo,omitempty"`
	ModulePath           string   `json:"module_path,omitempty"`
	RootDirName          string   `json:"root_dir_name,omitempty"`
	Reserved             bool     `json:"reserved"`
	Activated            bool     `json:"activated"`
	HandlerFullNames     []string `json:"star_handler_full_names,omitempty"`
	DisplayName          string   `json:"display_name,omitempty"`
	LogoPath             string   `json:"logo_path,omitempty"`
}

// PluginYAML plugin.yaml 定义
type PluginYAML struct {
	Name        string            `json:"name"`
	Author      string            `json:"author,omitempty"`
	Desc        string            `json:"desc,omitempty"`
	Version     string            `json:"version,omitempty"`
	Repo        string            `json:"repo,omitempty"`
	DisplayName string            `json:"display_name,omitempty"`
	Components  []PluginComponent `json:"components,omitempty"`
}

// PluginComponent 插件组件定义
type PluginComponent struct {
	Class    string `json:"class"`
	IsActive *bool  `json:"active,omitempty"`
}

// ============================================================
// Handler 注册表（对齐 astrbot-sdk/runtime/stars/registry/）
// ============================================================

// HandlerMetadata Handler 元数据
type HandlerMetadata struct {
	HandlerFullName string            `json:"handler_full_name"`
	EventType       string            `json:"event_type"` // "AdapterMessageEvent"
	EventFilters    []EventFilterMeta `json:"event_filters,omitempty"`
}

// EventFilterMeta 事件过滤器元数据
type EventFilterMeta struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// ============================================================
// Tool/Function Calling（对齐 astr_agent_sdk/tool.py）
// ============================================================

// FunctionTool 函数工具定义
type FunctionTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"` // JSON Schema
	Active      bool                   `json:"active"`
}

// ToolSet 工具集合
type ToolSet struct {
	Tools []FunctionTool `json:"tools"`
}

// OpenAIFuncDef OpenAI 格式的函数定义
type OpenAIFuncDef struct {
	Type     string       `json:"type"`
	Function OpenAIFunc   `json:"function"`
}

// OpenAIFunc OpenAI 函数体
type OpenAIFunc struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ToOpenAISchema 转换为 OpenAI 函数调用格式
func (ts *ToolSet) ToOpenAISchema() []OpenAIFuncDef {
	result := make([]OpenAIFuncDef, 0, len(ts.Tools))
	for _, tool := range ts.Tools {
		result = append(result, OpenAIFuncDef{
			Type: "function",
			Function: OpenAIFunc{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			},
		})
	}
	return result
}

// ============================================================
// 消息类型（对齐 astr_agent_sdk/message.py）
// ============================================================

// Message 对话消息
type Message struct {
	Role    string      `json:"role"` // "system" | "user" | "assistant" | "tool"
	Content interface{} `json:"content"` // string | []ContentPart
}

// ContentPart 内容片段
type ContentPart struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"` // "function"
	Function ToolCallFunction  `json:"function"`
}

// ToolCallFunction 工具调用函数体
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ============================================================
// Streaming 类型（对齐 astrbot-sdk/runtime/types.py）
// ============================================================

// StreamStartNotification 流式开始通知
type StreamStartNotification struct {
	ID              string `json:"id"`
	HandlerFullName string `json:"handler_full_name"`
}

// StreamUpdateNotification 流式更新通知
type StreamUpdateNotification struct {
	ID              string      `json:"id"`
	HandlerFullName string      `json:"handler_full_name"`
	Data            interface{} `json:"data"`
}

// StreamEndNotification 流式结束通知
type StreamEndNotification struct {
	ID              string `json:"id"`
	HandlerFullName string `json:"handler_full_name"`
}
