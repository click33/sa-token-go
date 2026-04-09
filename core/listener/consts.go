package listener

// Event defines authentication event type Event 定义认证事件类型
type Event string

const (
	// EventLogin indicates login event EventLogin 表示用户登录事件
	EventLogin Event = "login"
	// EventLogout indicates logout event EventLogout 表示用户登出事件
	EventLogout Event = "logout"
	// EventKickout indicates kickout event EventKickout 表示用户被踢下线事件
	EventKickout Event = "kickout"
	// EventReplace indicates replace event EventReplace 表示用户被顶下线事件
	EventReplace Event = "replace"
	// EventDisable indicates disable event EventDisable 表示账号被禁用事件
	EventDisable Event = "disable"
	// EventUntie indicates untie event EventUntie 表示账号解禁事件
	EventUntie Event = "untie"
	// EventRenew indicates renew event EventRenew 表示 Token 续期事件
	EventRenew Event = "renew"
	// EventCreateSession indicates create session event EventCreateSession 表示 Session 创建事件
	EventCreateSession Event = "createSession"
	// EventDestroySession indicates destroy session event EventDestroySession 表示 Session 销毁事件
	EventDestroySession Event = "destroySession"
	// EventPermissionCheck indicates permission check event EventPermissionCheck 表示权限检查事件
	EventPermissionCheck Event = "permissionCheck"
	// EventRoleCheck indicates role check event EventRoleCheck 表示角色检查事件
	EventRoleCheck Event = "roleCheck"
	// EventDisableService indicates disable service event EventDisableService 表示账号服务被封禁事件
	EventDisableService Event = "disableService"
	// EventUntieService indicates untie service event EventUntieService 表示账号服务解禁事件
	EventUntieService Event = "untieService"
	// EventAll indicates wildcard event EventAll 表示匹配所有事件的通配符事件
	EventAll Event = "*"
)

const (
	// ExtraKeyPermission stores permission key ExtraKeyPermission 存储单个权限字段
	ExtraKeyPermission = "permission"
	// ExtraKeyPermissions stores permissions key ExtraKeyPermissions 存储多个权限字段
	ExtraKeyPermissions = "permissions"
	// ExtraKeyRole stores role key ExtraKeyRole 存储单个角色字段
	ExtraKeyRole = "role"
	// ExtraKeyRoles stores roles key ExtraKeyRoles 存储多个角色字段
	ExtraKeyRoles = "roles"
	// ExtraKeyLogic stores logic key ExtraKeyLogic 存储逻辑类型字段
	ExtraKeyLogic = "logic"
	// ExtraKeyResult stores result key ExtraKeyResult 存储检查结果字段
	ExtraKeyResult = "result"
	// ExtraKeyService stores service key ExtraKeyService 存储服务模块字段
	ExtraKeyService = "service"
	// ExtraKeyLevel stores level key ExtraKeyLevel 存储封禁等级字段
	ExtraKeyLevel = "level"
)

const (
	// LogicAnd indicates AND logic LogicAnd 表示需要满足所有条件的 AND 逻辑
	LogicAnd = "AND"
	// LogicOr indicates OR logic LogicOr 表示满足任一条件即可的 OR 逻辑
	LogicOr = "OR"
)
