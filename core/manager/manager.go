package manager

import (
	"fmt"
	"strings"
	"time"

	"github.com/click33/sa-token-go/core/adapter"
	"github.com/click33/sa-token-go/core/config"
	"github.com/click33/sa-token-go/core/listener"
	"github.com/click33/sa-token-go/core/oauth2"
	"github.com/click33/sa-token-go/core/security"
	"github.com/click33/sa-token-go/core/session"
	"github.com/click33/sa-token-go/core/token"
)

// Constants for storage keys and default values | 存储键和默认值常量
const (
	DefaultDevice   = "default"
	DefaultPrefix   = "satoken"
	DisableValue    = "1"
	DefaultNonceTTL = 5 * time.Minute

	// Key prefixes | 键前缀
	TokenKeyPrefix   = "token:"
	AccountKeyPrefix = "account:"
	DisableKeyPrefix = "disable:"

	// Session keys | Session键
	SessionKeyLoginID     = "loginId"
	SessionKeyDevice      = "device"
	SessionKeyLoginTime   = "loginTime"
	SessionKeyPermissions = "permissions"
	SessionKeyRoles       = "roles"

	// Wildcard for permissions | 权限通配符
	PermissionWildcard  = "*"
	PermissionSeparator = ":"
)

// Error variables | 错误变量
var (
	ErrAccountDisabled  = fmt.Errorf("account is disabled")
	ErrNotLogin         = fmt.Errorf("not login")
	ErrTokenNotFound    = fmt.Errorf("token not found")
	ErrInvalidTokenData = fmt.Errorf("invalid token data")
)

// TokenInfo Token information | Token信息
type TokenInfo struct {
	LoginID    string `json:"loginId"`
	Device     string `json:"device"`
	CreateTime int64  `json:"createTime"`
	ActiveTime int64  `json:"activeTime"` // Last active time | 最后活跃时间
	Tag        string `json:"tag,omitempty"`
}

// Manager Authentication manager | 认证管理器
type Manager struct {
	storage        adapter.Storage
	config         *config.Config
	generator      *token.Generator
	prefix         string
	nonceManager   *security.NonceManager
	refreshManager *security.RefreshTokenManager
	oauth2Server   *oauth2.OAuth2Server
	eventManager   *listener.Manager
}

// NewManager Creates a new manager | 创建管理器
func NewManager(storage adapter.Storage, cfg *config.Config) *Manager {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Use configured prefix, fallback to default | 使用配置的前缀，回退到默认值
	prefix := cfg.KeyPrefix
	if prefix == "" {
		prefix = DefaultPrefix
	}

	return &Manager{
		storage:        storage,
		config:         cfg,
		generator:      token.NewGenerator(cfg),
		prefix:         prefix,
		nonceManager:   security.NewNonceManager(storage, prefix, DefaultNonceTTL),
		refreshManager: security.NewRefreshTokenManager(storage, prefix, cfg),
		oauth2Server:   oauth2.NewOAuth2Server(storage, prefix),
		eventManager:   listener.NewManager(),
	}
}

// ============ Helper Methods | 辅助方法 ============

// getDevice extracts device type from optional parameter | 从可选参数中提取设备类型
func getDevice(device []string) string {
	if len(device) > 0 && device[0] != "" {
		return device[0]
	}
	return DefaultDevice
}

// getExpiration calculates expiration duration from config | 从配置计算过期时间
func (m *Manager) getExpiration() time.Duration {
	if m.config.Timeout > 0 {
		return time.Duration(m.config.Timeout) * time.Second
	}
	return 0
}

// assertString safely converts interface to string | 安全地将interface转换为string
func assertString(v any) (string, bool) {
	s, ok := v.(string)
	return s, ok
}

// ============ Login Authentication | 登录认证 ============

