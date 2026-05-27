package stputil

import (
	"context"
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

// Security type aliases | 安全模块类型别名
type (
	ApiKeyInfo        = security.ApiKeyInfo
	TempTokenInfo     = security.TempTokenInfo
	SameTokenTemplate = security.SameTokenTemplate
)

// Global Manager instance | 全局Manager实例
var (
	globalLogic *StpLogic
	//globalManager *manager.Manager
	once sync.Once
	mu   sync.RWMutex
)

// SetManager sets the global Manager (must be called first) | 设置全局Manager（必须先调用此方法）
func SetManager(mgr *manager.Manager) {
	if globalLogic == nil {
		globalLogic = NewStpLogic(mgr)
	}
	globalLogic.SetManager(mgr)
}

// GetManager gets the global Manager | 获取全局Manager
func GetManager() *manager.Manager {
	if globalLogic == nil {
		panic("StpUtil not initialized, please call SetManager() first")
	}
	if globalLogic.GetManager() == nil {
		panic("StpUtil not initialized, please call SetManager() first")
	}
	return globalLogic.GetManager()
}

// CloseManager closes global Manager and releases resources | 关闭全局 Manager 并释放资源
func CloseManager() {
	if globalLogic != nil {
		globalLogic.CloseManager()
		globalLogic.SetManager(nil) // 置 nil 避免后续误用
	}
}

// StpLogic — global facade

// SetStpLogic sets the global StpLogic instance | 设置全局 StpLogic 实例
func SetStpLogic(logic *StpLogic) {
	mu.Lock()
	defer mu.Unlock()
	globalLogic = logic
}

// GetStpLogic gets the global StpLogic instance | 获取全局 StpLogic 实例
func GetStpLogic() *StpLogic {
	mu.Lock()
	defer mu.Unlock()
	return globalLogic
}

// ============ Authentication | 登录认证 ============

// Login performs user login | 用户登录
func Login(loginID interface{}, device ...string) (string, error) {
	return globalLogic.Login(toString(loginID), device...)
}

// LoginByToken performs login with specified token | 使用指定Token登录
func LoginByToken(loginID interface{}, tokenValue string, device ...string) error {
	return globalLogic.LoginByToken(toString(loginID), tokenValue, device...)
}

// Logout performs user logout | 用户登出
func Logout(loginID interface{}, device ...string) error {
	return globalLogic.Logout(toString(loginID), device...)
}

// LogoutByToken performs logout by token | 根据Token登出
func LogoutByToken(tokenValue string) error {
	return globalLogic.LogoutByToken(tokenValue)
}

// IsLogin checks if the user is logged in | 检查用户是否已登录
func IsLogin(tokenValue string) bool {
	return globalLogic.IsLogin(tokenValue)
}

// CheckLogin checks login status (throws error if not logged in) | 检查登录状态（未登录抛出错误）
func CheckLogin(tokenValue string) error {
	return globalLogic.CheckLogin(tokenValue)
}

// GetLoginID gets the login ID from token | 从Token获取登录ID
func GetLoginID(tokenValue string) (string, error) {
	return globalLogic.GetLoginID(tokenValue)
}

// GetLoginIDNotCheck gets login ID without checking | 获取登录ID（不检查）
func GetLoginIDNotCheck(tokenValue string) (string, error) {
	return globalLogic.GetLoginIDNotCheck(tokenValue)
}

// GetTokenValue gets the token value for a login ID | 获取登录ID对应的Token值
func GetTokenValue(loginID interface{}, device ...string) (string, error) {
	return globalLogic.GetTokenValue(toString(loginID), device...)
}

// GetTokenInfo gets token information | 获取Token信息
func GetTokenInfo(tokenValue string) (*manager.TokenInfo, error) {
	return globalLogic.GetTokenInfo(tokenValue)
}

// ============ Kickout | 踢人下线 ============

// Kickout kicks out a user session | 踢人下线
func Kickout(loginID interface{}, device ...string) error {
	return globalLogic.Kickout(toString(loginID), device...)
}

// ============ Account Disable | 账号封禁 ============

// Disable disables an account for specified duration | 封禁账号（指定时长）
func Disable(loginID interface{}, duration time.Duration) error {
	return globalLogic.Disable(toString(loginID), duration)
}

// Untie re-enables a disabled account | 解封账号
func Untie(loginID interface{}) error {
	return globalLogic.Untie(toString(loginID))
}

// IsDisable checks if an account is disabled | 检查账号是否被封禁
func IsDisable(loginID interface{}) bool {
	return globalLogic.IsDisable(toString(loginID))
}

// GetDisableTime gets remaining disable time in seconds | 获取剩余封禁时间（秒）
func GetDisableTime(loginID interface{}) (int64, error) {
	return globalLogic.GetDisableTime(toString(loginID))
}

// ============ Session Management | Session管理 ============

// GetSession gets session by login ID | 根据登录ID获取Session
func GetSession(loginID interface{}) (*session.Session, error) {
	return globalLogic.GetSession(toString(loginID))
}

// GetSessionByToken gets session by token | 根据Token获取Session
func GetSessionByToken(tokenValue string) (*session.Session, error) {
	return globalLogic.GetSessionByToken(tokenValue)
}

// DeleteSession deletes a session | 删除Session
func DeleteSession(loginID interface{}) error {
	return globalLogic.DeleteSession(toString(loginID))
}

// ============ Permission Verification | 权限验证 ============

// SetPermissions sets permissions for a login ID | 设置用户权限
func SetPermissions(loginID interface{}, permissions []string) error {
	return globalLogic.SetPermissions(toString(loginID), permissions)
}

// GetPermissions gets permission list | 获取权限列表
func GetPermissions(loginID interface{}) ([]string, error) {
	return globalLogic.GetPermissions(toString(loginID))
}

// HasPermission checks if has specified permission | 检查是否拥有指定权限
func HasPermission(loginID interface{}, permission string) bool {
	return globalLogic.HasPermission(toString(loginID), permission)
}

// HasPermissionsAnd checks if has all permissions (AND logic) | 检查是否拥有所有权限（AND逻辑）
func HasPermissionsAnd(loginID interface{}, permissions []string) bool {
	return globalLogic.HasPermissionsAnd(toString(loginID), permissions)
}

// HasPermissionsOr checks if has any permission (OR logic) | 检查是否拥有任一权限（OR逻辑）
func HasPermissionsOr(loginID interface{}, permissions []string) bool {
	return globalLogic.HasPermissionsOr(toString(loginID), permissions)
}

// ============ Role Management | 角色管理 ============

// SetRoles sets roles for a login ID | 设置用户角色
func SetRoles(loginID interface{}, roles []string) error {
	return globalLogic.SetRoles(toString(loginID), roles)
}

// GetRoles gets role list | 获取角色列表
func GetRoles(loginID interface{}) ([]string, error) {
	return globalLogic.GetRoles(toString(loginID))
}

// HasRole checks if has specified role | 检查是否拥有指定角色
func HasRole(loginID interface{}, role string) bool {
	return globalLogic.HasRole(toString(loginID), role)
}

// HasRolesAnd checks if has all roles (AND logic) | 检查是否拥有所有角色（AND逻辑）
func HasRolesAnd(loginID interface{}, roles []string) bool {
	return globalLogic.HasRolesAnd(toString(loginID), roles)
}

// HasRolesOr 检查是否拥有任一角色（OR）
func HasRolesOr(loginID interface{}, roles []string) bool {
	return globalLogic.HasRolesOr(toString(loginID), roles)
}

// ============ Token标签 ============

// SetTokenTag 设置Token标签
func SetTokenTag(tokenValue, tag string) error {
	return globalLogic.SetTokenTag(tokenValue, tag)
}

// GetTokenTag 获取Token标签
func GetTokenTag(tokenValue string) (string, error) {
	return globalLogic.GetTokenTag(tokenValue)
}

// ============ 会话查询 ============

// GetTokenValueList 获取指定账号的所有Token
func GetTokenValueList(loginID interface{}) ([]string, error) {
	return globalLogic.GetTokenValueList(loginID)
}

// GetSessionCount 获取指定账号的Session数量
func GetSessionCount(loginID interface{}) (int, error) {
	return globalLogic.GetSessionCount(loginID)
}

// ============ 辅助方法 ============

// toString 将interface{}转换为string
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return intToString(val)
	case int64:
		return int64ToString(val)
	case uint:
		return uintToString(val)
	case uint64:
		return uint64ToString(val)
	default:
		return ""
	}
}

