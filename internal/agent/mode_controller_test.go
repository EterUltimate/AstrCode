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
	"strings"
	"testing"

	"github.com/EterUltimate/astrcode/internal/mode"
)

func TestModeController_IsToolAllowed_CodeMode(t *testing.T) {
	manager := mode.NewManager(mode.DefaultConfig())
	controller := NewModeController(manager)

	// Code mode 应该允许所有工具
	tools := []string{"read", "write", "execute", "delete", "custom_tool"}
	for _, tool := range tools {
		if !controller.IsToolAllowed(tool) {
			t.Errorf("Expected tool '%s' to be allowed in Code mode", tool)
		}
	}
}

func TestModeController_IsToolAllowed_PlanMode_Allowed(t *testing.T) {
	config := &mode.Config{
		DefaultMode:     mode.ModePlan,
		AllowSwitch:     true,
		RequireApproval: true,
	}
	manager := mode.NewManager(config)
	controller := NewModeController(manager)

	// Plan mode 应该允许读取工具
	allowedTools := []string{"read", "search", "list_dir", "grep"}
	for _, tool := range allowedTools {
		if !controller.IsToolAllowed(tool) {
			t.Errorf("Expected tool '%s' to be allowed in Plan mode", tool)
		}
	}
}

func TestModeController_IsToolAllowed_PlanMode_Denied(t *testing.T) {
	config := &mode.Config{
		DefaultMode:     mode.ModePlan,
		AllowSwitch:     true,
		RequireApproval: true,
	}
	manager := mode.NewManager(config)
	controller := NewModeController(manager)

	// Plan mode 应该拒绝写入工具
	deniedTools := []string{"write", "execute", "delete", "modify"}
	for _, tool := range deniedTools {
		if controller.IsToolAllowed(tool) {
			t.Errorf("Expected tool '%s' to be denied in Plan mode", tool)
		}
	}
}

func TestModeController_ValidatePlanExit_ValidPlan(t *testing.T) {
	config := &mode.Config{
		DefaultMode:     mode.ModePlan,
		AllowSwitch:     true,
		RequireApproval: true,
	}
	manager := mode.NewManager(config)
	controller := NewModeController(manager)

	validPlan := `
# Implementation Plan

## Overview
This plan implements a new feature.

## Steps
1. Create the database schema
2. Implement the API endpoints
3. Add unit tests
4. Update documentation

## Edge Cases
- Handle invalid input
- Consider error handling for network failures
`

	err := controller.ValidatePlanExit(validPlan)
	if err != nil {
		t.Errorf("Expected valid plan to pass validation, got error: %v", err)
	}
}

func TestModeController_ValidatePlanExit_MissingHeading(t *testing.T) {
	config := &mode.Config{
		DefaultMode:     mode.ModePlan,
		AllowSwitch:     true,
		RequireApproval: true,
	}
	manager := mode.NewManager(config)
	controller := NewModeController(manager)

	invalidPlan := `
# My Plan

1. Step one
2. Step two
`

	err := controller.ValidatePlanExit(invalidPlan)
	if err == nil {
		t.Error("Expected error for plan missing 'Implementation Plan' heading")
	}

	if !strings.Contains(err.Error(), "Implementation Plan") {
		t.Errorf("Expected error message to mention 'Implementation Plan', got: %v", err)
	}
}

func TestModeController_ValidatePlanExit_MissingSteps(t *testing.T) {
	config := &mode.Config{
		DefaultMode:     mode.ModePlan,
		AllowSwitch:     true,
		RequireApproval: true,
	}
	manager := mode.NewManager(config)
	controller := NewModeController(manager)

	invalidPlan := `
# Implementation Plan

This is a plan without clear steps.
Just some text here.
`

	err := controller.ValidatePlanExit(invalidPlan)
	if err == nil {
		t.Error("Expected error for plan missing clear steps")
	}

	if !strings.Contains(err.Error(), "steps") {
		t.Errorf("Expected error message to mention 'steps', got: %v", err)
	}
}

func TestModeController_ValidatePlanExit_NoApprovalRequired(t *testing.T) {
	config := &mode.Config{
		DefaultMode:     mode.ModeCode,
		AllowSwitch:     true,
		RequireApproval: false, // 不需要审批
	}
	manager := mode.NewManager(config)
	controller := NewModeController(manager)

	// 即使计划无效，如果不需要审批也应该通过
	invalidPlan := "Just some random text"

	err := controller.ValidatePlanExit(invalidPlan)
	if err != nil {
		t.Errorf("Expected no error when approval not required, got: %v", err)
	}
}

func TestModeController_GetModeInfo(t *testing.T) {
	config := &mode.Config{
		DefaultMode:      mode.ModePlan,
		AllowSwitch:      true,
		RequireApproval:  true,
		AutoGeneratePlan: true,
	}
	manager := mode.NewManager(config)
	controller := NewModeController(manager)

	info := controller.GetModeInfo()

	if info["current_mode"] != string(mode.ModePlan) {
		t.Errorf("Expected current_mode to be 'plan', got: %v", info["current_mode"])
	}

	if info["allow_switch"] != true {
		t.Errorf("Expected allow_switch to be true, got: %v", info["allow_switch"])
	}

	if info["require_approval"] != true {
		t.Errorf("Expected require_approval to be true, got: %v", info["require_approval"])
	}

	if info["auto_generate_plan"] != true {
		t.Errorf("Expected auto_generate_plan to be true, got: %v", info["auto_generate_plan"])
	}
}

func TestModeController_SwitchModes(t *testing.T) {
	manager := mode.NewManager(mode.DefaultConfig())
	controller := NewModeController(manager)

	// 初始应该是 Code mode
	if manager.GetMode() != mode.ModeCode {
		t.Errorf("Expected initial mode to be Code, got: %v", manager.GetMode())
	}

	// 切换到 Plan mode
	err := controller.SwitchToPlan()
	if err != nil {
		t.Errorf("Failed to switch to Plan mode: %v", err)
	}

	if manager.GetMode() != mode.ModePlan {
		t.Errorf("Expected mode to be Plan after switch, got: %v", manager.GetMode())
	}

	// 切换回 Code mode
	err = controller.SwitchToCode()
	if err != nil {
		t.Errorf("Failed to switch to Code mode: %v", err)
	}

	if manager.GetMode() != mode.ModeCode {
		t.Errorf("Expected mode to be Code after switch, got: %v", manager.GetMode())
	}
}

func TestModeController_Toggle(t *testing.T) {
	manager := mode.NewManager(mode.DefaultConfig())
	controller := NewModeController(manager)

	// 初始是 Code mode
	if manager.GetMode() != mode.ModeCode {
		t.Errorf("Expected initial mode to be Code, got: %v", manager.GetMode())
	}

	// Toggle 应该切换到 Plan mode
	err := controller.Toggle()
	if err != nil {
		t.Errorf("Failed to toggle mode: %v", err)
	}

	if manager.GetMode() != mode.ModePlan {
		t.Errorf("Expected mode to be Plan after toggle, got: %v", manager.GetMode())
	}

	// 再次 Toggle 应该切换回 Code mode
	err = controller.Toggle()
	if err != nil {
		t.Errorf("Failed to toggle mode: %v", err)
	}

	if manager.GetMode() != mode.ModeCode {
		t.Errorf("Expected mode to be Code after second toggle, got: %v", manager.GetMode())
	}
}
