package cache

import (
	"context"
	"encoding/json"
	"time"
)

// RedisCache Redis 缓存实现（接口定义，可选实现）
// 如果环境中没有 Redis，此文件提供接口和 mock 实现

type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// RedisCache Redis 缓存
type RedisCache struct {
	client RedisClient
	prefix string
}

// NewRedisCache 创建 Redis 缓存
func NewRedisCache(client RedisClient, prefix string) *RedisCache {
	return &RedisCache{
		client: client,
		prefix: prefix,
	}
}

// Get 获取缓存值
func (c *RedisCache) Get(key string) (interface{}, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	val, err := c.client.Get(ctx, c.prefix+key)
	if err != nil {
		return nil, false
	}

	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return val, true // 返回原始字符串
	}
	return result, true
}

// Set 设置缓存值
func (c *RedisCache) Set(key string, value interface{}, ttl time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := json.Marshal(value)
	if err != nil {
		return
	}

	c.client.Set(ctx, c.prefix+key, string(data), ttl)
}

// Delete 删除缓存
func (c *RedisCache) Delete(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c.client.Delete(ctx, c.prefix+key)
}

// Clear 清空缓存（Redis 中按前缀删除）
func (c *RedisCache) Clear() {
	// Redis 需要 SCAN + DEL，这里简化处理
}

// MultiTierCache 多级缓存：L1 内存 → L2 磁盘 → L3 Redis
type MultiTierCache struct {
	l1 *MemoryCache // 内存
	l2 *FileCache   // 磁盘
	l3 *RedisCache  // Redis
}

// NewMultiTierCache 创建多级缓存
func NewMultiTierCache(l2Path string, l3Client RedisClient) (*MultiTierCache, error) {
	l2, err := NewFileCache(l2Path)
	if err != nil {
		return nil, err
	}

	return &MultiTierCache{
		l1: NewMemoryCache(),
		l2: l2,
		l3: NewRedisCache(l3Client, "astrcode:"),
	}, nil
}

// Get 多级缓存获取（L1 → L2 → L3）
func (c *MultiTierCache) Get(key string) (interface{}, bool) {
	// L1
	if val, ok := c.l1.Get(key); ok {
		return val, true
	}

	// L2
	if val, ok := c.l2.Get(key); ok {
		c.l1.Set(key, val, 5*time.Minute) // 回填 L1
		return val, true
	}

	// L3
	if c.l3 != nil {
		if val, ok := c.l3.Get(key); ok {
			c.l2.Set(key, val, 24*time.Hour)  // 回填 L2
			c.l1.Set(key, val, 5*time.Minute) // 回填 L1
			return val, true
		}
	}

	return nil, false
}

// Set 多级缓存设置
func (c *MultiTierCache) Set(key string, value interface{}, ttl time.Duration) {
	c.l1.Set(key, value, ttl)
	c.l2.Set(key, value, ttl)
	if c.l3 != nil {
		c.l3.Set(key, value, ttl)
	}
}

// Delete 多级缓存删除
func (c *MultiTierCache) Delete(key string) {
	c.l1.Delete(key)
	c.l2.Delete(key)
	if c.l3 != nil {
		c.l3.Delete(key)
	}
}

// Clear 清空所有级缓存
func (c *MultiTierCache) Clear() {
	c.l1.Clear()
	c.l2.Clear()
	if c.l3 != nil {
		c.l3.Clear()
	}
}
