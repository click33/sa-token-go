// @Author daixk 2026/1/22 17:33:00
package manager

import (
	"context"
	"errors"
	"fmt"
	"github.com/click33/sa-token-go/core/listener"
	"github.com/click33/sa-token-go/core/nonce"
	"github.com/click33/sa-token-go/core/oauth2"
	"strings"
	"time"

	djson "github.com/click33/sa-token-go/com/codec/json"
	"github.com/click33/sa-token-go/com/generator/sgenerator"
	"github.com/click33/sa-token-go/com/log/nop"
	"github.com/click33/sa-token-go/com/pool/ants"
	"github.com/click33/sa-token-go/com/storage/memory"
	"github.com/click33/sa-token-go/core/adapter"
	"github.com/click33/sa-token-go/core/config"
	"github.com/click33/sa-token-go/core/serror"
	"github.com/click33/sa-token-go/core/utils"
)

// NewManager creates a new Manager instance with the provided components. NewManager 创建一个新的 Manager 实例,使用提供的组件。
func NewManager(
	cfg *config.Config,
	generator adapter.Generator,
	storage adapter.Storage,
	serializer adapter.Codec,
	logger adapter.Log,
	pool adapter.Pool,
	CustomPermissionListFunc, CustomRoleListFunc func(loginID, authType string) ([]string, error),
	CustomPermissionListExtFunc, CustomRoleListExtFunc func(loginID, device, deviceId, authType string) ([]string, error),
) *Manager {

	// Use default config if cfg is nil cfg 为 nil 时使用默认配置
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Create default token generator if generator is nil generator 为 nil 时创建默认 Token 生成器
	if generator == nil {
		generator = sgenerator.NewGenerator(cfg.Timeout, cfg.JwtSecretKey, cfg.TokenStyle)
	}

	// Use memory storage if storage is nil storage 为 nil 时使用内存存储
	if storage == nil {
		storage = memory.NewStorage()
	}

	// Use JSON serializer if serializer is nil serializer 为 nil 时使用 JSON 序列化器
	if serializer == nil {
		serializer = djson.NewJSONSerializer()
	}

	// Use nop logger if logger is nil logger 为 nil 时使用空日志记录器
	if logger == nil {
		logger = nop.NewNopLogger()
	}

	// Create default goroutine pool if AutoRenew is enabled and pool is nil 启用自动续期且 pool 为 nil 时使用默认协程池
	if cfg.AutoRenew && pool == nil {
		pool = ants.NewRenewPoolManagerWithDefaultConfig()
	}

	// Return initialized Manager instance 返回初始化完成的 Manager 实例
	return &Manager{
		config:                      cfg,
		generator:                   generator,
		storage:                     storage,
		serializer:                  serializer,
		logger:                      logger,
		pool:                        pool,
		nonceManager:                nonce.NewNonceManager(cfg.AuthType, cfg.KeyPrefix, storage, nonce.DefaultNonceTTL),
		oauth2Manager:               oauth2.NewOAuth2Server(cfg.AuthType, cfg.KeyPrefix, storage, serializer),
		eventManager:                listener.NewManager(logger),
		CustomPermissionListFunc:    CustomPermissionListFunc,
		CustomRoleListFunc:          CustomRoleListFunc,
		CustomPermissionListExtFunc: CustomPermissionListExtFunc,
		CustomRoleListExtFunc:       CustomRoleListExtFunc,
	}
}

// CloseManager closes the manager and releases all resources. CloseManager 关闭管理器并释放所有资源。
func (m *Manager) CloseManager() {
	// Wait for all async events to complete 等待所有异步事件完成
	if m.eventManager != nil {
		m.eventManager.Wait()
	}

	// Flush and close logger if it implements LogControl interface 若日志记录器实现了 LogControl 接口则执行 Flush 和 Close
	if logControl, ok := m.logger.(adapter.LogControl); ok {
		logControl.Flush()
		logControl.Close()
	}

	// Safely stop goroutine pool and set to nil 安全关闭协程池并置空
	if m.pool != nil {
		m.pool.Stop()
		m.pool = nil
	}
}

// Login performs user login and returns a token. Login 执行用户登录并返回 token。
func (m *Manager) Login(ctx context.Context, loginID string, deviceAndDeviceId ...string) (string, error) {
	return m.LoginWithTimeout(ctx, loginID, 0, deviceAndDeviceId...)
}

// LoginWithTimeout performs user login with a custom token timeout and returns a token. LoginWithTimeout 执行用户登录并返回 token，使用指定的过期时间（0 或负数则使用全局配置）。
func (m *Manager) LoginWithTimeout(ctx context.Context, loginID string, timeout time.Duration, deviceAndDeviceId ...string) (string, error) {
	if loginID == "" {
		return "", serror.ErrIDIsEmpty
	}

	if m.isDisable(ctx, loginID) {
		return "", serror.ErrAccountDisabled
	}

	device, deviceId := m.getDeviceAndDeviceId(deviceAndDeviceId...)

	// Load existing session 尝试加载现有 session
	sess, err := m.loginGetSession(ctx, loginID)
	if err != nil {
		return "", err
	}

	// Handle concurrency strategy 处理并发策略
	if sess != nil {
		token, handled, handleErr := m.handleConcurrency(ctx, sess, loginID, device)
		if handleErr != nil {
			return "", handleErr
		}
		if handled {
			if token != "" {
				return token, nil // 复用 token
			}
		}
	}

	// Generate new token 生成新 token
	token, err := m.generator.Generate(loginID, device, deviceId)
	if err != nil {
		return "", err
	}

	// Record create time 记录创建时间
	createTime := time.Now().Unix()

	// Get or create session 获取或创建 session
	sess, err = m.getSession(ctx, loginID, true)
	if err != nil {
		return "", err
	}

	// Increase history terminal count 递增历史终端计数
	sess.HistoryTerminalCount++

	// Append terminal info 添加终端信息
	sess.TerminalInfos = append(sess.TerminalInfos, TerminalInfo{
		Token:      token,
		LoginID:    loginID,
		Device:     device,
		DeviceId:   deviceId,
		CreateTime: createTime,
		Index:      sess.HistoryTerminalCount, // 设置历史登录顺序索引
	})

	// Calculate expiration duration 计算过期时长
	expiration := m.getExpiration()
	if timeout > 0 {
		expiration = timeout
	}

	// Save session 保存 session
	if err = m.saveToStorage(ctx, m.getSessionKey(loginID), *sess, expiration); err != nil {
		return "", err
	}

	// Save token info 保存 token info
	if err = m.saveToStorage(ctx, m.getTokenKey(token), TokenInfo{
		AuthType:   m.config.AuthType,
		LoginID:    loginID,
		Device:     device,
		DeviceId:   deviceId,
		CreateTime: createTime,
	}, expiration); err != nil {
		return "", err
	}

	// Initialize token metadata 初始化 token 元数据
	if m.config.RenewInterval > 0 {
		if err = m.storage.Set(ctx, m.getRenewKey(token), time.Now().Unix(), time.Duration(m.config.RenewInterval)*time.Second); err != nil {
			return "", fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
		}
	}
	if m.config.ActiveTimeout > 0 {
		if err = m.storage.Set(ctx, m.getActiveKey(token), time.Now().Unix(), expiration); err != nil {
			return "", fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
		}
	}

	// Trigger login event 触发登录事件
	m.triggerEvent(listener.EventLogin, loginID, device, deviceId, token, nil)

	return token, nil
}

// LoginByToken performs login renewal based on an existing token. LoginByToken 根据 Token 续期登录。
func (m *Manager) LoginByToken(ctx context.Context, tokenValue string) error {
	if tokenValue == "" {
		return serror.ErrInvalidToken
	}

	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err != nil {
		return err
	}

	// Check account disable status 检查账号是否被封禁
	if m.isDisable(ctx, tokenInfo.LoginID) {
		return serror.ErrAccountDisabled
	}

	sess, err := m.getSession(ctx, tokenInfo.LoginID)
	if err != nil {
		return err
	}

	// Validate token in session terminals 验证 token 是否在 session 的 TerminalInfos 中
	found := false
	for _, ti := range sess.TerminalInfos {
		if ti.Token == tokenValue {
			found = true
			break
		}
	}
	if !found {
		return serror.ErrInvalidToken
	}

	// Renew token and session asynchronously 异步续期 Token 和 Session
	renewFunc := func() {
		sessionKey := m.getSessionKey(tokenInfo.LoginID)
		tokenKey := m.getTokenKey(tokenValue)

		// Renew session 续期 session
		if err := m.storage.Expire(context.Background(), sessionKey, m.getExpiration()); err != nil {
			m.logger.Errorf("LoginByToken: failed to expire session for loginID=%s, error=%v", tokenInfo.LoginID, err)
		}
		// Renew token 续期 Token
		if err := m.storage.Expire(context.Background(), tokenKey, m.getExpiration()); err != nil {
			m.logger.Errorf("LoginByToken: failed to expire token for token=%s, error=%v", tokenValue, err)
		}

		// Update metadata 更新 metadata
		if m.config.RenewInterval > 0 {
			if err := m.storage.Set(context.Background(), m.getRenewKey(tokenValue), time.Now().Unix(), time.Duration(m.config.RenewInterval)*time.Second); err != nil {
				m.logger.Errorf("LoginByToken: failed to set renew key for token=%s, error=%v", tokenValue, err)
			}
		}
		if m.config.ActiveTimeout > 0 {
			if err := m.storage.Set(context.Background(), m.getActiveKey(tokenValue), time.Now().Unix(), m.getExpiration()); err != nil {
				m.logger.Errorf("LoginByToken: failed to set active key for token=%s, error=%v", tokenValue, err)
			}
		}

		// Trigger renew event 触发续期事件
		m.triggerEvent(listener.EventRenew, tokenInfo.LoginID, tokenInfo.Device, tokenInfo.DeviceId, tokenValue, nil)
	}

	if m.pool != nil {
		_ = m.pool.Submit(renewFunc)
	} else {
		go renewFunc()
	}

	return nil
}

// Logout logs out a user by token. Logout 根据 Token 登出用户。
func (m *Manager) Logout(ctx context.Context, tokenValue string) error {
	if tokenValue == "" {
		return serror.ErrInvalidToken
	}

	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err != nil {
		return err
	}

	return m.logoutTerminals(ctx, tokenInfo.LoginID, func(sess *Session) []TerminalInfo {
		if info, ok := sess.removeTerminalByToken(tokenValue); ok {
			return []TerminalInfo{info}
		}
		return nil
	})
}

// LogoutByDevice logs out all terminals of a specific device type. LogoutByDevice 根据设备类型登出所有该设备的终端。
func (m *Manager) LogoutByDevice(ctx context.Context, loginID string, device string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}

	return m.logoutTerminals(ctx, loginID, func(sess *Session) []TerminalInfo {
		return sess.removeTerminalByDevice(device)
	})
}

// LogoutByDeviceAndDeviceId logs out a user by device type and device ID. LogoutByDeviceAndDeviceId 根据设备类型和设备ID登出用户。
func (m *Manager) LogoutByDeviceAndDeviceId(ctx context.Context, loginID string, deviceAndDeviceId ...string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}

	device, deviceId := m.getDeviceAndDeviceId(deviceAndDeviceId...)
	return m.logoutTerminals(ctx, loginID, func(sess *Session) []TerminalInfo {
		return sess.removeTerminalByDeviceAndDeviceId(device, deviceId)
	})
}

// LogoutByLoginID logs out all terminals for the specified loginID. LogoutByLoginID 登出指定 loginID 的所有终端。
func (m *Manager) LogoutByLoginID(ctx context.Context, loginID string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}

	return m.logoutTerminals(ctx, loginID, func(sess *Session) []TerminalInfo {
		return sess.removeAllTerminals()
	})
}

// Kickout kicks out a user by token. Kickout 根据 Token 踢人下线。
func (m *Manager) Kickout(ctx context.Context, tokenValue string) error {
	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err != nil {
		return err
	}

	return m.processTerminals(ctx, tokenInfo.LoginID, func(sess *Session) []TerminalInfo {
		if info, ok := sess.removeTerminalByToken(tokenValue); ok {
			return []TerminalInfo{info}
		}
		return nil
	}, TokenStateKickOut)
}

