package config

import (
	"fmt"
	"github.com/click33/sa-token-go/core/adapter"
	"strings"
)

// Config defines runtime config Config 定义运行时配置
type Config struct {
	// AuthType stores auth type AuthType 存储认证体系类型
	AuthType string

	// KeyPrefix stores storage key prefix KeyPrefix 存储存储键前缀
	KeyPrefix string

	// TokenName stores token name TokenName 存储 Token 名称
	TokenName string

	// Timeout stores token timeout seconds Timeout 存储 Token 超时时间秒数
	Timeout int64

	// AutoRenew controls auto renew AutoRenew 控制是否在校验时自动续期
	AutoRenew bool

	// RenewMaxRefresh stores renew trigger threshold RenewMaxRefresh 存储自动续期触发阈值
	RenewMaxRefresh int64

	// RenewInterval stores minimum renew interval RenewInterval 存储同一 Token 的最小续期间隔
	RenewInterval int64

	// ActiveTimeout stores max inactive duration ActiveTimeout 存储最大不活跃时长
	ActiveTimeout int64

	// ConcurrencyScope stores concurrency scope ConcurrencyScope 存储并发控制作用域
	ConcurrencyScope ConcurrencyScope

	// IsConcurrent controls concurrent login IsConcurrent 控制是否允许同一账号并发登录
	IsConcurrent bool

	// IsShare controls shared token IsShare 控制并发登录时是否共享同一 Token
	IsShare bool

	// MaxLoginCount stores max login count MaxLoginCount 存储同一账号最大登录数量
	MaxLoginCount int64

	// IsReadBody controls body token read IsReadBody 控制是否尝试从请求体读取 Token
	IsReadBody bool

	// IsReadHeader controls header token read IsReadHeader 控制是否尝试从 HTTP Header 读取 Token
	IsReadHeader bool

	// IsReadCookie controls cookie token read IsReadCookie 控制是否尝试从 Cookie 读取 Token
	IsReadCookie bool

	// TokenStyle stores token style TokenStyle 存储 Token 生成风格
	TokenStyle adapter.TokenStyle

	// JwtSecretKey stores JWT secret JwtSecretKey 存储 JWT 模式密钥
	JwtSecretKey string

	// IsLog controls logging IsLog 控制是否开启操作日志
	IsLog bool

	// IsPrintBanner controls banner print IsPrintBanner 控制是否打印启动 Banner
	IsPrintBanner bool

	// AsyncEvent controls async event AsyncEvent 控制是否异步触发事件
	AsyncEvent bool

	// CookieConfig stores cookie config CookieConfig 存储 Cookie 配置
	CookieConfig *CookieConfig
}

// CookieConfig defines cookie config CookieConfig 定义 Cookie 配置结构体
type CookieConfig struct {
	// Domain stores cookie domain Domain 存储 Cookie 作用域
	Domain string

	// Path stores cookie path Path 存储 Cookie 路径
	Path string

	// Secure controls HTTPS only Secure 控制是否仅在 HTTPS 下生效
	Secure bool

	// HttpOnly controls JavaScript access HttpOnly 控制是否禁止 JavaScript 访问 Cookie
	HttpOnly bool

	// SameSite stores sameSite mode SameSite 存储 SameSite 属性
	SameSite SameSiteMode

	// MaxAge stores cookie max age seconds MaxAge 存储 Cookie 过期时间秒数
	MaxAge int64
}

// DefaultConfig returns default config DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		AuthType:         DefaultAuthType,
		KeyPrefix:        DefaultKeyPrefix,
		TokenName:        DefaultTokenName,
		Timeout:          DefaultTimeout,
		AutoRenew:        true,
		RenewMaxRefresh:  DefaultTimeout / 2,
		RenewInterval:    NoLimit,
		ActiveTimeout:    NoLimit,
		ConcurrencyScope: ConcurrencyScopeAccount,
		IsConcurrent:     true,
		IsShare:          true,
		MaxLoginCount:    DefaultMaxLoginCount,
		IsReadBody:       false,
		IsReadHeader:     true,
		IsReadCookie:     false,
		TokenStyle:       adapter.TokenStyleUUID,
		JwtSecretKey:     "",
		IsLog:            false,
		IsPrintBanner:    true,
		AsyncEvent:       true,
		CookieConfig:     DefaultCookieConfig(),
	}
}

// DefaultCookieConfig returns default cookie config DefaultCookieConfig 返回默认 Cookie 配置
func DefaultCookieConfig() *CookieConfig {
	return &CookieConfig{
		Domain:   "",
		Path:     DefaultCookiePath,
		Secure:   false,
		HttpOnly: true,
		SameSite: SameSiteLax,
		MaxAge:   0,
	}
}

