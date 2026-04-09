package context

import (
	"context"
	"strings"
	"time"

	"github.com/click33/sa-token-go/core/adapter"
	"github.com/click33/sa-token-go/core/manager"
	"github.com/click33/sa-token-go/core/oauth2"
	"github.com/click33/sa-token-go/core/serror"
)

const (
	bearerPrefix = "Bearer "
	authHeader   = "Authorization"
)

// SaTokenContext defines current request context SaTokenContext 定义当前请求的 Sa-Token 上下文
type SaTokenContext struct {
	reqCtx  adapter.RequestContext
	manager *manager.Manager
}

// NewContext creates sa-token context NewContext 创建新的 SaTokenContext 上下文
func NewContext(reqCtx adapter.RequestContext, mgr *manager.Manager) *SaTokenContext {
	return &SaTokenContext{
		reqCtx:  reqCtx,
		manager: mgr,
	}
}

// GetTokenValue gets token from current request GetTokenValue 按 Header 到 Cookie 再到 Query 顺序获取 Token
func (c *SaTokenContext) GetTokenValue() string {
	cfg := c.manager.GetConfig()

	// Try header first 优先从 Header 读取 Token
	if cfg.IsReadHeader {
		// Try configured token header 优先读取配置的 Token Header
		if token := strings.TrimSpace(c.reqCtx.GetHeader(cfg.TokenName)); token != "" {
			return token
		}

		// Try authorization bearer token 其次尝试 Authorization Bearer Token
		if auth := c.reqCtx.GetHeader(authHeader); auth != "" {
			if token := extractBearerToken(auth); token != "" {
				return token
			}
		}
	}

	// Try cookie next 然后从 Cookie 读取 Token
	if cfg.IsReadCookie {
		if token := strings.TrimSpace(c.reqCtx.GetCookie(cfg.TokenName)); token != "" {
			return token
		}
	}

	// Try query finally 最后从 Query 参数读取 Token
	if token := strings.TrimSpace(c.reqCtx.GetQuery(cfg.TokenName)); token != "" {
		return token
	}

	return ""
}

// GetRequestContext returns raw request context GetRequestContext 获取原始请求上下文
func (c *SaTokenContext) GetRequestContext() adapter.RequestContext {
	return c.reqCtx
}

// GetManager returns related manager GetManager 获取关联的认证管理器
func (c *SaTokenContext) GetManager() *manager.Manager {
	return c.manager
}

// extractBearerToken extracts bearer token extractBearerToken 从 Authorization 头中提取 Bearer Token
func extractBearerToken(auth string) string {
	auth = strings.TrimSpace(auth)
	if auth == "" {
		return ""
	}

	// Check bearer prefix 检查 Bearer 前缀
	if len(auth) > 7 && strings.EqualFold(auth[:7], bearerPrefix) {
		return strings.TrimSpace(auth[7:])
	}

	// Return raw auth for compatibility 不符合 Bearer 格式时兼容返回原值
	return auth
}

// IsLogin checks login state IsLogin 检查当前 token 是否已登录
func (c *SaTokenContext) IsLogin(ctx context.Context) bool {
	token := c.GetTokenValue()
	if token == "" {
		return false
	}
	return c.manager.IsLogin(ctx, token)
}

// CheckLogin checks login state with error CheckLogin 检查当前 token 是否已登录并在未登录时返回错误
func (c *SaTokenContext) CheckLogin(ctx context.Context) error {
	token := c.GetTokenValue()
	if token == "" {
		return serror.ErrNotLogin
	}
	return c.manager.CheckLogin(ctx, token)
}

// GetLoginID gets login id by token GetLoginID 获取当前 token 关联的登录 ID
func (c *SaTokenContext) GetLoginID(ctx context.Context) (string, error) {
	token := c.GetTokenValue()
	if token == "" {
		return "", serror.ErrNotLogin
	}
	return c.manager.GetLoginID(ctx, token)
}

// LoginByToken logs in by current token LoginByToken 使用当前 token 登录
func (c *SaTokenContext) LoginByToken(ctx context.Context) error {
	token := c.GetTokenValue()
	if token == "" {
		return serror.ErrNotLogin
	}
	return c.manager.LoginByToken(ctx, token)
}

