package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Cache 通用缓存接口
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
}

// MemoryCache 内存缓存实现
type MemoryCache struct {
	data map[string]cacheItem
	mu   sync.RWMutex
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// NewMemoryCache 创建新的内存缓存
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string]cacheItem),
	}
	go cache.cleanup()
	return cache
}

// Get 获取缓存值
func (c *MemoryCache) Get(key string) (interface{}, bool) {
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
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
}

// Delete 删除缓存
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// Clear 清空缓存
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]cacheItem)
}

// cleanup 定期清理过期缓存
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.data {
			if now.After(item.expiration) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

// HashKey 生成缓存键
func HashKey(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// SkillCache Skill 结果缓存
type SkillCache struct {
	cache Cache
}

// NewSkillCache 创建 Skill 缓存
func NewSkillCache() *SkillCache {
	return &SkillCache{
		cache: NewMemoryCache(),
	}
}

// GetSkillResult 获取 Skill 执行结果
func (c *SkillCache) GetSkillResult(skill string, args map[string]interface{}) (string, bool) {
	key := fmt.Sprintf("skill:%s:%v", skill, args)
	hashKey := HashKey(key)

	value, ok := c.cache.Get(hashKey)
	if !ok {
		return "", false
	}

	result, ok := value.(string)
	return result, ok
}

// SetSkillResult 缓存 Skill 执行结果
func (c *SkillCache) SetSkillResult(skill string, args map[string]interface{}, result string, ttl time.Duration) {
	key := fmt.Sprintf("skill:%s:%v", skill, args)
	hashKey := HashKey(key)
	c.cache.Set(hashKey, result, ttl)
}
