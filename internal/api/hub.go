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

// Hub WebSocket 广播中心
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan *model.WSEvent
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.RWMutex
}

// NewHub 创建广播中心
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan *model.WSEvent, 256),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Run 启动广播循环
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.clients[conn] = true
			h.mu.Unlock()
			log.Printf("WebSocket client connected (total: %d)", len(h.clients))

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
			h.mu.Unlock()
			log.Printf("WebSocket client disconnected (total: %d)", len(h.clients))

		case event := <-h.broadcast:
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			h.mu.RLock()
			for conn := range h.clients {
				conn.WriteMessage(websocket.TextMessage, data)
			}
			h.mu.RUnlock()
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

// ClientCount 客户端数量
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
	h.register <- conn

	// 发送连接确认
	conn.WriteJSON(map[string]interface{}{
		"type":      "connected",
		"timestamp": time.Now().UnixMilli(),
		"message":   "AstrCode execution stream",
	})

	// 读循环（处理断开）
	defer func() {
		h.unregister <- conn
	}()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		// 客户端发送的消息（心跳/ping）暂不处理
	}
}
