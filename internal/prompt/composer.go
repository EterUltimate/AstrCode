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

package prompt

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// CacheLevel 缓存级别
type CacheLevel string

const (
	CacheLevelStable     CacheLevel = "stable"      // 稳定缓存（系统提示，极少变化）
	CacheLevelSemiStable CacheLevel = "semi_stable" // 半稳定缓存（项目配置，偶尔变化）
	CacheLevelInherited  CacheLevel = "inherited"   // 继承缓存（父 session 继承）
	CacheLevelDynamic    CacheLevel = "dynamic"     // 动态缓存（当前对话历史，每次变化）
)

// PromptBlock Prompt 块
type PromptBlock struct {
	ID         string                         // 唯一标识
	Content    string                         // 内容
	Priority   int                            // 优先级（越高越靠前）
	CacheLevel CacheLevel                     // 缓存级别
	Metadata   map[string]string              // 元数据
	Condition  func(ctx context.Context) bool // 条件渲染函数（可选）
}

// PromptContributor Prompt 贡献者接口
type PromptContributor interface {
	Name() string                                          // 贡献者名称
	Contribute(ctx context.Context) ([]PromptBlock, error) // 贡献 Prompt 块
}

// PromptCache Prompt 缓存
type PromptCache struct {
	mu    sync.RWMutex
	cache map[string]*CachedBlock
	ttl   time.Duration // 缓存过期时间
}

// CachedBlock 缓存的 Prompt 块
type CachedBlock struct {
	Block     PromptBlock
	Timestamp time.Time
	Checksum  string // 用于检测内容变化
}

// NewPromptCache 创建 Prompt 缓存
func NewPromptCache(ttl time.Duration) *PromptCache {
	if ttl == 0 {
		ttl = 5 * time.Minute // 默认 5 分钟
	}

	return &PromptCache{
		cache: make(map[string]*CachedBlock),
		ttl:   ttl,
	}
}

// Get 从缓存获取
func (pc *PromptCache) Get(key string) (*PromptBlock, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	cached, exists := pc.cache[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Since(cached.Timestamp) > pc.ttl {
		delete(pc.cache, key)
		return nil, false
	}

	return &cached.Block, true
}

// Set 设置缓存
func (pc *PromptCache) Set(key string, block PromptBlock) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.cache[key] = &CachedBlock{
		Block:     block,
		Timestamp: time.Now(),
		Checksum:  computeChecksum(block.Content),
	}
}

// Invalidate 使缓存失效
func (pc *PromptCache) Invalidate(key string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	delete(pc.cache, key)
}

// Clear 清空缓存
func (pc *PromptCache) Clear() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.cache = make(map[string]*CachedBlock)
}

// computeChecksum 计算内容校验和（简化版）
func computeChecksum(content string) string {
	// 实际项目中应该使用 crypto/sha256
	return fmt.Sprintf("%d", len(content))
}

// Composer Prompt 组装器
type Composer struct {
	contributors []PromptContributor
	cache        *PromptCache
	mu           sync.RWMutex
}

// NewComposer 创建 Prompt 组装器
func NewComposer(cache *PromptCache) *Composer {
	if cache == nil {
		cache = NewPromptCache(5 * time.Minute)
	}

	return &Composer{
		contributors: make([]PromptContributor, 0),
		cache:        cache,
	}
}

// RegisterContributor 注册贡献者
func (c *Composer) RegisterContributor(contributor PromptContributor) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.contributors = append(c.contributors, contributor)
}

// Assemble 组装 Prompt
func (c *Composer) Assemble(ctx context.Context) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var allBlocks []PromptBlock

	// 1. 收集所有 contributor 的 blocks
	for _, contributor := range c.contributors {
		blocks, err := contributor.Contribute(ctx)
		if err != nil {
			return "", fmt.Errorf("contributor '%s' failed: %w", contributor.Name(), err)
		}
		allBlocks = append(allBlocks, blocks...)
	}

	// 2. 应用条件过滤
	filteredBlocks := make([]PromptBlock, 0, len(allBlocks))
	for _, block := range allBlocks {
		if block.Condition != nil && !block.Condition(ctx) {
			continue // 条件不满足，跳过
		}
		filteredBlocks = append(filteredBlocks, block)
	}

	// 3. 去重（基于 ID）
	dedupedBlocks := c.deduplicateBlocks(filteredBlocks)

	// 4. 按 priority 排序（降序）
	sort.Slice(dedupedBlocks, func(i, j int) bool {
		return dedupedBlocks[i].Priority > dedupedBlocks[j].Priority
	})

	// 5. 组装最终 prompt
	var sb strings.Builder
	for i, block := range dedupedBlocks {
		if i > 0 {
			sb.WriteString("\n\n") // 块之间添加空行
		}
		sb.WriteString(block.Content)
	}

	return sb.String(), nil
}

// deduplicateBlocks 去重（保留高优先级的）
func (c *Composer) deduplicateBlocks(blocks []PromptBlock) []PromptBlock {
	seen := make(map[string]int) // ID -> index in result
	var result []PromptBlock

	for _, block := range blocks {
		if idx, exists := seen[block.ID]; exists {
			// 已存在，比较优先级
			if block.Priority > result[idx].Priority {
				result[idx] = block // 替换为更高优先级的
			}
		} else {
			// 新块
			seen[block.ID] = len(result)
			result = append(result, block)
		}
	}

	return result
}

// GetStats 获取组装器统计信息
func (c *Composer) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"contributors": len(c.contributors),
		"cache_size":   len(c.cache.cache),
	}
}

// ClearCache 清空缓存
func (c *Composer) ClearCache() {
	c.cache.Clear()
}
