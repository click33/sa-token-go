package granttype

import "github.com/sa-tokens/sa-token-go/core/adapter"

// Result is a generic grant handler output | 授权处理结果
type Result struct {
	Data any
}

// ServerLike abstracts OAuth2Server token operations | 抽象 OAuth2Server 颁发能力（避免 granttype import oauth2）
type ServerLike interface {
	IssueAccessToken(userID, clientID string, scopes []string) (any, error)
	IssueRefreshToken(refreshToken, clientID, clientSecret string) (any, error)
	ConsumeAuthorizationCode(code, clientID, clientSecret, redirectURI string) (any, error)
}

// TemplateLike abstracts OAuth2Template checks | 抽象模板校验
type TemplateLike interface {
	CheckClientCredential(clientID, clientSecret string) error
	CheckGrantType(clientID, grantType string) error
	CheckContractScope(clientID string, scopes []string) error
}

// Deps runtime dependencies for handlers | Handler 运行时依赖
type Deps struct {
	Server   ServerLike
	Template TemplateLike
	UserAuth interface {
		CheckCredential(username, password string) (loginID string, ok bool)
	}
}

// Handler implements one OAuth2 grant_type | grant_type 处理器
type Handler interface {
	Type() string
	Authorize(ctx adapter.RequestContext, deps Deps) (*Result, error)
}
