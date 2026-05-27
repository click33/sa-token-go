package stputil

import (
	"fmt"
	"sync"

	"github.com/sa-tokens/sa-token-go/core/adapter"
	"github.com/sa-tokens/sa-token-go/core/manager"
)

// MultiStpLogic manages multiple StpLogic instances by login type | 多账号体系管理器
type MultiStpLogic struct {
	mu      sync.RWMutex
	storage adapter.Storage
	logics  map[string]*StpLogic
}

// NewMultiStpLogic creates a new multi-account manager | 创建多账号体系管理器
func NewMultiStpLogic(storage adapter.Storage) *MultiStpLogic {
	return &MultiStpLogic{
		storage: storage,
		logics:  make(map[string]*StpLogic),
	}
}

// Register registers a StpLogic for a login type | 注册指定账号体系的StpLogic
func (m *MultiStpLogic) Register(loginType string, logic *StpLogic) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logics[loginType] = logic
}

// Get gets the StpLogic for a login type | 获取指定账号体系的StpLogic
func (m *MultiStpLogic) Get(loginType string) (*StpLogic, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	logic, ok := m.logics[loginType]
	return logic, ok
}

// GetOrCreate gets or creates a StpLogic for a login type with the given Manager | 获取或创建指定账号体系的StpLogic
func (m *MultiStpLogic) GetOrCreate(loginType string, mgr *manager.Manager) *StpLogic {
	m.mu.Lock()
	defer m.mu.Unlock()
	if logic, ok := m.logics[loginType]; ok {
		return logic
	}
	logic := NewStpLogic(mgr)
	m.logics[loginType] = logic
	return logic
}

// LoginByType performs login for a specific account type | 指定账号体系登录
func (m *MultiStpLogic) LoginByType(loginType string, loginID interface{}, device ...string) (string, error) {
	logic, ok := m.Get(loginType)
	if !ok {
		return "", fmt.Errorf("login type %q not registered", loginType)
	}
	return logic.Login(loginID, device...)
}

// GetLoginIDByType gets login ID for a specific account type | 获取指定账号体系的登录ID
func (m *MultiStpLogic) GetLoginIDByType(loginType string, tokenValue string) (string, error) {
	logic, ok := m.Get(loginType)
	if !ok {
		return "", fmt.Errorf("login type %q not registered", loginType)
	}
	return logic.GetLoginID(tokenValue)
}

// IsLoginByType checks login status for a specific account type | 检查指定账号体系的登录状态
func (m *MultiStpLogic) IsLoginByType(loginType string, tokenValue string) bool {
	logic, ok := m.Get(loginType)
	if !ok {
		return false
	}
	return logic.IsLogin(tokenValue)
}

// LogoutByType performs logout for a specific account type | 指定账号体系登出
func (m *MultiStpLogic) LogoutByType(loginType string, loginID interface{}, device ...string) error {
	logic, ok := m.Get(loginType)
	if !ok {
		return fmt.Errorf("login type %q not registered", loginType)
	}
	return logic.Logout(loginID, device...)
}

// GetTokenValueByType gets token value for a specific account type | 获取指定账号体系的Token值
func (m *MultiStpLogic) GetTokenValueByType(loginType string, loginID interface{}, device ...string) (string, error) {
	logic, ok := m.Get(loginType)
	if !ok {
		return "", fmt.Errorf("login type %q not registered", loginType)
	}
	return logic.GetTokenValue(loginID, device...)
}

// ListTypes returns all registered login types | 返回所有已注册的账号体系
func (m *MultiStpLogic) ListTypes() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	types := make([]string, 0, len(m.logics))
	for t := range m.logics {
		types = append(types, t)
	}
	return types
}
