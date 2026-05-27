package stputil

import (
	"context"
	"errors"
	"testing"

	core "github.com/sa-tokens/sa-token-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetTokenValueToCtx_GetLoginIDFromCtx(t *testing.T) {
	setupTestManager()
	tok, err := Login("ctx-user")
	require.NoError(t, err)

	ctx := context.Background()
	ctx = SetTokenValueToCtx(ctx, tok)
	id, err := GetLoginIDFromCtx(ctx)
	require.NoError(t, err)
	assert.Equal(t, "ctx-user", id)
}

func TestGetLoginIDFromCtx_SwitchToPriority(t *testing.T) {
	setupTestManager()
	ctx := context.Background()
	ctx = SetTokenValueToCtx(ctx, "ignored")
	ctx = SwitchTo(ctx, "ghost")
	id, err := GetLoginIDFromCtx(ctx)
	require.NoError(t, err)
	assert.Equal(t, "ghost", id)
}

func TestGetLoginIDFromCtx_NoToken_ErrNotLogin(t *testing.T) {
	setupTestManager()
	_, err := GetLoginIDFromCtx(context.Background())
	assert.Error(t, err)
	assert.True(t, errors.Is(err, core.ErrNotLogin))
}