// KickoutByDevice kicks out all terminals of a specific device type. KickoutByDevice 根据设备类型踢人下线（踢掉该设备类型的所有终端）。
func (m *Manager) KickoutByDevice(ctx context.Context, loginID string, device string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}

	return m.processTerminals(ctx, loginID, func(sess *Session) []TerminalInfo {
		return sess.removeTerminalByDevice(device)
	}, TokenStateKickOut)
}

// KickoutByDeviceAndDeviceId kicks out a user by device type and device ID. KickoutByDeviceAndDeviceId 根据设备类型和设备ID踢人下线。
func (m *Manager) KickoutByDeviceAndDeviceId(ctx context.Context, loginID string, deviceAndDeviceId ...string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}

	device, deviceId := m.getDeviceAndDeviceId(deviceAndDeviceId...)

	return m.processTerminals(ctx, loginID, func(sess *Session) []TerminalInfo {
		return sess.removeTerminalByDeviceAndDeviceId(device, deviceId)
	}, TokenStateKickOut)
}

// KickoutByLoginID kicks out all terminals for the specified loginID. KickoutByLoginID 踢出指定 loginID 的所有终端。
func (m *Manager) KickoutByLoginID(ctx context.Context, loginID string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}

	return m.processTerminals(ctx, loginID, func(sess *Session) []TerminalInfo {
		return sess.removeAllTerminals()
	}, TokenStateKickOut)
}

// Replace replaces a user session by token. Replace 根据 Token 顶人下线。
func (m *Manager) Replace(ctx context.Context, tokenValue string) error {
	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err != nil {
		return err
	}

	return m.processTerminals(ctx, tokenInfo.LoginID, func(sess *Session) []TerminalInfo {
		if info, ok := sess.removeTerminalByToken(tokenValue); ok {
			return []TerminalInfo{info}
		}
		return nil
	}, TokenStateReplaced)
}

// ReplaceByDevice replaces all terminals of a specific device type. ReplaceByDevice 根据设备类型顶人下线（顶掉该设备类型的所有终端）。
func (m *Manager) ReplaceByDevice(ctx context.Context, loginID string, device string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}

	return m.processTerminals(ctx, loginID, func(sess *Session) []TerminalInfo {
		return sess.removeTerminalByDevice(device)
	}, TokenStateReplaced)
}

// ReplaceByDeviceAndDeviceId replaces a user session by device type and device ID. ReplaceByDeviceAndDeviceId 根据设备类型和设备ID顶人下线。
func (m *Manager) ReplaceByDeviceAndDeviceId(ctx context.Context, loginID string, deviceAndDeviceId ...string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}

	device, deviceId := m.getDeviceAndDeviceId(deviceAndDeviceId...)
	return m.processTerminals(ctx, loginID, func(sess *Session) []TerminalInfo {
		return sess.removeTerminalByDeviceAndDeviceId(device, deviceId)
	}, TokenStateReplaced)
}

// ReplaceByLoginID replaces all terminals for the specified loginID. ReplaceByLoginID 顶替指定 loginID 的所有终端。
func (m *Manager) ReplaceByLoginID(ctx context.Context, loginID string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}

	return m.processTerminals(ctx, loginID, func(sess *Session) []TerminalInfo {
		return sess.removeAllTerminals()
	}, TokenStateReplaced)
}

// IsLogin checks if a user is logged in. IsLogin 检查用户是否登录。
func (m *Manager) IsLogin(ctx context.Context, tokenValue string) bool {
	return m.checkLoginInternal(ctx, tokenValue) == nil
}

// CheckLogin checks if a user is logged in and returns an error if not. CheckLogin 检查用户是否登录，如果未登录则返回错误。
func (m *Manager) CheckLogin(ctx context.Context, tokenValue string) error {
	return m.checkLoginInternal(ctx, tokenValue)
}

// GetLoginID retrieves the login ID from a token. GetLoginID 根据 Token 获取登录 ID。
func (m *Manager) GetLoginID(ctx context.Context, tokenValue string) (string, error) {
	// Get tokenInfo 获取 tokenInfo
	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err != nil {
		return "", err
	}

	return tokenInfo.LoginID, nil
}

// GetTokenInfo retrieves token information. GetTokenInfo 根据 Token 获取 TokenInfo 信息。
func (m *Manager) GetTokenInfo(ctx context.Context, tokenValue string) (*TokenInfo, error) {
	return m.getTokenInfo(ctx, tokenValue)
}

// GetDevice retrieves the device type for a token. GetDevice 获取 Token 的设备类型。
func (m *Manager) GetDevice(ctx context.Context, tokenValue string) (string, error) {
	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err != nil {
		return "", err
	}
	return tokenInfo.Device, nil
}

// GetDeviceId retrieves the device ID for a token. GetDeviceId 获取 Token 的设备 ID。
func (m *Manager) GetDeviceId(ctx context.Context, tokenValue string) (string, error) {
	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err != nil {
		return "", err
	}
	return tokenInfo.DeviceId, nil
}

// GetTokenCreateTime retrieves the creation time for a token. GetTokenCreateTime 获取 Token 的创建时间戳。
func (m *Manager) GetTokenCreateTime(ctx context.Context, tokenValue string) (int64, error) {
	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err != nil {
		return 0, err
	}
	return tokenInfo.CreateTime, nil
}

// GetTokenTTL retrieves the remaining time-to-live for a token in seconds. GetTokenTTL 获取 Token 的剩余有效时间（秒）。
func (m *Manager) GetTokenTTL(ctx context.Context, tokenValue string) (int64, error) {
	ttl, err := m.storage.TTL(ctx, m.getTokenKey(tokenValue))
	if err != nil {
		return 0, fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
	}

	// Normalize TTL sentinel values 统一 TTL 哨兵值语义
	seconds := int64(ttl)
	switch {
	case seconds == -2:
		return -2, nil
	case seconds == -1:
		return -1, nil
	case seconds > 0:
		return int64(ttl.Seconds()), nil
	default:
		return 0, nil
	}
}

// Disable disables an account for a specified duration. Disable 封禁账号指定时长。
func (m *Manager) Disable(ctx context.Context, loginID string, duration time.Duration, reason ...string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}

	// Load session before disable 先尝试加载 Session（如果存储出错，在保存封禁信息前就返回，保证原子性）
	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		// Ignore missing session and return other storage errors 如果只是 session 不存在，不算错误；其他存储错误则返回
		if !errors.Is(err, serror.ErrSessionNotFound) {
			return err
		}
		// Continue disable when sess is nil 否则 sess == nil，继续执行封禁操作（幂等）
	}

	// Build and save disable info 构建并保存封禁信息
	disableInfo := DisableInfo{
		DisableTime: time.Now().Unix(),
	}
	if len(reason) > 0 && reason[0] != "" {
		disableInfo.DisableReason = reason[0]
	}

	if err = m.saveToStorage(ctx, m.getDisableKey(loginID), disableInfo, duration); err != nil {
		return err
	}

	// Delete session 删除 Session
	if err = m.storage.Delete(ctx, m.getSessionKey(loginID)); err != nil {
		return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
	}

	// Trigger session destroy event 触发销毁 Session 事件
	m.triggerEvent(listener.EventDestroySession, loginID, "", "", "", nil)

	// Clean related token data when terminals exist 如果有终端信息，清理所有相关 token 数据
	if sess != nil && len(sess.TerminalInfos) > 0 {
		tokens := make([]string, len(sess.TerminalInfos))
		tokenKeys := make([]string, len(sess.TerminalInfos))
		for i, info := range sess.TerminalInfos {
			tokens[i] = info.Token
			tokenKeys[i] = m.getTokenKey(info.Token)
		}

		// Delete primary token keys 删除主 token keys
		if err = m.storage.Delete(ctx, tokenKeys...); err != nil {
			return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
		}

		// Clean token metadata 清理附属 metadata（续期、活跃时间）
		if err = m.cleanTokenMetadata(ctx, tokens); err != nil {
			return err
		}
	}

	// Trigger disable event 触发封禁事件
	m.triggerEvent(listener.EventDisable, loginID, "", "", "", map[string]any{
		"reason":   disableInfo.DisableReason,
		"duration": duration.Seconds(),
	})

	return nil
}

// Untie removes the disable status from an account. Untie 解封账号。
func (m *Manager) Untie(ctx context.Context, loginID string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}

	if err := m.storage.Delete(ctx, m.getDisableKey(loginID)); err != nil {
		return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
	}

	// Trigger untie event 触发解禁事件
	m.triggerEvent(listener.EventUntie, loginID, "", "", "", nil)

	return nil
}

// IsDisable checks if an account is disabled. IsDisable 检查账号是否被封禁。
func (m *Manager) IsDisable(ctx context.Context, loginID string) bool {
	return m.isDisable(ctx, loginID)
}

// GetDisableInfo retrieves disable information for an account. GetDisableInfo 获取账号的封禁信息。
func (m *Manager) GetDisableInfo(ctx context.Context, loginID string) (*DisableInfo, error) {
	disableInfoData, err := m.storage.Get(ctx, m.getDisableKey(loginID))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
	}

	// Return explicit error when disable key is missing 如果 key 不存在（用户未被封禁），返回明确的错误
	if disableInfoData == nil {
		return nil, serror.ErrAccountNotDisabled
	}

	bytesData, err := utils.ToBytes(disableInfoData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", serror.ErrTypeConvert, err)
	}

	var disableInfo DisableInfo
	if err = m.serializer.Decode(bytesData, &disableInfo); err != nil {
		return nil, fmt.Errorf("%w: %v", serror.ErrSerializeFailed, err)
	}

	return &disableInfo, nil
}

// GetDisableTTL retrieves the remaining disable time for an account in seconds. GetDisableTTL 获取账号剩余封禁时间（秒）。 Returns: -2: account is not disabled (未封禁) -1: account is permanently disabled (永久封禁) >0: remaining seconds until unban (剩余封禁秒数)
func (m *Manager) GetDisableTTL(ctx context.Context, loginID string) (int64, error) {
	ttl, err := m.storage.TTL(ctx, m.getDisableKey(loginID))
	if err != nil {
		return 0, fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
	}

	// Explain TTL semantics 存储适配器返回 time.Duration 类型，直接转换为 int64 即可，标准 Redis TTL 语义：-2 key 不存在，-1 key 无过期时间，>0 剩余秒数
	seconds := int64(ttl)

	switch {
	case seconds == -2:
		return -2, nil // 未封禁
	case seconds == -1:
		return -1, nil // 永久封禁
	case seconds > 0:
		return int64(ttl.Seconds()), nil
	default:
		return 0, nil
	}
}

// DisableService disables a specific service for an account. DisableService 封禁账号的指定服务。
func (m *Manager) DisableService(ctx context.Context, loginID, service string, duration time.Duration, reason ...string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}
	if service == "" {
		return serror.ErrInvalidParam
	}

	info := ServiceDisableInfo{
		Service:     service,
		Level:       0,
		DisableTime: time.Now().Unix(),
	}
	if len(reason) > 0 && reason[0] != "" {
		info.DisableReason = reason[0]
	}

	if err := m.saveToStorage(ctx, m.getDisableServiceKey(loginID, service), info, duration); err != nil {
		return err
	}

	m.triggerEvent(listener.EventDisableService, loginID, "", "", "", map[string]any{
		listener.ExtraKeyService: service,
		"reason":                 info.DisableReason,
		"duration":               duration.Seconds(),
	})

	return nil
}

// DisableServiceLevel disables a specific service for an account with a level. DisableServiceLevel 封禁账号的指定服务并设置封禁等级。
func (m *Manager) DisableServiceLevel(ctx context.Context, loginID, service string, level int, duration time.Duration, reason ...string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}
	if service == "" {
		return serror.ErrInvalidParam
	}

	info := ServiceDisableInfo{
		Service:     service,
		Level:       level,
		DisableTime: time.Now().Unix(),
	}
	if len(reason) > 0 && reason[0] != "" {
		info.DisableReason = reason[0]
	}

	if err := m.saveToStorage(ctx, m.getDisableServiceKey(loginID, service), info, duration); err != nil {
		return err
	}

	m.triggerEvent(listener.EventDisableService, loginID, "", "", "", map[string]any{
		listener.ExtraKeyService: service,
		listener.ExtraKeyLevel:   level,
		"reason":                 info.DisableReason,
		"duration":               duration.Seconds(),
	})

	return nil
}

