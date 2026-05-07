package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Config Pipeline 配置
type Config struct {
	MaxConcurrency int           // 最大并发数（默认 5）
	BatchSize      int           // 批处理大小（默认 10）
	Timeout        time.Duration // 整体超时时间
	RetryCount     int           // 重试次数（默认 3）
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		MaxConcurrency: 5,
		BatchSize:      10,
		Timeout:        5 * time.Minute,
		RetryCount:     3,
	}
}

// Task 表示一个可执行任务
type Task struct {
	ID       string
	Priority int // 优先级（越高越先执行）
	Execute  func(ctx context.Context) (interface{}, error)
}

// Result 任务执行结果
type Result struct {
	TaskID   string
	Result   interface{}
	Error    error
	Duration time.Duration
	Attempt  int
}

// BatchResult 批处理结果
type BatchResult struct {
	Results   []*Result
	TotalTime time.Duration
	Success   int
	Failed    int
}

// Executor Pipeline 执行器
type Executor struct {
	config *Config
	sem    chan struct{} // 信号量控制并发
}

// NewExecutor 创建新的 Pipeline 执行器
func NewExecutor(config *Config) *Executor {
	if config == nil {
		config = DefaultConfig()
	}
	
	return &Executor{
		config: config,
		sem:    make(chan struct{}, config.MaxConcurrency),
	}
}

// ExecuteBatch 批量执行任务（并行，最多 MaxConcurrency 并发）
func (e *Executor) ExecuteBatch(ctx context.Context, tasks []*Task) (*BatchResult, error) {
	if len(tasks) == 0 {
		return &BatchResult{}, nil
	}

	// 按优先级排序（高优先级优先）
	sortedTasks := e.sortByPriority(tasks)

	// 分批处理
	var allResults []*Result
	startTime := time.Now()

	for i := 0; i < len(sortedTasks); i += e.config.BatchSize {
		end := i + e.config.BatchSize
		if end > len(sortedTasks) {
			end = len(sortedTasks)
		}
		
		batch := sortedTasks[i:end]
		batchResult, err := e.executeBatch(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("batch execution failed: %w", err)
		}
		
		allResults = append(allResults, batchResult.Results...)
	}

	totalTime := time.Since(startTime)
	
	// 统计成功/失败
	success := 0
	failed := 0
	for _, r := range allResults {
		if r.Error == nil {
			success++
		} else {
			failed++
		}
	}

	return &BatchResult{
		Results:   allResults,
		TotalTime: totalTime,
		Success:   success,
		Failed:    failed,
	}, nil
}

// executeBatch 执行一批任务
func (e *Executor) executeBatch(ctx context.Context, tasks []*Task) (*BatchResult, error) {
	var wg sync.WaitGroup
	results := make([]*Result, len(tasks))
	mu := sync.Mutex{}

	for i, task := range tasks {
		wg.Add(1)
		
		go func(index int, t *Task) {
			defer wg.Done()
			
			// 获取信号量（控制并发）
			e.sem <- struct{}{}
			defer func() { <-e.sem }()
			
			// 执行任务（带重试）
			result := e.executeWithRetry(ctx, t)
			
			mu.Lock()
			results[index] = result
			mu.Unlock()
		}(i, task)
	}

	wg.Wait()
	
	return &BatchResult{
		Results: results,
	}, nil
}

// executeWithRetry 执行单个任务（带重试）
func (e *Executor) executeWithRetry(ctx context.Context, task *Task) *Result {
	var lastErr error
	
	for attempt := 1; attempt <= e.config.RetryCount; attempt++ {
		startTime := time.Now()
		
		// 创建带超时的上下文
		taskCtx, cancel := context.WithTimeout(ctx, e.config.Timeout)
		result, err := task.Execute(taskCtx)
		cancel()
		
		duration := time.Since(startTime)
		
		if err == nil {
			return &Result{
				TaskID:   task.ID,
				Result:   result,
				Error:    nil,
				Duration: duration,
				Attempt:  attempt,
			}
		}
		
		lastErr = err
		
		// 如果不是最后一次尝试，等待后重试
		if attempt < e.config.RetryCount {
			backoff := time.Duration(attempt) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return &Result{
					TaskID:   task.ID,
					Result:   nil,
					Error:    ctx.Err(),
					Duration: duration,
					Attempt:  attempt,
				}
			case <-time.After(backoff):
				// 继续重试
			}
		}
	}
	
	return &Result{
		TaskID:   task.ID,
		Result:   nil,
		Error:    fmt.Errorf("failed after %d attempts: %w", e.config.RetryCount, lastErr),
		Duration: 0,
		Attempt:  e.config.RetryCount,
	}
}

// sortByPriority 按优先级排序（降序）
func (e *Executor) sortByPriority(tasks []*Task) []*Task {
	sorted := make([]*Task, len(tasks))
	copy(sorted, tasks)
	
	// 简单的冒泡排序（任务数量通常不大）
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Priority < sorted[j+1].Priority {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	
	return sorted
}

// GetStats 获取执行统计信息
func (br *BatchResult) GetStats() string {
	return fmt.Sprintf("Executed %d tasks: %d succeeded, %d failed in %v",
		len(br.Results), br.Success, br.Failed, br.TotalTime)
}

// GetFailedResults 获取失败的结果
func (br *BatchResult) GetFailedResults() []*Result {
	var failed []*Result
	for _, r := range br.Results {
		if r.Error != nil {
			failed = append(failed, r)
		}
	}
	return failed
}

// GetSuccessfulResults 获取成功的结果
func (br *BatchResult) GetSuccessfulResults() []*Result {
	var success []*Result
	for _, r := range br.Results {
		if r.Error == nil {
			success = append(success, r)
		}
	}
	return success
}
