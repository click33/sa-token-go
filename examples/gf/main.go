// @Author daixk 2026/2/2 11:46:00
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/click33/sa-token-go/com/storage/redis"
	satoken "github.com/click33/sa-token-go/integrations/gf"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	ctx := gctx.New()

	// Initialize SaToken Manager initManager 初始化 SaToken 管理器
	initManager(ctx)

	// Create HTTP server Create HTTP server 创建 HTTP 服务
	s := g.Server()

	// Register middleware Register middleware 注册中间件
	s.Use(satoken.RegisterSaTokenContextMiddleware(ctx))

	// Public routes Public routes 公共路由
	s.Group("/api", func(group *ghttp.RouterGroup) {
		group.POST("/login", handleLogin)
		group.GET("/public", handlePublic)
	})

	// Protected routes Protected routes 受保护路由
	s.Group("/api/user", func(group *ghttp.RouterGroup) {
		group.Middleware(satoken.AuthMiddleware(ctx))
		group.GET("/info", handleUserInfo)
		group.POST("/logout", handleLogout)
	})

	// Admin routes Admin routes 管理员路由
	s.Group("/api/admin", func(group *ghttp.RouterGroup) {
		group.Middleware(satoken.RoleMiddleware(ctx, []string{"admin"}))
		group.GET("/users", handleAdminUsers)
		group.POST("/disable", handleDisableUser)
		group.POST("/enable", handleEnableUser)
	})

	// Permission routes Permission routes 权限路由
	s.Group("/api/resource", func(group *ghttp.RouterGroup) {
		group.Middleware(satoken.PermissionMiddleware(ctx, []string{"resource:read"}))
		group.GET("/list", handleResourceList)
	})

	// Annotation routes Annotation routes 注解路由
	s.Group("/api/annotation", func(group *ghttp.RouterGroup) {
		group.GET("/profile", satoken.CheckLoginMiddleware(ctx, handleProfile, handleAuthFail))
		group.GET("/admin-data", satoken.CheckRoleMiddleware(ctx, []string{"admin"}, handleAdminData, handleAuthFail))
		group.GET("/sensitive", satoken.CheckPermissionMiddleware(ctx, []string{"data:read"}, handleSensitiveData, handleAuthFail))
		group.GET("/super", satoken.CheckAllMiddleware(ctx, []string{"super-admin"}, []string{"all:access"}, handleSuperData, handleAuthFail))
	})

	// Start server Start server 启动服务
	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("服务器启动在 http://localhost:8080")
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("可用的接口:")
	fmt.Println("  POST /api/login          - Login (username: admin, password: 123456)")
	fmt.Println("  GET  /api/public         - Public endpoint")
	fmt.Println("  GET  /api/user/info      - Get user info (requires login)")
	fmt.Println("  POST /api/user/logout    - Logout")
	fmt.Println("  GET  /api/admin/users    - Admin users list (requires admin role)")
	fmt.Println("  POST /api/admin/disable  - Disable user (requires admin role)")
	fmt.Println("  POST /api/admin/enable   - Enable user (requires admin role)")
	fmt.Println("  GET  /api/resource/list  - Resource list (requires resource:read permission)")
	fmt.Println("  GET  /api/annotation/*   - Annotation-based routes")

	s.SetPort(8080)
	s.Run()
}

// initManager initializes the SaToken manager initManager 初始化 SaToken 管理器
func initManager(ctx context.Context) {
	// Use Redis storage Use Redis storage 使用 Redis 存储
	storage, err := redis.NewStorage("redis://:root@192.168.19.104:6379/0?dial_timeout=3&read_timeout=10s&max_retries=2")
	if err != nil {
		panic("Failed to connect to Redis: " + err.Error())
	}

	// Build manager Build manager 构建管理器
	builder := satoken.NewDefaultBuilder()
	mgr := builder.
		SetStorage(storage).
		Timeout(3600).
		ActiveTimeout(1800).
		MaxLoginCount(3).
		Build()

	// Set manager Set manager 设置管理器
	satoken.SetManager(mgr)

	fmt.Println("SaToken Manager initialized successfully with GoFrame Redis storage")
	fmt.Println("SaToken 管理器初始化成功（使用 GoFrame Redis 存储）")
}

