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
	"fmt"

	"github.com/EterUltimate/astrcode/internal/model"
)

// SystemPromptContributor 系统提示贡献者
type SystemPromptContributor struct {
	SystemPrompt string
}

// NewSystemPromptContributor 创建系统提示贡献者
func NewSystemPromptContributor(systemPrompt string) *SystemPromptContributor {
	return &SystemPromptContributor{
		SystemPrompt: systemPrompt,
	}
}

// Name 返回贡献者名称
func (c *SystemPromptContributor) Name() string {
	return "system_prompt"
}

// Contribute 贡献系统提示块
func (c *SystemPromptContributor) Contribute(ctx context.Context) ([]PromptBlock, error) {
	if c.SystemPrompt == "" {
		return nil, fmt.Errorf("system prompt is empty")
	}

	return []PromptBlock{
		{
			ID:         "system_prompt",
			Content:    c.SystemPrompt,
			Priority:   1000, // 最高优先级
			CacheLevel: CacheLevelStable,
			Metadata: map[string]string{
				"type": "system",
			},
		},
	}, nil
}

// SkillsContributor 技能列表贡献者
type SkillsContributor struct {
	Skills []model.Skill
}

// NewSkillsContributor 创建技能贡献者
func NewSkillsContributor(skills []model.Skill) *SkillsContributor {
	return &SkillsContributor{
		Skills: skills,
	}
}

// Name 返回贡献者名称
func (c *SkillsContributor) Name() string {
	return "skills"
}

// Contribute 贡献技能列表块
func (c *SkillsContributor) Contribute(ctx context.Context) ([]PromptBlock, error) {
	if len(c.Skills) == 0 {
		return nil, fmt.Errorf("no skills available")
	}

	content := "## Available Skills\n\n"
	for i, skill := range c.Skills {
		content += fmt.Sprintf("%d. **%s**: %s\n", i+1, skill.Name, skill.Description)
	}

	return []PromptBlock{
		{
			ID:         "skills_list",
			Content:    content,
			Priority:   800,
			CacheLevel: CacheLevelSemiStable,
			Metadata: map[string]string{
				"type":       "skills",
				"skill_count": fmt.Sprintf("%d", len(c.Skills)),
			},
		},
	}, nil
}

// ContextContributor 上下文贡献者（对话历史）
type ContextContributor struct {
	Messages []Message
	MaxMessages int // 最大消息数
}

// Message 对话消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// NewContextContributor 创建上下文贡献者
func NewContextContributor(messages []Message, maxMessages int) *ContextContributor {
	if maxMessages == 0 {
		maxMessages = 50 // 默认保留最近 50 条消息
	}

	return &ContextContributor{
		Messages:    messages,
		MaxMessages: maxMessages,
	}
}

// Name 返回贡献者名称
func (c *ContextContributor) Name() string {
	return "context"
}

// Contribute 贡献对话历史块
func (c *ContextContributor) Contribute(ctx context.Context) ([]PromptBlock, error) {
	if len(c.Messages) == 0 {
		return []PromptBlock{}, nil // 空历史，返回空块
	}

	// 只保留最近 N 条消息
	messages := c.Messages
	if len(messages) > c.MaxMessages {
		messages = messages[len(messages)-c.MaxMessages:]
	}

	content := "## Conversation History\n\n"
	for _, msg := range messages {
		content += fmt.Sprintf("**%s**: %s\n\n", msg.Role, msg.Content)
	}

	return []PromptBlock{
		{
			ID:         "conversation_history",
			Content:    content,
			Priority:   500,
			CacheLevel: CacheLevelDynamic,
			Metadata: map[string]string{
				"type":        "context",
				"message_count": fmt.Sprintf("%d", len(messages)),
			},
		},
	}, nil
}

// TaskContributor 任务描述贡献者
type TaskContributor struct {
	TaskDescription string
}

// NewTaskContributor 创建任务贡献者
func NewTaskContributor(taskDesc string) *TaskContributor {
	return &TaskContributor{
		TaskDescription: taskDesc,
	}
}

// Name 返回贡献者名称
func (c *TaskContributor) Name() string {
	return "task"
}

// Contribute 贡献任务描述块
func (c *TaskContributor) Contribute(ctx context.Context) ([]PromptBlock, error) {
	if c.TaskDescription == "" {
		return nil, fmt.Errorf("task description is empty")
	}

	content := fmt.Sprintf("## Current Task\n\n%s\n", c.TaskDescription)

	return []PromptBlock{
		{
			ID:         "task_description",
			Content:    content,
			Priority:   900,
			CacheLevel: CacheLevelDynamic,
			Metadata: map[string]string{
				"type": "task",
			},
		},
	}, nil
}

// InstructionsContributor 指令贡献者（特定场景的指令）
type InstructionsContributor struct {
	Instructions string
	Condition    func(ctx context.Context) bool
}

// NewInstructionsContributor 创建指令贡献者
func NewInstructionsContributor(instructions string, condition func(ctx context.Context) bool) *InstructionsContributor {
	return &InstructionsContributor{
		Instructions: instructions,
		Condition:    condition,
	}
}

// Name 返回贡献者名称
func (c *InstructionsContributor) Name() string {
	return "instructions"
}

// Contribute 贡献指令块
func (c *InstructionsContributor) Contribute(ctx context.Context) ([]PromptBlock, error) {
	if c.Instructions == "" {
		return nil, fmt.Errorf("instructions are empty")
	}

	content := fmt.Sprintf("## Instructions\n\n%s\n", c.Instructions)

	return []PromptBlock{
		{
			ID:         "instructions",
			Content:    content,
			Priority:   700,
			CacheLevel: CacheLevelSemiStable,
			Metadata: map[string]string{
				"type": "instructions",
			},
			Condition: c.Condition,
		},
	}, nil
}
