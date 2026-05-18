package gozero

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/click33/sa-token-go/core"
	"github.com/click33/sa-token-go/core/config"
	"github.com/click33/sa-token-go/storage/memory"
)

func TestPathAuthMiddleware_PropagatesContext(t *testing.T) {
	st := memory.NewStorage()
	cfg := config.DefaultConfig()
	cfg.TokenName = "Authorization"
	cfg.IsReadHeader = true
	mgr := core.NewManager(st, cfg)
	p := NewPlugin(mgr)

	token, _ := mgr.Login("user1")
	pathCfg := core.NewPathAuthConfig().SetInclude([]string{"/secure"})

	var gotLoginID string
	var gotSa bool
	handler := p.PathAuthMiddleware(pathCfg)(func(w http.ResponseWriter, r *http.Request) {
		id, ok := GetLoginIDFromCtx(r)
		assert.True(t, ok)
		gotLoginID = id
		_, gotSa = GetSaToken(r)
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/secure", nil)
	req.Header.Set("Authorization", token)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "user1", gotLoginID)
	assert.True(t, gotSa)
}
