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

package agent

import (
	"fmt"
	"strings"

	"github.com/EterUltimate/astrcode/internal/mode"
)

// ModeController 控制 Agent 的运行模式（Code/Plan）
type ModeController struct {
	manager *mode.Manager
}

// NewModeController 创建模式控制器
func NewModeController(manager *mode.Manager) *ModeController {
	return &ModeController{
		manager: manager,
	}
}

// IsToolAllowed 检查工具是否在当前模式下允许使用
func (mc *ModeController) IsToolAllowed(toolName string) bool {
	currentMode := mc.manager.GetMode()

	if currentMode == mode.ModePlan {
		// Plan mode 只允许读取工具和计划管理工具
		allowedTools := map[string]bool{
			"read":          true,
			"search":        true,
			"list_dir":      true,
			"grep":          true,
			"plan_create":   true,
			"plan_update":   true,
			"plan_view":     true,
			"plan_validate": true,
		}
		return allowedTools[toolName]
	}

	// Code mode 允许所有工具
	return true
}

// ValidatePlanExit Exit Gate: 验证计划是否符合退出条件
func (mc *ModeController) ValidatePlanExit(planContent string) error {
	if !mc.manager.ShouldRequireApproval() {
		return nil // 不需要审批，直接允许
	}

	// 检查清单：
	// 1. 是否包含 ## Implementation Plan heading
	if !strings.Contains(planContent, "## Implementation Plan") &&
		!strings.Contains(planContent, "# Implementation Plan") {
		return fmt.Errorf("plan must contain 'Implementation Plan' heading")
	}

	// 2. 是否有明确的步骤列表（至少一个有序或无序列表）
	hasSteps := strings.Contains(planContent, "1.") ||
		strings.Contains(planContent, "- ") ||
		strings.Contains(planContent, "* ")

	if !hasSteps {
		return fmt.Errorf("plan must contain clear steps (numbered or bullet list)")
	}

	// 3. 是否考虑了边界情况（可选但推荐）
	// 检查是否包含关键词：edge case, boundary, error handling, validation
	keywords := []string{"edge case", "boundary", "error handling", "validation", "corner case"}
	hasConsiderations := false
	for _, keyword := range keywords {
		if strings.Contains(strings.ToLower(planContent), keyword) {
			hasConsiderations = true
			break
		}
	}

	// 如果没有明确考虑边界情况，给出警告但不阻止
	if !hasConsiderations {
		// 这里可以记录日志或返回警告，但不阻止退出
		// log.Println("Warning: plan may not consider edge cases")
		// 空分支是预期的，用于未来扩展
	}

	return nil
}

// GetModeInfo 获取当前模式信息
func (mc *ModeController) GetModeInfo() map[string]interface{} {
	stats := mc.manager.GetStats()
	return stats
}

// SwitchToCode 切换到代码模式
func (mc *ModeController) SwitchToCode() error {
	return mc.manager.SwitchToCode()
}

// SwitchToPlan 切换到计划模式
func (mc *ModeController) SwitchToPlan() error {
	return mc.manager.SwitchToPlan()
}

// Toggle 切换模式
func (mc *ModeController) Toggle() error {
	return mc.manager.Toggle()
}
