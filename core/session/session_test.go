package session

import (
	"errors"
	"testing"
	"time"

	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountSessionPersistAndLoad(t *testing.T) {
	st := memory.NewStorage()
	prefix := "t:"
	s := NewSession("u1", st, prefix)
	require.NoError(t, s.Set("k", "v"))
	require.NoError(t, s.Persist(time.Hour))

	loaded, err := Load("u1", st, prefix)
	require.NoError(t, err)
	v, ok := loaded.Get("k")
	assert.True(t, ok)
	assert.Equal(t, "v", v)
}

func TestTokenSessionLoadWithKeyPrefix(t *testing.T) {
	st := memory.NewStorage()
	prefix := "t:"
	tok := "tok-abc"
	s := NewTokenSession(tok, st, prefix, "login")
	require.NoError(t, s.Set("scope", "read"))
	require.NoError(t, s.Persist(time.Hour))

	loaded, err := LoadWithKeyPrefix(tok, st, prefix, TokenSessionKeyPrefix)
	require.NoError(t, err)
	assert.Equal(t, TypeToken, loaded.Type)
	v, _ := loaded.Get("scope")
	assert.Equal(t, "read", v)
}

func TestAnonTokenSessionPersist(t *testing.T) {
	st := memory.NewStorage()
	prefix := "t:"
	s := NewAnonTokenSession("anon-1", st, prefix, "login")
	require.NoError(t, s.Set("x", 1))
	require.NoError(t, s.Persist(0))

	loaded, err := LoadWithKeyPrefix("anon-1", st, prefix, TokenSessionKeyPrefix)
	require.NoError(t, err)
	assert.Equal(t, TypeAnon, loaded.Type)
}

func TestLoadMissingSession_ErrSessionNotFound(t *testing.T) {
	st := memory.NewStorage()
	_, err := Load("nope", st, "p:")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrSessionNotFound))
}
