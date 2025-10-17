package session

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/click33/sa-token-go/core/adapter"
)

// Session session object for storing user data | 会话对象，用于存储用户数据
type Session struct {
	ID         string                 `json:"id"`         // Session ID | Session标识
	CreateTime int64                  `json:"createTime"` // Creation time | 创建时间
	Data       map[string]interface{} `json:"data"`       // Session data | 数据
	mu         sync.RWMutex           `json:"-"`          // Read-write lock | 读写锁
	storage    adapter.Storage        `json:"-"`          // Storage backend | 存储
	prefix     string                 `json:"-"`          // Key prefix | 键前缀
}

// NewSession creates a new session | 创建新的Session
func NewSession(id string, storage adapter.Storage, prefix string) *Session {
	return &Session{
		ID:         id,
		CreateTime: time.Now().Unix(),
		Data:       make(map[string]interface{}),
		storage:    storage,
		prefix:     prefix,
	}
}

// Set sets value | 设置值
func (s *Session) Set(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Data[key] = value
	return s.save()
}

// Get gets value | 获取值
func (s *Session) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.Data[key]
	return value, exists
}

// GetString gets string value | 获取字符串值
func (s *Session) GetString(key string) string {
	if value, exists := s.Get(key); exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt gets integer value | 获取整数值
func (s *Session) GetInt(key string) int {
	if value, exists := s.Get(key); exists {
		switch v := value.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

// GetInt64 获取int64值
func (s *Session) GetInt64(key string) int64 {
	if value, exists := s.Get(key); exists {
		switch v := value.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return 0
}

// GetBool 获取布尔值
func (s *Session) GetBool(key string) bool {
	if value, exists := s.Get(key); exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

// Has 检查键是否存在
func (s *Session) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.Data[key]
	return exists
}

// Delete 删除键
func (s *Session) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.Data, key)
	return s.save()
}

// Clear 清空所有数据
func (s *Session) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Data = make(map[string]interface{})
	return s.save()
}

// Keys 获取所有键
func (s *Session) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.Data))
	for key := range s.Data {
		keys = append(keys, key)
	}
	return keys
}

// Size 获取数据数量
func (s *Session) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.Data)
}

// save 保存到存储
func (s *Session) save() error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := s.prefix + "session:" + s.ID
	return s.storage.Set(key, string(data), 0)
}

// Load 从存储加载
func Load(id string, storage adapter.Storage, prefix string) (*Session, error) {
	key := prefix + "session:" + id
	data, err := storage.Get(key)
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, fmt.Errorf("session not found")
	}

	var session Session
	if err := json.Unmarshal([]byte(data.(string)), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	session.storage = storage
	session.prefix = prefix
	return &session, nil
}

// Destroy 销毁Session
func (s *Session) Destroy() error {
	key := s.prefix + "session:" + s.ID
	return s.storage.Delete(key)
}
