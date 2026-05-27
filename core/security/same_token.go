package security

import (
	"fmt"
	"time"

	"github.com/sa-tokens/sa-token-go/core/adapter"
	"github.com/sa-tokens/sa-token-go/core/errs"
	"github.com/sa-tokens/sa-token-go/core/utils"
)

const (
	// SameTokenHeader is the HTTP header name for same-token | 服务间调用令牌的 HTTP 头名
	SameTokenHeader = "SA-SAME-TOKEN"

	sameTokenKey     = "var:same-token"
	pastSameTokenKey = "var:past-same-token"
)

// SameTokenTemplate manages service-to-service authentication tokens | 服务间调用令牌管理器
type SameTokenTemplate struct {
	storage   adapter.Storage
	keyPrefix string
	timeout   time.Duration
}

// NewSameTokenTemplate creates a new same-token template | 创建服务间调用令牌管理器
func NewSameTokenTemplate(storage adapter.Storage, prefix string, timeout time.Duration) *SameTokenTemplate {
	if timeout <= 0 {
		timeout = 24 * time.Hour
	}
	return &SameTokenTemplate{
		storage:   storage,
		keyPrefix: prefix + ":",
		timeout:   timeout,
	}
}

// GetToken returns the current same-token, creating one if needed | 获取当前令牌（不存在则自动创建）
func (t *SameTokenTemplate) GetToken() (string, error) {
	val, err := t.storage.Get(t.keyPrefix + sameTokenKey)
	if err == nil && val != nil {
		if s, ok := val.(string); ok && s != "" {
			return s, nil
		}
	}
	// No token exists, create one
	return t.refreshToken()
}

// RefreshToken rotates the token: current → past, generate new | 刷新令牌（当前令牌变为旧令牌，生成新令牌）
func (t *SameTokenTemplate) RefreshToken() (string, error) {
	return t.refreshToken()
}

func (t *SameTokenTemplate) refreshToken() (string, error) {
	// Archive current token as past
	current, _ := t.getTokenFromStorage(sameTokenKey)
	if current != "" {
		ttl, _ := t.storage.TTL(t.keyPrefix + sameTokenKey)
		if ttl > 0 {
			_ = t.storage.Set(t.keyPrefix+pastSameTokenKey, current, ttl)
		} else {
			_ = t.storage.Set(t.keyPrefix+pastSameTokenKey, current, t.timeout)
		}
	}

	// Generate new token
	newToken := utils.RandomString(64)
	if err := t.storage.Set(t.keyPrefix+sameTokenKey, newToken, t.timeout); err != nil {
		return "", fmt.Errorf("failed to store same-token: %w", err)
	}
	return newToken, nil
}

// CheckToken validates a same-token value | 验证令牌
func (t *SameTokenTemplate) CheckToken(tokenValue string) error {
	if tokenValue == "" {
		return errs.ErrSameTokenInvalid
	}

	// Check current token
	current, _ := t.getTokenFromStorage(sameTokenKey)
	if current != "" && current == tokenValue {
		return nil
	}

	// Check past token (grace window)
	past, _ := t.getTokenFromStorage(pastSameTokenKey)
	if past != "" && past == tokenValue {
		return nil
	}

	return errs.ErrSameTokenInvalid
}

// IsValid checks if a same-token is valid without returning error | 检查令牌是否有效
func (t *SameTokenTemplate) IsValid(tokenValue string) bool {
	return t.CheckToken(tokenValue) == nil
}

func (t *SameTokenTemplate) getTokenFromStorage(key string) (string, error) {
	val, err := t.storage.Get(t.keyPrefix + key)
	if err != nil || val == nil {
		return "", err
	}
	if s, ok := val.(string); ok {
		return s, nil
	}
	return "", nil
}
