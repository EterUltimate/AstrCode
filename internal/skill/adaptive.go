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

// AdaptivePlanner 自适应规划器 — 动态调整 Plan
type AdaptivePlanner struct {
	planner   *Planner
	llmClient *llm.Client
}

// NewAdaptivePlanner 创建自适应规划器
func NewAdaptivePlanner(planner *Planner, llmClient *llm.Client) *AdaptivePlanner {
	return &AdaptivePlanner{
		planner:   planner,
		llmClient: llmClient,
	}
}

// ReplanOnFailure 步骤失败时重新规划
func (ap *AdaptivePlanner) ReplanOnFailure(ctx context.Context, originalPlan *model.Plan, failedStep *model.Step, failureError error, availableSkills []model.Skill) (*model.Plan, error) {
	prompt := buildReplanPrompt(originalPlan, failedStep, failureError, availableSkills)

	response, err := ap.llmClient.Complete(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("replan LLM call failed: %w", err)
	}

	newPlan, err := parsePlan(response)
	if err != nil {
		return nil, fmt.Errorf("parse replan: %w", err)
	}

	// 为每个步骤分配 ID 和默认值
	for i := range newPlan.Steps {
		newPlan.Steps[i].ID = fmt.Sprintf("step_replan_%d_%d", time.Now().Unix(), i)
		if newPlan.Steps[i].Status == "" {
			newPlan.Steps[i].Status = model.StepStatusPending
		}
		if newPlan.Steps[i].MaxRetries == 0 {
			newPlan.Steps[i].MaxRetries = 2
		}
		if newPlan.Steps[i].TimeoutSec == 0 {
			newPlan.Steps[i].TimeoutSec = 30
		}
	}

	newPlan.ID = fmt.Sprintf("plan_replan_%d", time.Now().UnixNano())
	newPlan.CreatedAt = time.Now().Unix()

	// 验证新计划
	if err := ValidatePlan(newPlan); err != nil {
		return nil, fmt.Errorf("replan validation failed: %w", err)
	}

	return newPlan, nil
}

