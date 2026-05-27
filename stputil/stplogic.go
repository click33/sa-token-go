package stputil

import (
	"fmt"
	"sync"
	"time"

	core "github.com/sa-tokens/sa-token-go/core"
	"github.com/sa-tokens/sa-token-go/core/adapter"
	"github.com/sa-tokens/sa-token-go/core/manager"
	"github.com/sa-tokens/sa-token-go/core/oauth2"
	"github.com/sa-tokens/sa-token-go/core/security"
	"github.com/sa-tokens/sa-token-go/core/session"
)

var (
	TokenValueKey  = "stplogic:tokenvalue"
	LoginIdKey     = "stplogic:loginid"
	PermissionsKey = "stplogic:permissions"
	RolesKey       = "stplogic:roles"
)

type StpLogic struct {
	manager *manager.Manager
	mu      sync.RWMutex
}

func NewStpLogic(mrg *manager.Manager) *StpLogic {
	return &StpLogic{manager: mrg}
}

// GetManager gets the global Manager | 获取全局Manager
func (s *StpLogic) GetManager() *manager.Manager {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.manager == nil {
		panic("StpLogic not initialized.")
	}
	return s.manager
}

func (s *StpLogic) SetManager(manager *manager.Manager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.manager = manager
}

// ============ Authentication | 登录认证 ============

// Login performs user login | 用户登录
func (s *StpLogic) Login(loginID interface{}, device ...string) (string, error) {
	return s.manager.Login(toString(loginID), device...)
}

// LoginByToken performs login with specified token | 使用指定Token登录
func (s *StpLogic) LoginByToken(loginID interface{}, tokenValue string, device ...string) error {
	return s.manager.LoginByToken(toString(loginID), tokenValue, device...)
}

// Logout performs user logout | 用户登出
func (s *StpLogic) Logout(loginID interface{}, device ...string) error {
	return s.manager.Logout(toString(loginID), device...)
}

// LogoutByToken performs logout by token | 根据Token登出
func (s *StpLogic) LogoutByToken(tokenValue string) error {
	return s.manager.LogoutByToken(tokenValue)
}

// IsLogin checks if the user is logged in | 检查用户是否已登录
func (s *StpLogic) IsLogin(tokenValue string) bool {
	return s.manager.IsLogin(tokenValue)
}

// CheckLogin checks login status (throws error if not logged in) | 检查登录状态（未登录抛出错误）
func (s *StpLogic) CheckLogin(tokenValue string) error {
	return s.manager.CheckLogin(tokenValue)
}

// GetLoginID gets the login ID from token | 从Token获取登录ID
func (s *StpLogic) GetLoginID(tokenValue string) (string, error) {
	return s.manager.GetLoginID(tokenValue)
}

// GetLoginIDNotCheck gets login ID without checking | 获取登录ID（不检查）
func (s *StpLogic) GetLoginIDNotCheck(tokenValue string) (string, error) {
	return s.manager.GetLoginIDNotCheck(tokenValue)
}

// GetTokenValue gets the token value for a login ID | 获取登录ID对应的Token值
func (s *StpLogic) GetTokenValue(loginID interface{}, device ...string) (string, error) {
	return s.manager.GetTokenValue(toString(loginID), device...)
}

// GetTokenInfo gets token information | 获取Token信息
func (s *StpLogic) GetTokenInfo(tokenValue string) (*manager.TokenInfo, error) {
	return s.manager.GetTokenInfo(tokenValue)
}

// ============ Kickout | 踢人下线 ============

// Kickout kicks out a user session | 踢人下线
func (s *StpLogic) Kickout(loginID interface{}, device ...string) error {
	return s.manager.Kickout(toString(loginID), device...)
}

// KickoutByToken kicks out by token value | 按 Token 踢下线
func (s *StpLogic) KickoutByToken(tokenValue string) error {
	return s.manager.KickoutByToken(tokenValue)
}

// ============ Account Disable | 账号封禁 ============

// Disable disables an account for specified duration | 封禁账号（指定时长）
func (s *StpLogic) Disable(loginID interface{}, duration time.Duration) error {
	return s.manager.Disable(toString(loginID), duration)
}

// Untie re-enables a disabled account | 解封账号
func (s *StpLogic) Untie(loginID interface{}) error {
	return s.manager.Untie(toString(loginID))
}

// IsDisable checks if an account is disabled | 检查账号是否被封禁
func (s *StpLogic) IsDisable(loginID interface{}) bool {
	return s.manager.IsDisable(toString(loginID))
}

