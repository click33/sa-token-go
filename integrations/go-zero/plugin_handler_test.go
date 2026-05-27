package gozero

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sa-tokens/sa-token-go/core"
	"github.com/sa-tokens/sa-token-go/core/config"
	"github.com/sa-tokens/sa-token-go/storage/memory"
)

func TestLogoutHandler_ByLoginID(t *testing.T) {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.TokenName = "Authorization"
	cfg.IsReadHeader = true
	mgr := core.NewManager(st, cfg)
	p := NewPlugin(mgr)

	token, _ := mgr.Login("user1")
	assert.True(t, mgr.IsLogin(token))

	chain := p.AuthMiddleware()(p.LogoutHandler)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("Authorization", token)
	chain(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, mgr.IsLogin(token))
}

func TestLoginHandler_SetsCookieWhenEnabled(t *testing.T) {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.IsReadCookie = true
	cfg.TokenName = "satoken"
	mgr := core.NewManager(st, cfg)
	p := NewPlugin(mgr)

	body, _ := json.Marshal(map[string]string{"username": "u1", "password": "p"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	p.LoginHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	cookies := w.Result().Cookies()
	assert.NotEmpty(t, cookies)
}
