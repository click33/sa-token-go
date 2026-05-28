package session

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sa-tokens/sa-token-go/core/adapter"
)

// Error variables (session cannot import root core: import cycle via core/satoken → manager → session)
// 错误变量（session 不能导入根 core：core/satoken → manager → session 会循环依赖）
var (
	ErrSessionNotFound    = fmt.Errorf("session not found: the session may have expired or been deleted")
	ErrInvalidSessionData = fmt.Errorf("invalid session data: stored payload is malformed")
)

// Session stores user data for Account-Session, Token-Session, or Custom-Session | 会话对象
type Session struct {
	ID         string          `json:"id"`
	Type       string          `json:"type,omitempty"`
	LoginType  string          `json:"loginType,omitempty"`
	LoginID    string          `json:"loginId,omitempty"`
	Token      string          `json:"token,omitempty"`
	CreateTime int64           `json:"createTime"`
	Data       map[string]any  `json:"data"`
	mu         sync.RWMutex    `json:"-"`
	storage    adapter.Storage `json:"-"`
	prefix     string          `json:"-"`
	keyPrefix  string          `json:"-"` // not persisted; derived on load | 存储前缀（不落库，加载时推导）
}

// NewSession creates an Account-Session keyed by loginId | 创建账号会话
func NewSession(id string, storage adapter.Storage, prefix string) *Session {
	return &Session{
		ID:         id,
		Type:       TypeAccount,
		LoginID:    id,
		CreateTime: time.Now().Unix(),
		Data:       make(map[string]any),
		storage:    storage,
		prefix:     prefix,
		keyPrefix:  AccountSessionKeyPrefix,
	}
}

// NewTokenSession creates a Token-Session keyed by token value | 创建 Token 会话
func NewTokenSession(tokenValue string, storage adapter.Storage, prefix, loginType string) *Session {
	if loginType == "" {
		loginType = "login"
	}
	return &Session{
		ID:         tokenValue,
		Type:       TypeToken,
		LoginType:  loginType,
		Token:      tokenValue,
		CreateTime: time.Now().Unix(),
		Data:       make(map[string]any),
		storage:    storage,
		prefix:     prefix,
		keyPrefix:  TokenSessionKeyPrefix,
	}
}

// NewAnonTokenSession creates an anonymous Token-Session | 创建匿名 Token 会话
func NewAnonTokenSession(tokenValue string, storage adapter.Storage, prefix, loginType string) *Session {
	if loginType == "" {
		loginType = "login"
	}
	return &Session{
		ID:         tokenValue,
		Type:       TypeAnon,
		LoginType:  loginType,
		Token:      tokenValue,
		CreateTime: time.Now().Unix(),
		Data:       make(map[string]any),
		storage:    storage,
		prefix:     prefix,
		keyPrefix:  TokenSessionKeyPrefix,
	}
}

func (s *Session) normalizeMeta() {
	if s.keyPrefix == "" {
		s.keyPrefix = AccountSessionKeyPrefix
	}
	if s.Type == "" {
		if s.keyPrefix == TokenSessionKeyPrefix {
			s.Type = TypeToken
		} else if s.keyPrefix == CustomSessionKeyPrefix {
			s.Type = TypeCustom
		} else {
			s.Type = TypeAccount
		}
	}
}

// Persist saves the session with optional TTL | 持久化会话
func (s *Session) Persist(ttl time.Duration) error {
	if ttl > 0 {
		return s.saveWithTTL(ttl)
	}
	return s.save()
}

// ============ Data Operations | 数据操作 ============

// Set Sets value | 设置值
func (s *Session) Set(key string, value any, ttl ...time.Duration) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.Data[key] = value
	if len(ttl) > 0 && ttl[0] > 0 {
		return s.saveWithTTL(ttl[0])
	}

	return s.save()
}

// SetMulti sets multiple key-value pairs | 设置多个键值对
func (s *Session) SetMulti(values map[string]any, ttl ...time.Duration) error {
	if len(values) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for key, value := range values {
		if key == "" {
			return fmt.Errorf("key cannot be empty")
		}
		s.Data[key] = value
	}

	if len(ttl) > 0 && ttl[0] > 0 {
		return s.saveWithTTL(ttl[0])
	}

	return s.save()
}