// Logout logs out current token Logout 登出当前 token
func (c *SaTokenContext) Logout(ctx context.Context) error {
	token := c.GetTokenValue()
	if token == "" {
		return serror.ErrNotLogin
	}
	return c.manager.Logout(ctx, token)
}

// GetTokenInfo gets token info GetTokenInfo 获取当前 token 的信息
func (c *SaTokenContext) GetTokenInfo(ctx context.Context) (*manager.TokenInfo, error) {
	token := c.GetTokenValue()
	if token == "" {
		return nil, serror.ErrNotLogin
	}
	return c.manager.GetTokenInfo(ctx, token)
}

// GetDevice gets token device GetDevice 获取当前 token 的设备类型
func (c *SaTokenContext) GetDevice(ctx context.Context) (string, error) {
	token := c.GetTokenValue()
	if token == "" {
		return "", serror.ErrNotLogin
	}
	return c.manager.GetDevice(ctx, token)
}

// GetDeviceId gets token device id GetDeviceId 获取当前 token 的设备 ID
func (c *SaTokenContext) GetDeviceId(ctx context.Context) (string, error) {
	token := c.GetTokenValue()
	if token == "" {
		return "", serror.ErrNotLogin
	}
	return c.manager.GetDeviceId(ctx, token)
}

// GetTokenCreateTime gets token create time GetTokenCreateTime 获取当前 token 的创建时间
func (c *SaTokenContext) GetTokenCreateTime(ctx context.Context) (int64, error) {
	token := c.GetTokenValue()
	if token == "" {
		return 0, serror.ErrNotLogin
	}
	return c.manager.GetTokenCreateTime(ctx, token)
}

// GetTokenTTL gets token ttl GetTokenTTL 获取当前 token 的剩余有效期
func (c *SaTokenContext) GetTokenTTL(ctx context.Context) (int64, error) {
	token := c.GetTokenValue()
	if token == "" {
		return 0, serror.ErrNotLogin
	}
	return c.manager.GetTokenTTL(ctx, token)
}

// Kickout kicks out current token Kickout 踢出当前 token
func (c *SaTokenContext) Kickout(ctx context.Context) error {
	token := c.GetTokenValue()
	if token == "" {
		return serror.ErrNotLogin
	}
	return c.manager.Kickout(ctx, token)
}

// Replace replaces current token Replace 顶替当前 token
func (c *SaTokenContext) Replace(ctx context.Context) error {
	token := c.GetTokenValue()
	if token == "" {
		return serror.ErrNotLogin
	}
	return c.manager.Replace(ctx, token)
}

// LogoutByDevice logs out current user by device LogoutByDevice 按设备登出当前用户
func (c *SaTokenContext) LogoutByDevice(ctx context.Context, device string) error {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return err
	}
	return c.manager.LogoutByDevice(ctx, loginID, device)
}

// LogoutByDeviceAndDeviceId logs out current user by device and id LogoutByDeviceAndDeviceId 按设备和设备 ID 登出当前用户
func (c *SaTokenContext) LogoutByDeviceAndDeviceId(ctx context.Context, deviceAndDeviceId ...string) error {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return err
	}
	return c.manager.LogoutByDeviceAndDeviceId(ctx, loginID, deviceAndDeviceId...)
}

// KickoutByDevice kicks out current user by device KickoutByDevice 按设备踢出当前用户
func (c *SaTokenContext) KickoutByDevice(ctx context.Context, device string) error {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return err
	}
	return c.manager.KickoutByDevice(ctx, loginID, device)
}

// KickoutByDeviceAndDeviceId kicks out current user by device and id KickoutByDeviceAndDeviceId 按设备和设备 ID 踢出当前用户
func (c *SaTokenContext) KickoutByDeviceAndDeviceId(ctx context.Context, deviceAndDeviceId ...string) error {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return err
	}
	return c.manager.KickoutByDeviceAndDeviceId(ctx, loginID, deviceAndDeviceId...)
}

// ReplaceByDevice replaces current user by device ReplaceByDevice 按设备顶替当前用户
func (c *SaTokenContext) ReplaceByDevice(ctx context.Context, device string) error {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return err
	}
	return c.manager.ReplaceByDevice(ctx, loginID, device)
}

