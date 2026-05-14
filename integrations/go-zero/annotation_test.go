package gozero

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/click33/sa-token-go/core/config"
	"github.com/click33/sa-token-go/core/manager"
	"github.com/click33/sa-token-go/storage/memory"
	"github.com/click33/sa-token-go/stputil"
)

func setupTestManager() *manager.Manager {
	storage := memory.NewStorage()
	cfg := &config.Config{
		TokenName:     "Authorization",
		Timeout:       2592000,
		IsConcurrent:  true,
		IsShare:       true,
		MaxLoginCount: -1,
		IsReadHeader:  true,
		IsReadCookie:  false,
	}
	mgr := manager.NewManager(storage, cfg)
	stputil.SetManager(mgr)
	return mgr
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

func okHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":0,"message":"success"}`))
	}
}

func TestCheckLogin_Success(t *testing.T) {
	setupTestManager()
	token := mockLogin("user1")

	handler := GetHandler(okHandler(), &Annotation{CheckLogin: true})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", token)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCheckLogin_Failed(t *testing.T) {
	setupTestManager()

	handler := GetHandler(okHandler(), &Annotation{CheckLogin: true})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	handler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCheckRole_WithValidRole(t *testing.T) {
	setupTestManager()
	token := mockLoginWithRole("user1", []string{"Admin"})

	handler := GetHandler(okHandler(), &Annotation{CheckRole: []string{"Admin"}})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", token)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCheckRole_WithInvalidRole(t *testing.T) {
	setupTestManager()
	token := mockLoginWithRole("user1", []string{"User"})

	handler := GetHandler(okHandler(), &Annotation{CheckRole: []string{"Admin"}})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", token)
	handler(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "role denied")
}

func TestCheckRole_MultipleRoles(t *testing.T) {
	setupTestManager()
	token := mockLoginWithRole("user1", []string{"SuperAdmin"})

	handler := GetHandler(okHandler(), &Annotation{CheckRole: []string{"Admin", "SuperAdmin"}})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", token)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCheckPermission_WithValidPermission(t *testing.T) {
	setupTestManager()
	token := mockLoginWithPermission("user1", []string{"user:read"})

	handler := GetHandler(okHandler(), &Annotation{CheckPermission: []string{"user:read"}})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/data", nil)
	req.Header.Set("Authorization", token)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCheckPermission_WithInvalidPermission(t *testing.T) {
	setupTestManager()
	token := mockLoginWithPermission("user1", []string{"user:read"})

	handler := GetHandler(okHandler(), &Annotation{CheckPermission: []string{"admin:write"}})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", token)
	handler(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "permission denied")
}

func TestCheckDisable_NotDisabled(t *testing.T) {
	setupTestManager()
	token := mockLogin("user1")

	handler := GetHandler(okHandler(), &Annotation{CheckDisable: true})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", token)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIgnore_SkipsAuthentication(t *testing.T) {
	setupTestManager()

	handler := GetHandler(okHandler(), &Annotation{Ignore: true})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/public", nil)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetHandler_WithNilHandler(t *testing.T) {
	setupTestManager()

	assert.NotPanics(t, func() {
		h := GetHandler(nil, &Annotation{CheckLogin: true})
		assert.NotNil(t, h)
	})
}

func TestCheckLoginMiddleware(t *testing.T) {
	setupTestManager()
	token := mockLogin("user1")

	mw := CheckLoginMiddleware()
	handler := mw(okHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", token)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCheckRoleMiddleware(t *testing.T) {
	setupTestManager()
	token := mockLoginWithRole("user1", []string{"Admin"})

	mw := CheckRoleMiddleware("Admin")
	handler := mw(okHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", token)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCheckPermissionMiddleware(t *testing.T) {
	setupTestManager()
	token := mockLoginWithPermission("user1", []string{"user:read"})

	mw := CheckPermissionMiddleware("user:read")
	handler := mw(okHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/data", nil)
	req.Header.Set("Authorization", token)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCheckDisableMiddleware(t *testing.T) {
	setupTestManager()
	token := mockLogin("user1")

	mw := CheckDisableMiddleware()
	handler := mw(okHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", token)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIgnoreMiddleware(t *testing.T) {
	setupTestManager()

	mw := IgnoreMiddleware()
	handler := mw(okHandler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/public", nil)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