// Get Gets value | 获取值
func (s *Session) Get(key string) (any, bool) {
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

// Clear Clears all data | 清空所有数据
func (s *Session) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Data = make(map[string]any)
	return s.save()
}

// Keys Gets all keys | 获取所有键
func (s *Session) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.Data))
	for key := range s.Data {
		keys = append(keys, key)
	}
	return keys
}

// Size Gets data count | 获取数据数量
func (s *Session) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.Data)
}

// IsEmpty Checks if session has no data | 检查Session是否为空
func (s *Session) IsEmpty() bool {
	return s.Size() == 0
}

// Renew extends the session TTL without modifying content | 续期 Session 的 TTL，但不修改内容
func (s *Session) Renew(ttl time.Duration) error {
	if ttl <= 0 {
		return nil // 不允许设置 0 TTL，避免误删
	}

	key := s.getStorageKey()
	return s.storage.Expire(key, ttl)
}

// ============ Internal Methods | 内部方法 ============

// save Saves session to storage | 保存到存储
func (s *Session) save() error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidSessionData, err)
	}

	key := s.getStorageKey()
	return s.storage.Set(key, string(data), 0)
}

// saveWithTTL saves session with TTL | 带 TTL 保存 Session
func (s *Session) saveWithTTL(ttl time.Duration) error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidSessionData, err)
	}

	key := s.getStorageKey()
	return s.storage.Set(key, string(data), ttl)
}

// getStorageKey Gets storage key for this session | 获取Session的存储键
func (s *Session) getStorageKey() string {
	s.normalizeMeta()
	return s.prefix + s.keyPrefix + s.ID
}

// ============ Static Methods | 静态方法 ============

// Load Loads session from storage | 从存储加载
func Load(id string, storage adapter.Storage, prefix string) (*Session, error) {
	if id == "" {
		return nil, fmt.Errorf("session id cannot be empty")
	}

	key := prefix + AccountSessionKeyPrefix + id
	data, err := storage.Get(key)
	if err != nil && data != nil {
		return nil, err
	}
	if data == nil {
		return nil, ErrSessionNotFound
	}

	var (
		raw     []byte
		session Session
	)

	// Support both string and []byte | 同时兼容 string 和 []byte
	switch v := data.(type) {
	case string:
		raw = []byte(v)

	case []byte:
		raw = v

	default:
		return nil, ErrInvalidSessionData
	}

	if err := json.Unmarshal(raw, &session); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSessionData, err)
	}

	session.storage = storage
	session.prefix = prefix
	session.keyPrefix = AccountSessionKeyPrefix
	session.normalizeMeta()
	return &session, nil
}

// LoadWithKeyPrefix loads a session using an explicit storage namespace | 按前缀加载会话
func LoadWithKeyPrefix(id string, storage adapter.Storage, prefix, keyPrefix string) (*Session, error) {
	if id == "" {
		return nil, fmt.Errorf("session id cannot be empty")
	}
	if keyPrefix == "" {
		keyPrefix = AccountSessionKeyPrefix
	}

	key := prefix + keyPrefix + id
	data, err := storage.Get(key)
	if err != nil && data != nil {
		return nil, err
	}
	if data == nil {
		// Many backends (e.g. memory) return a typed "not found" error with nil data.
		// Treat as missing session so callers can use errors.Is(ErrSessionNotFound).
		// 常见存储在 key 不存在时返回 err+nil data，这里统一映射为 ErrSessionNotFound
		return nil, ErrSessionNotFound
	}

	var (
		raw     []byte
		session Session
	)

	switch v := data.(type) {
	case string:
		raw = []byte(v)
	case []byte:
		raw = v
	default:
		return nil, ErrInvalidSessionData
	}

	if err := json.Unmarshal(raw, &session); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSessionData, err)
	}

	session.storage = storage
	session.prefix = prefix
	session.keyPrefix = keyPrefix
	session.normalizeMeta()
	return &session, nil
}

// Destroy Destroys session | 销毁Session
func (s *Session) Destroy() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.getStorageKey()
	return s.storage.Delete(key)
}
