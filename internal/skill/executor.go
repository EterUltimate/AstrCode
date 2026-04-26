package skill

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EterUltimate/astrcode/internal/model"
	"github.com/EterUltimate/astrcode/internal/sdk"
)

// EventSink 步骤事件回调
type EventSink func(event *model.WSEvent)

// Executor 负责执行 Skill Plan（Phase 4: 事件驱动 + 实时推送）
type Executor struct {
	sdkClient *sdk.AstrBotClient
	sink      EventSink
}

// NewExecutor 创建新的执行器
func NewExecutor(sdkClient *sdk.AstrBotClient) *Executor {
	return &Executor{sdkClient: sdkClient}
}

// NewExecutorWithSink 创建带事件推送的执行器
func NewExecutorWithSink(sdkClient *sdk.AstrBotClient, sink EventSink) *Executor {
	return &Executor{sdkClient: sdkClient, sink: sink}
}

// SetEventSink 设置事件回调
func (e *Executor) SetEventSink(sink EventSink) {
	e.sink = sink
}

// emit 发送事件
func (e *Executor) emit(eventType, taskID, planID, stepID string, data interface{}) {
	if e.sink == nil {
		return
	}
	evt := model.NewWSEvent(eventType, taskID, stepID, data)
	evt.PlanID = planID
	e.sink(evt)
}

// Execute 执行计划（支持依赖图、并行、事件推送）
func (e *Executor) Execute(ctx context.Context, plan *model.Plan) (*model.Task, error) {
	taskID := generateTaskID()
	task := &model.Task{
		ID:        taskID,
		Status:    model.TaskStatusExecuting,
		CreatedAt: time.Now().Unix(),
	}

	e.emit("task_created", taskID, plan.ID, "", map[string]string{
		"status":      "executing",
		"description": plan.Description,
	})

	graph := plan.BuildGraph()
	timeline := &model.StepTimeline{}

	startNano := time.Now().UnixNano()

	for !graph.IsComplete() {
		if graph.HasFailed() {
			task.Status = model.TaskStatusFailed
			task.Error = "plan execution failed"
			task.CompletedAt = time.Now().Unix()
			e.emit("task_failed", taskID, plan.ID, "", map[string]string{
				"error": task.Error,
			})
			return task, fmt.Errorf(task.Error)
		}

		readySteps := graph.GetReadySteps()
		if len(readySteps) == 0 {
			task.Error = "deadlock: no ready steps but plan not complete"
			task.Status = model.TaskStatusFailed
			task.CompletedAt = time.Now().Unix()
			e.emit("task_failed", taskID, plan.ID, "", map[string]string{
				"error": task.Error,
			})
			return task, fmt.Errorf(task.Error)
		}

		if plan.Parallel && plan.MaxParallel > 1 {
			var wg sync.WaitGroup
			sem := make(chan struct{}, plan.MaxParallel)

			for _, step := range readySteps {
				wg.Add(1)
				sem <- struct{}{}
				go func(s *model.Step) {
					defer wg.Done()
					defer func() { <-sem }()
					e.executeStepWithRetryAndEvents(ctx, s, taskID, plan.ID, timeline)
				}(step)
			}
			wg.Wait()
		} else {
			for _, step := range readySteps {
				e.executeStepWithRetryAndEvents(ctx, step, taskID, plan.ID, timeline)
			}
		}
	}

	totalMs := (time.Now().UnixNano() - startNano) / 1e6
	if graph.HasFailed() {
		task.Status = model.TaskStatusFailed
	} else {
		task.Status = model.TaskStatusCompleted
	}
	task.CompletedAt = time.Now().Unix()

	e.emit("task_completed", taskID, plan.ID, "", map[string]interface{}{
		"status":    task.Status,
		"duration":  totalMs,
		"timeline":  timeline.Items,
	})
	return task, nil
}

