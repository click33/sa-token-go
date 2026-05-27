package security

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sa-tokens/sa-token-go/core/adapter"
	"github.com/sa-tokens/sa-token-go/core/errs"
	"github.com/sa-tokens/sa-token-go/core/utils"
)

// ApiKeyInfo stores API key metadata | API Key 信息
type ApiKeyInfo struct {
	Key        string `json:"key"`
	LoginID    string `json:"loginId"`
	Title      string `json:"title,omitempty"`
	CreateTime int64  `json:"createTime"`
	ExpireTime int64  `json:"expireTime,omitempty"` // 0 means never expire
	Disabled   bool   `json:"disabled,omitempty"`
	Extra      string `json:"extra,omitempty"`
}

// ApiKeyManager manages API keys | API Key 管理器
type ApiKeyManager struct {
	storage   adapter.Storage
	keyPrefix string
}

// NewApiKeyManager creates a new API key manager | 创建 API Key 管理器
func NewApiKeyManager(storage adapter.Storage, prefix string) *ApiKeyManager {
	return &ApiKeyManager{
		storage:   storage,
		keyPrefix: prefix + "apikey:",
	}
}

// CreateApiKey creates a new API key | 创建 API Key
func (m *ApiKeyManager) CreateApiKey(loginID, title string, expireSeconds int64, extra string) (*ApiKeyInfo, error) {
	key := utils.RandomString(32)
	now := time.Now().Unix()

	info := &ApiKeyInfo{
		Key:        key,
		LoginID:    loginID,
		Title:      title,
		CreateTime: now,
		Extra:      extra,
	}

	if expireSeconds > 0 {
		info.ExpireTime = now + expireSeconds
	}

	data, err := json.Marshal(info)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal api key info: %w", err)
	}

	ttl := time.Duration(0)
	if expireSeconds > 0 {
		ttl = time.Duration(expireSeconds) * time.Second
	}

	if err := m.storage.Set(m.keyPrefix+key, string(data), ttl); err != nil {
		return nil, fmt.Errorf("failed to store api key: %w", err)
	}

	return info, nil
}

// GetApiKeyInfo retrieves API key info | 获取 API Key 信息
func (m *ApiKeyManager) GetApiKeyInfo(key string) (*ApiKeyInfo, error) {
	data, err := m.storage.Get(m.keyPrefix + key)
	if err != nil {
		return nil, errs.ErrApiKeyNotFound
	}
	if data == nil {
		return nil, errs.ErrApiKeyNotFound
	}

	var info ApiKeyInfo
	var bytes []byte
	switch v := data.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return nil, fmt.Errorf("unexpected type for api key data")
	}

	if err := json.Unmarshal(bytes, &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal api key info: %w", err)
	}

	return &info, nil
}

// VerifyApiKey verifies an API key and returns its info | 验证 API Key 并返回信息
func (m *ApiKeyManager) VerifyApiKey(key string) (*ApiKeyInfo, error) {
	info, err := m.GetApiKeyInfo(key)
	if err != nil {
		return nil, err
	}

	if info.Disabled {
		return nil, errs.ErrApiKeyDisabled
	}

	if info.ExpireTime > 0 && time.Now().Unix() > info.ExpireTime {
		return nil, errs.ErrApiKeyExpired
	}

	return info, nil
}

// DeleteApiKey deletes an API key | 删除 API Key
func (m *ApiKeyManager) DeleteApiKey(key string) error {
	return m.storage.Delete(m.keyPrefix + key)
}

// DisableApiKey disables an API key | 禁用 API Key
func (m *ApiKeyManager) DisableApiKey(key string) error {
	info, err := m.GetApiKeyInfo(key)
	if err != nil {
		return err
	}
	info.Disabled = true

	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	ttl := time.Duration(0)
	if info.ExpireTime > 0 {
		remaining := info.ExpireTime - time.Now().Unix()
		if remaining > 0 {
			ttl = time.Duration(remaining) * time.Second
		}
	}

	return m.storage.Set(m.keyPrefix+key, string(data), ttl)
}

// EnableApiKey re-enables a disabled API key | 启用 API Key
func (m *ApiKeyManager) EnableApiKey(key string) error {
	info, err := m.GetApiKeyInfo(key)
	if err != nil {
		return err
	}
	info.Disabled = false

	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	ttl := time.Duration(0)
	if info.ExpireTime > 0 {
		remaining := info.ExpireTime - time.Now().Unix()
		if remaining > 0 {
			ttl = time.Duration(remaining) * time.Second
		}
	}

	return m.storage.Set(m.keyPrefix+key, string(data), ttl)
}
