package security

import (
	"testing"

	"github.com/sa-tokens/sa-token-go/core/errs"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiKeyManager_CreateAndGet(t *testing.T) {
	storage := memory.NewStorage()
	mgr := NewApiKeyManager(storage, "test:")

	info, err := mgr.CreateApiKey("1001", "My App Key", 3600, "")
	require.NoError(t, err)
	assert.NotEmpty(t, info.Key)
	assert.Equal(t, "1001", info.LoginID)
	assert.Equal(t, "My App Key", info.Title)
	assert.False(t, info.Disabled)
	assert.True(t, info.ExpireTime > 0)

	// Retrieve
	got, err := mgr.GetApiKeyInfo(info.Key)
	require.NoError(t, err)
	assert.Equal(t, info.Key, got.Key)
	assert.Equal(t, "1001", got.LoginID)
}

func TestApiKeyManager_VerifyApiKey_Valid(t *testing.T) {
	storage := memory.NewStorage()
	mgr := NewApiKeyManager(storage, "test:")

	info, err := mgr.CreateApiKey("1001", "key", 3600, "")
	require.NoError(t, err)

	verified, err := mgr.VerifyApiKey(info.Key)
	require.NoError(t, err)
	assert.Equal(t, info.Key, verified.Key)
}

func TestApiKeyManager_VerifyApiKey_Disabled(t *testing.T) {
	storage := memory.NewStorage()
	mgr := NewApiKeyManager(storage, "test:")

	info, err := mgr.CreateApiKey("1001", "key", 3600, "")
	require.NoError(t, err)

	err = mgr.DisableApiKey(info.Key)
	require.NoError(t, err)

	_, err = mgr.VerifyApiKey(info.Key)
	assert.ErrorIs(t, err, errs.ErrApiKeyDisabled)
}

func TestApiKeyManager_VerifyApiKey_EnableAfterDisable(t *testing.T) {
	storage := memory.NewStorage()
	mgr := NewApiKeyManager(storage, "test:")

	info, err := mgr.CreateApiKey("1001", "key", 3600, "")
	require.NoError(t, err)

	err = mgr.DisableApiKey(info.Key)
	require.NoError(t, err)

	err = mgr.EnableApiKey(info.Key)
	require.NoError(t, err)

	verified, err := mgr.VerifyApiKey(info.Key)
	require.NoError(t, err)
	assert.Equal(t, info.Key, verified.Key)
}

func TestApiKeyManager_VerifyApiKey_NotFound(t *testing.T) {
	storage := memory.NewStorage()
	mgr := NewApiKeyManager(storage, "test:")

	_, err := mgr.VerifyApiKey("nonexistent")
	assert.ErrorIs(t, err, errs.ErrApiKeyNotFound)
}

func TestApiKeyManager_DeleteApiKey(t *testing.T) {
	storage := memory.NewStorage()
	mgr := NewApiKeyManager(storage, "test:")

	info, err := mgr.CreateApiKey("1001", "key", 3600, "")
	require.NoError(t, err)

	err = mgr.DeleteApiKey(info.Key)
	require.NoError(t, err)

	_, err = mgr.GetApiKeyInfo(info.Key)
	assert.Error(t, err)
}

func TestApiKeyManager_NeverExpire(t *testing.T) {
	storage := memory.NewStorage()
	mgr := NewApiKeyManager(storage, "test:")

	info, err := mgr.CreateApiKey("1001", "permanent", 0, "")
	require.NoError(t, err)
	assert.Equal(t, int64(0), info.ExpireTime)

	verified, err := mgr.VerifyApiKey(info.Key)
	require.NoError(t, err)
	assert.Equal(t, info.Key, verified.Key)
}

func TestApiKeyManager_WithExtra(t *testing.T) {
	storage := memory.NewStorage()
	mgr := NewApiKeyManager(storage, "test:")

	info, err := mgr.CreateApiKey("1001", "key", 3600, `{"scope":"read"}`)
	require.NoError(t, err)
	assert.Equal(t, `{"scope":"read"}`, info.Extra)
}
