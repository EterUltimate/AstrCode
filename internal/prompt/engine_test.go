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
	"testing"

	"github.com/EterUltimate/astrcode/internal/model"
)

func TestBuildDecisionPrompt(t *testing.T) {
	engine := NewEngine()
	skills := []model.Skill{
		{Name: "read_code", Description: "读取代码"},
		{Name: "fix_bug", Description: "修复bug"},
	}

	prompt := engine.BuildDecisionPrompt("读取并修复bug", skills)
	if len(prompt) == 0 {
		t.Error("prompt should not be empty")
	}
	// 应该包含任务描述和技能列表
	if !contains(prompt, "读取并修复bug") {
		t.Error("prompt should contain task description")
	}
	if !contains(prompt, "read_code") {
		t.Error("prompt should contain skill name")
	}
}

func TestBuildPlanPrompt(t *testing.T) {
	engine := NewEngine()
	skills := []model.Skill{{Name: "deploy", Description: "部署服务"}}
	decision := &DecisionOutput{
		Action:    "plan",
		Reasoning: "需要多步骤部署",
	}

	prompt := engine.BuildPlanPrompt("部署到生产", skills, decision)
	if !contains(prompt, "部署到生产") {
		t.Error("prompt should contain task")
	}
	if !contains(prompt, "deploy") {
		t.Error("prompt should contain skill")
	}
}

func TestParseSkillCall(t *testing.T) {
	engine := NewEngine()

	name, args, err := engine.ParseSkillCall("read_code|path=/main.go|line=42")
	if err != nil {
		t.Fatal(err)
	}
	if name != "read_code" {
		t.Errorf("expected read_code, got %s", name)
	}
	if args["path"] != "/main.go" {
		t.Errorf("expected /main.go, got %v", args["path"])
	}
}

func TestParseDecision(t *testing.T) {
	engine := NewEngine()

	// Valid JSON input
	decision, err := engine.ParseDecision(`{"action": "plan", "skills": ["read_code"], "confidence": 0.9}`)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Action != "plan" {
		t.Errorf("expected plan, got %s", decision.Action)
	}

	// No JSON → error
	_, err = engine.ParseDecision("plain text without json")
	if err == nil {
		t.Error("expected error for non-JSON input")
	}
}

func TestFormatSkillsForPrompt(t *testing.T) {
	skills := []model.Skill{
		{Name: "a", Description: "desc a"},
		{Name: "b", Description: "desc b"},
	}
	result := FormatSkillsForPrompt(skills)
	if !contains(result, "a") || !contains(result, "b") {
		t.Error("should contain both skill names")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
