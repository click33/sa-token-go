package manager

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/click33/sa-token-go/core/adapter"
	"github.com/click33/sa-token-go/core/config"
	"github.com/click33/sa-token-go/core/oauth2"
	"github.com/click33/sa-token-go/core/security"
	"github.com/click33/sa-token-go/core/session"
	"github.com/click33/sa-token-go/core/token"
)

// TokenInfo Token信息
type TokenInfo struct {
	LoginID    string `json:"loginId"`
	Device     string `json:"device"`
	CreateTime int64  `json:"createTime"`
	ActiveTime int64  `json:"activeTime"` // 最后活跃时间
	Tag        string `json:"tag,omitempty"`
}

// Manager 认证管理器
type Manager struct {
	storage        adapter.Storage
	config         *config.Config
	generator      *token.Generator
	prefix         string
	nonceManager   *security.NonceManager
	refreshManager *security.RefreshTokenManager
	oauth2Server   *oauth2.OAuth2Server
}

// NewManager 创建管理器
func NewManager(storage adapter.Storage, cfg *config.Config) *Manager {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	return &Manager{
		storage:        storage,
		config:         cfg,
		generator:      token.NewGenerator(cfg),
		prefix:         "satoken:",
		nonceManager:   security.NewNonceManager(storage, 5*time.Minute),
		refreshManager: security.NewRefreshTokenManager(storage, cfg),
		oauth2Server:   oauth2.NewOAuth2Server(storage),
	}
}

// ============ 登录认证 ============

// Login 登录，返回Token
func (m *Manager) Login(loginID string, device ...string) (string, error) {
	deviceType := "default"
	if len(device) > 0 {
		deviceType = device[0]
	}

	// 检查是否被封禁
	if m.IsDisable(loginID) {
		return "", fmt.Errorf("account is disabled")
	}

	// 如果不允许并发登录，先踢掉旧的
	if !m.config.IsConcurrent {
		m.kickout(loginID, deviceType)
	}

	// 生成Token
	tokenValue, err := m.generator.Generate(loginID, deviceType)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// 计算过期时间
	var expiration time.Duration
	if m.config.Timeout > 0 {
		expiration = time.Duration(m.config.Timeout) * time.Second
	}

	now := time.Now().Unix()
	// 保存Token信息
	tokenInfo := &TokenInfo{
		LoginID:    loginID,
		Device:     deviceType,
		CreateTime: now,
		ActiveTime: now,
	}

	if err := m.saveTokenInfo(tokenValue, tokenInfo, expiration); err != nil {
		return "", err
	}

	// 保存账号-Token映射
	accountKey := m.getAccountKey(loginID, deviceType)
	if err := m.storage.Set(accountKey, tokenValue, expiration); err != nil {
		return "", fmt.Errorf("failed to save account mapping: %w", err)
	}

	// 创建Session
	sess := session.NewSession(loginID, m.storage, m.prefix)
	sess.Set("loginId", loginID)
	sess.Set("device", deviceType)
	sess.Set("loginTime", now)

	return tokenValue, nil
}

// LoginByToken 使用指定Token登录（用于token无感刷新）
func (m *Manager) LoginByToken(loginID string, tokenValue string, device ...string) error {
	deviceType := "default"
	if len(device) > 0 {
		deviceType = device[0]
	}

	var expiration time.Duration
	if m.config.Timeout > 0 {
		expiration = time.Duration(m.config.Timeout) * time.Second
	}

	now := time.Now().Unix()
	tokenInfo := &TokenInfo{
		LoginID:    loginID,
		Device:     deviceType,
		CreateTime: now,
		ActiveTime: now,
	}

	if err := m.saveTokenInfo(tokenValue, tokenInfo, expiration); err != nil {
		return err
	}

	accountKey := m.getAccountKey(loginID, deviceType)
	return m.storage.Set(accountKey, tokenValue, expiration)
}

