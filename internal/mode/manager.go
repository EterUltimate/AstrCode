package mode

import (
	"fmt"
	"sync"
)

// RunMode 运行模式
type RunMode string

const (
	ModeCode RunMode = "code" // 代码模式：直接执行
	ModePlan RunMode = "plan" // 计划模式：先生成计划，用户确认后执行
)

// Config 模式配置
type Config struct {
	DefaultMode      RunMode // 默认模式
	AllowSwitch      bool    // 是否允许切换模式
	RequireApproval  bool    // 计划模式是否需要用户确认
	AutoGeneratePlan bool    // 是否自动生成计划
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		DefaultMode:      ModeCode,
		AllowSwitch:      true,
		RequireApproval:  true,
		AutoGeneratePlan: true,
	}
}

// Manager 模式管理器
type Manager struct {
	mu     sync.RWMutex
	mode   RunMode
	config *Config
}

// NewManager 创建模式管理器
func NewManager(config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	return &Manager{
		mode:   config.DefaultMode,
		config: config,
	}
}

// GetMode 获取当前模式
func (m *Manager) GetMode() RunMode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mode
}

// SetMode 设置模式
func (m *Manager) SetMode(mode RunMode) error {
	if !m.config.AllowSwitch {
		return fmt.Errorf("mode switching is not allowed")
	}

	if mode != ModeCode && mode != ModePlan {
		return fmt.Errorf("invalid mode: %s", mode)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.mode = mode

	return nil
}

// SwitchToCode 切换到代码模式
func (m *Manager) SwitchToCode() error {
	return m.SetMode(ModeCode)
}

// SwitchToPlan 切换到计划模式
func (m *Manager) SwitchToPlan() error {
	return m.SetMode(ModePlan)
}

// IsCodeMode 是否为代码模式
func (m *Manager) IsCodeMode() bool {
	return m.GetMode() == ModeCode
}

// IsPlanMode 是否为计划模式
func (m *Manager) IsPlanMode() bool {
	return m.GetMode() == ModePlan
}

// ShouldRequireApproval 是否需要用户确认
func (m *Manager) ShouldRequireApproval() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.mode == ModePlan && m.config.RequireApproval
}

// ShouldAutoGeneratePlan 是否应自动生成计划
func (m *Manager) ShouldAutoGeneratePlan() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.mode == ModePlan && m.config.AutoGeneratePlan
}

// GetModeDescription 获取模式描述
func (m *Manager) GetModeDescription() string {
	mode := m.GetMode()

	switch mode {
	case ModeCode:
		return "Code Mode: Execute tasks directly without planning"
	case ModePlan:
		return "Plan Mode: Generate execution plan first, then execute after approval"
	default:
		return "Unknown mode"
	}
}

// Toggle 切换模式
func (m *Manager) Toggle() error {
	current := m.GetMode()

	if current == ModeCode {
		return m.SwitchToPlan()
	} else {
		return m.SwitchToCode()
	}
}

// GetStats 获取模式统计信息
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"current_mode":       string(m.mode),
		"allow_switch":       m.config.AllowSwitch,
		"require_approval":   m.config.RequireApproval,
		"auto_generate_plan": m.config.AutoGeneratePlan,
	}
}
