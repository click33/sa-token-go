package oauth2

import (
	"errors"
	"fmt"
	"testing"

	"github.com/sa-tokens/sa-token-go/core/errs"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type staticLoader map[string]*ClientModel

func (s staticLoader) LoadClient(id string) (*ClientModel, error) {
	v, ok := s[id]
	if !ok {
		return nil, fmt.Errorf("missing")
	}
	return v, nil
}

func tpl(t *testing.T) *OAuth2Template {
	t.Helper()
	st := memory.NewStorage()
	prefix := "satoken:"
	loader := staticLoader{
		"c1": {
			ClientID:          "c1",
			ClientSecret:      "secret",
			AllowRedirectURIs: []string{"http://127.0.0.1/cb"},
			ContractScopes:    []string{"openid", "profile"},
			GrantTypes:        []string{"authorization_code", "password", "client_credentials", "refresh_token"},
		},
	}
	return NewOAuth2Template(st, prefix, loader)
}

func TestCheckRedirectURI_OK(t *testing.T) {
	tpl := tpl(t)
	assert.NoError(t, tpl.CheckRedirectURI("c1", "http://127.0.0.1/cb"))
}

func TestCheckRedirectURI_AtSign(t *testing.T) {
	tpl := tpl(t)
	err := tpl.CheckRedirectURI("c1", "http://evil@127.0.0.1/cb")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errs.ErrOAuth2RedirectURIContainsAt))
}

func TestCheckRedirectURI_NotAllowed(t *testing.T) {
	tpl := tpl(t)
	err := tpl.CheckRedirectURI("c1", "http://evil.com/cb")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errs.ErrOAuth2IllegalRedirectURI))
}

func TestCheckContractScope_NotContracted(t *testing.T) {
	tpl := tpl(t)
	err := tpl.CheckContractScope("c1", []string{"admin"})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errs.ErrOAuth2ScopeNotContracted))
}

func TestIsNeedCarefulConfirm_GrantThenOk(t *testing.T) {
	tpl := tpl(t)
	require.NoError(t, tpl.SaveGrantScope("u1", "c1", []string{"openid"}))
	assert.False(t, tpl.IsNeedCarefulConfirm("u1", "c1", []string{"openid"}))
	assert.True(t, tpl.IsNeedCarefulConfirm("u1", "c1", []string{"profile"}))
}
