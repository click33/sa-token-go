package stputil

import (
	"testing"

	"github.com/sa-tokens/sa-token-go/core/config"
	"github.com/sa-tokens/sa-token-go/core/manager"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestManager() *manager.Manager {
	storage := memory.NewStorage()
	cfg := config.DefaultConfig()
	return manager.NewManager(storage, cfg)
}

func TestMultiStpLogic_RegisterAndGet(t *testing.T) {
	storage := memory.NewStorage()
	multi := NewMultiStpLogic(storage)
	mgr := newTestManager()

	logic := NewStpLogic(mgr)
	multi.Register("user", logic)

	got, ok := multi.Get("user")
	assert.True(t, ok)
	assert.Equal(t, logic, got)
}

func TestMultiStpLogic_Get_NotFound(t *testing.T) {
	storage := memory.NewStorage()
	multi := NewMultiStpLogic(storage)

	_, ok := multi.Get("nonexistent")
	assert.False(t, ok)
}

func TestMultiStpLogic_GetOrCreate(t *testing.T) {
	storage := memory.NewStorage()
	multi := NewMultiStpLogic(storage)
	mgr := newTestManager()

	// First call creates
	logic1 := multi.GetOrCreate("admin", mgr)
	assert.NotNil(t, logic1)

	// Second call returns same instance
	logic2 := multi.GetOrCreate("admin", mgr)
	assert.Equal(t, logic1, logic2)
}

func TestMultiStpLogic_LoginByType(t *testing.T) {
	storage := memory.NewStorage()
	multi := NewMultiStpLogic(storage)
	mgr := newTestManager()
	multi.Register("user", NewStpLogic(mgr))

	token, err := multi.LoginByType("user", "1001")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify login
	assert.True(t, multi.IsLoginByType("user", token))
}

func TestMultiStpLogic_LoginByType_NotRegistered(t *testing.T) {
	storage := memory.NewStorage()
	multi := NewMultiStpLogic(storage)

	_, err := multi.LoginByType("unknown", "1001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not registered")
}

func TestMultiStpLogic_GetLoginIDByType(t *testing.T) {
	storage := memory.NewStorage()
	multi := NewMultiStpLogic(storage)
	mgr := newTestManager()
	multi.Register("user", NewStpLogic(mgr))

	token, err := multi.LoginByType("user", "1001")
	require.NoError(t, err)

	id, err := multi.GetLoginIDByType("user", token)
	require.NoError(t, err)
	assert.Equal(t, "1001", id)
}

func TestMultiStpLogic_LogoutByType(t *testing.T) {
	storage := memory.NewStorage()
	multi := NewMultiStpLogic(storage)
	mgr := newTestManager()
	multi.Register("user", NewStpLogic(mgr))

	token, err := multi.LoginByType("user", "1001")
	require.NoError(t, err)
	assert.True(t, multi.IsLoginByType("user", token))

	err = multi.LogoutByType("user", "1001")
	require.NoError(t, err)
	assert.False(t, multi.IsLoginByType("user", token))
}

func TestMultiStpLogic_IsLoginByType_NotRegistered(t *testing.T) {
	storage := memory.NewStorage()
	multi := NewMultiStpLogic(storage)

	assert.False(t, multi.IsLoginByType("unknown", "token"))
}

func TestMultiStpLogic_GetTokenValueByType(t *testing.T) {
	storage := memory.NewStorage()
	multi := NewMultiStpLogic(storage)
	mgr := newTestManager()
	multi.Register("user", NewStpLogic(mgr))

	token, err := multi.LoginByType("user", "1001")
	require.NoError(t, err)

	got, err := multi.GetTokenValueByType("user", "1001")
	require.NoError(t, err)
	assert.Equal(t, token, got)
}

func TestMultiStpLogic_ListTypes(t *testing.T) {
	storage := memory.NewStorage()
	multi := NewMultiStpLogic(storage)
	mgr := newTestManager()

	multi.Register("user", NewStpLogic(mgr))
	multi.Register("admin", NewStpLogic(mgr))

	types := multi.ListTypes()
	assert.Len(t, types, 2)
	assert.Contains(t, types, "user")
	assert.Contains(t, types, "admin")
}

func TestMultiStpLogic_MultipleTypes(t *testing.T) {
	storage := memory.NewStorage()
	multi := NewMultiStpLogic(storage)

	// Use separate managers so each type has its own token space
	userMgr := manager.NewManager(memory.NewStorage(), config.DefaultConfig())
	adminMgr := manager.NewManager(memory.NewStorage(), config.DefaultConfig())
	multi.Register("user", NewStpLogic(userMgr))
	multi.Register("admin", NewStpLogic(adminMgr))

	// Login as user
	userToken, err := multi.LoginByType("user", "1001")
	require.NoError(t, err)

	// Login as admin
	adminToken, err := multi.LoginByType("admin", "1001")
	require.NoError(t, err)

	// Both should be valid
	assert.True(t, multi.IsLoginByType("user", userToken))
	assert.True(t, multi.IsLoginByType("admin", adminToken))

	// Tokens should be different (separate managers)
	assert.NotEqual(t, userToken, adminToken)
}
