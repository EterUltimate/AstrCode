package context

import (
	"testing"
)

func TestCompressor_ShouldCompress(t *testing.T) {
	compressor := NewCompressor(1000, 0.8)

	tests := []struct {
		name     string
		tokens   int
		expected bool
	}{
		{"Below threshold", 700, false},
		{"At threshold", 800, true},
		{"Above threshold", 900, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compressor.ShouldCompress(tt.tokens)
			if result != tt.expected {
				t.Errorf("ShouldCompress(%d) = %v, want %v", tt.tokens, result, tt.expected)
			}
		})
	}
}

func TestCompressor_CompressMessages(t *testing.T) {
	compressor := NewCompressor(1000, 0.8)

	messages := []Message{
		{Role: "system", Content: "You are a helpful assistant"},
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
		{Role: "user", Content: "How are you?"},
		{Role: "assistant", Content: "I'm good, thanks!"},
		{Role: "user", Content: "What's the weather?"},
		{Role: "assistant", Content: "It's sunny today"},
	}

	// 估算 token 数（应该超过阈值）
	currentTokens := 850

	result, err := compressor.CompressMessages(messages, currentTokens)
	if err != nil {
		t.Fatalf("CompressMessages failed: %v", err)
	}

	if result.OriginalTokens != currentTokens {
		t.Errorf("Expected original tokens %d, got %d", currentTokens, result.OriginalTokens)
	}

	if len(result.RemovedMessages) == 0 {
		t.Error("Expected some messages to be removed")
	}

	t.Logf("Compression stats: %s", result.GetCompressionStats())
}

func TestCompressor_NoCompressionNeeded(t *testing.T) {
	compressor := NewCompressor(1000, 0.8)

	messages := []Message{
		{Role: "user", Content: "Hello"},
	}

	_, err := compressor.CompressMessages(messages, 100)
	if err == nil {
		t.Error("Expected error when compression not needed")
	}
}

func TestCompressor_RemoveOldestMessages(t *testing.T) {
	compressor := NewCompressor(1000, 0.8)

	messages := []Message{
		{Role: "system", Content: "System message"},
		{Role: "user", Content: "Old message 1"},
		{Role: "assistant", Content: "Old response 1"},
		{Role: "user", Content: "Old message 2"},
		{Role: "assistant", Content: "Old response 2"},
		{Role: "user", Content: "Recent message"},
		{Role: "assistant", Content: "Recent response"},
	}

	preserved, removed := compressor.removeOldestMessages(messages)

	// 应该保留系统消息 + 最近 2 条 = 3 条（但实际可能保留更多，取决于实现）
	expectedMinPreserved := 3
	if len(preserved) < expectedMinPreserved {
		t.Errorf("Expected at least %d preserved messages, got %d", expectedMinPreserved, len(preserved))
	}

	// 验证系统消息被保留
	if preserved[0].Role != "system" {
		t.Error("System message should be preserved")
	}

	t.Logf("Removed %d messages at indices: %v", len(removed), removed)
}

func TestCompressor_EstimateTokens(t *testing.T) {
	compressor := NewCompressor(1000, 0.8)

	messages := []Message{
		{Role: "user", Content: "Hello world"},
		{Role: "assistant", Content: "Hi there! How can I help you today?"},
	}

	tokens := compressor.estimateTokens(messages)

	if tokens <= 0 {
		t.Error("Expected positive token count")
	}

	t.Logf("Estimated tokens: %d", tokens)
}

func TestCompressor_CompressionRatio(t *testing.T) {
	compressor := NewCompressor(1000, 0.8)

	// 创建大量消息
	var messages []Message
	for i := 0; i < 20; i++ {
		messages = append(messages, Message{
			Role:    "user",
			Content: "This is a test message number " + string(rune('0'+i)),
		})
	}

	currentTokens := 900
	result, err := compressor.CompressMessages(messages, currentTokens)
	if err != nil {
		t.Fatalf("CompressMessages failed: %v", err)
	}

	if result.CompressionRatio >= 1.0 {
		t.Error("Compression ratio should be less than 1.0")
	}

	if result.CompressionRatio <= 0 {
		t.Error("Compression ratio should be greater than 0")
	}

	t.Logf("Compression ratio: %.2f", result.CompressionRatio)
}

func TestCompressor_PreserveSystemMessages(t *testing.T) {
	compressor := NewCompressor(1000, 0.8)

	messages := []Message{
		{Role: "system", Content: "System instruction 1"},
		{Role: "system", Content: "System instruction 2"},
		{Role: "user", Content: "User message 1"},
		{Role: "assistant", Content: "Assistant response 1"},
		{Role: "user", Content: "User message 2"},
		{Role: "assistant", Content: "Assistant response 2"},
	}

	preserved, _ := compressor.removeOldestMessages(messages)

	// 验证所有系统消息都被保留
	systemCount := 0
	for _, msg := range preserved {
		if msg.Role == "system" {
			systemCount++
		}
	}

	if systemCount != 2 {
		t.Errorf("Expected 2 system messages preserved, got %d", systemCount)
	}
}

func TestCompressionResult_GetCompressionStats(t *testing.T) {
	result := &CompressionResult{
		OriginalTokens:   1000,
		CompressedTokens: 600,
		CompressionRatio: 0.6,
	}

	stats := result.GetCompressionStats()

	if stats == "" {
		t.Error("Expected non-empty stats string")
	}

	t.Logf("Stats: %s", stats)
}
