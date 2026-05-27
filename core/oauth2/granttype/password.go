package granttype

import (
	"github.com/sa-tokens/sa-token-go/core/adapter"
	"github.com/sa-tokens/sa-token-go/core/errs"
)

// PasswordHandler resource-owner password grant | 密码模式
type PasswordHandler struct{}

func (PasswordHandler) Type() string { return "password" }

// Authorize validates client and issues access token | 校验客户端并颁发 access_token
func (PasswordHandler) Authorize(ctx adapter.RequestContext, deps Deps) (*Result, error) {
	clientID := ctx.GetPostForm("client_id")
	clientSecret := ctx.GetPostForm("client_secret")
	username := ctx.GetPostForm("username")
	password := ctx.GetPostForm("password")
	scope := ctx.GetPostForm("scope")
	if clientID == "" || username == "" || password == "" {
		return nil, errs.ErrOAuth2ParamMissing("client_id/username/password")
	}
	if err := deps.Template.CheckClientCredential(clientID, clientSecret); err != nil {
		return nil, err
	}
	if err := deps.Template.CheckGrantType(clientID, "password"); err != nil {
		return nil, err
	}
	if deps.UserAuth == nil {
		return nil, errs.ErrFeatureNotSupportedNamed("oauth2-password-grant")
	}
	loginID, ok := deps.UserAuth.CheckCredential(username, password)
	if !ok {
		return nil, errs.ErrOAuth2InvalidUserCredential
	}
	scopes := parseScopes(scope)
	if len(scopes) > 0 {
		if err := deps.Template.CheckContractScope(clientID, scopes); err != nil {
			return nil, err
		}
	}
	tok, err := deps.Server.IssueAccessToken(loginID, clientID, scopes)
	if err != nil {
		return nil, err
	}
	return &Result{Data: tok}, nil
}
