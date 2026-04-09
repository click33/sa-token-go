package nonce

import (
	"time"
)

const (
	// DefaultNonceTTL stores default nonce ttl DefaultNonceTTL 存储默认 nonce 过期时间
	DefaultNonceTTL = 5 * time.Minute
	// NonceLength stores nonce byte length NonceLength 存储 Nonce 字节长度
	NonceLength = 32
	// NonceKeySuffix stores nonce key suffix NonceKeySuffix 存储 Nonce 键后缀
	NonceKeySuffix = "nonce:"
)

const (
	// DefaultRefreshTTL stores default refresh ttl DefaultRefreshTTL 存储默认刷新令牌过期时间
	DefaultRefreshTTL = 30 * 24 * time.Hour
	// DefaultAccessTTL stores default access ttl DefaultAccessTTL 存储默认访问令牌过期时间
	DefaultAccessTTL = 2 * time.Hour
	// RefreshTokenLength stores refresh token length RefreshTokenLength 存储刷新令牌字节长度
	RefreshTokenLength = 32
	// RefreshKeySuffix stores refresh key suffix RefreshKeySuffix 存储刷新令牌键后缀
	RefreshKeySuffix = "refresh:"
)