// Login Performs user login and returns token | 登录，返回Token
func (m *Manager) Login(loginID string, device ...string) (string, error) {
	deviceType := getDevice(device)

	// Check if account is disabled | 检查是否被封禁
	if m.IsDisable(loginID) {
		return "", ErrAccountDisabled
	}

	// Kick out old session if concurrent login is not allowed | 如果不允许并发登录，先踢掉旧的
	if !m.config.IsConcurrent {
		m.kickout(loginID, deviceType)
	}

	// Generate token | 生成Token
	tokenValue, err := m.generator.Generate(loginID, deviceType)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	expiration := m.getExpiration()

	// Save token-loginID mapping (符合 Java sa-token 设计) | 保存 Token-LoginID 映射
	tokenKey := m.getTokenKey(tokenValue)
	if err := m.storage.Set(tokenKey, loginID, expiration); err != nil {
		return "", fmt.Errorf("failed to save token: %w", err)
	}

	// Save account-token mapping | 保存账号-Token映射
	accountKey := m.getAccountKey(loginID, deviceType)
	if err := m.storage.Set(accountKey, tokenValue, expiration); err != nil {
		return "", fmt.Errorf("failed to save account mapping: %w", err)
	}

	// Create session | 创建Session
	sess := session.NewSession(loginID, m.storage, m.prefix)
	sess.Set(SessionKeyLoginID, loginID)
	sess.Set(SessionKeyDevice, deviceType)
	sess.Set(SessionKeyLoginTime, time.Now().Unix())

	// Trigger login event | 触发登录事件
	if m.eventManager != nil {
		m.eventManager.Trigger(&listener.EventData{
			Event:   listener.EventLogin,
			LoginID: loginID,
			Token:   tokenValue,
			Device:  deviceType,
		})
	}

	return tokenValue, nil
}

// LoginByToken Login with specified token (for seamless token refresh) | 使用指定Token登录（用于token无感刷新）
func (m *Manager) LoginByToken(loginID string, tokenValue string, device ...string) error {
	deviceType := getDevice(device)
	expiration := m.getExpiration()

	// Save token-loginID mapping (符合 Java sa-token 设计) | 保存 Token-LoginID 映射
	tokenKey := m.getTokenKey(tokenValue)
	if err := m.storage.Set(tokenKey, loginID, expiration); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	accountKey := m.getAccountKey(loginID, deviceType)
	return m.storage.Set(accountKey, tokenValue, expiration)
}

// Logout Performs user logout | 登出
func (m *Manager) Logout(loginID string, device ...string) error {
	deviceType := getDevice(device)
	accountKey := m.getAccountKey(loginID, deviceType)

	tokenValue, err := m.storage.Get(accountKey)
	if err != nil || tokenValue == nil {
		return nil // Already logged out | 已经登出
	}

	// Delete token | 删除Token
	tokenStr, ok := assertString(tokenValue)
	if !ok {
		return nil
	}

	tokenKey := m.getTokenKey(tokenStr)
	m.storage.Delete(tokenKey)

	// Delete account mapping | 删除账号映射
	m.storage.Delete(accountKey)

	// Trigger logout event | 触发登出事件
	if m.eventManager != nil {
		m.eventManager.Trigger(&listener.EventData{
			Event:   listener.EventLogout,
			LoginID: loginID,
			Device:  deviceType,
		})
	}

	return nil
}

// LogoutByToken Logout by token | 根据Token登出
func (m *Manager) LogoutByToken(tokenValue string) error {
	if tokenValue == "" {
		return nil
	}

	// Get loginID before deletion for event | 删除前获取loginID用于事件
	loginID, _ := m.getLoginIDByToken(tokenValue)

	tokenKey := m.getTokenKey(tokenValue)
	err := m.storage.Delete(tokenKey)

	// Trigger logout event | 触发登出事件
	if m.eventManager != nil && loginID != "" {
		m.eventManager.Trigger(&listener.EventData{
			Event:   listener.EventLogout,
			LoginID: loginID,
			Token:   tokenValue,
		})
	}

	return err
}

