package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	core "github.com/sa-tokens/sa-token-go/core"
	"github.com/sa-tokens/sa-token-go/core/manager"
	"github.com/sa-tokens/sa-token-go/plugins/sso/common"
)

// Config SSO client configuration | SSO 客户端配置
type Config struct {
	ServerURL string
	ClientID  string
	SecretKey string
}

// Signer signs outbound SSO requests | 签名器
type Signer struct {
	Secret string
}

// Sign builds HMAC signature for params | 签名参数表
func (s Signer) Sign(params map[string]string) string {
	return common.HMACSign(s.Secret, params)
}

// Template SSO client template | SSO 客户端模板
type Template struct {
	cfg    *Config
	signer Signer
	mgr    *manager.Manager // optional: local logout on SLO callback | 可选：回调时本地登出
}

// NewTemplate constructs client template | 构造客户端模板
func NewTemplate(cfg *Config) *Template {
	if cfg == nil {
		cfg = &Config{}
	}
	return &Template{cfg: cfg, signer: Signer{Secret: cfg.SecretKey}}
}

// SetManager wires manager for local logout (logout/logoutCall) | 绑定 Manager 用于本地登出
func (t *Template) SetManager(m *manager.Manager) {
	t.mgr = m
}

// BuildServerAuthURL builds redirect to SSO server /sso/auth | 构造跳转认证 URL
func (t *Template) BuildServerAuthURL(clientLoginURL, back string) string {
	if t.cfg.ServerURL == "" {
		return ""
	}
	u, err := url.Parse(t.cfg.ServerURL + "/sso/auth")
	if err != nil {
		return ""
	}
	q := u.Query()
	q.Set("redirect", clientLoginURL+"?back="+url.QueryEscape(back))
	u.RawQuery = q.Encode()
	return u.String()
}

// CheckTicket exchanges ticket for loginId via server /sso/checkTicket | 向服务端校验 ticket
func (t *Template) CheckTicket(ticket, ssoLogoutCallURL string) (loginID string, err error) {
	if t.cfg.ServerURL == "" {
		return "", core.ErrSsoServerUnreachable
	}
	form := url.Values{}
	form.Set("ticket", ticket)
	form.Set("client", t.cfg.ClientID)
	if ssoLogoutCallURL != "" {
		form.Set("ssoLogoutCall", ssoLogoutCallURL)
	}
	if t.cfg.SecretKey != "" {
		params := map[string]string{"ticket": ticket, "client": t.cfg.ClientID}
		form.Set("sign", t.signer.Sign(params))
	}
	endpoint := strings.TrimRight(t.cfg.ServerURL, "/") + "/sso/checkTicket"
	resp, err := http.PostForm(endpoint, form)
	if err != nil {
		return "", core.ErrSsoServerUnreachable
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", core.NewError(core.CodeBadRequest, "invalid ticket", core.ErrSsoInvalidTicket).
			WithContext("ticket", ticket).WithContext("status", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var parsed struct {
		LoginID string `json:"loginId"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil || parsed.LoginID == "" {
		return "", core.NewError(core.CodeBadRequest, "invalid ticket", core.ErrSsoInvalidTicket).
			WithContext("ticket", ticket)
	}
	return parsed.LoginID, nil
}
