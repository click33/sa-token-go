package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/click33/sa-token-go/core/adapter"
	"github.com/redis/go-redis/v9"
)

// Storage Redis存储实现
type Storage struct {
	client    *redis.Client
	ctx       context.Context
	keyPrefix string
}

// Config Redis配置
type Config struct {
	Host     string
	Port     int
	Password string
	Database int
	PoolSize int
}

// NewStorage 通过Redis URL创建存储
func NewStorage(url string, keyPrefix string) (adapter.Storage, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}

	client := redis.NewClient(opts)

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Storage{
		client:    client,
		ctx:       ctx,
		keyPrefix: keyPrefix,
	}, nil
}

// NewStorageFromConfig 通过配置创建存储
func NewStorageFromConfig(cfg *Config, keyPrefix string) (adapter.Storage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.Database,
		PoolSize: cfg.PoolSize,
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Storage{
		client:    client,
		ctx:       ctx,
		keyPrefix: keyPrefix,
	}, nil
}

// NewStorageFromClient 从已有的Redis客户端创建存储
func NewStorageFromClient(client *redis.Client, keyPrefix string) adapter.Storage {
	return &Storage{
		client:    client,
		ctx:       context.Background(),
		keyPrefix: keyPrefix,
	}
}

// getKey 获取完整的键名
func (s *Storage) getKey(key string) string {
	return s.keyPrefix + key
}

// Set 设置键值对
func (s *Storage) Set(key string, value interface{}, expiration time.Duration) error {
	return s.client.Set(s.ctx, s.getKey(key), value, expiration).Err()
}

// Get 获取值
func (s *Storage) Get(key string) (interface{}, error) {
	val, err := s.client.Get(s.ctx, s.getKey(key)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	if err != nil {
		return nil, err
	}
	return val, nil
}

// Delete 删除键
func (s *Storage) Delete(key string) error {
	return s.client.Del(s.ctx, s.getKey(key)).Err()
}

// Exists 检查键是否存在
func (s *Storage) Exists(key string) bool {
	result, err := s.client.Exists(s.ctx, s.getKey(key)).Result()
	if err != nil {
		return false
	}
	return result > 0
}

// Keys 获取匹配模式的所有键
func (s *Storage) Keys(pattern string) ([]string, error) {
	fullPattern := s.getKey(pattern)
	keys, err := s.client.Keys(s.ctx, fullPattern).Result()
	if err != nil {
		return nil, err
	}

	// 移除键前缀
	result := make([]string, len(keys))
	prefixLen := len(s.keyPrefix)
	for i, key := range keys {
		if len(key) > prefixLen {
			result[i] = key[prefixLen:]
		} else {
			result[i] = key
		}
	}

	return result, nil
}

// Expire 设置键的过期时间
func (s *Storage) Expire(key string, expiration time.Duration) error {
	return s.client.Expire(s.ctx, s.getKey(key), expiration).Err()
}

// TTL 获取键的剩余生存时间
func (s *Storage) TTL(key string) (time.Duration, error) {
	return s.client.TTL(s.ctx, s.getKey(key)).Result()
}

// Clear 清空所有数据（使用前缀匹配删除）
func (s *Storage) Clear() error {
	pattern := s.keyPrefix + "*"
	keys, err := s.client.Keys(s.ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return s.client.Del(s.ctx, keys...).Err()
	}

	return nil
}

// Close 关闭连接
func (s *Storage) Close() error {
	return s.client.Close()
}

// GetClient 获取Redis客户端（用于高级操作）
func (s *Storage) GetClient() *redis.Client {
	return s.client
}

// Builder Redis存储构建器
type Builder struct {
	host     string
	port     int
	password string
	database int
	poolSize int
	prefix   string
}

// NewBuilder 创建构建器
func NewBuilder() *Builder {
	return &Builder{
		host:     "localhost",
		port:     6379,
		password: "",
		database: 0,
		poolSize: 10,
		prefix:   "satoken:",
	}
}

// Host 设置主机
func (b *Builder) Host(host string) *Builder {
	b.host = host
	return b
}

// Port 设置端口
func (b *Builder) Port(port int) *Builder {
	b.port = port
	return b
}

// Password 设置密码
func (b *Builder) Password(password string) *Builder {
	b.password = password
	return b
}

// Database 设置数据库
func (b *Builder) Database(database int) *Builder {
	b.database = database
	return b
}

// PoolSize 设置连接池大小
func (b *Builder) PoolSize(poolSize int) *Builder {
	b.poolSize = poolSize
	return b
}

// KeyPrefix 设置键前缀
func (b *Builder) KeyPrefix(prefix string) *Builder {
	b.prefix = prefix
	return b
}

// Build 构建存储
func (b *Builder) Build() (adapter.Storage, error) {
	return NewStorageFromConfig(&Config{
		Host:     b.host,
		Port:     b.port,
		Password: b.password,
		Database: b.database,
		PoolSize: b.poolSize,
	}, b.prefix)
}
