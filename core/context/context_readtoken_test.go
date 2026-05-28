package context

import (
	"testing"

	"github.com/sa-tokens/sa-token-go/core/adapter"
	"github.com/sa-tokens/sa-token-go/core/config"
	"github.com/sa-tokens/sa-token-go/core/manager"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testRequestCtx 测试用 RequestContext 桩，仅填充 Header/Cookie/Query
// testRequestCtx is a minimal RequestContext stub for ReadTokenFromRequest tests
type testRequestCtx struct {
	header map[string]string
	cookie map[string]string
	query  map[string]string
}

func (t *testRequestCtx) GetHeader(key string) string {
	if t.header == nil {
		return ""
	}
	return t.header[key]
}
func (t *testRequestCtx) GetHeaders() map[string][]string { return nil }
func (t *testRequestCtx) GetQuery(key string) string {
	if t.query == nil {
		return ""
	}
	return t.query[key]
}
func (t *testRequestCtx) GetQueryAll() map[string][]string { return nil }
func (t *testRequestCtx) GetPostForm(string) string        { return "" }
func (t *testRequestCtx) GetCookie(key string) string {
	if t.cookie == nil {
		return ""
	}
	return t.cookie[key]
}
func (t *testRequestCtx) GetBody() ([]byte, error) { return nil, nil }
func (t *testRequestCtx) GetClientIP() string      { return "" }
func (t *testRequestCtx) GetMethod() string        { return "GET" }
func (t *testRequestCtx) GetPath() string          { return "/" }
func (t *testRequestCtx) GetURL() string           { return "" }
func (t *testRequestCtx) GetUserAgent() string     { return "" }
func (t *testRequestCtx) SetHeader(_, _ string)    {}
func (t *testRequestCtx) SetCookie(_, _ string, _ int, _, _ string, _, _ bool) {
}
func (t *testRequestCtx) SetCookieWithOptions(_ *adapter.CookieOptions) {}
func (t *testRequestCtx) Set(string, any)                               {}
func (t *testRequestCtx) Get(string) (any, bool)                        { return nil, false }
func (t *testRequestCtx) GetString(string) string                       { return "" }
func (t *testRequestCtx) MustGet(string) any                            { return nil }
func (t *testRequestCtx) Abort()                                        {}
func (t *testRequestCtx) IsAborted() bool                               { return false }

func testManager(cfg *config.Config) *manager.Manager {
	return manager.NewManager(memory.NewStorage(), cfg)
}

func TestResolveTokenName_FallbackAuthorization(t *testing.T) {
	cfg := &config.Config{TokenName: ""}
	assert.Equal(t, AuthHeaderName, ResolveTokenName(cfg))
	assert.Equal(t, "mytok", ResolveTokenName(&config.Config{TokenName: "mytok"}))
}

func TestReadTokenFromRequest_HeaderCookieQuery(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.TokenName = "satoken"
	cfg.IsReadHeader = true
	cfg.IsReadCookie = true
	m := testManager(cfg)

	rc := &testRequestCtx{header: map[string]string{"satoken": "  hval  "}}
	assert.Equal(t, "hval", ReadTokenFromRequest(rc, m))

	rc = &testRequestCtx{cookie: map[string]string{"satoken": "cval"}}
	assert.Equal(t, "cval", ReadTokenFromRequest(rc, m))

	rc = &testRequestCtx{query: map[string]string{"satoken": "qval"}}
	assert.Equal(t, "qval", ReadTokenFromRequest(rc, m))
}

func TestReadTokenFromRequest_EmptyTokenName_AuthorizationBearer(t *testing.T) {
	cfg := config.DefaultConfig()
	// 仅空白字符：ResolveTokenName 回退 Authorization，且仍通过 Validate 的非空校验
	cfg.TokenName = "   "
	cfg.IsReadHeader = true
	cfg.IsReadCookie = true
	require.NoError(t, cfg.Validate())
	m := testManager(cfg)
	rc := &testRequestCtx{header: map[string]string{AuthHeaderName: "Bearer abc123"}}
	assert.Equal(t, "abc123", ReadTokenFromRequest(rc, m))
}

func TestReadTokenFromRequest_IsReadHeaderOff_SkipsHeader(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.TokenName = "satoken"
	cfg.IsReadHeader = false
	cfg.IsReadCookie = true
	m := testManager(cfg)
	rc := &testRequestCtx{
		header: map[string]string{"satoken": "should-not-use"},
		cookie: map[string]string{"satoken": "from-cookie"},
	}
	assert.Equal(t, "from-cookie", ReadTokenFromRequest(rc, m))
}

func TestReadTokenFromRequest_IsReadCookieOff_QueryStillWorks(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.TokenName = "satoken"
	cfg.IsReadHeader = false
	cfg.IsReadCookie = false
	m := testManager(cfg)
	rc := &testRequestCtx{query: map[string]string{"satoken": "from-query"}}
	assert.Equal(t, "from-query", ReadTokenFromRequest(rc, m))
}
