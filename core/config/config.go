package config

// TokenStyle Token generation style | Token生成风格
type TokenStyle string

const (
	// TokenStyleUUID UUID style | UUID风格
	TokenStyleUUID TokenStyle = "uuid"
	// TokenStyleSimple Simple random string | 简单随机字符串
	TokenStyleSimple TokenStyle = "simple"
	// TokenStyleRandom32 32-bit random string | 32位随机字符串
	TokenStyleRandom32 TokenStyle = "random32"
	// TokenStyleRandom64 64-bit random string | 64位随机字符串
	TokenStyleRandom64 TokenStyle = "random64"
	// TokenStyleRandom128 128-bit random string | 128位随机字符串
	TokenStyleRandom128 TokenStyle = "random128"
	// TokenStyleJWT JWT style | JWT风格
	TokenStyleJWT TokenStyle = "jwt"
	// TokenStyleHash SHA256 hash-based style | SHA256哈希风格
	TokenStyleHash TokenStyle = "hash"
	// TokenStyleTimestamp Timestamp-based style | 时间戳风格
	TokenStyleTimestamp TokenStyle = "timestamp"
	// TokenStyleTik Short ID style (like TikTok) | Tik风格短ID（类似抖音）
	TokenStyleTik TokenStyle = "tik"
)

// Config Sa-Token configuration | Sa-Token配置
type Config struct {
	// TokenName Token name (also used as Cookie name) | Token名称（同时也是Cookie名称）
	TokenName string

	// Timeout Token expiration time in seconds, -1 for never expire | Token超时时间（单位：秒，-1代表永不过期）
	Timeout int64

	// ActiveTimeout Token minimum activity frequency in seconds. If Token is not accessed for this time, it will be frozen. -1 means no limit | Token最低活跃频率（单位：秒），如果Token超过此时间没有访问，则会被冻结。-1代表不限制，永不冻结
	ActiveTimeout int64

	// IsConcurrent Allow concurrent login for the same account (true=allow concurrent login, false=new login kicks out old login) | 是否允许同一账号并发登录（为true时允许一起登录，为false时新登录挤掉旧登录）
	IsConcurrent bool

	// IsShare Share the same Token for concurrent logins (true=share one Token, false=create new Token for each login) | 在多人登录同一账号时，是否共用一个Token（为true时所有登录共用一个Token，为false时每次登录新建一个Token）
	IsShare bool

	// MaxLoginCount Maximum number of concurrent logins for the same account, -1 means no limit (only effective when IsConcurrent=true and IsShare=false) | 同一账号最大登录数量，-1代表不限（只有在IsConcurrent=true，IsShare=false时此配置才有效）
	MaxLoginCount int

	// IsReadBody Try to read Token from request body (default: false) | 是否尝试从请求体里读取Token（默认：false）
	IsReadBody bool

	// IsReadHeader Try to read Token from HTTP Header (default: true, recommended) | 是否尝试从Header里读取Token（默认：true，推荐）
	IsReadHeader bool

	// IsReadCookie Try to read Token from Cookie (default: false) | 是否尝试从Cookie里读取Token（默认：false）
	IsReadCookie bool

	// TokenStyle Token generation style | Token风格
	TokenStyle TokenStyle

	// DataRefreshPeriod Auto-refresh period in seconds, -1 means no auto-refresh | 自动续签（单位：秒），-1代表不自动续签
	DataRefreshPeriod int64

	// TokenSessionCheckLogin Check if Token-Session is kicked out when logging in (true=check on login, false=skip check) | Token-Session在登录时是否检查（true=登录时验证是否被踢下线，false=不作此检查）
	TokenSessionCheckLogin bool

	// AutoRenew Auto-renew Token expiration time on each validation | 是否自动续期（每次验证Token时，都会延长Token的有效期）
	AutoRenew bool

	// JwtSecretKey JWT secret key (only effective when TokenStyle=JWT) | JWT密钥（只有TokenStyle=JWT时，此配置才生效）
	JwtSecretKey string

	// IsLog Enable operation logging | 是否输出操作日志
	IsLog bool

	// IsPrintBanner Print startup banner (default: true) | 是否打印启动 Banner（默认：true）
	IsPrintBanner bool

	// CookieConfig Cookie configuration | Cookie配置
	CookieConfig *CookieConfig
}

