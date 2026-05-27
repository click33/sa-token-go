package gozero

import (
	"time"

	"github.com/sa-tokens/sa-token-go/core"
	"github.com/sa-tokens/sa-token-go/stputil"
)

// ============ Re-export core types | 重新导出核心类型 ============

// Configuration related types | 配置相关类型
type (
	Config       = core.Config
	CookieConfig = core.CookieConfig
	TokenStyle   = core.TokenStyle
)

// Token style constants | Token风格常量
const (
	TokenStyleUUID      = core.TokenStyleUUID
	TokenStyleSimple    = core.TokenStyleSimple
	TokenStyleRandom32  = core.TokenStyleRandom32
	TokenStyleRandom64  = core.TokenStyleRandom64
	TokenStyleRandom128 = core.TokenStyleRandom128
	TokenStyleJWT       = core.TokenStyleJWT
	TokenStyleHash      = core.TokenStyleHash
	TokenStyleTimestamp = core.TokenStyleTimestamp
	TokenStyleTik       = core.TokenStyleTik
)

// Core types | 核心类型
type (
	Manager             = core.Manager
	TokenInfo           = core.TokenInfo
	Session             = core.Session
	TokenGenerator      = core.TokenGenerator
	SaTokenContext      = core.SaTokenContext
	Builder             = core.Builder
	NonceManager        = core.NonceManager
	RefreshTokenInfo    = core.RefreshTokenInfo
	RefreshTokenManager = core.RefreshTokenManager
	OAuth2Server        = core.OAuth2Server
	OAuth2Client        = core.OAuth2Client
	OAuth2AccessToken   = core.OAuth2AccessToken
	OAuth2GrantType     = core.OAuth2GrantType
)

// Adapter interfaces | 适配器接口
type (
	Storage        = core.Storage
	RequestContext = core.RequestContext
)

// Event related types | 事件相关类型
type (
	EventListener  = core.EventListener
	EventManager   = core.EventManager
	EventData      = core.EventData
	Event          = core.Event
	ListenerFunc   = core.ListenerFunc
	ListenerConfig = core.ListenerConfig
)

// Event constants | 事件常量
const (
	EventLogin           = core.EventLogin
	EventLogout          = core.EventLogout
	EventKickout         = core.EventKickout
	EventDisable         = core.EventDisable
	EventUntie           = core.EventUntie
	EventRenew           = core.EventRenew
	EventCreateSession   = core.EventCreateSession
	EventDestroySession  = core.EventDestroySession
	EventPermissionCheck = core.EventPermissionCheck
	EventRoleCheck       = core.EventRoleCheck
	EventAll             = core.EventAll
)

// OAuth2 grant type constants | OAuth2授权类型常量
const (
	GrantTypeAuthorizationCode = core.GrantTypeAuthorizationCode
	GrantTypeRefreshToken      = core.GrantTypeRefreshToken
	GrantTypeClientCredentials = core.GrantTypeClientCredentials
	GrantTypePassword          = core.GrantTypePassword
)

// Utility functions | 工具函数
var (
	RandomString   = core.RandomString
	IsEmpty        = core.IsEmpty
	IsNotEmpty     = core.IsNotEmpty
	DefaultString  = core.DefaultString
	ContainsString = core.ContainsString
	RemoveString   = core.RemoveString
	UniqueStrings  = core.UniqueStrings
	MergeStrings   = core.MergeStrings
	MatchPattern   = core.MatchPattern
)

// ============ Core constructor functions | 核心构造函数 ============

func DefaultConfig() *Config                                      { return core.DefaultConfig() }
func NewManager(storage Storage, cfg *Config) *Manager            { return core.NewManager(storage, cfg) }
func NewContext(ctx RequestContext, mgr *Manager) *SaTokenContext { return core.NewContext(ctx, mgr) }
func NewSession(id string, storage Storage, prefix string) *Session {
	return core.NewSession(id, storage, prefix)
}
func LoadSession(id string, storage Storage, prefix string) (*Session, error) {
	return core.LoadSession(id, storage, prefix)
}
func NewTokenGenerator(cfg *Config) *TokenGenerator { return core.NewTokenGenerator(cfg) }
func NewEventManager() *EventManager                { return core.NewEventManager() }
func NewBuilder() *Builder                          { return core.NewBuilder() }
func NewNonceManager(storage Storage, prefix string, ttl ...int64) *NonceManager {
	return core.NewNonceManager(storage, prefix, ttl...)
}
func NewRefreshTokenManager(storage Storage, prefix string, cfg *Config) *RefreshTokenManager {
	return core.NewRefreshTokenManager(storage, prefix, cfg)
}
func NewOAuth2Server(storage Storage, prefix string) *OAuth2Server {
	return core.NewOAuth2Server(storage, prefix)
}

