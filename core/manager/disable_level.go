package manager

import (
	"strconv"
	"time"

	"github.com/sa-tokens/sa-token-go/core/errs"
)

const (
	// DefaultDisableService default service dimension for tiered disable | 分级封禁默认 service
	DefaultDisableService = "login"
)

// DisableLevel sets disable marker for loginID+service | 分级封禁
func (m *Manager) DisableLevel(loginID, service string, level int, duration time.Duration) error {
	if loginID == "" {
		return errs.ErrLoginIDEmpty
	}
	if service == "" {
		service = DefaultDisableService
	}
	if level < MinDisableLevel {
		return errs.ErrInvalidDisableLevel
	}
	key := m.getDisableServiceKey(loginID, service)
	return m.storage.Set(key, strconv.Itoa(level), duration)
}

// GetDisableLevel returns current level or NotDisableLevel | 获取封禁等级
func (m *Manager) GetDisableLevel(loginID, service string) int {
	if loginID == "" {
		return NotDisableLevel
	}
	if service == "" {
		service = DefaultDisableService
	}
	v, err := m.storage.Get(m.getDisableServiceKey(loginID, service))
	if err != nil || v == nil {
		return NotDisableLevel
	}
	str, ok := assertString(v)
	if !ok {
		return NotDisableLevel
	}
	n, err := strconv.Atoi(str)
	if err != nil {
		return NotDisableLevel
	}
	return n
}

// IsDisableLevel reports whether current level >= threshold | 是否达到封禁等级
func (m *Manager) IsDisableLevel(loginID, service string, level int) bool {
	curr := m.GetDisableLevel(loginID, service)
	return curr != NotDisableLevel && curr >= level
}

// CheckDisableLevel returns error when disabled at threshold | 校验封禁等级
func (m *Manager) CheckDisableLevel(loginID, service string, level int) error {
	if m.IsDisableLevel(loginID, service, level) {
		if service == "" {
			service = DefaultDisableService
		}
		return errs.ErrDisableLevelExceededWithContext(loginID, service, level)
	}
	return nil
}

// UntieDisableServices removes tiered disable keys | 解封指定 service
func (m *Manager) UntieDisableServices(loginID string, services ...string) error {
	if loginID == "" {
		return errs.ErrLoginIDEmpty
	}
	if len(services) == 0 {
		return errs.ErrDisableServiceEmpty
	}
	for _, svc := range services {
		_ = m.storage.Delete(m.getDisableServiceKey(loginID, svc))
	}
	return nil
}

func (m *Manager) getDisableServiceKey(loginID, service string) string {
	return m.prefix + "disable:" + service + ":" + loginID
}