// executeStepWithRetryAndEvents 带重试和事件推送的步骤执行
func (e *Executor) executeStepWithRetryAndEvents(ctx context.Context, step *model.Step, taskID, planID string, timeline *model.StepTimeline) {
	step.Status = model.StepStatusRunning
	e.emit("step_started", taskID, planID, step.ID, map[string]interface{}{
		"type":   step.Type,
		"skill":  step.Skill,
		"status": "running",
	})
	timeline.Add(step.ID, "running", fmt.Sprintf("开始执行 %s: %s", step.Type, step.Skill))

	startNano := time.Now().UnixNano()

	for attempt := 0; attempt <= step.MaxRetries; attempt++ {
		step.RetryCount = attempt

		stepCtx, cancel := context.WithTimeout(ctx, time.Duration(step.TimeoutSec)*time.Second)
		result, err := e.executeStep(stepCtx, step)
		cancel()

		if err == nil {
			step.Status = model.StepStatusCompleted
			step.Result = result
			dur := (time.Now().UnixNano() - startNano) / 1e6
			timeline.AddWithDuration(step.ID, "completed", step.Skill, startNano)
			e.emit("step_completed", taskID, planID, step.ID, map[string]interface{}{
				"type":     step.Type,
				"skill":    step.Skill,
				"status":   "completed",
				"duration": dur,
				"attempt":  attempt + 1,
			})
			return
		}

		step.Error = err.Error()

		if attempt < step.MaxRetries {
			timeline.Add(step.ID, "retry", fmt.Sprintf("第 %d 次重试", attempt+1))
			e.emit("step_retry", taskID, planID, step.ID, map[string]interface{}{
				"attempt": attempt + 1,
				"max":     step.MaxRetries,
				"error":   err.Error(),
			})
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	step.Status = model.StepStatusFailed
	timeline.AddWithDuration(step.ID, "failed", step.Error, startNano)
	e.emit("step_failed", taskID, planID, step.ID, map[string]interface{}{
		"type":    step.Type,
		"skill":   step.Skill,
		"status":  "failed",
		"error":   step.Error,
		"retries": step.RetryCount + 1,
	})
}

// executeStep 执行单个步骤
func (e *Executor) executeStep(ctx context.Context, step *model.Step) (interface{}, error) {
	switch step.Type {
	case model.StepTypeSkill:
		return e.executeSkillStep(ctx, step)
	case model.StepTypeHandler:
		return e.executeHandlerStep(ctx, step)
	case model.StepTypeLLM:
		return e.executeLLMStep(ctx, step)
	case model.StepTypeCondition:
		return e.executeConditionStep(ctx, step)
	default:
		return nil, fmt.Errorf("unknown step type: %s", step.Type)
	}
}

func (e *Executor) executeSkillStep(ctx context.Context, step *model.Step) (interface{}, error) {
	result, err := e.sdkClient.CallSkill(ctx, step.Skill, step.Args)
	if err != nil {
		return nil, fmt.Errorf("skill %s failed: %w", step.Skill, err)
	}
	return result, nil
}

func (e *Executor) executeHandlerStep(ctx context.Context, step *model.Step) (interface{}, error) {
	event := &model.AstrMessageEvent{
		MessageStr: step.Skill,
		SessionID:  "astrcode",
		PlatformMeta: model.PlatformMetadata{
			ID:   "astrcode",
			Name: "AstrCode",
		},
	}
	result, err := e.sdkClient.CallHandler(ctx, step.Skill, event)
	if err != nil {
		return nil, fmt.Errorf("handler %s failed: %w", step.Skill, err)
	}
	return result, nil
}

func (e *Executor) executeLLMStep(ctx context.Context, step *model.Step) (interface{}, error) {
	return step.Args, nil
}

func (e *Executor) executeConditionStep(ctx context.Context, step *model.Step) (interface{}, error) {
	if step.Condition == nil {
		return nil, fmt.Errorf("condition step has no condition")
	}
	return step.Condition.ThenSteps, nil
}

// ExecuteHandler 直接调用 AstrBot Handler
func (e *Executor) ExecuteHandler(ctx context.Context, handlerFullName string, event *model.AstrMessageEvent) (*model.MessageEventResult, error) {
	return e.sdkClient.CallHandler(ctx, handlerFullName, event)
}

func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}
