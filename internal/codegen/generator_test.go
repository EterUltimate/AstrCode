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

// TestGenerator_GeneratePlugin_BasicRequirement 测试基本需求生成
func TestGenerator_GeneratePlugin_BasicRequirement(t *testing.T) {
	// 跳过需要真实 LLM 的测试
	t.Skip("Skipping test that requires real LLM client")

	gen := NewGenerator(nil) // LLM client 为 nil，测试模板逻辑

	ctx := context.Background()
	requirement := "创建一个天气查询插件，支持城市名称查询"

	plugin, err := gen.GeneratePlugin(ctx, requirement, nil)

	if err != nil {
		t.Fatalf("GeneratePlugin failed: %v", err)
	}

	if plugin == nil {
		t.Fatal("Expected plugin, got nil")
	}

	// 验证生成的文件
	if _, ok := plugin.Files["plugin.yaml"]; !ok {
		t.Error("Missing plugin.yaml in generated files")
	}

	if _, ok := plugin.Files["main.py"]; !ok {
		t.Error("Missing main.py in generated files")
	}

	// 验证 PluginYaml 和 MainPy 不为空
	if plugin.PluginYaml == "" {
		t.Error("PluginYaml is empty")
	}

	if plugin.MainPy == "" {
		t.Error("MainPy is empty")
	}

	// 验证 Spec 包含基本信息
	if plugin.Spec.Name == "" {
		t.Error("Plugin spec name is empty")
	}

	if plugin.Spec.Description == "" {
		t.Error("Plugin spec description is empty")
	}
}

// TestGenerator_GeneratePlugin_ComplexRequirement 测试复杂需求生成
func TestGenerator_GeneratePlugin_ComplexRequirement(t *testing.T) {
	t.Skip("Skipping test that requires real LLM client")
	gen := NewGenerator(nil)

	ctx := context.Background()
	requirement := `
		创建一个多功能助手插件，包含以下功能：
		1. 待办事项管理（添加、删除、查询）
		2. 提醒功能（设置定时提醒）
		3. 笔记记录（创建、搜索笔记）
		需要持久化存储，支持配置数据目录
	`

	plugin, err := gen.GeneratePlugin(ctx, requirement, nil)

	if err != nil {
		t.Fatalf("GeneratePlugin failed: %v", err)
	}

	// 验证生成了多个处理器
	if len(plugin.Spec.Handlers) == 0 {
		t.Error("Expected multiple handlers for complex requirement")
	}

	// 验证包含配置项
	if len(plugin.Spec.Config) == 0 {
		t.Error("Expected config items for persistence")
	}

	// 验证依赖库
	if len(plugin.Spec.Dependencies) == 0 {
		t.Log("Warning: No dependencies specified")
	}
}

// TestGenerator_GeneratePlugin_EmptyRequirement 测试空需求
func TestGenerator_GeneratePlugin_EmptyRequirement(t *testing.T) {
	t.Skip("Skipping test that requires real LLM client")
	gen := NewGenerator(nil)

	ctx := context.Background()
	requirement := ""

	plugin, err := gen.GeneratePlugin(ctx, requirement, nil)

	// 应该能够处理空需求（可能使用默认值）
	if err != nil {
		t.Logf("Empty requirement handled with error: %v", err)
	}

	if plugin != nil {
		t.Log("Plugin generated even with empty requirement")
	}
}

// TestGenerator_GeneratePlugin_WithSkills 测试带 Skills 的生成
func TestGenerator_GeneratePlugin_WithSkills(t *testing.T) {
	t.Skip("Skipping test that requires real LLM client")
	gen := NewGenerator(nil)

	ctx := context.Background()
	requirement := "创建一个代码审查插件"

	// 模拟相关的 skills
	// 实际使用时需要从 model 包导入 Skill 结构
	_ = []struct {
		Name        string
		Description string
	}{
		{"code_analysis", "分析代码结构和质量"},
		{"security_check", "检查安全漏洞"},
	}

	// 转换为 model.Skill 格式（这里简化处理）
	// 实际使用时需要从 model 包导入

	plugin, err := gen.GeneratePlugin(ctx, requirement, nil)

	if err != nil {
		t.Fatalf("GeneratePlugin with skills failed: %v", err)
	}

	if plugin == nil {
		t.Fatal("Expected plugin with skills context")
	}
}