func intToString(i int) string {
	return int64ToString(int64(i))
}

func int64ToString(i int64) string {
	// 简单实现，可以用 strconv.FormatInt(i, 10) 但为了减少依赖
	if i == 0 {
		return "0"
	}

	negative := i < 0
	if negative {
		i = -i
	}

	var result []byte
	for i > 0 {
		result = append([]byte{byte('0' + i%10)}, result...)
		i /= 10
	}

	if negative {
		result = append([]byte{'-'}, result...)
	}

	return string(result)
}

func uintToString(u uint) string {
	return uint64ToString(uint64(u))
}

func uint64ToString(u uint64) string {
	if u == 0 {
		return "0"
	}

	var result []byte
	for u > 0 {
		result = append([]byte{byte('0' + u%10)}, result...)
		u /= 10
	}

	return string(result)
}

func GenerateNonce() (string, error) {
	if globalLogic == nil {
		panic("Manager not initialized. Call stputil.SetManager() first")
	}
	return globalLogic.GenerateNonce()
}

func VerifyNonce(nonce string) bool {
	if globalLogic == nil {
		panic("Manager not initialized. Call stputil.SetManager() first")
	}
	return globalLogic.VerifyNonce(nonce)
}

func LoginWithRefreshToken(loginID interface{}, device ...string) (*security.RefreshTokenInfo, error) {
	if globalLogic == nil {
		panic("Manager not initialized. Call stputil.SetManager() first")
	}
	deviceType := "default"
	if len(device) > 0 {
		deviceType = device[0]
	}
	return globalLogic.LoginWithRefreshToken(fmt.Sprintf("%v", loginID), deviceType)
}

