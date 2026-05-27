package server

import (
	"net/url"

	core "github.com/sa-tokens/sa-token-go/core"
	"github.com/sa-tokens/sa-token-go/core/adapter"
)

const (
	pathSsoAuth        = "/sso/auth"
	pathSsoDoLogin     = "/sso/doLogin"
	pathSsoCheckTicket = "/sso/checkTicket"
	pathSsoSignout     = "/sso/signout"
)

// Processor HTTP entry for SSO server | SSO 服务端 HTTP 入口
type Processor struct {
	tpl *Template
}

// NewProcessor constructs processor | 构造处理器
func NewProcessor(tpl *Template) *Processor {
	return &Processor{tpl: tpl}
}

// Dispatch routes by path | 路由分发
func (p *Processor) Dispatch(ctx adapter.RequestContext) (bool, any, error) {
	if ctx == nil {
		return false, nil, nil
	}
	switch ctx.GetPath() {
	case pathSsoAuth:
		v, err := p.auth(ctx)
		return true, v, err
	case pathSsoCheckTicket:
		v, err := p.checkTicket(ctx)
		return true, v, err
	case pathSsoSignout:
		v, err := p.signout(ctx)
		return true, v, err
	case pathSsoDoLogin:
		v, err := p.doLogin(ctx)
		return true, v, err
	default:
		return false, nil, nil
	}
}

func (p *Processor) readToken(ctx adapter.RequestContext) string {
	m := p.tpl.mgr
	return m.CutTokenPrefix(ctx.GetHeader(m.GetTokenName()))
}

func (p *Processor) auth(ctx adapter.RequestContext) (any, error) {
	redirect := ctx.GetQuery("redirect")
	clientID := ctx.GetQuery("client")
	if redirect == "" {
		return nil, core.NewOAuth2RequiredParamError("redirect")
	}
	if err := p.tpl.CheckRedirectURL(redirect); err != nil {
		return nil, err
	}
	tokenValue := p.readToken(ctx)
	loginID, err := p.tpl.mgr.GetLoginID(tokenValue)
	if err != nil {
		return map[string]any{"need_login": true, "redirect": redirect}, nil
	}
	ticket, err := p.tpl.CreateTicketAndSave(clientID, loginID, tokenValue)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(redirect)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("ticket", ticket)
	u.RawQuery = q.Encode()
	return map[string]any{"redirect": u.String()}, nil
}

func (p *Processor) checkTicket(ctx adapter.RequestContext) (any, error) {
	ticket := ctx.GetPostForm("ticket")
	client := ctx.GetPostForm("client")
	sloCallback := ctx.GetPostForm("ssoLogoutCall")
	if ticket == "" {
		return nil, core.NewOAuth2RequiredParamError("ticket")
	}
	tm, err := p.tpl.CheckTicketAndDelete(ticket, client)
	if err != nil {
		return nil, err
	}
	if sloCallback != "" {
		_ = p.tpl.RegisterSloCallback(tm.LoginID, tm.Client, sloCallback)
	}
	return map[string]any{"loginId": tm.LoginID}, nil
}

func (p *Processor) signout(ctx adapter.RequestContext) (any, error) {
	loginID := ctx.GetPostForm("loginId")
	if loginID == "" {
		return nil, core.NewOAuth2RequiredParamError("loginId")
	}
	if err := p.tpl.SsoLogout(loginID); err != nil {
		return nil, err
	}
	return map[string]any{"signout": true}, nil
}

func (p *Processor) doLogin(ctx adapter.RequestContext) (any, error) {
	loginID := ctx.GetPostForm("name")
	if loginID == "" {
		return nil, core.NewOAuth2RequiredParamError("name")
	}
	tok, err := p.tpl.mgr.Login(loginID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"token": tok, "loginId": loginID}, nil
}