// buildReplanPrompt 构建重新规划 Prompt
func buildReplanPrompt(originalPlan *model.Plan, failedStep *model.Step, failureError error, availableSkills []model.Skill) string {
	var sb strings.Builder
	sb.WriteString("你是一个自适应任务规划专家。原计划的某个步骤失败了，请生成替代执行方案。\n\n")

	sb.WriteString("原失败步骤:\n")
	sb.WriteString(fmt.Sprintf("- ID: %s\n", failedStep.ID))
	sb.WriteString(fmt.Sprintf("- 类型: %s\n", failedStep.Type))
	sb.WriteString(fmt.Sprintf("- 技能: %s\n", failedStep.Skill))
	sb.WriteString(fmt.Sprintf("- 参数: %v\n", failedStep.Args))
	sb.WriteString(fmt.Sprintf("- 错误: %v\n", failureError))

	sb.WriteString("\n可用技能:\n")
	for _, s := range availableSkills {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", s.Name, s.Description))
	}

	sb.WriteString("\n请输出 JSON 格式的替代计划:\n")
	sb.WriteString(`{
  "description": "替代方案描述",
  "steps": [
    {
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

// FallbackStrategy 降级策略
type FallbackStrategy string

const (
	FallbackRetry       FallbackStrategy = "retry"       // 重试
	FallbackSkip        FallbackStrategy = "skip"        // 跳过
	FallbackAlternative FallbackStrategy = "alternative" // 使用替代方案
	FallbackManual      FallbackStrategy = "manual"      // 转人工
)

// FallbackHandler 降级处理器
type FallbackHandler struct {
	strategies map[string]FallbackStrategy // step type -> strategy
}

// NewFallbackHandler 创建降级处理器
func NewFallbackHandler() *FallbackHandler {
	return &FallbackHandler{
		strategies: map[string]FallbackStrategy{
			string(model.StepTypeSkill):   FallbackRetry,
			string(model.StepTypeHandler): FallbackAlternative,
			string(model.StepTypeLLM):     FallbackRetry,
		},
	}
}

// SetStrategy 设置步骤类型的降级策略
func (fh *FallbackHandler) SetStrategy(stepType model.StepType, strategy FallbackStrategy) {
	fh.strategies[string(stepType)] = strategy
}

// HandleFailure 处理步骤失败
func (fh *FallbackHandler) HandleFailure(step *model.Step, err error) (FallbackStrategy, map[string]interface{}, error) {
	strategy, ok := fh.strategies[string(step.Type)]
	if !ok {
		strategy = FallbackRetry
	}

	switch strategy {
	case FallbackRetry:
		if step.RetryCount < step.MaxRetries {
			return FallbackRetry, nil, nil
		}
		// 重试耗尽，尝试替代方案
		return FallbackAlternative, nil, nil

	case FallbackSkip:
		return FallbackSkip, nil, nil

	case FallbackAlternative:
		// 尝试找到替代技能
		altArgs := make(map[string]interface{})
		altArgs["original_skill"] = step.Skill
		altArgs["original_args"] = step.Args
		altArgs["error"] = err.Error()
		return FallbackAlternative, altArgs, nil

	case FallbackManual:
		return FallbackManual, nil, fmt.Errorf("step %s failed and requires manual intervention: %v", step.ID, err)
	}

	return FallbackRetry, nil, nil
}

// AdaptiveExecutor 自适应执行器
type AdaptiveExecutor struct {
	*Executor
	adaptivePlanner *AdaptivePlanner
	fallbackHandler *FallbackHandler
}

// NewAdaptiveExecutor 创建自适应执行器
func NewAdaptiveExecutor(sdkClient interface{}, adaptivePlanner *AdaptivePlanner, fallbackHandler *FallbackHandler) *AdaptiveExecutor {
	// 这里简化处理，实际需要类型断言
	return &AdaptiveExecutor{
		adaptivePlanner: adaptivePlanner,
		fallbackHandler: fallbackHandler,
	}
}

// ExecuteWithFallback 带降级策略的执行
func (ae *AdaptiveExecutor) ExecuteWithFallback(ctx context.Context, plan *model.Plan, availableSkills []model.Skill) (*model.Task, error) {
	task := &model.Task{
		ID:        generateTaskID(),
		Status:    model.TaskStatusExecuting,
		CreatedAt: time.Now().Unix(),
	}

	graph := plan.BuildGraph()

	for !graph.IsComplete() {
		if graph.HasFailed() {
			// 尝试降级处理
			for _, step := range graph.Nodes {
				if step.Status == model.StepStatusFailed {
					strategy, altArgs, err := ae.fallbackHandler.HandleFailure(step, fmt.Errorf(step.Error))
					if err != nil {
						task.Status = model.TaskStatusFailed
						task.Error = err.Error()
						task.CompletedAt = time.Now().Unix()
						return task, err
					}

					switch strategy {
					case FallbackRetry:
						step.Status = model.StepStatusPending
						step.RetryCount = 0

					case FallbackSkip:
						step.Status = model.StepStatusSkipped

					case FallbackAlternative:
						// 重新规划
						newPlan, replanErr := ae.adaptivePlanner.ReplanOnFailure(ctx, plan, step, fmt.Errorf(step.Error), availableSkills)
						if replanErr != nil {
							// 重新规划失败，跳过此步骤
							step.Status = model.StepStatusSkipped
						} else {
							// 使用新计划继续执行
							// 简化：将新步骤加入当前图
							_ = newPlan
							_ = altArgs
							step.Status = model.StepStatusSkipped
						}
					}
				}
			}
		}

		readySteps := graph.GetReadySteps()
		if len(readySteps) == 0 {
			break
		}

		// 串行执行（简化版）
		for _, step := range readySteps {
			// 这里需要调用实际的 executeStep，简化处理
			_ = step
		}
	}

	task.Status = model.TaskStatusCompleted
	task.CompletedAt = time.Now().Unix()
	return task, nil
}

// StepResult 步骤执行结果
type StepResult struct {
	StepID    string
	Success   bool
	Result    interface{}
	Error     error
	Duration  time.Duration
	RetryUsed int
}

// ExecutionReport 执行报告
type ExecutionReport struct {
	TaskID      string        `json:"task_id"`
	PlanID      string        `json:"plan_id"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`
	StepResults []StepResult  `json:"step_results"`
	Status      string        `json:"status"`
	Error       string        `json:"error,omitempty"`
}

// GenerateReport 生成执行报告
func GenerateReport(taskID, planID string, results []StepResult, startTime time.Time) *ExecutionReport {
	endTime := time.Now()
	status := "completed"
	var errStr string

	for _, r := range results {
		if !r.Success {
			status = "failed"
			if r.Error != nil {
				errStr = r.Error.Error()
			}
			break
		}
	}

	return &ExecutionReport{
		TaskID:      taskID,
		PlanID:      planID,
		StartTime:   startTime,
		EndTime:     endTime,
		Duration:    endTime.Sub(startTime),
		StepResults: results,
		Status:      status,
		Error:       errStr,
	}
}

// ToJSON 将报告转为 JSON
func (r *ExecutionReport) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}