func RefreshAccessToken(refreshToken string) (*security.RefreshTokenInfo, error) {
	if globalLogic == nil {
		panic("Manager not initialized. Call stputil.SetManager() first")
	}
	return globalLogic.RefreshAccessToken(refreshToken)
}

func RevokeRefreshToken(refreshToken string) error {
	if globalLogic == nil {
		panic("Manager not initialized. Call stputil.SetManager() first")
	}
	return globalLogic.RevokeRefreshToken(refreshToken)
}

func GetOAuth2Server() *oauth2.OAuth2Server {
	if globalLogic == nil {
		panic("Manager not initialized. Call stputil.SetManager() first")
	}
	return globalLogic.GetOAuth2Server()
}

// ============ Check Functions for Token-based operations | 基于Token的检查函数 ============

// CheckDisable checks if the account associated with the token is disabled | 检查Token对应账号是否被封禁
func CheckDisable(tokenValue string) error {
	loginID, err := GetLoginID(tokenValue)
	if err != nil {
		return err
	}
	if IsDisable(loginID) {
		return core.NewAccountDisabledError(loginID)
	}
	return nil
}

// CheckPermission checks if the token has the specified permission | 检查Token是否拥有指定权限
func CheckPermission(tokenValue string, permission string) error {
	loginID, err := GetLoginID(tokenValue)
	if err != nil {
		return err
	}
	if !HasPermission(loginID, permission) {
		return core.NewPermissionDeniedError(permission)
	}
	return nil
}

// CheckPermissionAnd checks if the token has all specified permissions | 检查Token是否拥有所有指定权限
func CheckPermissionAnd(tokenValue string, permissions []string) error {
	loginID, err := GetLoginID(tokenValue)
	if err != nil {
		return err
	}
	if !HasPermissionsAnd(loginID, permissions) {
		return core.NewPermissionDeniedListError(permissions)
	}
	return nil
}

// CheckPermissionOr checks if the token has any of the specified permissions | 检查Token是否拥有任一指定权限
func CheckPermissionOr(tokenValue string, permissions []string) error {
	loginID, err := GetLoginID(tokenValue)
	if err != nil {
		return err
	}
	if !HasPermissionsOr(loginID, permissions) {
		return core.NewPermissionDeniedListError(permissions)
	}
	return nil
}

// GetPermissionList gets permission list for the token | 获取Token对应的权限列表
func GetPermissionList(tokenValue string) ([]string, error) {
	loginID, err := GetLoginID(tokenValue)
	if err != nil {
		return nil, err
	}
	return GetPermissions(loginID)
}

// CheckRole checks if the token has the specified role | 检查Token是否拥有指定角色
func CheckRole(tokenValue string, role string) error {
	loginID, err := GetLoginID(tokenValue)
	if err != nil {
		return err
	}
	if !HasRole(loginID, role) {
		return core.NewRoleDeniedError(role)
	}
	return nil
}

// CheckRoleAnd checks if the token has all specified roles | 检查Token是否拥有所有指定角色
func CheckRoleAnd(tokenValue string, roles []string) error {
	loginID, err := GetLoginID(tokenValue)
	if err != nil {
		return err
	}
	if !HasRolesAnd(loginID, roles) {
		return core.NewRoleDeniedListError(roles)
	}
	return nil
}

