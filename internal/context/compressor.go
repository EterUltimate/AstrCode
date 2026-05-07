package context

import (
	"fmt"
	"strings"
)

// Compressor 实现上下文压缩
type Compressor struct {
	maxTokens      int     // 最大 token 数
	thresholdRatio float64 // 触发压缩的阈值比例（默认 0.8）
}

// NewCompressor 创建压缩器
func NewCompressor(maxTokens int, thresholdRatio float64) *Compressor {
	if thresholdRatio <= 0 || thresholdRatio > 1 {
		thresholdRatio = 0.8
	}

	return &Compressor{
		maxTokens:      maxTokens,
		thresholdRatio: thresholdRatio,
	}
}

// CompressionResult 压缩结果
type CompressionResult struct {
	OriginalTokens   int     `json:"original_tokens"`
	CompressedTokens int     `json:"compressed_tokens"`
	CompressionRatio float64 `json:"compression_ratio"`
	RemovedMessages  []int   `json:"removed_messages"`  // 被移除的消息索引
	Summary          string  `json:"summary,omitempty"` // 生成的摘要
}

// ShouldCompress 检查是否需要压缩
func (c *Compressor) ShouldCompress(currentTokens int) bool {
	threshold := int(float64(c.maxTokens) * c.thresholdRatio)
	return currentTokens >= threshold
}

// CompressMessages 压缩消息列表
func (c *Compressor) CompressMessages(messages []Message, currentTokens int) (*CompressionResult, error) {
	if !c.ShouldCompress(currentTokens) {
		return nil, fmt.Errorf("no compression needed: %d tokens < threshold %d",
			currentTokens, int(float64(c.maxTokens)*c.thresholdRatio))
	}

	result := &CompressionResult{
		OriginalTokens: currentTokens,
	}

	// 策略 1: 移除最旧的非系统消息
	compressedMessages, removedIndices := c.removeOldestMessages(messages)
	result.RemovedMessages = removedIndices

	// 估算压缩后的 token 数
	result.CompressedTokens = c.estimateTokens(compressedMessages)

	if result.CompressedTokens > 0 {
		result.CompressionRatio = float64(result.CompressedTokens) / float64(result.OriginalTokens)
	}

	return result, nil
}

// CompressWithSummary 使用摘要进行压缩（需要 LLM 支持）
func (c *Compressor) CompressWithSummary(messages []Message, currentTokens int, summaryGenerator func([]Message) (string, error)) (*CompressionResult, error) {
	if !c.ShouldCompress(currentTokens) {
		return nil, fmt.Errorf("no compression needed")
	}

	result := &CompressionResult{
		OriginalTokens: currentTokens,
	}

	// 保留系统消息和最近的消息
	var messagesToSummarize []Message
	var preservedMessages []Message

	for i, msg := range messages {
		if msg.Role == "system" || i >= len(messages)-3 {
			preservedMessages = append(preservedMessages, msg)
		} else {
			messagesToSummarize = append(messagesToSummarize, msg)
		}
	}

	// 生成摘要
	if len(messagesToSummarize) > 0 && summaryGenerator != nil {
		summary, err := summaryGenerator(messagesToSummarize)
		if err != nil {
			return nil, fmt.Errorf("failed to generate summary: %w", err)
		}
		result.Summary = summary

		// 将摘要作为第一条消息
		summaryMessage := Message{
			Role:    "assistant",
			Content: fmt.Sprintf("[Previous conversation summary]\n%s", summary),
		}

		finalMessages := append([]Message{summaryMessage}, preservedMessages...)
		result.CompressedTokens = c.estimateTokens(finalMessages)
	} else {
		result.CompressedTokens = c.estimateTokens(preservedMessages)
	}

	if result.CompressedTokens > 0 {
		result.CompressionRatio = float64(result.CompressedTokens) / float64(result.OriginalTokens)
	}

	return result, nil
}

// removeOldestMessages 移除最旧的 N 条非系统消息
func (c *Compressor) removeOldestMessages(messages []Message) ([]Message, []int) {
	var preserved []Message
	var removed []int

	systemMessages := 0
	for _, msg := range messages {
		if msg.Role == "system" {
			systemMessages++
		}
	}

	// 至少保留系统消息和最近的 2 条消息
	minPreserve := systemMessages + 2

	if len(messages) <= minPreserve {
		return messages, nil
	}

	// 计算需要移除的数量
	toRemove := len(messages) - minPreserve

	// 标记要移除的消息索引
	for i := 0; i < toRemove; i++ {
		if messages[i].Role != "system" {
			removed = append(removed, i)
		}
	}

	// 构建保留的消息列表
	for i, msg := range messages {
		isRemoved := false
		for _, idx := range removed {
			if i == idx {
				isRemoved = true
				break
			}
		}
		if !isRemoved {
			preserved = append(preserved, msg)
		}
	}

	return preserved, removed
}

// estimateTokens 估算 token 数量（简化版本，实际应使用 tokenizer）
func (c *Compressor) estimateTokens(messages []Message) int {
	totalTokens := 0
	for _, msg := range messages {
		// 粗略估算：每 4 个字符约 1 个 token
		tokens := len(strings.Fields(msg.Content)) + 10 // +10 for role and metadata
		totalTokens += tokens
	}
	return totalTokens
}

// Message 表示一条消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GetCompressionStats 获取压缩统计信息
func (r *CompressionResult) GetCompressionStats() string {
	return fmt.Sprintf("Compression: %d → %d tokens (%.1f%% reduction, ratio: %.2f)",
		r.OriginalTokens,
		r.CompressedTokens,
		(1-r.CompressionRatio)*100,
		r.CompressionRatio,
	)
}
