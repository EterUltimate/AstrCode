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
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/EterUltimate/astrcode/internal/agent"
	"github.com/EterUltimate/astrcode/internal/model"
)

// Server HTTP + WebSocket API 服务器
type Server struct {
	agent  *agent.Agent
	hub    *Hub
	store  *model.TaskStore
	mux    *http.ServeMux
	server *http.Server
}

// NewServer 创建新的 API 服务器
func NewServer(ag *agent.Agent, hub *Hub, store *model.TaskStore, addr string) *Server {
	s := &Server{
		agent: ag,
		hub:   hub,
		store: store,
	}

	mux := http.NewServeMux()
	s.mux = mux

	// 任务 API
	mux.HandleFunc("/api/task", s.handleTask)
	mux.HandleFunc("/api/task/", s.handleTaskStatus)

	// 代码生成 API (Phase 2: Dev Assistant)
	mux.HandleFunc("/api/generate", s.handleGeneratePlugin)
	mux.HandleFunc("/api/review", s.handleReviewCode)
	mux.HandleFunc("/api/deploy", s.handleDeployPlugin)

	// 技能/计划/执行 API
	mux.HandleFunc("/api/skills", s.handleSkills)
	mux.HandleFunc("/api/plan", s.handlePlan)
	mux.HandleFunc("/api/execute", s.handleExecute)

	// 可视化/状态 API
	mux.HandleFunc("/api/tasks", s.handleTaskList)
	mux.HandleFunc("/api/snapshot/", s.handleSnapshot)

	// WebSocket
	mux.HandleFunc("/ws", hub.HandleWS)
	
	// Phase 6: WebSocket 统计 API
	mux.HandleFunc("/api/ws/stats", s.handleWSStats)

	// 健康检查
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/health", s.handleHealth)

	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return s
}

// Start 启动服务器
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// AddStaticHandler 添加静态文件处理器
func (s *Server) AddStaticHandler(pattern string, handler http.Handler) {
	if s.mux != nil {
		// 使用 HandleFunc 包装以支持 SPA 路由
		s.mux.Handle(pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 如果请求的是 API 路径,让其他 handler 处理
			if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/ws" {
				http.NotFound(w, r)
				return
			}
			handler.ServeHTTP(w, r)
		}))
	}
}

// ============================================================
// 任务 API
// ============================================================

// handleTask 提交任务（完整流程，异步执行）
func (s *Server) handleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Task  string `json:"task"`
		Async bool   `json:"async,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Async {
		// 异步执行
		task, err := s.agent.CreateTask(req.Task)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); encodeErr != nil {
				_ = encodeErr // TODO: Add proper logging
			}
			return
		}

		// 存储任务
		s.store.CreateTask(task)

		// 后台执行
		go func() {
			s.agent.ProcessTaskAsync(context.Background(), task)
			s.store.UpdateTask(task)
		}()

		w.Header().Set("Content-Type", "application/json")
		if encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
			"task_id": task.ID,
			"status":  task.Status,
			"ws":      "/ws",
		}); encodeErr != nil {
			_ = encodeErr // TODO: Add proper logging
		}
		return
	}

	// 同步执行
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	result, err := s.agent.ProcessTask(ctx, req.Task)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
			"task":  result,
		}); encodeErr != nil {
			_ = encodeErr // TODO: Add proper logging
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(result); encodeErr != nil {
		_ = encodeErr // TODO: Add proper logging
	}
}

// handleTaskStatus 查询任务状态
func (s *Server) handleTaskStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}
	taskID := parts[3]

	task, ok := s.store.GetTask(taskID)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": "task not found"}); encodeErr != nil {
			_ = encodeErr // TODO: Add proper logging
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(task); encodeErr != nil {
		_ = encodeErr // TODO: Add proper logging
	}
}

// handleTaskList 列出所有任务
func (s *Server) handleTaskList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tasks := s.store.ListTasks(100)
	running := s.store.ListRunning()

	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
		"total":   len(tasks),
		"running": len(running),
		"tasks":   tasks,
	}); encodeErr != nil {
		_ = encodeErr // TODO: Add proper logging
	}
}

// handleSnapshot 执行快照（可视化数据）
func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}
	taskID := parts[3]

	task, ok := s.store.GetTask(taskID)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": "task not found"}); encodeErr != nil {
			_ = encodeErr // TODO: Add proper logging
		}
		return
	}

	// 获取关联的 Plan
	var plan *model.Plan
	if task.Result != "" {
		// 尝试从 planCache 获取
	} else {
		plan, _ = s.store.GetPlan(task.ID)
	}

	snapshot := model.BuildSnapshot(task, plan, nil)

	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(snapshot); encodeErr != nil {
		_ = encodeErr // TODO: Add proper logging
	}
}

// ============================================================
// 技能/计划/执行 API
// ============================================================

// handleSkills 获取可用技能列表
func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stars := s.agent.GetStarManager().GetAllStars()
	allSkills := s.agent.GetAllSkills()

	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
		"stars":  stars,
		"skills": allSkills,
		"count":  len(allSkills),
	}); encodeErr != nil {
		_ = encodeErr // TODO: Add proper logging
	}
}

// handlePlan 生成计划（不执行）
func (s *Server) handlePlan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Task string `json:"task"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	plan, err := s.agent.GeneratePlan(ctx, req.Task)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); encodeErr != nil {
			_ = encodeErr // TODO: Add proper logging
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(plan); encodeErr != nil {
		_ = encodeErr // TODO: Add proper logging
	}
}