// ReplaceByDeviceAndDeviceId replaces current user by device and id ReplaceByDeviceAndDeviceId 按设备和设备 ID 顶替当前用户
func (c *SaTokenContext) ReplaceByDeviceAndDeviceId(ctx context.Context, deviceAndDeviceId ...string) error {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return err
	}
	return c.manager.ReplaceByDeviceAndDeviceId(ctx, loginID, deviceAndDeviceId...)
}

// LogoutByLoginID logs out all terminals LogoutByLoginID 登出当前用户的所有终端
func (c *SaTokenContext) LogoutByLoginID(ctx context.Context) error {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return err
	}
	return c.manager.LogoutByLoginID(ctx, loginID)
}

// KickoutByLoginID kicks out all terminals KickoutByLoginID 踢出当前用户的所有终端
func (c *SaTokenContext) KickoutByLoginID(ctx context.Context) error {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return err
	}
	return c.manager.KickoutByLoginID(ctx, loginID)
}

// ReplaceByLoginID replaces all terminals ReplaceByLoginID 顶替当前用户的所有终端
func (c *SaTokenContext) ReplaceByLoginID(ctx context.Context) error {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return err
	}
	return c.manager.ReplaceByLoginID(ctx, loginID)
}

// GetTokenValueList gets token list GetTokenValueList 获取当前登录用户的全部 token 列表
func (c *SaTokenContext) GetTokenValueList(ctx context.Context, checkAlive ...bool) ([]string, error) {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return nil, err
	}
	return c.manager.GetTokenValueListByLoginID(ctx, loginID, checkAlive...)
}

// GetTokenValueListByDevice gets token list by device GetTokenValueListByDevice 按设备获取当前用户的 token 列表
func (c *SaTokenContext) GetTokenValueListByDevice(ctx context.Context, device string, checkAlive ...bool) ([]string, error) {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return nil, err
	}
	return c.manager.GetTokenValueListByDevice(ctx, loginID, device, checkAlive...)
}

// GetTokenValueListByDeviceAndDeviceId gets token list by device and id GetTokenValueListByDeviceAndDeviceId 按设备和设备 ID 获取当前用户的 token 列表
func (c *SaTokenContext) GetTokenValueListByDeviceAndDeviceId(ctx context.Context, device, deviceId string, checkAlive ...bool) ([]string, error) {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return nil, err
	}
	return c.manager.GetTokenValueListByDeviceAndDeviceId(ctx, loginID, device, deviceId, checkAlive...)
}

// GetOnlineTerminalCount gets online terminal count GetOnlineTerminalCount 获取当前用户的在线终端数量
func (c *SaTokenContext) GetOnlineTerminalCount(ctx context.Context) (int, error) {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return 0, err
	}
	return c.manager.GetOnlineTerminalCount(ctx, loginID)
}

// GetOnlineTerminalCountByDevice gets online terminal count by device GetOnlineTerminalCountByDevice 按设备获取当前用户的在线终端数量
func (c *SaTokenContext) GetOnlineTerminalCountByDevice(ctx context.Context, device string) (int, error) {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return 0, err
	}
	return c.manager.GetOnlineTerminalCountByDevice(ctx, loginID, device)
}

// GetOnlineTerminalCountByDeviceAndDeviceId gets online terminal count by device and id GetOnlineTerminalCountByDeviceAndDeviceId 按设备和设备 ID 获取当前用户的在线终端数量
func (c *SaTokenContext) GetOnlineTerminalCountByDeviceAndDeviceId(ctx context.Context, device, deviceId string) (int, error) {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return 0, err
	}
	return c.manager.GetOnlineTerminalCountByDeviceAndDeviceId(ctx, loginID, device, deviceId)
}

// GetRoles gets role list GetRoles 获取当前登录用户的角色列表
func (c *SaTokenContext) GetRoles(ctx context.Context) ([]string, error) {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return nil, err
	}
	return c.manager.GetRoles(ctx, loginID)
}

