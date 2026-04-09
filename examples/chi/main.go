// @Author daixk 2026/2/2 16:30:00
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/click33/sa-token-go/com/storage/redis"
	satoken "github.com/click33/sa-token-go/integrations/chi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Response defines unified response body Response 定义统一响应体
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// LoginRequest defines login payload LoginRequest 定义登录参数
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserRequest defines username payload UserRequest 定义用户名参数
type UserRequest struct {
	Username string `json:"username"`
}

func main() {
	ctx := context.Background()

	// initManager initializes manager initManager 初始化管理器
	initManager(ctx)

	// Create router Create router 创建路由
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(satoken.RegisterSaTokenContextMiddleware())

	// Register public routes Register public routes 注册公开路由
	r.Route("/api", func(api chi.Router) {
		api.Post("/login", handleLogin)
		api.Get("/public", handlePublic)
	})

	// Register user routes Register user routes 注册用户路由
	r.Route("/api/user", func(user chi.Router) {
		user.Use(satoken.AuthMiddleware())
		user.Get("/info", handleUserInfo)
		user.Post("/logout", handleLogout)
	})

	// Register admin routes Register admin routes 注册管理路由
	r.Route("/api/admin", func(admin chi.Router) {
		admin.Use(satoken.RoleMiddleware([]string{"admin"}))
		admin.Get("/users", handleAdminUsers)
		admin.Post("/disable", handleDisableUser)
		admin.Post("/enable", handleEnableUser)
	})

	// Register resource routes Register resource routes 注册权限路由
	r.Route("/api/resource", func(resource chi.Router) {
		resource.Use(satoken.PermissionMiddleware([]string{"resource:read"}))
		resource.Get("/list", handleResourceList)
	})

	// Register annotation routes Register annotation routes 注册注解路由
	r.Route("/api/annotation", func(annotation chi.Router) {
		annotation.Get("/profile", satoken.GetHandler(handleProfile, &satoken.Annotation{CheckLogin: true}))
		annotation.Get("/admin-data", satoken.GetHandler(handleAdminData, &satoken.Annotation{
			CheckLogin: true,
			CheckRole:  []string{"admin"},
		}))
		annotation.Get("/sensitive", satoken.GetHandler(handleSensitiveData, &satoken.Annotation{
			CheckLogin:      true,
			CheckPermission: []string{"data:read"},
		}))
		annotation.Get("/super", satoken.GetHandler(handleSuperData, &satoken.Annotation{
			CheckLogin:      true,
			CheckRole:       []string{"super-admin"},
			CheckPermission: []string{"all:access"},
		}))
	})

	// Print routes Print routes 打印路由信息
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
	fmt.Println("  GET  /api/annotation/*")

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}

// initManager initializes manager initManager 初始化管理器
func initManager(ctx context.Context) {
	// Create redis storage Create redis storage 创建 Redis 存储
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

	fmt.Println("SaToken manager initialized successfully with Chi storage")
	_ = ctx
}

// writeJSON writes response writeJSON 写入响应
func writeJSON(w http.ResponseWriter, status int, data Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// decodeJSON decodes request body decodeJSON 解析请求体
func decodeJSON(r *http.Request, target interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

// handleLogin handles login handleLogin 处理登录
func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusOK, Response{
			Code:    satoken.CodeBadRequest,
			Message: "Invalid request",
			Data:    nil,
		})
		return
	}

	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusOK, Response{
			Code:    satoken.CodeBadRequest,
			Message: "Username and password are required",
			Data:    nil,
		})
		return
	}

	if req.Username != "admin" || req.Password != "123456" {
		writeJSON(w, http.StatusOK, Response{
			Code:    satoken.CodeNotLogin,
			Message: "Invalid username or password",
			Data:    nil,
		})
		return
	}

	// Seed auth data after login Seed auth data after login 登录后写入示例权限数据
	token, err := satoken.Login(r.Context(), req.Username)
	if err != nil {
		writeJSON(w, http.StatusOK, Response{
			Code:    satoken.CodeServerError,
			Message: fmt.Sprintf("Login failed: %v", err),
			Data:    nil,
		})
		return
	}
	_ = satoken.AddRoles(r.Context(), req.Username, []string{"admin", "super-admin"})
	_ = satoken.AddPermissions(r.Context(), req.Username, []string{"resource:read", "resource:write", "data:read", "all:access"})

	writeJSON(w, http.StatusOK, Response{
		Code:    satoken.CodeSuccess,
		Message: "Login successful",
		Data: map[string]interface{}{
			"token":    token,
			"username": req.Username,
		},
	})
}

// handlePublic handles public endpoint handlePublic 处理公开接口
func handlePublic(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Response{
		Code:    satoken.CodeSuccess,
		Message: "This is a public endpoint",
		Data:    "Anyone can access this",
	})
}

