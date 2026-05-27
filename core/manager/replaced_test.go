package manager

import (
	"testing"

	"github.com/sa-tokens/sa-token-go/core/config"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplacedRange_CurrDeviceOnly(t *testing.T) {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.IsConcurrent = true
	cfg.IsShare = false
	cfg.MaxLoginCount = -1
	cfg.ReplacedRange = "CURR_DEVICE"
	m := NewManager(st, cfg)

	tPC, err := m.Login("u1", "pc")
	require.NoError(t, err)
	tMobile, err := m.Login("u1", "mobile")
	require.NoError(t, err)

	require.NoError(t, m.Replaced("u1", "pc"))
	assert.False(t, m.IsLogin(tPC))
	assert.True(t, m.IsLogin(tMobile))
}

func TestReplacedRange_AllDevices(t *testing.T) {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.IsConcurrent = true
	cfg.IsShare = false
	cfg.MaxLoginCount = -1
	cfg.ReplacedRange = "ALL_DEVICE"
	m := NewManager(st, cfg)

	t1, err := m.Login("u2", "a")
	require.NoError(t, err)
	t2, err := m.Login("u2", "b")
	require.NoError(t, err)

	require.NoError(t, m.Replaced("u2", "a"))
	assert.False(t, m.IsLogin(t1))
	assert.False(t, m.IsLogin(t2))
}
