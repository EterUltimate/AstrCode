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

package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/EterUltimate/astrcode/internal/cache"
	"github.com/EterUltimate/astrcode/internal/llm"
	"github.com/EterUltimate/astrcode/internal/model"
	"github.com/EterUltimate/astrcode/internal/prompt"
	"github.com/EterUltimate/astrcode/internal/rag"
	"github.com/EterUltimate/astrcode/internal/sdk"
	"github.com/EterUltimate/astrcode/internal/skill"
)

// Agent AstrCode 核心 Agent = AstrBot 智能开发助手
// 职责：理解开发者意图 → 检索相关 Skill → 生成插件代码 → 审查与部署
type Agent struct {
	retriever    *skill.Retriever
	planner      *skill.Planner
	executor     *skill.Executor
	skillLoader  *skill.Loader
	starManager  *skill.StarManager
	skillCache   *cache.SkillCache
	planCache    *cache.MemoryCache
	sdkClient    *sdk.AstrBotClient
	promptEngine *prompt.Engine
	llmClient    *llm.Client
	eventSink    skill.EventSink // Phase 4: 事件回调
}

// NewAgent 创建新的 Agent
func NewAgent(llmClient *llm.Client, sdkClient *sdk.AstrBotClient) *Agent {
	return &Agent{
		retriever:    skill.NewRetriever(),
		planner:      skill.NewPlanner(llmClient),
		executor:     skill.NewExecutor(sdkClient),
		skillLoader:  skill.NewLoader("./skills"),
		starManager:  skill.NewStarManager(),
		skillCache:   cache.NewSkillCache(),
		planCache:    cache.NewMemoryCache(),
		sdkClient:    sdkClient,
		promptEngine: prompt.NewEngine(),
		llmClient:    llmClient,
	}
}

// NewAgentWithVector 创建带向量检索的 Agent
func NewAgentWithVector(llmClient *llm.Client, sdkClient *sdk.AstrBotClient, embedding *rag.EmbeddingClient, store rag.VectorStore) *Agent {
	index := rag.NewSkillIndex(store, embedding)
	return &Agent{
		retriever:    skill.NewRetrieverWithIndex(index),
		planner:      skill.NewPlanner(llmClient),
		executor:     skill.NewExecutor(sdkClient),
		skillLoader:  skill.NewLoader("./skills"),
		starManager:  skill.NewStarManager(),
		skillCache:   cache.NewSkillCache(),
		planCache:    cache.NewMemoryCache(),
		sdkClient:    sdkClient,
		promptEngine: prompt.NewEngine(),
		llmClient:    llmClient,
	}
}

// SetEventSink 设置事件回调（WebSocket 推送用）
func (a *Agent) SetEventSink(sink skill.EventSink) {
	a.eventSink = sink
	a.executor.SetEventSink(sink)
}

// RegisterSkill 注册 Skill
func (a *Agent) RegisterSkill(s model.Skill) {
	a.retriever.Register(s)
}

// RegisterSkillWithIndex 注册并索引 Skill
func (a *Agent) RegisterSkillWithIndex(ctx context.Context, s model.Skill) error {
	return a.retriever.RegisterAndIndex(ctx, s)
}

// LoadSkills 从目录加载 SKILL.md
func (a *Agent) LoadSkills(dir string) ([]model.Skill, error) {
	a.skillLoader = skill.NewLoader(dir)
	skills, err := a.skillLoader.LoadFromDirectory()
	if err != nil {
		return nil, err
	}
	for _, s := range skills {
		a.retriever.Register(s)
	}
	return skills, nil
}

// DiscoverStars 从 plugin.yaml 发现插件
func (a *Agent) DiscoverStars(dir string) ([]*model.StarMetadata, error) {
	return a.starManager.DiscoverAllStars(dir)
}

// GetStarManager 获取 Star 管理器
func (a *Agent) GetStarManager() *skill.StarManager {
	return a.starManager
}

// GetAllSkills 获取所有已注册的 Skill
func (a *Agent) GetAllSkills() []model.Skill {
	return a.retriever.AllSkills()
}