// UntieService removes the disable status of a specific service for an account. UntieService 解封账号的指定服务。
func (m *Manager) UntieService(ctx context.Context, loginID, service string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}
	if service == "" {
		return serror.ErrInvalidParam
	}

	if err := m.storage.Delete(ctx, m.getDisableServiceKey(loginID, service)); err != nil {
		return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
	}

	m.triggerEvent(listener.EventUntieService, loginID, "", "", "", map[string]any{
		listener.ExtraKeyService: service,
	})

	return nil
}

// IsDisableService checks if a specific service is disabled for an account. IsDisableService 检查账号的指定服务是否被封禁。
func (m *Manager) IsDisableService(ctx context.Context, loginID, service string) bool {
	if loginID == "" || service == "" {
		return false
	}
	return m.storage.Exists(ctx, m.getDisableServiceKey(loginID, service))
}

// IsDisableServiceLevel checks if a specific service is disabled at or above the given level. IsDisableServiceLevel 检查账号的指定服务是否达到指定封禁等级。
func (m *Manager) IsDisableServiceLevel(ctx context.Context, loginID, service string, level int) bool {
	info, err := m.GetDisableServiceInfo(ctx, loginID, service)
	if err != nil {
		return false
	}
	return info.Level >= level
}

// CheckDisableService checks if any of the specified services are disabled, returns error if disabled. CheckDisableService 校验账号的指定服务是否被封禁，被封禁则返回 error。
func (m *Manager) CheckDisableService(ctx context.Context, loginID string, services ...string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}
	for _, service := range services {
		if m.IsDisableService(ctx, loginID, service) {
			return fmt.Errorf("%w: service=%s", serror.ErrServiceDisabled, service)
		}
	}
	return nil
}

// CheckDisableServiceLevel checks if a service is disabled at or above the given level, returns error if so. CheckDisableServiceLevel 校验账号的指定服务是否达到指定封禁等级，达到则返回 error。
func (m *Manager) CheckDisableServiceLevel(ctx context.Context, loginID, service string, level int) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}
	if m.IsDisableServiceLevel(ctx, loginID, service, level) {
		return fmt.Errorf("%w: service=%s, level=%d", serror.ErrServiceDisabled, service, level)
	}
	return nil
}

// GetDisableServiceInfo retrieves the disable info for a specific service. GetDisableServiceInfo 获取账号指定服务的封禁信息。
func (m *Manager) GetDisableServiceInfo(ctx context.Context, loginID, service string) (*ServiceDisableInfo, error) {
	if loginID == "" {
		return nil, serror.ErrIDIsEmpty
	}
	if service == "" {
		return nil, serror.ErrInvalidParam
	}

	data, err := m.storage.Get(ctx, m.getDisableServiceKey(loginID, service))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
	}
	if data == nil {
		return nil, serror.ErrServiceNotDisabled
	}

	bytesData, err := utils.ToBytes(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", serror.ErrTypeConvert, err)
	}

	var info ServiceDisableInfo
	if err = m.serializer.Decode(bytesData, &info); err != nil {
		return nil, fmt.Errorf("%w: %v", serror.ErrSerializeFailed, err)
	}

	return &info, nil
}

// GetDisableServiceTTL retrieves the remaining disable time for a specific service in seconds. GetDisableServiceTTL 获取账号指定服务的剩余封禁时间（秒）。
func (m *Manager) GetDisableServiceTTL(ctx context.Context, loginID, service string) (int64, error) {
	ttl, err := m.storage.TTL(ctx, m.getDisableServiceKey(loginID, service))
	if err != nil {
		return 0, fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
	}

	seconds := int64(ttl)
	switch {
	case seconds == -2:
		return -2, nil
	case seconds == -1:
		return -1, nil
	default:
		return int64(ttl.Seconds()), nil
	}
}

// GetSession retrieves session information for a login ID. GetSession 获取指定登录 ID 的会话信息。
func (m *Manager) GetSession(ctx context.Context, loginID string) (*Session, error) {
	if loginID == "" {
		return nil, serror.ErrIDIsEmpty
	}
	return m.getSession(ctx, loginID)
}

// GetSessionByToken retrieves session information by token. GetSessionByToken 通过 Token 值获取会话信息。
func (m *Manager) GetSessionByToken(ctx context.Context, tokenValue string) (*Session, error) {
	// Get tokenInfo 获取 tokenInfo
	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err != nil {
		return nil, err
	}

	return m.getSession(ctx, tokenInfo.LoginID)
}

// GetTokenValueListByLoginID retrieves all tokens for a login ID. GetTokenValueListByLoginID 获取指定登录 ID 的所有 Token。
func (m *Manager) GetTokenValueListByLoginID(ctx context.Context, loginID string, checkAlive ...bool) ([]string, error) {
	if loginID == "" {
		return nil, serror.ErrIDIsEmpty
	}

	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		// Return errors only for real storage failures 仅当存储层真正出错时才返回 error；session 不存在视为 nil
		if !errors.Is(err, serror.ErrSessionNotFound) {
			return nil, err
		}
		return []string{}, nil
	}
	if sess == nil {
		return []string{}, nil
	}

	return m.filterTokens(ctx, sess.TerminalInfos, len(checkAlive) > 0 && checkAlive[0])
}

// GetTokenValueListByDevice retrieves all tokens for a specific device type. GetTokenValueListByDevice 获取指定设备类型的所有 Token。
func (m *Manager) GetTokenValueListByDevice(ctx context.Context, loginID, device string, checkAlive ...bool) ([]string, error) {
	if loginID == "" {
		return []string{}, serror.ErrIDIsEmpty
	}
	if device == "" {
		return []string{}, serror.ErrInvalidToken
	}

	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		if !errors.Is(err, serror.ErrSessionNotFound) {
			return nil, err
		}
		return []string{}, nil
	}
	if sess == nil {
		return []string{}, nil
	}

	matched := sess.getTerminalsByDevice(device)
	return m.filterTokens(ctx, matched, len(checkAlive) > 0 && checkAlive[0])
}

// GetTokenValueListByDeviceAndDeviceId retrieves all tokens for a specific device type and device ID. GetTokenValueListByDeviceAndDeviceId 获取指定设备类型和设备 ID 的所有 Token。
func (m *Manager) GetTokenValueListByDeviceAndDeviceId(ctx context.Context, loginID, device, deviceId string, checkAlive ...bool) ([]string, error) {
	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		if !errors.Is(err, serror.ErrSessionNotFound) {
			return nil, err
		}
		return []string{}, nil
	}
	if sess == nil {
		return []string{}, nil
	}

	matched := sess.getTerminalsByDeviceAndDeviceId(device, deviceId)
	return m.filterTokens(ctx, matched, len(checkAlive) > 0 && checkAlive[0])
}

// GetOnlineTerminalCount retrieves the count of online terminals for a user. GetOnlineTerminalCount 获取用户的在线终端数量。
func (m *Manager) GetOnlineTerminalCount(ctx context.Context, loginID string) (int, error) {
	if loginID == "" {
		return 0, serror.ErrIDIsEmpty
	}

	tokens, err := m.GetTokenValueListByLoginID(ctx, loginID, true)
	if err != nil {
		return 0, err
	}
	return len(tokens), nil
}

// GetOnlineTerminalCountByDevice retrieves the count of online terminals for a specific device type. GetOnlineTerminalCountByDevice 获取用户在指定设备类型的在线终端数量。
func (m *Manager) GetOnlineTerminalCountByDevice(ctx context.Context, loginID, device string) (int, error) {
	tokens, err := m.GetTokenValueListByDevice(ctx, loginID, device, true)
	if err != nil {
		return 0, err
	}
	return len(tokens), nil
}

// GetOnlineTerminalCountByDeviceAndDeviceId retrieves the count of online terminals for a specific device type and device ID. GetOnlineTerminalCountByDeviceAndDeviceId 获取用户在指定设备类型和设备ID的在线终端数量。
func (m *Manager) GetOnlineTerminalCountByDeviceAndDeviceId(ctx context.Context, loginID, device, deviceId string) (int, error) {
	tokens, err := m.GetTokenValueListByDeviceAndDeviceId(ctx, loginID, device, deviceId, true)
	if err != nil {
		return 0, err
	}
	return len(tokens), nil
}

// GetTerminalListByLoginID retrieves all terminal info for a login ID, optionally filtered by device. GetTerminalListByLoginID 获取指定登录 ID 的所有终端信息列表，可选按设备类型过滤。
func (m *Manager) GetTerminalListByLoginID(ctx context.Context, loginID string, device ...string) ([]TerminalInfo, error) {
	if loginID == "" {
		return nil, serror.ErrIDIsEmpty
	}

	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		if errors.Is(err, serror.ErrSessionNotFound) {
			return []TerminalInfo{}, nil
		}
		return nil, err
	}
	if sess == nil {
		return []TerminalInfo{}, nil
	}

	if len(device) > 0 && device[0] != "" {
		return sess.getTerminalsByDevice(device[0]), nil
	}

	// Return copy to avoid external mutation 返回副本，避免外部修改影响内部数据
	result := make([]TerminalInfo, len(sess.TerminalInfos))
	copy(result, sess.TerminalInfos)
	return result, nil
}

// GetTerminalInfoByToken retrieves terminal info for a specific token. GetTerminalInfoByToken 根据 Token 获取终端详情。
func (m *Manager) GetTerminalInfoByToken(ctx context.Context, tokenValue string) (*TerminalInfo, error) {
	if tokenValue == "" {
		return nil, serror.ErrInvalidToken
	}

	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err != nil {
		return nil, err
	}

	sess, err := m.getSession(ctx, tokenInfo.LoginID)
	if err != nil {
		return nil, err
	}

	for _, ti := range sess.TerminalInfos {
		if ti.Token == tokenValue {
			return &ti, nil
		}
	}

	return nil, serror.ErrInvalidToken
}

// GetTokenValueByLoginID retrieves the latest token for a login ID, optionally filtered by device. GetTokenValueByLoginID 获取指定登录 ID 的最新 Token，可选按设备类型过滤。
func (m *Manager) GetTokenValueByLoginID(ctx context.Context, loginID string, device ...string) (string, error) {
	if loginID == "" {
		return "", serror.ErrIDIsEmpty
	}

	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		return "", err
	}

	if len(device) > 0 && device[0] != "" {
		if ti, ok := sess.getLatestTerminalByDevice(device[0]); ok {
			return ti.Token, nil
		}
		return "", serror.ErrInvalidToken
	}

	// Return latest token 返回最后一个（最新的）
	if len(sess.TerminalInfos) == 0 {
		return "", serror.ErrInvalidToken
	}
	return sess.TerminalInfos[len(sess.TerminalInfos)-1].Token, nil
}

// SearchTokenValue searches token keys by keyword with pagination. SearchTokenValue 根据关键词搜索 Token，支持分页。 keyword: 搜索关键词（模糊匹配），start: 起始索引，size: 返回数量（-1 返回全部）
func (m *Manager) SearchTokenValue(ctx context.Context, keyword string, start, size int) ([]string, error) {
	pattern := m.config.KeyPrefix + m.config.AuthType + "*" + keyword + "*"
	return m.searchKeys(ctx, pattern, start, size)
}

// SearchSessionId searches session keys by keyword with pagination. SearchSessionId 根据关键词搜索 Session ID，支持分页。 keyword: 搜索关键词（模糊匹配），start: 起始索引，size: 返回数量（-1 返回全部）
func (m *Manager) SearchSessionId(ctx context.Context, keyword string, start, size int) ([]string, error) {
	pattern := m.config.KeyPrefix + m.config.AuthType + SessionKeyPrefix + "*" + keyword + "*"
	return m.searchKeys(ctx, pattern, start, size)
}

// TerminalVisitor is a callback function for terminal traversal. TerminalVisitor 终端遍历回调函数。 Return false to stop traversal. 返回 false 停止遍历。
type TerminalVisitor func(terminal TerminalInfo) bool

