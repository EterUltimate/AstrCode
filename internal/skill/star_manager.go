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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EterUltimate/astrcode/internal/model"
	"gopkg.in/yaml.v3"
)

// StarManager Star/Plugin 发现与管理（对齐 astrbot-sdk/runtime/stars/star_manager.py）
type StarManager struct {
	stars    map[string]*model.StarMetadata
	handlers map[string]*model.HandlerMetadata
}

// NewStarManager 创建 Star 管理器
func NewStarManager() *StarManager {
	return &StarManager{
		stars:    make(map[string]*model.StarMetadata),
		handlers: make(map[string]*model.HandlerMetadata),
	}
}

// DiscoverStar 从 plugin.yaml 发现插件
func (sm *StarManager) DiscoverStar(rootDir string) ([]*model.StarMetadata, error) {
	yamlPath := filepath.Join(rootDir, "plugin.yaml")
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no plugin.yaml found in %s", rootDir)
	}

	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("read plugin.yaml: %w", err)
	}

	var plugin model.PluginYAML
	if err := yaml.Unmarshal(data, &plugin); err != nil {
		return nil, fmt.Errorf("parse plugin.yaml: %w", err)
	}

	if plugin.Name == "" {
		return nil, fmt.Errorf("plugin name is required in plugin.yaml")
	}

	// 查找 logo
	logoPath := ""
	if _, err := os.Stat(filepath.Join(rootDir, "logo.png")); err == nil {
		logoPath = filepath.Join(rootDir, "logo.png")
	}

	// 收集 handler 全名
	handlerFullNames := make([]string, 0)
	for _, comp := range plugin.Components {
		if comp.Class != "" {
			// 类路径格式: module.path:ClassName
			// Handler 全名: module.path:ClassName.handler_name
			parts := strings.Split(comp.Class, ":")
			if len(parts) == 2 {
				handlerFullNames = append(handlerFullNames, comp.Class+".on_message")
			}
		}
	}

	modulePath := fmt.Sprintf("%s.main", plugin.Name)
	meta := &model.StarMetadata{
		Name:             plugin.Name,
		Author:           plugin.Author,
		Desc:             plugin.Desc,
		Version:          plugin.Version,
		Repo:             plugin.Repo,
		ModulePath:       modulePath,
		RootDirName:      filepath.Base(rootDir),
		Reserved:         false,
		Activated:        true,
		HandlerFullNames: handlerFullNames,
		DisplayName:      plugin.DisplayName,
		LogoPath:         logoPath,
	}

	sm.stars[modulePath] = meta

	// 为每个 handler 注册元数据
	for _, hfn := range handlerFullNames {
		sm.handlers[hfn] = &model.HandlerMetadata{
			HandlerFullName: hfn,
			EventType:       "AdapterMessageEvent",
		}
	}

	return []*model.StarMetadata{meta}, nil
}

// DiscoverAllStars 从目录发现所有插件
func (sm *StarManager) DiscoverAllStars(rootDir string) ([]*model.StarMetadata, error) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	var allMeta []*model.StarMetadata
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirPath := filepath.Join(rootDir, entry.Name())
		metas, err := sm.DiscoverStar(dirPath)
		if err != nil {
			continue
		}
		allMeta = append(allMeta, metas...)
	}

	return allMeta, nil
}

// GetStar 获取 Star 元数据
func (sm *StarManager) GetStar(modulePath string) (*model.StarMetadata, bool) {
	meta, ok := sm.stars[modulePath]
	return meta, ok
}

// GetAllStars 获取所有 Star
func (sm *StarManager) GetAllStars() []*model.StarMetadata {
	result := make([]*model.StarMetadata, 0, len(sm.stars))
	for _, meta := range sm.stars {
		result = append(result, meta)
	}
	return result
}

// GetHandlersForStar 获取 Star 的所有 Handler
func (sm *StarManager) GetHandlersForStar(modulePath string) []*model.HandlerMetadata {
	star, ok := sm.stars[modulePath]
	if !ok {
		return nil
	}

	result := make([]*model.HandlerMetadata, 0)
	for _, hfn := range star.HandlerFullNames {
		if h, ok := sm.handlers[hfn]; ok {
			result = append(result, h)
		}
	}
	return result
}

// DiscoverFromJSON 从 JSON 导入 Star 元数据（兼容不同来源）
func (sm *StarManager) DiscoverFromJSON(jsonData []byte) ([]*model.StarMetadata, error) {
	var metas []*model.StarMetadata
	if err := json.Unmarshal(jsonData, &metas); err != nil {
		return nil, fmt.Errorf("parse star metadata json: %w", err)
	}

	for _, meta := range metas {
		sm.stars[meta.ModulePath] = meta
		for _, hfn := range meta.HandlerFullNames {
			sm.handlers[hfn] = &model.HandlerMetadata{
				HandlerFullName: hfn,
				EventType:       "AdapterMessageEvent",
			}
		}
	}
	return metas, nil
}
