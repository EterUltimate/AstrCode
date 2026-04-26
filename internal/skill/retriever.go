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

package skill

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/EterUltimate/astrcode/internal/model"
	"github.com/EterUltimate/astrcode/internal/rag"
)

// Retriever 负责 Skill 检索
type Retriever struct {
	skills    []model.Skill
	index     *rag.SkillIndex
	useVector bool
	mu        sync.RWMutex
}

// NewRetriever 创建新的 Skill 检索器
func NewRetriever() *Retriever {
	return &Retriever{
		skills:    make([]model.Skill, 0),
		useVector: false,
	}
}

// NewRetrieverWithIndex 创建带向量索引的检索器
func NewRetrieverWithIndex(index *rag.SkillIndex) *Retriever {
	return &Retriever{
		skills:    make([]model.Skill, 0),
		index:     index,
		useVector: true,
	}
}

// Register 注册 Skill
func (r *Retriever) Register(skill model.Skill) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.skills = append(r.skills, skill)
}

// RegisterAndIndex 注册并索引 Skill
func (r *Retriever) RegisterAndIndex(ctx context.Context, skill model.Skill) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.skills = append(r.skills, skill)

	if r.index != nil {
		return r.index.IndexSkill(ctx, skill)
	}
	return nil
}

// Retrieve 根据任务检索 Top-K 相关 Skill
func (r *Retriever) Retrieve(ctx context.Context, task string, topK int) ([]model.Skill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.skills) == 0 {
		return nil, fmt.Errorf("no skills registered")
	}

	// 优先使用向量检索
	if r.useVector && r.index != nil {
		skills, err := r.index.Search(ctx, task, topK)
		if err == nil && len(skills) > 0 {
			return skills, nil
		}
		// 向量检索失败时回退到文本匹配
	}

	// 回退：关键词匹配
	return r.keywordRetrieve(task, topK)
}

// keywordRetrieve 基于关键词的检索
func (r *Retriever) keywordRetrieve(task string, topK int) ([]model.Skill, error) {
	type scoredSkill struct {
		skill model.Skill
		score float64
	}

	scored := make([]scoredSkill, 0, len(r.skills))
	for _, s := range r.skills {
		score := calculateKeywordScore(task, s)
		if score > 0 {
			scored = append(scored, scoredSkill{skill: s, score: score})
		}
	}

	// 按分数排序
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// 取 Top-K
	if topK > len(scored) {
		topK = len(scored)
	}
	if topK == 0 {
		return nil, fmt.Errorf("no matching skills found")
	}

	result := make([]model.Skill, topK)
	for i := 0; i < topK; i++ {
		result[i] = scored[i].skill
	}
	return result, nil
}

// calculateKeywordScore 计算关键词匹配分数
func calculateKeywordScore(task string, skill model.Skill) float64 {
	taskLower := strings.ToLower(task)
	nameLower := strings.ToLower(skill.Name)
	descLower := strings.ToLower(skill.Description)

	score := 0.0

	// 名称完全匹配权重最高
	if strings.Contains(nameLower, taskLower) || strings.Contains(taskLower, nameLower) {
		score += 10.0
	}

	// 关键词重叠
	taskWords := strings.Fields(taskLower)
	for _, word := range taskWords {
		if len(word) < 2 {
			continue
		}
		if strings.Contains(nameLower, word) {
			score += 2.0
		}
		if strings.Contains(descLower, word) {
			score += 1.0
		}
	}

	return score
}

// AllSkills 获取所有已注册的技能
func (r *Retriever) AllSkills() []model.Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.Skill, len(r.skills))
	copy(result, r.skills)
	return result
}
