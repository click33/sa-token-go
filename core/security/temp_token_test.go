package security

import (
	"testing"

	"github.com/sa-tokens/sa-token-go/core/errs"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTempTokenManager_CreateAndVerify(t *testing.T) {
	storage := memory.NewStorage()
	mgr := NewTempTokenManager(storage, "test:")

	info, err := mgr.CreateTempToken("1001", 300, "")
	require.NoError(t, err)
	assert.NotEmpty(t, info.Token)
	assert.Equal(t, "1001", info.LoginID)
	assert.False(t, info.Used)

	// Verify (consumes)
	verified, err := mgr.VerifyTempToken(info.Token)
	require.NoError(t, err)
	assert.Equal(t, "1001", verified.LoginID)
	assert.True(t, verified.Used)
}

func TestTempTokenManager_VerifyTwice_Used(t *testing.T) {
	storage := memory.NewStorage()
	mgr := NewTempTokenManager(storage, "test:")

	info, err := mgr.CreateTempToken("1001", 300, "")
	require.NoError(t, err)

	// First verify succeeds
	_, err = mgr.VerifyTempToken(info.Token)
	require.NoError(t, err)

	// Second verify fails (already used)
	_, err = mgr.VerifyTempToken(info.Token)
	assert.ErrorIs(t, err, errs.ErrTempTokenUsed)
}

func TestTempTokenManager_Verify_NotFound(t *testing.T) {
	storage := memory.NewStorage()
	mgr := NewTempTokenManager(storage, "test:")

	_, err := mgr.VerifyTempToken("nonexistent")
	assert.ErrorIs(t, err, errs.ErrTempTokenNotFound)
}

func TestTempTokenManager_GetTempTokenInfo(t *testing.T) {
	storage := memory.NewStorage()
	mgr := NewTempTokenManager(storage, "test:")

	info, err := mgr.CreateTempToken("1001", 300, "extra-data")
	require.NoError(t, err)

	got, err := mgr.GetTempTokenInfo(info.Token)
	require.NoError(t, err)
	assert.Equal(t, info.Token, got.Token)
	assert.Equal(t, "1001", got.LoginID)
	assert.Equal(t, "extra-data", got.Extra)
}

func TestTempTokenManager_DeleteTempToken(t *testing.T) {
	storage := memory.NewStorage()
	mgr := NewTempTokenManager(storage, "test:")

	info, err := mgr.CreateTempToken("1001", 300, "")
	require.NoError(t, err)

	err = mgr.DeleteTempToken(info.Token)
	require.NoError(t, err)

	_, err = mgr.GetTempTokenInfo(info.Token)
	assert.Error(t, err)
}

func TestTempTokenManager_DefaultExpire(t *testing.T) {
	storage := memory.NewStorage()
	mgr := NewTempTokenManager(storage, "test:")

	// expireSeconds=0 should default to 300s
	info, err := mgr.CreateTempToken("1001", 0, "")
	require.NoError(t, err)
	assert.True(t, info.ExpireTime > info.CreateTime)
	assert.True(t, info.ExpireTime-info.CreateTime == 300)
}

func TestTempTokenManager_WithExtra(t *testing.T) {
	storage := memory.NewStorage()
	mgr := NewTempTokenManager(storage, "test:")

	info, err := mgr.CreateTempToken("1001", 300, `{"action":"reset_password"}`)
	require.NoError(t, err)
	assert.Equal(t, `{"action":"reset_password"}`, info.Extra)
}
