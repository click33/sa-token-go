package security

import (
	"testing"

	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestNonceManager() *NonceManager {
	st := memory.NewStorage()
	return NewNonceManager(st, "test:", DefaultNonceTTL)
}

func TestNonceManager_GenerateAndVerify(t *testing.T) {
	nm := newTestNonceManager()
	nonce, err := nm.Generate()
	require.NoError(t, err)
	assert.NotEmpty(t, nonce)
	assert.Len(t, nonce, 64) // 32 bytes hex = 64 chars

	// First verify should succeed
	assert.True(t, nm.Verify(nonce))

	// Second verify should fail (one-time use)
	assert.False(t, nm.Verify(nonce))
}

func TestNonceManager_VerifyEmpty(t *testing.T) {
	nm := newTestNonceManager()
	assert.False(t, nm.Verify(""))
}

func TestNonceManager_VerifyInvalid(t *testing.T) {
	nm := newTestNonceManager()
	assert.False(t, nm.Verify("nonexistent-nonce-value"))
}

func TestNonceManager_IsValid(t *testing.T) {
	nm := newTestNonceManager()
	nonce, err := nm.Generate()
	require.NoError(t, err)

	// IsValid should not consume
	assert.True(t, nm.IsValid(nonce))
	assert.True(t, nm.IsValid(nonce)) // still valid

	// Verify consumes it
	assert.True(t, nm.Verify(nonce))
	assert.False(t, nm.IsValid(nonce))
}

func TestNonceManager_VerifyAndConsume(t *testing.T) {
	nm := newTestNonceManager()
	nonce, err := nm.Generate()
	require.NoError(t, err)

	assert.NoError(t, nm.VerifyAndConsume(nonce))
	assert.Error(t, nm.VerifyAndConsume(nonce)) // already consumed
	assert.Error(t, nm.VerifyAndConsume("invalid"))
}

func TestNonceManager_IsValidEmpty(t *testing.T) {
	nm := newTestNonceManager()
	assert.False(t, nm.IsValid(""))
}
