package security

import (
	"testing"
	"time"

	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSameToken() *SameTokenTemplate {
	storage := memory.NewStorage()
	return NewSameTokenTemplate(storage, "test", 10*time.Second)
}

func TestSameToken_GetToken_CreatesOnFirstCall(t *testing.T) {
	st := newTestSameToken()

	token, err := st.GetToken()
	require.NoError(t, err)
	assert.Len(t, token, 64)
}

func TestSameToken_GetToken_ReturnsSameOnSubsequentCalls(t *testing.T) {
	st := newTestSameToken()

	token1, err := st.GetToken()
	require.NoError(t, err)

	token2, err := st.GetToken()
	require.NoError(t, err)

	assert.Equal(t, token1, token2)
}

func TestSameToken_RefreshToken_RotatesToken(t *testing.T) {
	st := newTestSameToken()

	oldToken, err := st.GetToken()
	require.NoError(t, err)

	newToken, err := st.RefreshToken()
	require.NoError(t, err)

	assert.NotEqual(t, oldToken, newToken)
	assert.Len(t, newToken, 64)
}

func TestSameToken_CheckToken_CurrentToken(t *testing.T) {
	st := newTestSameToken()

	token, err := st.GetToken()
	require.NoError(t, err)

	// Current token should be valid
	assert.NoError(t, st.CheckToken(token))
}

func TestSameToken_CheckToken_PastToken(t *testing.T) {
	st := newTestSameToken()

	oldToken, err := st.GetToken()
	require.NoError(t, err)

	// Refresh moves old token to past
	_, err = st.RefreshToken()
	require.NoError(t, err)

	// Past token should still be valid (grace window)
	assert.NoError(t, st.CheckToken(oldToken))
}

func TestSameToken_CheckToken_InvalidToken(t *testing.T) {
	st := newTestSameToken()

	_, err := st.GetToken()
	require.NoError(t, err)

	assert.Error(t, st.CheckToken("invalid-token"))
}

func TestSameToken_CheckToken_EmptyToken(t *testing.T) {
	st := newTestSameToken()

	assert.Error(t, st.CheckToken(""))
}

func TestSameToken_IsValid(t *testing.T) {
	st := newTestSameToken()

	token, err := st.GetToken()
	require.NoError(t, err)

	assert.True(t, st.IsValid(token))
	assert.False(t, st.IsValid("bad"))
	assert.False(t, st.IsValid(""))
}

func TestSameToken_PastToken_OverwrittenOnSecondRefresh(t *testing.T) {
	st := newTestSameToken()

	token1, err := st.GetToken()
	require.NoError(t, err)

	// First refresh: token1 → past
	_, err = st.RefreshToken()
	require.NoError(t, err)

	// Second refresh: token2 → past, token1 no longer valid
	_, err = st.RefreshToken()
	require.NoError(t, err)

	assert.Error(t, st.CheckToken(token1))
}

func TestSameToken_Constants(t *testing.T) {
	assert.Equal(t, "SA-SAME-TOKEN", SameTokenHeader)
}
