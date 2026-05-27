package granttype

import (
	"github.com/sa-tokens/sa-token-go/core/adapter"
	"github.com/sa-tokens/sa-token-go/core/errs"
)

// RefreshTokenHandler refresh_token grant | 刷新令牌模式
type RefreshTokenHandler struct{}

func (RefreshTokenHandler) Type() string { return "refresh_token" }

// Authorize refreshes access token | 刷新 access_token
func (RefreshTokenHandler) Authorize(ctx adapter.RequestContext, deps Deps) (*Result, error) {
	clientID := ctx.GetPostForm("client_id")
	clientSecret := ctx.GetPostForm("client_secret")
	refreshToken := ctx.GetPostForm("refresh_token")
	if refreshToken == "" {
		return nil, errs.ErrOAuth2ParamMissing("refresh_token")
	}
	if clientID != "" {
		if err := deps.Template.CheckClientCredential(clientID, clientSecret); err != nil {
			return nil, err
		}
		if err := deps.Template.CheckGrantType(clientID, "refresh_token"); err != nil {
			return nil, err
		}
	}
	tok, err := deps.Server.IssueRefreshToken(refreshToken, clientID, clientSecret)
	if err != nil {
		return nil, err
	}
	return &Result{Data: tok}, nil
}
