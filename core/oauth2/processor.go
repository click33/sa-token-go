package oauth2

import (
	"github.com/sa-tokens/sa-token-go/core/adapter"
	"github.com/sa-tokens/sa-token-go/core/errs"
	"github.com/sa-tokens/sa-token-go/core/oauth2/granttype"
)

// APIPaths configurable OAuth2 HTTP paths | OAuth2 路由路径
type APIPaths struct {
	Authorize   string
	Token       string
	Refresh     string
	Revoke      string
	DoLogin     string
	DoConfirm   string
	ClientToken string
}

// ManagerLike minimal manager surface for OAuth2 processor | Processor 所需 Manager 能力
type ManagerLike interface {
	GetLoginID(token string) (string, error)
	IsLogin(token string) bool
	Login(loginID string, device ...string) (string, error)
	CutTokenPrefix(raw string) string
	GetTokenName() string
}

// OAuth2ServerProcessor routes OAuth2 HTTP endpoints | OAuth2 路由处理器
type OAuth2ServerProcessor struct {
	template *OAuth2Template
	server   *OAuth2Server
	manager  ManagerLike
	paths    APIPaths
	grants   *granttype.Registry
	userAuth UserCredentialChecker
}

// NewOAuth2ServerProcessor constructs processor | 构造处理器
func NewOAuth2ServerProcessor(tpl *OAuth2Template, srv *OAuth2Server, mgr ManagerLike, paths APIPaths, grants *granttype.Registry, userAuth UserCredentialChecker) *OAuth2ServerProcessor {
	if grants == nil {
		grants = granttype.NewRegistry()
		grants.Register(granttype.AuthorizationCodeHandler{})
		grants.Register(granttype.PasswordHandler{})
		grants.Register(granttype.ClientCredentialsHandler{})
		grants.Register(granttype.RefreshTokenHandler{})
	}
	return &OAuth2ServerProcessor{
		template: tpl,
		server:   srv,
		manager:  mgr,
		paths:    paths,
		grants:   grants,
		userAuth: userAuth,
	}
}

// Dispatch handles a request; returns handled, payload, err | 分发请求
func (p *OAuth2ServerProcessor) Dispatch(ctx adapter.RequestContext) (bool, any, error) {
	if ctx == nil {
		return false, nil, nil
	}
	path := ctx.GetPath()
	switch path {
	case p.paths.Authorize:
		v, err := p.authorize(ctx)
		return true, v, err
	case p.paths.Token:
		v, err := p.token(ctx)
		return true, v, err
	case p.paths.Refresh:
		v, err := p.refresh(ctx)
		return true, v, err
	case p.paths.Revoke:
		v, err := p.revoke(ctx)
		return true, v, err
	case p.paths.DoLogin:
		v, err := p.doLogin(ctx)
		return true, v, err
	case p.paths.DoConfirm:
		v, err := p.doConfirm(ctx)
		return true, v, err
	case p.paths.ClientToken:
		v, err := p.clientToken(ctx)
		return true, v, err
	}
	return false, nil, nil
}

func (p *OAuth2ServerProcessor) readToken(ctx adapter.RequestContext) string {
	if p.manager == nil {
		return ""
	}
	name := p.manager.GetTokenName()
	return p.manager.CutTokenPrefix(ctx.GetHeader(name))
}