// GetRolesByToken gets role list by token GetRolesByToken 使用当前 token 获取角色列表
func (c *SaTokenContext) GetRolesByToken(ctx context.Context) ([]string, error) {
	token := c.GetTokenValue()
	if token == "" {
		return nil, serror.ErrNotLogin
	}
	return c.manager.GetRolesByToken(ctx, token)
}

// HasRole checks role HasRole 检查当前用户是否拥有指定角色
func (c *SaTokenContext) HasRole(ctx context.Context, role string) bool {
	token := c.GetTokenValue()
	if token == "" {
		return false
	}
	return c.manager.HasRoleByToken(ctx, token, role)
}

// HasRoles checks any role HasRoles 检查当前用户是否拥有任意指定角色
func (c *SaTokenContext) HasRoles(ctx context.Context, roles []string) bool {
	token := c.GetTokenValue()
	if token == "" {
		return false
	}
	return c.manager.HasRolesOrByToken(ctx, token, roles)
}

// HasRolesAnd checks all roles HasRolesAnd 检查当前用户是否拥有全部指定角色
func (c *SaTokenContext) HasRolesAnd(ctx context.Context, roles []string) bool {
	token := c.GetTokenValue()
	if token == "" {
		return false
	}
	return c.manager.HasRolesAndByToken(ctx, token, roles)
}

// AddRoles adds roles AddRoles 为当前用户添加角色
func (c *SaTokenContext) AddRoles(ctx context.Context, roles []string) error {
	token := c.GetTokenValue()
	if token == "" {
		return serror.ErrNotLogin
	}
	return c.manager.AddRolesByToken(ctx, token, roles)
}

// RemoveRoles removes roles RemoveRoles 从当前用户移除角色
func (c *SaTokenContext) RemoveRoles(ctx context.Context, roles []string) error {
	token := c.GetTokenValue()
	if token == "" {
		return serror.ErrNotLogin
	}
	return c.manager.RemoveRolesByToken(ctx, token, roles)
}

// GetPermissions gets permission list GetPermissions 获取当前登录用户的权限列表
func (c *SaTokenContext) GetPermissions(ctx context.Context) ([]string, error) {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return nil, err
	}
	return c.manager.GetPermissions(ctx, loginID)
}

// GetPermissionsByToken gets permission list by token GetPermissionsByToken 使用当前 token 获取权限列表
func (c *SaTokenContext) GetPermissionsByToken(ctx context.Context) ([]string, error) {
	token := c.GetTokenValue()
	if token == "" {
		return nil, serror.ErrNotLogin
	}
	return c.manager.GetPermissionsByToken(ctx, token)
}

// HasPermission checks permission HasPermission 检查当前用户是否拥有指定权限
func (c *SaTokenContext) HasPermission(ctx context.Context, permission string) bool {
	token := c.GetTokenValue()
	if token == "" {
		return false
	}
	return c.manager.HasPermissionByToken(ctx, token, permission)
}

// HasPermissions checks any permission HasPermissions 检查当前用户是否拥有任意指定权限
func (c *SaTokenContext) HasPermissions(ctx context.Context, permissions []string) bool {
	token := c.GetTokenValue()
	if token == "" {
		return false
	}
	return c.manager.HasPermissionsOrByToken(ctx, token, permissions)
}

// HasPermissionsAnd checks all permissions HasPermissionsAnd 检查当前用户是否拥有全部指定权限
func (c *SaTokenContext) HasPermissionsAnd(ctx context.Context, permissions []string) bool {
	token := c.GetTokenValue()
	if token == "" {
		return false
	}
	return c.manager.HasPermissionsAndByToken(ctx, token, permissions)
}

// AddPermissions adds permissions AddPermissions 为当前用户添加权限
func (c *SaTokenContext) AddPermissions(ctx context.Context, permissions []string) error {
	token := c.GetTokenValue()
	if token == "" {
		return serror.ErrNotLogin
	}
	return c.manager.AddPermissionsByToken(ctx, token, permissions)
}

// RemovePermissions removes permissions RemovePermissions 从当前用户移除权限
func (c *SaTokenContext) RemovePermissions(ctx context.Context, permissions []string) error {
	token := c.GetTokenValue()
	if token == "" {
		return serror.ErrNotLogin
	}
	return c.manager.RemovePermissionsByToken(ctx, token, permissions)
}

