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
	"fmt"
	"strings"

	"github.com/EterUltimate/astrcode/internal/llm"
)

// Reviewer 代码审查器
// 检查生成的插件代码质量、安全性和最佳实践
type Reviewer struct {
	llmClient *llm.Client
}

// NewReviewer 创建代码审查器
func NewReviewer(llmClient *llm.Client) *Reviewer {
	return &Reviewer{
		llmClient: llmClient,
	}
}

// ReviewResult 审查结果
type ReviewResult struct {
	Issues      []Issue  // 发现的问题
	Score       int      // 代码评分 (0-100)
	Suggestions []string // 改进建议
	Passed      bool     // 是否通过审查
}

// Issue 问题项
type Issue struct {
	Severity string // critical/warning/info
	File     string
	Line     int
	Message  string
	Fix      string // 修复建议
}

// Review 审查插件代码
func (r *Reviewer) Review(ctx context.Context, pluginFiles map[string]string) (*ReviewResult, error) {
	result := &ReviewResult{
		Issues:      []Issue{},
		Score:       100,
		Suggestions: []string{},
		Passed:      true,
	}

	// 1. 静态规则检查
	r.staticCheck(pluginFiles, result)

	// 2. AI 辅助审查
	if err := r.aiReview(ctx, pluginFiles, result); err != nil {
		return nil, fmt.Errorf("AI review failed: %w", err)
	}

	// 计算最终评分
	r.calculateScore(result)

	// 判断是否通过
	result.Passed = len(r.getCriticalIssues(result)) == 0

	return result, nil
}

// staticCheck 静态规则检查
func (r *Reviewer) staticCheck(files map[string]string, result *ReviewResult) {
	for filename, content := range files {
		lines := strings.Split(content, "\n")

		for i, line := range lines {
			lineNum := i + 1

			// 检查硬编码密钥
			if strings.Contains(strings.ToLower(line), "api_key") &&
				strings.Contains(line, "=") &&
				!strings.Contains(line, "config") &&
				!strings.Contains(line, "os.environ") {
				result.Issues = append(result.Issues, Issue{
					Severity: "critical",
					File:     filename,
					Line:     lineNum,
					Message:  "发现硬编码 API Key，存在安全风险",
					Fix:      "使用 config.yaml 或环境变量存储敏感信息",
				})
				result.Score -= 20
			}

			// 检查裸 except
			if strings.TrimSpace(line) == "except:" {
				result.Issues = append(result.Issues, Issue{
					Severity: "warning",
					File:     filename,
					Line:     lineNum,
					Message:  "使用裸 except，可能掩盖重要错误",
					Fix:      "改为 except Exception as e:",
				})
				result.Score -= 5
			}

			// 检查 print 语句（应使用 logging）
			if strings.HasPrefix(strings.TrimSpace(line), "print(") {
				result.Issues = append(result.Issues, Issue{
					Severity: "info",
					File:     filename,
					Line:     lineNum,
					Message:  "建议使用 logging 模块而非 print",
					Fix:      "import logging; logger.info(...)",
				})
				result.Score -= 2
			}
		}

		// 检查是否导入必要模块
		if filename == "main.py" {
			if !strings.Contains(content, "from astrbot") {
				result.Issues = append(result.Issues, Issue{
					Severity: "critical",
					File:     filename,
					Line:     1,
					Message:  "未导入 astrbot SDK",
					Fix:      "添加: from astrbot.api import *",
				})
				result.Score -= 30
			}
		}
	}
}

// aiReview AI 辅助审查
func (r *Reviewer) aiReview(ctx context.Context, files map[string]string, result *ReviewResult) error {
	// 构建审查 prompt
	var codeSummary strings.Builder
	for filename, content := range files {
		codeSummary.WriteString(fmt.Sprintf("\n=== %s ===\n%s\n", filename, content))
	}

	prompt := fmt.Sprintf(`你是一个资深 Python 代码审查专家。请审查以下 AstrBot 插件代码：

%s

请从以下维度评估：
1. 代码安全性（SQL注入、XSS、命令注入等）
2. 性能优化（异步处理、缓存、资源管理）
3. 最佳实践（错误处理、日志记录、代码结构）
4. AstrBot 规范（Handler 注册、事件处理、配置管理）

输出 JSON 格式：
{
  "issues": [{"severity": "critical/warning/info", "message": "...", "suggestion": "..."}],
  "score": 0-100,
  "passed": true/false
}

只输出 JSON。`, codeSummary.String())

	resp, err := r.llmClient.Complete(ctx, prompt)
	if err != nil {
		return err
	}
	_ = resp // TODO: 解析 AI 审查结果

	return nil
}

// calculateScore 计算最终评分
func (r *Reviewer) calculateScore(result *ReviewResult) {
	if result.Score < 0 {
		result.Score = 0
	}
	if result.Score > 100 {
		result.Score = 100
	}
}

// getCriticalIssues 获取严重问题
func (r *Reviewer) getCriticalIssues(result *ReviewResult) []Issue {
	critical := []Issue{}
	for _, issue := range result.Issues {
		if issue.Severity == "critical" {
			critical = append(critical, issue)
		}
	}
	return critical
}

// GenerateReport 生成审查报告
func (r *Reviewer) GenerateReport(result *ReviewResult) string {
	var report strings.Builder

	report.WriteString("## 代码审查报告\n\n")
	report.WriteString(fmt.Sprintf("**评分**: %d/100\n", result.Score))
	report.WriteString(fmt.Sprintf("**状态**: %s\n\n", map[bool]string{true: "✅ 通过", false: "❌ 未通过"}[result.Passed]))

	if len(result.Issues) > 0 {
		report.WriteString("### 发现的问题\n\n")
		for _, issue := range result.Issues {
			emoji := map[string]string{
				"critical": "🔴",
				"warning":  "⚠️",
				"info":     "💡",
			}[issue.Severity]

			report.WriteString(fmt.Sprintf("%s **%s** (%s:%d)\n", emoji, issue.Severity, issue.File, issue.Line))
			report.WriteString(fmt.Sprintf("   - %s\n", issue.Message))
			if issue.Fix != "" {
				report.WriteString(fmt.Sprintf("   - 修复: %s\n\n", issue.Fix))
			}
		}
	}

	if len(result.Suggestions) > 0 {
		report.WriteString("### 改进建议\n\n")
		for _, suggestion := range result.Suggestions {
			report.WriteString(fmt.Sprintf("- %s\n", suggestion))
		}
		report.WriteString("\n")
	}

	return report.String()
}