// Logout 登出
func (m *Manager) Logout(loginID string, device ...string) error {
	deviceType := "default"
	if len(device) > 0 {
		deviceType = device[0]
	}

	accountKey := m.getAccountKey(loginID, deviceType)
	tokenValue, err := m.storage.Get(accountKey)
	if err != nil || tokenValue == nil {
		return nil // 已经登出
	}

	// 删除Token
	tokenKey := m.getTokenKey(tokenValue.(string))
	m.storage.Delete(tokenKey)

	// 删除账号映射
	m.storage.Delete(accountKey)

	return nil
}

// LogoutByToken 根据Token登出
func (m *Manager) LogoutByToken(tokenValue string) error {
	tokenKey := m.getTokenKey(tokenValue)
	m.storage.Delete(tokenKey)
	return nil
}

// kickout 踢人下线
func (m *Manager) kickout(loginID string, device string) error {
	accountKey := m.getAccountKey(loginID, device)
	tokenValue, err := m.storage.Get(accountKey)
	if err != nil || tokenValue == nil {
		return nil
	}

	tokenKey := m.getTokenKey(tokenValue.(string))
	return m.storage.Delete(tokenKey)
}

// Kickout 踢人下线（公开方法）
func (m *Manager) Kickout(loginID string, device ...string) error {
	deviceType := "default"
	if len(device) > 0 {
		deviceType = device[0]
	}
	return m.kickout(loginID, deviceType)
}

// ============ Token验证 ============

// IsLogin 检查是否登录
func (m *Manager) IsLogin(tokenValue string) bool {
	if tokenValue == "" {
		return false
	}

	tokenKey := m.getTokenKey(tokenValue)
	exists := m.storage.Exists(tokenKey)
	if !exists {
		return false
	}

	// 更新活跃时间并检查活跃超时
	if m.config.ActiveTimeout > 0 {
		info, _ := m.getTokenInfo(tokenValue)
		if info != nil {
			elapsed := time.Now().Unix() - info.ActiveTime
			if elapsed > m.config.ActiveTimeout {
				m.LogoutByToken(tokenValue)
				return false
			}
		}
	}

	// ✨ 异步自动续期（提高性能）
	if m.config.AutoRenew && m.config.Timeout > 0 {
		go func() {
			expiration := time.Duration(m.config.Timeout) * time.Second

			// 延长Token存储的过期时间
			m.storage.Expire(tokenKey, expiration)

			// 更新活跃时间
			info, _ := m.getTokenInfo(tokenValue)
			if info != nil {
				info.ActiveTime = time.Now().Unix()
				m.saveTokenInfo(tokenValue, info, expiration)
			}
		}()
	}

	return true
}

// CheckLogin 检查登录（未登录抛出错误）
func (m *Manager) CheckLogin(tokenValue string) error {
	if !m.IsLogin(tokenValue) {
		return fmt.Errorf("not login")
	}
	return nil
}

// GetLoginID 根据Token获取登录ID
func (m *Manager) GetLoginID(tokenValue string) (string, error) {
	if !m.IsLogin(tokenValue) {
		return "", fmt.Errorf("not login")
	}

	info, err := m.getTokenInfo(tokenValue)
	if err != nil {
		return "", err
	}

	return info.LoginID, nil
}

// GetLoginIDNotCheck 获取登录ID（不检查Token是否有效）
func (m *Manager) GetLoginIDNotCheck(tokenValue string) (string, error) {
	info, err := m.getTokenInfo(tokenValue)
	if err != nil {
		return "", err
	}
	return info.LoginID, nil
}

// GetTokenValue 根据登录ID获取Token
func (m *Manager) GetTokenValue(loginID string, device ...string) (string, error) {
	deviceType := "default"
	if len(device) > 0 {
		deviceType = device[0]
	}

	accountKey := m.getAccountKey(loginID, deviceType)
	tokenValue, err := m.storage.Get(accountKey)
	if err != nil || tokenValue == nil {
		return "", fmt.Errorf("token not found for login id: %s", loginID)
	}

	return tokenValue.(string), nil
}

