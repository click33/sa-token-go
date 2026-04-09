// @Author daixk 2026/2/2
// Quick Start Example - Comprehensive Test Suite for SaToken Framework Quick Start Example - SaToken 框架完整测试套件

package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"

	"github.com/click33/sa-token-go/com/storage/redis"
	satoken "github.com/click33/sa-token-go/integrations/gin"
)

// Response defines unified response payload 统一响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// LoginRequest defines login request payload 登录请求
type LoginRequest struct {
	LoginID  string `json:"loginId" binding:"required"`
	Device   string `json:"device"`
	DeviceId string `json:"deviceId"`
}

// LoginByTokenRequest defines token refresh request 通过 Token 登录请求
type LoginByTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// PermissionRequest defines permission request payload 权限请求
type PermissionRequest struct {
	LoginID     string   `json:"loginId" binding:"required"`
	Permissions []string `json:"permissions" binding:"required"`
}

// RoleRequest defines role request payload 角色请求
type RoleRequest struct {
	LoginID string   `json:"loginId" binding:"required"`
	Roles   []string `json:"roles" binding:"required"`
}

// DisableRequest defines disable request payload 封禁请求
type DisableRequest struct {
	LoginID  string `json:"loginId" binding:"required"`
	Duration int64  `json:"duration" binding:"required"` // 封禁时长（秒）
	Reason   string `json:"reason"`
}

// DeviceRequest defines device request payload 设备请求
type DeviceRequest struct {
	LoginID  string `json:"loginId" binding:"required"`
	Device   string `json:"device"`
	DeviceId string `json:"deviceId"`
}

// TokenRequest defines token request payload Token 请求
type TokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// success writes success response 成功响应
func success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	})
}

// fail writes failure response 失败响应
func fail(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: 500,
		Msg:  msg,
	})
}

// unauthorized writes 401 response 未授权响应 返回401未授权响应
func unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code: 401,
		Msg:  msg,
	})
	c.Abort()
}

// authMiddleware checks login state 登录验证中间件 验证用户是否已登录，并将token存入上下文
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Token From Header 从Header中获取token
		token := c.GetHeader("Authorization")
		if token == "" {
			unauthorized(c, "未提供认证token")
			return
		}

		// Validate Token 验证token是否有效
		if err := satoken.CheckLogin(c.Request.Context(), token); err != nil {
			unauthorized(c, "token无效或已过期: "+err.Error())
			return
		}

		// Store Token In Context 将token存入上下文
		c.Set("token", token)

		// Continue Request Handling 继续处理请求
		c.Next()
	}
}

// initSaToken initializes the SaToken framework initSaToken 初始化 SaToken 框架
func initSaToken() error {
	// Use Redis Storage And URL Format 使用 Redis 存储与 Redis URL 格式
	storage, err := redis.NewStorage("redis://:root@192.168.19.104:6379/0?dial_timeout=3&read_timeout=10s&max_retries=2")
	if err != nil {
		panic("Failed to connect to Redis: " + err.Error())
	}

	// Build Manager With Builder 使用 Builder 构建管理器
	mgr := satoken.NewDefaultBuilder().
		TokenName("token").    // Token 名称
		Timeout(7200).         // 超时时间 2 小时
		RenewMaxRefresh(1800). // 续期触发阈值 30 分钟
		IsConcurrent(true).    // 允许并发登录
		MaxLoginCount(5).      // 最大并发登录数 5
		IsReadHeader(true).    // 从 Header 读取 Token
		IsLog(true).           // 开启日志
		IsPrintBanner(true).   // 打印启动 Banner
		SetStorage(storage).   // 设置存储适配器
		Build()

	// Set Global Manager 设置全局管理器
	satoken.SetManager(mgr)

	fmt.Println("SaToken 框架初始化成功")
	return nil
}

// handleLogin handles login 用户登录
func handleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	var token string
	var err error

	// Choose Login Method By Device Info 根据是否提供设备信息选择登录方式
	if req.Device != "" && req.DeviceId != "" {
		token, err = satoken.Login(c.Request.Context(), req.LoginID, req.Device, req.DeviceId)
	} else if req.Device != "" {
		token, err = satoken.Login(c.Request.Context(), req.LoginID, req.Device)
	} else {
		token, err = satoken.Login(c.Request.Context(), req.LoginID)
	}

	if err != nil {
		fail(c, "登录失败: "+err.Error())
		return
	}

	success(c, gin.H{
		"token":   token,
		"loginId": req.LoginID,
	})
}

