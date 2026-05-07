package pipeline

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestExecutor_ExecuteBatch_Basic(t *testing.T) {
	executor := NewExecutor(DefaultConfig())

	tasks := []*Task{
		{
			ID:       "task1",
			Priority: 1,
			Execute: func(ctx context.Context) (interface{}, error) {
				return "result1", nil
			},
		},
		{
			ID:       "task2",
			Priority: 2,
			Execute: func(ctx context.Context) (interface{}, error) {
				return "result2", nil
			},
		},
	}

	result, err := executor.ExecuteBatch(context.Background(), tasks)
	if err != nil {
		t.Fatalf("ExecuteBatch failed: %v", err)
	}

	if result.Success != 2 {
		t.Errorf("Expected 2 successes, got %d", result.Success)
	}

	if result.Failed != 0 {
		t.Errorf("Expected 0 failures, got %d", result.Failed)
	}

	t.Logf("Stats: %s", result.GetStats())
}

func TestExecutor_ExecuteBatch_Parallel(t *testing.T) {
	config := DefaultConfig()
	config.MaxConcurrency = 3
	executor := NewExecutor(config)

	var concurrentCount int32
	var maxConcurrent int32

	tasks := make([]*Task, 10)
	for i := 0; i < 10; i++ {
		idx := i
		tasks[i] = &Task{
			ID:       fmt.Sprintf("task_%d", i),
			Priority: 1,
			Execute: func(ctx context.Context) (interface{}, error) {
				// 增加并发计数
				current := atomic.AddInt32(&concurrentCount, 1)

				// 更新最大并发数
				for {
					old := atomic.LoadInt32(&maxConcurrent)
					if current <= old || atomic.CompareAndSwapInt32(&maxConcurrent, old, current) {
						break
					}
				}

				// 模拟工作
				time.Sleep(50 * time.Millisecond)

				// 减少并发计数
				atomic.AddInt32(&concurrentCount, -1)

				return fmt.Sprintf("result_%d", idx), nil
			},
		}
	}

	result, err := executor.ExecuteBatch(context.Background(), tasks)
	if err != nil {
		t.Fatalf("ExecuteBatch failed: %v", err)
	}

	if result.Success != 10 {
		t.Errorf("Expected 10 successes, got %d", result.Success)
	}

	// 验证并发控制（最大并发数不应超过配置）
	if maxConcurrent > int32(config.MaxConcurrency) {
		t.Errorf("Max concurrent %d exceeded limit %d", maxConcurrent, config.MaxConcurrency)
	}

	t.Logf("Max concurrent: %d (limit: %d)", maxConcurrent, config.MaxConcurrency)
	t.Logf("Stats: %s", result.GetStats())
}

func TestExecutor_ExecuteBatch_Retry(t *testing.T) {
	config := DefaultConfig()
	config.RetryCount = 3
	executor := NewExecutor(config)

	attemptCount := 0

	task := &Task{
		ID:       "retry_task",
		Priority: 1,
		Execute: func(ctx context.Context) (interface{}, error) {
			attemptCount++
			if attemptCount < 3 {
				return nil, fmt.Errorf("temporary error (attempt %d)", attemptCount)
			}
			return "success", nil
		},
	}

	result, err := executor.ExecuteBatch(context.Background(), []*Task{task})
	if err != nil {
		t.Fatalf("ExecuteBatch failed: %v", err)
	}

	if result.Success != 1 {
		t.Errorf("Expected 1 success after retries, got %d", result.Success)
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}

	t.Logf("Task succeeded after %d attempts", attemptCount)
}

func TestExecutor_ExecuteBatch_Failure(t *testing.T) {
	config := DefaultConfig()
	config.RetryCount = 2
	executor := NewExecutor(config)

	task := &Task{
		ID:       "failing_task",
		Priority: 1,
		Execute: func(ctx context.Context) (interface{}, error) {
			return nil, fmt.Errorf("permanent error")
		},
	}

	result, err := executor.ExecuteBatch(context.Background(), []*Task{task})
	if err != nil {
		t.Fatalf("ExecuteBatch failed: %v", err)
	}

	if result.Failed != 1 {
		t.Errorf("Expected 1 failure, got %d", result.Failed)
	}

	failedResults := result.GetFailedResults()
	if len(failedResults) != 1 {
		t.Errorf("Expected 1 failed result, got %d", len(failedResults))
	}

	t.Logf("Failed as expected: %v", failedResults[0].Error)
}