// ForEachTerminal iterates over all terminals for a login ID and calls the visitor function. ForEachTerminal 遍历指定登录 ID 的所有终端，对每个终端调用回调函数。
func (m *Manager) ForEachTerminal(ctx context.Context, loginID string, visitor TerminalVisitor) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}

	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		if errors.Is(err, serror.ErrSessionNotFound) {
			return nil
		}
		return err
	}

	for _, ti := range sess.TerminalInfos {
		if !visitor(ti) {
			break
		}
	}
	return nil
}

// ForEachTerminalByDevice iterates over terminals filtered by device type. ForEachTerminalByDevice 遍历指定设备类型的终端。
func (m *Manager) ForEachTerminalByDevice(ctx context.Context, loginID, device string, visitor TerminalVisitor) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}

	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		if errors.Is(err, serror.ErrSessionNotFound) {
			return nil
		}
		return err
	}

	for _, ti := range sess.TerminalInfos {
		if ti.Device == device {
			if !visitor(ti) {
				break
			}
		}
	}
	return nil
}

// RenewTimeout manually renews the timeout of a token. RenewTimeout 手动续期指定 Token 的过期时间。
func (m *Manager) RenewTimeout(ctx context.Context, tokenValue string, timeout time.Duration) error {
	if tokenValue == "" {
		return serror.ErrInvalidToken
	}

	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err != nil {
		return err
	}

	// Renew token key 续期 token key
	if err = m.storage.Expire(ctx, m.getTokenKey(tokenValue), timeout); err != nil {
		return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
	}

	// Renew session key 续期 session key
	if err = m.storage.Expire(ctx, m.getSessionKey(tokenInfo.LoginID), timeout); err != nil {
		return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
	}

	// Trigger renew event 触发续期事件
	m.triggerEvent(listener.EventRenew, tokenInfo.LoginID, tokenInfo.Device, tokenInfo.DeviceId, tokenValue, map[string]any{
		"timeout": timeout.Seconds(),
	})

	return nil
}

// AddPermissions adds permissions to a user. AddPermissions 为用户添加权限。
func (m *Manager) AddPermissions(ctx context.Context, loginID string, permissions []string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}

	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		return err
	}

	sess.addPermissions(permissions...)
	err = m.saveToStorage(ctx, m.getSessionKey(loginID), *sess)
	if err != nil {
		return err
	}

	return nil
}

// AddPermissionsByToken adds permissions to a user by token. AddPermissionsByToken 根据 Token 为用户添加权限。
func (m *Manager) AddPermissionsByToken(ctx context.Context, tokenValue string, permissions []string) error {
	sess, err := m.GetSessionByToken(ctx, tokenValue)
	if err != nil {
		return err
	}

	// Add permissions 添加权限
	sess.addPermissions(permissions...)
	// Save session 保存 Session
	err = m.saveToStorage(ctx, m.getSessionKey(sess.LoginID), *sess)
	if err != nil {
		return err
	}

	return nil
}

// RemovePermissions removes permissions from a user. RemovePermissions 删除用户的指定权限。
func (m *Manager) RemovePermissions(ctx context.Context, loginID string, permissions []string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}

	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		return err
	}

	sess.removePermissions(permissions...)
	err = m.saveToStorage(ctx, m.getSessionKey(loginID), *sess)
	if err != nil {
		return err
	}

	return nil
}

// RemovePermissionsByToken removes permissions from a user by token. RemovePermissionsByToken 根据 Token 删除用户的指定权限。
func (m *Manager) RemovePermissionsByToken(ctx context.Context, tokenValue string, permissions []string) error {
	sess, err := m.GetSessionByToken(ctx, tokenValue)
	if err != nil {
		return err
	}

	// Remove permissions 删除权限
	sess.removePermissions(permissions...)
	// Save session 保存 Session
	err = m.saveToStorage(ctx, m.getSessionKey(sess.LoginID), *sess)
	if err != nil {
		return err
	}

	return nil
}

// GetPermissions retrieves the permission list for a user. GetPermissions 获取用户的权限列表。
func (m *Manager) GetPermissions(ctx context.Context, loginID string) ([]string, error) {
	if loginID == "" {
		return nil, serror.ErrIDIsEmpty
	}

	// Use custom permission list function 使用自定义权限列表获取函数
	if m.CustomPermissionListFunc != nil {
		return m.CustomPermissionListFunc(loginID, m.config.AuthType)
	}

	// Get session 获取 Session
	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		return nil, err
	}

	return sess.Permissions, nil
}

// GetPermissionsByToken retrieves the permission list by token. GetPermissionsByToken 根据 Token 获取权限列表。
func (m *Manager) GetPermissionsByToken(ctx context.Context, tokenValue string) ([]string, error) {
	// Get session 获取 Session
	sess, err := m.GetSessionByToken(ctx, tokenValue)
	if err != nil {
		return nil, err
	}

	return sess.Permissions, nil
}

// HasPermission checks if a user has a specific permission. HasPermission 检查用户是否拥有指定权限。
func (m *Manager) HasPermission(ctx context.Context, loginID string, permission string) bool {
	if loginID == "" {
		return false
	}

	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		m.logger.Errorf("HasPermission: failed to get session for loginID=%s, error=%v", loginID, err)
		return false
	}

	// Get permissions with Func then Session priority 获取权限列表（两级优先级：Func > Session）
	permissions := sess.Permissions
	if m.CustomPermissionListFunc != nil {
		customPerms, err := m.CustomPermissionListFunc(loginID, m.config.AuthType)
		if err == nil && customPerms != nil {
			permissions = customPerms
		}
	}

	hasPermission := false
	for _, p := range permissions {
		if m.matchPermission(p, permission) {
			hasPermission = true
			break
		}
	}

	// Trigger permission check event 触发权限检查事件
	m.triggerEvent(listener.EventPermissionCheck, loginID, "", "", "", map[string]any{
		listener.ExtraKeyPermission: permission,
		listener.ExtraKeyResult:     hasPermission,
	})

	return hasPermission
}

// HasPermissionByToken checks if a user has a specific permission by token. HasPermissionByToken 根据 Token 检查用户是否拥有指定权限。
func (m *Manager) HasPermissionByToken(ctx context.Context, tokenValue string, permission string) bool {
	sess, err := m.GetSessionByToken(ctx, tokenValue)
	if err != nil {
		return false
	}

	// Get device and deviceId 获取 device/deviceId
	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err != nil {
		m.logger.Errorf("HasPermissionByToken: failed to get token info for token=%s, error=%v", tokenValue, err)
		return false
	}
	device, deviceId := tokenInfo.Device, tokenInfo.DeviceId

	// Get permissions with Ext Func Session priority 获取权限列表（三级优先级：Ext > Func > Session）
	permissions := sess.Permissions
	if m.CustomPermissionListExtFunc != nil {
		customPerms, err := m.CustomPermissionListExtFunc(sess.LoginID, device, deviceId, m.config.AuthType)
		if err == nil && customPerms != nil {
			permissions = customPerms
		}
	} else if m.CustomPermissionListFunc != nil {
		customPerms, err := m.CustomPermissionListFunc(sess.LoginID, m.config.AuthType)
		if err == nil && customPerms != nil {
			permissions = customPerms
		}
	}

	hasPermission := false
	for _, p := range permissions {
		if m.matchPermission(p, permission) {
			hasPermission = true
			break
		}
	}

	// Trigger permission check event 触发权限检查事件
	m.triggerEvent(listener.EventPermissionCheck, sess.LoginID, device, deviceId, tokenValue, map[string]any{
		listener.ExtraKeyPermission: permission,
		listener.ExtraKeyResult:     hasPermission,
	})

	return hasPermission
}

// HasPermissionsAnd checks if a user has all specified permissions (AND logic). HasPermissionsAnd 检查用户是否拥有所有指定权限（AND 逻辑）。
func (m *Manager) HasPermissionsAnd(ctx context.Context, loginID string, permissions []string) bool {
	if loginID == "" {
		return false
	}

	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		m.logger.Errorf("HasPermissionsAnd: failed to get session for loginID=%s, error=%v", loginID, err)
		return false
	}

	// Get permissions with Func then Session priority 获取权限列表（两级优先级：Func > Session）
	permList := sess.Permissions
	if m.CustomPermissionListFunc != nil {
		customPerms, err := m.CustomPermissionListFunc(loginID, m.config.AuthType)
		if err == nil && customPerms != nil {
			permList = customPerms
		}
	}

	// Check each required permission 校验每一个必需权限
	hasAll := true
	for _, need := range permissions {
		if !m.hasPermissionInList(permList, need) {
			hasAll = false
			break
		}
	}

	// Trigger permission check event 触发权限检查事件
	m.triggerEvent(listener.EventPermissionCheck, loginID, "", "", "", map[string]any{
		listener.ExtraKeyPermissions: permissions,
		listener.ExtraKeyLogic:       listener.LogicAnd,
		listener.ExtraKeyResult:      hasAll,
	})

	return hasAll
}

// HasPermissionsAndByToken checks if a user has all specified permissions by token (AND logic). HasPermissionsAndByToken 根据 Token 检查用户是否拥有所有指定权限（AND 逻辑）。
func (m *Manager) HasPermissionsAndByToken(ctx context.Context, tokenValue string, permissions []string) bool {
	sess, err := m.GetSessionByToken(ctx, tokenValue)
	if err != nil {
		return false
	}

	// Get device and deviceId 获取 device/deviceId
	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err != nil {
		m.logger.Errorf("HasPermissionsAndByToken: failed to get token info for token=%s, error=%v", tokenValue, err)
		return false
	}
	device, deviceId := tokenInfo.Device, tokenInfo.DeviceId

	// Get permissions with Ext Func Session priority 获取权限列表（三级优先级：Ext > Func > Session）
	permList := sess.Permissions
	if m.CustomPermissionListExtFunc != nil {
		customPerms, err := m.CustomPermissionListExtFunc(sess.LoginID, device, deviceId, m.config.AuthType)
		if err == nil && customPerms != nil {
			permList = customPerms
		}
	} else if m.CustomPermissionListFunc != nil {
		customPerms, err := m.CustomPermissionListFunc(sess.LoginID, m.config.AuthType)
		if err == nil && customPerms != nil {
			permList = customPerms
		}
	}

	// Check each required permission 校验每一个必需权限
	hasAll := true
	for _, need := range permissions {
		if !m.hasPermissionInList(permList, need) {
			hasAll = false
			break
		}
	}

	// Trigger permission check event 触发权限检查事件
	m.triggerEvent(listener.EventPermissionCheck, sess.LoginID, device, deviceId, tokenValue, map[string]any{
		listener.ExtraKeyPermissions: permissions,
		listener.ExtraKeyLogic:       listener.LogicAnd,
		listener.ExtraKeyResult:      hasAll,
	})

	return hasAll
}

// HasPermissionsOr checks if a user has any of the specified permissions (OR logic). HasPermissionsOr 检查用户是否拥有任一指定权限（OR 逻辑）。
func (m *Manager) HasPermissionsOr(ctx context.Context, loginID string, permissions []string) bool {
	if loginID == "" {
		return false
	}

	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		m.logger.Errorf("HasPermissionsOr: failed to get session for loginID=%s, error=%v", loginID, err)
		return false
	}

	// Get permissions with Func then Session priority 获取权限列表（两级优先级：Func > Session）
	permList := sess.Permissions
	if m.CustomPermissionListFunc != nil {
		customPerms, err := m.CustomPermissionListFunc(loginID, m.config.AuthType)
		if err == nil && customPerms != nil {
			permList = customPerms
		}
	}

	// Pass on any matching permission 任一权限匹配即通过
	hasAny := false
	for _, need := range permissions {
		if m.hasPermissionInList(permList, need) {
			hasAny = true
			break
		}
	}

	// Trigger permission check event 触发权限检查事件
	m.triggerEvent(listener.EventPermissionCheck, loginID, "", "", "", map[string]any{
		listener.ExtraKeyPermissions: permissions,
		listener.ExtraKeyLogic:       listener.LogicOr,
		listener.ExtraKeyResult:      hasAny,
	})

	return hasAny
}

