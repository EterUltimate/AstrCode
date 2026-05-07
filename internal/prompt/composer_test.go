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

package prompt

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/EterUltimate/astrcode/internal/model"
)

func TestPromptCache_Basic(t *testing.T) {
	cache := NewPromptCache(5 * time.Minute)

	block := PromptBlock{
		ID:         "test_block",
		Content:    "Test content",
		Priority:   100,
		CacheLevel: CacheLevelStable,
	}

	// 设置缓存
	cache.Set("test_key", block)

	// 获取缓存
	cached, exists := cache.Get("test_key")
	if !exists {
		t.Error("Expected block to exist in cache")
	}

	if cached.ID != block.ID {
		t.Errorf("Expected ID '%s', got '%s'", block.ID, cached.ID)
	}

	if cached.Content != block.Content {
		t.Errorf("Expected content '%s', got '%s'", block.Content, cached.Content)
	}
}

func TestPromptCache_Expiration(t *testing.T) {
	cache := NewPromptCache(100 * time.Millisecond) // 很短的 TTL

	block := PromptBlock{
		ID:      "test_block",
		Content: "Test content",
	}

	cache.Set("test_key", block)

	// 立即获取应该存在
	_, exists := cache.Get("test_key")
	if !exists {
		t.Error("Expected block to exist immediately after set")
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 再次获取应该不存在
	_, exists = cache.Get("test_key")
	if exists {
		t.Error("Expected block to be expired")
	}
}

func TestPromptCache_Invalidate(t *testing.T) {
	cache := NewPromptCache(5 * time.Minute)

	block := PromptBlock{
		ID:      "test_block",
		Content: "Test content",
	}

	cache.Set("test_key", block)

	// 使缓存失效
	cache.Invalidate("test_key")

	// 获取应该不存在
	_, exists := cache.Get("test_key")
	if exists {
		t.Error("Expected block to be invalidated")
	}
}

func TestPromptCache_Clear(t *testing.T) {
	cache := NewPromptCache(5 * time.Minute)

	// 添加多个块
	for i := 0; i < 5; i++ {
		key := string(rune('a' + i))
		block := PromptBlock{
			ID:      key,
			Content: "Content " + key,
		}
		cache.Set(key, block)
	}

	// 清空缓存
	cache.Clear()

	// 检查缓存大小
	stats := map[string]interface{}{"cache_size": len(cache.cache)}
	if stats["cache_size"] != 0 {
		t.Errorf("Expected cache size 0 after clear, got %v", stats["cache_size"])
	}
}

func TestComposer_RegisterContributor(t *testing.T) {
	composer := NewComposer(nil)

	contributor := NewSystemPromptContributor("Test system prompt")
	composer.RegisterContributor(contributor)

	stats := composer.GetStats()
	if stats["contributors"] != 1 {
		t.Errorf("Expected 1 contributor, got %v", stats["contributors"])
	}
}

func TestComposer_Assemble_Basic(t *testing.T) {
	composer := NewComposer(nil)

	// 注册系统提示贡献者
	systemContributor := NewSystemPromptContributor("You are an AI assistant.")
	composer.RegisterContributor(systemContributor)

	// 注册任务贡献者
	taskContributor := NewTaskContributor("Write a Python function")
	composer.RegisterContributor(taskContributor)

	// 组装 prompt
	ctx := context.Background()
	prompt, err := composer.Assemble(ctx)
	if err != nil {
		t.Fatalf("Failed to assemble prompt: %v", err)
	}

	// 验证内容
	if !strings.Contains(prompt, "You are an AI assistant.") {
		t.Error("Expected prompt to contain system prompt")
	}

	if !strings.Contains(prompt, "Write a Python function") {
		t.Error("Expected prompt to contain task description")
	}

	// 验证顺序（系统提示应该在前面，因为优先级更高）
	systemIdx := strings.Index(prompt, "You are an AI assistant.")
	taskIdx := strings.Index(prompt, "Write a Python function")

	if systemIdx > taskIdx {
		t.Error("Expected system prompt to appear before task description")
	}
}

func TestComposer_Assemble_WithSkills(t *testing.T) {
	composer := NewComposer(nil)

	// 注册技能贡献者
	skills := []model.Skill{
		{Name: "skill1", Description: "First skill"},
		{Name: "skill2", Description: "Second skill"},
	}
	skillsContributor := NewSkillsContributor(skills)
	composer.RegisterContributor(skillsContributor)

	ctx := context.Background()
	prompt, err := composer.Assemble(ctx)
	if err != nil {
		t.Fatalf("Failed to assemble prompt: %v", err)
	}

	if !strings.Contains(prompt, "## Available Skills") {
		t.Error("Expected prompt to contain skills section")
	}

	if !strings.Contains(prompt, "skill1") || !strings.Contains(prompt, "skill2") {
		t.Error("Expected prompt to contain skill names")
	}
}

func TestComposer_Assemble_WithContext(t *testing.T) {
	composer := NewComposer(nil)

	// 注册上下文贡献者
	messages := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}
	contextContributor := NewContextContributor(messages, 10)
	composer.RegisterContributor(contextContributor)

	ctx := context.Background()
	prompt, err := composer.Assemble(ctx)
	if err != nil {
		t.Fatalf("Failed to assemble prompt: %v", err)
	}

	if !strings.Contains(prompt, "## Conversation History") {
		t.Error("Expected prompt to contain conversation history section")
	}

	if !strings.Contains(prompt, "Hello") || !strings.Contains(prompt, "Hi there!") {
		t.Error("Expected prompt to contain message content")
	}
}

