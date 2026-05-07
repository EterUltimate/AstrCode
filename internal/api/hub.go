// AstrCode - Agent orchestration engine for AstrBot
// Copyright (C) 2026 EterUltimate
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/EterUltimate/astrcode/internal/model"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Client 表示一个 WebSocket 客户端连接
type Client struct {
	conn          *websocket.Conn
	subscriptions map[string]bool // 订阅的事件类型或任务ID
	send          chan *model.WSEvent
}

// Hub WebSocket 广播中心（Phase 6: 增强版）
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *model.WSEvent
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex

	// Phase 6: 事件历史缓冲区（用于新客户端回放）
	eventHistory []*model.WSEvent
	maxHistory   int
	historyMu    sync.RWMutex

	// Phase 6: 统计信息
	stats   HubStats
	statsMu sync.RWMutex
}

// HubStats Hub 统计信息
type HubStats struct {
	TotalConnections  int64     `json:"total_connections"`
	ActiveConnections int       `json:"active_connections"`
	TotalEventsSent   int64     `json:"total_events_sent"`
	LastEventTime     time.Time `json:"last_event_time"`
	PeakConnections   int       `json:"peak_connections"`
}

// NewHub 创建广播中心
func NewHub() *Hub {
	return &Hub{
		clients:      make(map[*Client]bool),
		broadcast:    make(chan *model.WSEvent, 256),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		maxHistory:   100, // 保留最近 100 个事件
		eventHistory: make([]*model.WSEvent, 0, 100),
	}
}

// Run 启动广播循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

			// 更新统计
			h.statsMu.Lock()
			h.stats.TotalConnections++
			h.stats.ActiveConnections = len(h.clients)
			if h.stats.ActiveConnections > h.stats.PeakConnections {
				h.stats.PeakConnections = h.stats.ActiveConnections
			}
			h.statsMu.Unlock()

			log.Printf("WebSocket client connected (total: %d)", len(h.clients))

			// Phase 6: 发送历史事件回放（最近 10 个事件）
			go h.replayRecentEvents(client, 10)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				client.conn.Close()
			}
			h.mu.Unlock()

			// 更新统计
			h.statsMu.Lock()
			h.stats.ActiveConnections = len(h.clients)
			h.statsMu.Unlock()

			log.Printf("WebSocket client disconnected (total: %d)", len(h.clients))

		case event := <-h.broadcast:
			// Phase 6: 添加到事件历史
			h.addToHistory(event)

			h.mu.RLock()
			for client := range h.clients {
				// Phase 6: 检查客户端是否订阅了该事件
				if !h.isSubscribed(client, event) {
					continue
				}

				select {
				case client.send <- event:
					// 成功发送
				default:
					// 客户端接收缓冲区满，跳过
					log.Println("Client send buffer full, dropping event")
				}
			}
			h.mu.RUnlock()

			// 更新统计
			h.statsMu.Lock()
			h.stats.TotalEventsSent++
			h.stats.LastEventTime = time.Now()
			h.statsMu.Unlock()
		}
	}
}

// Broadcast 广播事件
func (h *Hub) Broadcast(event *model.WSEvent) {
	select {
	case h.broadcast <- event:
	default:
		log.Println("WebSocket broadcast channel full, dropping event")
	}
}

// Phase 6: addToHistory 添加事件到历史记录
func (h *Hub) addToHistory(event *model.WSEvent) {
	h.historyMu.Lock()
	defer h.historyMu.Unlock()

	h.eventHistory = append(h.eventHistory, event)
	if len(h.eventHistory) > h.maxHistory {
		// 移除最旧的事件
		h.eventHistory = h.eventHistory[1:]
	}
}

// Phase 6: replayRecentEvents 回放最近的事件给新客户端
func (h *Hub) replayRecentEvents(client *Client, count int) {
	h.historyMu.RLock()
	defer h.historyMu.RUnlock()

	start := len(h.eventHistory) - count
	if start < 0 {
		start = 0
	}

	recentEvents := h.eventHistory[start:]
	for _, event := range recentEvents {
		if h.isSubscribed(client, event) {
			select {
			case client.send <- event:
			default:
				// 缓冲区满，停止回放
				return
			}
		}
	}
}

// Phase 6: isSubscribed 检查客户端是否订阅了该事件
func (h *Hub) isSubscribed(client *Client, event *model.WSEvent) bool {
	// 如果没有订阅任何内容，默认接收所有事件
	if len(client.subscriptions) == 0 {
		return true
	}

	// 检查是否订阅了该事件类型
	if client.subscriptions[event.Type] {
		return true
	}

	// 检查是否订阅了该任务ID
	if event.TaskID != "" && client.subscriptions["task:"+event.TaskID] {
		return true
	}

	return false
}

// Phase 6: Subscribe 客户端订阅事件
func (h *Hub) Subscribe(client *Client, subscription string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client.subscriptions == nil {
		client.subscriptions = make(map[string]bool)
	}
	client.subscriptions[subscription] = true
}

// Phase 6: Unsubscribe 客户端取消订阅
func (h *Hub) Unsubscribe(client *Client, subscription string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client.subscriptions != nil {
		delete(client.subscriptions, subscription)
	}
}

// Phase 6: GetStats 获取 Hub 统计信息
func (h *Hub) GetStats() HubStats {
	h.statsMu.RLock()
	defer h.statsMu.RUnlock()

	stats := h.stats
	stats.ActiveConnections = len(h.clients)
	return stats
}

// ClientCount 客户端数量（保持向后兼容）
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// HandleWS WebSocket 连接处理
func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Phase 6: 创建客户端对象
	client := &Client{
		conn:          conn,
		send:          make(chan *model.WSEvent, 64), // 每个客户端的发送缓冲区
		subscriptions: make(map[string]bool),
	}

	h.register <- client

	// 发送连接确认
	if err := conn.WriteJSON(map[string]interface{}{
		"type":      "connected",
		"timestamp": time.Now().UnixMilli(),
		"message":   "AstrCode execution stream",
	}); err != nil {
		_ = err
	}

	// Phase 6: 启动写协程
	go h.writePump(client)

	// 读循环（处理断开和订阅请求）
	defer func() {
		h.unregister <- client
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Phase 6: 处理客户端消息（订阅/取消订阅）
		h.handleClientMessage(client, message)
	}
}

// Phase 6: writePump 向客户端发送消息
func (h *Hub) writePump(client *Client) {
	ticker := time.NewTicker(30 * time.Second) // 心跳间隔
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case event, ok := <-client.send:
			if !ok {
				// 通道已关闭
				_ = client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data, err := json.Marshal(event)
			if err != nil {
				continue
			}

			if err := client.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			// 发送心跳
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Phase 6: handleClientMessage 处理客户端消息
func (h *Hub) handleClientMessage(client *Client, message []byte) {
	var msg struct {
		Type         string `json:"type"`
		Subscription string `json:"subscription,omitempty"`
	}

	if err := json.Unmarshal(message, &msg); err != nil {
		return
	}

	switch msg.Type {
	case "subscribe":
		if msg.Subscription != "" {
			h.Subscribe(client, msg.Subscription)
		}
	case "unsubscribe":
		if msg.Subscription != "" {
			h.Unsubscribe(client, msg.Subscription)
		}
	case "ping":
		// 响应 pong
		_ = client.conn.WriteJSON(map[string]string{"type": "pong"})
	}
}
