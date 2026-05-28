package granttype

import (
	"github.com/sa-tokens/sa-token-go/core/adapter"
	"github.com/sa-tokens/sa-token-go/core/errs"
)

// AuthorizationCodeHandler authorization_code grant | 授权码模式
type AuthorizationCodeHandler struct{}

func (AuthorizationCodeHandler) Type() string { return "authorization_code" }

// Authorize exchanges code for tokens | 用 code 换 token
func (AuthorizationCodeHandler) Authorize(ctx adapter.RequestContext, deps Deps) (*Result, error) {
	clientID := ctx.GetPostForm("client_id")
	clientSecret := ctx.GetPostForm("client_secret")
	code := ctx.GetPostForm("code")
	redirectURI := ctx.GetPostForm("redirect_uri")
	if code == "" {
		return nil, errs.ErrOAuth2ParamMissing("code")
	}
	if clientID == "" {
		return nil, errs.ErrOAuth2ParamMissing("client_id")
	}
	if err := deps.Template.CheckClientCredential(clientID, clientSecret); err != nil {
		return nil, err
	}
	if err := deps.Template.CheckGrantType(clientID, "authorization_code"); err != nil {
		return nil, err
	}
	tok, err := deps.Server.ConsumeAuthorizationCode(code, clientID, clientSecret, redirectURI)
	if err != nil {
		return nil, err
	}
	return &Result{Data: tok}, nil
}