// HasPermissionsOrByToken checks if a user has any of the specified permissions by token (OR logic). HasPermissionsOrByToken 根据 Token 检查用户是否拥有任一指定权限（OR 逻辑）。
func (m *Manager) HasPermissionsOrByToken(ctx context.Context, tokenValue string, permissions []string) bool {
	sess, err := m.GetSessionByToken(ctx, tokenValue)
	if err != nil {
		return false
	}

	// Get device and deviceId 获取 device/deviceId
	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err != nil {
		m.logger.Errorf("HasPermissionsOrByToken: failed to get token info for token=%s, error=%v", tokenValue, err)
		return false
	}
	device, deviceId := tokenInfo.Device, tokenInfo.DeviceId

	// Get permissions with Ext Func Session priority 获取权限列表（三级优先级：Ext > Func > Session）
	permList := sess.Permissions
	if m.CustomPermissionListExtFunc != nil {
		customPerms, err := m.CustomPermissionListExtFunc(sess.LoginID, device, deviceId, m.config.AuthType)
		if err == nil && customPerms != nil {
			permList = customPerms
		}
	} else if m.CustomPermissionListFunc != nil {
		customPerms, err := m.CustomPermissionListFunc(sess.LoginID, m.config.AuthType)
		if err == nil && customPerms != nil {
			permList = customPerms
		}
	}

	// Pass on any matching permission 任一权限匹配即通过
	hasAny := false
	for _, need := range permissions {
		if m.hasPermissionInList(permList, need) {
			hasAny = true
			break
		}
	}

	// Trigger permission check event 触发权限检查事件
	m.triggerEvent(listener.EventPermissionCheck, sess.LoginID, device, deviceId, tokenValue, map[string]any{
		listener.ExtraKeyPermissions: permissions,
		listener.ExtraKeyLogic:       listener.LogicOr,
		listener.ExtraKeyResult:      hasAny,
	})

	return hasAny
}

// AddRoles adds roles to a user. AddRoles 为用户添加角色。
func (m *Manager) AddRoles(ctx context.Context, loginID string, roles []string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}

	// Get session 获取 Session
	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		return err
	}

	sess.addRoles(roles...)
	err = m.saveToStorage(ctx, m.getSessionKey(loginID), *sess)
	if err != nil {
		return err
	}

	return nil
}

// AddRolesByToken adds roles to a user by token. AddRolesByToken 根据 Token 为用户添加角色。
func (m *Manager) AddRolesByToken(ctx context.Context, tokenValue string, roles []string) error {
	// Get session 获取 Session
	sess, err := m.GetSessionByToken(ctx, tokenValue)
	if err != nil {
		return err
	}

	// Add roles 添加角色
	sess.addRoles(roles...)
	// Save session 保存 Session
	err = m.saveToStorage(ctx, m.getSessionKey(sess.LoginID), *sess)
	if err != nil {
		return err
	}

	return nil
}

// RemoveRoles removes roles from a user. RemoveRoles 删除用户的指定角色。
func (m *Manager) RemoveRoles(ctx context.Context, loginID string, roles []string) error {
	if loginID == "" {
		return serror.ErrIDIsEmpty
	}

	// Get session 获取 Session
	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		return err
	}

	sess.removeRoles(roles...)
	err = m.saveToStorage(ctx, m.getSessionKey(loginID), *sess)
	if err != nil {
		return err
	}

	return nil
}

// RemoveRolesByToken removes roles from a user by token. RemoveRolesByToken 根据 Token 删除用户的指定角色。
func (m *Manager) RemoveRolesByToken(ctx context.Context, tokenValue string, roles []string) error {
	// Get session 获取 Session
	sess, err := m.GetSessionByToken(ctx, tokenValue)
	if err != nil {
		return err
	}

	// Remove roles 删除角色
	sess.removeRoles(roles...)
	// Save session 保存 Session
	err = m.saveToStorage(ctx, m.getSessionKey(sess.LoginID), *sess)
	if err != nil {
		return err
	}

	return nil
}

// GetRoles retrieves the role list for a user. GetRoles 获取用户的角色列表。
func (m *Manager) GetRoles(ctx context.Context, loginID string) ([]string, error) {
	if loginID == "" {
		return nil, serror.ErrIDIsEmpty
	}

	// Use custom role list function 使用自定义角色列表获取函数
	if m.CustomRoleListFunc != nil {
		return m.CustomRoleListFunc(loginID, m.config.AuthType)
	}

	// Get session 获取 Session
	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		return nil, err
	}

	return sess.Roles, nil
}

// GetRolesByToken retrieves the role list by token. GetRolesByToken 根据 Token 获取角色列表。
func (m *Manager) GetRolesByToken(ctx context.Context, tokenValue string) ([]string, error) {
	// Get session 获取 Session
	sess, err := m.GetSessionByToken(ctx, tokenValue)
	if err != nil {
		return nil, err
	}

	return sess.Roles, nil
}

// HasRole checks if a user has a specific role. HasRole 检查用户是否拥有指定角色。
func (m *Manager) HasRole(ctx context.Context, loginID string, role string) bool {
	if loginID == "" {
		return false
	}

	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		m.logger.Errorf("HasRole: failed to get session for loginID=%s, error=%v", loginID, err)
		return false
	}

	// Get roles with Func then Session priority 获取角色列表（两级优先级：Func > Session）
	roles := sess.Roles
	if m.CustomRoleListFunc != nil {
		customRoles, err := m.CustomRoleListFunc(loginID, m.config.AuthType)
		if err == nil && customRoles != nil {
			roles = customRoles
		}
	}

	hasRole := false
	for _, r := range roles {
		if r == role {
			hasRole = true
			break
		}
	}

	// Trigger role check event 触发角色检查事件
	m.triggerEvent(listener.EventRoleCheck, loginID, "", "", "", map[string]any{
		listener.ExtraKeyRole:   role,
		listener.ExtraKeyResult: hasRole,
	})

	return hasRole
}

// HasRoleByToken checks if a user has a specific role by token. HasRoleByToken 根据 Token 检查用户是否拥有指定角色。
func (m *Manager) HasRoleByToken(ctx context.Context, tokenValue string, role string) bool {
	sess, err := m.GetSessionByToken(ctx, tokenValue)
	if err != nil {
		return false
	}

	// Get device and deviceId 获取 device/deviceId
	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err != nil {
		m.logger.Errorf("HasRoleByToken: failed to get token info for token=%s, error=%v", tokenValue, err)
		return false
	}
	device, deviceId := tokenInfo.Device, tokenInfo.DeviceId

	// Get roles with Ext Func Session priority 获取角色列表（三级优先级：Ext > Func > Session）
	roles := sess.Roles
	if m.CustomRoleListExtFunc != nil {
		customRoles, err := m.CustomRoleListExtFunc(sess.LoginID, device, deviceId, m.config.AuthType)
		if err == nil && customRoles != nil {
			roles = customRoles
		}
	} else if m.CustomRoleListFunc != nil {
		customRoles, err := m.CustomRoleListFunc(sess.LoginID, m.config.AuthType)
		if err == nil && customRoles != nil {
			roles = customRoles
		}
	}

	hasRole := false
	for _, r := range roles {
		if r == role {
			hasRole = true
			break
		}
	}

	// Trigger role check event 触发角色检查事件
	m.triggerEvent(listener.EventRoleCheck, sess.LoginID, device, deviceId, tokenValue, map[string]any{
		listener.ExtraKeyRole:   role,
		listener.ExtraKeyResult: hasRole,
	})

	return hasRole
}

// HasRolesAnd checks if a user has all specified roles (AND logic). HasRolesAnd 检查用户是否拥有所有指定角色（AND 逻辑）。
func (m *Manager) HasRolesAnd(ctx context.Context, loginID string, roles []string) bool {
	if loginID == "" {
		return false
	}

	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		m.logger.Errorf("HasRolesAnd: failed to get session for loginID=%s, error=%v", loginID, err)
		return false
	}

	// Get roles with Func then Session priority 获取角色列表（两级优先级：Func > Session）
	roleList := sess.Roles
	if m.CustomRoleListFunc != nil {
		customRoles, err := m.CustomRoleListFunc(loginID, m.config.AuthType)
		if err == nil && customRoles != nil {
			roleList = customRoles
		}
	}

	// Check each required role 校验每一个必需角色
	hasAll := true
	for _, need := range roles {
		found := false
		for _, r := range roleList {
			if r == need {
				found = true
				break
			}
		}
		if !found {
			hasAll = false
			break
		}
	}

	// Trigger role check event 触发角色检查事件
	m.triggerEvent(listener.EventRoleCheck, loginID, "", "", "", map[string]any{
		listener.ExtraKeyRoles:  roles,
		listener.ExtraKeyLogic:  listener.LogicAnd,
		listener.ExtraKeyResult: hasAll,
	})

	return hasAll
}

// HasRolesAndByToken checks if a user has all specified roles by token (AND logic). HasRolesAndByToken 根据 Token 检查用户是否拥有所有指定角色（AND 逻辑）。
func (m *Manager) HasRolesAndByToken(ctx context.Context, tokenValue string, roles []string) bool {
	sess, err := m.GetSessionByToken(ctx, tokenValue)
	if err != nil {
		return false
	}

	// Get device and deviceId 获取 device/deviceId
	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err != nil {
		m.logger.Errorf("HasRolesAndByToken: failed to get token info for token=%s, error=%v", tokenValue, err)
		return false
	}
	device, deviceId := tokenInfo.Device, tokenInfo.DeviceId

	// Get roles with Ext Func Session priority 获取角色列表（三级优先级：Ext > Func > Session）
	roleList := sess.Roles
	if m.CustomRoleListExtFunc != nil {
		customRoles, err := m.CustomRoleListExtFunc(sess.LoginID, device, deviceId, m.config.AuthType)
		if err == nil && customRoles != nil {
			roleList = customRoles
		}
	} else if m.CustomRoleListFunc != nil {
		customRoles, err := m.CustomRoleListFunc(sess.LoginID, m.config.AuthType)
		if err == nil && customRoles != nil {
			roleList = customRoles
		}
	}

	// Check each required role 校验每一个必需角色
	hasAll := true
	for _, need := range roles {
		found := false
		for _, r := range roleList {
			if r == need {
				found = true
				break
			}
		}
		if !found {
			hasAll = false
			break
		}
	}

	// Trigger role check event 触发角色检查事件
	m.triggerEvent(listener.EventRoleCheck, sess.LoginID, device, deviceId, tokenValue, map[string]any{
		listener.ExtraKeyRoles:  roles,
		listener.ExtraKeyLogic:  listener.LogicAnd,
		listener.ExtraKeyResult: hasAll,
	})

	return hasAll
}

// HasRolesOr checks if a user has any of the specified roles (OR logic). HasRolesOr 检查用户是否拥有任一指定角色（OR 逻辑）。
func (m *Manager) HasRolesOr(ctx context.Context, loginID string, roles []string) bool {
	if loginID == "" {
		return false
	}

	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		m.logger.Errorf("HasRolesOr: failed to get session for loginID=%s, error=%v", loginID, err)
		return false
	}

	// Get roles with Func then Session priority 获取角色列表（两级优先级：Func > Session）
	roleList := sess.Roles
	if m.CustomRoleListFunc != nil {
		customRoles, err := m.CustomRoleListFunc(loginID, m.config.AuthType)
		if err == nil && customRoles != nil {
			roleList = customRoles
		}
	}

	// Pass on any matching role 任一角色匹配即通过
	hasAny := false
	for _, need := range roles {
		for _, r := range roleList {
			if r == need {
				hasAny = true
				break
			}
		}
		if hasAny {
			break
		}
	}

	// Trigger role check event 触发角色检查事件
	m.triggerEvent(listener.EventRoleCheck, loginID, "", "", "", map[string]any{
		listener.ExtraKeyRoles:  roles,
		listener.ExtraKeyLogic:  listener.LogicOr,
		listener.ExtraKeyResult: hasAny,
	})

	return hasAny
}