// kickout Kick user offline (private) | 踢人下线（私有）
func (m *Manager) kickout(loginID string, device string) error {
	accountKey := m.getAccountKey(loginID, device)
	tokenValue, err := m.storage.Get(accountKey)
	if err != nil || tokenValue == nil {
		return nil
	}

	tokenStr, ok := assertString(tokenValue)
	if !ok {
		return nil
	}

	tokenKey := m.getTokenKey(tokenStr)

	// Trigger kickout event | 触发踢出事件
	if m.eventManager != nil {
		m.eventManager.Trigger(&listener.EventData{
			Event:   listener.EventKickout,
			LoginID: loginID,
			Token:   tokenStr,
			Device:  device,
		})
	}

	return m.storage.Delete(tokenKey)
}

// Kickout Kick user offline (public method) | 踢人下线（公开方法）
func (m *Manager) Kickout(loginID string, device ...string) error {
	deviceType := getDevice(device)
	return m.kickout(loginID, deviceType)
}

// ============ Token Validation | Token验证 ============

// IsLogin Checks if user is logged in | 检查是否登录
func (m *Manager) IsLogin(tokenValue string) bool {
	if tokenValue == "" {
		return false
	}

	tokenKey := m.getTokenKey(tokenValue)
	if !m.storage.Exists(tokenKey) {
		return false
	}

	// Async auto-renew for better performance | 异步自动续期（提高性能）
	// Note: ActiveTimeout feature removed to comply with Java sa-token design
	if m.config.AutoRenew && m.config.Timeout > 0 {
		go m.renewToken(tokenKey)
	}

	return true
}

// renewToken Renews token expiration asynchronously | 异步续期Token
func (m *Manager) renewToken(tokenKey string) {
	expiration := m.getExpiration()
	// Extend token storage expiration | 延长Token存储的过期时间
	m.storage.Expire(tokenKey, expiration)
}

// CheckLogin Checks login status (throws error if not logged in) | 检查登录（未登录抛出错误）
func (m *Manager) CheckLogin(tokenValue string) error {
	if !m.IsLogin(tokenValue) {
		return ErrNotLogin
	}
	return nil
}

// GetLoginID Gets login ID from token | 根据Token获取登录ID
func (m *Manager) GetLoginID(tokenValue string) (string, error) {
	if !m.IsLogin(tokenValue) {
		return "", ErrNotLogin
	}

	info, err := m.getTokenInfo(tokenValue)
	if err != nil {
		return "", err
	}

	return info.LoginID, nil
}

// GetLoginIDNotCheck Gets login ID without checking token validity | 获取登录ID（不检查Token是否有效）
func (m *Manager) GetLoginIDNotCheck(tokenValue string) (string, error) {
	info, err := m.getTokenInfo(tokenValue)
	if err != nil {
		return "", err
	}
	return info.LoginID, nil
}

// GetTokenValue Gets token by login ID | 根据登录ID获取Token
func (m *Manager) GetTokenValue(loginID string, device ...string) (string, error) {
	deviceType := getDevice(device)
	accountKey := m.getAccountKey(loginID, deviceType)

	tokenValue, err := m.storage.Get(accountKey)
	if err != nil || tokenValue == nil {
		return "", fmt.Errorf("token not found for login id: %s", loginID)
	}

	tokenStr, ok := assertString(tokenValue)
	if !ok {
		return "", fmt.Errorf("invalid token value type")
	}

	return tokenStr, nil
}

// GetTokenInfo Gets token information | 获取Token信息
func (m *Manager) GetTokenInfo(tokenValue string) (*TokenInfo, error) {
	return m.getTokenInfo(tokenValue)
}

// ============ Account Disable | 账号封禁 ============

// Disable Disables an account | 封禁账号
func (m *Manager) Disable(loginID string, duration time.Duration) error {
	key := m.getDisableKey(loginID)
	return m.storage.Set(key, DisableValue, duration)
}

// Untie Re-enables a disabled account | 解封账号
func (m *Manager) Untie(loginID string) error {
	key := m.getDisableKey(loginID)
	return m.storage.Delete(key)
}

// IsDisable Checks if account is disabled | 检查账号是否被封禁
func (m *Manager) IsDisable(loginID string) bool {
	key := m.getDisableKey(loginID)
	return m.storage.Exists(key)
}

