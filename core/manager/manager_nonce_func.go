package manager

import (
	"context"
	"time"
)

// GenerateNonce generates nonce GenerateNonce 生成新的 nonce 并使用默认有效期
func (m *Manager) GenerateNonce(ctx context.Context) (string, error) {
	return m.nonceManager.Generate(ctx)
}

// GenerateNonceWithTimeout generates nonce with timeout GenerateNonceWithTimeout 使用指定有效期生成新的 nonce
func (m *Manager) GenerateNonceWithTimeout(ctx context.Context, timeout time.Duration) (string, error) {
	return m.nonceManager.GenerateWithTimeout(ctx, timeout)
}

// VerifyNonce verifies and consumes nonce VerifyNonce 验证并消费一次性 nonce
func (m *Manager) VerifyNonce(ctx context.Context, nonce string) bool {
	return m.nonceManager.Verify(ctx, nonce)
}

// VerifyAndConsumeNonce verifies nonce with error VerifyAndConsumeNonce 验证并消费 nonce 且在无效时返回错误
func (m *Manager) VerifyAndConsumeNonce(ctx context.Context, nonce string) error {
	return m.nonceManager.VerifyAndConsume(ctx, nonce)
}

// IsNonceValid checks nonce validity IsNonceValid 检查 nonce 是否有效且不消费
func (m *Manager) IsNonceValid(ctx context.Context, nonce string) bool {
	return m.nonceManager.IsValid(ctx, nonce)
}

// GetNonceTTL gets nonce ttl GetNonceTTL 获取 nonce 的剩余有效时间秒数
func (m *Manager) GetNonceTTL(ctx context.Context, nonce string) (int64, error) {
	return m.nonceManager.GetTTL(ctx, nonce)
}
