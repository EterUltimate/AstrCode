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

package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PersistCache 持久化缓存接口
type PersistCache interface {
	Cache
	Save() error
	Load() error
}

// FileCache 基于文件的持久化缓存
type FileCache struct {
	data  map[string]cacheItem
	mu    sync.RWMutex
	path  string
	dirty bool
}

// NewFileCache 创建文件缓存
func NewFileCache(path string) (*FileCache, error) {
	fc := &FileCache{
		data:  make(map[string]cacheItem),
		path:  path,
		dirty: false,
	}

	// 尝试加载已有缓存
	if err := fc.Load(); err != nil {
		// 文件不存在是正常的
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("load cache: %w", err)
		}
	}

	// 启动自动保存
	go fc.autoSave()

	return fc, nil
}

// Get 获取缓存值
func (c *FileCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.data[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(item.expiration) {
		return nil, false
	}

	return item.value, true
}

// Set 设置缓存值
func (c *FileCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
	c.dirty = true
}

// Delete 删除缓存
func (c *FileCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
	c.dirty = true
}

// Clear 清空缓存
func (c *FileCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]cacheItem)
	c.dirty = true
}

// Save 保存缓存到文件
func (c *FileCache) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.dirty {
		return nil
	}

	// 过滤过期项
	now := time.Now()
	cleanData := make(map[string]persistItem)
	for key, item := range c.data {
		if now.Before(item.expiration) {
			cleanData[key] = persistItem{
				Value:      item.value,
				Expiration: item.expiration.Unix(),
			}
		}
	}

	// 确保目录存在
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	// 写入临时文件后重命名（原子操作）
	tmpPath := c.path + ".tmp"
	data, err := json.Marshal(cleanData)
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("write cache file: %w", err)
	}

	if err := os.Rename(tmpPath, c.path); err != nil {
		return fmt.Errorf("rename cache file: %w", err)
	}

	c.dirty = false
	return nil
}

// Load 从文件加载缓存
func (c *FileCache) Load() error {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return err
	}

	var loaded map[string]persistItem
	if err := json.Unmarshal(data, &loaded); err != nil {
		return fmt.Errorf("unmarshal cache: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range loaded {
		exp := time.Unix(item.Expiration, 0)
		if now.Before(exp) {
			c.data[key] = cacheItem{
				value:      item.Value,
				expiration: exp,
			}
		}
	}

	return nil
}

// autoSave 自动保存（每 30 秒）
func (c *FileCache) autoSave() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := c.Save(); err != nil {
			// 静默处理保存错误
			continue
		}
	}
}

// persistItem 持久化项结构
type persistItem struct {
	Value      interface{} `json:"value"`
	Expiration int64       `json:"expiration"`
}

// ============================================================
// HybridCache 混合缓存：内存 + 磁盘
// ============================================================

// HybridCache 混合缓存
type HybridCache struct {
	memory *MemoryCache
	disk   *FileCache
}

// NewHybridCache 创建混合缓存
func NewHybridCache(diskPath string) (*HybridCache, error) {
	disk, err := NewFileCache(diskPath)
	if err != nil {
		return nil, err
	}

	return &HybridCache{
		memory: NewMemoryCache(),
		disk:   disk,
	}, nil
}

// Get 获取缓存（先内存后磁盘）
func (c *HybridCache) Get(key string) (interface{}, bool) {
	// 先查内存
	if val, ok := c.memory.Get(key); ok {
		return val, true
	}

	// 再查磁盘
	if val, ok := c.disk.Get(key); ok {
		// 回填内存
		c.memory.Set(key, val, 5*time.Minute)
		return val, true
	}

	return nil, false
}

// Set 设置缓存（内存 + 磁盘）
func (c *HybridCache) Set(key string, value interface{}, ttl time.Duration) {
	c.memory.Set(key, value, ttl)
	c.disk.Set(key, value, ttl)
}

// Delete 删除缓存
func (c *HybridCache) Delete(key string) {
	c.memory.Delete(key)
	c.disk.Delete(key)
}

// Clear 清空缓存
func (c *HybridCache) Clear() {
	c.memory.Clear()
	c.disk.Clear()
}
