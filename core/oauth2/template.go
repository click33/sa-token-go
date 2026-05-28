package oauth2

import (
	"fmt"
	"strings"

	"github.com/sa-tokens/sa-token-go/core/adapter"
	"github.com/sa-tokens/sa-token-go/core/errs"
)

// OAuth2Template protocol helpers | OAuth2 模板（授权码等协议步骤）
type OAuth2Template struct {
	storage   adapter.Storage
	keyPrefix string
	loader    DataLoader
	generator DataGenerator
}

// NewOAuth2Template constructs a template | 构造模板
func NewOAuth2Template(storage adapter.Storage, keyPrefix string, loader DataLoader) *OAuth2Template {
	return &OAuth2Template{
		storage:   storage,
		keyPrefix: keyPrefix,
		loader:    loader,
	}
}

// CheckClientModel loads and validates client_id | 校验 client
func (t *OAuth2Template) CheckClientModel(clientID string) (*ClientModel, error) {
	if clientID == "" || t.loader == nil {
		return nil, fmt.Errorf("%w: client_id=%s", errs.ErrOAuth2InvalidClientID, clientID)
	}
	cm, err := t.loader.LoadClient(clientID)
	if err != nil {
		return nil, err
	}
	if cm == nil || cm.ClientID == "" {
		return nil, fmt.Errorf("%w: client_id=%s", errs.ErrOAuth2InvalidClientID, clientID)
	}
	return cm, nil
}

// CheckRedirectURI validates redirect_uri | 校验 redirect_uri
func (t *OAuth2Template) CheckRedirectURI(clientID, rawURL string) error {
	if !isHTTPURL(rawURL) {
		return fmt.Errorf("%w: redirect_uri=%s", errs.ErrOAuth2InvalidRedirectURI, rawURL)
	}
	u := rawURL
	if i := strings.IndexByte(u, '?'); i != -1 {
		u = u[:i]
	}
	if strings.ContainsRune(u, '@') {
		return fmt.Errorf("%w: redirect_uri=%s", errs.ErrOAuth2RedirectURIContainsAt, rawURL)
	}
	cm, err := t.CheckClientModel(clientID)
	if err != nil {
		return err
	}
	if !matchAllowURL(cm.AllowRedirectURIs, u) {
		return fmt.Errorf("%w: redirect_uri=%s", errs.ErrOAuth2IllegalRedirectURI, rawURL)
	}
	return nil
}

// CheckContractScope ensures requested scopes are contracted | 校验签约 scope
func (t *OAuth2Template) CheckContractScope(clientID string, scopes []string) error {
	cm, err := t.CheckClientModel(clientID)
	if err != nil {
		return err
	}
	for _, s := range scopes {
		if !containsStr(cm.ContractScopes, s) {
			return fmt.Errorf("%w: client_id=%s scope=%s", errs.ErrOAuth2ScopeNotContracted, clientID, s)
		}
	}
	return nil
}

// CheckClientCredential validates client_id and client_secret | 校验客户端密钥
func (t *OAuth2Template) CheckClientCredential(clientID, clientSecret string) error {
	cm, err := t.CheckClientModel(clientID)
	if err != nil {
		return err
	}
	if cm.ClientSecret != "" && cm.ClientSecret != clientSecret {
		return errs.ErrOAuth2InvalidClientSecret
	}
	return nil
}

// CheckGrantType validates client allows grant_type | 校验 grant_type 是否允许
func (t *OAuth2Template) CheckGrantType(clientID, grantType string) error {
	cm, err := t.CheckClientModel(clientID)
	if err != nil {
		return err
	}
	if len(cm.GrantTypes) == 0 {
		return nil
	}
	for _, gt := range cm.GrantTypes {
		if gt == grantType {
			return nil
		}
	}
	return fmt.Errorf("%w: client_id=%s grant_type=%s", errs.ErrOAuth2GrantTypeNotAllowed, clientID, grantType)
}

// HigherScopes scopes that always require consent UI | 高级 scope（需确认页）
func (t *OAuth2Template) HigherScopes() []string {
	return nil
}

// LowerScopes scopes ignored when deciding consent | 低级 scope（可忽略）
func (t *OAuth2Template) LowerScopes() []string {
	return nil
}

func (t *OAuth2Template) grantStorageKey(loginID, clientID string) string {
	return t.keyPrefix + "oauth2:grantScope:" + clientID + ":" + loginID
}

// IsGrantScope reports whether scopes were already granted | 是否已授权 scope
func (t *OAuth2Template) IsGrantScope(loginID, clientID string, scopes []string) bool {
	if loginID == "" || len(scopes) == 0 {
		return true
	}
	v, err := t.storage.Get(t.grantStorageKey(loginID, clientID))
	if err != nil || v == nil {
		return false
	}
	saved, _ := v.(string)
	if saved == "" {
		return false
	}
	existing := strings.Split(saved, ",")
	for _, need := range scopes {
		if !containsStr(existing, strings.TrimSpace(need)) {
			return false
		}
	}
	return true
}

// SaveGrantScope persists granted scopes | 保存已授权 scope
func (t *OAuth2Template) SaveGrantScope(loginID, clientID string, scopes []string) error {
	key := t.grantStorageKey(loginID, clientID)
	prev := ""
	if v, err := t.storage.Get(key); err == nil && v != nil {
		if s, ok := v.(string); ok {
			prev = s
		}
	}
	merged := scopes
	if prev != "" {
		parts := strings.Split(prev, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" || containsStr(merged, p) {
				continue
			}
			merged = append(merged, p)
		}
	}
	val := strings.Join(merged, ",")
	return t.storage.Set(key, val, 0)
}

// IsNeedCarefulConfirm decides whether consent UI is required | 是否需要确认授权
func (t *OAuth2Template) IsNeedCarefulConfirm(loginID, clientID string, scopes []string) bool {
	if len(scopes) == 0 {
		return false
	}
	if intersectStr(scopes, t.HigherScopes()) {
		return true
	}
	scopes = subtractStr(scopes, t.LowerScopes())
	if len(scopes) == 0 {
		return false
	}
	return !t.IsGrantScope(loginID, clientID, scopes)
}
