package manager

import (
	"time"

	"github.com/sa-tokens/sa-token-go/core/errs"
)

// SafeKeyPrefix second-level auth storage namespace | 二级认证存储前缀
const SafeKeyPrefix = "safe:"

// DefaultSafeService default service tag for safe-auth | 二级认证默认 service
const DefaultSafeService = "important"

func (m *Manager) safeDefaultService(service string) string {
	if service != "" {
		return service
	}
	if m.config != nil && m.config.SafeAuthDefaultService != "" {
		return m.config.SafeAuthDefaultService
	}
	return DefaultSafeService
}

// OpenSafe enables second-level authentication | 开启二级认证
func (m *Manager) OpenSafe(tokenValue, service string, safeTime int64) error {
	if !m.IsLogin(tokenValue) {
		return ErrNotLogin
	}
	service = m.safeDefaultService(service)
	if safeTime <= 0 {
		return errs.ErrSafeTimeInvalid
	}
	key := m.getSafeKey(tokenValue, service)
	return m.storage.Set(key, "1", time.Duration(safeTime)*time.Second)
}

// IsSafe reports whether second-level auth is active | 是否已通过二级认证
func (m *Manager) IsSafe(tokenValue, service string) bool {
	if tokenValue == "" {
		return false
	}
	service = m.safeDefaultService(service)
	if !m.IsLogin(tokenValue) {
		return false
	}
	return m.storage.Exists(m.getSafeKey(tokenValue, service))
}

// CheckSafe asserts second-level auth | 校验二级认证
func (m *Manager) CheckSafe(tokenValue, service string) error {
	if !m.IsSafe(tokenValue, service) {
		return errs.ErrNotPassedSafeAuthWithService(m.safeDefaultService(service))
	}
	return nil
}

// GetSafeTime returns remaining TTL seconds, -2 if inactive | 二级认证剩余时间（秒）
func (m *Manager) GetSafeTime(tokenValue, service string) (int64, error) {
	if tokenValue == "" {
		return -2, nil
	}
	service = m.safeDefaultService(service)
	ttl, err := m.storage.TTL(m.getSafeKey(tokenValue, service))
	if err != nil {
		return -2, err
	}
	return int64(ttl.Seconds()), nil
}

// CloseSafe removes second-level auth marker | 关闭二级认证
func (m *Manager) CloseSafe(tokenValue, service string) error {
	if tokenValue == "" {
		return nil
	}
	service = m.safeDefaultService(service)
	return m.storage.Delete(m.getSafeKey(tokenValue, service))
}

func (m *Manager) getSafeKey(tokenValue, service string) string {
	return m.prefix + SafeKeyPrefix + service + ":" + tokenValue
}