func TestComposer_Assemble_ConditionalBlock(t *testing.T) {
	composer := NewComposer(nil)

	// 注册条件指令贡献者
	conditionMet := true
	instructionsContributor := NewInstructionsContributor(
		"Special instructions",
		func(ctx context.Context) bool {
			return conditionMet
		},
	)
	composer.RegisterContributor(instructionsContributor)

	ctx := context.Background()
	prompt, err := composer.Assemble(ctx)
	if err != nil {
		t.Fatalf("Failed to assemble prompt: %v", err)
	}

	if !strings.Contains(prompt, "Special instructions") {
		t.Error("Expected prompt to contain conditional instructions when condition is met")
	}

	// 现在条件不满足
	conditionMet = false
	prompt2, err := composer.Assemble(ctx)
	if err != nil {
		t.Fatalf("Failed to assemble prompt: %v", err)
	}

	if strings.Contains(prompt2, "Special instructions") {
		t.Error("Expected prompt to NOT contain conditional instructions when condition is not met")
	}
}

func TestComposer_Assemble_Deduplication(t *testing.T) {
	composer := NewComposer(nil)

	// 注册两个相同 ID 的贡献者，但优先级不同
	block1 := PromptBlock{
		ID:       "duplicate",
		Content:  "Low priority content",
		Priority: 100,
	}
	block2 := PromptBlock{
		ID:       "duplicate",
		Content:  "High priority content",
		Priority: 200,
	}

	// 手动添加重复块（模拟两个 contributor 返回相同 ID）
	testContributor := &testContributor{
		blocks: []PromptBlock{block1, block2},
	}
	composer.RegisterContributor(testContributor)

	ctx := context.Background()
	prompt, err := composer.Assemble(ctx)
	if err != nil {
		t.Fatalf("Failed to assemble prompt: %v", err)
	}

	// 应该只保留高优先级的块
	if strings.Contains(prompt, "Low priority content") {
		t.Error("Expected low priority content to be deduplicated")
	}

	if !strings.Contains(prompt, "High priority content") {
		t.Error("Expected high priority content to be present")
	}
}

func TestComposer_Assemble_EmptyContributors(t *testing.T) {
	composer := NewComposer(nil)

	ctx := context.Background()
	prompt, err := composer.Assemble(ctx)
	if err != nil {
		t.Fatalf("Failed to assemble prompt: %v", err)
	}

	if prompt != "" {
		t.Errorf("Expected empty prompt when no contributors, got: %s", prompt)
	}
}

func TestComposer_GetStats(t *testing.T) {
	composer := NewComposer(nil)

	// 注册几个贡献者
	composer.RegisterContributor(NewSystemPromptContributor("System"))
	composer.RegisterContributor(NewTaskContributor("Task"))

	stats := composer.GetStats()

	if stats["contributors"] != 2 {
		t.Errorf("Expected 2 contributors, got %v", stats["contributors"])
	}

	if stats["cache_size"] == nil {
		t.Error("Expected cache_size in stats")
	}
}

func TestComposer_ClearCache(t *testing.T) {
	composer := NewComposer(nil)

	// 添加一些缓存
	cache := NewPromptCache(5 * time.Minute)
	cache.Set("key1", PromptBlock{ID: "block1"})
	cache.Set("key2", PromptBlock{ID: "block2"})

	composer.cache = cache

	// 清空缓存
	composer.ClearCache()

	if len(composer.cache.cache) != 0 {
		t.Errorf("Expected cache to be cleared, got size %d", len(composer.cache.cache))
	}
}

// testContributor 测试用的贡献者
type testContributor struct {
	blocks []PromptBlock
}

func (tc *testContributor) Name() string {
	return "test"
}

func (tc *testContributor) Contribute(ctx context.Context) ([]PromptBlock, error) {
	return tc.blocks, nil
}
