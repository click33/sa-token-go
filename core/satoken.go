package core

import (
	"time"

	"github.com/click33/sa-token-go/core/adapter"
	"github.com/click33/sa-token-go/core/builder"
	"github.com/click33/sa-token-go/core/config"
	"github.com/click33/sa-token-go/core/context"
	"github.com/click33/sa-token-go/core/listener"
	"github.com/click33/sa-token-go/core/manager"
	"github.com/click33/sa-token-go/core/oauth2"
	"github.com/click33/sa-token-go/core/security"
	"github.com/click33/sa-token-go/core/session"
	"github.com/click33/sa-token-go/core/token"
	"github.com/click33/sa-token-go/core/utils"
)

// Version Sa-Token-Go version | Sa-Token-Go版本
const Version = "0.1.0"

// Export main types and functions for external use | 导出主要类型和函数，方便外部使用

// Configuration related types | 配置相关类型
type (
	Config       = config.Config
	CookieConfig = config.CookieConfig
	TokenStyle   = config.TokenStyle
)

// Token style constants | Token风格常量
const (
	TokenStyleUUID      = config.TokenStyleUUID
	TokenStyleSimple    = config.TokenStyleSimple
	TokenStyleRandom32  = config.TokenStyleRandom32
	TokenStyleRandom64  = config.TokenStyleRandom64
	TokenStyleRandom128 = config.TokenStyleRandom128
	TokenStyleJWT       = config.TokenStyleJWT
	TokenStyleHash      = config.TokenStyleHash
	TokenStyleTimestamp = config.TokenStyleTimestamp
	TokenStyleTik       = config.TokenStyleTik
)

// Core types | 核心类型
type (
	Manager             = manager.Manager
	TokenInfo           = manager.TokenInfo
	Session             = session.Session
	TokenGenerator      = token.Generator
	SaTokenContext      = context.SaTokenContext
	Builder             = builder.Builder
	NonceManager        = security.NonceManager
	RefreshTokenInfo    = security.RefreshTokenInfo
	RefreshTokenManager = security.RefreshTokenManager
	OAuth2Server        = oauth2.OAuth2Server
	OAuth2Client        = oauth2.Client
	OAuth2AccessToken   = oauth2.AccessToken
	OAuth2GrantType     = oauth2.GrantType
)

// Adapter interfaces | 适配器接口
type (
	Storage        = adapter.Storage
	RequestContext = adapter.RequestContext
)

// Event related types | 事件相关类型
type (
	EventListener  = listener.Listener
	EventManager   = listener.Manager
	EventData      = listener.EventData
	Event          = listener.Event
	ListenerFunc   = listener.ListenerFunc
	ListenerConfig = listener.ListenerConfig
)

// Event constants | 事件常量
const (
	EventLogin           = listener.EventLogin
	EventLogout          = listener.EventLogout
	EventKickout         = listener.EventKickout
	EventDisable         = listener.EventDisable
	EventUntie           = listener.EventUntie
	EventRenew           = listener.EventRenew
	EventCreateSession   = listener.EventCreateSession
	EventDestroySession  = listener.EventDestroySession
	EventPermissionCheck = listener.EventPermissionCheck
	EventRoleCheck       = listener.EventRoleCheck
	EventAll             = listener.EventAll
)

const (
	GrantTypeAuthorizationCode = oauth2.GrantTypeAuthorizationCode
	GrantTypeRefreshToken      = oauth2.GrantTypeRefreshToken
	GrantTypeClientCredentials = oauth2.GrantTypeClientCredentials
	GrantTypePassword          = oauth2.GrantTypePassword
)

// Utility functions | 工具函数
var (
	RandomString   = utils.RandomString
	IsEmpty        = utils.IsEmpty
	IsNotEmpty     = utils.IsNotEmpty
	DefaultString  = utils.DefaultString
	ContainsString = utils.ContainsString
	RemoveString   = utils.RemoveString
	UniqueStrings  = utils.UniqueStrings
	MergeStrings   = utils.MergeStrings
	MatchPattern   = utils.MatchPattern
)

// DefaultConfig returns default configuration | 返回默认配置
func DefaultConfig() *Config {
	return config.DefaultConfig()
}

// NewManager creates a new authentication manager | 创建新的认证管理器
func NewManager(storage Storage, cfg *Config) *Manager {
	return manager.NewManager(storage, cfg)
}

// NewContext creates a new Sa-Token context | 创建新的Sa-Token上下文
func NewContext(ctx RequestContext, mgr *Manager) *SaTokenContext {
	return context.NewContext(ctx, mgr)
}

// NewSession creates a new session | 创建新的Session
func NewSession(id string, storage Storage, prefix string) *Session {
	return session.NewSession(id, storage, prefix)
}

// LoadSession loads an existing session | 加载已存在的Session
func LoadSession(id string, storage Storage, prefix string) (*Session, error) {
	return session.Load(id, storage, prefix)
}

// NewTokenGenerator creates a new token generator | 创建新的Token生成器
func NewTokenGenerator(cfg *Config) *TokenGenerator {
	return token.NewGenerator(cfg)
}

// NewEventManager creates a new event manager | 创建新的事件管理器
func NewEventManager() *EventManager {
	return listener.NewManager()
}

// NewBuilder creates a new builder for fluent configuration | 创建新的Builder构建器（用于流式配置）
func NewBuilder() *Builder {
	return builder.NewBuilder()
}

func NewNonceManager(storage Storage, ttl ...int64) *NonceManager {
	var duration time.Duration
	if len(ttl) > 0 && ttl[0] > 0 {
		duration = time.Duration(ttl[0]) * time.Second
	}
	return security.NewNonceManager(storage, duration)
}

func NewRefreshTokenManager(storage Storage, cfg *Config) *RefreshTokenManager {
	return security.NewRefreshTokenManager(storage, cfg)
}

func NewOAuth2Server(storage Storage) *OAuth2Server {
	return oauth2.NewOAuth2Server(storage)
}
