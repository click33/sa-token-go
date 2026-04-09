package manager

const (
	// DisableKeyPrefix stores disable key prefix DisableKeyPrefix 存储禁用状态存储前缀
	DisableKeyPrefix = "disable:"
	// DisableServiceKeyPrefix stores service disable prefix DisableServiceKeyPrefix 存储分类禁用状态存储前缀
	DisableServiceKeyPrefix = "disable:service:"
	// SessionKeyPrefix stores session key prefix SessionKeyPrefix 存储会话存储前缀
	SessionKeyPrefix = "session:"
	// RenewKeyPrefix stores renew key prefix RenewKeyPrefix 存储 Token 续期存储前缀
	RenewKeyPrefix = "renew:"
	// ActivePrefix stores active key prefix ActivePrefix 存储活跃时间存储前缀
	ActivePrefix = "active:"

	// SessionKeyLoginID stores session login id key SessionKeyLoginID 存储登录 ID 键名
	SessionKeyLoginID = "loginId"
	// SessionKeyDevice stores session device key SessionKeyDevice 存储设备类型键名
	SessionKeyDevice = "device"
	// SessionKeyLoginTime stores session login time key SessionKeyLoginTime 存储登录时间键名
	SessionKeyLoginTime = "loginTime"
	// SessionKeyPermissions stores permissions key SessionKeyPermissions 存储权限列表键名
	SessionKeyPermissions = "permissions"
	// SessionKeyRoles stores roles key SessionKeyRoles 存储角色列表键名
	SessionKeyRoles = "roles"

	// PermissionWildcard stores permission wildcard PermissionWildcard 存储全局权限通配符
	PermissionWildcard = "*"
	// PermissionSeparator stores permission separator PermissionSeparator 存储权限段分隔符
	PermissionSeparator = ":"
)

// TokenState defines token logical state TokenState 定义 Token 逻辑状态
type TokenState string

const (
	// TokenStateLogout indicates logout state TokenStateLogout 表示主动登出状态
	TokenStateLogout TokenState = "LOGOUT"
	// TokenStateKickOut indicates kickout state TokenStateKickOut 表示被踢下线状态
	TokenStateKickOut TokenState = "KICK_OUT"
	// TokenStateReplaced indicates replaced state TokenStateReplaced 表示被顶下线状态
	TokenStateReplaced TokenState = "REPLACED"
)
