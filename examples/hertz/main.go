// @Author daixk 2026/4/8 17:40:00
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/click33/sa-token-go/com/storage/redis"
	satoken "github.com/click33/sa-token-go/integrations/hertz"
	hertzapp "github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type LoginRequest struct {
	Username string `json:"username" query:"username" form:"username"`
	Password string `json:"password" query:"password" form:"password"`
}

type UserRequest struct {
	Username string `json:"username" query:"username" form:"username"`
}

func main() {
	ctx := context.Background()

	// initManager initializes manager initManager 初始化管理器
	initManager(ctx)

	// Create hertz server Create hertz server 创建 Hertz 服务
	h := server.Default(server.WithHostPorts(":8080"))
	h.Use(satoken.RegisterSaTokenContextMiddleware(ctx))

	// Register public routes Register public routes 注册公开路由
	api := h.Group("/api")
	api.POST("/login", handleLogin)
	api.GET("/public", handlePublic)

	// Register user routes Register user routes 注册用户路由
	user := h.Group("/api/user")
	user.Use(satoken.AuthMiddleware(ctx))
	user.GET("/info", handleUserInfo)
	user.POST("/logout", handleLogout)

	// Register admin routes Register admin routes 注册管理路由
	admin := h.Group("/api/admin")
	admin.Use(satoken.RoleMiddleware(ctx, []string{"admin"}))
	admin.GET("/users", handleAdminUsers)
	admin.POST("/disable", handleDisableUser)
	admin.POST("/enable", handleEnableUser)

	// Register permission routes Register permission routes 注册权限路由
	resource := h.Group("/api/resource")
	resource.Use(satoken.PermissionMiddleware(ctx, []string{"resource:read"}))
	resource.GET("/list", handleResourceList)

	// Register annotation routes Register annotation routes 注册注解路由
	annotation := h.Group("/api/annotation")
	annotation.GET("/profile", satoken.CheckLoginMiddleware(ctx, handleProfile, handleAuthFail))
	annotation.GET("/admin-data", satoken.CheckRoleMiddleware(ctx, []string{"admin"}, handleAdminData, handleAuthFail))
	annotation.GET("/sensitive", satoken.CheckPermissionMiddleware(ctx, []string{"data:read"}, handleSensitiveData, handleAuthFail))
	annotation.GET("/super", satoken.CheckAllMiddleware(ctx, []string{"super-admin"}, []string{"all:access"}, handleSuperData, handleAuthFail))

	// Print routes Print routes 打印路由
	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("Available endpoints:")
	fmt.Println("  POST /api/login")
	fmt.Println("  GET  /api/public")
	fmt.Println("  GET  /api/user/info")
	fmt.Println("  POST /api/user/logout")
	fmt.Println("  GET  /api/admin/users")
	fmt.Println("  POST /api/admin/disable")
	fmt.Println("  POST /api/admin/enable")
	fmt.Println("  GET  /api/resource/list")
	fmt.Println("  GET  /api/annotation/profile")
	fmt.Println("  GET  /api/annotation/admin-data")
	fmt.Println("  GET  /api/annotation/sensitive")
	fmt.Println("  GET  /api/annotation/super")

	h.Spin()
}

// initManager initializes manager initManager 初始化管理器
func initManager(ctx context.Context) {
	// Create storage Create storage 创建存储
	storage, err := redis.NewStorage("redis://:root@192.168.19.104:6379/0?dial_timeout=3&read_timeout=10s&max_retries=2")
	if err != nil {
		panic("failed to connect redis: " + err.Error())
	}

	// Build manager Build manager 构建管理器
	builder := satoken.NewDefaultBuilder()
	manager := builder.
		SetStorage(storage).
		Timeout(3600).
		ActiveTimeout(1800).
		MaxLoginCount(3).
		Build()

	satoken.SetManager(manager)

	fmt.Println("SaToken manager initialized successfully with Hertz storage")
	_ = ctx
}

