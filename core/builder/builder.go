package builder

import (
	sjson "github.com/click33/sa-token-go/com/codec/json"
	"github.com/click33/sa-token-go/com/generator/sgenerator"
	"github.com/click33/sa-token-go/com/log/nop"
	"github.com/click33/sa-token-go/com/log/slog"
	"github.com/click33/sa-token-go/com/pool/ants"
	"github.com/click33/sa-token-go/com/storage/memory"
	"github.com/click33/sa-token-go/core/adapter"
	"github.com/click33/sa-token-go/core/banner"
	"github.com/click33/sa-token-go/core/config"
	"github.com/click33/sa-token-go/core/manager"
	"strings"
	"time"
)

// Builder defines manager builder Builder 定义 Manager 构建器
type Builder struct {
	authType         string                  // authType stores auth type authType 存储认证体系类型
	keyPrefix        string                  // keyPrefix stores storage key prefix keyPrefix 存储存储键前缀
	tokenName        string                  // tokenName stores token name tokenName 存储 Token 名称
	timeout          int64                   // timeout stores token timeout seconds timeout 存储 Token 超时时间秒数
	autoRenew        bool                    // autoRenew controls auto renew autoRenew 控制是否启用自动续期
	renewMaxRefresh  int64                   // renewMaxRefresh stores renew trigger threshold renewMaxRefresh 存储续期触发阈值
	renewInterval    int64                   // renewInterval stores minimum renew interval renewInterval 存储最小续期间隔
	activeTimeout    int64                   // activeTimeout stores max inactive duration activeTimeout 存储最大不活跃时长
	concurrencyScope config.ConcurrencyScope // concurrencyScope stores concurrency scope concurrencyScope 存储并发控制作用域
	isConcurrent     bool                    // isConcurrent controls concurrent login isConcurrent 控制是否允许并发登录
	isShare          bool                    // isShare controls shared token isShare 控制是否共享同一 Token
	maxLoginCount    int64                   // maxLoginCount stores max login count maxLoginCount 存储最大并发登录数量
	isReadBody       bool                    // isReadBody controls body token read isReadBody 控制是否从请求体读取 Token
	isReadHeader     bool                    // isReadHeader controls header token read isReadHeader 控制是否从 Header 读取 Token
	isReadCookie     bool                    // isReadCookie controls cookie token read isReadCookie 控制是否从 Cookie 读取 Token
	tokenStyle       adapter.TokenStyle      // tokenStyle stores token style tokenStyle 存储 Token 生成风格
	jwtSecretKey     string                  // jwtSecretKey stores JWT secret key jwtSecretKey 存储 JWT 密钥
	isLog            bool                    // isLog controls logging isLog 控制是否开启日志输出
	isPrintBanner    bool                    // isPrintBanner controls banner print isPrintBanner 控制是否打印启动 Banner
	asyncEvent       bool                    // asyncEvent controls async event asyncEvent 控制是否异步触发事件

	cookieConfig    *config.CookieConfig  // cookieConfig stores cookie config cookieConfig 存储 Cookie 配置
	renewPoolConfig *ants.RenewPoolConfig // renewPoolConfig stores renew pool config renewPoolConfig 存储续期协程池配置
	logConfig       *slog.LoggerConfig    // logConfig stores logger config logConfig 存储日志配置

	generator adapter.Generator // generator stores token generator generator 存储 Token 生成器
	storage   adapter.Storage   // storage stores storage adapter storage 存储存储适配器
	codec     adapter.Codec     // codec stores codec adapter codec 存储编解码器适配器
	log       adapter.Log       // log stores log adapter log 存储日志适配器
	pool      adapter.Pool      // pool stores async task pool pool 存储异步任务协程池组件

	customPermissionListFunc    func(loginID, authType string) ([]string, error)                   // customPermissionListFunc stores custom permission callback customPermissionListFunc 存储自定义权限列表回调
	customRoleListFunc          func(loginID, authType string) ([]string, error)                   // customRoleListFunc stores custom role callback customRoleListFunc 存储自定义角色列表回调
	customPermissionListExtFunc func(loginID, device, deviceId, authType string) ([]string, error) // customPermissionListExtFunc stores custom extended permission callback customPermissionListExtFunc 存储扩展权限列表回调
	customRoleListExtFunc       func(loginID, device, deviceId, authType string) ([]string, error) // customRoleListExtFunc stores custom extended role callback customRoleListExtFunc 存储扩展角色列表回调
}

