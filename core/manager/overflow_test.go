package manager

import (
	"testing"
	"time"

	"github.com/sa-tokens/sa-token-go/core/config"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func baseOverflowManager(mode string) *Manager {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.IsConcurrent = true
	cfg.IsShare = false
	cfg.MaxLoginCount = 2
	cfg.OverflowLogoutMode = mode
	return NewManager(st, cfg)
}

func TestOverflowLogoutMode_Logout(t *testing.T) {
	m := baseOverflowManager("LOGOUT")
	_, err := m.Login("ov", "d1")
	require.NoError(t, err)
	_, err = m.Login("ov", "d2")
	require.NoError(t, err)
	t3, err := m.Login("ov", "d3")
	require.NoError(t, err)
	list, err := m.GetTokenValueListByLoginID("ov")
	require.NoError(t, err)
	assert.Len(t, list, 2)
	assert.NotEmpty(t, t3)
}

func TestOverflowLogoutMode_Kickout(t *testing.T) {
	m := baseOverflowManager("KICKOUT")
	_, err := m.Login("kv", "d1")
	require.NoError(t, err)
	_, err = m.Login("kv", "d2")
	require.NoError(t, err)
	tokens, err := m.GetTokenValueListByLoginID("kv")
	require.NoError(t, err)
	victim := m.pickOverflowVictimToken(tokens)
	_, err = m.Login("kv", "d3")
	require.NoError(t, err)
	assert.False(t, m.IsLogin(victim))
}

func TestOverflowLogoutMode_Replaced(t *testing.T) {
	m := baseOverflowManager("REPLACED")
	_, err := m.Login("rp", "d1")
	require.NoError(t, err)
	time.Sleep(5 * time.Millisecond)
	_, err = m.Login("rp", "d2")
	require.NoError(t, err)
	tokens, err := m.GetTokenValueListByLoginID("rp")
	require.NoError(t, err)
	require.Len(t, tokens, 2)
	victim := m.pickOverflowVictimToken(tokens)
	_, err = m.Login("rp", "d3")
	require.NoError(t, err)
	raw, err := m.storage.Get(m.getTokenKey(victim))
	require.NoError(t, err)
	str, ok := assertString(raw)
	require.True(t, ok)
	assert.Equal(t, string(TokenStateReplaced), str)
}