// handleUserInfo handles user info handleUserInfo 处理用户信息
func handleUserInfo(w http.ResponseWriter, r *http.Request) {
	loginID, err := satoken.GetLoginIDByCtx(r.Context())
	if err != nil {
		writeJSON(w, http.StatusOK, Response{
			Code:    satoken.CodeNotLogin,
			Message: fmt.Sprintf("Failed to get login ID: %v", err),
			Data:    nil,
		})
		return
	}

	tokenInfo, err := satoken.GetTokenInfoByCtx(r.Context())
	if err != nil {
		writeJSON(w, http.StatusOK, Response{
			Code:    satoken.CodeServerError,
			Message: fmt.Sprintf("Failed to get token info: %v", err),
			Data:    nil,
		})
		return
	}

	// Read cached context Read cached context 读取缓存上下文
	saCtx, ok := satoken.GetSaTokenContextByCtx(r.Context())
	if !ok {
		writeJSON(w, http.StatusOK, Response{
			Code:    satoken.CodeServerError,
			Message: "Failed to get SaToken context",
			Data:    nil,
		})
		return
	}

	roles, _ := satoken.GetRoles(r.Context(), loginID)
	permissions, _ := satoken.GetPermissions(r.Context(), loginID)

	writeJSON(w, http.StatusOK, Response{
		Code:    satoken.CodeSuccess,
		Message: "User info retrieved successfully",
		Data: map[string]interface{}{
			"loginID":     loginID,
			"tokenValue":  saCtx.GetTokenValue(),
			"device":      tokenInfo.Device,
			"createTime":  time.Unix(tokenInfo.CreateTime, 0).Format("2006-01-02 15:04:05"),
			"roles":       roles,
			"permissions": permissions,
		},
	})
}

// handleLogout handles logout handleLogout 处理退出登录
func handleLogout(w http.ResponseWriter, r *http.Request) {
	saCtx, ok := satoken.GetSaTokenContextByCtx(r.Context())
	if !ok {
		writeJSON(w, http.StatusOK, Response{
			Code:    satoken.CodeNotLogin,
			Message: "Not logged in",
			Data:    nil,
		})
		return
	}

	if err := satoken.Logout(r.Context(), saCtx.GetTokenValue()); err != nil {
		writeJSON(w, http.StatusOK, Response{
			Code:    satoken.CodeServerError,
			Message: fmt.Sprintf("Logout failed: %v", err),
			Data:    nil,
		})
		return
	}

	writeJSON(w, http.StatusOK, Response{
		Code:    satoken.CodeSuccess,
		Message: "Logout successful",
		Data:    nil,
	})
}

// handleAdminUsers handles admin users handleAdminUsers 处理管理用户列表
func handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Response{
		Code:    satoken.CodeSuccess,
		Message: "Admin users list",
		Data: []map[string]interface{}{
			{"id": 1, "username": "admin", "role": "admin"},
			{"id": 2, "username": "user1", "role": "user"},
		},
	})
}

// handleDisableUser handles disable handleDisableUser 处理封禁用户
func handleDisableUser(w http.ResponseWriter, r *http.Request) {
	var req UserRequest
	if err := decodeJSON(r, &req); err != nil || req.Username == "" {
		writeJSON(w, http.StatusOK, Response{
			Code:    satoken.CodeBadRequest,
			Message: "Username is required",
			Data:    nil,
		})
		return
	}

	if err := satoken.Disable(r.Context(), req.Username, time.Hour, "Violated terms of service"); err != nil {
		writeJSON(w, http.StatusOK, Response{
			Code:    satoken.CodeServerError,
			Message: fmt.Sprintf("Failed to disable user: %v", err),
			Data:    nil,
		})
		return
	}

	writeJSON(w, http.StatusOK, Response{
		Code:    satoken.CodeSuccess,
		Message: fmt.Sprintf("User %s has been disabled for 1 hour", req.Username),
		Data:    nil,
	})
}

// handleEnableUser handles enable handleEnableUser 处理解除封禁
func handleEnableUser(w http.ResponseWriter, r *http.Request) {
	var req UserRequest
	if err := decodeJSON(r, &req); err != nil || req.Username == "" {
		writeJSON(w, http.StatusOK, Response{
			Code:    satoken.CodeBadRequest,
			Message: "Username is required",
			Data:    nil,
		})
		return
	}

	if err := satoken.Untie(r.Context(), req.Username); err != nil {
		writeJSON(w, http.StatusOK, Response{
			Code:    satoken.CodeServerError,
			Message: fmt.Sprintf("Failed to enable user: %v", err),
			Data:    nil,
		})
		return
	}

	writeJSON(w, http.StatusOK, Response{
		Code:    satoken.CodeSuccess,
		Message: fmt.Sprintf("User %s has been enabled", req.Username),
		Data:    nil,
	})
}

// handleResourceList handles resource list handleResourceList 处理资源列表
func handleResourceList(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Response{
		Code:    satoken.CodeSuccess,
		Message: "Resource list",
		Data: []map[string]interface{}{
			{"id": 1, "name": "Resource 1", "type": "document"},
			{"id": 2, "name": "Resource 2", "type": "image"},
		},
	})
}

// handleProfile handles profile handleProfile 处理个人资料
func handleProfile(w http.ResponseWriter, r *http.Request) {
	loginID, _ := satoken.GetLoginIDByCtx(r.Context())

	writeJSON(w, http.StatusOK, Response{
		Code:    satoken.CodeSuccess,
		Message: "Profile data",
		Data: map[string]interface{}{
			"username": loginID,
			"email":    loginID + "@example.com",
		},
	})
}

// handleAdminData handles admin data handleAdminData 处理管理数据
func handleAdminData(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Response{
		Code:    satoken.CodeSuccess,
		Message: "Admin data",
		Data:    "This is admin-only data",
	})
}

// handleSensitiveData handles sensitive data handleSensitiveData 处理敏感数据
func handleSensitiveData(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Response{
		Code:    satoken.CodeSuccess,
		Message: "Sensitive data",
		Data:    "This is sensitive data requiring data:read permission",
	})
}

// handleSuperData handles super data handleSuperData 处理超级管理数据
func handleSuperData(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Response{
		Code:    satoken.CodeSuccess,
		Message: "Super admin data",
		Data:    "This requires super-admin role and all:access permission",
	})
}