// NewBuilder creates builder with default config NewBuilder 创建使用默认配置的构建器
func NewBuilder() *Builder {
	return &Builder{
		authType:         config.DefaultAuthType,
		keyPrefix:        config.DefaultKeyPrefix,
		tokenName:        config.DefaultTokenName,
		timeout:          config.DefaultTimeout,
		autoRenew:        true,
		renewMaxRefresh:  config.DefaultTimeout / 2,
		renewInterval:    config.NoLimit,
		activeTimeout:    config.NoLimit,
		concurrencyScope: config.ConcurrencyScopeAccount,
		isConcurrent:     true,
		isShare:          false,
		maxLoginCount:    config.NoLimit,
		isReadBody:       false,
		isReadHeader:     true,
		isReadCookie:     false,
		tokenStyle:       adapter.TokenStyleUUID,
		jwtSecretKey:     sgenerator.DefaultJWTSecret,
		isLog:            false,
		isPrintBanner:    true,
		asyncEvent:       true,

		cookieConfig:    config.DefaultCookieConfig(),
		renewPoolConfig: ants.DefaultRenewPoolConfig(),
	}
}

// AuthType sets auth type with suffix fix AuthType 设置认证体系类型并自动补全冒号
func (b *Builder) AuthType(authType string) *Builder {
	if authType == "" {
		b.authType = config.DefaultAuthType
	} else if !strings.HasSuffix(authType, ":") {
		b.authType = authType + ":"
	} else {
		b.authType = authType
	}
	return b
}

// KeyPrefix sets key prefix with suffix fix KeyPrefix 设置存储键前缀并自动补全冒号
func (b *Builder) KeyPrefix(keyPrefix string) *Builder {
	if keyPrefix == "" {
		b.keyPrefix = config.DefaultKeyPrefix
	} else if !strings.HasSuffix(keyPrefix, ":") {
		b.keyPrefix = keyPrefix + ":"
	} else {
		b.keyPrefix = keyPrefix
	}
	return b
}

// TokenName sets token name TokenName 设置 Token 名称
func (b *Builder) TokenName(name string) *Builder {
	b.tokenName = name
	return b
}

// Timeout sets timeout seconds Timeout 设置超时时间秒数
func (b *Builder) Timeout(seconds int64) *Builder {
	b.timeout = seconds
	return b
}

// TimeoutDuration sets timeout by duration TimeoutDuration 按时间段设置超时时间
func (b *Builder) TimeoutDuration(d time.Duration) *Builder {
	b.timeout = int64(d.Seconds())
	return b
}

// AutoRenew sets auto renew switch AutoRenew 设置是否启用自动续期
func (b *Builder) AutoRenew(autoRenew bool) *Builder {
	b.autoRenew = autoRenew
	return b
}

// RenewMaxRefresh sets renew trigger threshold RenewMaxRefresh 设置自动续期触发阈值
func (b *Builder) RenewMaxRefresh(seconds int64) *Builder {
	b.renewMaxRefresh = seconds
	return b
}

// RenewInterval sets minimum renew interval RenewInterval 设置最小续期间隔
func (b *Builder) RenewInterval(seconds int64) *Builder {
	b.renewInterval = seconds
	return b
}

// ActiveTimeout sets max inactive duration ActiveTimeout 设置最大不活跃时长
func (b *Builder) ActiveTimeout(seconds int64) *Builder {
	b.activeTimeout = seconds
	return b
}

// ConcurrencyScope sets concurrency scope ConcurrencyScope 设置并发控制作用域
func (b *Builder) ConcurrencyScope(concurrencyScope config.ConcurrencyScope) *Builder {
	b.concurrencyScope = concurrencyScope
	return b
}

// IsConcurrent sets concurrent login switch IsConcurrent 设置是否允许并发登录
func (b *Builder) IsConcurrent(concurrent bool) *Builder {
	b.isConcurrent = concurrent
	return b
}

// IsShare sets shared token switch IsShare 设置是否共享同一 Token
func (b *Builder) IsShare(share bool) *Builder {
	b.isShare = share
	return b
}

// MaxLoginCount sets max login count MaxLoginCount 设置最大并发登录数量
func (b *Builder) MaxLoginCount(count int64) *Builder {
	b.maxLoginCount = count
	return b
}

// IsReadBody sets body read switch IsReadBody 设置是否从请求体读取 Token
func (b *Builder) IsReadBody(isRead bool) *Builder {
	b.isReadBody = isRead
	return b
}