// GetDisableTime gets remaining disable time in seconds | 获取剩余封禁时间（秒）
func (s *StpLogic) GetDisableTime(loginID interface{}) (int64, error) {
	return s.manager.GetDisableTime(toString(loginID))
}

// ============ Session Management | Session管理 ============

// GetSession gets session by login ID | 根据登录ID获取Session
func (s *StpLogic) GetSession(loginID interface{}) (*session.Session, error) {
	return s.manager.GetSession(toString(loginID))
}

// GetSessionByToken gets session by token | 根据Token获取Session
func (s *StpLogic) GetSessionByToken(tokenValue string) (*session.Session, error) {
	return s.manager.GetSessionByToken(tokenValue)
}

// DeleteSession deletes a session | 删除Session
func (s *StpLogic) DeleteSession(loginID interface{}) error {
	return s.manager.DeleteSession(toString(loginID))
}

// ============ Permission Verification | 权限验证 ============

// SetPermissions sets permissions for a login ID | 设置用户权限
func (s *StpLogic) SetPermissions(loginID interface{}, permissions []string) error {
	return s.manager.SetPermissions(toString(loginID), permissions)
}

// GetPermissions gets permission list | 获取权限列表
func (s *StpLogic) GetPermissions(loginID interface{}) ([]string, error) {
	return s.manager.GetPermissions(toString(loginID))
}

// HasPermission checks if has specified permission | 检查是否拥有指定权限
func (s *StpLogic) HasPermission(loginID interface{}, permission string) bool {
	return s.manager.HasPermission(toString(loginID), permission)
}

// HasPermissionsAnd checks if has all permissions (AND logic) | 检查是否拥有所有权限（AND逻辑）
func (s *StpLogic) HasPermissionsAnd(loginID interface{}, permissions []string) bool {
	return s.manager.HasPermissionsAnd(toString(loginID), permissions)
}

// HasPermissionsOr checks if has any permission (OR logic) | 检查是否拥有任一权限（OR逻辑）
func (s *StpLogic) HasPermissionsOr(loginID interface{}, permissions []string) bool {
	return s.manager.HasPermissionsOr(toString(loginID), permissions)
}

// ============ Role Management | 角色管理 ============

// SetRoles sets roles for a login ID | 设置用户角色
func (s *StpLogic) SetRoles(loginID interface{}, roles []string) error {
	return s.manager.SetRoles(toString(loginID), roles)
}

// GetRoles gets role list | 获取角色列表
func (s *StpLogic) GetRoles(loginID interface{}) ([]string, error) {
	return s.manager.GetRoles(toString(loginID))
}

// HasRole checks if has specified role | 检查是否拥有指定角色
func (s *StpLogic) HasRole(loginID interface{}, role string) bool {
	return s.manager.HasRole(toString(loginID), role)
}

// HasRolesAnd checks if has all roles (AND logic) | 检查是否拥有所有角色（AND逻辑）
func (s *StpLogic) HasRolesAnd(loginID interface{}, roles []string) bool {
	return s.manager.HasRolesAnd(toString(loginID), roles)
}

// HasRolesOr 检查是否拥有任一角色（OR）
func (s *StpLogic) HasRolesOr(loginID interface{}, roles []string) bool {
	return s.manager.HasRolesOr(toString(loginID), roles)
}

// ============ Token标签 ============

// SetTokenTag 设置Token标签
func (s *StpLogic) SetTokenTag(tokenValue, tag string) error {
	return s.manager.SetTokenTag(tokenValue, tag)
}

// GetTokenTag 获取Token标签
func (s *StpLogic) GetTokenTag(tokenValue string) (string, error) {
	return s.manager.GetTokenTag(tokenValue)
}

// ============ 会话查询 ============

// GetTokenValueList 获取指定账号的所有Token
func (s *StpLogic) GetTokenValueList(loginID interface{}) ([]string, error) {
	return s.manager.GetTokenValueListByLoginID(toString(loginID))
}

// GetSessionCount 获取指定账号的Session数量
func (s *StpLogic) GetSessionCount(loginID interface{}) (int, error) {
	return s.manager.GetSessionCountByLoginID(toString(loginID))
}

func (s *StpLogic) GenerateNonce() (string, error) {
	if s.manager == nil {
		panic("Manager not initialized.")
	}
	return s.manager.GenerateNonce()
}

func (s *StpLogic) VerifyNonce(nonce string) bool {
	if s.manager == nil {
		panic("Manager not initialized.")
	}
	return s.manager.VerifyNonce(nonce)
}