// Validate validates config Validate 验证配置是否合法
func (c *Config) Validate() error {
	// Validate basic format 验证基础格式
	if c.TokenName == "" {
		return fmt.Errorf("TokenName 不能为空")
	}
	if strings.ContainsAny(c.TokenName, "\t\r\n") {
		return fmt.Errorf("TokenName 不能包含制表符或换行符，当前值：%q", c.TokenName)
	}
	if len(c.TokenName) > 64 {
		return fmt.Errorf("TokenName 长度不能超过 64 个字符，当前长度：%d", len(c.TokenName))
	}

	if c.AuthType == "" {
		return fmt.Errorf("AuthType 不能为空")
	}
	if strings.ContainsAny(c.AuthType, "\t\r\n") {
		return fmt.Errorf("AuthType 不能包含制表符或换行符，当前值：%q", c.AuthType)
	}
	if len(c.AuthType) > 64 {
		return fmt.Errorf("AuthType 长度不能超过 64 个字符，当前长度：%d", len(c.AuthType))
	}

	// Validate numeric range 验证数值范围
	if err := c.checkNoLimits(); err != nil {
		return err
	}

	// Validate concurrency scope 验证并发作用域
	switch c.ConcurrencyScope {
	case ConcurrencyScopeAccount, ConcurrencyScopeDevice:
	default:
		return fmt.Errorf("ConcurrencyScope 必须为 %q 或 %q，当前值：%q",
			ConcurrencyScopeAccount, ConcurrencyScopeDevice, c.ConcurrencyScope)
	}

	// Validate token style settings 验证 Token 风格相关配置
	if c.TokenStyle == adapter.TokenStyleJWT && c.JwtSecretKey == "" {
		return fmt.Errorf("TokenStyle 为 JWT 时，JwtSecretKey 不能为空")
	}

	// Auto fix renew threshold 自动修正续期阈值
	if c.AutoRenew && c.Timeout != NoLimit && c.RenewMaxRefresh != NoLimit && c.RenewMaxRefresh > c.Timeout {
		c.RenewMaxRefresh = c.Timeout / 2
		if c.RenewMaxRefresh <= 0 {
			c.RenewMaxRefresh = c.Timeout
		}
	}

	// Validate time relation 验证时间关系
	if c.AutoRenew && c.ActiveTimeout != NoLimit && c.RenewInterval != NoLimit && c.RenewInterval >= c.ActiveTimeout {
		return fmt.Errorf("RenewInterval (%d) 必须小于 ActiveTimeout (%d)，否则活跃用户可能被踢出", c.RenewInterval, c.ActiveTimeout)
	}

	// Validate token sources 验证 Token 读取来源
	if !c.IsReadHeader && !c.IsReadCookie && !c.IsReadBody {
		return fmt.Errorf("至少需要启用 IsReadHeader、IsReadCookie 或 IsReadBody 中的一项")
	}

	// Validate cookie config 验证 Cookie 配置
	if c.IsReadCookie && c.CookieConfig == nil {
		return fmt.Errorf("启用 IsReadCookie 时，CookieConfig 不能为空")
	}
	if c.CookieConfig != nil {
		if err := c.validateCookieConfig(); err != nil {
			return err
		}
	}

	return nil
}

// validateCookieConfig validates cookie config validateCookieConfig 验证 Cookie 配置是否合法
func (c *Config) validateCookieConfig() error {
	cc := c.CookieConfig

	if cc.Path == "" {
		return fmt.Errorf("CookieConfig.Path 不能为空")
	}

	switch cc.SameSite {
	case SameSiteLax, SameSiteStrict, SameSiteNone, "":
	default:
		return fmt.Errorf("无效的 CookieConfig.SameSite 值：%v", cc.SameSite)
	}

	if cc.SameSite == SameSiteNone && !cc.Secure {
		return fmt.Errorf("SameSite 为 None 时，Secure 必须为 true（浏览器强制要求）")
	}

	return nil
}

// Clone clones config Clone 克隆配置
func (c *Config) Clone() *Config {
	newConfig := *c
	if c.CookieConfig != nil {
		cookieConfig := *c.CookieConfig
		newConfig.CookieConfig = &cookieConfig
	}
	return &newConfig
}

// SetAuthType sets auth type SetAuthType 设置认证体系类型
func (c *Config) SetAuthType(authType string) *Config {
	c.AuthType = authType
	return c
}

// SetKeyPrefix sets key prefix SetKeyPrefix 设置存储键前缀
func (c *Config) SetKeyPrefix(keyPrefix string) *Config {
	c.KeyPrefix = keyPrefix
	return c
}

