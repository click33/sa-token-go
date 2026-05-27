package gozero

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sa-tokens/sa-token-go/core"
	"github.com/sa-tokens/sa-token-go/core/config"
	"github.com/sa-tokens/sa-token-go/storage/memory"
)

func TestTokenInterceptor_HeaderToken(t *testing.T) {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.TokenName = "Authorization"
	cfg.IsReadHeader = true
	mgr := core.NewManager(st, cfg)
	p := NewPlugin(mgr)

	handler := p.TokenInterceptor()(func(w http.ResponseWriter, r *http.Request) {
		tok := GetTokenFromCtx(r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(tok))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "test-token-value")
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test-token-value", w.Body.String())
}

func TestTokenInterceptor_NoToken(t *testing.T) {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.TokenName = "Authorization"
	cfg.IsReadHeader = true
	mgr := core.NewManager(st, cfg)
	p := NewPlugin(mgr)

	handler := p.TokenInterceptor()(func(w http.ResponseWriter, r *http.Request) {
		tok := GetTokenFromCtx(r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(tok))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestTokenInterceptor_QueryToken(t *testing.T) {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.TokenName = "satoken"
	cfg.IsReadHeader = false
	cfg.IsReadBody = false
	cfg.IsReadCookie = false
	mgr := core.NewManager(st, cfg)
	p := NewPlugin(mgr)

	handler := p.TokenInterceptor()(func(w http.ResponseWriter, r *http.Request) {
		tok := GetTokenFromCtx(r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(tok))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?satoken=query-token", nil)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "query-token", w.Body.String())
}

func TestGetTokenFromCtx_NilRequest(t *testing.T) {
	assert.Empty(t, GetTokenFromCtx(nil))
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.TokenName = "Authorization"
	cfg.IsReadHeader = true
	mgr := core.NewManager(st, cfg)
	p := NewPlugin(mgr)

	token, _ := mgr.Login("user1")

	handler := p.AuthMiddleware()(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", token)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.TokenName = "Authorization"
	cfg.IsReadHeader = true
	mgr := core.NewManager(st, cfg)
	p := NewPlugin(mgr)

	handler := p.AuthMiddleware()(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	handler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPermissionRequired_Valid(t *testing.T) {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.TokenName = "Authorization"
	cfg.IsReadHeader = true
	mgr := core.NewManager(st, cfg)
	p := NewPlugin(mgr)

	token, _ := mgr.Login("user1")
	mgr.SetPermissions("user1", []string{"admin:write"})

	handler := p.PermissionRequired("admin:write")(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", token)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPermissionRequired_Denied(t *testing.T) {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.TokenName = "Authorization"
	cfg.IsReadHeader = true
	mgr := core.NewManager(st, cfg)
	p := NewPlugin(mgr)

	token, _ := mgr.Login("user1")

	handler := p.PermissionRequired("admin:write")(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", token)
	handler(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRoleRequired_Valid(t *testing.T) {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.TokenName = "Authorization"
	cfg.IsReadHeader = true
	mgr := core.NewManager(st, cfg)
	p := NewPlugin(mgr)

	token, _ := mgr.Login("user1")
	mgr.SetRoles("user1", []string{"Admin"})

	handler := p.RoleRequired("Admin")(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", token)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleRequired_Denied(t *testing.T) {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.TokenName = "Authorization"
	cfg.IsReadHeader = true
	mgr := core.NewManager(st, cfg)
	p := NewPlugin(mgr)

	token, _ := mgr.Login("user1")

	handler := p.RoleRequired("Admin")(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", token)
	handler(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetSaToken_WithValidContext(t *testing.T) {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.TokenName = "Authorization"
	cfg.IsReadHeader = true
	mgr := core.NewManager(st, cfg)
	p := NewPlugin(mgr)

	token, _ := mgr.Login("user1")

	var capturedCtx *core.SaTokenContext
	handler := p.AuthMiddleware()(func(w http.ResponseWriter, r *http.Request) {
		saCtx, ok := GetSaToken(r)
		assert.True(t, ok)
		assert.NotNil(t, saCtx)
		capturedCtx = saCtx
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", token)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotNil(t, capturedCtx)
}

func TestGetSaToken_WithoutAuth(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test", nil)
	saCtx, ok := GetSaToken(req)
	assert.False(t, ok)
	assert.Nil(t, saCtx)
}