// IsReadHeader sets header read switch IsReadHeader 设置是否从 HTTP Header 读取 Token
func (b *Builder) IsReadHeader(isRead bool) *Builder {
	b.isReadHeader = isRead
	return b
}

// IsReadCookie sets cookie read switch IsReadCookie 设置是否从 Cookie 读取 Token
func (b *Builder) IsReadCookie(isRead bool) *Builder {
	b.isReadCookie = isRead
	return b
}

// TokenStyle sets token style TokenStyle 设置 Token 生成风格
func (b *Builder) TokenStyle(style adapter.TokenStyle) *Builder {
	b.tokenStyle = style
	return b
}

// JwtSecretKey sets JWT secret key JwtSecretKey 设置 JWT 密钥
func (b *Builder) JwtSecretKey(key string) *Builder {
	b.jwtSecretKey = key
	return b
}

// IsLog sets log switch IsLog 设置是否开启日志输出
func (b *Builder) IsLog(isLog bool) *Builder {
	b.isLog = isLog
	return b
}

// IsPrintBanner sets banner print switch IsPrintBanner 设置是否打印启动 Banner
func (b *Builder) IsPrintBanner(isPrint bool) *Builder {
	b.isPrintBanner = isPrint
	return b
}

// AsyncEvent sets async event switch AsyncEvent 设置是否异步触发事件
func (b *Builder) AsyncEvent(asyncEvent bool) *Builder {
	b.asyncEvent = asyncEvent
	return b
}

// CookieDomain sets cookie domain CookieDomain 设置 Cookie 作用域 Domain
func (b *Builder) CookieDomain(domain string) *Builder {
	if b.cookieConfig == nil {
		b.cookieConfig = &config.CookieConfig{}
	}
	b.cookieConfig.Domain = domain
	return b
}

// CookiePath sets cookie path CookiePath 设置 Cookie 路径
func (b *Builder) CookiePath(path string) *Builder {
	if b.cookieConfig == nil {
		b.cookieConfig = &config.CookieConfig{}
	}
	b.cookieConfig.Path = path
	return b
}

// CookieSecure sets cookie secure switch CookieSecure 设置 Cookie 是否仅在 HTTPS 下生效
func (b *Builder) CookieSecure(secure bool) *Builder {
	if b.cookieConfig == nil {
		b.cookieConfig = &config.CookieConfig{}
	}
	b.cookieConfig.Secure = secure
	return b
}

// CookieHttpOnly sets cookie httpOnly switch CookieHttpOnly 设置 Cookie 是否禁止 JavaScript 访问
func (b *Builder) CookieHttpOnly(httpOnly bool) *Builder {
	if b.cookieConfig == nil {
		b.cookieConfig = &config.CookieConfig{}
	}
	b.cookieConfig.HttpOnly = httpOnly
	return b
}

// CookieSameSite sets cookie sameSite CookieSameSite 设置 Cookie SameSite 属性
func (b *Builder) CookieSameSite(sameSite config.SameSiteMode) *Builder {
	if b.cookieConfig == nil {
		b.cookieConfig = &config.CookieConfig{}
	}
	b.cookieConfig.SameSite = sameSite
	return b
}

// CookieMaxAge sets cookie max age CookieMaxAge 设置 Cookie 过期时间秒数
func (b *Builder) CookieMaxAge(maxAge int64) *Builder {
	if b.cookieConfig == nil {
		b.cookieConfig = &config.CookieConfig{}
	}
	b.cookieConfig.MaxAge = maxAge
	return b
}

// CookieConfig sets full cookie config CookieConfig 设置完整 Cookie 配置
func (b *Builder) CookieConfig(cfg *config.CookieConfig) *Builder {
	b.cookieConfig = cfg
	return b
}

// RenewPoolMinSize sets renew pool min size RenewPoolMinSize 设置续期协程池最小协程数
func (b *Builder) RenewPoolMinSize(size int) *Builder {
	if b.renewPoolConfig == nil {
		b.renewPoolConfig = &ants.RenewPoolConfig{}
	}
	b.renewPoolConfig.MinSize = size
	return b
}

// RenewPoolMaxSize sets renew pool max size RenewPoolMaxSize 设置续期协程池最大协程数
func (b *Builder) RenewPoolMaxSize(size int) *Builder {
	if b.renewPoolConfig == nil {
		b.renewPoolConfig = &ants.RenewPoolConfig{}
	}
	b.renewPoolConfig.MaxSize = size
	return b
}