// TestGenerator_analyzeRequirement 测试需求分析
func TestGenerator_analyzeRequirement(t *testing.T) {
	t.Skip("Skipping test that requires real LLM client")
	gen := NewGenerator(nil)

	ctx := context.Background()

	tests := []struct {
		name        string
		requirement string
		wantName    string // 期望的插件名关键词
	}{
		{
			name:        "simple weather plugin",
			requirement: "创建天气查询插件",
			wantName:    "weather",
		},
		{
			name:        "todo manager",
			requirement: "待办事项管理器",
			wantName:    "todo",
		},
		{
			name:        "code reviewer",
			requirement: "代码审查工具",
			wantName:    "code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := gen.analyzeRequirement(ctx, tt.requirement, nil)
			if err != nil {
				t.Fatalf("analyzeRequirement failed: %v", err)
			}

			if spec.Name == "" {
				t.Error("Spec name should not be empty")
			}

			// 检查名称是否包含期望的关键词（不区分大小写）
			if !strings.Contains(strings.ToLower(spec.Name), strings.ToLower(tt.wantName)) {
				t.Logf("Warning: Expected name to contain '%s', got '%s'", tt.wantName, spec.Name)
			}
		})
	}
}

// TestGenerator_generatePluginYaml 测试 YAML 生成
func TestGenerator_generatePluginYaml(t *testing.T) {
	gen := NewGenerator(nil)

	spec := &PluginSpec{
		Name:        "test-plugin",
		Description: "A test plugin",
		Handlers: []HandlerSpec{
			{
				Name:        "test_handler",
				EventType:   "message",
				Description: "Test handler",
			},
		},
		Config: map[string]string{
			"api_key": "your_api_key_here",
		},
		Dependencies: []string{"requests"},
	}

	yaml, err := gen.generatePluginYaml(spec)

	if err != nil {
		t.Fatalf("generatePluginYaml failed: %v", err)
	}

	if yaml == "" {
		t.Fatal("Generated YAML is empty")
	}

	// 验证 YAML 包含关键字段
	requiredFields := []string{"name:", "description:", "version:"}
	for _, field := range requiredFields {
		if !strings.Contains(yaml, field) {
			t.Errorf("YAML missing required field: %s", field)
		}
	}
}

// TestGenerator_generateMainPy 测试 Python 代码生成
func TestGenerator_generateMainPy(t *testing.T) {
	gen := NewGenerator(nil)

	spec := &PluginSpec{
		Name:        "example",
		Description: "Example plugin",
		Handlers: []HandlerSpec{
			{
				Name:        "hello_handler",
				EventType:   "command",
				Description: "Say hello",
			},
			{
				Name:        "help_handler",
				EventType:   "command",
				Description: "Show help",
			},
		},
	}

	py, err := gen.generateMainPy(spec)

	if err != nil {
		t.Fatalf("generateMainPy failed: %v", err)
	}

	if py == "" {
		t.Fatal("Generated Python code is empty")
	}

	// 验证 Python 代码包含必要元素
	requiredElements := []string{"import", "class", "def"}
	for _, elem := range requiredElements {
		if !strings.Contains(py, elem) {
			t.Errorf("Python code missing element: %s", elem)
		}
	}
}

// TestGenerator_GeneratePlugin_ErrorHandling 测试错误处理
func TestGenerator_GeneratePlugin_ErrorHandling(t *testing.T) {
	t.Skip("Skipping test that requires real LLM client")
	gen := NewGenerator(nil)

	ctx := context.Background()

	// 测试超时上下文
	ctxTimeout, cancel := context.WithTimeout(ctx, 0)
	defer cancel()

	_, err := gen.GeneratePlugin(ctxTimeout, "test", nil)

	// 可能会因为超时而失败
	if err != nil {
		t.Logf("Expected error with timeout context: %v", err)
	}
}

// BenchmarkGenerator_GeneratePlugin 性能基准测试
func BenchmarkGenerator_GeneratePlugin(b *testing.B) {
	gen := NewGenerator(nil)
	ctx := context.Background()
	requirement := "创建一个简单的问候插件"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.GeneratePlugin(ctx, requirement, nil)
	}
}
