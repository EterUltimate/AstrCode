package prompt

import (
	"fmt"
	"strings"

	"github.com/EterUltimate/astrcode/internal/model"
)

// Engine Prompt 引擎 — 三阶段 Prompt 管理
type Engine struct {
	systemPrompt string
}

// NewEngine 创建新的 Prompt 引擎
func NewEngine() *Engine {
	return &Engine{
		systemPrompt: "你是一个智能任务编排助手。你的职责是理解用户任务，选择合适的技能，并生成高效的执行计划。",
	}
}

// SetSystemPrompt 设置系统 Prompt
func (e *Engine) SetSystemPrompt(prompt string) {
	e.systemPrompt = prompt
}

// ============================================================
// Phase 1: 决策 Prompt（Decide）
// ============================================================

// DecisionOutput 决策输出
type DecisionOutput struct {
	Action       string   `json:"action"`              // "plan" | "direct" | "clarify"
	Skills       []string `json:"skills"`              // 选中的技能名
	Confidence   float64  `json:"confidence"`          // 置信度 0-1
	Reasoning    string   `json:"reasoning"`           // 决策理由
	NeedsClarify bool     `json:"needs_clarify"`       // 是否需要澄清
	Questions    []string `json:"questions,omitempty"` // 澄清问题
}

// BuildDecisionPrompt 构建决策 Prompt
func (e *Engine) BuildDecisionPrompt(task string, skills []model.Skill) string {
	var sb strings.Builder
	sb.WriteString(e.systemPrompt)
	sb.WriteString("\n\n## 决策阶段\n")
	sb.WriteString("分析用户任务，决定如何处理。\n\n")

	sb.WriteString("可用技能:\n")
	for _, s := range skills {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", s.Name, s.Description))
	}

	sb.WriteString(fmt.Sprintf("\n用户任务: %s\n\n", task))
	sb.WriteString("请输出 JSON 决策:\n")
	sb.WriteString(`{
  "action": "plan",
  "skills": ["skill1", "skill2"],
  "confidence": 0.9,
  "reasoning": "任务需要...",
  "needs_clarify": false,
  "questions": []
}`)
	sb.WriteString("\n\naction 说明:\n")
	sb.WriteString("- plan: 需要生成多步骤计划\n")
	sb.WriteString("- direct: 可以直接调用单个技能\n")
	sb.WriteString("- clarify: 任务不明确，需要用户澄清\n")

	return sb.String()
}

// ============================================================
// Phase 2: 规划 Prompt（Plan）
// ============================================================

// BuildPlanPrompt 构建规划 Prompt
func (e *Engine) BuildPlanPrompt(task string, skills []model.Skill, decision *DecisionOutput) string {
	var sb strings.Builder
	sb.WriteString(e.systemPrompt)
	sb.WriteString("\n\n## 规划阶段\n")
	sb.WriteString(fmt.Sprintf("任务: %s\n", task))
	sb.WriteString(fmt.Sprintf("决策: %s\n", decision.Reasoning))

	sb.WriteString("\n选中技能:\n")
	for _, s := range skills {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", s.Name, s.Description))
	}

	sb.WriteString("\n请生成结构化执行计划（JSON）:\n")
	sb.WriteString(`{
  "description": "计划描述",
  "steps": [
    {
      "id": "step_0",
      "type": "skill",
      "skill": "技能名",
      "args": {},
      "depends_on": [],
      "description": "步骤说明",
      "max_retries": 2,
      "timeout_sec": 30
    }
  ]
}`)

	return sb.String()
}

// ============================================================
// Phase 3: 调用/解析 Prompt（Call / Parse）
// ============================================================

// BuildSkillCallPrompt 构建 Skill 调用 Prompt
func (e *Engine) BuildSkillCallPrompt(step model.Step, context map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString("## 执行步骤\n")
	sb.WriteString(fmt.Sprintf("步骤: %s\n", step.Description))
	sb.WriteString(fmt.Sprintf("技能: %s\n", step.Skill))
	sb.WriteString(fmt.Sprintf("参数: %v\n", step.Args))

	if len(context) > 0 {
		sb.WriteString("\n上下文:\n")
		for k, v := range context {
			sb.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
		}
	}

	sb.WriteString("\n请执行此步骤并输出结果。")
	return sb.String()
}

// BuildParsePrompt 构建结果解析 Prompt
func (e *Engine) BuildParsePrompt(rawResult string, expectedFormat string) string {
	var sb strings.Builder
	sb.WriteString("## 结果解析\n")
	sb.WriteString(fmt.Sprintf("原始结果: %s\n", rawResult))
	sb.WriteString(fmt.Sprintf("期望格式: %s\n", expectedFormat))
	sb.WriteString("\n请解析并格式化结果。")
	return sb.String()
}

// ============================================================
// 工具方法
// ============================================================

// ParseDecision 解析 LLM 输出的决策 JSON
func (e *Engine) ParseDecision(response string) (*DecisionOutput, error) {
	// 提取 JSON
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || start >= end {
		return nil, fmt.Errorf("no JSON found in response")
	}

	// 这里简化处理，实际应该用 json.Unmarshal
	return &DecisionOutput{
		Action:     "plan",
		Confidence: 0.8,
		Reasoning:  response,
	}, nil
}

// ParseSkillCall 解析 Skill 调用
func (e *Engine) ParseSkillCall(response string) (string, map[string]interface{}, error) {
	parts := strings.Split(response, "|")
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("invalid skill call format")
	}

	skillName := strings.TrimSpace(parts[0])
	args := make(map[string]interface{})

	for i := 1; i < len(parts); i++ {
		kv := strings.SplitN(parts[i], "=", 2)
		if len(kv) == 2 {
			args[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	return skillName, args, nil
}

// FormatSkillsForPrompt 格式化技能列表用于 Prompt
func FormatSkillsForPrompt(skills []model.Skill) string {
	var sb strings.Builder
	for i, s := range skills {
		sb.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, s.Name, s.Description))
	}
	return sb.String()
}