func (s *StpLogic) LoginWithRefreshToken(loginID interface{}, device ...string) (*security.RefreshTokenInfo, error) {
	if s.manager == nil {
		panic("Manager not initialized.")
	}
	deviceType := "default"
	if len(device) > 0 {
		deviceType = device[0]
	}
	return s.manager.LoginWithRefreshToken(fmt.Sprintf("%v", loginID), deviceType)
}

func (s *StpLogic) RefreshAccessToken(refreshToken string) (*security.RefreshTokenInfo, error) {
	if s.manager == nil {
		panic("Manager not initialized.")
	}
	return s.manager.RefreshAccessToken(refreshToken)
}

func (s *StpLogic) RevokeRefreshToken(refreshToken string) error {
	if s.manager == nil {
		panic("Manager not initialized.")
	}
	return s.manager.RevokeRefreshToken(refreshToken)
}

func (s *StpLogic) GetOAuth2Server() *oauth2.OAuth2Server {
	if s.manager == nil {
		panic("Manager not initialized.")
	}
	return s.manager.GetOAuth2Server()
}

// ============ Check Functions for Token-based operations | 基于Token的检查函数 ============

// CheckDisable checks if the account associated with the token is disabled | 检查Token对应账号是否被封禁
func (s *StpLogic) CheckDisable(tokenValue string) error {
	loginID, err := s.GetLoginID(tokenValue)
	if err != nil {
		return err
	}
	if s.IsDisable(loginID) {
		return core.NewAccountDisabledError(loginID)
	}
	return nil
}

// CheckPermission checks if the token has the specified permission | 检查Token是否拥有指定权限
func (s *StpLogic) CheckPermission(tokenValue string, permission string) error {
	loginID, err := s.GetLoginID(tokenValue)
	if err != nil {
		return err
	}
	if !s.HasPermission(loginID, permission) {
		return core.NewPermissionDeniedError(permission)
	}
	return nil
}

// CheckPermissionAnd checks if the token has all specified permissions | 检查Token是否拥有所有指定权限
func (s *StpLogic) CheckPermissionAnd(tokenValue string, permissions []string) error {
	loginID, err := s.GetLoginID(tokenValue)
	if err != nil {
		return err
	}
	if !s.HasPermissionsAnd(loginID, permissions) {
		return core.NewPermissionDeniedListError(permissions)
	}
	return nil
}

// CheckPermissionOr checks if the token has any of the specified permissions | 检查Token是否拥有任一指定权限
func (s *StpLogic) CheckPermissionOr(tokenValue string, permissions []string) error {
	loginID, err := s.GetLoginID(tokenValue)
	if err != nil {
		return err
	}
	if !s.HasPermissionsOr(loginID, permissions) {
		return core.NewPermissionDeniedListError(permissions)
	}
	return nil
}

// GetPermissionList gets permission list for the token | 获取Token对应的权限列表
func (s *StpLogic) GetPermissionList(tokenValue string) ([]string, error) {
	loginID, err := s.GetLoginID(tokenValue)
	if err != nil {
		return nil, err
	}
	return s.GetPermissions(loginID)
}

// CheckRole checks if the token has the specified role | 检查Token是否拥有指定角色
func (s *StpLogic) CheckRole(tokenValue string, role string) error {
	loginID, err := s.GetLoginID(tokenValue)
	if err != nil {
		return err
	}
	if !s.HasRole(loginID, role) {
		return core.NewRoleDeniedError(role)
	}
	return nil
}

// CheckRoleAnd checks if the token has all specified roles | 检查Token是否拥有所有指定角色
func (s *StpLogic) CheckRoleAnd(tokenValue string, roles []string) error {
	loginID, err := s.GetLoginID(tokenValue)
	if err != nil {
		return err
	}
	if !s.HasRolesAnd(loginID, roles) {
		return core.NewRoleDeniedListError(roles)
	}
	return nil
}

// CheckRoleOr checks if the token has any of the specified roles | 检查Token是否拥有任一指定角色
func (s *StpLogic) CheckRoleOr(tokenValue string, roles []string) error {
	loginID, err := s.GetLoginID(tokenValue)
	if err != nil {
		return err
	}
	if !s.HasRolesOr(loginID, roles) {
		return core.NewRoleDeniedListError(roles)
	}
	return nil
}

// GetRoleList gets role list for the token | 获取Token对应的角色列表
func (s *StpLogic) GetRoleList(tokenValue string) ([]string, error) {
	loginID, err := GetLoginID(tokenValue)
	if err != nil {
		return nil, err
	}
	return GetRoles(loginID)
}

