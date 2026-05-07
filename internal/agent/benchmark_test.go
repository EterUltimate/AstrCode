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

package agent

import (
	"context"
	"testing"

	"github.com/EterUltimate/astrcode/internal/model"
	"github.com/EterUltimate/astrcode/internal/mode"
	"github.com/EterUltimate/astrcode/internal/prompt"
)

// BenchmarkModeController_IsToolAllowed 基准测试：工具权限检查性能
func BenchmarkModeController_IsToolAllowed(b *testing.B) {
	manager := mode.NewManager(mode.DefaultConfig())
	controller := NewModeController(manager)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.IsToolAllowed("test_tool")
	}
}

// BenchmarkPromptComposer_Assemble 基准测试：Prompt 组装性能（无缓存）
func BenchmarkPromptComposer_Assemble_NoCache(b *testing.B) {
	composer := prompt.NewComposer(nil)

	// 注册贡献者
	systemContributor := prompt.NewSystemPromptContributor("You are an AI assistant.")
	composer.RegisterContributor(systemContributor)

	taskContributor := prompt.NewTaskContributor("Write a Python function to sort a list")
	composer.RegisterContributor(taskContributor)

	skills := []model.Skill{
		{Name: "skill1", Description: "First skill"},
		{Name: "skill2", Description: "Second skill"},
		{Name: "skill3", Description: "Third skill"},
	}
	skillsContributor := prompt.NewSkillsContributor(skills)
	composer.RegisterContributor(skillsContributor)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := composer.Assemble(ctx)
		if err != nil {
			b.Fatalf("Failed to assemble prompt: %v", err)
		}
	}
}

// BenchmarkPromptComposer_Assemble_WithCache 基准测试：Prompt 组装性能（有缓存）
func BenchmarkPromptComposer_Assemble_WithCache(b *testing.B) {
	cache := prompt.NewPromptCache(5 * 60 * 1e9) // 5 minutes
	composer := prompt.NewComposer(cache)

	// 注册贡献者
	systemContributor := prompt.NewSystemPromptContributor("You are an AI assistant.")
	composer.RegisterContributor(systemContributor)

	taskContributor := prompt.NewTaskContributor("Write a Python function to sort a list")
	composer.RegisterContributor(taskContributor)

	skills := []model.Skill{
		{Name: "skill1", Description: "First skill"},
		{Name: "skill2", Description: "Second skill"},
		{Name: "skill3", Description: "Third skill"},
	}
	skillsContributor := prompt.NewSkillsContributor(skills)
	composer.RegisterContributor(skillsContributor)

	ctx := context.Background()

	// 预热缓存
	_, err := composer.Assemble(ctx)
	if err != nil {
		b.Fatalf("Failed to assemble prompt: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := composer.Assemble(ctx)
		if err != nil {
			b.Fatalf("Failed to assemble prompt: %v", err)
		}
	}
}

// BenchmarkPromptComposer_Deduplication 基准测试：去重性能
func BenchmarkPromptComposer_Deduplication(b *testing.B) {
	composer := prompt.NewComposer(nil)

	// 创建多个重复的 contributor
	for i := 0; i < 10; i++ {
		block := prompt.PromptBlock{
			ID:       "duplicate_block",
			Content:  "Duplicate content",
			Priority: i * 10,
		}
		testContributor := &testContributorForBenchmark{block: block}
		composer.RegisterContributor(testContributor)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := composer.Assemble(ctx)
		if err != nil {
			b.Fatalf("Failed to assemble prompt: %v", err)
		}
	}
}

// testContributorForBenchmark 基准测试用的贡献者
type testContributorForBenchmark struct {
	block prompt.PromptBlock
}

func (tc *testContributorForBenchmark) Name() string {
	return "benchmark_test"
}

func (tc *testContributorForBenchmark) Contribute(ctx context.Context) ([]prompt.PromptBlock, error) {
	return []prompt.PromptBlock{tc.block}, nil
}
