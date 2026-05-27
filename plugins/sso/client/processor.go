package client

import (
	"net/http"
	"net/url"
	"strings"

	core "github.com/sa-tokens/sa-token-go/core"
	"github.com/sa-tokens/sa-token-go/core/adapter"
)

const (
	pathClientLogin      = "/sso/login"
	pathClientLogout     = "/sso/logout"
	pathClientLogoutCall = "/sso/logoutCall"
)

// Processor HTTP entry for SSO client | SSO 客户端 HTTP 入口
type Processor struct {
	tpl *Template
}

// NewProcessor constructs client processor | 构造处理器
func NewProcessor(tpl *Template) *Processor {
	return &Processor{tpl: tpl}
}

// Dispatch routes client-side SSO paths | 路由分发
func (p *Processor) Dispatch(ctx adapter.RequestContext) (bool, any, error) {
	if ctx == nil {
		return false, nil, nil
	}
	switch ctx.GetPath() {
	case pathClientLogin:
		v, err := p.login(ctx)
		return true, v, err
	case pathClientLogout:
		v, err := p.logout(ctx)
		return true, v, err
	case pathClientLogoutCall:
		v, err := p.logoutCall(ctx)
		return true, v, err
	default:
		return false, nil, nil
	}
}

func (p *Processor) login(ctx adapter.RequestContext) (any, error) {
	ticket := ctx.GetQuery("ticket")
	back := ctx.GetQuery("back")
	if ticket == "" {
		return map[string]any{
			"redirect": p.tpl.BuildServerAuthURL(ctx.GetURL(), back),
		}, nil
	}
	loginID, err := p.tpl.CheckTicket(ticket, "")
	if err != nil {
		return nil, err
	}
	return map[string]any{"loginId": loginID, "back": back}, nil
}

func (p *Processor) logout(ctx adapter.RequestContext) (any, error) {
	loginID := ctx.GetQuery("loginId")
	if loginID == "" {
		return nil, core.NewOAuth2RequiredParamError("loginId")
	}
	if p.tpl.mgr != nil {
		_ = p.tpl.mgr.Logout(loginID)
	}
	if p.tpl.cfg.ServerURL != "" {
		form := url.Values{}
		form.Set("loginId", loginID)
		if p.tpl.cfg.SecretKey != "" {
			form.Set("sign", p.tpl.signer.Sign(map[string]string{"loginId": loginID}))
		}
		endpoint := strings.TrimRight(p.tpl.cfg.ServerURL, "/") + "/sso/signout"
		resp, err := http.PostForm(endpoint, form)
		if err != nil {
			return nil, core.ErrSsoServerUnreachable
		}
		_ = resp.Body.Close()
	}
	return map[string]any{"signout": true, "loginId": loginID}, nil
}

func (p *Processor) logoutCall(ctx adapter.RequestContext) (any, error) {
	loginID := ctx.GetPostForm("loginId")
	client := ctx.GetPostForm("client")
	if loginID == "" {
		return nil, core.NewOAuth2RequiredParamError("loginId")
	}
	if p.tpl.cfg.SecretKey != "" {
		want := p.tpl.signer.Sign(map[string]string{
			"loginId": loginID,
			"client":  client,
			"evict":   ctx.GetPostForm("evict"),
		})
		if want != ctx.GetPostForm("sign") {
			return nil, core.NewSsoSignatureInvalidError(client)
		}
	}
	if p.tpl.mgr != nil {
		_ = p.tpl.mgr.Logout(loginID)
	}
	return map[string]any{"received": true, "loginId": loginID}, nil
}