// handleLogin handles user login handleLogin 处理用户登录
func handleLogin(r *ghttp.Request) {
	username := r.Get("username").String()
	password := r.Get("password").String()

	// Validate input Validate input 校验输入
	if username == "" || password == "" {
		r.Response.WriteJson(g.Map{
			"code":    satoken.CodeBadRequest,
			"message": "Username and password are required",
			"data":    nil,
		})
		return
	}

	// Mock authentication Mock authentication 模拟认证
	if username != "admin" || password != "123456" {
		r.Response.WriteJson(g.Map{
			"code":    satoken.CodeNotLogin,
			"message": "Invalid username or password",
			"data":    nil,
		})
		return
	}

	// Login and issue token Login and issue token 登录并签发 token
	token, err := satoken.Login(r.Context(), username)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    satoken.CodeServerError,
			"message": fmt.Sprintf("Login failed: %v", err),
			"data":    nil,
		})
		return
	}

	// Seed roles and permissions Seed roles and permissions 初始化角色和权限
	// _ = satoken.AddRoles(r.Context(), username, []string{"admin", "super-admin"})
	// _ = satoken.AddPermissions(r.Context(), username, []string{"resource:read", "resource:write", "data:read", "all:access"})

	r.Response.WriteJson(g.Map{
		"code":    satoken.CodeSuccess,
		"message": "Login successful",
		"data": g.Map{
			"token":    token,
			"username": username,
		},
	})
}

// handlePublic handles public endpoint handlePublic 处理公共接口
func handlePublic(r *ghttp.Request) {
	r.Response.WriteJson(g.Map{
		"code":    satoken.CodeSuccess,
		"message": "This is a public endpoint",
		"data":    "Anyone can access this",
	})
}

// handleUserInfo handles user info request handleUserInfo 处理用户信息请求
func handleUserInfo(r *ghttp.Request) {
	loginID, err := satoken.GetLoginIDByCtx(r.Context())
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    satoken.CodeNotLogin,
			"message": fmt.Sprintf("Failed to get login ID: %v", err),
			"data":    nil,
		})
		return
	}

	tokenInfo, err := satoken.GetTokenInfoByCtx(r.Context())
	if err != nil {
		r.Response.WriteJson(g.Map{
			"code":    satoken.CodeServerError,
			"message": fmt.Sprintf("Failed to get token info: %v", err),
			"data":    nil,
		})
		return
	}

	// Get SaToken context Get SaToken context 获取 SaToken 上下文
	saCtx, ok := satoken.GetSaTokenContextByCtx(r.Context())
	if !ok {
		r.Response.WriteJson(g.Map{
			"code":    satoken.CodeServerError,
			"message": "Failed to get SaToken context",
			"data":    nil,
		})
		return
	}

	roles, _ := satoken.GetRoles(r.Context(), loginID)
	permissions, _ := satoken.GetPermissions(r.Context(), loginID)

	r.Response.WriteJson(g.Map{
		"code":    satoken.CodeSuccess,
		"message": "User info retrieved successfully",
		"data": g.Map{
			"loginID":     loginID,
			"tokenValue":  saCtx.GetTokenValue(),
			"device":      tokenInfo.Device,
			"createTime":  time.Unix(tokenInfo.CreateTime, 0).Format("2006-01-02 15:04:05"),
			"roles":       roles,
			"permissions": permissions,
		},
	})
}

// handleLogout handles user logout handleLogout 处理用户登出
func handleLogout(r *ghttp.Request) {
	// Get SaToken context Get SaToken context 获取 SaToken 上下文
	saCtx, ok := satoken.GetSaTokenContextByCtx(r.Context())
	if !ok {
		r.Response.WriteJson(g.Map{
			"code":    satoken.CodeNotLogin,
			"message": "Not logged in",
			"data":    nil,
		})
		return
	}

	if err := satoken.Logout(r.Context(), saCtx.GetTokenValue()); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    satoken.CodeServerError,
			"message": fmt.Sprintf("Logout failed: %v", err),
			"data":    nil,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    satoken.CodeSuccess,
		"message": "Logout successful",
		"data":    nil,
	})
}

// handleAdminUsers handles admin users list handleAdminUsers 处理管理员用户列表
func handleAdminUsers(r *ghttp.Request) {
	r.Response.WriteJson(g.Map{
		"code":    satoken.CodeSuccess,
		"message": "Admin users list",
		"data": []g.Map{
			{"id": 1, "username": "admin", "role": "admin"},
			{"id": 2, "username": "user1", "role": "user"},
		},
	})
}

