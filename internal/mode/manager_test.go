package mode

import (
	"testing"
)

func TestManager_DefaultMode(t *testing.T) {
	mgr := NewManager(nil)

	if mgr.GetMode() != ModeCode {
		t.Errorf("Expected default mode 'code', got '%s'", mgr.GetMode())
	}
}

func TestManager_SwitchMode(t *testing.T) {
	mgr := NewManager(DefaultConfig())

	// 切换到计划模式
	err := mgr.SwitchToPlan()
	if err != nil {
		t.Fatalf("SwitchToPlan failed: %v", err)
	}

	if !mgr.IsPlanMode() {
		t.Error("Should be in plan mode")
	}

	// 切换回代码模式
	err = mgr.SwitchToCode()
	if err != nil {
		t.Fatalf("SwitchToCode failed: %v", err)
	}

	if !mgr.IsCodeMode() {
		t.Error("Should be in code mode")
	}
}

func TestManager_Toggle(t *testing.T) {
	mgr := NewManager(DefaultConfig())

	// 初始为 code 模式
	if !mgr.IsCodeMode() {
		t.Error("Should start in code mode")
	}

	// Toggle 到 plan 模式
	mgr.Toggle()
	if !mgr.IsPlanMode() {
		t.Error("Should be in plan mode after toggle")
	}

	// Toggle 回 code 模式
	mgr.Toggle()
	if !mgr.IsCodeMode() {
		t.Error("Should be in code mode after second toggle")
	}
}

func TestManager_DisallowSwitch(t *testing.T) {
	config := DefaultConfig()
	config.AllowSwitch = false
	mgr := NewManager(config)

	err := mgr.SwitchToPlan()
	if err == nil {
		t.Error("Expected error when switching is not allowed")
	}
}

func TestManager_InvalidMode(t *testing.T) {
	mgr := NewManager(DefaultConfig())

	err := mgr.SetMode("invalid")
	if err == nil {
		t.Error("Expected error for invalid mode")
	}
}

func TestManager_ShouldRequireApproval(t *testing.T) {
	// 计划模式 + RequireApproval = true
	config1 := DefaultConfig()
	config1.RequireApproval = true
	mgr1 := NewManager(config1)
	mgr1.SwitchToPlan()

	if !mgr1.ShouldRequireApproval() {
		t.Error("Should require approval in plan mode with RequireApproval=true")
	}

	// 计划模式 + RequireApproval = false
	config2 := DefaultConfig()
	config2.RequireApproval = false
	mgr2 := NewManager(config2)
	mgr2.SwitchToPlan()

	if mgr2.ShouldRequireApproval() {
		t.Error("Should not require approval when RequireApproval=false")
	}

	// 代码模式
	mgr3 := NewManager(DefaultConfig())
	if mgr3.ShouldRequireApproval() {
		t.Error("Should not require approval in code mode")
	}
}

func TestManager_ShouldAutoGeneratePlan(t *testing.T) {
	// 计划模式 + AutoGeneratePlan = true
	config1 := DefaultConfig()
	config1.AutoGeneratePlan = true
	mgr1 := NewManager(config1)
	mgr1.SwitchToPlan()

	if !mgr1.ShouldAutoGeneratePlan() {
		t.Error("Should auto-generate plan in plan mode with AutoGeneratePlan=true")
	}

	// 计划模式 + AutoGeneratePlan = false
	config2 := DefaultConfig()
	config2.AutoGeneratePlan = false
	mgr2 := NewManager(config2)
	mgr2.SwitchToPlan()

	if mgr2.ShouldAutoGeneratePlan() {
		t.Error("Should not auto-generate plan when AutoGeneratePlan=false")
	}

	// 代码模式
	mgr3 := NewManager(DefaultConfig())
	if mgr3.ShouldAutoGeneratePlan() {
		t.Error("Should not auto-generate plan in code mode")
	}
}

func TestManager_GetModeDescription(t *testing.T) {
	mgr := NewManager(DefaultConfig())

	desc := mgr.GetModeDescription()
	if desc == "" {
		t.Error("Expected non-empty description")
	}

	// 切换到计划模式
	mgr.SwitchToPlan()
	desc = mgr.GetModeDescription()
	if desc == "" {
		t.Error("Expected non-empty description for plan mode")
	}
}

func TestManager_GetStats(t *testing.T) {
	config := DefaultConfig()
	config.RequireApproval = false
	mgr := NewManager(config)

	stats := mgr.GetStats()

	if stats["current_mode"] != "code" {
		t.Errorf("Expected current_mode 'code', got '%v'", stats["current_mode"])
	}

	if stats["allow_switch"] != true {
		t.Errorf("Expected allow_switch true, got '%v'", stats["allow_switch"])
	}

	if stats["require_approval"] != false {
		t.Errorf("Expected require_approval false, got '%v'", stats["require_approval"])
	}
}

func TestManager_ConcurrentAccess(t *testing.T) {
	mgr := NewManager(DefaultConfig())

	done := make(chan bool, 10)

	// 并发读取和切换模式
	for i := 0; i < 10; i++ {
		go func() {
			mgr.GetMode()
			mgr.Toggle()
			mgr.IsCodeMode()
			mgr.IsPlanMode()
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 如果没有 panic，说明线程安全
	t.Log("Concurrent access test passed")
}