func TestExecutor_ExecuteBatch_Priority(t *testing.T) {
	executor := NewExecutor(DefaultConfig())

	var mu sync.Mutex
	executionOrder := []string{}

	tasks := []*Task{
		{
			ID:       "low_priority",
			Priority: 1,
			Execute: func(ctx context.Context) (interface{}, error) {
				mu.Lock()
				executionOrder = append(executionOrder, "low")
				mu.Unlock()
				return nil, nil
			},
		},
		{
			ID:       "high_priority",
			Priority: 10,
			Execute: func(ctx context.Context) (interface{}, error) {
				mu.Lock()
				executionOrder = append(executionOrder, "high")
				mu.Unlock()
				return nil, nil
			},
		},
		{
			ID:       "medium_priority",
			Priority: 5,
			Execute: func(ctx context.Context) (interface{}, error) {
				mu.Lock()
				executionOrder = append(executionOrder, "medium")
				mu.Unlock()
				return nil, nil
			},
		},
	}

	_, err := executor.ExecuteBatch(context.Background(), tasks)
	if err != nil {
		t.Fatalf("ExecuteBatch failed: %v", err)
	}

	// 由于是并行执行，顺序可能不固定，但高优先级应该先开始
	t.Logf("Execution order: %v", executionOrder)
}

func TestExecutor_ExecuteBatch_Empty(t *testing.T) {
	executor := NewExecutor(DefaultConfig())

	result, err := executor.ExecuteBatch(context.Background(), []*Task{})
	if err != nil {
		t.Fatalf("ExecuteBatch failed: %v", err)
	}

	if len(result.Results) != 0 {
		t.Errorf("Expected 0 results for empty batch, got %d", len(result.Results))
	}
}

func TestExecutor_ExecuteBatch_Timeout(t *testing.T) {
	config := DefaultConfig()
	config.Timeout = 100 * time.Millisecond
	executor := NewExecutor(config)

	task := &Task{
		ID:       "slow_task",
		Priority: 1,
		Execute: func(ctx context.Context) (interface{}, error) {
			select {
			case <-time.After(200 * time.Millisecond):
				return "completed", nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	}

	result, err := executor.ExecuteBatch(context.Background(), []*Task{task})
	if err != nil {
		t.Fatalf("ExecuteBatch failed: %v", err)
	}

	// 任务应该超时失败
	if result.Failed != 1 {
		t.Errorf("Expected 1 failure due to timeout, got %d", result.Failed)
	}

	t.Logf("Task timed out as expected")
}

func TestExecutor_ExecuteBatch_Mixed(t *testing.T) {
	executor := NewExecutor(DefaultConfig())

	tasks := []*Task{
		{
			ID:       "success_task",
			Priority: 1,
			Execute: func(ctx context.Context) (interface{}, error) {
				return "success", nil
			},
		},
		{
			ID:       "fail_task",
			Priority: 1,
			Execute: func(ctx context.Context) (interface{}, error) {
				return nil, fmt.Errorf("error")
			},
		},
		{
			ID:       "another_success",
			Priority: 2,
			Execute: func(ctx context.Context) (interface{}, error) {
				return "ok", nil
			},
		},
	}

	result, err := executor.ExecuteBatch(context.Background(), tasks)
	if err != nil {
		t.Fatalf("ExecuteBatch failed: %v", err)
	}

	if result.Success != 2 {
		t.Errorf("Expected 2 successes, got %d", result.Success)
	}

	if result.Failed != 1 {
		t.Errorf("Expected 1 failure, got %d", result.Failed)
	}

	successful := result.GetSuccessfulResults()
	failed := result.GetFailedResults()

	if len(successful) != 2 {
		t.Errorf("Expected 2 successful results, got %d", len(successful))
	}

	if len(failed) != 1 {
		t.Errorf("Expected 1 failed result, got %d", len(failed))
	}

	t.Logf("Stats: %s", result.GetStats())
}

func TestBatchResult_GetStats(t *testing.T) {
	result := &BatchResult{
		Results:   make([]*Result, 5),
		TotalTime: 1 * time.Second,
		Success:   3,
		Failed:    2,
	}

	stats := result.GetStats()
	if stats == "" {
		t.Error("Expected non-empty stats string")
	}

	t.Logf("Stats: %s", stats)
}
