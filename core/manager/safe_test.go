package manager

import (
	"errors"
	"testing"

	"github.com/sa-tokens/sa-token-go/core/config"
	"github.com/sa-tokens/sa-token-go/core/errs"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenSafe_CheckSafe_CloseSafe(t *testing.T) {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.IsConcurrent = true
	cfg.IsShare = true
	cfg.MaxLoginCount = -1
	m := NewManager(st, cfg)

	tok, err := m.Login("u-safe")
	require.NoError(t, err)

	require.NoError(t, m.OpenSafe(tok, "pay", 120))
	assert.True(t, m.IsSafe(tok, "pay"))
	assert.NoError(t, m.CheckSafe(tok, "pay"))

	require.NoError(t, m.CloseSafe(tok, "pay"))
	err = m.CheckSafe(tok, "pay")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errs.ErrNotPassedSafeAuth))
}
