package config

// SameSiteMode defines cookie sameSite mode SameSiteMode 定义 Cookie SameSite 属性值
type SameSiteMode string

const (
	// SameSiteStrict uses strict mode SameSiteStrict 使用严格模式
	SameSiteStrict SameSiteMode = "Strict"
	// SameSiteLax uses lax mode SameSiteLax 使用宽松模式
	SameSiteLax SameSiteMode = "Lax"
	// SameSiteNone uses none mode SameSiteNone 使用无约束模式
	SameSiteNone SameSiteMode = "None"
)

// ConcurrencyScope defines concurrency scope ConcurrencyScope 定义并发控制作用域
type ConcurrencyScope string

const (
	// ConcurrencyScopeAccount uses account scope ConcurrencyScopeAccount 使用账号级别作用域
	ConcurrencyScopeAccount ConcurrencyScope = "account"
	// ConcurrencyScopeDevice uses device scope ConcurrencyScopeDevice 使用设备级别作用域
	ConcurrencyScopeDevice ConcurrencyScope = "device"
)

const (
	// DefaultTokenName stores default token name DefaultTokenName 存储默认 Token 名称
	DefaultTokenName = "sa-token-go"
	// DefaultKeyPrefix stores default key prefix DefaultKeyPrefix 存储默认存储键前缀
	DefaultKeyPrefix = "sa-token-go:"
	// DefaultAuthType stores default auth type DefaultAuthType 存储默认认证体系类型
	DefaultAuthType = "auth:"
	// DefaultTimeout stores default timeout DefaultTimeout 存储默认 Token 超时时间
	DefaultTimeout = 2592000
	// DefaultMaxLoginCount stores default max login count DefaultMaxLoginCount 存储默认最大并发登录数
	DefaultMaxLoginCount = 12
	// DefaultCookiePath stores default cookie path DefaultCookiePath 存储默认 Cookie 路径
	DefaultCookiePath = "/"
	// NoLimit marks unlimited value NoLimit 标记无限制取值
	NoLimit = -1
)
