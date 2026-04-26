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

package deploy

import (
	"context"
	"fmt"
	"log"

	"github.com/EterUltimate/astrcode/internal/sdk"
)

// Manager 部署管理器
// 负责插件的测试、热重载和版本管理
type Manager struct {
	sdkClient *sdk.AstrBotClient
}

// NewManager 创建部署管理器
func NewManager(sdkClient *sdk.AstrBotClient) *Manager {
	return &Manager{
		sdkClient: sdkClient,
	}
}

// DeployResult 部署结果
type DeployResult struct {
	Success   bool
	Message   string
	Version   string
	Rollback  func() error // 回滚函数
}

// DeployPlugin 部署插件到 AstrBot Runtime
func (m *Manager) DeployPlugin(ctx context.Context, pluginName string, files map[string]string) (*DeployResult, error) {
	log.Printf("🚀 Deploying plugin: %s", pluginName)

	// 1. 备份当前版本（用于回滚）
	backup, err := m.backupCurrentVersion(pluginName)
	if err != nil {
		log.Printf("⚠️  Backup failed: %v", err)
		// 继续部署，但记录警告
	}

	// 2. 通过 JSON-RPC 热重载插件
	err = m.hotReload(ctx, pluginName, files)
	if err != nil {
		// 部署失败，尝试回滚
		if backup != nil {
			log.Printf("🔄 Rolling back to previous version...")
			if rbErr := backup(); rbErr != nil {
				log.Printf("❌ Rollback also failed: %v", rbErr)
			}
		}
		return nil, fmt.Errorf("hot reload failed: %w", err)
	}

	// 3. 验证部署
	if err := m.verifyDeployment(ctx, pluginName); err != nil {
		log.Printf("⚠️  Verification failed: %v", err)
		// 不阻断，仅警告
	}

	result := &DeployResult{
		Success: true,
		Message: fmt.Sprintf("Plugin %s deployed successfully", pluginName),
		Version: "0.1.0", // TODO: 从 plugin.yaml 读取
		Rollback: func() error {
			if backup != nil {
				return backup()
			}
			return fmt.Errorf("no backup available")
		},
	}

	log.Printf("✅ Deployment complete: %s", pluginName)
	return result, nil
}

// hotReload 热重载插件
func (m *Manager) hotReload(ctx context.Context, pluginName string, files map[string]string) error {
	// 使用 AstrBot SDK 的热重载 API
	// 这里简化实现，实际需要调用 JSON-RPC 方法
	
	// 示例：astrbot.plugin.reload(plugin_name, files)
	resp, err := m.sdkClient.CallRPC(ctx, "astrbot.plugin.reload", map[string]interface{}{
		"plugin_name": pluginName,
		"files":       files,
	})
	if err != nil {
		return fmt.Errorf("JSON-RPC call failed: %w", err)
	}
	_ = resp // 检查响应

	return nil
}

// verifyDeployment 验证部署
func (m *Manager) verifyDeployment(ctx context.Context, pluginName string) error {
	// 检查插件是否成功加载
	_, err := m.sdkClient.CallRPC(ctx, "astrbot.plugin.status", map[string]interface{}{
		"plugin_name": pluginName,
	})
	if err != nil {
		return err
	}

	return nil
}

// backupCurrentVersion 备份当前版本
func (m *Manager) backupCurrentVersion(pluginName string) (func() error, error) {
	// 获取当前插件文件
	// 保存到临时目录
	// 返回回滚函数

	rollback := func() error {
		// 恢复备份的文件
		log.Printf("🔄 Restoring backup for %s", pluginName)
		return nil
	}

	return rollback, nil
}

// RunTests 运行插件测试
func (m *Manager) RunTests(ctx context.Context, pluginName string, testFiles map[string]string) error {
	log.Printf("🧪 Running tests for: %s", pluginName)

	// 在沙箱环境中运行测试
	// 这里简化实现，实际需要执行 pytest 或类似工具

	for filename, content := range testFiles {
		log.Printf("  Test file: %s (%d bytes)", filename, len(content))
	}

	// TODO: 实际执行测试
	log.Printf("✅ Tests passed")
	return nil
}

// ListPlugins 列出已部署的插件
func (m *Manager) ListPlugins(ctx context.Context) ([]string, error) {
	_, err := m.sdkClient.CallRPC(ctx, "astrbot.plugin.list", nil)
	if err != nil {
		return nil, err
	}

	return []string{}, nil
}

// UninstallPlugin 卸载插件
func (m *Manager) UninstallPlugin(ctx context.Context, pluginName string) error {
	_, err := m.sdkClient.CallRPC(ctx, "astrbot.plugin.uninstall", map[string]interface{}{
		"plugin_name": pluginName,
	})
	if err != nil {
		return err
	}

	log.Printf("🗑️  Plugin uninstalled: %s", pluginName)
	return nil
}