// GetTokenInfo 获取Token信息
func (m *Manager) GetTokenInfo(tokenValue string) (*TokenInfo, error) {
	return m.getTokenInfo(tokenValue)
}

// ============ 账号封禁 ============

// Disable 封禁账号
func (m *Manager) Disable(loginID string, duration time.Duration) error {
	key := m.prefix + "disable:" + loginID
	return m.storage.Set(key, "1", duration)
}

// Untie 解封账号
func (m *Manager) Untie(loginID string) error {
	key := m.prefix + "disable:" + loginID
	return m.storage.Delete(key)
}

// IsDisable 检查账号是否被封禁
func (m *Manager) IsDisable(loginID string) bool {
	key := m.prefix + "disable:" + loginID
	return m.storage.Exists(key)
}

// GetDisableTime 获取账号剩余封禁时间（秒）
func (m *Manager) GetDisableTime(loginID string) (int64, error) {
	key := m.prefix + "disable:" + loginID
	ttl, err := m.storage.TTL(key)
	if err != nil {
		return -2, err
	}
	return int64(ttl.Seconds()), nil
}

// ============ Session管理 ============

// GetSession 获取Session
func (m *Manager) GetSession(loginID string) (*session.Session, error) {
	sess, err := session.Load(loginID, m.storage, m.prefix)
	if err != nil {
		sess = session.NewSession(loginID, m.storage, m.prefix)
	}
	return sess, nil
}

// GetSessionByToken 根据Token获取Session
func (m *Manager) GetSessionByToken(tokenValue string) (*session.Session, error) {
	loginID, err := m.GetLoginID(tokenValue)
	if err != nil {
		return nil, err
	}
	return m.GetSession(loginID)
}

// DeleteSession 删除Session
func (m *Manager) DeleteSession(loginID string) error {
	sess, err := m.GetSession(loginID)
	if err != nil {
		return err
	}
	return sess.Destroy()
}

// ============ 权限验证 ============

// SetPermissions 设置权限
func (m *Manager) SetPermissions(loginID string, permissions []string) error {
	sess, err := m.GetSession(loginID)
	if err != nil {
		return err
	}
	return sess.Set("permissions", permissions)
}

// GetPermissions 获取权限列表
func (m *Manager) GetPermissions(loginID string) ([]string, error) {
	sess, err := m.GetSession(loginID)
	if err != nil {
		return nil, err
	}

	perms, exists := sess.Get("permissions")
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

// matchPermission 权限匹配（支持通配符）
func (m *Manager) matchPermission(pattern, permission string) bool {
	if pattern == "*" || pattern == permission {
		return true
	}

	// 支持通配符，例如 user:* 匹配 user:add, user:delete等
	if strings.HasSuffix(pattern, ":*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(permission, prefix)
	}

	// 支持 user:*:view 这样的模式
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, ":")
		permParts := strings.Split(permission, ":")
		if len(parts) != len(permParts) {
			return false
		}
		for i, part := range parts {
			if part != "*" && part != permParts[i] {
				return false
			}
		}
		return true
	}

	return false
}

// ============ 角色验证 ============

// SetRoles 设置角色
func (m *Manager) SetRoles(loginID string, roles []string) error {
	sess, err := m.GetSession(loginID)
	if err != nil {
		return err
	}
	return sess.Set("roles", roles)
}

