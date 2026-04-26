package skill

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/EterUltimate/astrcode/internal/llm"
	"github.com/EterUltimate/astrcode/internal/model"
)

// Planner 负责生成任务执行计划
type Planner struct {
	llmClient *llm.Client
}

// NewPlanner 创建新的计划生成器
func NewPlanner(llmClient *llm.Client) *Planner {
	return &Planner{llmClient: llmClient}
}

// GeneratePlan 根据任务和可用 Skill 生成执行计划
func (p *Planner) GeneratePlan(ctx context.Context, task string, skills []model.Skill) (*model.Plan, error) {
	prompt := buildPlanningPrompt(task, skills)

	response, err := p.llmClient.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	plan, err := parsePlan(response)
	if err != nil {
		return nil, fmt.Errorf("parse plan failed: %w", err)
	}

	// 为每个步骤分配 ID 和默认值
	for i := range plan.Steps {
		plan.Steps[i].ID = fmt.Sprintf("step_%d", i)
		if plan.Steps[i].Status == "" {
			plan.Steps[i].Status = model.StepStatusPending
		}
		if plan.Steps[i].MaxRetries == 0 {
			plan.Steps[i].MaxRetries = 2
		}
		if plan.Steps[i].TimeoutSec == 0 {
			plan.Steps[i].TimeoutSec = 30
		}
	}

	plan.ID = fmt.Sprintf("plan_%d", time.Now().UnixNano())
	plan.CreatedAt = time.Now().Unix()

	return plan, nil
}

// buildPlanningPrompt 构建计划生成 prompt
func buildPlanningPrompt(task string, skills []model.Skill) string {
	var sb strings.Builder
	sb.WriteString("你是一个任务规划专家。请根据用户任务和可用技能，生成结构化的执行计划。\n\n")
	sb.WriteString("可用技能:\n")
	for _, s := range skills {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", s.Name, s.Description))
	}

	sb.WriteString(fmt.Sprintf("\n用户任务: %s\n\n", task))
	sb.WriteString("请输出 JSON 格式的执行计划，支持以下字段:\n")
	sb.WriteString(`{
  "description": "计划描述",
  "parallel": false,
  "max_parallel": 3,
  "steps": [
    {
      "type": "skill",
      "skill": "技能名",
      "args": {"参数名": "参数值"},
      "depends_on": [],
      "description": "步骤描述",
      "max_retries": 2,
      "timeout_sec": 30
    }
  ]
}`)
	sb.WriteString("\n\n规则:\n")
	sb.WriteString("1. 每个步骤必须有唯一的 type（skill/handler/llm/condition）\n")
	sb.WriteString("2. depends_on 指定前置步骤 ID（可选）\n")
	sb.WriteString("3. 如果步骤可以并行执行，设置 parallel=true\n")
	sb.WriteString("4. 复杂任务拆分为多个小步骤\n")
	sb.WriteString("\n计划:")
	return sb.String()
}

// parsePlan 解析 LLM 输出的计划
func parsePlan(response string) (*model.Plan, error) {
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || start >= end {
		return nil, fmt.Errorf("no JSON found in response")
	}

	jsonStr := response[start : end+1]
	var plan model.Plan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return nil, fmt.Errorf("unmarshal plan: %w", err)
	}

	if len(plan.Steps) == 0 {
		return nil, fmt.Errorf("plan has no steps")
	}

	return &plan, nil
}

// ValidatePlan 验证计划的有效性
func ValidatePlan(plan *model.Plan) error {
	if len(plan.Steps) == 0 {
		return fmt.Errorf("plan has no steps")
	}

	stepIDs := make(map[string]bool)
	for _, step := range plan.Steps {
		if step.ID == "" {
			return fmt.Errorf("step has no ID")
		}
		if stepIDs[step.ID] {
			return fmt.Errorf("duplicate step ID: %s", step.ID)
		}
		stepIDs[step.ID] = true
	}

	// 验证依赖是否存在
	for _, step := range plan.Steps {
		for _, depID := range step.DependsOn {
			if !stepIDs[depID] {
				return fmt.Errorf("step %s depends on unknown step %s", step.ID, depID)
			}
		}
	}

	// 检查循环依赖
	if hasCycle(plan) {
		return fmt.Errorf("plan has circular dependencies")
	}

	return nil
}

// hasCycle 检查计划是否有循环依赖
func hasCycle(plan *model.Plan) bool {
	graph := make(map[string][]string)
	for _, step := range plan.Steps {
		graph[step.ID] = step.DependsOn
	}

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(string) bool
	dfs = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, dep := range graph[node] {
			if !visited[dep] {
				if dfs(dep) {
					return true
				}
			} else if recStack[dep] {
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for id := range graph {
		if !visited[id] {
			if dfs(id) {
				return true
			}
		}
	}
	return false
}
