package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/EterUltimate/astrcode/internal/model"
)

// AstrBotClient AstrBot SDK 客户端（对齐 astrbot-sdk JSON-RPC 协议）
type AstrBotClient struct {
	BaseURL      string
	Token        string
	HTTPClient   *http.Client
	MaxRetries   int
	RequestID    int
	mu           sync.Mutex
}

// NewAstrBotClient 创建新的 SDK 客户端
func NewAstrBotClient(baseURL, token string) *AstrBotClient {
	return &AstrBotClient{
		BaseURL:    baseURL,
		Token:      token,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		MaxRetries: 3,
	}
}

// nextRequestID 生成递增请求 ID
func (c *AstrBotClient) nextRequestID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.RequestID++
	return fmt.Sprintf("%d", c.RequestID)
}

// ============================================================
// JSON-RPC 通信
// ============================================================

// CallRPC 发送 JSON-RPC 请求并等待响应
func (c *AstrBotClient) CallRPC(ctx context.Context, method string, params map[string]interface{}) (map[string]interface{}, error) {
	req := &model.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  method,
		Params:  params,
	}

	var lastErr error
	for i := 0; i < c.MaxRetries; i++ {
		result, err := c.doRPCRequest(ctx, req)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if i < c.MaxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	return nil, fmt.Errorf("RPC call %s failed after %d retries: %w", method, c.MaxRetries, lastErr)
}

func (c *AstrBotClient) doRPCRequest(ctx context.Context, req *model.JSONRPCRequest) (map[string]interface{}, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/rpc", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应：尝试 success 或 error
	var baseResp struct {
		JSONRPC string `json:"jsonrpc"`
		ID      string `json:"id"`
		Result  json.RawMessage `json:"result,omitempty"`
		Error   *model.JSONRPCErrorData `json:"error,omitempty"`
	}
	if err := json.Unmarshal(body, &baseResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if baseResp.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", baseResp.Error.Code, baseResp.Error.Message)
	}

	if baseResp.Result != nil {
		var result map[string]interface{}
		if err := json.Unmarshal(baseResp.Result, &result); err != nil {
			return nil, fmt.Errorf("unmarshal result: %w", err)
		}
		return result, nil
	}

	return map[string]interface{}{}, nil
}

// ============================================================
// AstrBot 高级 API
// ============================================================

// Handshake 与 AstrBot 执行握手，获取插件元数据
func (c *AstrBotClient) Handshake(ctx context.Context) (*model.StarMetadata, error) {
	result, err := c.CallRPC(ctx, "handshake", map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("handshake failed: %w", err)
	}

	// 从 result 中提取 metadata
	meta := &model.StarMetadata{
		Name:      getStringField(result, "name"),
		Author:    getStringField(result, "author"),
		Desc:      getStringField(result, "desc"),
		Version:   getStringField(result, "version"),
		Activated: true,
	}
	return meta, nil
}

// CallHandler 调用指定的 Handler
func (c *AstrBotClient) CallHandler(ctx context.Context, handlerFullName string, event *model.AstrMessageEvent) (*model.MessageEventResult, error) {
	params := map[string]interface{}{
		"handler_full_name": handlerFullName,
		"event":             event,
	}

	result, err := c.CallRPC(ctx, "call_handler", params)
	if err != nil {
		return nil, fmt.Errorf("call handler %s failed: %w", handlerFullName, err)
	}

	// 解析 result 为 MessageEventResult
	resultBytes, _ := json.Marshal(result)
	var msgResult model.MessageEventResult
	if err := json.Unmarshal(resultBytes, &msgResult); err != nil {
		return nil, fmt.Errorf("parse handler result: %w", err)
	}
	return &msgResult, nil
}

// CallSkill 调用指定 Skill（兼容旧接口）
func (c *AstrBotClient) CallSkill(ctx context.Context, skill string, args map[string]interface{}) (string, error) {
	params := map[string]interface{}{
		"skill": skill,
		"args":  args,
	}

	result, err := c.CallRPC(ctx, "skill.call", params)
	if err != nil {
		return "", fmt.Errorf("call skill %s failed: %w", skill, err)
	}

	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// CallTool 调用 Function Tool
func (c *AstrBotClient) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	params := map[string]interface{}{
		"tool":      toolName,
		"arguments": arguments,
	}

	result, err := c.CallRPC(ctx, "tool.call", params)
	if err != nil {
		return "", fmt.Errorf("call tool %s failed: %w", toolName, err)
	}

	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// ListTools 获取可用的 Tool 列表
func (c *AstrBotClient) ListTools(ctx context.Context) ([]model.FunctionTool, error) {
	result, err := c.CallRPC(ctx, "tool.list", nil)
	if err != nil {
		return nil, fmt.Errorf("list tools failed: %w", err)
	}

	toolsRaw, ok := result["tools"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid tools format")
	}

	toolsBytes, _ := json.Marshal(toolsRaw)
	var tools []model.FunctionTool
	if err := json.Unmarshal(toolsBytes, &tools); err != nil {
		return nil, fmt.Errorf("parse tools: %w", err)
	}
	return tools, nil
}

func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