// handleDisableUser handles disabling a user handleDisableUser 处理封禁用户
func handleDisableUser(r *ghttp.Request) {
	targetUser := r.Get("username").String()
	if targetUser == "" {
		r.Response.WriteJson(g.Map{
			"code":    satoken.CodeBadRequest,
			"message": "Username is required",
			"data":    nil,
		})
		return
	}

	if err := satoken.Disable(r.Context(), targetUser, time.Hour, "Violated terms of service"); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    satoken.CodeServerError,
			"message": fmt.Sprintf("Failed to disable user: %v", err),
			"data":    nil,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    satoken.CodeSuccess,
		"message": fmt.Sprintf("User %s has been disabled for 1 hour", targetUser),
		"data":    nil,
	})
}

// handleEnableUser handles enabling a user handleEnableUser 处理解封用户
func handleEnableUser(r *ghttp.Request) {
	targetUser := r.Get("username").String()
	if targetUser == "" {
		r.Response.WriteJson(g.Map{
			"code":    satoken.CodeBadRequest,
			"message": "Username is required",
			"data":    nil,
		})
		return
	}

	if err := satoken.Untie(r.Context(), targetUser); err != nil {
		r.Response.WriteJson(g.Map{
			"code":    satoken.CodeServerError,
			"message": fmt.Sprintf("Failed to enable user: %v", err),
			"data":    nil,
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    satoken.CodeSuccess,
		"message": fmt.Sprintf("User %s has been enabled", targetUser),
		"data":    nil,
	})
}

// handleResourceList handles resource list handleResourceList 处理资源列表
func handleResourceList(r *ghttp.Request) {
	r.Response.WriteJson(g.Map{
		"code":    satoken.CodeSuccess,
		"message": "Resource list",
		"data": []g.Map{
			{"id": 1, "name": "Resource 1", "type": "document"},
			{"id": 2, "name": "Resource 2", "type": "image"},
		},
	})
}

// handleProfile handles profile request handleProfile 处理个人资料请求
func handleProfile(r *ghttp.Request) {
	loginID, _ := satoken.GetLoginIDByCtx(r.Context())
	r.Response.WriteJson(g.Map{
		"code":    satoken.CodeSuccess,
		"message": "Profile data",
		"data": g.Map{
			"username": loginID,
			"email":    loginID + "@example.com",
		},
	})
}

// handleAdminData handles admin data request handleAdminData 处理管理员数据请求
func handleAdminData(r *ghttp.Request) {
	r.Response.WriteJson(g.Map{
		"code":    satoken.CodeSuccess,
		"message": "Admin data",
		"data":    "This is admin-only data",
	})
}

// handleSensitiveData handles sensitive data request handleSensitiveData 处理敏感数据请求
func handleSensitiveData(r *ghttp.Request) {
	r.Response.WriteJson(g.Map{
		"code":    satoken.CodeSuccess,
		"message": "Sensitive data",
		"data":    "This is sensitive data requiring data:read permission",
	})
}

// handleSuperData handles super admin data request handleSuperData 处理超级管理员数据请求
func handleSuperData(r *ghttp.Request) {
	r.Response.WriteJson(g.Map{
		"code":    satoken.CodeSuccess,
		"message": "Super admin data",
		"data":    "This requires super-admin role and all:access permission",
	})
}

// handleAuthFail handles authentication failure handleAuthFail 处理认证失败
func handleAuthFail(r *ghttp.Request, err error) {
	var code int
	var message string

	switch err {
	case satoken.ErrNotLogin:
		code = satoken.CodeNotLogin
		message = "Not logged in"
	case satoken.ErrPermissionDenied:
		code = satoken.CodePermissionDenied
		message = "Permission denied"
	case satoken.ErrRoleDenied:
		code = satoken.CodePermissionDenied
		message = "Role denied"
	case satoken.ErrAccountDisabled:
		code = satoken.CodeAccountDisabled
		message = "Account disabled"
	default:
		code = satoken.CodeServerError
		message = err.Error()
	}

	r.Response.WriteJson(g.Map{
		"code":    code,
		"message": message,
		"data":    nil,
	})
}