// GetRoles 获取角色列表
func (m *Manager) GetRoles(loginID string) ([]string, error) {
	sess, err := m.GetSession(loginID)
	if err != nil {
		return nil, err
	}

	roles, exists := sess.Get("roles")
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

// ============ Token标签 ============

// SetTokenTag 设置Token标签
func (m *Manager) SetTokenTag(tokenValue, tag string) error {
	info, err := m.getTokenInfo(tokenValue)
	if err != nil {
		return err
	}

	info.Tag = tag

	var expiration time.Duration
	if m.config.Timeout > 0 {
		expiration = time.Duration(m.config.Timeout) * time.Second
	}

	return m.saveTokenInfo(tokenValue, info, expiration)
}

// GetTokenTag 获取Token标签
func (m *Manager) GetTokenTag(tokenValue string) (string, error) {
	info, err := m.getTokenInfo(tokenValue)
	if err != nil {
		return "", err
	}
	return info.Tag, nil
}

// ============ 会话查询 ============

// GetTokenValueListByLoginID 获取指定账号的所有Token
func (m *Manager) GetTokenValueListByLoginID(loginID string) ([]string, error) {
	pattern := m.prefix + "account:" + loginID + ":*"
	keys, err := m.storage.Keys(pattern)
	if err != nil {
		return nil, err
	}

	tokens := make([]string, 0)
	for _, key := range keys {
		value, err := m.storage.Get(key)
		if err == nil && value != nil {
			tokens = append(tokens, value.(string))
		}
	}

	return tokens, nil
}

// GetSessionCountByLoginID 获取指定账号的Session数量
func (m *Manager) GetSessionCountByLoginID(loginID string) (int, error) {
	tokens, err := m.GetTokenValueListByLoginID(loginID)
	if err != nil {
		return 0, err
	}
	return len(tokens), nil
}

// ============ 辅助方法 ============

// getTokenKey 获取Token存储键
func (m *Manager) getTokenKey(tokenValue string) string {
	return m.prefix + "token:" + tokenValue
}

// getAccountKey 获取账号存储键
func (m *Manager) getAccountKey(loginID, device string) string {
	return m.prefix + "account:" + loginID + ":" + device
}

// saveTokenInfo 保存Token信息
func (m *Manager) saveTokenInfo(tokenValue string, info *TokenInfo, expiration time.Duration) error {
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal token info: %w", err)
	}

	tokenKey := m.getTokenKey(tokenValue)
	return m.storage.Set(tokenKey, string(data), expiration)
}

// getTokenInfo 获取Token信息
func (m *Manager) getTokenInfo(tokenValue string) (*TokenInfo, error) {
	tokenKey := m.getTokenKey(tokenValue)
	data, err := m.storage.Get(tokenKey)
	if err != nil || data == nil {
		return nil, fmt.Errorf("token not found")
	}

	var info TokenInfo
	if err := json.Unmarshal([]byte(data.(string)), &info); err != nil {
		return nil, fmt.Errorf("invalid token data: %w", err)
	}

	return &info, nil
}

// toStringSlice 将interface{}转换为[]string
func (m *Manager) toStringSlice(v interface{}) []string {
	switch val := v.(type) {
	case []string:
		return val
	case []interface{}:
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

// GetConfig 获取配置
func (m *Manager) GetConfig() *config.Config {
	return m.config
}

// GetStorage 获取存储
func (m *Manager) GetStorage() adapter.Storage {
	return m.storage
}

func (m *Manager) GenerateNonce() (string, error) {
	return m.nonceManager.Generate()
}

func (m *Manager) VerifyNonce(nonce string) bool {
	return m.nonceManager.Verify(nonce)
}

func (m *Manager) LoginWithRefreshToken(loginID, device string) (*security.RefreshTokenInfo, error) {
	return m.refreshManager.GenerateTokenPair(loginID, device)
}

func (m *Manager) RefreshAccessToken(refreshToken string) (*security.RefreshTokenInfo, error) {
	return m.refreshManager.RefreshAccessToken(refreshToken)
}

func (m *Manager) RevokeRefreshToken(refreshToken string) error {
	return m.refreshManager.RevokeRefreshToken(refreshToken)
}

func (m *Manager) GetOAuth2Server() *oauth2.OAuth2Server {
	return m.oauth2Server
}