// handleLoginByToken handles login by token 通过 Token 续期登录
func handleLoginByToken(c *gin.Context) {
	var req LoginByTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	if err := satoken.LoginByToken(c.Request.Context(), req.Token); err != nil {
		fail(c, "Token 续期失败: "+err.Error())
		return
	}

	success(c, "Token 续期成功")
}

// handleLogout handles logout 用户登出
func handleLogout(c *gin.Context) {
	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	if err := satoken.Logout(c.Request.Context(), tokenStr); err != nil {
		fail(c, "登出失败: "+err.Error())
		return
	}

	success(c, "登出成功")
}

// handleLogoutByDevice handles logout by device 根据设备类型登出
func handleLogoutByDevice(c *gin.Context) {
	var req DeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	if err := satoken.LogoutByDevice(c.Request.Context(), req.LoginID, req.Device); err != nil {
		fail(c, "登出失败: "+err.Error())
		return
	}

	success(c, "登出成功")
}

// handleLogoutByDeviceAndDeviceId handles logout by device and device id 根据设备类型和设备ID登出
func handleLogoutByDeviceAndDeviceId(c *gin.Context) {
	var req DeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	if err := satoken.LogoutByDeviceAndDeviceId(c.Request.Context(), req.LoginID, req.Device, req.DeviceId); err != nil {
		fail(c, "登出失败: "+err.Error())
		return
	}

	success(c, "登出成功")
}