// RenewPoolScaleUpRate sets renew pool scale up rate RenewPoolScaleUpRate 设置协程池扩容触发比例
func (b *Builder) RenewPoolScaleUpRate(rate float64) *Builder {
	if b.renewPoolConfig == nil {
		b.renewPoolConfig = &ants.RenewPoolConfig{}
	}
	b.renewPoolConfig.ScaleUpRate = rate
	return b
}

// RenewPoolScaleDownRate sets renew pool scale down rate RenewPoolScaleDownRate 设置协程池缩容触发比例
func (b *Builder) RenewPoolScaleDownRate(rate float64) *Builder {
	if b.renewPoolConfig == nil {
		b.renewPoolConfig = &ants.RenewPoolConfig{}
	}
	b.renewPoolConfig.ScaleDownRate = rate
	return b
}

// RenewPoolCheckInterval sets pool check interval RenewPoolCheckInterval 设置协程池扩缩容检查间隔
func (b *Builder) RenewPoolCheckInterval(interval time.Duration) *Builder {
	if b.renewPoolConfig == nil {
		b.renewPoolConfig = &ants.RenewPoolConfig{}
	}
	b.renewPoolConfig.CheckInterval = interval
	return b
}

// RenewPoolExpiry sets pool worker expiry RenewPoolExpiry 设置协程池空闲协程过期时间
func (b *Builder) RenewPoolExpiry(duration time.Duration) *Builder {
	if b.renewPoolConfig == nil {
		b.renewPoolConfig = &ants.RenewPoolConfig{}
	}
	b.renewPoolConfig.Expiry = duration
	return b
}

// RenewPoolPrintStatusInterval sets pool status print interval RenewPoolPrintStatusInterval 设置协程池状态打印间隔
func (b *Builder) RenewPoolPrintStatusInterval(interval time.Duration) *Builder {
	if b.renewPoolConfig == nil {
		b.renewPoolConfig = &ants.RenewPoolConfig{}
	}
	b.renewPoolConfig.PrintStatusInterval = interval
	return b
}

// RenewPoolPreAlloc sets pool pre alloc switch RenewPoolPreAlloc 设置协程池是否预分配内存
func (b *Builder) RenewPoolPreAlloc(preAlloc bool) *Builder {
	if b.renewPoolConfig == nil {
		b.renewPoolConfig = &ants.RenewPoolConfig{}
	}
	b.renewPoolConfig.PreAlloc = preAlloc
	return b
}

// RenewPoolNonBlocking sets pool non blocking switch RenewPoolNonBlocking 设置协程池是否为非阻塞模式
func (b *Builder) RenewPoolNonBlocking(nonBlocking bool) *Builder {
	if b.renewPoolConfig == nil {
		b.renewPoolConfig = &ants.RenewPoolConfig{}
	}
	b.renewPoolConfig.NonBlocking = nonBlocking
	return b
}

// RenewPoolConfig sets full renew pool config RenewPoolConfig 设置完整续期协程池配置
func (b *Builder) RenewPoolConfig(cfg *ants.RenewPoolConfig) *Builder {
	b.renewPoolConfig = cfg
	return b
}

// LoggerPath sets logger path LoggerPath 设置日志文件目录路径
func (b *Builder) LoggerPath(path string) *Builder {
	if b.logConfig == nil {
		b.logConfig = &slog.LoggerConfig{}
	}
	b.logConfig.Path = path
	return b
}

// LoggerFileFormat sets logger file format LoggerFileFormat 设置日志文件命名格式
func (b *Builder) LoggerFileFormat(format string) *Builder {
	if b.logConfig == nil {
		b.logConfig = &slog.LoggerConfig{}
	}
	b.logConfig.FileFormat = format
	return b
}

// LoggerPrefix sets logger prefix LoggerPrefix 设置日志行前缀
func (b *Builder) LoggerPrefix(prefix string) *Builder {
	if b.logConfig == nil {
		b.logConfig = &slog.LoggerConfig{}
	}
	b.logConfig.Prefix = prefix
	return b
}

// LoggerLevel sets logger level LoggerLevel 设置日志最低输出级别
func (b *Builder) LoggerLevel(level slog.LogLevel) *Builder {
	if b.logConfig == nil {
		b.logConfig = &slog.LoggerConfig{}
	}
	b.logConfig.Level = level
	return b
}

// LoggerTimeFormat sets logger time format LoggerTimeFormat 设置日志时间戳格式
func (b *Builder) LoggerTimeFormat(format string) *Builder {
	if b.logConfig == nil {
		b.logConfig = &slog.LoggerConfig{}
	}
	b.logConfig.TimeFormat = format
	return b
}

