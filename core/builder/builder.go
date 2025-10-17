package builder

import (
	"time"

	"github.com/click33/sa-token-go/core/adapter"
	"github.com/click33/sa-token-go/core/banner"
	"github.com/click33/sa-token-go/core/config"
	"github.com/click33/sa-token-go/core/manager"
)

// Builder Sa-Token builder for fluent configuration | Sa-Token构建器，用于流式配置
type Builder struct {
	storage       adapter.Storage
	tokenName     string
	timeout       int64
	activeTimeout int64
	isConcurrent  bool
	isShare       bool
	maxLoginCount int
	tokenStyle    config.TokenStyle
	autoRenew     bool
	jwtSecretKey  string
	isLog         bool
	isPrintBanner bool
	isReadBody    bool
	isReadHeader  bool
	isReadCookie  bool
}

// NewBuilder creates a new builder | 创建新的构建器
func NewBuilder() *Builder {
	return &Builder{
		tokenName:     "satoken",
		timeout:       2592000, // 30 days | 30天
		activeTimeout: -1,
		isConcurrent:  true,
		isShare:       true,
		maxLoginCount: 12,
		tokenStyle:    config.TokenStyleUUID,
		autoRenew:     true,
		isLog:         false,
		isPrintBanner: true,  // Print banner by default | 默认打印 Banner
		isReadBody:    false, // Don't read from body by default | 默认不从 Body 读取
		isReadHeader:  true,  // Read from header by default | 默认从 Header 读取
		isReadCookie:  false, // Don't read from cookie by default | 默认不从 Cookie 读取
	}
}

// Storage sets storage adapter | 设置存储适配器
func (b *Builder) Storage(storage adapter.Storage) *Builder {
	b.storage = storage
	return b
}

// TokenName sets token name | 设置Token名称
func (b *Builder) TokenName(name string) *Builder {
	b.tokenName = name
	return b
}

// Timeout sets timeout in seconds | 设置超时时间（秒）
func (b *Builder) Timeout(seconds int64) *Builder {
	b.timeout = seconds
	return b
}

// TimeoutDuration sets timeout with duration | 设置超时时间（时间段）
func (b *Builder) TimeoutDuration(d time.Duration) *Builder {
	b.timeout = int64(d.Seconds())
	return b
}

// ActiveTimeout sets active timeout in seconds | 设置活跃超时（秒）
func (b *Builder) ActiveTimeout(seconds int64) *Builder {
	b.activeTimeout = seconds
	return b
}

// IsConcurrent sets whether to allow concurrent login | 设置是否允许并发登录
func (b *Builder) IsConcurrent(concurrent bool) *Builder {
	b.isConcurrent = concurrent
	return b
}

// IsShare sets whether to share token | 设置是否共享Token
func (b *Builder) IsShare(share bool) *Builder {
	b.isShare = share
	return b
}

// MaxLoginCount sets maximum login count | 设置最大登录数量
func (b *Builder) MaxLoginCount(count int) *Builder {
	b.maxLoginCount = count
	return b
}

// TokenStyle sets token generation style | 设置Token风格
func (b *Builder) TokenStyle(style config.TokenStyle) *Builder {
	b.tokenStyle = style
	return b
}

// AutoRenew sets whether to auto-renew token | 设置是否自动续期
func (b *Builder) AutoRenew(autoRenew bool) *Builder {
	b.autoRenew = autoRenew
	return b
}

// JwtSecretKey sets JWT secret key | 设置JWT密钥
func (b *Builder) JwtSecretKey(key string) *Builder {
	b.jwtSecretKey = key
	return b
}

// IsLog sets whether to enable logging | 设置是否输出日志
func (b *Builder) IsLog(isLog bool) *Builder {
	b.isLog = isLog
	return b
}

// IsPrintBanner sets whether to print startup banner | 设置是否打印启动Banner
func (b *Builder) IsPrintBanner(isPrint bool) *Builder {
	b.isPrintBanner = isPrint
	return b
}

// IsReadBody sets whether to read token from request body | 设置是否从请求体读取Token
func (b *Builder) IsReadBody(isRead bool) *Builder {
	b.isReadBody = isRead
	return b
}

// IsReadHeader sets whether to read token from header | 设置是否从Header读取Token
func (b *Builder) IsReadHeader(isRead bool) *Builder {
	b.isReadHeader = isRead
	return b
}

// IsReadCookie sets whether to read token from cookie | 设置是否从Cookie读取Token
func (b *Builder) IsReadCookie(isRead bool) *Builder {
	b.isReadCookie = isRead
	return b
}

// Build builds Manager and prints startup banner | 构建Manager并打印启动Banner
func (b *Builder) Build() *manager.Manager {
	if b.storage == nil {
		panic("storage is required, please call Storage() method")
	}

	cfg := &config.Config{
		TokenName:              b.tokenName,
		Timeout:                b.timeout,
		ActiveTimeout:          b.activeTimeout,
		IsConcurrent:           b.isConcurrent,
		IsShare:                b.isShare,
		MaxLoginCount:          b.maxLoginCount,
		IsReadBody:             b.isReadBody,
		IsReadHeader:           b.isReadHeader,
		IsReadCookie:           b.isReadCookie,
		TokenStyle:             b.tokenStyle,
		DataRefreshPeriod:      -1,
		TokenSessionCheckLogin: true,
		AutoRenew:              b.autoRenew,
		JwtSecretKey:           b.jwtSecretKey,
		IsLog:                  b.isLog,
		IsPrintBanner:          b.isPrintBanner,
		CookieConfig: &config.CookieConfig{
			Domain:   "",
			Path:     "/",
			Secure:   false,
			HttpOnly: true,
			SameSite: "Lax",
			MaxAge:   0,
		},
	}

	// Print startup banner with full configuration | 打印启动Banner和完整配置
	// Only skip printing when both IsLog=false AND IsPrintBanner=false | 只有当 IsLog=false 且 IsPrintBanner=false 时才不打印
	if b.isPrintBanner || b.isLog {
		banner.PrintWithConfig(cfg)
	}

	mgr := manager.NewManager(b.storage, cfg)

	// Note: If you use the stputil package, it will automatically set the global Manager | 注意：如果你使用了 stputil 包，它会自动设置全局 Manager
	// We don't directly call stputil.SetManager here to avoid hard dependencies | 这里不直接调用 stputil.SetManager，避免强依赖

	return mgr
}
