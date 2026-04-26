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

package codegen

import (
	"context"
	"strings"
	"testing"
)

// TestReviewer_Review_BasicPlugin 测试基本插件审查
func TestReviewer_Review_BasicPlugin(t *testing.T) {
	t.Skip("Skipping test that requires real LLM client for AI review")
	reviewer := NewReviewer(nil)

	ctx := context.Background()

	// 模拟一个基本的插件文件
	pluginFiles := map[string]string{
		"plugin.yaml": `name: test-plugin
version: "1.0.0"
description: A test plugin
`,
		"main.py": `import asyncio

class TestPlugin:
    def __init__(self):
        self.name = "test"
    
    async def handle_message(self, event):
        return "Hello"
`,
	}

	result, err := reviewer.Review(ctx, pluginFiles)

	if err != nil {
		t.Fatalf("Review failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected review result, got nil")
	}

	// 验证评分在合理范围内
	if result.Score < 0 || result.Score > 100 {
		t.Errorf("Score out of range: %d", result.Score)
	}

	t.Logf("Review score: %d, Issues: %d, Passed: %v",
		result.Score, len(result.Issues), result.Passed)
}

// TestReviewer_Review_SecurityIssues 测试安全问题检测
func TestReviewer_Review_SecurityIssues(t *testing.T) {
	t.Skip("Skipping test that requires real LLM client for AI review")
	reviewer := NewReviewer(nil)

	ctx := context.Background()

	// 包含安全问题的代码
	insecureCode := map[string]string{
		"main.py": `
api_key = "sk-1234567890abcdef"
password = "admin123"

def connect():
    # Hardcoded credentials
    return api_key
`,
	}

	result, err := reviewer.Review(ctx, insecureCode)

	if err != nil {
		t.Fatalf("Review failed: %v", err)
	}

	// 应该检测到安全问题
	criticalIssues := reviewer.getCriticalIssues(result)
	if len(criticalIssues) == 0 {
		t.Log("Warning: No critical security issues detected")
	} else {
		t.Logf("Found %d critical issues", len(criticalIssues))
		for _, issue := range criticalIssues {
			t.Logf("  - %s: %s", issue.Severity, issue.Message)
		}
	}
}

// TestReviewer_staticCheck 测试静态检查
func TestReviewer_staticCheck(t *testing.T) {
	reviewer := NewReviewer(nil)

	tests := []struct {
		name      string
		files     map[string]string
		wantIssue bool
	}{
		{
			name: "hardcoded API key",
			files: map[string]string{
				"config.py": `API_KEY = "secret_key_123"`,
			},
			wantIssue: true,
		},
		{
			name: "proper config usage",
			files: map[string]string{
				"config.py": `API_KEY = config.get("api_key")`,
			},
			wantIssue: false,
		},
		{
			name: "print statements",
			files: map[string]string{
				"main.py": `print("debug info")`,
			},
			wantIssue: true, // 生产代码不应有 print
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ReviewResult{
				Issues:      []Issue{},
				Score:       100,
				Suggestions: []string{},
				Passed:      true,
			}

			reviewer.staticCheck(tt.files, result)

			hasIssue := len(result.Issues) > 0
			if hasIssue != tt.wantIssue {
				t.Errorf("Expected issue=%v, got issue=%v (issues: %v)",
					tt.wantIssue, hasIssue, result.Issues)
			}
		})
	}
}

// TestReviewer_calculateScore 测试评分计算
func TestReviewer_calculateScore(t *testing.T) {
	reviewer := NewReviewer(nil)

	tests := []struct {
		name         string
		initialScore int
		issues       []Issue
		expectedMax  int
		expectedMin  int
	}{
		{
			name:         "no issues",
			initialScore: 100,
			issues:       []Issue{},
			expectedMax:  100,
			expectedMin:  100,
		},
		{
			name:         "one critical issue (already deducted in staticCheck)",
			initialScore: 80, // staticCheck would deduct 20
			issues: []Issue{
				{Severity: "critical", Message: "Security issue"},
			},
			expectedMax: 80,
			expectedMin: 80,
		},
		{
			name:         "multiple warnings (already deducted)",
			initialScore: 90, // staticCheck would deduct 5 each
			issues: []Issue{
				{Severity: "warning", Message: "Warning 1"},
				{Severity: "warning", Message: "Warning 2"},
			},
			expectedMax: 90,
			expectedMin: 90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ReviewResult{
				Issues:      tt.issues,
				Score:       tt.initialScore,
				Suggestions: []string{},
				Passed:      true,
			}

			reviewer.calculateScore(result)

			if result.Score > tt.expectedMax || result.Score < tt.expectedMin {
				t.Errorf("Score %d out of expected range [%d, %d]",
					result.Score, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

// TestReviewer_getCriticalIssues 测试关键问题筛选
func TestReviewer_getCriticalIssues(t *testing.T) {
	reviewer := NewReviewer(nil)

	result := &ReviewResult{
		Issues: []Issue{
			{Severity: "critical", Message: "Critical 1"},
			{Severity: "warning", Message: "Warning 1"},
			{Severity: "critical", Message: "Critical 2"},
			{Severity: "info", Message: "Info 1"},
		},
	}

	critical := reviewer.getCriticalIssues(result)

	if len(critical) != 2 {
		t.Errorf("Expected 2 critical issues, got %d", len(critical))
	}

	for _, issue := range critical {
		if issue.Severity != "critical" {
			t.Errorf("Expected critical severity, got %s", issue.Severity)
		}
	}
}

// TestReviewer_Review_EmptyFiles 测试空文件
func TestReviewer_Review_EmptyFiles(t *testing.T) {
	t.Skip("Skipping test that requires real LLM client for AI review")
	reviewer := NewReviewer(nil)

	ctx := context.Background()
	emptyFiles := map[string]string{}

	result, err := reviewer.Review(ctx, emptyFiles)

	if err != nil {
		t.Fatalf("Review failed: %v", err)
	}

	if result.Score != 100 {
		t.Logf("Empty files got score: %d", result.Score)
	}
}

// TestReviewer_Review_LargePlugin 测试大型插件
func TestReviewer_Review_LargePlugin(t *testing.T) {
	t.Skip("Skipping test that requires real LLM client for AI review")
	reviewer := NewReviewer(nil)

	ctx := context.Background()

	// 模拟一个较复杂的插件
	largePlugin := map[string]string{
		"plugin.yaml": `name: complex-plugin
version: "2.0.0"
description: A complex plugin with multiple features
author: Test Author
`,
		"main.py": strings.Repeat(`
class Feature:
    def __init__(self):
        self.data = {}
    
    def process(self):
        return self.data
`, 10),
		"utils.py": `
def helper():
    pass
`,
		"tests/test_main.py": `
def test_feature():
    assert True
`,
	}

	result, err := reviewer.Review(ctx, largePlugin)

	if err != nil {
		t.Fatalf("Review failed: %v", err)
	}

	t.Logf("Large plugin review - Score: %d, Issues: %d",
		result.Score, len(result.Issues))
}

// BenchmarkReviewer_Review 性能基准测试
func BenchmarkReviewer_Review(b *testing.B) {
	reviewer := NewReviewer(nil)
	ctx := context.Background()

	pluginFiles := map[string]string{
		"main.py": `
class Plugin:
    def handle(self):
        return "ok"
`,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = reviewer.Review(ctx, pluginFiles)
	}
}