// handleExecute 直接执行步骤
func (s *Server) handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Handler string                  `json:"handler"`
		Event   *model.AstrMessageEvent `json:"event"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if req.Event == nil {
		req.Event = &model.AstrMessageEvent{
			MessageStr: req.Handler,
			SessionID:  "api",
			PlatformMeta: model.PlatformMetadata{
				ID:   "api",
				Name: "AstrCode API",
			},
		}
	}

	result, err := s.agent.ProcessEvent(ctx, req.Handler, req.Event)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); encodeErr != nil {
			_ = encodeErr // TODO: Add proper logging
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(result); encodeErr != nil {
		_ = encodeErr // TODO: Add proper logging
	}
}

// handleHealth 健康检查
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "ok",
		"time":       time.Now().Format(time.RFC3339),
		"version":    "0.4.0",
		"phase":      "4",
		"ws_clients": s.hub.ClientCount(),
		"features": []string{
			"json-rpc-sdk",
			"websocket-transport",
			"star-discovery",
			"skill-retriever",
			"plan-execution",
			"adaptive-planning",
			"fallback-strategies",
			"persist-cache",
			"task-store",
			"ws-broadcast",
			"execution-snapshot",
			"async-task",
			"code-generation", // NEW: Phase 2
			"code-review",     // NEW: Phase 2
			"hot-deploy",      // NEW: Phase 2
		},
	}); encodeErr != nil {
		_ = encodeErr // TODO: Add proper logging
	}
}

// Phase 6: handleWSStats WebSocket 统计信息
func (s *Server) handleWSStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	stats := s.hub.GetStats()
	if encodeErr := json.NewEncoder(w).Encode(stats); encodeErr != nil {
		_ = encodeErr
	}
}

// ============================================================
// Phase 2: Dev Assistant API Handlers
// ============================================================

// handleGeneratePlugin 处理插件生成请求
func (s *Server) handleGeneratePlugin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Requirement string `json:"requirement"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"}); encodeErr != nil {
			_ = encodeErr // TODO: Add proper logging
		}
		return
	}

	// TODO: 调用 codegen.Generator
	// 这里返回占位响应
	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "ok",
		"message":     "Plugin generation endpoint ready",
		"requirement": req.Requirement,
	}); encodeErr != nil {
		_ = encodeErr // TODO: Add proper logging
	}
}

// handleReviewCode 处理代码审查请求
func (s *Server) handleReviewCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Files map[string]string `json:"files"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"}); encodeErr != nil {
			_ = encodeErr // TODO: Add proper logging
		}
		return
	}

	// TODO: 调用 codegen.Reviewer
	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "ok",
		"message":     "Code review endpoint ready",
		"files_count": len(req.Files),
	}); encodeErr != nil {
		_ = encodeErr // TODO: Add proper logging
	}
}

// handleDeployPlugin 处理插件部署请求
func (s *Server) handleDeployPlugin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PluginName string            `json:"plugin_name"`
		Files      map[string]string `json:"files"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"}); encodeErr != nil {
			_ = encodeErr // TODO: Add proper logging
		}
		return
	}

	// TODO: 调用 deploy.Manager
	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "ok",
		"message":     "Plugin deployment endpoint ready",
		"plugin_name": req.PluginName,
	}); encodeErr != nil {
		_ = encodeErr // TODO: Add proper logging
	}
}
