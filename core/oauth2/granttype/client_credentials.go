package granttype

import (
	"github.com/sa-tokens/sa-token-go/core/adapter"
	"github.com/sa-tokens/sa-token-go/core/errs"
)

// ClientCredentialsHandler client_credentials grant | 客户端模式
type ClientCredentialsHandler struct{}

func (ClientCredentialsHandler) Type() string { return "client_credentials" }

// Authorize issues token for the client itself | 以 client 为主体颁发 token
func (ClientCredentialsHandler) Authorize(ctx adapter.RequestContext, deps Deps) (*Result, error) {
	clientID := ctx.GetPostForm("client_id")
	clientSecret := ctx.GetPostForm("client_secret")
	scope := ctx.GetPostForm("scope")
	if clientID == "" {
		return nil, errs.ErrOAuth2ParamMissing("client_id")
	}
	if err := deps.Template.CheckClientCredential(clientID, clientSecret); err != nil {
		return nil, err
	}
	if err := deps.Template.CheckGrantType(clientID, "client_credentials"); err != nil {
		return nil, err
	}
	scopes := parseScopes(scope)
	if len(scopes) > 0 {
		if err := deps.Template.CheckContractScope(clientID, scopes); err != nil {
			return nil, err
		}
	}
	tok, err := deps.Server.IssueAccessToken(clientID, clientID, scopes)
	if err != nil {
		return nil, err
	}
	return &Result{Data: tok}, nil
}
