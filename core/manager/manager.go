package manager

import (
	"github.com/click33/sa-token-go/core/adapter"
	"github.com/click33/sa-token-go/core/config"
	"github.com/click33/sa-token-go/core/listener"
	"github.com/click33/sa-token-go/core/nonce"
	"github.com/click33/sa-token-go/core/oauth2"
)

// Manager defines auth manager Manager 定义认证管理器
type Manager struct {
	config *config.Config // config stores auth config config 存储全局认证配置

	generator  adapter.Generator // generator stores token generator generator 存储 Token 生成器
	storage    adapter.Storage   // storage stores storage adapter storage 存储存储适配器
	serializer adapter.Codec     // serializer stores codec adapter serializer 存储编解码器适配器
	logger     adapter.Log       // logger stores log adapter logger 存储日志适配器
	pool       adapter.Pool      // pool stores async task pool pool 存储异步任务协程池组件

	nonceManager  *nonce.NonceManager  // nonceManager stores nonce manager nonceManager 存储 Nonce 管理器
	oauth2Manager *oauth2.OAuth2Server // oauth2Manager stores oauth2 manager oauth2Manager 存储 OAuth2 管理器
	eventManager  *listener.Manager    // eventManager stores event manager eventManager 存储事件管理器

	CustomPermissionListFunc    func(loginID, authType string) ([]string, error)                   // CustomPermissionListFunc stores custom permission callback CustomPermissionListFunc 存储自定义权限列表获取函数
	CustomRoleListFunc          func(loginID, authType string) ([]string, error)                   // CustomRoleListFunc stores custom role callback CustomRoleListFunc 存储自定义角色列表获取函数
	CustomPermissionListExtFunc func(loginID, device, deviceId, authType string) ([]string, error) // CustomPermissionListExtFunc stores extended permission callback CustomPermissionListExtFunc 存储扩展权限列表获取函数
	CustomRoleListExtFunc       func(loginID, device, deviceId, authType string) ([]string, error) // CustomRoleListExtFunc stores extended role callback CustomRoleListExtFunc 存储扩展角色列表获取函数
}

// TokenInfo defines token info TokenInfo 定义 Token 信息
type TokenInfo struct {
	AuthType   string `json:"authType"`   // AuthType stores auth type AuthType 存储认证体系类型
	LoginID    string `json:"loginId"`    // LoginID stores login id LoginID 存储登录 ID
	Device     string `json:"device"`     // Device stores device type Device 存储设备类型
	DeviceId   string `json:"deviceId"`   // DeviceId stores device id DeviceId 存储设备 ID
	CreateTime int64  `json:"createTime"` // CreateTime stores create time CreateTime 存储创建时间戳
}

// Session defines session object Session 定义用于存储用户数据的会话对象
type Session struct {
	AuthType             string         `json:"authType"`                       // AuthType stores auth type AuthType 存储认证体系类型
	LoginID              string         `json:"loginId"`                        // LoginID stores login id LoginID 存储登录 ID
	CreateTime           int64          `json:"createTime"`                     // CreateTime stores create time CreateTime 存储创建时间
	TerminalInfos        []TerminalInfo `json:"terminalInfos,omitempty"`        // TerminalInfos stores terminal list TerminalInfos 存储终端信息列表
	Permissions          []string       `json:"permissions,omitempty"`          // Permissions stores permission list Permissions 存储权限列表
	Roles                []string       `json:"roles,omitempty"`                // Roles stores role list Roles 存储角色列表
	HistoryTerminalCount int64          `json:"historyTerminalCount,omitempty"` // HistoryTerminalCount stores history count HistoryTerminalCount 存储历史总计登录设备数量
}

// TerminalInfo defines terminal info TerminalInfo 定义终端信息
type TerminalInfo struct {
	Token      string `json:"token"`      // Token stores token value Token 存储令牌值
	LoginID    string `json:"loginId"`    // LoginID stores login id LoginID 存储登录 ID
	Device     string `json:"device"`     // Device stores device type Device 存储设备类型
	DeviceId   string `json:"deviceId"`   // DeviceId stores device id DeviceId 存储设备 ID
	CreateTime int64  `json:"createTime"` // CreateTime stores token create time CreateTime 存储 Token 创建时间戳
	Index      int64  `json:"index"`      // Index stores history order Index 存储历史登录顺序索引
}

// DisableInfo defines account disable info DisableInfo 定义账号封禁信息
type DisableInfo struct {
	DisableTime   int64  `json:"disableTime"`   // DisableTime stores disable time DisableTime 存储封禁时间戳
	DisableReason string `json:"disableReason"` // DisableReason stores disable reason DisableReason 存储封禁原因
}

// ServiceDisableInfo defines service disable info ServiceDisableInfo 定义分类封禁信息
type ServiceDisableInfo struct {
	Service       string `json:"service"`       // Service stores disabled service Service 存储被封禁的服务模块
	Level         int    `json:"level"`         // Level stores disable level Level 存储封禁等级
	DisableTime   int64  `json:"disableTime"`   // DisableTime stores disable time DisableTime 存储封禁时间戳
	DisableReason string `json:"disableReason"` // DisableReason stores disable reason DisableReason 存储封禁原因
}