// LoggerStdout sets stdout switch LoggerStdout 设置是否将日志输出到控制台
func (b *Builder) LoggerStdout(stdout bool) *Builder {
	if b.logConfig == nil {
		b.logConfig = &slog.LoggerConfig{}
	}
	b.logConfig.Stdout = stdout
	return b
}

// LoggerStdoutOnly sets stdout only switch LoggerStdoutOnly 设置是否仅输出到控制台
func (b *Builder) LoggerStdoutOnly(stdoutOnly bool) *Builder {
	if b.logConfig == nil {
		b.logConfig = &slog.LoggerConfig{}
	}
	b.logConfig.StdoutOnly = stdoutOnly
	return b
}

// LoggerQueueSize sets logger queue size LoggerQueueSize 设置日志异步写入队列大小
func (b *Builder) LoggerQueueSize(size int) *Builder {
	if b.logConfig == nil {
		b.logConfig = &slog.LoggerConfig{}
	}
	b.logConfig.QueueSize = size
	return b
}

// LoggerRotateSize sets rotate size LoggerRotateSize 设置日志文件滚动大小阈值
func (b *Builder) LoggerRotateSize(size int64) *Builder {
	if b.logConfig == nil {
		b.logConfig = &slog.LoggerConfig{}
	}
	b.logConfig.RotateSize = size
	return b
}

// LoggerRotateExpire sets rotate expire LoggerRotateExpire 设置日志文件按时间滚动间隔
func (b *Builder) LoggerRotateExpire(expire time.Duration) *Builder {
	if b.logConfig == nil {
		b.logConfig = &slog.LoggerConfig{}
	}
	b.logConfig.RotateExpire = expire
	return b
}

// LoggerRotateBackupLimit sets backup limit LoggerRotateBackupLimit 设置日志备份文件最大数量
func (b *Builder) LoggerRotateBackupLimit(limit int) *Builder {
	if b.logConfig == nil {
		b.logConfig = &slog.LoggerConfig{}
	}
	b.logConfig.RotateBackupLimit = limit
	return b
}

// LoggerRotateBackupDays sets backup retain days LoggerRotateBackupDays 设置日志备份文件保留天数
func (b *Builder) LoggerRotateBackupDays(days int) *Builder {
	if b.logConfig == nil {
		b.logConfig = &slog.LoggerConfig{}
	}
	b.logConfig.RotateBackupDays = days
	return b
}

// LoggerConfig sets full logger config LoggerConfig 设置完整日志配置
func (b *Builder) LoggerConfig(cfg *slog.LoggerConfig) *Builder {
	b.logConfig = cfg
	return b
}

// SetGenerator sets token generator SetGenerator 设置 Token 生成器
func (b *Builder) SetGenerator(generator adapter.Generator) *Builder {
	b.generator = generator
	return b
}

// SetStorage sets storage adapter SetStorage 设置存储适配器
func (b *Builder) SetStorage(storage adapter.Storage) *Builder {
	b.storage = storage
	return b
}

// SetCodec sets codec adapter SetCodec 设置编解码器适配器
func (b *Builder) SetCodec(codec adapter.Codec) *Builder {
	b.codec = codec
	return b
}

// SetLog sets log adapter SetLog 设置日志适配器
func (b *Builder) SetLog(log adapter.Log) *Builder {
	b.log = log
	return b
}

// SetPool sets async task pool SetPool 设置异步任务协程池
func (b *Builder) SetPool(pool adapter.Pool) *Builder {
	b.pool = pool
	return b
}

// SetCustomPermissionListFunc sets permission callback SetCustomPermissionListFunc 设置自定义权限列表获取函数
func (b *Builder) SetCustomPermissionListFunc(f func(loginID, authType string) ([]string, error)) *Builder {
	b.customPermissionListFunc = f
	return b
}

// SetCustomRoleListFunc sets role callback SetCustomRoleListFunc 设置自定义角色列表获取函数
func (b *Builder) SetCustomRoleListFunc(f func(loginID, authType string) ([]string, error)) *Builder {
	b.customRoleListFunc = f
	return b
}

// SetCustomPermissionListExtFunc sets extended permission callback SetCustomPermissionListExtFunc 设置扩展权限列表获取函数
func (b *Builder) SetCustomPermissionListExtFunc(f func(loginID, device, deviceId, authType string) ([]string, error)) *Builder {
	b.customPermissionListExtFunc = f
	return b
}

