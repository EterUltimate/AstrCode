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
	"testing"

	"github.com/EterUltimate/astrcode/internal/model"
)

func TestRetrieverRegister(t *testing.T) {
	r := NewRetriever()
	r.Register(model.Skill{Name: "read_code", Description: "读取代码"})
	r.Register(model.Skill{Name: "fix_bug", Description: "修复 bug"})

	skills := r.AllSkills()
	if len(skills) != 2 {
		t.Errorf("expected 2 skills, got %d", len(skills))
	}
}

func TestRetrieverKeywordMatch(t *testing.T) {
	r := NewRetriever()
	r.Register(model.Skill{Name: "read_code", Description: "读取代码文件内容"})
	r.Register(model.Skill{Name: "fix_bug", Description: "修复代码中的 bug"})
	r.Register(model.Skill{Name: "deploy", Description: "部署服务到生产环境"})

	// 匹配 "代码" 关键词
	skills, err := r.keywordRetrieve("读取代码", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) == 0 {
		t.Error("expected some results")
	}

	// read_code 应该排在前面
	if skills[0].Name != "read_code" {
		t.Errorf("expected read_code first, got %s", skills[0].Name)
	}
}

func TestStarManagerDiscover(t *testing.T) {
	sm := NewStarManager()

	// 不存在的路径
	_, err := sm.DiscoverStar("/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestStarManagerGet(t *testing.T) {
	sm := NewStarManager()

	// 手动添加
	meta := &model.StarMetadata{
		Name:       "test_star",
		ModulePath: "test_star.main",
		Activated:  true,
	}
	sm.stars["test_star.main"] = meta

	got, ok := sm.GetStar("test_star.main")
	if !ok {
		t.Error("expected to find star")
	}
	if got.Name != "test_star" {
		t.Errorf("expected test_star, got %s", got.Name)
	}

	all := sm.GetAllStars()
	if len(all) != 1 {
		t.Errorf("expected 1 star, got %d", len(all))
	}
}

func TestFallbackHandler(t *testing.T) {
	fh := NewFallbackHandler()

	step := &model.Step{
		ID:         "s1",
		Type:       model.StepTypeSkill,
		Skill:      "test",
		MaxRetries: 2,
		RetryCount: 0,
	}

	// First failure → retry
	strategy, _, _ := fh.HandleFailure(step, errTest)
	if strategy != FallbackRetry {
		t.Errorf("expected retry, got %s", strategy)
	}

	// Exhausted retries → alternative
	step.RetryCount = 2
	strategy, _, _ = fh.HandleFailure(step, errTest)
	if strategy != FallbackAlternative {
		t.Errorf("expected alternative, got %s", strategy)
	}

	// Skip strategy
	fh.SetStrategy(model.StepTypeHandler, FallbackSkip)
	step.Type = model.StepTypeHandler
	strategy, _, _ = fh.HandleFailure(step, errTest)
	if strategy != FallbackSkip {
		t.Errorf("expected skip, got %s", strategy)
	}
}

var errTest = fmtErr("test error")

type fmtErr string

func (e fmtErr) Error() string { return string(e) }
