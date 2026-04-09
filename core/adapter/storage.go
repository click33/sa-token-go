package adapter

import (
	"context"
	"time"
)

// Storage defines storage interface for token and session data Storage 定义用于存储 Token 和 Session 数据的接口
type Storage interface {
	// Set stores key-value pair with optional expiration Set 设置键值对并可选指定过期时间
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	// Get gets value by key Get 获取键对应的值
	Get(ctx context.Context, key string) (any, error)
	// GetAndDelete gets and deletes key atomically GetAndDelete 原子获取并删除键
	GetAndDelete(ctx context.Context, key string) (any, error)
	// Delete deletes one or more keys Delete 删除一个或多个键
	Delete(ctx context.Context, keys ...string) error
	// Exists checks whether key exists Exists 检查键是否存在
	Exists(ctx context.Context, key string) bool

	// Keys gets all keys matching pattern Keys 获取匹配模式的所有键
	Keys(ctx context.Context, pattern string) ([]string, error)
	// Expire sets key expiration Expire 设置键的过期时间
	Expire(ctx context.Context, key string, expiration time.Duration) error
	// TTL gets remaining lifetime of key TTL 获取键的剩余生存时间
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Clear clears all data Clear 清空所有数据
	Clear(ctx context.Context) error
	// Ping checks whether storage is reachable Ping 检查存储是否可访问
	Ping(ctx context.Context) error
}
