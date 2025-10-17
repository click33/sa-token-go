package adapter

import "time"

// Storage defines storage interface for Token and Session data | 定义存储接口，用于存储Token和Session数据
type Storage interface {
	// Set sets key-value pair with optional expiration time (0 means never expire) | 设置键值对，可选过期时间（0表示永不过期）
	Set(key string, value interface{}, expiration time.Duration) error

	// Get gets value by key | 获取键对应的值
	Get(key string) (interface{}, error)

	// Delete deletes key | 删除键
	Delete(key string) error

	// Exists checks if key exists | 检查键是否存在
	Exists(key string) bool

	// Keys gets all keys matching pattern | 获取匹配模式的所有键
	Keys(pattern string) ([]string, error)

	// Expire sets expiration time for key | 设置键的过期时间
	Expire(key string, expiration time.Duration) error

	// TTL gets remaining time to live | 获取键的剩余生存时间
	TTL(key string) (time.Duration, error)

	// Clear clears all data (for testing) | 清空所有数据（用于测试）
	Clear() error
}