// IsDisable checks disable state IsDisable 检查当前用户是否被封禁
func (c *SaTokenContext) IsDisable(ctx context.Context) bool {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return false
	}
	return c.manager.IsDisable(ctx, loginID)
}

// GetDisableInfo gets disable info GetDisableInfo 获取当前用户的封禁信息
func (c *SaTokenContext) GetDisableInfo(ctx context.Context) (*manager.DisableInfo, error) {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return nil, err
	}
	return c.manager.GetDisableInfo(ctx, loginID)
}

// GetDisableTTL gets disable ttl GetDisableTTL 获取当前用户的封禁剩余时间
func (c *SaTokenContext) GetDisableTTL(ctx context.Context) (int64, error) {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return 0, err
	}
	return c.manager.GetDisableTTL(ctx, loginID)
}

// Disable disables current user Disable 封禁当前用户账号指定时长
func (c *SaTokenContext) Disable(ctx context.Context, duration time.Duration, reason ...string) error {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return err
	}
	return c.manager.Disable(ctx, loginID, duration, reason...)
}

// Untie removes disable state Untie 解封当前用户账号
func (c *SaTokenContext) Untie(ctx context.Context) error {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return err
	}
	return c.manager.Untie(ctx, loginID)
}

// GetSession gets session by login id GetSession 获取当前登录用户的 Session
func (c *SaTokenContext) GetSession(ctx context.Context) (*manager.Session, error) {
	loginID, err := c.GetLoginID(ctx)
	if err != nil {
		return nil, err
	}
	return c.manager.GetSession(ctx, loginID)
}

// GetSessionByToken gets session by token GetSessionByToken 使用当前 token 获取 Session
func (c *SaTokenContext) GetSessionByToken(ctx context.Context) (*manager.Session, error) {
	token := c.GetTokenValue()
	if token == "" {
		return nil, serror.ErrNotLogin
	}
	return c.manager.GetSessionByToken(ctx, token)
}

// GenerateNonce generates nonce GenerateNonce 生成新的 nonce
func (c *SaTokenContext) GenerateNonce(ctx context.Context) (string, error) {
	return c.manager.GenerateNonce(ctx)
}

// VerifyNonce verifies and consumes nonce VerifyNonce 验证并消费 nonce
func (c *SaTokenContext) VerifyNonce(ctx context.Context, nonce string) bool {
	return c.manager.VerifyNonce(ctx, nonce)
}

// VerifyAndConsumeNonce verifies nonce with error VerifyAndConsumeNonce 验证并消费 nonce 且在无效时返回错误
func (c *SaTokenContext) VerifyAndConsumeNonce(ctx context.Context, nonce string) error {
	return c.manager.VerifyAndConsumeNonce(ctx, nonce)
}

// IsNonceValid checks nonce validity IsNonceValid 检查 nonce 是否有效且不消费
func (c *SaTokenContext) IsNonceValid(ctx context.Context, nonce string) bool {
	return c.manager.IsNonceValid(ctx, nonce)
}

// ValidateOAuth2AccessToken validates access token ValidateOAuth2AccessToken 验证 OAuth2 访问令牌
func (c *SaTokenContext) ValidateOAuth2AccessToken(ctx context.Context, accessToken string) bool {
	return c.manager.ValidateOAuth2AccessToken(ctx, accessToken)
}

// ValidateOAuth2AccessTokenAndGetInfo validates access token and gets info ValidateOAuth2AccessTokenAndGetInfo 验证 OAuth2 访问令牌并获取信息
func (c *SaTokenContext) ValidateOAuth2AccessTokenAndGetInfo(ctx context.Context, accessToken string) (*oauth2.AccessToken, error) {
	return c.manager.ValidateOAuth2AccessTokenAndGetInfo(ctx, accessToken)
}

// RevokeOAuth2Token revokes oauth2 token RevokeOAuth2Token 撤销 OAuth2 访问令牌及其刷新令牌
func (c *SaTokenContext) RevokeOAuth2Token(ctx context.Context, accessToken string) error {
	return c.manager.RevokeOAuth2Token(ctx, accessToken)
}
