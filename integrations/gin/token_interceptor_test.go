package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sa-tokens/sa-token-go/core"
	"github.com/sa-tokens/sa-token-go/core/config"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/stretchr/testify/assert"
)

// TestTokenInterceptor_QueryToken Query 传参 apikey 场景：拦截器应写入 satoken_token 并可被 GetTokenFromCtx 读出
func TestTokenInterceptor_QueryToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.TokenName = "satoken"
	cfg.IsReadHeader = false
	cfg.IsReadCookie = false
	m := core.NewManager(st, cfg)
	p := NewPlugin(m)

	r := gin.New()
	r.Use(p.TokenInterceptor())
	r.GET("/t", func(c *gin.Context) {
		c.String(http.StatusOK, GetTokenFromCtx(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/t?satoken=qtoken-value", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "qtoken-value", w.Body.String())
}

// TestTokenInterceptor_NoToken 无 token 时拦截器写入空串，便于下游统一判空
func TestTokenInterceptor_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	m := core.NewManager(st, cfg)
	p := NewPlugin(m)

	r := gin.New()
	r.Use(p.TokenInterceptor())
	r.GET("/t", func(c *gin.Context) {
		c.String(http.StatusOK, GetTokenFromCtx(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "", w.Body.String())
}