// GetTokenSession loads Token-Session (not Account-Session) | 获取 Token-Session
func (s *StpLogic) GetTokenSession(tokenValue string) (*session.Session, error) {
	return s.manager.GetTokenSession(tokenValue, false)
}

// GetTokenSessionOrCreate loads or creates Token-Session | 获取或创建 Token-Session
func (s *StpLogic) GetTokenSessionOrCreate(tokenValue string) (*session.Session, error) {
	return s.manager.GetTokenSession(tokenValue, true)
}

// GetAnonTokenSession anonymous Token-Session | 匿名 Token-Session
func (s *StpLogic) GetAnonTokenSession(currentToken string, ctx adapter.RequestContext) (*session.Session, error) {
	return s.manager.GetAnonTokenSession(currentToken, ctx)
}

// DeleteTokenSession removes Token-Session storage | 删除 Token-Session
func (s *StpLogic) DeleteTokenSession(tokenValue string) error {
	return s.manager.DeleteTokenSession(tokenValue)
}

// Replaced overrun logout by loginId+device | 顶号下线
func (s *StpLogic) Replaced(loginID interface{}, device ...string) error {
	return s.manager.Replaced(toString(loginID), device...)
}

// ReplacedByToken overrun logout by token | 按 Token 顶号下线
func (s *StpLogic) ReplacedByToken(tokenValue string) error {
	return s.manager.ReplacedByToken(tokenValue)
}

// OpenSafe second-level auth | 二级认证
func (s *StpLogic) OpenSafe(tokenValue, service string, safeTime int64) error {
	return s.manager.OpenSafe(tokenValue, service, safeTime)
}

func (s *StpLogic) IsSafe(tokenValue, service string) bool {
	return s.manager.IsSafe(tokenValue, service)
}

func (s *StpLogic) CheckSafe(tokenValue, service string) error {
	return s.manager.CheckSafe(tokenValue, service)
}

func (s *StpLogic) GetSafeTime(tokenValue, service string) (int64, error) {
	return s.manager.GetSafeTime(tokenValue, service)
}

func (s *StpLogic) CloseSafe(tokenValue, service string) error {
	return s.manager.CloseSafe(tokenValue, service)
}

func (s *StpLogic) DisableLevel(loginID interface{}, service string, level int, duration time.Duration) error {
	return s.manager.DisableLevel(toString(loginID), service, level, duration)
}

func (s *StpLogic) IsDisableLevel(loginID interface{}, service string, level int) bool {
	return s.manager.IsDisableLevel(toString(loginID), service, level)
}

func (s *StpLogic) CheckDisableLevel(loginID interface{}, service string, level int) error {
	return s.manager.CheckDisableLevel(toString(loginID), service, level)
}

func (s *StpLogic) GetDisableLevel(loginID interface{}, service string) int {
	return s.manager.GetDisableLevel(toString(loginID), service)
}

func (s *StpLogic) UntieDisable(loginID interface{}, services ...string) error {
	return s.manager.UntieDisableServices(toString(loginID), services...)
}

// LoginRememberMe performs login with remember-me mode | 记住我模式登录
func (s *StpLogic) LoginRememberMe(loginID interface{}, device ...string) (string, error) {
	return s.manager.LoginRememberMe(toString(loginID), device...)
}

// IsRememberMeLogin checks if token was created with remember-me | 检查是否为记住我模式登录
func (s *StpLogic) IsRememberMeLogin(tokenValue string) (bool, error) {
	return s.manager.IsRememberMeLogin(tokenValue)
}

func (s *StpLogic) UpdateLastActiveToNow(tokenValue string) error {
	return s.manager.UpdateLastActiveToNow(tokenValue)
}

// ============ API Key | API Key 管理 ============

func (s *StpLogic) CreateApiKey(loginID, title string, expireSeconds int64, extra string) (*security.ApiKeyInfo, error) {
	return s.manager.CreateApiKey(loginID, title, expireSeconds, extra)
}

func (s *StpLogic) GetApiKeyInfo(key string) (*security.ApiKeyInfo, error) {
	return s.manager.GetApiKeyInfo(key)
}

func (s *StpLogic) VerifyApiKey(key string) (*security.ApiKeyInfo, error) {
	return s.manager.VerifyApiKey(key)
}

func (s *StpLogic) DeleteApiKey(key string) error {
	return s.manager.DeleteApiKey(key)
}

func (s *StpLogic) DisableApiKey(key string) error {
	return s.manager.DisableApiKey(key)
}

func (s *StpLogic) EnableApiKey(key string) error {
	return s.manager.EnableApiKey(key)
}

// ============ Signature | 参数签名 ============

