package model

// StepType 步骤类型
type StepType string

const (
	StepTypeSkill   StepType = "skill"   // 调用 Skill
	StepTypeHandler StepType = "handler" // 调用 AstrBot Handler
	StepTypeLLM     StepType = "llm"     // LLM 推理
	StepTypeCondition StepType = "condition" // 条件分支
)

// StepStatus 步骤状态
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
)

// Step 计划步骤（增强版）
type Step struct {
	ID           string                 `json:"id"`
	Type         StepType               `json:"type"`
	Skill        string                 `json:"skill,omitempty"`    // Skill/Handler 名称
	Args         map[string]interface{} `json:"args,omitempty"`
	DependsOn    []string               `json:"depends_on,omitempty"` // 依赖的步骤 ID
	Condition    *Condition             `json:"condition,omitempty"`  // 条件分支
	Description  string                 `json:"description,omitempty"`
	Status       StepStatus             `json:"status"`
	Result       interface{}            `json:"result,omitempty"`
	Error        string                 `json:"error,omitempty"`
	RetryCount   int                    `json:"retry_count"`
	MaxRetries   int                    `json:"max_retries"`
	TimeoutSec   int                    `json:"timeout_sec"`
}

// Condition 条件分支
type Condition struct {
	Expression string `json:"expression"` // 条件表达式
	ThenSteps  []Step `json:"then_steps"` // 条件为真时执行
	ElseSteps  []Step `json:"else_steps,omitempty"` // 条件为假时执行
}

// Plan 任务执行计划（增强版）
type Plan struct {
	ID          string   `json:"id"`
	Description string   `json:"description,omitempty"`
	Steps       []Step   `json:"steps"`
	Parallel    bool     `json:"parallel,omitempty"` // 是否支持并行
	MaxParallel int      `json:"max_parallel,omitempty"`
	CreatedAt   int64    `json:"created_at"`
}

// PlanGraph 计划依赖图
type PlanGraph struct {
	Nodes map[string]*Step      `json:"nodes"`
	Edges map[string][]string   `json:"edges"` // step_id -> 依赖它的步骤 IDs
}

// BuildGraph 从 Plan 构建依赖图
func (p *Plan) BuildGraph() *PlanGraph {
	graph := &PlanGraph{
		Nodes: make(map[string]*Step),
		Edges: make(map[string][]string),
	}

	for i := range p.Steps {
		step := &p.Steps[i]
		graph.Nodes[step.ID] = step
	}

	for _, step := range p.Steps {
		for _, depID := range step.DependsOn {
			if graph.Edges[depID] == nil {
				graph.Edges[depID] = []string{}
			}
			graph.Edges[depID] = append(graph.Edges[depID], step.ID)
		}
	}

	return graph
}

// GetReadySteps 获取可以执行的步骤（所有依赖已完成）
func (g *PlanGraph) GetReadySteps() []*Step {
	var ready []*Step
	for _, step := range g.Nodes {
		if step.Status != StepStatusPending {
			continue
		}
		allDepsDone := true
		for _, depID := range step.DependsOn {
			dep := g.Nodes[depID]
			if dep == nil || dep.Status != StepStatusCompleted {
				allDepsDone = false
				break
			}
		}
		if allDepsDone {
			ready = append(ready, step)
		}
	}
	return ready
}

// IsComplete 检查计划是否完成
func (g *PlanGraph) IsComplete() bool {
	for _, step := range g.Nodes {
		if step.Status == StepStatusPending || step.Status == StepStatusRunning {
			return false
		}
	}
	return true
}

// HasFailed 检查是否有失败步骤
func (g *PlanGraph) HasFailed() bool {
	for _, step := range g.Nodes {
		if step.Status == StepStatusFailed {
			return true
		}
	}
	return false
}
