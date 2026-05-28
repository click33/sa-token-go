package config

import (
	"testing"

	"github.com/sa-tokens/sa-token-go/core/pool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, DefaultTokenName, cfg.TokenName)
	assert.Equal(t, int64(DefaultTimeout), cfg.Timeout)
	assert.Equal(t, int64(DefaultTimeout/2), cfg.MaxRefresh)
	assert.Equal(t, int64(NoLimit), cfg.RenewInterval)
	assert.Equal(t, int64(NoLimit), cfg.ActiveTimeout)
	assert.True(t, cfg.IsConcurrent)
	assert.True(t, cfg.IsShare)
	assert.Equal(t, DefaultMaxLoginCount, cfg.MaxLoginCount)
	assert.False(t, cfg.IsReadBody)
	assert.True(t, cfg.IsReadHeader)
	assert.False(t, cfg.IsReadCookie)
	assert.Equal(t, TokenStyleUUID, cfg.TokenStyle)
	assert.True(t, cfg.TokenSessionCheckLogin)
	assert.True(t, cfg.AutoRenew)
	assert.Equal(t, "login", cfg.LoginType)
	assert.Equal(t, "CURR_DEVICE", cfg.ReplacedRange)
	assert.Equal(t, "LOGOUT", cfg.OverflowLogoutMode)
	assert.NotNil(t, cfg.CookieConfig)
}

func TestConfig_Validate_ValidConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.NoError(t, cfg.Validate())
}

func TestConfig_Validate_EmptyTokenName(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TokenName = ""
	assert.Error(t, cfg.Validate())
	assert.Contains(t, cfg.Validate().Error(), "TokenName")
}

func TestConfig_Validate_InvalidTokenStyle(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TokenStyle = "invalid"
	assert.Error(t, cfg.Validate())
	assert.Contains(t, cfg.Validate().Error(), "TokenStyle")
}

func TestConfig_Validate_JWTWithoutSecret(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TokenStyle = TokenStyleJWT
	cfg.JwtSecretKey = ""
	assert.Error(t, cfg.Validate())
	assert.Contains(t, cfg.Validate().Error(), "JwtSecretKey")
}

func TestConfig_Validate_TimeoutTooLow(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Timeout = -2
	assert.Error(t, cfg.Validate())
}

func TestConfig_Validate_MaxRefreshTooLow(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRefresh = -2
	assert.Error(t, cfg.Validate())
}

func TestConfig_Validate_NoReadSource(t *testing.T) {
	cfg := DefaultConfig()
	cfg.IsReadHeader = false
	cfg.IsReadCookie = false
	cfg.IsReadBody = false
	assert.Error(t, cfg.Validate())
	assert.Contains(t, cfg.Validate().Error(), "at least one")
}

func TestConfig_Validate_MaxRefreshAutoAdjust(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Timeout = 100
	cfg.MaxRefresh = 200 // greater than timeout
	require.NoError(t, cfg.Validate())
	assert.Equal(t, int64(50), cfg.MaxRefresh)
}

func TestConfig_Validate_RenewPoolConfig_Invalid(t *testing.T) {
	cfg := DefaultConfig()

	cfg.RenewPoolConfig = &pool.RenewPoolConfig{MinSize: 0}
	assert.Error(t, cfg.Validate())

	cfg.RenewPoolConfig = &pool.RenewPoolConfig{MinSize: 10, MaxSize: 5}
	assert.Error(t, cfg.Validate())

	cfg.RenewPoolConfig = &pool.RenewPoolConfig{MinSize: 10, MaxSize: 10, ScaleUpRate: 0}
	assert.Error(t, cfg.Validate())

	cfg.RenewPoolConfig = &pool.RenewPoolConfig{
		MinSize: 10, MaxSize: 10, ScaleUpRate: 0.5, ScaleDownRate: -1,
		CheckInterval: 1, Expiry: 1,
	}
	assert.Error(t, cfg.Validate())
}

func TestConfig_Clone(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TokenName = "test"
	cfg.CookieConfig.Domain = "example.com"

	clone := cfg.Clone()
	assert.Equal(t, "test", clone.TokenName)
	assert.Equal(t, "example.com", clone.CookieConfig.Domain)

	// Modify clone, original should not change
	clone.TokenName = "modified"
	clone.CookieConfig.Domain = "other.com"
	assert.Equal(t, "test", cfg.TokenName)
	assert.Equal(t, "example.com", cfg.CookieConfig.Domain)
}

func TestConfig_EffectiveLoginType(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "login", cfg.EffectiveLoginType())

	cfg.LoginType = "admin"
	assert.Equal(t, "admin", cfg.EffectiveLoginType())

	var nilCfg *Config
	assert.Equal(t, "login", nilCfg.EffectiveLoginType())
}

func TestTokenStyle_IsValid(t *testing.T) {
	valid := []TokenStyle{
		TokenStyleUUID, TokenStyleSimple, TokenStyleRandom32,
		TokenStyleRandom64, TokenStyleRandom128, TokenStyleJWT,
		TokenStyleHash, TokenStyleTimestamp, TokenStyleTik,
	}
	for _, s := range valid {
		assert.True(t, s.IsValid(), "expected %s to be valid", s)
	}
	assert.False(t, TokenStyle("invalid").IsValid())
}

func TestConfig_SetterMethods(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SetTokenName("test").
		SetTimeout(100).
		SetMaxRefresh(50).
		SetIsConcurrent(false).
		SetIsShare(false).
		SetMaxLoginCount(5)

	assert.Equal(t, "test", cfg.TokenName)
	assert.Equal(t, int64(100), cfg.Timeout)
	assert.Equal(t, int64(50), cfg.MaxRefresh)
	assert.False(t, cfg.IsConcurrent)
	assert.False(t, cfg.IsShare)
	assert.Equal(t, 5, cfg.MaxLoginCount)
}
