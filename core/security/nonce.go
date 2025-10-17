package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/click33/sa-token-go/core/adapter"
)

// Nonce Anti-Replay Attack Implementation
// Nonce 防重放攻击实现
//
// Flow | 流程:
// 1. Generate() - Create unique nonce and store with TTL | 生成唯一nonce并存储（带过期时间）
// 2. Verify() - Check existence and delete (one-time use) | 检查存在性并删除（一次性使用）
// 3. Auto-expire after TTL (default 5min) | TTL后自动过期（默认5分钟）
//
// Usage | 用法:
//   nonce, _ := manager.GenerateNonce()
//   valid := manager.VerifyNonce(nonce)  // true
//   valid = manager.VerifyNonce(nonce)   // false (replay prevented)

// NonceManager Nonce manager for anti-replay attacks | Nonce管理器，用于防重放攻击
type NonceManager struct {
	storage adapter.Storage
	ttl     time.Duration
	mu      sync.RWMutex
}

// NewNonceManager creates a new nonce manager | 创建新的Nonce管理器
// ttl: time to live, default 5 minutes | 过期时间，默认5分钟
func NewNonceManager(storage adapter.Storage, ttl time.Duration) *NonceManager {
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	return &NonceManager{
		storage: storage,
		ttl:     ttl,
	}
}

// Generate generates a new nonce and stores it | 生成新的nonce并存储
// Returns 64-char hex string | 返回64字符的十六进制字符串
func (nm *NonceManager) Generate() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	nonce := hex.EncodeToString(bytes)

	key := fmt.Sprintf("satoken:nonce:%s", nonce)
	if err := nm.storage.Set(key, time.Now().Unix(), nm.ttl); err != nil {
		return "", err
	}

	return nonce, nil
}

// Verify verifies nonce and consumes it (one-time use) | 验证nonce并消费它（一次性使用）
// Returns false if nonce doesn't exist or already used | 如果nonce不存在或已使用则返回false
func (nm *NonceManager) Verify(nonce string) bool {
	if nonce == "" {
		return false
	}

	key := fmt.Sprintf("satoken:nonce:%s", nonce)

	nm.mu.Lock()
	defer nm.mu.Unlock()

	if !nm.storage.Exists(key) {
		return false
	}

	nm.storage.Delete(key)
	return true
}

// VerifyAndConsume verifies and consumes nonce, returns error if invalid | 验证并消费nonce，无效时返回错误
func (nm *NonceManager) VerifyAndConsume(nonce string) error {
	if !nm.Verify(nonce) {
		return fmt.Errorf("invalid or expired nonce")
	}
	return nil
}

// Clean cleans expired nonces (handled by storage TTL) | 清理过期的nonce（由存储的TTL处理）
func (nm *NonceManager) Clean() {
}
