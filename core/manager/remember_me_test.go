package manager

import (
	"testing"

	"github.com/sa-tokens/sa-token-go/core/config"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginRememberMe_Basic(t *testing.T) {
	storage := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.Timeout = 60             // 60s normal timeout
	cfg.RememberMeTimeout = 3600 // 1h remember-me
	mgr := NewManager(storage, cfg)

	token, err := mgr.LoginRememberMe("1001")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Token should be valid
	assert.True(t, mgr.IsLogin(token))

	// Should be remember-me login
	isRemember, err := mgr.IsRememberMeLogin(token)
	require.NoError(t, err)
	assert.True(t, isRemember)

	// TokenInfo should have IsRememberMe=true
	info, err := mgr.GetTokenInfo(token)
	require.NoError(t, err)
	assert.True(t, info.IsRememberMe)
	assert.Equal(t, "1001", info.LoginID)
}

func TestLoginRememberMe_NormalLogin_NotRememberMe(t *testing.T) {
	storage := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.Timeout = 60
	cfg.RememberMeTimeout = 3600
	mgr := NewManager(storage, cfg)

	// Normal login
	token, err := mgr.Login("1001")
	require.NoError(t, err)

	// Should NOT be remember-me
	isRemember, err := mgr.IsRememberMeLogin(token)
	require.NoError(t, err)
	assert.False(t, isRemember)

	// TokenInfo should have IsRememberMe=false (or omitted)
	info, err := mgr.GetTokenInfo(token)
	require.NoError(t, err)
	assert.False(t, info.IsRememberMe)
}

func TestLoginRememberMe_WithDevice(t *testing.T) {
	storage := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.Timeout = 60
	cfg.RememberMeTimeout = 3600
	mgr := NewManager(storage, cfg)

	token, err := mgr.LoginRememberMe("1001", "mobile")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	info, err := mgr.GetTokenInfo(token)
	require.NoError(t, err)
	assert.Equal(t, "1001", info.LoginID)
	assert.Equal(t, "mobile", info.Device)
	assert.True(t, info.IsRememberMe)
}

func TestIsRememberMeLogin_InvalidToken(t *testing.T) {
	storage := memory.NewStorage()
	cfg := config.DefaultConfig()
	mgr := NewManager(storage, cfg)

	_, err := mgr.IsRememberMeLogin("nonexistent-token")
	assert.Error(t, err)
}

func TestLoginRememberMe_TimeoutConfig(t *testing.T) {
	storage := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.Timeout = 60
	cfg.RememberMeTimeout = 7200 // 2h
	mgr := NewManager(storage, cfg)

	token, err := mgr.LoginRememberMe("1001")
	require.NoError(t, err)

	// Token should be accessible
	assert.True(t, mgr.IsLogin(token))
}

func TestLoginRememberMe_DefaultTimeout(t *testing.T) {
	storage := memory.NewStorage()
	cfg := config.DefaultConfig()
	// Default RememberMeTimeout is 604800 (7 days)
	mgr := NewManager(storage, cfg)

	token, err := mgr.LoginRememberMe("1001")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	isRemember, err := mgr.IsRememberMeLogin(token)
	require.NoError(t, err)
	assert.True(t, isRemember)
}