func (s *StpLogic) Sign(params map[string]string, secret string) string {
	return s.manager.Sign(params, secret)
}

func (s *StpLogic) VerifySign(params map[string]string, secret, timestamp, nonce, signature string, maxAgeSeconds int64) error {
	return s.manager.VerifySign(params, secret, timestamp, nonce, signature, maxAgeSeconds)
}

// ============ Temp Token | 临时Token ============

func (s *StpLogic) CreateTempToken(loginID string, expireSeconds int64, extra string) (*security.TempTokenInfo, error) {
	return s.manager.CreateTempToken(loginID, expireSeconds, extra)
}

func (s *StpLogic) VerifyTempToken(token string) (*security.TempTokenInfo, error) {
	return s.manager.VerifyTempToken(token)
}

func (s *StpLogic) GetTempTokenInfo(token string) (*security.TempTokenInfo, error) {
	return s.manager.GetTempTokenInfo(token)
}

func (s *StpLogic) DeleteTempToken(token string) error {
	return s.manager.DeleteTempToken(token)
}

func (s *StpLogic) GetTokenLastActiveTime(tokenValue string) (int64, error) {
	return s.manager.GetTokenLastActiveTime(tokenValue)
}

func (s *StpLogic) CheckActiveTimeout(tokenValue string) error {
	return s.manager.CheckActiveTimeout(tokenValue)
}

func (s *StpLogic) GetTokenTimeout(tokenValue string) (int64, error) {
	return s.manager.GetTokenTimeout(tokenValue)
}

func (s *StpLogic) GetSessionTimeout(loginID interface{}) (int64, error) {
	return s.manager.GetSessionTimeout(toString(loginID))
}

func (s *StpLogic) GetTokenSessionTimeout(tokenValue string) (int64, error) {
	return s.manager.GetTokenSessionTimeout(tokenValue)
}

func (s *StpLogic) RenewTimeout(tokenValue string, sec int64) error {
	return s.manager.RenewTimeout(tokenValue, sec)
}

func (s *StpLogic) SearchTokenValue(keyword string, start, size int, asc bool) ([]string, error) {
	return s.manager.SearchTokenValue(keyword, start, size, asc)
}

func (s *StpLogic) SearchSessionID(keyword string, start, size int, asc bool) ([]string, error) {
	return s.manager.SearchSessionID(keyword, start, size, asc)
}

func (s *StpLogic) SearchTokenSessionID(keyword string, start, size int, asc bool) ([]string, error) {
	return s.manager.SearchTokenSessionID(keyword, start, size, asc)
}

func (s *StpLogic) GetTokenValueByLoginID(loginID interface{}, device ...string) (string, error) {
	return s.manager.GetTokenValue(toString(loginID), device...)
}

func (s *StpLogic) GetTerminalListByLoginID(loginID interface{}, device ...string) ([]string, error) {
	return s.manager.GetTerminalListByLoginID(toString(loginID), device...)
}

func (s *StpLogic) GetTerminalInfo(tokenValue string) (*manager.TokenInfo, error) {
	return s.manager.GetTerminalInfo(tokenValue)
}

func (s *StpLogic) GetLoginDeviceType(tokenValue string) (string, error) {
	return s.manager.GetLoginDeviceType(tokenValue)
}

func (s *StpLogic) GetLoginDeviceID(tokenValue string) (string, error) {
	return s.manager.GetLoginDeviceID(tokenValue)
}

func (s *StpLogic) IsTrustDeviceID(loginID interface{}, deviceID string) bool {
	return s.manager.IsTrustDeviceID(toString(loginID), deviceID)
}

// ============ Same-Token | 服务间调用令牌 ============

// GetSameToken returns the current same-token | 获取服务间调用令牌
func (s *StpLogic) GetSameToken() (string, error) {
	return s.manager.GetSameToken()
}

// RefreshSameToken rotates the same-token | 刷新服务间调用令牌
func (s *StpLogic) RefreshSameToken() (string, error) {
	return s.manager.RefreshSameToken()
}

// CheckSameToken validates a same-token value | 验证服务间调用令牌
func (s *StpLogic) CheckSameToken(tokenValue string) error {
	return s.manager.CheckSameToken(tokenValue)
}

// IsSameTokenValid checks if a same-token is valid | 检查服务间调用令牌是否有效
func (s *StpLogic) IsSameTokenValid(tokenValue string) bool {
	return s.manager.IsSameTokenValid(tokenValue)
}

// CloseManager Closes the manager and releases all resources | 关闭管理器并释放所有资源
func (s *StpLogic) CloseManager() {
	s.manager.CloseManager()
}
