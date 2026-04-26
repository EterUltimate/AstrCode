package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketTransport 基于 WebSocket 的 JSON-RPC 传输（对齐 astrbot-sdk/runtime/rpc/client/websocket.py）
type WebSocketTransport struct {
	URL               string
	Heartbeat         time.Duration
	ReconnectInterval time.Duration
	AutoReconnect     bool

	conn     *websocket.Conn
	mu       sync.Mutex
	handler  func([]byte)
	running  bool
	done     chan struct{}
	requests map[string]chan json.RawMessage
	reqMu    sync.Mutex
	reqID    int
}

// NewWebSocketTransport 创建 WebSocket 传输
func NewWebSocketTransport(url string) *WebSocketTransport {
	return &WebSocketTransport{
		URL:               url,
		Heartbeat:         30 * time.Second,
		ReconnectInterval: 5 * time.Second,
		AutoReconnect:     true,
		requests:          make(map[string]chan json.RawMessage),
	}
}

// Connect 建立连接
func (t *WebSocketTransport) Connect(ctx context.Context) error {
	dialer := websocket.Dialer{}
	conn, _, err := dialer.DialContext(ctx, t.URL, nil)
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}
	t.conn = conn
	t.running = true
	t.done = make(chan struct{})

	go t.readLoop()
	go t.heartbeatLoop()

	return nil
}

// SetMessageHandler 设置消息处理回调
func (t *WebSocketTransport) SetMessageHandler(handler func([]byte)) {
	t.handler = handler
}

// Send 发送 JSON-RPC 消息
func (t *WebSocketTransport) Send(msg interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.conn == nil {
		return fmt.Errorf("not connected")
	}

	return t.conn.WriteJSON(msg)
}

// Call 发送请求并等待响应
func (t *WebSocketTransport) Call(ctx context.Context, method string, params map[string]interface{}) (json.RawMessage, error) {
	t.reqMu.Lock()
	t.reqID++
	id := fmt.Sprintf("%d", t.reqID)
	respCh := make(chan json.RawMessage, 1)
	t.requests[id] = respCh
	t.reqMu.Unlock()

	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}

	if err := t.Send(req); err != nil {
		t.cleanupRequest(id)
		return nil, err
	}

	select {
	case <-ctx.Done():
		t.cleanupRequest(id)
		return nil, ctx.Err()
	case resp := <-respCh:
		return resp, nil
	case <-time.After(30 * time.Second):
		t.cleanupRequest(id)
		return nil, fmt.Errorf("request timeout")
	}
}

// Close 关闭连接
func (t *WebSocketTransport) Close() error {
	t.running = false
	close(t.done)

	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}

func (t *WebSocketTransport) readLoop() {
	for t.running {
		_, message, err := t.conn.ReadMessage()
		if err != nil {
			if t.running && t.AutoReconnect {
				time.Sleep(t.ReconnectInterval)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if err := t.Connect(ctx); err != nil {
					cancel()
					continue
				}
				cancel()
			}
			continue
		}

		// 尝试匹配 pending request
		var baseMsg struct {
			ID     string          `json:"id"`
			Result json.RawMessage `json:"result,omitempty"`
			Error  interface{}     `json:"error,omitempty"`
		}
		if err := json.Unmarshal(message, &baseMsg); err == nil && baseMsg.ID != "" {
			t.reqMu.Lock()
			if ch, ok := t.requests[baseMsg.ID]; ok {
				delete(t.requests, baseMsg.ID)
				t.reqMu.Unlock()
				ch <- baseMsg.Result
				continue
			}
			t.reqMu.Unlock()
		}

		// 通知上层处理
		if t.handler != nil {
			t.handler(message)
		}
	}
}

func (t *WebSocketTransport) heartbeatLoop() {
	ticker := time.NewTicker(t.Heartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-t.done:
			return
		case <-ticker.C:
			t.mu.Lock()
			if t.conn != nil {
				t.conn.WriteMessage(websocket.PingMessage, nil)
			}
			t.mu.Unlock()
		}
	}
}

func (t *WebSocketTransport) cleanupRequest(id string) {
	t.reqMu.Lock()
	delete(t.requests, id)
	t.reqMu.Unlock()
}