// SetCustomRoleListExtFunc sets extended role callback SetCustomRoleListExtFunc 设置扩展角色列表获取函数
func (b *Builder) SetCustomRoleListExtFunc(f func(loginID, device, deviceId, authType string) ([]string, error)) *Builder {
	b.customRoleListExtFunc = f
	return b
}

// JwtSecret enables JWT style and sets secret JwtSecret 设置 JWT 模式并指定密钥
func (b *Builder) JwtSecret(secret string) *Builder {
	b.tokenStyle = adapter.TokenStyleJWT
	b.jwtSecretKey = secret
	return b
}

// Clone clones builder with deep copy Clone 深拷贝当前构建器
func (b *Builder) Clone() *Builder {
	clone := *b
	if b.cookieConfig != nil {
		cookieCopy := *b.cookieConfig
		clone.cookieConfig = &cookieCopy
	}
	if b.renewPoolConfig != nil {
		poolCopy := *b.renewPoolConfig
		clone.renewPoolConfig = &poolCopy
	}
	if b.logConfig != nil {
		logCopy := *b.logConfig
		clone.logConfig = &logCopy
	}
	return &clone
}

// Build builds manager and prints banner Build 构建 Manager 实例并打印启动 Banner
func (b *Builder) Build() *manager.Manager {
	if b.cookieConfig == nil {
		b.cookieConfig = config.DefaultCookieConfig()
	}

	cfg := &config.Config{
		AuthType:         b.authType,
		KeyPrefix:        b.keyPrefix,
		TokenName:        b.tokenName,
		Timeout:          b.timeout,
		AutoRenew:        b.autoRenew,
		RenewMaxRefresh:  b.renewMaxRefresh,
		RenewInterval:    b.renewInterval,
		ActiveTimeout:    b.activeTimeout,
		ConcurrencyScope: b.concurrencyScope,
		IsConcurrent:     b.isConcurrent,
		IsShare:          b.isShare,
		MaxLoginCount:    b.maxLoginCount,
		IsReadBody:       b.isReadBody,
		IsReadHeader:     b.isReadHeader,
		IsReadCookie:     b.isReadCookie,
		TokenStyle:       b.tokenStyle,
		JwtSecretKey:     b.jwtSecretKey,
		IsLog:            b.isLog,
		IsPrintBanner:    b.isPrintBanner,
		AsyncEvent:       b.asyncEvent,
		CookieConfig:     b.cookieConfig,
	}

	err := cfg.Validate()
	if err != nil {
		panic("Build Manager Invalid config err: " + err.Error())
	}

	if b.generator == nil {
		b.generator = sgenerator.NewGenerator(b.timeout, b.jwtSecretKey, b.tokenStyle)
	}
	if b.storage == nil {
		b.storage = memory.NewStorage()
	}
	if b.codec == nil {
		b.codec = sjson.NewJSONSerializer()
	}

	if b.isLog {
		if b.log == nil {
			if b.logConfig == nil {
				b.logConfig = slog.DefaultLoggerConfig()
			}
			b.log, err = slog.NewLoggerWithConfig(b.logConfig)
			if err != nil {
				panic("Build Manager Invalid LoggerConfig err: " + err.Error())
			}
		}
	} else {
		b.log = nop.NewNopLogger()
	}

	if b.autoRenew {
		if b.pool == nil {
			if b.renewPoolConfig == nil {
				b.renewPoolConfig = ants.DefaultRenewPoolConfig()
			}
			err = b.renewPoolConfig.Validate()
			if err != nil {
				panic("Build Manager Invalid RenewPoolConfig err: " + err.Error())
			}
			b.pool, err = ants.NewRenewPoolManagerWithConfig(b.renewPoolConfig)
			if err != nil {
				panic("Build Manager NewRenewPoolManagerWithConfig err: " + err.Error())
			}
		}

		if b.renewPoolConfig.PrintStatusInterval > 0 {
			ticker := time.NewTicker(b.renewPoolConfig.PrintStatusInterval)
			go func() {
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						running, capacity, usage := b.pool.Stats()
						b.log.Infof(
							"RenewPool Status: Capacity=%d, Running=%d, Usage=%.2f%%",
							capacity, running, usage*100,
						)
					}
				}
			}()
		}
	}

	if b.isPrintBanner {
		banner.PrintBanner(cfg)
	}

	return manager.NewManager(cfg, b.generator, b.storage, b.codec, b.log, b.pool, b.customPermissionListFunc, b.customRoleListFunc, b.customPermissionListExtFunc, b.customRoleListExtFunc)
}
