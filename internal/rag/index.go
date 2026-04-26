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

package rag

import (
	"context"
	"fmt"

	"github.com/EterUltimate/astrcode/internal/model"
)

// SkillIndex Skill 向量索引
type SkillIndex struct {
	store     VectorStore
	embedding *EmbeddingClient
}

// NewSkillIndex 创建 Skill 索引
func NewSkillIndex(store VectorStore, embedding *EmbeddingClient) *SkillIndex {
	return &SkillIndex{
		store:     store,
		embedding: embedding,
	}
}

// IndexSkill 索引单个 Skill
func (idx *SkillIndex) IndexSkill(ctx context.Context, skill model.Skill) error {
	// 使用 name + description 作为索引文本
	text := skill.Name + " " + skill.Description
	if skill.Summary != "" {
		text += " " + skill.Summary
	}

	embedding, err := idx.embedding.EmbedText(ctx, text)
	if err != nil {
		return fmt.Errorf("embed skill %s: %w", skill.Name, err)
	}

	metadata := map[string]interface{}{
		"name":        skill.Name,
		"description": skill.Description,
		"path":        skill.Path,
	}

	if err := idx.store.Add(skill.Name, embedding, metadata); err != nil {
		return fmt.Errorf("add to store: %w", err)
	}

	return nil
}

// IndexSkills 批量索引 Skills
func (idx *SkillIndex) IndexSkills(ctx context.Context, skills []model.Skill) error {
	for _, skill := range skills {
		if err := idx.IndexSkill(ctx, skill); err != nil {
			return err
		}
	}
	return nil
}

// Search 根据任务搜索相关 Skill
func (idx *SkillIndex) Search(ctx context.Context, task string, topK int) ([]model.Skill, error) {
	queryEmbedding, err := idx.embedding.EmbedText(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	results, err := idx.store.Search(queryEmbedding, topK)
	if err != nil {
		return nil, fmt.Errorf("search store: %w", err)
	}

	skills := make([]model.Skill, 0, len(results))
	for _, r := range results {
		meta := r.Metadata
		skills = append(skills, model.Skill{
			Name:        getString(meta, "name"),
			Description: getString(meta, "description"),
			Path:        getString(meta, "path"),
		})
	}

	return skills, nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
