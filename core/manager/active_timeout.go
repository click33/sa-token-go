package manager

import (
	"encoding/json"
	"time"

	"github.com/sa-tokens/sa-token-go/core/errs"
	"github.com/sa-tokens/sa-token-go/core/session"
)

// UpdateLastActiveToNow refreshes TokenInfo.ActiveTime | 更新最后活跃时间
func (m *Manager) UpdateLastActiveToNow(tokenValue string) error {
	info, err := m.getTokenInfo(tokenValue, false)
	if err != nil {
		return err
	}
	if info == nil {
		return ErrInvalidTokenData
	}
	info.ActiveTime = time.Now().Unix()
	data, err := json.Marshal(info)
	if err != nil {
		return errs.ErrMarshalTokenInfo(err)
	}
	return m.storage.SetKeepTTL(m.getTokenKey(tokenValue), data)
}

// GetTokenLastActiveTime returns TokenInfo.ActiveTime | 最后活跃时间戳
func (m *Manager) GetTokenLastActiveTime(tokenValue string) (int64, error) {
	info, err := m.getTokenInfo(tokenValue, false)
	if err != nil {
		return 0, err
	}
	if info == nil {
		return 0, ErrInvalidTokenData
	}
	return info.ActiveTime, nil
}

// CheckActiveTimeout returns active-timeout error when idle too long | 校验活跃超时
func (m *Manager) CheckActiveTimeout(tokenValue string) error {
	if m.config.ActiveTimeout <= 0 {
		return nil
	}
	info, err := m.getTokenInfo(tokenValue)
	if err != nil {
		return err
	}
	if info == nil {
		return ErrInvalidTokenData
	}
	if time.Now().Unix()-info.ActiveTime > m.config.ActiveTimeout {
		return errs.ErrActiveTimeoutWithToken(tokenValue)
	}
	return nil
}

// GetTokenTimeout returns remaining TTL seconds for token mapping (-1 no limit / key missing semantics) | Token 剩余 TTL
func (m *Manager) GetTokenTimeout(tokenValue string) (int64, error) {
	if tokenValue == "" {
		return -2, nil
	}
	d, err := m.storage.TTL(m.getTokenKey(tokenValue))
	if err != nil {
		return -2, err
	}
	sec := int64(d.Seconds())
	if d < 0 {
		return -1, nil
	}
	return sec, nil
}

// GetSessionTimeout returns TTL for Account-Session | 账号 Session TTL
func (m *Manager) GetSessionTimeout(loginID string) (int64, error) {
	key := m.prefix + session.AccountSessionKeyPrefix + loginID
	d, err := m.storage.TTL(key)
	if err != nil {
		return -2, err
	}
	if d < 0 {
		return -1, nil
	}
	return int64(d.Seconds()), nil
}

// GetTokenSessionTimeout returns TTL for Token-Session storage | Token-Session TTL
func (m *Manager) GetTokenSessionTimeout(tokenValue string) (int64, error) {
	key := m.prefix + session.TokenSessionKeyPrefix + tokenValue
	d, err := m.storage.TTL(key)
	if err != nil {
		return -2, err
	}
	if d < 0 {
		return -1, nil
	}
	return int64(d.Seconds()), nil
}

// RenewTimeout renews token + account + account-session TTL | 续期
func (m *Manager) RenewTimeout(tokenValue string, seconds int64) error {
	if tokenValue == "" || seconds <= 0 {
		return nil
	}
	dur := time.Duration(seconds) * time.Second
	info, err := m.getTokenInfo(tokenValue, false)
	if err != nil || info == nil {
		return err
	}
	_ = m.storage.Expire(m.getTokenKey(tokenValue), dur)
	_ = m.storage.Expire(m.getAccountKey(info.LoginID, info.Device), dur)
	if sess, err := m.GetSession(info.LoginID); err == nil && sess != nil {
		_ = sess.Renew(dur)
	}
	return nil
}