// HasRolesOrByToken checks if a user has any of the specified roles by token (OR logic). HasRolesOrByToken 根据 Token 检查用户是否拥有任一指定角色（OR 逻辑）。
func (m *Manager) HasRolesOrByToken(ctx context.Context, tokenValue string, roles []string) bool {
	sess, err := m.GetSessionByToken(ctx, tokenValue)
	if err != nil {
		return false
	}

	// Get device and deviceId 获取 device/deviceId
	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err != nil {
		m.logger.Errorf("HasRolesOrByToken: failed to get token info for token=%s, error=%v", tokenValue, err)
		return false
	}
	device, deviceId := tokenInfo.Device, tokenInfo.DeviceId

	// Get roles with Ext Func Session priority 获取角色列表（三级优先级：Ext > Func > Session）
	roleList := sess.Roles
	if m.CustomRoleListExtFunc != nil {
		customRoles, err := m.CustomRoleListExtFunc(sess.LoginID, device, deviceId, m.config.AuthType)
		if err == nil && customRoles != nil {
			roleList = customRoles
		}
	} else if m.CustomRoleListFunc != nil {
		customRoles, err := m.CustomRoleListFunc(sess.LoginID, m.config.AuthType)
		if err == nil && customRoles != nil {
			roleList = customRoles
		}
	}

	// Pass on any matching role 任一角色匹配即通过
	hasAny := false
	for _, need := range roles {
		for _, r := range roleList {
			if r == need {
				hasAny = true
				break
			}
		}
		if hasAny {
			break
		}
	}

	// Trigger role check event 触发角色检查事件
	m.triggerEvent(listener.EventRoleCheck, sess.LoginID, device, deviceId, tokenValue, map[string]any{
		listener.ExtraKeyRoles:  roles,
		listener.ExtraKeyLogic:  listener.LogicOr,
		listener.ExtraKeyResult: hasAny,
	})

	return hasAny
}

// CheckPermission checks if a user has a specific permission, returns error if not. CheckPermission 校验用户是否拥有指定权限，无权限返回 error。
func (m *Manager) CheckPermission(ctx context.Context, loginID string, permission string) error {
	if !m.HasPermission(ctx, loginID, permission) {
		return fmt.Errorf("%w: %s", serror.ErrPermissionDenied, permission)
	}
	return nil
}

// CheckPermissionAnd checks if a user has all specified permissions, returns error if not. CheckPermissionAnd 校验用户是否拥有所有指定权限，缺少任一权限返回 error。
func (m *Manager) CheckPermissionAnd(ctx context.Context, loginID string, permissions []string) error {
	if !m.HasPermissionsAnd(ctx, loginID, permissions) {
		return serror.ErrPermissionDenied
	}
	return nil
}

// CheckPermissionOr checks if a user has any of the specified permissions, returns error if none. CheckPermissionOr 校验用户是否拥有任一指定权限，全部缺少返回 error。
func (m *Manager) CheckPermissionOr(ctx context.Context, loginID string, permissions []string) error {
	if !m.HasPermissionsOr(ctx, loginID, permissions) {
		return serror.ErrPermissionDenied
	}
	return nil
}

// CheckRole checks if a user has a specific role, returns error if not. CheckRole 校验用户是否拥有指定角色，无角色返回 error。
func (m *Manager) CheckRole(ctx context.Context, loginID string, role string) error {
	if !m.HasRole(ctx, loginID, role) {
		return fmt.Errorf("%w: %s", serror.ErrRoleDenied, role)
	}
	return nil
}

// CheckRoleAnd checks if a user has all specified roles, returns error if not. CheckRoleAnd 校验用户是否拥有所有指定角色，缺少任一角色返回 error。
func (m *Manager) CheckRoleAnd(ctx context.Context, loginID string, roles []string) error {
	if !m.HasRolesAnd(ctx, loginID, roles) {
		return serror.ErrRoleDenied
	}
	return nil
}

// CheckRoleOr checks if a user has any of the specified roles, returns error if none. CheckRoleOr 校验用户是否拥有任一指定角色，全部缺少返回 error。
func (m *Manager) CheckRoleOr(ctx context.Context, loginID string, roles []string) error {
	if !m.HasRolesOr(ctx, loginID, roles) {
		return serror.ErrRoleDenied
	}
	return nil
}

// CheckDisable checks if an account is disabled, returns error if disabled. CheckDisable 校验账号是否被封禁，被封禁返回 error。
func (m *Manager) CheckDisable(ctx context.Context, loginID string) error {
	if m.IsDisable(ctx, loginID) {
		return serror.ErrAccountDisabled
	}
	return nil
}

// GetConfig retrieves the manager configuration. GetConfig 获取管理器配置。
func (m *Manager) GetConfig() *config.Config {
	return m.config
}

// GetGenerator retrieves the token generator. GetGenerator 获取 Token 生成器。
func (m *Manager) GetGenerator() adapter.Generator {
	return m.generator
}

// GetStorage retrieves the storage adapter. GetStorage 获取存储适配器。
func (m *Manager) GetStorage() adapter.Storage {
	return m.storage
}

// GetSerializer retrieves the serializer adapter. GetSerializer 获取序列化器适配器。
func (m *Manager) GetSerializer() adapter.Codec {
	return m.serializer
}

// GetLogger retrieves the logger adapter. GetLogger 获取日志适配器。
func (m *Manager) GetLogger() adapter.Log {
	return m.logger
}

// GetPool retrieves the goroutine pool. GetPool 获取协程池。
func (m *Manager) GetPool() adapter.Pool {
	return m.pool
}

// GetCustomPermissionListFunc retrieves the custom permission list function. GetCustomPermissionListFunc 获取自定义权限列表获取函数。
func (m *Manager) GetCustomPermissionListFunc() func(loginID, authType string) ([]string, error) {
	return m.CustomPermissionListFunc
}

// GetCustomRoleListFunc retrieves the custom role list function. GetCustomRoleListFunc 获取自定义角色列表获取函数。
func (m *Manager) GetCustomRoleListFunc() func(loginID, authType string) ([]string, error) {
	return m.CustomRoleListFunc
}

// GetNonceManager retrieves the nonce manager. GetNonceManager 获取 Nonce 管理器。
func (m *Manager) GetNonceManager() *nonce.NonceManager {
	return m.nonceManager
}

// GetOAuth2Manager retrieves the OAuth2 manager. GetOAuth2Manager 获取 OAuth2 管理器。
func (m *Manager) GetOAuth2Manager() *oauth2.OAuth2Server {
	return m.oauth2Manager
}

// getSession retrieves session information (internal method). getSession 获取会话信息（内部方法）。
func (m *Manager) getSession(ctx context.Context, loginID string, autoCreate ...bool) (*Session, error) {
	sessData, err := m.storage.Get(ctx, m.getSessionKey(loginID))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
	}
	if sessData == nil {
		if len(autoCreate) > 0 && autoCreate[0] {
			newSession := &Session{
				AuthType:      m.config.AuthType,
				LoginID:       loginID,
				CreateTime:    time.Now().Unix(),
				TerminalInfos: make([]TerminalInfo, 0),
				Permissions:   make([]string, 0),
				Roles:         make([]string, 0),
			}
			// Trigger session create event 触发创建 Session 事件
			m.triggerEvent(listener.EventCreateSession, loginID, "", "", "", nil)
			return newSession, nil
		}

		return nil, serror.ErrSessionNotFound
	}

	bytesData, err := utils.ToBytes(sessData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", serror.ErrTypeConvert, err)
	}

	var sess Session
	err = m.serializer.Decode(bytesData, &sess)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", serror.ErrSerializeFailed, err)
	}

	return &sess, nil
}

// getTokenInfo retrieves token information (internal method). getTokenInfo 获取 Token 信息（内部方法）。
func (m *Manager) getTokenInfo(ctx context.Context, tokenValue string) (*TokenInfo, error) {
	tokenInfoData, err := m.storage.Get(ctx, m.getTokenKey(tokenValue))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
	}
	if tokenInfoData == nil {
		return nil, serror.ErrInvalidToken
	}

	tokenInfoBytes, err := utils.ToBytes(tokenInfoData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", serror.ErrTypeConvert, err)
	}

	switch string(tokenInfoBytes) {
	case string(TokenStateLogout):
		return nil, serror.ErrInvalidToken
	case string(TokenStateKickOut):
		return nil, serror.ErrTokenKickout
	case string(TokenStateReplaced):
		return nil, serror.ErrTokenReplaced
	}

	var tokenInfo TokenInfo
	err = m.serializer.Decode(tokenInfoBytes, &tokenInfo)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", serror.ErrTypeConvert, err)
	}

	return &tokenInfo, nil
}

// loginGetSession retrieves session for login operation (internal method). loginGetSession 获取登录操作的会话信息（内部方法）。
func (m *Manager) loginGetSession(ctx context.Context, loginID string) (*Session, error) {
	sessData, err := m.storage.Get(ctx, m.getSessionKey(loginID))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
	}
	if sessData == nil {
		return nil, nil
	}

	bytesData, err := utils.ToBytes(sessData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", serror.ErrTypeConvert, err)
	}

	var sess Session
	err = m.serializer.Decode(bytesData, &sess)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", serror.ErrSerializeFailed, err)
	}

	return &sess, nil
}

// checkLoginInternal performs the core login validation logic (internal method). checkLoginInternal 执行登录状态的核心验证逻辑（内部方法）。
func (m *Manager) checkLoginInternal(ctx context.Context, tokenValue string) error {
	// Get tokenInfo 获取 tokenInfo
	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err != nil {
		return err
	}

	// Check disable status after token lookup 获取 token 后检查封禁状态
	if m.isDisable(ctx, tokenInfo.LoginID) {
		return serror.ErrAccountDisabled
	}

	// Check max inactive timeout 检查最大不活跃时长
	if m.config.ActiveTimeout > 0 {
		timeStampAny, err := m.storage.Get(ctx, m.getActiveKey(tokenValue))
		if err == nil && timeStampAny != nil {
			timeStamp, err := utils.ToInt64(timeStampAny)
			if err != nil {
				_ = m.storage.Delete(ctx, m.getActiveKey(tokenValue))
			} else if time.Now().Unix()-timeStamp > m.config.ActiveTimeout {
				// Kick out token when inactive timeout exceeded Token 已超过最大不活跃时长，执行踢出操作
				_ = m.Kickout(ctx, tokenValue)
				return serror.ErrTokenKickout
			}
		}
	}

	// Renew asynchronously 异步续期
	if m.config.AutoRenew && m.config.Timeout > 0 {
		if ttl, err := m.storage.TTL(ctx, m.getTokenKey(tokenValue)); err == nil && ttl > 0 {
			ttlSeconds := int64(ttl.Seconds())
			if ttlSeconds > 0 &&
				(m.config.RenewMaxRefresh <= 0 || ttlSeconds <= m.config.RenewMaxRefresh) &&
				(m.config.RenewInterval <= 0 || !m.storage.Exists(ctx, m.getRenewKey(tokenValue))) {

				renewFunc := func() {
					m.renewFunc(context.Background(), tokenValue, tokenInfo.LoginID)
				}

				if m.pool != nil {
					_ = m.pool.Submit(renewFunc)
				} else {
					go renewFunc()
				}
			}
		}
	}

	// Update active timeout asynchronously 异步活跃时长
	if m.config.ActiveTimeout > 0 {
		activeFunc := func() {
			if err := m.storage.Set(ctx, m.getActiveKey(tokenValue), time.Now().Unix(), m.getExpiration()); err != nil {
				m.logger.Errorf("checkLoginInternal: failed to set active key for token=%s, error=%v", tokenValue, err)
			}
		}
		if m.pool != nil {
			_ = m.pool.Submit(activeFunc)
		} else {
			go activeFunc()
		}
	}

	return nil
}

// cleanExpiredTerminals removes expired tokens from session (internal method). cleanExpiredTerminals 清理会话中已过期的 token（内部方法）。
func (m *Manager) cleanExpiredTerminals(ctx context.Context, sess *Session) error {
	if sess == nil || len(sess.TerminalInfos) == 0 {
		return nil
	}

	var validTerminals []TerminalInfo
	hasExpired := false

	for _, ti := range sess.TerminalInfos {
		// Treat successful token lookup as valid 尝试获取 token 信息，如果成功则说明 token 有效
		_, err := m.getTokenInfo(ctx, ti.Token)
		if err == nil {
			validTerminals = append(validTerminals, ti)
			continue
		}

		// Remove only terminal state errors 仅清理明确的终端状态错误
		if errors.Is(err, serror.ErrInvalidToken) ||
			errors.Is(err, serror.ErrTokenExpired) ||
			errors.Is(err, serror.ErrTokenKickout) ||
			errors.Is(err, serror.ErrTokenReplaced) {
			hasExpired = true
			continue
		}

		return err
	}

	// Update session when expired tokens exist 如果有过期的 token，更新 session
	if hasExpired {
		sess.TerminalInfos = validTerminals
		if err := m.saveToStorage(ctx, m.getSessionKey(sess.LoginID), *sess); err != nil {
			return err
		}
	}

	return nil
}