// CheckRoleOr checks if the token has any of the specified roles | 检查Token是否拥有任一指定角色
func CheckRoleOr(tokenValue string, roles []string) error {
	loginID, err := GetLoginID(tokenValue)
	if err != nil {
		return err
	}
	if !HasRolesOr(loginID, roles) {
		return core.NewRoleDeniedListError(roles)
	}
	return nil
}

// GetRoleList gets role list for the token | 获取Token对应的角色列表
func GetRoleList(tokenValue string) ([]string, error) {
	loginID, err := GetLoginID(tokenValue)
	if err != nil {
		return nil, err
	}
	return GetRoles(loginID)
}

// GetTokenSession loads Token-Session (token-scoped) | 获取 Token-Session
func GetTokenSession(tokenValue string) (*session.Session, error) {
	return globalLogic.GetTokenSession(tokenValue)
}

// GetTokenSessionOrCreate loads or creates Token-Session | 获取或创建 Token-Session
func GetTokenSessionOrCreate(tokenValue string) (*session.Session, error) {
	return globalLogic.GetTokenSessionOrCreate(tokenValue)
}

// GetAnonTokenSession anonymous Token-Session | 匿名 Token-Session
func GetAnonTokenSession(currentToken string, ctx adapter.RequestContext) (*session.Session, error) {
	return globalLogic.GetAnonTokenSession(currentToken, ctx)
}

// DeleteTokenSession removes Token-Session | 删除 Token-Session
func DeleteTokenSession(tokenValue string) error {
	return globalLogic.DeleteTokenSession(tokenValue)
}

// Replaced overrun logout | 顶号下线
func Replaced(loginID interface{}, device ...string) error {
	return globalLogic.Replaced(toString(loginID), device...)
}

// ReplacedByToken overrun logout by token | 按 Token 顶号下线
func ReplacedByToken(tokenValue string) error {
	return globalLogic.ReplacedByToken(tokenValue)
}

// KickoutByToken kicks by token | 按 Token 踢下线
func KickoutByToken(tokenValue string) error {
	return globalLogic.KickoutByToken(tokenValue)
}

// SetTokenValueToCtx writes token into context for downstream GetLoginIDFromCtx | 将 Token 写入上下文
func SetTokenValueToCtx(parent context.Context, token string) context.Context {
	return context.WithValue(parent, TokenValueKey, token)
}

// GetLoginIDFromCtx prefers SwitchTo loginId, else resolves token from context | 优先 SwitchTo，再从 Token 解析
func GetLoginIDFromCtx(ctx context.Context) (string, error) {
	if id := GetSwitchLoginID(ctx); id != "" {
		return id, nil
	}
	tv := GetTokenValueFromCtx(ctx)
	if tv == "" {
		return "", core.ErrNotLogin
	}
	return GetLoginID(tv)
}

