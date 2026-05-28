package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	core "github.com/sa-tokens/sa-token-go/core"
	"github.com/sa-tokens/sa-token-go/core/manager"
	"github.com/sa-tokens/sa-token-go/core/session"
	"github.com/sa-tokens/sa-token-go/core/utils"
	"github.com/sa-tokens/sa-token-go/plugins/sso/common"
)

const ssoClientInfoKey = "SSO_CLIENT_INFO_LIST"

// ClientCfg per-client SSO settings | 单 client 配置
type ClientCfg struct {
	ClientID string
}

// Config SSO server configuration | SSO 服务端配置
type Config struct {
	TicketTimeout int64
	AllowURL      string
	IsSlo         bool
	SecretKey     string
	MaxRegClient  int
	Clients       map[string]*ClientCfg
}

// ClientInfo SLO callback registration | SLO 回调注册信息
type ClientInfo struct {
	Client         string `json:"client"`
	SloCallbackURL string `json:"sloCallbackUrl"`
}

// TicketModel one-shot ticket | 一次性票据
type TicketModel struct {
	Ticket     string `json:"ticket"`
	Client     string `json:"client"`
	LoginID    string `json:"loginId"`
	TokenValue string `json:"tokenValue"`
}

// Template SSO server template | SSO 服务端模板
type Template struct {
	mgr *manager.Manager
	cfg *Config
}

// NewTemplate constructs SSO server template | 构造模板
func NewTemplate(mgr *manager.Manager, cfg *Config) *Template {
	if cfg == nil {
		cfg = &Config{TicketTimeout: 300, MaxRegClient: -1}
	}
	return &Template{mgr: mgr, cfg: cfg}
}

func (t *Template) keyTicket(ticket string) string {
	return t.mgr.GetConfig().KeyPrefix + "sso:ticket:" + ticket
}

func (t *Template) keyTicketIndex(client, loginID string) string {
	return t.mgr.GetConfig().KeyPrefix + "sso:ticketIdx:" + client + ":" + loginID
}

// CheckRedirectURL validates redirect against AllowURL allow list | 校验 redirect 白名单
func (t *Template) CheckRedirectURL(target string) error {
	allow := strings.TrimSpace(t.cfg.AllowURL)
	if allow == "" || allow == "*" {
		return nil
	}
	for _, p := range strings.Split(allow, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasSuffix(p, "*") {
			if strings.HasPrefix(target, strings.TrimSuffix(p, "*")) {
				return nil
			}
			continue
		}
		if p == target {
			return nil
		}
	}
	return core.NewError(core.CodeBadRequest, "invalid sso redirect url", core.ErrSsoInvalidRedirect).
		WithContext("redirect", target)
}

// CheckClient ensures client is registered when Clients map is set | 校验 client 已注册
func (t *Template) CheckClient(client string) error {
	if t.cfg.Clients == nil {
		return nil
	}
	if _, ok := t.cfg.Clients[client]; !ok {
		return core.ErrSsoClientNotRegistered
	}
	return nil
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func buildForm(params map[string]string) string {
	var b strings.Builder
	first := true
	for k, v := range params {
		if !first {
			b.WriteByte('&')
		}
		first = false
		b.WriteString(url.QueryEscape(k))
		b.WriteByte('=')
		b.WriteString(url.QueryEscape(v))
	}
	return b.String()
}

// CreateTicketAndSave creates ticket + index | 创建票据
func (t *Template) CreateTicketAndSave(client, loginID, tokenValue string) (string, error) {
	if err := t.CheckClient(client); err != nil {
		return "", err
	}
	tk := utils.RandomString(64)
	tm := &TicketModel{Ticket: tk, Client: client, LoginID: loginID, TokenValue: tokenValue}
	ttl := time.Duration(t.cfg.TicketTimeout) * time.Second
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	if err := t.mgr.GetStorage().Set(t.keyTicket(tk), tm, ttl); err != nil {
		return "", err
	}
	if err := t.mgr.GetStorage().Set(t.keyTicketIndex(client, loginID), tk, ttl); err != nil {
		return "", err
	}
	return tk, nil
}

// CheckTicketAndDelete validates and consumes ticket | 校验并消费票据
func (t *Template) CheckTicketAndDelete(ticket, client string) (*TicketModel, error) {
	v, err := t.mgr.GetStorage().Get(t.keyTicket(ticket))
	if err != nil || v == nil {
		return nil, core.ErrSsoInvalidTicket
	}
	var tm TicketModel
	switch x := v.(type) {
	case string:
		if err := json.Unmarshal([]byte(x), &tm); err != nil {
			return nil, core.ErrSsoInvalidTicket
		}
	case []byte:
		if err := json.Unmarshal(x, &tm); err != nil {
			return nil, core.ErrSsoInvalidTicket
		}
	case *TicketModel:
		tm = *x
	default:
		return nil, core.ErrSsoInvalidTicket
	}
	if client != "*" && client != tm.Client {
		return nil, core.NewSsoTicketClientMismatchError(ticket, tm.Client, client)
	}
	_ = t.mgr.GetStorage().Delete(t.keyTicket(ticket))
	_ = t.mgr.GetStorage().Delete(t.keyTicketIndex(tm.Client, tm.LoginID))
	return &tm, nil
}

func getClientInfoList(sess *session.Session) []ClientInfo {
	raw, ok := sess.Get(ssoClientInfoKey)
	if !ok || raw == nil {
		return nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var list []ClientInfo
	if err := json.Unmarshal(b, &list); err != nil {
		return nil
	}
	return list
}

// RegisterSloCallback registers SLO callback URL on Account-Session | 注册 SLO 回调
func (t *Template) RegisterSloCallback(loginID, client, sloCallbackURL string) error {
	sess, err := t.mgr.GetSession(loginID)
	if err != nil {
		return err
	}
	list := getClientInfoList(sess)
	list = append(list, ClientInfo{Client: client, SloCallbackURL: sloCallbackURL})
	if t.cfg.MaxRegClient > 0 && len(list) > t.cfg.MaxRegClient {
		evict := list[0]
		list = list[1:]
		go t.notifyClientLogout(loginID, evict, true)
	}
	cfg := t.mgr.GetConfig()
	var ttl time.Duration
	if cfg != nil && cfg.Timeout > 0 {
		ttl = time.Duration(cfg.Timeout) * time.Second
	}
	return sess.Set(ssoClientInfoKey, list, ttl)
}

func (t *Template) notifyClientLogout(loginID string, ci ClientInfo, evict bool) {
	if ci.SloCallbackURL == "" {
		return
	}
	params := map[string]string{
		"loginId": loginID,
		"client":  ci.Client,
		"evict":   boolStr(evict),
	}
	if t.cfg.SecretKey != "" {
		params["sign"] = common.HMACSign(t.cfg.SecretKey, params)
	}
	form := buildForm(params)
	req, err := http.NewRequest(http.MethodPost, ci.SloCallbackURL, strings.NewReader(form))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
}

// SsoLogout notifies clients and logs out locally | 单点注销
func (t *Template) SsoLogout(loginID string) error {
	sess, err := t.mgr.GetSession(loginID)
	if err != nil {
		return err
	}
	for _, ci := range getClientInfoList(sess) {
		go t.notifyClientLogout(loginID, ci, false)
	}
	return t.mgr.Logout(loginID)
}
