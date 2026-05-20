package iris

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/click33/sa-token-go/core/config"
	"github.com/click33/sa-token-go/core/manager"
	"github.com/click33/sa-token-go/storage/memory"
	"github.com/click33/sa-token-go/stputil"
	irisfw "github.com/kataras/iris/v12"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRouter 创建测试应用和初始化 sa-token | create test app and init sa-token
func setupTestRouter() *irisfw.Application {
	app := irisfw.New()
	app.Logger().SetLevel("disable")

	storage := memory.NewStorage()
	cfg := &config.Config{
		TokenName:     "satoken",
		Timeout:       2592000,
		IsConcurrent:  true,
		IsShare:       true,
		MaxLoginCount: -1,
		IsReadHeader:  true,
		IsReadCookie:  false,
	}
	mgr := manager.NewManager(storage, cfg)
	stputil.SetManager(mgr)

	return app
}

func mockLogin(loginID interface{}) string {
	token, _ := stputil.Login(loginID)
	return token
}

func mockLoginWithRole(loginID interface{}, roles []string) string {
	token, _ := stputil.Login(loginID)
	stputil.SetRoles(loginID, roles)
	return token
}

func mockLoginWithPermission(loginID interface{}, permissions []string) string {
	token, _ := stputil.Login(loginID)
	stputil.SetPermissions(loginID, permissions)
	return token
}

// serveOnce builds the iris app once and serves a single request.
func serveOnce(t *testing.T, app *irisfw.Application, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	require.NoError(t, app.Build())
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	return w
}

