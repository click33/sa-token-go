package memory

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/click33/sa-token-go/core/adapter"
)

// item 存储项
type item struct {
	value      interface{}
	expiration int64 // 过期时间戳（0表示永不过期）
}

// isExpired 检查是否过期
func (i *item) isExpired() bool {
	if i.expiration == 0 {
		return false
	}
	return time.Now().Unix() > i.expiration
}

// Storage 内存存储实现
type Storage struct {
	data map[string]*item
	mu   sync.RWMutex
}

// NewStorage 创建内存存储
func NewStorage() adapter.Storage {
	s := &Storage{
		data: make(map[string]*item),
	}
	// 启动清理协程
	go s.cleanup()
	return s
}

// Set 设置键值对
func (s *Storage) Set(key string, value interface{}, expiration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var exp int64
	if expiration > 0 {
		exp = time.Now().Add(expiration).Unix()
	}

	s.data[key] = &item{
		value:      value,
		expiration: exp,
	}

	return nil
}

// Get 获取值
func (s *Storage) Get(key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	if item.isExpired() {
		return nil, fmt.Errorf("key expired: %s", key)
	}

	return item.value, nil
}

// Delete 删除键
func (s *Storage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
	return nil
}

// Exists 检查键是否存在
func (s *Storage) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.data[key]
	if !exists {
		return false
	}

	if item.isExpired() {
		return false
	}

	return true
}

// Keys 获取匹配模式的所有键
func (s *Storage) Keys(pattern string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var keys []string
	for key, item := range s.data {
		if item.isExpired() {
			continue
		}
		if matchPattern(key, pattern) {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// Expire 设置键的过期时间
func (s *Storage) Expire(key string, expiration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	if expiration > 0 {
		item.expiration = time.Now().Add(expiration).Unix()
	} else {
		item.expiration = 0
	}

	return nil
}

// TTL 获取键的剩余生存时间
func (s *Storage) TTL(key string) (time.Duration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.data[key]
	if !exists {
		return -2 * time.Second, fmt.Errorf("key not found: %s", key)
	}

	if item.expiration == 0 {
		return -1 * time.Second, nil // 永不过期
	}

	ttl := item.expiration - time.Now().Unix()
	if ttl < 0 {
		return -2 * time.Second, nil // 已过期
	}

	return time.Duration(ttl) * time.Second, nil
}

// Clear 清空所有数据
func (s *Storage) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[string]*item)
	return nil
}

// cleanup 定期清理过期数据
func (s *Storage) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		for key, item := range s.data {
			if item.isExpired() {
				delete(s.data, key)
			}
		}
		s.mu.Unlock()
	}
}

// matchPattern 简单的模式匹配（支持 * 通配符）
func matchPattern(key, pattern string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}

	// 移除模式前缀的 **/
	pattern = strings.TrimPrefix(pattern, "**/")

	// 支持前缀匹配
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(key, prefix)
	}

	// 支持后缀匹配
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(key, suffix)
	}

	// 支持包含匹配
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(key, parts[0]) && strings.HasSuffix(key, parts[1])
		}
	}

	// 精确匹配
	return key == pattern
}
