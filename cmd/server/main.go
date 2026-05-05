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

package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EterUltimate/astrcode/internal/agent"
	"github.com/EterUltimate/astrcode/internal/api"
	"github.com/EterUltimate/astrcode/internal/cache"
	"github.com/EterUltimate/astrcode/internal/llm"
	"github.com/EterUltimate/astrcode/internal/model"
	"github.com/EterUltimate/astrcode/internal/rag"
	"github.com/EterUltimate/astrcode/internal/sdk"
	"github.com/EterUltimate/astrcode/internal/skill"
)

func main() {
	var (
		addr         = flag.String("addr", ":8080", "Server address")
		astrbotURL   = flag.String("astrbot-url", "http://localhost:6185", "AstrBot API URL")
		astrbotToken = flag.String("astrbot-token", "", "AstrBot API Token")
		llmProvider  = flag.String("llm-provider", "openai", "LLM provider: openai, gemini, claude")
		llmURL       = flag.String("llm-url", "https://api.openai.com", "LLM API URL")
		llmKey       = flag.String("llm-key", "", "LLM API Key")
		llmModel     = flag.String("llm-model", "gpt-4o", "LLM Model name")
		skillsDir    = flag.String("skills-dir", "./skills", "Skills directory")
		starsDir     = flag.String("stars-dir", "./stars", "Stars directory")
		cacheDir     = flag.String("cache-dir", "./cache", "Cache directory")
		useVector    = flag.Bool("use-vector", false, "Enable vector search (requires embedding API)")
		embedURL     = flag.String("embed-url", "", "Embedding API URL (defaults to llm-url)")
		embedModel   = flag.String("embed-model", "", "Embedding model (defaults to llm-model)")
		staticDir    = flag.String("static-dir", "./web", "Static files directory for dashboard")
	)
	flag.Parse()

	// 创建 SDK 客户端
	sdkClient := sdk.NewAstrBotClient(*astrbotURL, *astrbotToken)

	// 创建 LLM 客户端
	provider := llm.ProviderType(*llmProvider)
	llmClient := llm.NewClient(provider, *llmURL, *llmKey, *llmModel)

	// 创建 Agent
	var ag *agent.Agent
	if *useVector {
		embedURLVal := *embedURL
		if embedURLVal == "" {
			embedURLVal = *llmURL
		}
		embedModelVal := *embedModel
		if embedModelVal == "" {
			embedModelVal = *llmModel
		}
		embedding := rag.NewEmbeddingClient(embedURLVal, embedModelVal, *llmKey)
		store := rag.NewMemoryVectorStore()
		ag = agent.NewAgentWithVector(llmClient, sdkClient, embedding, store)
		log.Println("Vector search enabled")
	} else {
		ag = agent.NewAgent(llmClient, sdkClient)
	}

	// WebSocket Hub
	hub := api.NewHub()
	go hub.Run()

	// 事件回调：Executor → Hub → WebSocket 客户端
	ag.SetEventSink(func(event *model.WSEvent) {
		hub.Broadcast(event)
	})

	// TaskStore
	taskStore := model.NewTaskStore()

	// 加载 Skill
	loader := skill.NewLoader(*skillsDir)
	skills, err := loader.LoadFromDirectory()
	if err != nil {
		log.Printf("Warning: failed to load skills: %v", err)
	} else {
		log.Printf("Loaded %d skills", len(skills))
		for _, s := range skills {
			ag.RegisterSkill(s)
			log.Printf("  - %s: %s", s.Name, s.Description)
		}
	}

	// 发现 Stars
	if *starsDir != "" {
		if _, err := os.Stat(*starsDir); err == nil {
			stars, err := ag.DiscoverStars(*starsDir)
			if err != nil {
				log.Printf("Warning: failed to discover stars: %v", err)
			} else {
				log.Printf("Discovered %d stars", len(stars))
				for _, star := range stars {
					log.Printf("  - %s v%s by %s", star.Name, star.Version, star.Author)
				}
			}
		}
	}

	// 初始化持久化缓存
	if *cacheDir != "" {
		if err := os.MkdirAll(*cacheDir, 0755); err == nil {
			hybridCache, err := cache.NewHybridCache(*cacheDir + "/astrcode.cache")
			if err != nil {
				log.Printf("Warning: failed to create hybrid cache: %v", err)
			} else {
				log.Println("Hybrid cache initialized")
				_ = hybridCache
			}
		}
	}

	// 创建 API 服务器
	server := api.NewServer(ag, hub, taskStore, *addr)

	// 获取服务器的 mux 并添加静态文件支持
	if *staticDir != "" {
		if _, err := os.Stat(*staticDir); err == nil {
			fs := http.FileServer(http.Dir(*staticDir))
			// 将静态文件处理器添加到服务器的 mux
			server.AddStaticHandler("/", fs)
			log.Printf("Dashboard UI served from %s at /", *staticDir)
		} else {
			log.Printf("Warning: static directory %s not found", *staticDir)
		}
	}

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("AstrCode v0.4.0 server starting on %s", *addr)
		log.Printf("Endpoints: /api/task /api/skills /api/plan /api/execute /api/tasks /api/snapshot/{id} /ws /health")
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
