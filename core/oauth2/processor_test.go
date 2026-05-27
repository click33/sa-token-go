package oauth2_test

import (
	"fmt"
	"testing"

	"github.com/sa-tokens/sa-token-go/core/adapter"
	"github.com/sa-tokens/sa-token-go/core/config"
	"github.com/sa-tokens/sa-token-go/core/manager"
	oauth2 "github.com/sa-tokens/sa-token-go/core/oauth2"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubRC minimal RequestContext for processor tests | 最小请求上下文桩
type stubRC struct {
	path   string
	form   map[string]string
	header map[string]string
}

func (s *stubRC) GetHeader(key string) string {
	if s.header == nil {
		return ""
	}
	return s.header[key]
}
func (s *stubRC) GetHeaders() map[string][]string                      { return nil }
func (s *stubRC) GetQuery(string) string                               { return "" }
func (s *stubRC) GetQueryAll() map[string][]string                     { return nil }
func (s *stubRC) GetPostForm(key string) string                        { return s.form[key] }
func (s *stubRC) GetCookie(string) string                              { return "" }
func (s *stubRC) GetBody() ([]byte, error)                             { return nil, nil }
func (s *stubRC) GetClientIP() string                                  { return "" }
func (s *stubRC) GetMethod() string                                    { return "POST" }
func (s *stubRC) GetPath() string                                      { return s.path }
func (s *stubRC) GetURL() string                                       { return "" }
func (s *stubRC) GetUserAgent() string                                 { return "" }
func (s *stubRC) SetHeader(_, _ string)                                {}
func (s *stubRC) SetCookie(_, _ string, _ int, _, _ string, _, _ bool) {}
func (s *stubRC) SetCookieWithOptions(_ *adapter.CookieOptions)        {}
func (s *stubRC) Set(string, any)                                      {}
func (s *stubRC) Get(string) (any, bool)                               { return nil, false }
func (s *stubRC) GetString(string) string                              { return "" }
func (s *stubRC) MustGet(string) any                                   { return nil }
func (s *stubRC) Abort()                                               {}
func (s *stubRC) IsAborted() bool                                      { return false }

type staticLoader map[string]*oauth2.ClientModel

func (s staticLoader) LoadClient(id string) (*oauth2.ClientModel, error) {
	v, ok := s[id]
	if !ok {
		return nil, fmt.Errorf("missing client")
	}
	return v, nil
}

type testUserAuth struct{}

func (testUserAuth) CheckCredential(username, password string) (string, bool) {
	if username == "alice" && password == "pw" {
		return "alice", true
	}
	return "", false
}

func oauthProcessorFixture(t *testing.T) (*oauth2.OAuth2ServerProcessor, *oauth2.OAuth2Server, *manager.Manager) {
	t.Helper()
	st := memory.NewStorage()
	prefix := "satoken:"
	srv := oauth2.NewOAuth2Server(st, prefix)
	require.NoError(t, srv.RegisterClient(&oauth2.Client{
		ClientID:     "c1",
		ClientSecret: "sec",
		RedirectURIs: []string{"http://127.0.0.1/cb"},
		GrantTypes: []oauth2.GrantType{
			oauth2.GrantTypeAuthorizationCode,
			oauth2.GrantTypePassword,
			oauth2.GrantTypeClientCredentials,
			oauth2.GrantTypeRefreshToken,
		},
	}))

	loader := staticLoader{
		"c1": {
			ClientID:          "c1",
			ClientSecret:      "sec",
			AllowRedirectURIs: []string{"http://127.0.0.1/cb"},
			ContractScopes:    []string{"openid"},
			GrantTypes:        []string{"authorization_code", "password", "client_credentials", "refresh_token"},
		},
	}
	tpl := oauth2.NewOAuth2Template(st, prefix, loader)

	cfg := config.DefaultConfig()
	cfg.IsConcurrent = true
	cfg.IsShare = true
	cfg.MaxLoginCount = -1
	mgr := manager.NewManager(st, cfg)

	paths := oauth2.APIPaths{Token: "/oauth2/token"}
	p := oauth2.NewOAuth2ServerProcessor(tpl, srv, mgr, paths, nil, testUserAuth{})
	return p, srv, mgr
}

func TestProcessorToken_ClientCredentials(t *testing.T) {
	p, _, _ := oauthProcessorFixture(t)
	ctx := &stubRC{
		path: "/oauth2/token",
		form: map[string]string{
			"grant_type":    "client_credentials",
			"client_id":     "c1",
			"client_secret": "sec",
		},
	}
	ok, data, err := p.Dispatch(ctx)
	require.NoError(t, err)
	assert.True(t, ok)
	tok, ok := data.(*oauth2.AccessToken)
	require.True(t, ok)
	assert.NotEmpty(t, tok.Token)
}

func TestProcessorToken_Password(t *testing.T) {
	p, _, _ := oauthProcessorFixture(t)
	ctx := &stubRC{
		path: "/oauth2/token",
		form: map[string]string{
			"grant_type":    "password",
			"client_id":     "c1",
			"client_secret": "sec",
			"username":      "alice",
			"password":      "pw",
		},
	}
	ok, data, err := p.Dispatch(ctx)
	require.NoError(t, err)
	assert.True(t, ok)
	tok, ok := data.(*oauth2.AccessToken)
	require.True(t, ok)
	assert.NotEmpty(t, tok.Token)
}

func TestProcessorToken_AuthorizationCode(t *testing.T) {
	p, srv, _ := oauthProcessorFixture(t)
	ac, err := srv.GenerateAuthorizationCode("c1", "http://127.0.0.1/cb", "user9", []string{"openid"})
	require.NoError(t, err)

	ctx := &stubRC{
		path: "/oauth2/token",
		form: map[string]string{
			"grant_type":    "authorization_code",
			"client_id":     "c1",
			"client_secret": "sec",
			"code":          ac.Code,
			"redirect_uri":  "http://127.0.0.1/cb",
		},
	}
	ok, data, err := p.Dispatch(ctx)
	require.NoError(t, err)
	assert.True(t, ok)
	tok, ok := data.(*oauth2.AccessToken)
	require.True(t, ok)
	assert.NotEmpty(t, tok.Token)
}

func TestProcessorToken_RefreshToken(t *testing.T) {
	p, srv, _ := oauthProcessorFixture(t)
	ac, err := srv.GenerateAuthorizationCode("c1", "http://127.0.0.1/cb", "user8", nil)
	require.NoError(t, err)
	first := &stubRC{
		path: "/oauth2/token",
		form: map[string]string{
			"grant_type":    "authorization_code",
			"client_id":     "c1",
			"client_secret": "sec",
			"code":          ac.Code,
			"redirect_uri":  "http://127.0.0.1/cb",
		},
	}
	_, data, err := p.Dispatch(first)
	require.NoError(t, err)
	at := data.(*oauth2.AccessToken)
	require.NotEmpty(t, at.RefreshToken)

	ctx := &stubRC{
		path: "/oauth2/token",
		form: map[string]string{
			"grant_type":    "refresh_token",
			"client_id":     "c1",
			"client_secret": "sec",
			"refresh_token": at.RefreshToken,
		},
	}
	ok, data2, err := p.Dispatch(ctx)
	require.NoError(t, err)
	assert.True(t, ok)
	at2 := data2.(*oauth2.AccessToken)
	assert.NotEmpty(t, at2.Token)
}
