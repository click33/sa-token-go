package iris

import (
	"net/http"
	"net/http/httptest"
	"testing"

	irisfw "github.com/kataras/iris/v12"
	"github.com/sa-tokens/sa-token-go/core"
	"github.com/sa-tokens/sa-token-go/core/config"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTokenInterceptor_QueryToken Query 传参 apikey 场景：拦截器应写入 satoken_token 并可被 GetTokenFromCtx 读出
func TestTokenInterceptor_QueryToken(t *testing.T) {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.TokenName = "satoken"
	cfg.IsReadHeader = false
	cfg.IsReadCookie = false
	m := core.NewManager(st, cfg)
	p := NewPlugin(m)

	app := irisfw.New()
	app.Logger().SetLevel("disable")
	app.Use(p.TokenInterceptor())
	app.Get("/t", func(c irisfw.Context) {
		_, _ = c.WriteString(GetTokenFromCtx(c))
	})
	require.NoError(t, app.Build())

	req := httptest.NewRequest(http.MethodGet, "/t?satoken=qtoken-value", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "qtoken-value", w.Body.String())
}

// TestTokenInterceptor_NoToken 无 token 时拦截器写入空串，便于下游统一判空
func TestTokenInterceptor_NoToken(t *testing.T) {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	m := core.NewManager(st, cfg)
	p := NewPlugin(m)

	app := irisfw.New()
	app.Logger().SetLevel("disable")
	app.Use(p.TokenInterceptor())
	app.Get("/t", func(c irisfw.Context) {
		_, _ = c.WriteString(GetTokenFromCtx(c))
	})
	require.NoError(t, app.Build())

	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "", w.Body.String())
}