// CreateTask 创建任务（不执行）
func (a *Agent) CreateTask(taskContent string) (*model.Task, error) {
	return &model.Task{
		ID:        fmt.Sprintf("task_%d", time.Now().UnixNano()),
		Content:   taskContent,
		Status:    model.TaskStatusPending,
		CreatedAt: time.Now().Unix(),
	}, nil
}

// ProcessTaskAsync 异步执行任务（由 goroutine 调用）
func (a *Agent) ProcessTaskAsync(ctx context.Context, task *model.Task) {
	result, err := a.processTask(ctx, task.Content)
	if err != nil {
		task.Status = model.TaskStatusFailed
		task.Error = err.Error()
	} else {
		task.Status = result.Status
		task.Result = result.Result
		task.Error = result.Error
	}
	task.CompletedAt = time.Now().Unix()
}

// ProcessTask 处理用户任务（核心流程，同步）
func (a *Agent) ProcessTask(ctx context.Context, taskContent string) (*model.Task, error) {
	return a.processTask(ctx, taskContent)
}

// processTask 内部实现
func (a *Agent) processTask(ctx context.Context, taskContent string) (*model.Task, error) {
	task := &model.Task{
		ID:        fmt.Sprintf("task_%d", time.Now().UnixNano()),
		Content:   taskContent,
		Status:    model.TaskStatusPending,
		CreatedAt: time.Now().Unix(),
	}

	// 1. 决策阶段
	allSkills := a.retriever.AllSkills()
	decisionPrompt := a.promptEngine.BuildDecisionPrompt(taskContent, allSkills)
	decisionResp, err := a.llmClient.Complete(ctx, decisionPrompt)
	if err != nil {
		task.Status = model.TaskStatusFailed
		task.Error = fmt.Sprintf("decision failed: %v", err)
		task.CompletedAt = time.Now().Unix()
		return task, err
	}

	decision, err := a.promptEngine.ParseDecision(decisionResp)
	if err != nil {
		decision = &prompt.DecisionOutput{Action: "plan", Confidence: 0.5}
	}

	if decision.NeedsClarify {
		task.Status = model.TaskStatusFailed
		task.Error = "needs clarification"
		task.Result = fmt.Sprintf("需要澄清: %v", decision.Questions)
		task.CompletedAt = time.Now().Unix()
		return task, fmt.Errorf("needs clarification")
	}

	// 2. Skill 检索
	task.Status = model.TaskStatusPlanning
	skills, err := a.retriever.Retrieve(ctx, taskContent, 5)
	if err != nil {
		task.Status = model.TaskStatusFailed
		task.Error = fmt.Sprintf("skill retrieval failed: %v", err)
		task.CompletedAt = time.Now().Unix()
		return task, err
	}

	// 3. 生成执行计划
	plan, err := a.planner.GeneratePlan(ctx, taskContent, skills)
	if err != nil {
		task.Status = model.TaskStatusFailed
		task.Error = fmt.Sprintf("plan generation failed: %v", err)
		task.CompletedAt = time.Now().Unix()
		return task, err
	}

	if err := skill.ValidatePlan(plan); err != nil {
		task.Status = model.TaskStatusFailed
		task.Error = fmt.Sprintf("plan validation failed: %v", err)
		task.CompletedAt = time.Now().Unix()
		return task, err
	}

	a.planCache.Set(cache.HashKey("plan:"+taskContent), plan, 30*time.Minute)

	// 4. 执行计划
	task.Status = model.TaskStatusExecuting
	return a.executor.Execute(ctx, plan)
}

// GeneratePlan 生成计划（不执行，用于 /api/plan 预览）
func (a *Agent) GeneratePlan(ctx context.Context, taskContent string) (*model.Plan, error) {
	skills, err := a.retriever.Retrieve(ctx, taskContent, 5)
	if err != nil {
		return nil, err
	}
	return a.planner.GeneratePlan(ctx, taskContent, skills)
}

// ProcessEvent 处理 AstrBot 消息事件
func (a *Agent) ProcessEvent(ctx context.Context, handlerFullName string, event *model.AstrMessageEvent) (*model.MessageEventResult, error) {
	return a.executor.ExecuteHandler(ctx, handlerFullName, event)
}

// ListTools 列出可用的 Tools
func (a *Agent) ListTools(ctx context.Context) ([]model.FunctionTool, error) {
	return a.sdkClient.ListTools(ctx)
}
