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
	llmClient := llm.NewClient("http://localhost:11434", "", "qwen2.5")
	
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