// TestCheckRole_WithValidRole 测试具有有效角色的用户访问
func TestCheckRole_WithValidRole(t *testing.T) {
	app := setupTestRouter()

	app.Get("/admin", CheckRole("Admin"), func(c irisfw.Context) {
		_ = c.JSON(map[string]interface{}{"message": "success"})
	})

	token := mockLoginWithRole("user123", []string{"Admin"})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", token)

	w := serveOnce(t, app, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

// TestCheckRole_WithInvalidRole 测试没有所需角色的用户访问
func TestCheckRole_WithInvalidRole(t *testing.T) {
	app := setupTestRouter()

	app.Get("/admin", CheckRole("Admin"), func(c irisfw.Context) {
		_ = c.JSON(map[string]interface{}{"message": "success"})
	})

	token := mockLoginWithRole("user456", []string{"User"})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", token)

	w := serveOnce(t, app, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "role denied")
}

// TestCheckRole_MultipleRoles 测试多个角色的情况（OR 逻辑）
func TestCheckRole_MultipleRoles(t *testing.T) {
	app := setupTestRouter()

	app.Get("/manage", CheckRole("Admin", "SuperAdmin"), func(c irisfw.Context) {
		_ = c.JSON(map[string]interface{}{"message": "success"})
	})

	token := mockLoginWithRole("superuser", []string{"SuperAdmin"})

	req := httptest.NewRequest(http.MethodGet, "/manage", nil)
	req.Header.Set("Authorization", token)

	w := serveOnce(t, app, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

// TestCheckRole_NoToken 测试未提供 token 的情况
func TestCheckRole_NoToken(t *testing.T) {
	app := setupTestRouter()

	app.Get("/admin", CheckRole("Admin"), func(c irisfw.Context) {
		_ = c.JSON(map[string]interface{}{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := serveOnce(t, app, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "not logged")
}

// TestCheckRole_InvalidToken 测试无效 token 的情况
func TestCheckRole_InvalidToken(t *testing.T) {
	app := setupTestRouter()

	app.Get("/admin", CheckRole("Admin"), func(c irisfw.Context) {
		_ = c.JSON(map[string]interface{}{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "invalid-token-12345")

	w := serveOnce(t, app, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "not logged")
}

// TestCheckPermission_WithValidPermission 测试具有有效权限的用户访问
func TestCheckPermission_WithValidPermission(t *testing.T) {
	app := setupTestRouter()

	app.Get("/users", CheckPermission("user.read"), func(c irisfw.Context) {
		_ = c.JSON(map[string]interface{}{"message": "success"})
	})

	token := mockLoginWithPermission("user789", []string{"user.read"})

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set("Authorization", token)

	w := serveOnce(t, app, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

// TestCheckPermission_WithInvalidPermission 测试没有所需权限的用户访问
func TestCheckPermission_WithInvalidPermission(t *testing.T) {
	app := setupTestRouter()

	app.Get("/users", CheckPermission("user.delete"), func(c irisfw.Context) {
		_ = c.JSON(map[string]interface{}{"message": "success"})
	})

	token := mockLoginWithPermission("user789", []string{"user.read"})

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set("Authorization", token)

	w := serveOnce(t, app, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "permission denied")
}

// TestCheckLogin_Success 测试登录检查成功
func TestCheckLogin_Success(t *testing.T) {
	app := setupTestRouter()

	app.Get("/profile", CheckLogin(), func(c irisfw.Context) {
		_ = c.JSON(map[string]interface{}{"message": "profile data"})
	})

	token := mockLogin("user999")

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req.Header.Set("Authorization", token)

	w := serveOnce(t, app, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "profile data")
}

// TestCheckLogin_Failed 测试登录检查失败
func TestCheckLogin_Failed(t *testing.T) {
	app := setupTestRouter()

	app.Get("/profile", CheckLogin(), func(c irisfw.Context) {
		_ = c.JSON(map[string]interface{}{"message": "profile data"})
	})

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	w := serveOnce(t, app, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "not logged")
}

// TestCheckDisable_NotDisabled 测试账号未被封禁的情况
func TestCheckDisable_NotDisabled(t *testing.T) {
	app := setupTestRouter()

	app.Get("/resource", CheckDisable(), func(c irisfw.Context) {
		_ = c.JSON(map[string]interface{}{"message": "resource data"})
	})

	token := mockLogin("user101")

	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	req.Header.Set("Authorization", token)

	w := serveOnce(t, app, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "resource data")
}

// TestCheckDisable_IsDisabled 测试账号被封禁的情况
func TestCheckDisable_IsDisabled(t *testing.T) {
	app := setupTestRouter()

	app.Get("/resource", CheckDisable(), func(c irisfw.Context) {
		_ = c.JSON(map[string]interface{}{"message": "resource data"})
	})

	loginID := "user102"
	token := mockLogin(loginID)

	stputil.Disable(loginID, 3600)

	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	req.Header.Set("Authorization", token)

	w := serveOnce(t, app, req)
	// Disable 会踢掉所有会话，旧 token 立即失效，中间件先命中「未登录」而非封禁文案
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "not logged")
}

// TestIgnore_SkipsAuthentication 测试忽略认证装饰器
func TestIgnore_SkipsAuthentication(t *testing.T) {
	app := setupTestRouter()

	app.Get("/public", Ignore(), func(c irisfw.Context) {
		_ = c.JSON(map[string]interface{}{"message": "public data"})
	})

	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	w := serveOnce(t, app, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "public data")
}

// TestChainedMiddleware_CheckRoleAndHandler 测试链式中间件：CheckRole + 实际处理器
func TestChainedMiddleware_CheckRoleAndHandler(t *testing.T) {
	app := setupTestRouter()

	safeGroup := app.Party("/safe")
	safeGroup.Get("", CheckRole("SuperAdmin"), func(c irisfw.Context) {
		_ = c.JSON(map[string]interface{}{"message": "safe settings"})
	})

	token := mockLoginWithRole("admin123", []string{"SuperAdmin"})

	req := httptest.NewRequest(http.MethodGet, "/safe", nil)
	req.Header.Set("Authorization", token)

	w := serveOnce(t, app, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "safe settings")
}

// TestChainedMiddleware_CheckRoleAndHandler_NoRole 测试链式中间件：无角色访问
func TestChainedMiddleware_CheckRoleAndHandler_NoRole(t *testing.T) {
	app := setupTestRouter()

	safeGroup := app.Party("/safe")
	safeGroup.Get("", CheckRole("SuperAdmin"), func(c irisfw.Context) {
		_ = c.JSON(map[string]interface{}{"message": "safe settings"})
	})

	token := mockLoginWithRole("user123", []string{"User"})

	req := httptest.NewRequest(http.MethodGet, "/safe", nil)
	req.Header.Set("Authorization", token)

	w := serveOnce(t, app, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "role denied")
}

// TestGetHandler_WithNilHandler 测试 GetHandler 在 handler 为 nil 时的行为
func TestGetHandler_WithNilHandler(t *testing.T) {
	app := setupTestRouter()

	middleware := GetHandler(nil, &Annotation{CheckRole: []string{"Admin"}})

	app.Get("/test", middleware, func(c irisfw.Context) {
		_ = c.JSON(map[string]interface{}{"message": "test passed"})
	})

	token := mockLoginWithRole("testuser", []string{"Admin"})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", token)

	w := serveOnce(t, app, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test passed")
}

// TestMiddleware_CheckRole 测试 Middleware 函数的角色检查
func TestMiddleware_CheckRole(t *testing.T) {
	app := setupTestRouter()

	app.Get("/api/data", Middleware(&Annotation{CheckRole: []string{"Admin"}}), func(c irisfw.Context) {
		_ = c.JSON(map[string]interface{}{"data": "sensitive data"})
	})

	token := mockLoginWithRole("admin999", []string{"Admin"})

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.Header.Set("Authorization", token)

	w := serveOnce(t, app, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "sensitive data")
}

// TestParseTag 测试标签解析功能
func TestParseTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected *Annotation
	}{
		{name: "解析登录检查标签", tag: "sa_check_login", expected: &Annotation{CheckLogin: true}},
		{name: "解析角色检查标签", tag: "sa_check_role=Admin|SuperAdmin", expected: &Annotation{CheckRole: []string{"Admin", "SuperAdmin"}}},
		{name: "解析权限检查标签", tag: "sa_check_permission=user.read|user.write", expected: &Annotation{CheckPermission: []string{"user.read", "user.write"}}},
		{name: "解析忽略认证标签", tag: "sa_ignore", expected: &Annotation{Ignore: true}},
		{name: "解析封禁检查标签", tag: "sa_check_disable", expected: &Annotation{CheckDisable: true}},
		{name: "空标签", tag: "", expected: &Annotation{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseTag(tt.tag)
			assert.Equal(t, tt.expected.CheckLogin, result.CheckLogin)
			assert.Equal(t, tt.expected.CheckRole, result.CheckRole)
			assert.Equal(t, tt.expected.CheckPermission, result.CheckPermission)
			assert.Equal(t, tt.expected.CheckDisable, result.CheckDisable)
			assert.Equal(t, tt.expected.Ignore, result.Ignore)
		})
	}
}

// TestAnnotationValidate 测试注解验证功能
func TestAnnotationValidate(t *testing.T) {
	tests := []struct {
		name       string
		annotation *Annotation
		valid      bool
	}{
		{name: "有效的单一检查", annotation: &Annotation{CheckLogin: true}, valid: true},
		{name: "有效的忽略标记", annotation: &Annotation{Ignore: true, CheckLogin: true}, valid: true},
		{name: "有效的空注解", annotation: &Annotation{}, valid: true},
		{name: "无效的多重检查", annotation: &Annotation{CheckLogin: true, CheckRole: []string{"Admin"}}, valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.annotation.Validate()
			assert.Equal(t, tt.valid, result)
		})
	}
}