// GetDisableTime Gets remaining disable time in seconds | 获取账号剩余封禁时间（秒）
func (m *Manager) GetDisableTime(loginID string) (int64, error) {
	key := m.getDisableKey(loginID)
	ttl, err := m.storage.TTL(key)
	if err != nil {
		return -2, err
	}
	return int64(ttl.Seconds()), nil
}

// getDisableKey Gets disable storage key | 获取禁用存储键
func (m *Manager) getDisableKey(loginID string) string {
	return m.prefix + DisableKeyPrefix + loginID
}

// ============ Session Management | Session管理 ============

// GetSession Gets session by login ID | 获取Session
func (m *Manager) GetSession(loginID string) (*session.Session, error) {
	sess, err := session.Load(loginID, m.storage, m.prefix)
	if err != nil {
		sess = session.NewSession(loginID, m.storage, m.prefix)
	}
	return sess, nil
}

// GetSessionByToken Gets session by token | 根据Token获取Session
func (m *Manager) GetSessionByToken(tokenValue string) (*session.Session, error) {
	loginID, err := m.GetLoginID(tokenValue)
	if err != nil {
		return nil, err
	}
	return m.GetSession(loginID)
}

// DeleteSession Deletes session | 删除Session
func (m *Manager) DeleteSession(loginID string) error {
	sess, err := m.GetSession(loginID)
	if err != nil {
		return err
	}
	return sess.Destroy()
}

// ============ Permission Validation | 权限验证 ============

// SetPermissions Sets permissions for user | 设置权限
func (m *Manager) SetPermissions(loginID string, permissions []string) error {
	sess, err := m.GetSession(loginID)
	if err != nil {
		return err
	}
	return sess.Set(SessionKeyPermissions, permissions)
}

// GetPermissions Gets permission list | 获取权限列表
func (m *Manager) GetPermissions(loginID string) ([]string, error) {
	sess, err := m.GetSession(loginID)
	if err != nil {
		return nil, err
	}

	perms, exists := sess.Get(SessionKeyPermissions)
	if !exists {
		return []string{}, nil
	}

	return m.toStringSlice(perms), nil
}

// HasPermission 检查是否有指定权限
func (m *Manager) HasPermission(loginID string, permission string) bool {
	perms, err := m.GetPermissions(loginID)
	if err != nil {
		return false
	}

	for _, p := range perms {
		if m.matchPermission(p, permission) {
			return true
		}
	}

	return false
}

// HasPermissionsAnd 检查是否拥有所有权限（AND）
func (m *Manager) HasPermissionsAnd(loginID string, permissions []string) bool {
	for _, perm := range permissions {
		if !m.HasPermission(loginID, perm) {
			return false
		}
	}
	return true
}

// HasPermissionsOr 检查是否拥有任一权限（OR）
func (m *Manager) HasPermissionsOr(loginID string, permissions []string) bool {
	for _, perm := range permissions {
		if m.HasPermission(loginID, perm) {
			return true
		}
	}
	return false
}

// matchPermission Matches permission with wildcards support | 权限匹配（支持通配符）
func (m *Manager) matchPermission(pattern, permission string) bool {
	// Exact match or wildcard | 精确匹配或通配符
	if pattern == PermissionWildcard || pattern == permission {
		return true
	}

	// Pattern like "user:*" matches "user:add", "user:delete", etc. | 支持通配符，例如 user:* 匹配 user:add, user:delete等
	wildcardSuffix := PermissionSeparator + PermissionWildcard
	if strings.HasSuffix(pattern, wildcardSuffix) {
		prefix := strings.TrimSuffix(pattern, PermissionWildcard)
		return strings.HasPrefix(permission, prefix)
	}

	// Pattern like "user:*:view" | 支持 user:*:view 这样的模式
	if strings.Contains(pattern, PermissionWildcard) {
		parts := strings.Split(pattern, PermissionSeparator)
		permParts := strings.Split(permission, PermissionSeparator)
		if len(parts) != len(permParts) {
			return false
		}
		for i, part := range parts {
			if part != PermissionWildcard && part != permParts[i] {
				return false
			}
		}
		return true
	}

	return false
}