// SetTokenName sets token name SetTokenName 设置 Token 名称
func (c *Config) SetTokenName(name string) *Config {
	c.TokenName = name
	return c
}

// SetTimeout sets timeout SetTimeout 设置超时时间
func (c *Config) SetTimeout(timeout int64) *Config {
	c.Timeout = timeout
	return c
}

// SetRenewMaxRefresh sets renew threshold SetRenewMaxRefresh 设置自动续期触发阈值
func (c *Config) SetRenewMaxRefresh(refresh int64) *Config {
	c.RenewMaxRefresh = refresh
	return c
}

// SetRenewInterval sets renew interval SetRenewInterval 设置最小续期间隔
func (c *Config) SetRenewInterval(interval int64) *Config {
	c.RenewInterval = interval
	return c
}

// SetActiveTimeout sets active timeout SetActiveTimeout 设置最大不活跃时长
func (c *Config) SetActiveTimeout(timeout int64) *Config {
	c.ActiveTimeout = timeout
	return c
}

// SetIsConcurrent sets concurrent switch SetIsConcurrent 设置是否允许并发登录
func (c *Config) SetIsConcurrent(isConcurrent bool) *Config {
	c.IsConcurrent = isConcurrent
	return c
}

// SetIsShare sets share switch SetIsShare 设置是否共享 Token
func (c *Config) SetIsShare(isShare bool) *Config {
	c.IsShare = isShare
	return c
}

// SetMaxLoginCount sets max login count SetMaxLoginCount 设置最大登录数量
func (c *Config) SetMaxLoginCount(count int64) *Config {
	c.MaxLoginCount = count
	return c
}

// SetIsReadBody sets body read switch SetIsReadBody 设置是否从请求体读取 Token
func (c *Config) SetIsReadBody(isReadBody bool) *Config {
	c.IsReadBody = isReadBody
	return c
}

// SetIsReadHeader sets header read switch SetIsReadHeader 设置是否从 Header 读取 Token
func (c *Config) SetIsReadHeader(isReadHeader bool) *Config {
	c.IsReadHeader = isReadHeader
	return c
}

// SetIsReadCookie sets cookie read switch SetIsReadCookie 设置是否从 Cookie 读取 Token
func (c *Config) SetIsReadCookie(isReadCookie bool) *Config {
	c.IsReadCookie = isReadCookie
	return c
}

// SetTokenStyle sets token style SetTokenStyle 设置 Token 生成风格
func (c *Config) SetTokenStyle(style adapter.TokenStyle) *Config {
	c.TokenStyle = style
	return c
}

// SetJwtSecretKey sets JWT secret SetJwtSecretKey 设置 JWT 密钥
func (c *Config) SetJwtSecretKey(key string) *Config {
	c.JwtSecretKey = key
	return c
}

// SetAutoRenew sets auto renew switch SetAutoRenew 设置是否自动续期
func (c *Config) SetAutoRenew(autoRenew bool) *Config {
	c.AutoRenew = autoRenew
	return c
}

// SetIsLog sets log switch SetIsLog 设置是否开启日志
func (c *Config) SetIsLog(isLog bool) *Config {
	c.IsLog = isLog
	return c
}

// SetIsPrintBanner sets banner print switch SetIsPrintBanner 设置是否打印启动 Banner
func (c *Config) SetIsPrintBanner(isPrint bool) *Config {
	c.IsPrintBanner = isPrint
	return c
}

// SetAsyncEvent sets async event switch SetAsyncEvent 设置是否异步触发事件
func (c *Config) SetAsyncEvent(asyncEvent bool) *Config {
	c.AsyncEvent = asyncEvent
	return c
}

// SetCookieConfig sets cookie config SetCookieConfig 设置 Cookie 配置
func (c *Config) SetCookieConfig(cookieConfig *CookieConfig) *Config {
	if cookieConfig != nil {
		c.CookieConfig = cookieConfig
	}
	return c
}

// checkNoLimits validates no limit fields checkNoLimits 验证无限制数值字段
func (c *Config) checkNoLimits() error {
	fields := map[string]int64{
		"Timeout":         c.Timeout,
		"RenewMaxRefresh": c.RenewMaxRefresh,
		"RenewInterval":   c.RenewInterval,
		"ActiveTimeout":   c.ActiveTimeout,
		"MaxLoginCount":   c.MaxLoginCount,
	}

	for name, value := range fields {
		if value == -1 || value > 0 {
			continue
		}
		return fmt.Errorf("%s 必须为 -1（无限制）或大于 0，当前值：%d", name, value)
	}
	return nil
}