// GetTokenValueFromCtx reads token from context key stputil.TokenValueKey if set | 从上下文取 Token
func GetTokenValueFromCtx(ctx context.Context) string {
	if v := ctx.Value(TokenValueKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetTokenInfoFromCtx builds token info using GetTokenValueFromCtx | 从上下文取 TokenInfo
func GetTokenInfoFromCtx(ctx context.Context) (*manager.TokenInfo, error) {
	tv := GetTokenValueFromCtx(ctx)
	if tv == "" {
		return nil, core.ErrNotLogin
	}
	return GetTokenInfo(tv)
}

// LoginRememberMe performs login with remember-me mode | 记住我模式登录
func LoginRememberMe(loginID interface{}, device ...string) (string, error) {
	return globalLogic.LoginRememberMe(toString(loginID), device...)
}

// IsRememberMeLogin checks if token was created with remember-me | 检查是否为记住我模式登录
func IsRememberMeLogin(tokenValue string) (bool, error) {
	return globalLogic.IsRememberMeLogin(tokenValue)
}

// UpdateLastActiveToNow updates active time | 更新活跃时间
func UpdateLastActiveToNow(tokenValue string) error {
	return globalLogic.UpdateLastActiveToNow(tokenValue)
}

func GetTokenLastActiveTime(tokenValue string) (int64, error) {
	return globalLogic.GetTokenLastActiveTime(tokenValue)
}

func CheckActiveTimeout(tokenValue string) error {
	return globalLogic.CheckActiveTimeout(tokenValue)
}

func GetTokenTimeout(tokenValue string) (int64, error) {
	return globalLogic.GetTokenTimeout(tokenValue)
}

func GetSessionTimeout(loginID interface{}) (int64, error) {
	return globalLogic.GetSessionTimeout(toString(loginID))
}

func GetTokenSessionTimeout(tokenValue string) (int64, error) {
	return globalLogic.GetTokenSessionTimeout(tokenValue)
}

func RenewTimeout(tokenValue string, sec int64) error {
	return globalLogic.RenewTimeout(tokenValue, sec)
}

func SearchTokenValue(keyword string, start, size int, asc bool) ([]string, error) {
	return globalLogic.SearchTokenValue(keyword, start, size, asc)
}

func SearchSessionID(keyword string, start, size int, asc bool) ([]string, error) {
	return globalLogic.SearchSessionID(keyword, start, size, asc)
}

func SearchTokenSessionID(keyword string, start, size int, asc bool) ([]string, error) {
	return globalLogic.SearchTokenSessionID(keyword, start, size, asc)
}

func GetTokenValueByLoginID(loginID interface{}, device ...string) (string, error) {
	return globalLogic.GetTokenValueByLoginID(loginID, device...)
}

func GetTerminalListByLoginID(loginID interface{}, device ...string) ([]string, error) {
	return globalLogic.GetTerminalListByLoginID(toString(loginID), device...)
}

func GetTerminalInfo(tokenValue string) (*manager.TokenInfo, error) {
	return globalLogic.GetTerminalInfo(tokenValue)
}

func GetLoginDeviceType(tokenValue string) (string, error) {
	return globalLogic.GetLoginDeviceType(tokenValue)
}

func GetLoginDeviceID(tokenValue string) (string, error) {
	return globalLogic.GetLoginDeviceID(tokenValue)
}

func IsTrustDeviceID(loginID interface{}, deviceID string) bool {
	return globalLogic.IsTrustDeviceID(toString(loginID), deviceID)
}

// ============ API Key | API Key 管理 ============

func CreateApiKey(loginID, title string, expireSeconds int64, extra string) (*security.ApiKeyInfo, error) {
	return globalLogic.CreateApiKey(loginID, title, expireSeconds, extra)
}

func GetApiKeyInfo(key string) (*security.ApiKeyInfo, error) {
	return globalLogic.GetApiKeyInfo(key)
}

func VerifyApiKey(key string) (*security.ApiKeyInfo, error) {
	return globalLogic.VerifyApiKey(key)
}

func DeleteApiKey(key string) error {
	return globalLogic.DeleteApiKey(key)
}

func DisableApiKey(key string) error {
	return globalLogic.DisableApiKey(key)
}

func EnableApiKey(key string) error {
	return globalLogic.EnableApiKey(key)
}

// ============ Signature | 参数签名 ============

func Sign(params map[string]string, secret string) string {
	return globalLogic.Sign(params, secret)
}

func VerifySign(params map[string]string, secret, timestamp, nonce, signature string, maxAgeSeconds int64) error {
	return globalLogic.VerifySign(params, secret, timestamp, nonce, signature, maxAgeSeconds)
}

// ============ Temp Token | 临时Token ============

func CreateTempToken(loginID string, expireSeconds int64, extra string) (*security.TempTokenInfo, error) {
	return globalLogic.CreateTempToken(loginID, expireSeconds, extra)
}

func VerifyTempToken(token string) (*security.TempTokenInfo, error) {
	return globalLogic.VerifyTempToken(token)
}

func GetTempTokenInfo(token string) (*security.TempTokenInfo, error) {
	return globalLogic.GetTempTokenInfo(token)
}

func DeleteTempToken(token string) error {
	return globalLogic.DeleteTempToken(token)
}

// ============ Same-Token | 服务间调用令牌 ============

// GetSameToken returns the current same-token | 获取服务间调用令牌
func GetSameToken() (string, error) {
	return globalLogic.GetSameToken()
}

// RefreshSameToken rotates the same-token | 刷新服务间调用令牌
func RefreshSameToken() (string, error) {
	return globalLogic.RefreshSameToken()
}

// CheckSameToken validates a same-token value | 验证服务间调用令牌
func CheckSameToken(tokenValue string) error {
	return globalLogic.CheckSameToken(tokenValue)
}

// IsSameTokenValid checks if a same-token is valid | 检查服务间调用令牌是否有效
func IsSameTokenValid(tokenValue string) bool {
	return globalLogic.IsSameTokenValid(tokenValue)
}