// handleLogoutByLoginID handles logout by login id 根据 LoginID 登出所有终端
func handleLogoutByLoginID(c *gin.Context) {
	var req struct {
		LoginID string `json:"loginId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	if err := satoken.LogoutByLoginID(c.Request.Context(), req.LoginID); err != nil {
		fail(c, "登出失败: "+err.Error())
		return
	}

	success(c, "登出成功")
}

// handleIsLogin handles is login 检查用户是否登录
func handleIsLogin(c *gin.Context) {
	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	isLogin := satoken.IsLogin(c.Request.Context(), tokenStr)

	success(c, gin.H{
		"isLogin": isLogin,
	})
}

// handleCheckLogin validates login status 验证登录状态（未登录返回错误）
func handleCheckLogin(c *gin.Context) {
	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	if err := satoken.CheckLogin(c.Request.Context(), tokenStr); err != nil {
		fail(c, "未登录: "+err.Error())
		return
	}

	success(c, "已登录")
}

// handleGetLoginID handles get login id 获取登录ID
func handleGetLoginID(c *gin.Context) {
	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	loginID, err := satoken.GetLoginID(c.Request.Context(), tokenStr)
	if err != nil {
		fail(c, "获取登录ID失败: "+err.Error())
		return
	}

	success(c, gin.H{
		"loginId": loginID,
	})
}

// handleGetTokenInfo handles get token info 获取 Token 信息
func handleGetTokenInfo(c *gin.Context) {
	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	tokenInfo, err := satoken.GetTokenInfo(c.Request.Context(), tokenStr)
	if err != nil {
		fail(c, "获取Token信息失败: "+err.Error())
		return
	}

	success(c, tokenInfo)
}

// handleGetDevice handles get device 获取设备类型
func handleGetDevice(c *gin.Context) {
	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	device, err := satoken.GetDevice(c.Request.Context(), tokenStr)
	if err != nil {
		fail(c, "获取设备类型失败: "+err.Error())
		return
	}

	success(c, gin.H{
		"device": device,
	})
}

// handleGetDeviceId handles get device id 获取设备ID
func handleGetDeviceId(c *gin.Context) {
	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	deviceId, err := satoken.GetDeviceId(c.Request.Context(), tokenStr)
	if err != nil {
		fail(c, "获取设备ID失败: "+err.Error())
		return
	}

	success(c, gin.H{
		"deviceId": deviceId,
	})
}

// handleGetTokenCreateTime handles get token create time 获取 Token 创建时间
func handleGetTokenCreateTime(c *gin.Context) {
	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	createTime, err := satoken.GetTokenCreateTime(c.Request.Context(), tokenStr)
	if err != nil {
		fail(c, "获取Token创建时间失败: "+err.Error())
		return
	}

	success(c, gin.H{
		"createTime": createTime,
	})
}

// handleGetTokenTTL handles get token ttl 获取 Token 剩余有效时间
func handleGetTokenTTL(c *gin.Context) {
	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	ttl, err := satoken.GetTokenTTL(c.Request.Context(), tokenStr)
	if err != nil {
		fail(c, "获取Token TTL失败: "+err.Error())
		return
	}

	success(c, gin.H{
		"ttl": ttl,
	})
}

// handleGetOnlineTerminalCount handles get online terminal count 获取在线终端总数
func handleGetOnlineTerminalCount(c *gin.Context) {
	loginID := c.Param("loginId")
	if loginID == "" {
		fail(c, "loginId 不能为空")
		return
	}

	count, err := satoken.GetOnlineTerminalCount(c.Request.Context(), loginID)
	if err != nil {
		fail(c, "获取在线终端数失败: "+err.Error())
		return
	}

	success(c, gin.H{
		"count": count,
	})
}

// handleGetOnlineTerminalCountByDevice handles get online terminal count by device 获取指定设备的在线终端数
func handleGetOnlineTerminalCountByDevice(c *gin.Context) {
	loginID := c.Param("loginId")
	device := c.Param("device")
	fmt.Println("loginID: %s, device: %s", loginID, device)
	if loginID == "" || device == "" {
		fail(c, "loginId 和 device 不能为空")
		return
	}

	count, err := satoken.GetOnlineTerminalCountByDevice(c.Request.Context(), loginID, device)
	if err != nil {
		fail(c, "获取在线终端数失败: "+err.Error())
		return
	}

	success(c, gin.H{
		"count": count,
	})
}

// handleGetOnlineTerminalCountByDeviceAndDeviceId handles get online terminal count by device and device id 获取指定设备和设备ID的在线终端数
func handleGetOnlineTerminalCountByDeviceAndDeviceId(c *gin.Context) {
	loginID := c.Param("loginId")
	device := c.Param("device")
	deviceId := c.Param("deviceId")
	if loginID == "" || device == "" || deviceId == "" {
		fail(c, "loginId、device 和 deviceId 不能为空")
		return
	}

	count, err := satoken.GetOnlineTerminalCountByDeviceAndDeviceId(c.Request.Context(), loginID, device, deviceId)
	if err != nil {
		fail(c, "获取在线终端数失败: "+err.Error())
		return
	}

	success(c, gin.H{
		"count": count,
	})
}

// handleKickout handles kickout 根据 Token 踢人下线
func handleKickout(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	if err := satoken.Kickout(c.Request.Context(), req.Token); err != nil {
		fail(c, "踢人下线失败: "+err.Error())
		return
	}

	success(c, "踢人下线成功")
}

// handleKickoutByDevice handles kickout by device 根据设备类型踢人下线
func handleKickoutByDevice(c *gin.Context) {
	var req DeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	if err := satoken.KickoutByDevice(c.Request.Context(), req.LoginID, req.Device); err != nil {
		fail(c, "踢人下线失败: "+err.Error())
		return
	}

	success(c, "踢人下线成功")
}

// handleKickoutByDeviceAndDeviceId handles kickout by device and device id 根据设备和设备ID踢人下线
func handleKickoutByDeviceAndDeviceId(c *gin.Context) {
	var req DeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	if err := satoken.KickoutByDeviceAndDeviceId(c.Request.Context(), req.LoginID, req.Device, req.DeviceId); err != nil {
		fail(c, "踢人下线失败: "+err.Error())
		return
	}

	success(c, "踢人下线成功")
}

// handleKickoutByLoginID handles kickout by login id 根据 LoginID 踢出所有终端
func handleKickoutByLoginID(c *gin.Context) {
	var req struct {
		LoginID string `json:"loginId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	if err := satoken.KickoutByLoginID(c.Request.Context(), req.LoginID); err != nil {
		fail(c, "踢人下线失败: "+err.Error())
		return
	}

	success(c, "踢人下线成功")
}