func writeJSON(ctx *hertzapp.RequestContext, code int, message string, data interface{}) {
	ctx.JSON(200, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func getSaTokenContext(ctx *hertzapp.RequestContext) (*satoken.SaTokenContext, bool) {
	return satoken.GetSaTokenContext(ctx)
}

// handleAuthFail handles auth failure handleAuthFail 处理鉴权失败
func handleAuthFail(c context.Context, ctx *hertzapp.RequestContext, err error) {
	code := satoken.CodePermissionDenied
	switch {
	case errors.Is(err, satoken.ErrNotLogin), errors.Is(err, satoken.ErrInvalidToken), errors.Is(err, satoken.ErrTokenExpired):
		code = satoken.CodeNotLogin
	case errors.Is(err, satoken.ErrRoleDenied), errors.Is(err, satoken.ErrPermissionDenied), errors.Is(err, satoken.ErrAccountDisabled):
		code = satoken.CodePermissionDenied
	default:
		code = satoken.CodeServerError
	}

	writeJSON(ctx, code, err.Error(), nil)
	ctx.Abort()
	_ = c
}

// handleLogin handles login handleLogin 处理登录
func handleLogin(c context.Context, ctx *hertzapp.RequestContext) {
	var req LoginRequest
	if err := ctx.Bind(&req); err != nil {
		writeJSON(ctx, satoken.CodeBadRequest, "Invalid request", nil)
		return
	}

	if req.Username == "" || req.Password == "" {
		writeJSON(ctx, satoken.CodeBadRequest, "Username and password are required", nil)
		return
	}

	if req.Username != "admin" || req.Password != "123456" {
		writeJSON(ctx, satoken.CodeNotLogin, "Invalid username or password", nil)
		return
	}

	token, err := satoken.Login(c, req.Username)
	if err != nil {
		writeJSON(ctx, satoken.CodeServerError, fmt.Sprintf("Login failed: %v", err), nil)
		return
	}

	_ = satoken.AddRoles(c, req.Username, []string{"admin", "super-admin"})
	_ = satoken.AddPermissions(c, req.Username, []string{"resource:read", "resource:write", "data:read", "all:access"})

	writeJSON(ctx, satoken.CodeSuccess, "Login successful", map[string]interface{}{
		"token":    token,
		"username": req.Username,
	})
}

// handlePublic handles public api handlePublic 处理公开接口
func handlePublic(c context.Context, ctx *hertzapp.RequestContext) {
	writeJSON(ctx, satoken.CodeSuccess, "This is a public endpoint", "Anyone can access this")
	_ = c
}

// handleUserInfo handles user info handleUserInfo 处理用户信息
func handleUserInfo(c context.Context, ctx *hertzapp.RequestContext) {
	saCtx, ok := getSaTokenContext(ctx)
	if !ok {
		writeJSON(ctx, satoken.CodeServerError, "Failed to get SaToken context", nil)
		return
	}

	loginID, err := saCtx.GetLoginID(c)
	if err != nil {
		writeJSON(ctx, satoken.CodeNotLogin, fmt.Sprintf("Failed to get login ID: %v", err), nil)
		return
	}

	tokenInfo, err := saCtx.GetTokenInfo(c)
	if err != nil {
		writeJSON(ctx, satoken.CodeServerError, fmt.Sprintf("Failed to get token info: %v", err), nil)
		return
	}

	roles, _ := saCtx.GetRoles(c)
	permissions, _ := saCtx.GetPermissions(c)

	writeJSON(ctx, satoken.CodeSuccess, "User info retrieved successfully", map[string]interface{}{
		"loginID":     loginID,
		"tokenValue":  saCtx.GetTokenValue(),
		"device":      tokenInfo.Device,
		"createTime":  time.Unix(tokenInfo.CreateTime, 0).Format("2006-01-02 15:04:05"),
		"roles":       roles,
		"permissions": permissions,
	})
}

func handleLogout(c context.Context, ctx *hertzapp.RequestContext) {
	saCtx, ok := getSaTokenContext(ctx)
	if !ok {
		writeJSON(ctx, satoken.CodeNotLogin, "Not logged in", nil)
		return
	}

	if err := saCtx.Logout(c); err != nil {
		writeJSON(ctx, satoken.CodeServerError, fmt.Sprintf("Logout failed: %v", err), nil)
		return
	}

	writeJSON(ctx, satoken.CodeSuccess, "Logout successful", nil)
}

func handleAdminUsers(c context.Context, ctx *hertzapp.RequestContext) {
	saCtx, ok := getSaTokenContext(ctx)
	if !ok {
		writeJSON(ctx, satoken.CodeNotLogin, "Not logged in", nil)
		return
	}

	loginID, _ := saCtx.GetLoginID(c)
	writeJSON(ctx, satoken.CodeSuccess, "Admin users list", []map[string]interface{}{
		{"id": 1, "username": "admin", "role": "admin"},
		{"id": 2, "username": "user1", "role": "user"},
		{"id": 3, "username": loginID, "role": "current-admin"},
	})
}

// handleDisableUser handles disable user handleDisableUser 处理封禁用户
func handleDisableUser(c context.Context, ctx *hertzapp.RequestContext) {
	var req UserRequest
	if err := ctx.Bind(&req); err != nil || req.Username == "" {
		writeJSON(ctx, satoken.CodeBadRequest, "Username is required", nil)
		return
	}

	if err := satoken.Disable(c, req.Username, time.Hour, "Violated terms of service"); err != nil {
		writeJSON(ctx, satoken.CodeServerError, fmt.Sprintf("Failed to disable user: %v", err), nil)
		return
	}

	writeJSON(ctx, satoken.CodeSuccess, fmt.Sprintf("User %s has been disabled for 1 hour", req.Username), nil)
}

// handleEnableUser handles enable user handleEnableUser 处理解封用户
func handleEnableUser(c context.Context, ctx *hertzapp.RequestContext) {
	var req UserRequest
	if err := ctx.Bind(&req); err != nil || req.Username == "" {
		writeJSON(ctx, satoken.CodeBadRequest, "Username is required", nil)
		return
	}

	if err := satoken.Untie(c, req.Username); err != nil {
		writeJSON(ctx, satoken.CodeServerError, fmt.Sprintf("Failed to enable user: %v", err), nil)
		return
	}

	writeJSON(ctx, satoken.CodeSuccess, fmt.Sprintf("User %s has been enabled", req.Username), nil)
}

// handleResourceList handles resource list handleResourceList 处理资源列表
func handleResourceList(c context.Context, ctx *hertzapp.RequestContext) {
	saCtx, ok := getSaTokenContext(ctx)
	if !ok {
		writeJSON(ctx, satoken.CodeNotLogin, "Not logged in", nil)
		return
	}

	loginID, _ := saCtx.GetLoginID(c)
	writeJSON(ctx, satoken.CodeSuccess, "Resource list", []map[string]interface{}{
		{"id": 1, "name": "resource-a", "owner": loginID},
		{"id": 2, "name": "resource-b", "owner": "system"},
	})
}

// handleProfile handles profile api handleProfile 处理个人信息
func handleProfile(c context.Context, ctx *hertzapp.RequestContext) {
	saCtx, ok := getSaTokenContext(ctx)
	if !ok {
		writeJSON(ctx, satoken.CodeNotLogin, "Not logged in", nil)
		return
	}

	loginID, _ := saCtx.GetLoginID(c)
	writeJSON(ctx, satoken.CodeSuccess, "Profile data", map[string]interface{}{
		"loginID": loginID,
		"token":   saCtx.GetTokenValue(),
	})
}

func handleAdminData(c context.Context, ctx *hertzapp.RequestContext) {
	writeJSON(ctx, satoken.CodeSuccess, "Admin data", map[string]interface{}{
		"dashboard": "hertz-admin",
		"status":    "ok",
	})
	_ = c
}

// handleSensitiveData handles sensitive data handleSensitiveData 处理敏感数据
func handleSensitiveData(c context.Context, ctx *hertzapp.RequestContext) {
	writeJSON(ctx, satoken.CodeSuccess, "Sensitive data", map[string]interface{}{
		"scope":   "data:read",
		"records": 3,
	})
	_ = c
}

// handleSuperData handles super data handleSuperData 处理超级权限数据
func handleSuperData(c context.Context, ctx *hertzapp.RequestContext) {
	writeJSON(ctx, satoken.CodeSuccess, "Super admin data", map[string]interface{}{
		"system": "hertz",
		"level":  "super-admin",
	})
	_ = c
}