// ============ Global StpUtil functions | 全局StpUtil函数 ============

func SetManager(mgr *Manager) { stputil.SetManager(mgr) }
func GetManager() *Manager    { return stputil.GetManager() }
func Login(loginID interface{}, device ...string) (string, error) {
	return stputil.Login(loginID, device...)
}
func LoginByToken(loginID interface{}, tokenValue string, device ...string) error {
	return stputil.LoginByToken(loginID, tokenValue, device...)
}
func Logout(loginID interface{}, device ...string) error { return stputil.Logout(loginID, device...) }
func LogoutByToken(tokenValue string) error              { return stputil.LogoutByToken(tokenValue) }
func IsLogin(tokenValue string) bool                     { return stputil.IsLogin(tokenValue) }
func CheckLogin(tokenValue string) error                 { return stputil.CheckLogin(tokenValue) }
func GetLoginID(tokenValue string) (string, error)       { return stputil.GetLoginID(tokenValue) }
func GetLoginIDNotCheck(tokenValue string) (string, error) {
	return stputil.GetLoginIDNotCheck(tokenValue)
}
func GetTokenValue(loginID interface{}, device ...string) (string, error) {
	return stputil.GetTokenValue(loginID, device...)
}
func GetTokenInfo(tokenValue string) (*TokenInfo, error)  { return stputil.GetTokenInfo(tokenValue) }
func Kickout(loginID interface{}, device ...string) error { return stputil.Kickout(loginID, device...) }
func Disable(loginID interface{}, duration time.Duration) error {
	return stputil.Disable(loginID, duration)
}
func IsDisable(loginID interface{}) bool                { return stputil.IsDisable(loginID) }
func CheckDisableByToken(tokenValue string) error       { return stputil.CheckDisable(tokenValue) }
func GetDisableTime(loginID interface{}) (int64, error) { return stputil.GetDisableTime(loginID) }
func Untie(loginID interface{}) error                   { return stputil.Untie(loginID) }
func CheckPermissionByToken(tokenValue string, permission string) error {
	return stputil.CheckPermission(tokenValue, permission)
}
func HasPermission(loginID interface{}, permission string) bool {
	return stputil.HasPermission(loginID, permission)
}
func CheckPermissionAndByToken(tokenValue string, permissions []string) error {
	return stputil.CheckPermissionAnd(tokenValue, permissions)
}
func CheckPermissionOrByToken(tokenValue string, permissions []string) error {
	return stputil.CheckPermissionOr(tokenValue, permissions)
}
func GetPermissionListByToken(tokenValue string) ([]string, error) {
	return stputil.GetPermissionList(tokenValue)
}
func CheckRoleByToken(tokenValue string, role string) error {
	return stputil.CheckRole(tokenValue, role)
}
func HasRole(loginID interface{}, role string) bool { return stputil.HasRole(loginID, role) }
func CheckRoleAndByToken(tokenValue string, roles []string) error {
	return stputil.CheckRoleAnd(tokenValue, roles)
}
func CheckRoleOrByToken(tokenValue string, roles []string) error {
	return stputil.CheckRoleOr(tokenValue, roles)
}
func GetRoleListByToken(tokenValue string) ([]string, error) { return stputil.GetRoleList(tokenValue) }
func GetSession(loginID interface{}) (*Session, error)       { return stputil.GetSession(loginID) }
func GetSessionByToken(tokenValue string) (*Session, error) {
	return stputil.GetSessionByToken(tokenValue)
}
func GetTokenSession(tokenValue string) (*Session, error) { return stputil.GetTokenSession(tokenValue) }
func GenerateNonce() (string, error)                      { return stputil.GenerateNonce() }
func VerifyNonce(nonce string) bool                       { return stputil.VerifyNonce(nonce) }
func LoginWithRefreshToken(loginID interface{}, device ...string) (*RefreshTokenInfo, error) {
	return stputil.LoginWithRefreshToken(loginID, device...)
}
func RefreshAccessToken(refreshToken string) (*RefreshTokenInfo, error) {
	return stputil.RefreshAccessToken(refreshToken)
}
func RevokeRefreshToken(refreshToken string) error { return stputil.RevokeRefreshToken(refreshToken) }
func GetOAuth2Server() *OAuth2Server               { return stputil.GetOAuth2Server() }

const Version = core.Version