// handleConcurrency handles login concurrency strategy (internal method). handleConcurrency 处理登录并发策略（内部方法）。
func (m *Manager) handleConcurrency(
	ctx context.Context,
	sess *Session,
	loginID, device string,
) (reuseToken string, handled bool, err error) {
	// Clean expired tokens 清理已过期的 token
	if err = m.cleanExpiredTerminals(ctx, sess); err != nil {
		return "", false, err
	}

	if !m.config.IsConcurrent {
		// Kick old sessions when concurrency is disabled 不允许并发：踢掉旧会话
		if m.config.ConcurrencyScope == config.ConcurrencyScopeAccount {
			if err = m.removeTerminalInfosAndTokens(ctx, sess); err != nil {
				return "", false, err
			}
		} else if m.config.ConcurrencyScope == config.ConcurrencyScopeDevice {
			if err = m.removeTerminalInfosAndTokens(ctx, sess, device); err != nil {
				return "", false, err
			}
		}
		return "", true, nil
	}

	if m.config.IsShare {
		// Try token sharing reuse 允许共享：尝试复用
		var token string
		var shareErr error
		if m.config.ConcurrencyScope == config.ConcurrencyScopeAccount {
			token, shareErr = m.getTokenAndShare(ctx, sess)
		} else if m.config.ConcurrencyScope == config.ConcurrencyScopeDevice {
			token, shareErr = m.getTokenAndShare(ctx, sess, device)
		}
		if shareErr != nil {
			return "", false, shareErr
		}
		if token != "" {
			return token, true, nil
		}
	}

	if m.config.ConcurrencyScope == config.ConcurrencyScopeAccount {
		if m.config.MaxLoginCount > 0 && int64(len(sess.TerminalInfos)) >= m.config.MaxLoginCount {
			if err := m.removeOldestTerminalInfoAndToken(ctx, sess); err != nil {
				return "", false, err
			}
			return "", true, nil
		}
	} else if m.config.ConcurrencyScope == config.ConcurrencyScopeDevice {
		if m.config.MaxLoginCount > 0 && int64(len(sess.getTerminalsByDevice(device))) >= m.config.MaxLoginCount {
			if err := m.removeOldestTerminalInfoAndToken(ctx, sess, device); err != nil {
				return "", false, err
			}
			return "", true, nil
		}
	}

	return "", false, nil
}

// getTokenAndShare retrieves and shares a token 获取并共享 token
func (m *Manager) getTokenAndShare(ctx context.Context, sess *Session, device ...string) (string, error) {
	if len(sess.TerminalInfos) == 0 {
		return "", nil
	}

	// Get candidate terminals 获取候选的 terminals
	var candidates []TerminalInfo
	if len(device) > 0 {
		// Get terminals for specified device 指定设备：获取该设备的所有 terminals
		candidates = sess.getTerminalsByDevice(device[0])
	} else {
		// Get all terminals for account scope 账号级别：获取所有 terminals
		candidates = sess.TerminalInfos
	}

	if len(candidates) == 0 {
		return "", nil
	}

	// Revive latest token by resetting TTL 获取最后一个（最新的）token 并直接复活，若 token 仍在 session 中说明它未被踢出或顶替，即使过期也可通过重置 TTL 复活
	terminalInfo := candidates[len(candidates)-1]

	// Renew session 续期 session
	if err := m.storage.Expire(ctx, m.getSessionKey(terminalInfo.LoginID), m.getExpiration()); err != nil {
		m.logger.Errorf("getTokenAndShare: failed to expire session for loginID=%s, error=%v", terminalInfo.LoginID, err)
	}

	// Renew token and revive if expired 续期 Token（如果已过期，这会复活它）
	tokenInfo := TokenInfo{
		AuthType:   m.config.AuthType,
		LoginID:    terminalInfo.LoginID,
		Device:     terminalInfo.Device,
		DeviceId:   terminalInfo.DeviceId,
		CreateTime: terminalInfo.CreateTime,
	}
	if err := m.saveToStorage(ctx, m.getTokenKey(terminalInfo.Token), tokenInfo, m.getExpiration()); err != nil {
		return "", err
	}

	// Renew or reset metadata 续期或重新设置 metadata
	if m.config.RenewInterval > 0 {
		if err := m.storage.Set(ctx, m.getRenewKey(terminalInfo.Token), time.Now().Unix(), time.Duration(m.config.RenewInterval)*time.Second); err != nil {
			m.logger.Errorf("getTokenAndShare: failed to set renew key for token=%s, error=%v", terminalInfo.Token, err)
		}
	}
	// Set active timeout 设置最大不活跃时长
	if m.config.ActiveTimeout > 0 {
		if err := m.storage.Set(ctx, m.getActiveKey(terminalInfo.Token), time.Now().Unix(), m.getExpiration()); err != nil {
			m.logger.Errorf("getTokenAndShare: failed to set active key for token=%s, error=%v", terminalInfo.Token, err)
		}
	}

	return terminalInfo.Token, nil
}

// removeOldestTerminalInfoAndToken removes the oldest terminal and its token (internal method). removeOldestTerminalInfoAndToken 移除最旧的终端信息并删除对应的 Token（内部方法）。
func (m *Manager) removeOldestTerminalInfoAndToken(ctx context.Context, sess *Session, device ...string) error {
	if len(device) > 0 {
		terminalInfo, ok := sess.removeOldestTerminal(device...)
		if ok {
			// Save session data 保存会话数据
			if err := m.saveToStorage(ctx, m.getSessionKey(sess.LoginID), *sess); err != nil {
				return err
			}
			// Mark token as kicked out 设置 token 状态为踢出
			if err := m.storage.Set(ctx, m.getTokenKey(terminalInfo.Token), string(TokenStateKickOut), m.getExpiration()); err != nil {
				return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
			}
			// Clean metadata 清理 metadata
			if err := m.storage.Delete(ctx, m.getRenewKey(terminalInfo.Token)); err != nil {
				return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
			}
			if err := m.storage.Delete(ctx, m.getActiveKey(terminalInfo.Token)); err != nil {
				return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
			}
		}
		return nil
	}

	terminalInfo, ok := sess.removeOldestTerminal()
	if ok {
		// Save session data 保存会话数据
		if err := m.saveToStorage(ctx, m.getSessionKey(sess.LoginID), *sess); err != nil {
			return err
		}
		// Mark token as kicked out 设置 token 状态为踢出
		if err := m.storage.Set(ctx, m.getTokenKey(terminalInfo.Token), string(TokenStateKickOut), m.getExpiration()); err != nil {
			return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
		}
		// Clean metadata 清理 metadata
		if err := m.storage.Delete(ctx, m.getRenewKey(terminalInfo.Token)); err != nil {
			return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
		}
		if err := m.storage.Delete(ctx, m.getActiveKey(terminalInfo.Token)); err != nil {
			return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
		}
	}
	return nil
}

// removeTerminalInfosAndTokens removes terminal information and tokens (internal method). removeTerminalInfosAndTokens 移除终端信息和 Token（内部方法）。
func (m *Manager) removeTerminalInfosAndTokens(ctx context.Context, sess *Session, device ...string) error {
	if len(device) > 0 {
		// Remove terminals for specified device 移除指定设备类型的终端信息
		terminalInfos := sess.removeTerminalByDevice(device[0])

		// Delete session when no terminals remain 如果 session 中没有剩余终端，删除整个 session
		if len(sess.TerminalInfos) == 0 {
			if err := m.storage.Delete(ctx, m.getSessionKey(sess.LoginID)); err != nil {
				return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
			}
			// Trigger session destroy event 触发销毁 Session 事件
			m.triggerEvent(listener.EventDestroySession, sess.LoginID, "", "", "", nil)
		} else {
			// Save updated session otherwise 否则保存更新后的 session
			if err := m.saveToStorage(ctx, m.getSessionKey(sess.LoginID), *sess); err != nil {
				return err
			}
		}

		// Mark all tokens as kicked out 将所有的 Token 设置为踢出
		for _, info := range terminalInfos {
			err := m.storage.Set(ctx, m.getTokenKey(info.Token), string(TokenStateKickOut), m.getExpiration())
			if err != nil {
				return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
			}
		}

		// 清理附属 metadata
		tokens := make([]string, len(terminalInfos))
		for i, info := range terminalInfos {
			tokens[i] = info.Token
		}
		if err := m.cleanTokenMetadata(ctx, tokens); err != nil {
			return err
		}

		return nil
	}

	// Get old terminal infos 获取旧的终端信息
	oldTerminalInfos := sess.TerminalInfos

	// Remove terminal infos 移除终端信息
	sess.TerminalInfos = make([]TerminalInfo, 0)

	// Delete session when empty session 为空，删除整个 session
	if err := m.storage.Delete(ctx, m.getSessionKey(sess.LoginID)); err != nil {
		return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
	}

	// Trigger session destroy event 触发销毁 Session 事件
	m.triggerEvent(listener.EventDestroySession, sess.LoginID, "", "", "", nil)

	// Mark all tokens as kicked out 将所有的 Token 设置为踢出
	for _, terminalInfo := range oldTerminalInfos {
		err := m.storage.Set(ctx, m.getTokenKey(terminalInfo.Token), string(TokenStateKickOut), m.getExpiration())
		if err != nil {
			return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
		}
	}

	// 清理附属 metadata
	tokens := make([]string, len(oldTerminalInfos))
	for i, info := range oldTerminalInfos {
		tokens[i] = info.Token
	}
	if err := m.cleanTokenMetadata(ctx, tokens); err != nil {
		return err
	}

	return nil
}

// logoutTerminals performs common logout logic (internal method). logoutTerminals 通用登出逻辑：移除终端 + 删除 token + 清理 metadata（内部方法）。
func (m *Manager) logoutTerminals(
	ctx context.Context,
	loginID string,
	removalFunc func(*Session) []TerminalInfo,
) error {
	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		return err
	}
	if sess == nil {
		return nil // session 不存在，登出无害
	}

	removed := removalFunc(sess)
	if len(removed) == 0 {
		return nil
	}

	// Delete session when no terminals remain 如果 session 中没有剩余终端，删除整个 session
	if len(sess.TerminalInfos) == 0 {
		if err = m.storage.Delete(ctx, m.getSessionKey(loginID)); err != nil {
			return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
		}
		// Trigger session destroy event 触发销毁 Session 事件
		m.triggerEvent(listener.EventDestroySession, loginID, "", "", "", nil)
	} else {
		// Save updated session otherwise 否则保存更新后的 session
		if err = m.saveToStorage(ctx, m.getSessionKey(loginID), *sess); err != nil {
			return err
		}
	}

	// Extract token list 提取 token 列表
	tokens := make([]string, len(removed))
	tokenKeys := make([]string, len(removed))
	for i, info := range removed {
		tokens[i] = info.Token
		tokenKeys[i] = m.getTokenKey(info.Token)
	}

	// Delete primary token keys 删除主 token keys
	if err = m.storage.Delete(ctx, tokenKeys...); err != nil {
		return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
	}

	// 清理附属 metadata
	if err = m.cleanTokenMetadata(ctx, tokens); err != nil {
		return err
	}

	// Trigger logout event 触发登出事件
	for _, info := range removed {
		m.triggerEvent(listener.EventLogout, loginID, info.Device, info.DeviceId, info.Token, nil)
	}

	return nil
}

// cleanTokenMetadata cleans token metadata in batch (internal method). cleanTokenMetadata 批量清理 token 的附属元数据（续期 key、活跃时间 key）（内部方法）。
func (m *Manager) cleanTokenMetadata(ctx context.Context, tokens []string) error {
	if len(tokens) == 0 {
		return nil
	}

	keys := make([]string, 0, len(tokens)*2)
	for _, token := range tokens {
		keys = append(keys, m.getRenewKey(token), m.getActiveKey(token))
	}

	if err := m.storage.Delete(ctx, keys...); err != nil {
		return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
	}

	return nil
}

