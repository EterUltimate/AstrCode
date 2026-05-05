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

package test

import (
	"context"
	"testing"
	"time"

	"github.com/EterUltimate/astrcode/internal/agent"
	"github.com/EterUltimate/astrcode/internal/llm"
	"github.com/EterUltimate/astrcode/internal/model"
	"github.com/EterUltimate/astrcode/internal/sdk"
)

func TestAgentProcessTask(t *testing.T) {
	// 创建模拟客户端
	sdkClient := sdk.NewAstrBotClient("http://localhost:6185", "")
	llmClient := llm.NewClient(llm.ProviderOpenAI, "http://localhost:11434", "", "qwen2.5")

	// 创建 Agent
	ag := agent.NewAgent(llmClient, sdkClient)

	// 注册测试 Skill
	ag.RegisterSkill(model.Skill{
		Name:        "read_code",
		Description: "读取代码文件内容",
	})
	ag.RegisterSkill(model.Skill{
		Name:        "fix_bug",
		Description: "修复代码中的 bug",
	})

	// 测试任务处理
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	task, err := ag.ProcessTask(ctx, "读取代码并修复 bug")
	if err != nil {
		t.Logf("Task processing error (expected in test): %v", err)
	}

	if task != nil {
		t.Logf("Task ID: %s", task.ID)
		t.Logf("Task Status: %s", task.Status)
		t.Logf("Task Result: %s", task.Result)
	}
}
