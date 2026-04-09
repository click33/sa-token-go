package oauth2

import (
	"time"
)

const (
	// DefaultCodeExpiration stores code expiration DefaultCodeExpiration 存储授权码过期时间
	DefaultCodeExpiration = 10 * time.Minute
	// DefaultTokenExpiration stores token expiration DefaultTokenExpiration 存储访问令牌过期时间
	DefaultTokenExpiration = 2 * time.Hour
	// DefaultRefreshTTL stores refresh expiration DefaultRefreshTTL 存储刷新令牌过期时间
	DefaultRefreshTTL = 30 * 24 * time.Hour

	// CodeLength stores code byte length CodeLength 存储授权码字节长度
	CodeLength = 32
	// AccessTokenLength stores access token byte length AccessTokenLength 存储访问令牌字节长度
	AccessTokenLength = 32
	// RefreshTokenLength stores refresh token byte length RefreshTokenLength 存储刷新令牌字节长度
	RefreshTokenLength = 32

	// CodeKeySuffix stores code key suffix CodeKeySuffix 存储授权码键后缀
	CodeKeySuffix = "oauth2:code:"
	// TokenKeySuffix stores token key suffix TokenKeySuffix 存储令牌键后缀
	TokenKeySuffix = "oauth2:token:"
	// RefreshKeySuffix stores refresh key suffix RefreshKeySuffix 存储刷新令牌键后缀
	RefreshKeySuffix = "oauth2:refresh:"

	// TokenTypeBearer stores bearer token type TokenTypeBearer 存储 Bearer 令牌类型
	TokenTypeBearer = "Bearer"
)

// GrantType defines oauth2 grant type GrantType 定义 OAuth2 授权类型
type GrantType string

const (
	// GrantTypeAuthorizationCode stores authorization code mode GrantTypeAuthorizationCode 存储授权码模式
	GrantTypeAuthorizationCode GrantType = "authorization_code"
	// GrantTypeRefreshToken stores refresh token mode GrantTypeRefreshToken 存储刷新令牌模式
	GrantTypeRefreshToken GrantType = "refresh_token"
	// GrantTypeClientCredentials stores client credentials mode GrantTypeClientCredentials 存储客户端凭证模式
	GrantTypeClientCredentials GrantType = "client_credentials"
	// GrantTypePassword stores password mode GrantTypePassword 存储密码模式
	GrantTypePassword GrantType = "password"
)