func (p *OAuth2ServerProcessor) authorize(ctx adapter.RequestContext) (any, error) {
	respType := firstNonEmptyParam(ctx, "response_type")
	clientID := firstNonEmptyParam(ctx, "client_id")
	redirectURI := firstNonEmptyParam(ctx, "redirect_uri")
	rawScope := firstNonEmptyParam(ctx, "scope")
	state := firstNonEmptyParam(ctx, "state")
	if respType != "code" && respType != "token" {
		return nil, errs.ErrOAuth2InvalidResponseType
	}
	if clientID == "" {
		return nil, errs.ErrOAuth2ParamMissing("client_id")
	}
	if err := p.template.CheckRedirectURI(clientID, redirectURI); err != nil {
		return nil, err
	}
	scopes := ParseScopes(rawScope)
	if len(scopes) > 0 {
		if err := p.template.CheckContractScope(clientID, scopes); err != nil {
			return nil, err
		}
	}
	if p.manager == nil {
		return nil, errs.ErrNotLogin
	}
	tokenValue := p.readToken(ctx)
	loginID, err := p.manager.GetLoginID(tokenValue)
	if err != nil {
		return nil, errs.ErrNotLogin
	}
	if p.template.IsNeedCarefulConfirm(loginID, clientID, scopes) {
		return map[string]any{
			"need_confirm": true,
			"client_id":    clientID,
			"scope":        scopes,
			"redirect_uri": redirectURI,
			"state":        state,
		}, nil
	}
	if err := p.template.SaveGrantScope(loginID, clientID, scopes); err != nil {
		return nil, err
	}
	switch respType {
	case "code":
		ac, err := p.server.GenerateAuthorizationCode(clientID, redirectURI, loginID, scopes)
		if err != nil {
			return nil, err
		}
		return map[string]any{"redirect": redirectURI, "code": ac.Code, "state": state}, nil
	default:
		tk, err := p.server.IssueAccessToken(loginID, clientID, scopes)
		if err != nil {
			return nil, err
		}
		return tk, nil
	}
}

func (p *OAuth2ServerProcessor) token(ctx adapter.RequestContext) (any, error) {
	grant := ctx.GetPostForm("grant_type")
	if grant == "" {
		return nil, errs.ErrOAuth2ParamMissing("grant_type")
	}
	h := p.grants.Get(grant)
	if h == nil {
		return nil, errs.ErrOAuth2InvalidGrantType
	}
	deps := granttype.Deps{
		Server:   p.server,
		Template: p.template,
		UserAuth: p.userAuth,
	}
	r, err := h.Authorize(ctx, deps)
	if err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (p *OAuth2ServerProcessor) refresh(ctx adapter.RequestContext) (any, error) {
	h := p.grants.Get("refresh_token")
	if h == nil {
		return nil, errs.ErrOAuth2InvalidGrantType
	}
	r, err := h.Authorize(ctx, granttype.Deps{Server: p.server, Template: p.template})
	if err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (p *OAuth2ServerProcessor) revoke(ctx adapter.RequestContext) (any, error) {
	tok := ctx.GetPostForm("access_token")
	if tok == "" {
		tok = ctx.GetPostForm("token")
	}
	if tok == "" {
		return nil, errs.ErrOAuth2ParamMissing("access_token")
	}
	if err := p.server.RevokeToken(tok); err != nil {
		return nil, err
	}
	return map[string]any{"revoked": true}, nil
}

func (p *OAuth2ServerProcessor) doConfirm(ctx adapter.RequestContext) (any, error) {
	clientID := firstNonEmptyParam(ctx, "client_id")
	rawScope := firstNonEmptyParam(ctx, "scope")
	if p.manager == nil {
		return nil, errs.ErrNotLogin
	}
	tokenValue := p.readToken(ctx)
	loginID, err := p.manager.GetLoginID(tokenValue)
	if err != nil {
		return nil, errs.ErrNotLogin
	}
	if err := p.template.SaveGrantScope(loginID, clientID, ParseScopes(rawScope)); err != nil {
		return nil, err
	}
	return map[string]any{"granted": true}, nil
}

func (p *OAuth2ServerProcessor) doLogin(ctx adapter.RequestContext) (any, error) {
	username := ctx.GetPostForm("username")
	password := ctx.GetPostForm("password")
	if username == "" || password == "" {
		return nil, errs.ErrOAuth2ParamMissing("username/password")
	}
	if p.userAuth == nil {
		return nil, errs.ErrFeatureNotSupportedNamed("oauth2-doLogin")
	}
	if p.manager == nil {
		return nil, errs.ErrNotLogin
	}
	loginID, ok := p.userAuth.CheckCredential(username, password)
	if !ok {
		return nil, errs.ErrOAuth2InvalidUserCredential
	}
	tok, err := p.manager.Login(loginID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"token": tok, "loginId": loginID}, nil
}

func (p *OAuth2ServerProcessor) clientToken(ctx adapter.RequestContext) (any, error) {
	h := p.grants.Get("client_credentials")
	if h == nil {
		return nil, errs.ErrOAuth2InvalidGrantType
	}
	r, err := h.Authorize(ctx, granttype.Deps{Server: p.server, Template: p.template})
	if err != nil {
		return nil, err
	}
	return r.Data, nil
}
