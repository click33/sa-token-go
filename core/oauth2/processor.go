package oauth2

import (
	"strings"

	"github.com/sa-tokens/sa-token-go/core/adapter"
	"github.com/sa-tokens/sa-token-go/core/config"
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

// ManagerLike Processor 所需 Manager 能力
// 增加 GetConfig，以便在本包内联与 ReadTokenFromRequest 等价的读取逻辑
// 注意：不得 import core/context（manager → oauth2 → context → manager 会形成循环依赖）
type ManagerLike interface {
	GetLoginID(token string) (string, error)
	IsLogin(token string) bool
	Login(loginID string, device ...string) (string, error)
	CutTokenPrefix(raw string) string
	GetTokenName() string
	GetConfig() *config.Config
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

// readToken 从请求中读取主会话 Token
// 与 core/context.ReadTokenFromRequest 语义对齐（Header → Cookie → Query + Bearer 兜底 + CutTokenPrefix）
// 因 oauth2 被 manager 引用，不可 import context，故在此内联同等逻辑
func (p *OAuth2ServerProcessor) readToken(ctx adapter.RequestContext) string {
	if p.manager == nil || ctx == nil {
		return ""
	}
	mgr := p.manager
	cfg := mgr.GetConfig()
	// 与 context.ResolveTokenName 一致：有 TokenName 用配置，否则回退 Authorization
	name := "Authorization"
	if cfg != nil && strings.TrimSpace(cfg.TokenName) != "" {
		name = cfg.TokenName
	}

	readHeader := cfg == nil || cfg.IsReadHeader
	readCookie := cfg == nil || cfg.IsReadCookie

	// 1) Header
	if readHeader {
		if v := strings.TrimSpace(ctx.GetHeader(name)); v != "" {
			if strings.EqualFold(name, "Authorization") {
				if t := extractBearerToken(v); t != "" {
					return mgr.CutTokenPrefix(t)
				}
			}
			return mgr.CutTokenPrefix(v)
		}
		if !strings.EqualFold(name, "Authorization") {
			if auth := strings.TrimSpace(ctx.GetHeader("Authorization")); auth != "" {
				if t := extractBearerToken(auth); t != "" {
					return mgr.CutTokenPrefix(t)
				}
			}
		}
	}

	// 2) Cookie：开启 CookieAutoFillPrefix 时先拼 TokenPrefix 再裁剪
	if readCookie {
		if v := strings.TrimSpace(ctx.GetCookie(name)); v != "" {
			if cfg != nil && cfg.CookieAutoFillPrefix && cfg.TokenPrefix != "" {
				v = cfg.TokenPrefix + v
			}
			return mgr.CutTokenPrefix(v)
		}
	}

	// 3) Query
	if v := strings.TrimSpace(ctx.GetQuery(name)); v != "" {
		return mgr.CutTokenPrefix(v)
	}
	return ""
}

// extractBearerToken 从 Authorization 头提取 Bearer token（大小写不敏感）
func extractBearerToken(auth string) string {
	auth = strings.TrimSpace(auth)
	if auth == "" {
		return ""
	}
	const bearerPrefix = "Bearer "
	if len(auth) > 7 && strings.EqualFold(auth[:7], bearerPrefix) {
		return strings.TrimSpace(auth[7:])
	}
	return auth
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

// revoke 吊销 access_token
// 表单里的 token 也可能带 TokenPrefix，统一裁剪后再交给 RevokeToken
func (p *OAuth2ServerProcessor) revoke(ctx adapter.RequestContext) (any, error) {
	tok := ctx.GetPostForm("access_token")
	if tok == "" {
		tok = ctx.GetPostForm("token")
	}
	if tok == "" {
		return nil, errs.ErrOAuth2ParamMissing("access_token")
	}
	// 客户端可能原样回传带前缀的 access_token
	if p.manager != nil {
		tok = p.manager.CutTokenPrefix(tok)
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