// handleReplace handles replace 根据 Token 顶人下线
func handleReplace(c *gin.Context) {
	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	if err := satoken.Replace(c.Request.Context(), tokenStr); err != nil {
		fail(c, "顶人下线失败: "+err.Error())
		return
	}

	success(c, "顶人下线成功")
}

// handleReplaceByDevice handles replace by device 根据设备类型顶人下线
func handleReplaceByDevice(c *gin.Context) {
	var req DeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	if err := satoken.ReplaceByDevice(c.Request.Context(), req.LoginID, req.Device); err != nil {
		fail(c, "顶人下线失败: "+err.Error())
		return
	}

	success(c, "顶人下线成功")
}

// handleReplaceByDeviceAndDeviceId handles replace by device and device id 根据设备和设备ID顶人下线
func handleReplaceByDeviceAndDeviceId(c *gin.Context) {
	var req DeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	if err := satoken.ReplaceByDeviceAndDeviceId(c.Request.Context(), req.LoginID, req.Device, req.DeviceId); err != nil {
		fail(c, "顶人下线失败: "+err.Error())
		return
	}

	success(c, "顶人下线成功")
}

// handleReplaceByLoginID handles replace by login id 根据 LoginID 顶替所有终端
func handleReplaceByLoginID(c *gin.Context) {
	var req struct {
		LoginID string `json:"loginId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	if err := satoken.ReplaceByLoginID(c.Request.Context(), req.LoginID); err != nil {
		fail(c, "顶人下线失败: "+err.Error())
		return
	}

	success(c, "顶人下线成功")
}

// handleAddPermissions handles add permissions 为用户添加权限
func handleAddPermissions(c *gin.Context) {
	var req PermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	if err := satoken.AddPermissions(c.Request.Context(), req.LoginID, req.Permissions); err != nil {
		fail(c, "添加权限失败: "+err.Error())
		return
	}

	success(c, "添加权限成功")
}

// handleAddPermissionsByToken handles add permissions by token 根据 Token 添加权限
func handleAddPermissionsByToken(c *gin.Context) {
	var req struct {
		Permissions []string `json:"permissions" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	if err := satoken.AddPermissionsByToken(c.Request.Context(), tokenStr, req.Permissions); err != nil {
		fail(c, "添加权限失败: "+err.Error())
		return
	}

	success(c, "添加权限成功")
}

// handleRemovePermissions handles remove permissions 删除用户权限
func handleRemovePermissions(c *gin.Context) {
	var req PermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	if err := satoken.RemovePermissions(c.Request.Context(), req.LoginID, req.Permissions); err != nil {
		fail(c, "删除权限失败: "+err.Error())
		return
	}

	success(c, "删除权限成功")
}

// handleRemovePermissionsByToken handles remove permissions by token 根据 Token 删除权限
func handleRemovePermissionsByToken(c *gin.Context) {
	var req struct {
		Permissions []string `json:"permissions" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	if err := satoken.RemovePermissionsByToken(c.Request.Context(), tokenStr, req.Permissions); err != nil {
		fail(c, "删除权限失败: "+err.Error())
		return
	}

	success(c, "删除权限成功")
}

// handleGetPermissions handles get permissions 获取用户权限列表
func handleGetPermissions(c *gin.Context) {
	loginID := c.Param("loginId")
	if loginID == "" {
		fail(c, "loginId 不能为空")
		return
	}

	permissions, err := satoken.GetPermissions(c.Request.Context(), loginID)
	if err != nil {
		fail(c, "获取权限列表失败: "+err.Error())
		return
	}

	success(c, gin.H{
		"permissions": permissions,
	})
}

// handleGetPermissionsByToken handles get permissions by token 根据 Token 获取权限列表
func handleGetPermissionsByToken(c *gin.Context) {
	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	permissions, err := satoken.GetPermissionsByToken(c.Request.Context(), tokenStr)
	if err != nil {
		fail(c, "获取权限列表失败: "+err.Error())
		return
	}

	success(c, gin.H{
		"permissions": permissions,
	})
}

// handleHasPermission handles has permission 检查是否拥有指定权限
func handleHasPermission(c *gin.Context) {
	var req struct {
		LoginID    string `json:"loginId" binding:"required"`
		Permission string `json:"permission" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	has := satoken.HasPermission(c.Request.Context(), req.LoginID, req.Permission)

	success(c, gin.H{
		"hasPermission": has,
	})
}

// handleHasPermissionByToken handles has permission by token 根据 Token 检查权限
func handleHasPermissionByToken(c *gin.Context) {
	var req struct {
		Permission string `json:"permission" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	has := satoken.HasPermissionByToken(c.Request.Context(), tokenStr, req.Permission)

	success(c, gin.H{
		"hasPermission": has,
	})
}

// handleHasPermissionsAnd checks all permissions 检查是否拥有所有权限（AND逻辑）
func handleHasPermissionsAnd(c *gin.Context) {
	var req PermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	has := satoken.HasPermissionsAnd(c.Request.Context(), req.LoginID, req.Permissions)

	success(c, gin.H{
		"hasAllPermissions": has,
	})
}

// handleHasPermissionsAndByToken handles has permissions and by token 根据 Token 检查所有权限
func handleHasPermissionsAndByToken(c *gin.Context) {
	var req struct {
		Permissions []string `json:"permissions" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	has := satoken.HasPermissionsAndByToken(c.Request.Context(), tokenStr, req.Permissions)

	success(c, gin.H{
		"hasAllPermissions": has,
	})
}

// handleHasPermissionsOr checks any permission 检查是否拥有任一权限（OR逻辑）
func handleHasPermissionsOr(c *gin.Context) {
	var req PermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	has := satoken.HasPermissionsOr(c.Request.Context(), req.LoginID, req.Permissions)

	success(c, gin.H{
		"hasAnyPermission": has,
	})
}

// handleHasPermissionsOrByToken handles has permissions or by token 根据 Token 检查任一权限
func handleHasPermissionsOrByToken(c *gin.Context) {
	var req struct {
		Permissions []string `json:"permissions" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	has := satoken.HasPermissionsOrByToken(c.Request.Context(), tokenStr, req.Permissions)

	success(c, gin.H{
		"hasAnyPermission": has,
	})
}

// handleAddRoles handles add roles 为用户添加角色
func handleAddRoles(c *gin.Context) {
	var req RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	if err := satoken.AddRoles(c.Request.Context(), req.LoginID, req.Roles); err != nil {
		fail(c, "添加角色失败: "+err.Error())
		return
	}

	success(c, "添加角色成功")
}

// handleAddRolesByToken handles add roles by token 根据 Token 添加角色
func handleAddRolesByToken(c *gin.Context) {
	var req struct {
		Roles []string `json:"roles" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	if err := satoken.AddRolesByToken(c.Request.Context(), tokenStr, req.Roles); err != nil {
		fail(c, "添加角色失败: "+err.Error())
		return
	}

	success(c, "添加角色成功")
}

// handleRemoveRoles handles remove roles 删除用户角色
func handleRemoveRoles(c *gin.Context) {
	var req RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	if err := satoken.RemoveRoles(c.Request.Context(), req.LoginID, req.Roles); err != nil {
		fail(c, "删除角色失败: "+err.Error())
		return
	}

	success(c, "删除角色成功")
}

// handleRemoveRolesByToken handles remove roles by token 根据 Token 删除角色
func handleRemoveRolesByToken(c *gin.Context) {
	var req struct {
		Roles []string `json:"roles" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	if err := satoken.RemoveRolesByToken(c.Request.Context(), tokenStr, req.Roles); err != nil {
		fail(c, "删除角色失败: "+err.Error())
		return
	}

	success(c, "删除角色成功")
}

// handleGetRoles handles get roles 获取用户角色列表
func handleGetRoles(c *gin.Context) {
	loginID := c.Param("loginId")
	if loginID == "" {
		fail(c, "loginId 不能为空")
		return
	}

	roles, err := satoken.GetRoles(c.Request.Context(), loginID)
	if err != nil {
		fail(c, "获取角色列表失败: "+err.Error())
		return
	}

	success(c, gin.H{
		"roles": roles,
	})
}

// handleGetRolesByToken handles get roles by token 根据 Token 获取角色列表
func handleGetRolesByToken(c *gin.Context) {
	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	roles, err := satoken.GetRolesByToken(c.Request.Context(), tokenStr)
	if err != nil {
		fail(c, "获取角色列表失败: "+err.Error())
		return
	}

	success(c, gin.H{
		"roles": roles,
	})
}

// handleHasRole handles has role 检查是否拥有指定角色
func handleHasRole(c *gin.Context) {
	var req struct {
		LoginID string `json:"loginId" binding:"required"`
		Role    string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	has := satoken.HasRole(c.Request.Context(), req.LoginID, req.Role)

	success(c, gin.H{
		"hasRole": has,
	})
}

// handleHasRoleByToken handles has role by token 根据 Token 检查角色
func handleHasRoleByToken(c *gin.Context) {
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	has := satoken.HasRoleByToken(c.Request.Context(), tokenStr, req.Role)

	success(c, gin.H{
		"hasRole": has,
	})
}

// handleHasRolesAnd checks all roles 检查是否拥有所有角色（AND逻辑）
func handleHasRolesAnd(c *gin.Context) {
	var req RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	has := satoken.HasRolesAnd(c.Request.Context(), req.LoginID, req.Roles)

	success(c, gin.H{
		"hasAllRoles": has,
	})
}

// handleHasRolesAndByToken handles has roles and by token 根据 Token 检查所有角色
func handleHasRolesAndByToken(c *gin.Context) {
	var req struct {
		Roles []string `json:"roles" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	has := satoken.HasRolesAndByToken(c.Request.Context(), tokenStr, req.Roles)

	success(c, gin.H{
		"hasAllRoles": has,
	})
}

// handleHasRolesOr checks any role 检查是否拥有任一角色（OR逻辑）
func handleHasRolesOr(c *gin.Context) {
	var req RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	has := satoken.HasRolesOr(c.Request.Context(), req.LoginID, req.Roles)

	success(c, gin.H{
		"hasAnyRole": has,
	})
}

// handleHasRolesOrByToken handles has roles or by token 根据 Token 检查任一角色
func handleHasRolesOrByToken(c *gin.Context) {
	var req struct {
		Roles []string `json:"roles" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	has := satoken.HasRolesOrByToken(c.Request.Context(), tokenStr, req.Roles)

	success(c, gin.H{
		"hasAnyRole": has,
	})
}

// handleGetSession handles get session 获取指定登录ID的会话
func handleGetSession(c *gin.Context) {
	loginID := c.Param("loginId")
	if loginID == "" {
		fail(c, "loginId 不能为空")
		return
	}

	session, err := satoken.GetSession(c.Request.Context(), loginID)
	if err != nil {
		fail(c, "获取会话失败: "+err.Error())
		return
	}

	success(c, session)
}

// handleGetSessionByToken handles get session by token 通过 Token 获取会话
func handleGetSessionByToken(c *gin.Context) {
	// Get Token From Context 从上下文获取 token
	token, _ := c.Get("token")
	tokenStr := token.(string)

	session, err := satoken.GetSessionByToken(c.Request.Context(), tokenStr)
	if err != nil {
		fail(c, "获取会话失败: "+err.Error())
		return
	}

	success(c, session)
}

// handleGetTokenValueListByLoginID handles get token value list by login id 获取指定登录ID的所有Token
func handleGetTokenValueListByLoginID(c *gin.Context) {
	loginID := c.Param("loginId")
	if loginID == "" {
		fail(c, "loginId 不能为空")
		return
	}

	checkAlive := c.DefaultQuery("checkAlive", "true") == "true"

	tokens, err := satoken.GetTokenValueListByLoginID(c.Request.Context(), loginID, checkAlive)
	if err != nil {
		fail(c, "获取Token列表失败: "+err.Error())
		return
	}

	success(c, gin.H{
		"tokens": tokens,
	})
}

// handleGetTokenValueListByDevice handles get token value list by device 获取指定设备类型的所有Token
func handleGetTokenValueListByDevice(c *gin.Context) {
	loginID := c.Param("loginId")
	device := c.Param("device")
	if loginID == "" || device == "" {
		fail(c, "loginId 和 device 不能为空")
		return
	}

	checkAlive := c.DefaultQuery("checkAlive", "true") == "true"

	tokens, err := satoken.GetTokenValueListByDevice(c.Request.Context(), loginID, device, checkAlive)
	if err != nil {
		fail(c, "获取Token列表失败: "+err.Error())
		return
	}

	success(c, gin.H{
		"tokens": tokens,
	})
}

// handleDisable handles disable 封禁账号指定时长
func handleDisable(c *gin.Context) {
	var req DisableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	if err := satoken.Disable(c.Request.Context(), req.LoginID, time.Duration(req.Duration)*time.Second, req.Reason); err != nil {
		fail(c, "封禁账号失败: "+err.Error())
		return
	}

	success(c, "封禁账号成功")
}

// handleUntie handles untie 解封账号
func handleUntie(c *gin.Context) {
	var req struct {
		LoginID string `json:"loginId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误: "+err.Error())
		return
	}

	if err := satoken.Untie(c.Request.Context(), req.LoginID); err != nil {
		fail(c, "解封账号失败: "+err.Error())
		return
	}

	success(c, "解封账号成功")
}

// handleIsDisable handles is disable 检查账号是否被封禁
func handleIsDisable(c *gin.Context) {
	loginID := c.Param("loginId")
	if loginID == "" {
		fail(c, "loginId 不能为空")
		return
	}

	isDisabled := satoken.IsDisable(c.Request.Context(), loginID)

	success(c, gin.H{
		"isDisabled": isDisabled,
	})
}

// handleGetDisableInfo handles get disable info 获取账号封禁信息
func handleGetDisableInfo(c *gin.Context) {
	loginID := c.Param("loginId")
	if loginID == "" {
		fail(c, "loginId 不能为空")
		return
	}

	info, err := satoken.GetDisableInfo(c.Request.Context(), loginID)
	if err != nil {
		fail(c, "获取封禁信息失败: "+err.Error())
		return
	}

	success(c, info)
}

// handleGetDisableTTL handles get disable ttl 获取账号剩余封禁时间
func handleGetDisableTTL(c *gin.Context) {
	loginID := c.Param("loginId")
	if loginID == "" {
		fail(c, "loginId 不能为空")
		return
	}

	ttl, err := satoken.GetDisableTTL(c.Request.Context(), loginID)
	if err != nil {
		fail(c, "获取封禁TTL失败: "+err.Error())
		return
	}

	success(c, gin.H{
		"ttl": ttl,
	})
}

// setupRoutes registers all routes 注册所有路由
func setupRoutes(r *gin.Engine) {
	// Root Path 根路径
	r.GET("/", func(c *gin.Context) {
		success(c, gin.H{
			"message": "SaToken Quick Start API",
			"version": "1.0.0",
		})
	})

	api := r.Group("/api")

	auth := api.Group("/auth")
	{
		// Public Endpoints 无需验证的接口
		auth.POST("/login", handleLogin)
		auth.POST("/login-by-token", handleLoginByToken)

		// Protected Endpoints 需要验证的接口
		auth.POST("/logout", authMiddleware(), handleLogout)
		auth.POST("/logout-by-device", authMiddleware(), handleLogoutByDevice)
		auth.POST("/logout-by-device-id", authMiddleware(), handleLogoutByDeviceAndDeviceId)
		auth.POST("/logout-by-login-id", authMiddleware(), handleLogoutByLoginID)
		auth.POST("/is-login", authMiddleware(), handleIsLogin)
		auth.POST("/check-login", authMiddleware(), handleCheckLogin)
		auth.POST("/get-login-id", authMiddleware(), handleGetLoginID)
		auth.POST("/get-token-info", authMiddleware(), handleGetTokenInfo)
		auth.POST("/get-device", authMiddleware(), handleGetDevice)
		auth.POST("/get-device-id", authMiddleware(), handleGetDeviceId)
		auth.POST("/get-token-create-time", authMiddleware(), handleGetTokenCreateTime)
		auth.POST("/get-token-ttl", authMiddleware(), handleGetTokenTTL)
		auth.GET("/online-count/:loginId", authMiddleware(), handleGetOnlineTerminalCount)
		auth.GET("/online-count/:loginId/:device", authMiddleware(), handleGetOnlineTerminalCountByDevice)
		auth.GET("/online-count/:loginId/:device/:deviceId", authMiddleware(), handleGetOnlineTerminalCountByDeviceAndDeviceId)
	}

	online := api.Group("/online", authMiddleware())
	{
		online.POST("/kickout", handleKickout)
		online.POST("/kickout-by-device", handleKickoutByDevice)
		online.POST("/kickout-by-device-id", handleKickoutByDeviceAndDeviceId)
		online.POST("/kickout-by-login-id", handleKickoutByLoginID)
		online.POST("/replace", handleReplace)
		online.POST("/replace-by-device", handleReplaceByDevice)
		online.POST("/replace-by-device-id", handleReplaceByDeviceAndDeviceId)
		online.POST("/replace-by-login-id", handleReplaceByLoginID)
	}

	permission := api.Group("/permission", authMiddleware())
	{
		permission.POST("/add", handleAddPermissions)
		permission.POST("/add-by-token", handleAddPermissionsByToken)
		permission.POST("/remove", handleRemovePermissions)
		permission.POST("/remove-by-token", handleRemovePermissionsByToken)
		permission.GET("/list/:loginId", handleGetPermissions)
		permission.POST("/list-by-token", handleGetPermissionsByToken)
		permission.POST("/has", handleHasPermission)
		permission.POST("/has-by-token", handleHasPermissionByToken)
		permission.POST("/has-and", handleHasPermissionsAnd)
		permission.POST("/has-and-by-token", handleHasPermissionsAndByToken)
		permission.POST("/has-or", handleHasPermissionsOr)
		permission.POST("/has-or-by-token", handleHasPermissionsOrByToken)
	}

	role := api.Group("/role", authMiddleware())
	{
		role.POST("/add", handleAddRoles)
		role.POST("/add-by-token", handleAddRolesByToken)
		role.POST("/remove", handleRemoveRoles)
		role.POST("/remove-by-token", handleRemoveRolesByToken)
		role.GET("/list/:loginId", handleGetRoles)
		role.POST("/list-by-token", handleGetRolesByToken)
		role.POST("/has", handleHasRole)
		role.POST("/has-by-token", handleHasRoleByToken)
		role.POST("/has-and", handleHasRolesAnd)
		role.POST("/has-and-by-token", handleHasRolesAndByToken)
		role.POST("/has-or", handleHasRolesOr)
		role.POST("/has-or-by-token", handleHasRolesOrByToken)
	}

	session := api.Group("/session", authMiddleware())
	{
		session.GET("/:loginId", handleGetSession)
		session.POST("/by-token", handleGetSessionByToken)
		session.GET("/tokens/:loginId", handleGetTokenValueListByLoginID)
		session.GET("/tokens/:loginId/:device", handleGetTokenValueListByDevice)
	}

	disable := api.Group("/disable", authMiddleware())
	{
		disable.POST("/ban", handleDisable)
		disable.POST("/unban", handleUntie)
		disable.GET("/is-disabled/:loginId", handleIsDisable)
		disable.GET("/info/:loginId", handleGetDisableInfo)
		disable.GET("/ttl/:loginId", handleGetDisableTTL)
	}

	fmt.Println("所有路由注册完成")
}

func main() {
	fmt.Println("========================================")
	fmt.Println("SaToken Quick Start - 完整测试示例")
	fmt.Println("========================================")

	// Initialize SaToken Framework 初始化 SaToken 框架
	if err := initSaToken(); err != nil {
		fmt.Printf("初始化失败: %v\n", err)
		return
	}

	// Create Gin Engine 创建 Gin 引擎
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Register All Routes 注册所有路由
	setupRoutes(r)

	// Start Server 启动服务器
	port := ":8080"
	fmt.Printf("\n服务器启动成功，监听端口: %s\n", port)
	fmt.Println("========================================")
	fmt.Println("API 接口分类:")
	fmt.Println("  - 认证接口: http://localhost:8080/api/auth/*")
	fmt.Println("  - 在线管理: http://localhost:8080/api/online/*")
	fmt.Println("  - 权限管理: http://localhost:8080/api/permission/*")
	fmt.Println("  - 角色管理: http://localhost:8080/api/role/*")
	fmt.Println("  - Session:  http://localhost:8080/api/session/*")
	fmt.Println("  - 账号封禁: http://localhost:8080/api/disable/*")
	fmt.Println("========================================")

	if err := r.Run(port); err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
	}
}
