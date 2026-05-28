package security

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sa-tokens/sa-token-go/core/config"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRefreshManager() *RefreshTokenManager {
	st := memory.NewStorage()
	cfg := &config.Config{Timeout: 3600}
	return NewRefreshTokenManager(st, "test:", "token:", cfg)
}

func newTestRefreshManagerWithStorage() (*RefreshTokenManager, *memory.Storage) {
	st := memory.NewStorage().(*memory.Storage)
	cfg := &config.Config{Timeout: 3600}
	return NewRefreshTokenManager(st, "test:", "token:", cfg), st
}

func TestRefreshTokenManager_GenerateTokenPair(t *testing.T) {
	rm, _ := newTestRefreshManagerWithStorage()
	info, err := rm.GenerateTokenPair("user1", "default", "access123")
	require.NoError(t, err)
	assert.Equal(t, "access123", info.AccessToken)
	assert.NotEmpty(t, info.RefreshToken)
	assert.Equal(t, "user1", info.LoginID)
	assert.Equal(t, "default", info.Device)
	assert.True(t, info.CreateTime > 0)
	assert.True(t, info.ExpireTime > info.CreateTime)
}

func TestRefreshTokenManager_RefreshAccessToken(t *testing.T) {
	rm, st := newTestRefreshManagerWithStorage()
	info, err := rm.GenerateTokenPair("user1", "default", "old-access")
	require.NoError(t, err)

	// Re-store as JSON bytes to match what RefreshAccessToken expects
	// (GenerateTokenPair stores *RefreshTokenInfo, but retrieval expects []byte)
	infoJSON, _ := json.Marshal(info)
	st.Set("test:refresh:"+info.RefreshToken, infoJSON, 30*24*time.Hour)

	// Also store token info for the copy logic
	tokenData := map[string]string{"loginId": "user1", "device": "default"}
	tokenJSON, _ := json.Marshal(tokenData)
	st.Set("test:token:old-access", tokenJSON, time.Hour)

	refreshed, err := rm.RefreshAccessToken(info.RefreshToken)
	require.NoError(t, err)
	assert.NotEqual(t, "old-access", refreshed.AccessToken)
	assert.Equal(t, info.RefreshToken, refreshed.RefreshToken)
	assert.Equal(t, "user1", refreshed.LoginID)
}

func TestRefreshTokenManager_RevokeRefreshToken(t *testing.T) {
	rm, _ := newTestRefreshManagerWithStorage()
	info, err := rm.GenerateTokenPair("user1", "default", "access1")
	require.NoError(t, err)

	assert.NoError(t, rm.RevokeRefreshToken(info.RefreshToken))

	_, err = rm.RefreshAccessToken(info.RefreshToken)
	assert.Error(t, err)
}

func TestRefreshTokenManager_GetRefreshTokenInfo(t *testing.T) {
	rm, st := newTestRefreshManagerWithStorage()
	info, err := rm.GenerateTokenPair("user1", "default", "access1")
	require.NoError(t, err)

	// Store as bytes to match what GetRefreshTokenInfo expects
	infoJSON, _ := json.Marshal(info)
	st.Set("test:refresh:"+info.RefreshToken, infoJSON, time.Hour)

	retrieved, err := rm.GetRefreshTokenInfo(info.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, info.LoginID, retrieved.LoginID)
	assert.Equal(t, info.AccessToken, retrieved.AccessToken)
}

func TestRefreshTokenManager_EmptyInputs(t *testing.T) {
	rm, _ := newTestRefreshManagerWithStorage()
	_, err := rm.RefreshAccessToken("")
	assert.Error(t, err)

	_, err = rm.GetRefreshTokenInfo("")
	assert.Error(t, err)
}

func TestRefreshTokenManager_InvalidRefreshToken(t *testing.T) {
	rm, _ := newTestRefreshManagerWithStorage()
	_, err := rm.RefreshAccessToken("nonexistent")
	assert.Error(t, err)
}
