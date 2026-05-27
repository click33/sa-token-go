package stputil

import "github.com/sa-tokens/sa-token-go/core/manager"

// StpInterface is an alias to manager.StpInterface | 权限数据源接口（与 manager 一致）
type StpInterface = manager.StpInterface

// DefaultStpInterface is an alias to manager.DefaultStpInterface | 默认实现
type DefaultStpInterface = manager.DefaultStpInterface

// SetStpInterface injects the global StpInterface implementation | 注入全局 StpInterface
func SetStpInterface(impl StpInterface) {
	manager.SetGlobalStpInterface(impl)
}

// GetStpInterface returns the current global StpInterface | 获取全局 StpInterface
func GetStpInterface() StpInterface {
	return manager.GetGlobalStpInterface()
}