// CookieConfig Cookie configuration | Cookie配置
type CookieConfig struct {
	// Domain Cookie domain | 作用域
	Domain string

	// Path Cookie path | 路径
	Path string

	// Secure Only effective under HTTPS | 是否只在HTTPS下生效
	Secure bool

	// HttpOnly Prevent JavaScript access to Cookie | 是否禁止JS操作Cookie
	HttpOnly bool

	// SameSite SameSite attribute (Strict, Lax, None) | SameSite属性（Strict、Lax、None）
	SameSite string

	// MaxAge Cookie expiration time in seconds | 过期时间（单位：秒）
	MaxAge int
}

// DefaultConfig Returns default configuration | 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		TokenName:              "sa-token",
		Timeout:                2592000, // 30 days | 30天
		ActiveTimeout:          -1,      // No limit | 不限制
		IsConcurrent:           true,    // Allow concurrent login | 允许并发登录
		IsShare:                true,    // Share Token | 共享Token
		MaxLoginCount:          12,      // Max 12 logins | 最多12个
		IsReadBody:             false,   // Don't read from Body (default) | 不从Body读取（默认）
		IsReadHeader:           true,    // Read from Header (recommended) | 从Header读取（推荐）
		IsReadCookie:           false,   // Don't read from Cookie (default) | 不从Cookie读取（默认）
		TokenStyle:             TokenStyleUUID,
		DataRefreshPeriod:      -1,    // No auto-refresh | 不自动续签
		TokenSessionCheckLogin: true,  // Check on login | 登录时检查
		AutoRenew:              true,  // Auto-renew | 自动续期
		JwtSecretKey:           "",    // Empty by default | 默认空
		IsLog:                  false, // No logging | 不输出日志
		IsPrintBanner:          true,  // Print startup banner | 打印启动 Banner
		CookieConfig: &CookieConfig{
			Domain:   "",
			Path:     "/",
			Secure:   false,
			HttpOnly: true,
			SameSite: "Lax",
			MaxAge:   0,
		},
	}
}

// Clone Clone configuration | 克隆配置
func (c *Config) Clone() *Config {
	newConfig := *c
	if c.CookieConfig != nil {
		cookieConfig := *c.CookieConfig
		newConfig.CookieConfig = &cookieConfig
	}
	return &newConfig
}

// SetTokenName Set Token name | 设置Token名称
func (c *Config) SetTokenName(name string) *Config {
	c.TokenName = name
	return c
}

// SetTimeout Set timeout duration | 设置超时时间
func (c *Config) SetTimeout(timeout int64) *Config {
	c.Timeout = timeout
	return c
}

// SetActiveTimeout Set active timeout duration | 设置活跃超时时间
func (c *Config) SetActiveTimeout(timeout int64) *Config {
	c.ActiveTimeout = timeout
	return c
}

// SetIsConcurrent Set whether to allow concurrent login | 设置是否允许并发登录
func (c *Config) SetIsConcurrent(isConcurrent bool) *Config {
	c.IsConcurrent = isConcurrent
	return c
}

// SetIsShare Set whether to share Token | 设置是否共享Token
func (c *Config) SetIsShare(isShare bool) *Config {
	c.IsShare = isShare
	return c
}

// SetTokenStyle Set Token generation style | 设置Token风格
func (c *Config) SetTokenStyle(style TokenStyle) *Config {
	c.TokenStyle = style
	return c
}

// SetJwtSecretKey Set JWT secret key | 设置JWT密钥
func (c *Config) SetJwtSecretKey(key string) *Config {
	c.JwtSecretKey = key
	return c
}

// SetAutoRenew Set whether to auto-renew Token | 设置是否自动续期
func (c *Config) SetAutoRenew(autoRenew bool) *Config {
	c.AutoRenew = autoRenew
	return c
}

// SetIsLog Set whether to enable logging | 设置是否输出日志
func (c *Config) SetIsLog(isLog bool) *Config {
	c.IsLog = isLog
	return c
}