// TerminalRemovalFunc defines how to remove terminals from a session. TerminalRemovalFunc 定义如何从 Session 中移除终端。
type TerminalRemovalFunc func(sess *Session) []TerminalInfo

// processTerminals performs common terminal processing logic (internal method). processTerminals 通用终端处理逻辑（内部方法）。
func (m *Manager) processTerminals(
	ctx context.Context,
	loginID string,
	removalFunc TerminalRemovalFunc,
	state TokenState,
) error {
	// Load session 加载 Session
	sess, err := m.getSession(ctx, loginID)
	if err != nil {
		return err
	}

	// Apply removal strategy 执行移除策略
	removedTerminals := removalFunc(sess)

	// Update session when terminals are removed 如果有移除项，更新 session
	if len(removedTerminals) > 0 {
		// Delete session when no terminals remain 如果 session 中没有剩余终端，删除整个 session
		if len(sess.TerminalInfos) == 0 {
			if err = m.storage.Delete(ctx, m.getSessionKey(loginID)); err != nil {
				return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
			}
			// Trigger session destroy event 触发销毁 Session 事件
			m.triggerEvent(listener.EventDestroySession, loginID, "", "", "", nil)
		} else {
			// Save updated session otherwise 否则保存更新后的 session
			if err = m.saveToStorage(ctx, m.getSessionKey(loginID), *sess); err != nil {
				return err
			}
		}
	}

	// Clean each removed token 对每个被移除的 token 执行清理
	for _, info := range removedTerminals {
		token := info.Token

		// Set token state 设置 token 状态
		if err = m.storage.Set(ctx, m.getTokenKey(token), string(state), m.getExpiration()); err != nil {
			return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
		}

		// Delete renew key 删除续期 key
		if err = m.storage.Delete(ctx, m.getRenewKey(token)); err != nil {
			return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
		}

		// Delete active key 删除活跃时间 key
		if err = m.storage.Delete(ctx, m.getActiveKey(token)); err != nil {
			return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
		}
	}

	// Trigger matched event 触发对应事件
	var event listener.Event
	switch state {
	case TokenStateKickOut:
		event = listener.EventKickout
	case TokenStateReplaced:
		event = listener.EventReplace
	}

	if event != "" {
		for _, info := range removedTerminals {
			m.triggerEvent(event, loginID, info.Device, info.DeviceId, info.Token, nil)
		}
	}

	return nil
}

// isDisable checks if an account is disabled (internal method). isDisable 检查账号是否被封禁（内部方法）。
func (m *Manager) isDisable(ctx context.Context, loginID string) bool {
	return m.storage.Exists(ctx, m.getDisableKey(loginID))
}

// filterTokens filters tokens based on checkAlive flag (internal method). filterTokens 根据 checkAlive 决定是否验证 token 有效性，并返回 token 列表（内部方法）。
func (m *Manager) filterTokens(ctx context.Context, terminals []TerminalInfo, checkAlive bool) ([]string, error) {
	if len(terminals) == 0 {
		return []string{}, nil
	}

	if !checkAlive {
		// Return all tokens without alive check 不检查存活：直接返回所有 token（预分配容量）
		tokens := make([]string, len(terminals))
		for i, ti := range terminals {
			tokens[i] = ti.Token
		}
		return tokens, nil
	}

	// Check each token by GetTokenInfo 检查每个 token 是否有效（调用 GetTokenInfo）
	var tokens []string // 无法预知数量，动态 append
	for _, ti := range terminals {
		if _, err := m.GetTokenInfo(ctx, ti.Token); err == nil {
			tokens = append(tokens, ti.Token)
		}
		// Skip invalid tokens 若 token 无效（过期/被踢等），跳过
	}
	return tokens, nil
}

// matchPermission matches permission with wildcard support (internal method). matchPermission 权限匹配（支持通配符）（内部方法）。
func (m *Manager) matchPermission(pattern, permission string) bool {
	// Wildcard matches all permissions 全通配符匹配所有权限
	if pattern == PermissionWildcard {
		return true
	}

	// Exact match 精确匹配
	if pattern == permission {
		return true
	}

	// Return false when pattern has no wildcard 如果 pattern 不包含通配符，则不匹配
	if !strings.Contains(pattern, PermissionWildcard) {
		return false
	}

	// Auto detect separator from pattern 自动检测分隔符：优先使用 pattern 中的分隔符
	separator := PermissionSeparator // 默认使用 ":"
	if strings.Contains(pattern, "/") {
		separator = "/" // 如果包含 "/"，则使用 URL 路径格式
	}

	// Match wildcard by segments 通配符匹配：按段分割并逐段比较
	patternParts := strings.Split(pattern, separator)
	permParts := strings.Split(permission, separator)

	// Require equal segment count 段数必须一致（避免意外越权）
	if len(patternParts) != len(permParts) {
		return false
	}

	// Match each segment 逐段匹配
	for i := range patternParts {
		// Match wildcard segment 如果 pattern 的当前段是通配符，则该段匹配
		if patternParts[i] == PermissionWildcard {
			continue
		}
		// Require exact segment match 如果 pattern 的当前段不是通配符，则必须精确匹配
		if patternParts[i] != permParts[i] {
			return false
		}
	}

	return true
}

// hasPermissionInList checks if permission exists in permission list (internal method). hasPermissionInList 判断权限是否存在于权限列表中（内部方法）。
func (m *Manager) hasPermissionInList(perms []string, permission string) bool {
	for _, p := range perms {
		if m.matchPermission(p, permission) {
			return true
		}
	}
	return false
}

// renewFunc performs token renewal (internal method). renewFunc 续期函数（内部方法）。
func (m *Manager) renewFunc(ctx context.Context, tokenValue, loginID string) {
	// Validate empty parameters 参数为空校验
	if tokenValue == "" || loginID == "" {
		return
	}

	// Renew token 续期 Token
	if err := m.storage.Expire(ctx, m.getTokenKey(tokenValue), m.getExpiration()); err != nil {
		m.logger.Errorf("renewFunc: failed to expire token for token=%s, error=%v", tokenValue, err)
	}

	// Renew session 续期 Session
	if err := m.storage.Expire(ctx, m.getSessionKey(loginID), m.getExpiration()); err != nil {
		m.logger.Errorf("renewFunc: failed to expire session for loginID=%s, error=%v", loginID, err)
	}

	// Set renew interval marker 设置最小续期间隔标记
	if m.config.RenewInterval > 0 {
		if err := m.storage.Set(ctx, m.getRenewKey(tokenValue), time.Now().Unix(), time.Duration(m.config.RenewInterval)*time.Second); err != nil {
			m.logger.Errorf("renewFunc: failed to set renew key for token=%s, error=%v", tokenValue, err)
		}
	}

	// Trigger renew event 触发续期事件
	tokenInfo, err := m.getTokenInfo(ctx, tokenValue)
	if err == nil {
		m.triggerEvent(listener.EventRenew, loginID, tokenInfo.Device, tokenInfo.DeviceId, tokenValue, nil)
	} else {
		m.logger.Errorf("renewFunc: failed to get token info for token=%s, error=%v", tokenValue, err)
	}
}

// getTokenKey generates the storage key for a token. getTokenKey 获取 Token 存储键。
func (m *Manager) getTokenKey(tokenValue string) string {
	return m.config.KeyPrefix + m.config.AuthType + tokenValue
}

// getSessionKey generates the storage key for a session. getSessionKey 获取会话存储键。
func (m *Manager) getSessionKey(loginID string) string {
	return m.config.KeyPrefix + m.config.AuthType + SessionKeyPrefix + loginID
}

// getRenewKey generates the storage key for token renewal tracking. getRenewKey 获取 Token 续期追踪键。
func (m *Manager) getRenewKey(tokenValue string) string {
	return m.config.KeyPrefix + m.config.AuthType + RenewKeyPrefix + tokenValue
}

// getActiveKey generates the storage key for token activity tracking. getActiveKey 获取 Token 活跃时间追踪键。
func (m *Manager) getActiveKey(tokenValue string) string {
	return m.config.KeyPrefix + m.config.AuthType + ActivePrefix + tokenValue
}

// getDisableKey generates the storage key for account disable status. getDisableKey 获取账号禁用状态存储键。
func (m *Manager) getDisableKey(loginID string) string {
	return m.config.KeyPrefix + m.config.AuthType + DisableKeyPrefix + loginID
}

// getDisableServiceKey generates the storage key for service disable status. getDisableServiceKey 获取账号分类禁用状态存储键。
func (m *Manager) getDisableServiceKey(loginID, service string) string {
	return m.config.KeyPrefix + m.config.AuthType + DisableServiceKeyPrefix + loginID + ":" + service
}

// triggerEvent triggers an event through the event manager. triggerEvent 通过事件管理器触发事件（根据配置决定同步或异步）。
func (m *Manager) triggerEvent(event listener.Event, loginID, device, deviceId, token string, extra map[string]any) {
	if m.eventManager == nil {
		return
	}

	eventData := &listener.EventData{
		Event:     event,
		AuthType:  m.config.AuthType,
		LoginID:   loginID,
		Device:    device,
		DeviceId:  deviceId,
		Token:     token,
		Extra:     extra,
		Timestamp: time.Now().Unix(),
	}

	// Choose sync or async by config 根据配置决定同步或异步触发
	if m.config.AsyncEvent {
		// Trigger asynchronously 异步触发
		eventFunc := func() {
			m.eventManager.Trigger(eventData)
		}

		if m.pool != nil {
			_ = m.pool.Submit(eventFunc)
		} else {
			go eventFunc()
		}
	} else {
		// Trigger synchronously 同步触发
		m.eventManager.Trigger(eventData)
	}
}

// getExpiration calculates token expiration duration from configuration. getExpiration 从配置中计算 Token 过期时长。
func (m *Manager) getExpiration() time.Duration {
	if m.config.Timeout > 0 {
		return time.Duration(m.config.Timeout) * time.Second
	}
	return 0
}

// getDeviceAndDeviceId extracts device type and device ID from parameters. getDeviceAndDeviceId 获取设备类型和设备 ID。 规则：device 和 deviceId 是两个独立的过滤维度，互不影响
func (m *Manager) getDeviceAndDeviceId(deviceAndDeviceId ...string) (string, string) {
	device := ""
	deviceId := ""

	if len(deviceAndDeviceId) > 0 {
		device = strings.TrimSpace(deviceAndDeviceId[0])
	}

	if len(deviceAndDeviceId) > 1 {
		deviceId = strings.TrimSpace(deviceAndDeviceId[1])
	}

	return device, deviceId
}

// saveToStorage serializes and saves data to storage backend. saveToStorage 将指定类型的数据序列化并存储到存储后端。
func (m *Manager) saveToStorage(
	ctx context.Context,
	key string,
	value any,
	expiration ...time.Duration,
) error {

	// Serialize to bytes 序列化为字节
	bytesData, err := m.serializer.Encode(value)
	if err != nil {
		return fmt.Errorf("%w: %v", serror.ErrSerializeFailed, err)
	}

	// Build expiration duration 构建过期时长
	duration := m.getExpiration()
	if len(expiration) > 0 {
		duration = expiration[0]
	}

	// Persist to storage 存储到后端
	if err = m.storage.Set(ctx, key, bytesData, duration); err != nil {
		return fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
	}

	return nil
}

// searchKeys searches storage keys by pattern with pagination (internal method). searchKeys 根据模式搜索存储键并分页（内部方法）。
func (m *Manager) searchKeys(ctx context.Context, pattern string, start, size int) ([]string, error) {
	keys, err := m.storage.Keys(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", serror.ErrStorageUnavailable, err)
	}

	total := len(keys)
	if start < 0 {
		start = 0
	}
	if start >= total {
		return []string{}, nil
	}

	// Return all when size == -1 size == -1 表示返回全部
	end := total
	if size >= 0 {
		end = start + size
		if end > total {
			end = total
		}
	}

	return keys[start:end], nil
}