// ============ Role Validation | 角色验证 ============

// SetRoles Sets roles for user | 设置角色
func (m *Manager) SetRoles(loginID string, roles []string) error {
	sess, err := m.GetSession(loginID)
	if err != nil {
		return err
	}
	return sess.Set(SessionKeyRoles, roles)
}

// GetRoles Gets role list | 获取角色列表
func (m *Manager) GetRoles(loginID string) ([]string, error) {
	sess, err := m.GetSession(loginID)
	if err != nil {
		return nil, err
	}

	roles, exists := sess.Get(SessionKeyRoles)
	if !exists {
		return []string{}, nil
	}

	return m.toStringSlice(roles), nil
}

// HasRole 检查是否有指定角色
func (m *Manager) HasRole(loginID string, role string) bool {
	roles, err := m.GetRoles(loginID)
	if err != nil {
		return false
	}

	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasRolesAnd 检查是否拥有所有角色（AND）
func (m *Manager) HasRolesAnd(loginID string, roles []string) bool {
	for _, role := range roles {
		if !m.HasRole(loginID, role) {
			return false
		}
	}
	return true
}

// HasRolesOr 检查是否拥有任一角色（OR）
func (m *Manager) HasRolesOr(loginID string, roles []string) bool {
	for _, role := range roles {
		if m.HasRole(loginID, role) {
			return true
		}
	}
	return false
}

// ============ Token Tags | Token标签 ============

// SetTokenTag Sets token tag | 设置Token标签
func (m *Manager) SetTokenTag(tokenValue, tag string) error {
	// Tag feature not supported to comply with Java sa-token design
	// If you need custom metadata, use Session instead
	return fmt.Errorf("token tag feature not supported (use Session for custom metadata)")
}

// GetTokenTag Gets token tag | 获取Token标签
func (m *Manager) GetTokenTag(tokenValue string) (string, error) {
	// Tag feature not supported to comply with Java sa-token design
	return "", fmt.Errorf("token tag feature not supported (use Session for custom metadata)")
}

// ============ Session Query | 会话查询 ============

// GetTokenValueListByLoginID Gets all tokens for specified account | 获取指定账号的所有Token
func (m *Manager) GetTokenValueListByLoginID(loginID string) ([]string, error) {
	pattern := m.prefix + AccountKeyPrefix + loginID + ":*"
	keys, err := m.storage.Keys(pattern)
	if err != nil {
		return nil, err
	}

	tokens := make([]string, 0, len(keys))
	for _, key := range keys {
		value, err := m.storage.Get(key)
		if err == nil && value != nil {
			if tokenStr, ok := assertString(value); ok {
				tokens = append(tokens, tokenStr)
			}
		}
	}

	return tokens, nil
}

// GetSessionCountByLoginID Gets session count for specified account | 获取指定账号的Session数量
func (m *Manager) GetSessionCountByLoginID(loginID string) (int, error) {
	tokens, err := m.GetTokenValueListByLoginID(loginID)
	if err != nil {
		return 0, err
	}
	return len(tokens), nil
}

// ============ Internal Helper Methods | 内部辅助方法 ============

// getTokenKey Gets token storage key | 获取Token存储键
func (m *Manager) getTokenKey(tokenValue string) string {
	return m.prefix + TokenKeyPrefix + tokenValue
}

// getAccountKey Gets account storage key | 获取账号存储键
func (m *Manager) getAccountKey(loginID, device string) string {
	return m.prefix + AccountKeyPrefix + loginID + PermissionSeparator + device
}

// getLoginIDByToken Gets loginID by token (符合 Java sa-token 设计) | 通过 Token 获取 loginID
func (m *Manager) getLoginIDByToken(tokenValue string) (string, error) {
	tokenKey := m.getTokenKey(tokenValue)
	data, err := m.storage.Get(tokenKey)
	if err != nil || data == nil {
		return "", ErrTokenNotFound
	}

	loginID, ok := assertString(data)
	if !ok {
		return "", ErrInvalidTokenData
	}

	return loginID, nil
}

// getTokenInfo Gets token information (为了向后兼容) | 获取Token信息（向后兼容）
func (m *Manager) getTokenInfo(tokenValue string) (*TokenInfo, error) {
	loginID, err := m.getLoginIDByToken(tokenValue)
	if err != nil {
		return nil, err
	}

	// 构造简化的 TokenInfo，只包含必要信息
	return &TokenInfo{
		LoginID: loginID,
		Device:  DefaultDevice, // 从 token 无法获取设备信息
	}, nil
}

// toStringSlice Converts any to []string | 将any转换为[]string
func (m *Manager) toStringSlice(v any) []string {
	switch val := v.(type) {
	case []string:
		return val
	case []any:
		result := make([]string, 0, len(val))
		for _, item := range val {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		return []string{}
	}
}

// ============ Event Management | 事件管理 ============

// RegisterFunc registers a function as an event listener | 注册函数作为事件监听器
func (m *Manager) RegisterFunc(event listener.Event, fn func(*listener.EventData)) {
	if m.eventManager != nil {
		m.eventManager.RegisterFunc(event, fn)
	}
}

// Register registers an event listener | 注册事件监听器
func (m *Manager) Register(event listener.Event, listener listener.Listener) string {
	if m.eventManager != nil {
		return m.eventManager.Register(event, listener)
	}
	return ""
}

// RegisterWithConfig registers an event listener with config | 注册带配置的事件监听器
func (m *Manager) RegisterWithConfig(event listener.Event, listener listener.Listener, config listener.ListenerConfig) string {
	if m.eventManager != nil {
		return m.eventManager.RegisterWithConfig(event, listener, config)
	}
	return ""
}

// Unregister removes an event listener by ID | 根据ID移除事件监听器
func (m *Manager) Unregister(id string) bool {
	if m.eventManager != nil {
		return m.eventManager.Unregister(id)
	}
	return false
}

// TriggerEvent manually triggers an event | 手动触发事件
func (m *Manager) TriggerEvent(data *listener.EventData) {
	if m.eventManager != nil {
		m.eventManager.Trigger(data)
	}
}

// WaitEvents waits for all async event listeners to complete | 等待所有异步事件监听器完成
func (m *Manager) WaitEvents() {
	if m.eventManager != nil {
		m.eventManager.Wait()
	}
}

// GetEventManager gets the event manager | 获取事件管理器
func (m *Manager) GetEventManager() *listener.Manager {
	return m.eventManager
}

// ============ Public Getters | 公共获取器 ============

// GetConfig Gets configuration | 获取配置
func (m *Manager) GetConfig() *config.Config {
	return m.config
}

// GetStorage Gets storage | 获取存储
func (m *Manager) GetStorage() adapter.Storage {
	return m.storage
}

// ============ Security Features | 安全特性 ============

// GenerateNonce Generates a one-time nonce | 生成一次性随机数
func (m *Manager) GenerateNonce() (string, error) {
	return m.nonceManager.Generate()
}

// VerifyNonce Verifies a nonce | 验证随机数
func (m *Manager) VerifyNonce(nonce string) bool {
	return m.nonceManager.Verify(nonce)
}

// LoginWithRefreshToken Logs in with refresh token | 使用刷新令牌登录
func (m *Manager) LoginWithRefreshToken(loginID, device string) (*security.RefreshTokenInfo, error) {
	return m.refreshManager.GenerateTokenPair(loginID, device)
}

// RefreshAccessToken Refreshes access token | 刷新访问令牌
func (m *Manager) RefreshAccessToken(refreshToken string) (*security.RefreshTokenInfo, error) {
	return m.refreshManager.RefreshAccessToken(refreshToken)
}

// RevokeRefreshToken Revokes refresh token | 撤销刷新令牌
func (m *Manager) RevokeRefreshToken(refreshToken string) error {
	return m.refreshManager.RevokeRefreshToken(refreshToken)
}

// GetOAuth2Server Gets OAuth2 server instance | 获取OAuth2服务器实例
func (m *Manager) GetOAuth2Server() *oauth2.OAuth2Server {
	return m.oauth2Server
}
