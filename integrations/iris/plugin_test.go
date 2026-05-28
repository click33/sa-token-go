package iris

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	irisfw "github.com/kataras/iris/v12"
	"github.com/sa-tokens/sa-token-go/core"
	"github.com/sa-tokens/sa-token-go/core/config"
	"github.com/sa-tokens/sa-token-go/core/manager"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/sa-tokens/sa-token-go/stputil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestPlugin builds an isolated Plugin + Manager and points stputil's global Manager
// to the same instance so decorators (which read the global Manager) and Plugin
// middlewares share the same storage state.
// newTestPlugin 构造一份独立的 Plugin 与 Manager，并将 stputil 全局 Manager 指向同一实例，
// 使装饰器（依赖全局 Manager）与 Plugin 中间件共享同一存储状态。
func newTestPlugin(t *testing.T) (*Plugin, *core.Manager) {
	t.Helper()
	st := memory.NewStorage()
	cfg := &config.Config{
		TokenName:    "satoken",
		Timeout:      2592000,
		IsConcurrent: true,
		IsShare:      true,
		IsReadHeader: true,
	}
	mgr := manager.NewManager(st, cfg)
	stputil.SetManager(mgr)
	p := NewPlugin(mgr)
	return p, mgr
}

// buildApp constructs an Iris app with logging disabled to keep test output clean.
// buildApp 构造禁用日志的 Iris 应用，避免单测刷屏。
func buildApp(t *testing.T) *irisfw.Application {
	t.Helper()
	app := irisfw.New()
	app.Logger().SetLevel("disable")
	return app
}

// serve performs a single request and returns the recorder. | serve 执行一次请求并返回响应记录。
func serve(t *testing.T, app *irisfw.Application, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	require.NoError(t, app.Build())
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	return w
}

// ===== AuthMiddleware =====

// TestAuthMiddleware_NoToken: request without token should be rejected with 401.
// TestAuthMiddleware_NoToken：无 token 的请求应被 401 拒绝。
func TestAuthMiddleware_NoToken(t *testing.T) {
	p, _ := newTestPlugin(t)
	app := buildApp(t)
	app.Get("/api", p.AuthMiddleware(), func(c irisfw.Context) { _ = c.JSON(map[string]string{"ok": "1"}) })

	w := serve(t, app, httptest.NewRequest(http.MethodGet, "/api", nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAuthMiddleware_OK: valid token should pass and SaTokenContext should be exposed.
// TestAuthMiddleware_OK：合法 token 应通过校验，并能在 handler 中取到 SaTokenContext。
func TestAuthMiddleware_OK(t *testing.T) {
	p, mgr := newTestPlugin(t)
	app := buildApp(t)
	app.Get("/api", p.AuthMiddleware(), func(c irisfw.Context) {
		ctx, ok := GetSaToken(c)
		assert.True(t, ok)
		assert.NotNil(t, ctx)
		_ = c.JSON(map[string]string{"ok": "1"})
	})

	token, err := mgr.Login("u-1", "default")
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Authorization", token)

	w := serve(t, app, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ===== PathAuthMiddleware =====

// TestPathAuthMiddleware_RejectByInclude: Include matches and no token -> 401.
// TestPathAuthMiddleware_RejectByInclude：命中 Include 且无 token，应返回 401。
func TestPathAuthMiddleware_RejectByInclude(t *testing.T) {
	p, _ := newTestPlugin(t)
	app := buildApp(t)

	cfg := &core.PathAuthConfig{
		Include: []string{"/admin/**"},
	}
	app.Get("/admin/secret", p.PathAuthMiddleware(cfg), func(c irisfw.Context) {
		_ = c.JSON(map[string]string{"ok": "1"})
	})

	w := serve(t, app, httptest.NewRequest(http.MethodGet, "/admin/secret", nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestPathAuthMiddleware_AllowWhenLoggedIn: Include matches and token valid -> 200.
// TestPathAuthMiddleware_AllowWhenLoggedIn：命中 Include 且 token 合法，应放行。
func TestPathAuthMiddleware_AllowWhenLoggedIn(t *testing.T) {
	p, mgr := newTestPlugin(t)
	app := buildApp(t)

	cfg := &core.PathAuthConfig{
		Include: []string{"/admin/**"},
	}
	app.Get("/admin/secret", p.PathAuthMiddleware(cfg), func(c irisfw.Context) {
		_ = c.JSON(map[string]string{"ok": "1"})
	})

	token, err := mgr.Login("u-2", "default")
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/admin/secret", nil)
	req.Header.Set("Authorization", token)

	w := serve(t, app, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestPathAuthMiddleware_ExcludeBypass: path hits Exclude -> always allowed even w/o token.
// TestPathAuthMiddleware_ExcludeBypass：路径命中 Exclude，即使无 token 也应放行。
func TestPathAuthMiddleware_ExcludeBypass(t *testing.T) {
	p, _ := newTestPlugin(t)
	app := buildApp(t)

	cfg := &core.PathAuthConfig{
		Include: []string{"/admin/**"},
		Exclude: []string{"/admin/public/**"},
	}
	app.Get("/admin/public/info", p.PathAuthMiddleware(cfg), func(c irisfw.Context) {
		_ = c.JSON(map[string]string{"ok": "1"})
	})

	w := serve(t, app, httptest.NewRequest(http.MethodGet, "/admin/public/info", nil))
	assert.Equal(t, http.StatusOK, w.Code)
}

// ===== PermissionRequired / RoleRequired =====

// TestPermissionRequired_Denied: logged-in user lacking permission -> 403.
// TestPermissionRequired_Denied：已登录但缺权限，应返回 403。
func TestPermissionRequired_Denied(t *testing.T) {
	p, mgr := newTestPlugin(t)
	app := buildApp(t)
	app.Get("/p", p.PermissionRequired("user:delete"), func(c irisfw.Context) {
		_ = c.JSON(map[string]string{"ok": "1"})
	})

	token, err := mgr.Login("u-3", "default")
	require.NoError(t, err)
	require.NoError(t, stputil.SetPermissions("u-3", []string{"user:read"}))

	req := httptest.NewRequest(http.MethodGet, "/p", nil)
	req.Header.Set("Authorization", token)
	w := serve(t, app, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestRoleRequired_OK: logged-in user holding role -> 200.
// TestRoleRequired_OK：已登录且具备所需角色，应放行。
func TestRoleRequired_OK(t *testing.T) {
	p, mgr := newTestPlugin(t)
	app := buildApp(t)
	app.Get("/r", p.RoleRequired("admin"), func(c irisfw.Context) {
		_ = c.JSON(map[string]string{"ok": "1"})
	})

	token, err := mgr.Login("u-4", "default")
	require.NoError(t, err)
	require.NoError(t, stputil.SetRoles("u-4", []string{"admin"}))

	req := httptest.NewRequest(http.MethodGet, "/r", nil)
	req.Header.Set("Authorization", token)
	w := serve(t, app, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ===== LoginHandler =====

// TestLoginHandler_RejectEmptyUsername: empty username -> 400.
// TestLoginHandler_RejectEmptyUsername：用户名为空应返回 400。
func TestLoginHandler_RejectEmptyUsername(t *testing.T) {
	p, _ := newTestPlugin(t)
	app := buildApp(t)
	app.Post("/login", p.LoginHandler)

	body := bytes.NewBufferString(`{"username":"","password":"x"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := serve(t, app, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestLoginHandler_Success: non-empty username -> 200 with token in body.
// TestLoginHandler_Success：用户名非空应返回 200，且响应体包含 token 字段。
func TestLoginHandler_Success(t *testing.T) {
	p, _ := newTestPlugin(t)
	app := buildApp(t)
	app.Post("/login", p.LoginHandler)

	body := bytes.NewBufferString(`{"username":"u-5","password":"x"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := serve(t, app, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"token"`)
}

// ===== GetClientIP =====

// TestGetClientIP_PrefersXFF: X-Forwarded-For should win over RemoteAddr.
// TestGetClientIP_PrefersXFF：X-Forwarded-For 应优先于 RemoteAddr。
func TestGetClientIP_PrefersXFF(t *testing.T) {
	app := buildApp(t)
	app.Get("/ip", func(c irisfw.Context) {
		_, _ = c.WriteString(NewIrisContext(c).GetClientIP())
	})
	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 10.0.0.1")
	req.RemoteAddr = "127.0.0.1:55555"
	w := serve(t, app, req)
	assert.Equal(t, "1.2.3.4", w.Body.String())
}

// TestGetClientIP_FallbackRemoteAddr: no proxy headers -> RemoteAddr with port stripped.
// TestGetClientIP_FallbackRemoteAddr：无代理头时应返回 RemoteAddr 去端口后的纯 IP。
func TestGetClientIP_FallbackRemoteAddr(t *testing.T) {
	app := buildApp(t)
	app.Get("/ip", func(c irisfw.Context) {
		_, _ = c.WriteString(NewIrisContext(c).GetClientIP())
	})
	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "9.9.9.9:1234"
	w := serve(t, app, req)
	assert.Equal(t, "9.9.9.9", w.Body.String())
}

// ===== getHTTPStatusFromCode =====

// TestHTTPStatusMapping: business codes should map to expected HTTP status codes.
// TestHTTPStatusMapping：业务码到 HTTP 状态码的映射应符合预期。
func TestHTTPStatusMapping(t *testing.T) {
	cases := map[int]int{
		core.CodeNotLogin:         http.StatusUnauthorized,
		core.CodeTokenInvalid:     http.StatusUnauthorized,
		core.CodeAccountDisabled:  http.StatusForbidden,
		core.CodePermissionDenied: http.StatusForbidden,
		core.CodeBadRequest:       http.StatusBadRequest,
		core.CodeNotFound:         http.StatusNotFound,
		core.CodeServerError:      http.StatusInternalServerError,
	}
	for code, want := range cases {
		assert.Equal(t, want, getHTTPStatusFromCode(code), "code=%d", code)
	}
}
