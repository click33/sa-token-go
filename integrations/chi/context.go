package chi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/click33/sa-token-go/core/adapter"
)

// ChiContext adapts request context ChiContext 适配 Chi 请求上下文
type ChiContext struct {
	w       http.ResponseWriter
	r       *http.Request
	ctx     context.Context
	aborted bool
}

// NewChiContext creates request context adapter NewChiContext 创建请求上下文适配器
func NewChiContext(w http.ResponseWriter, r *http.Request) adapter.RequestContext {
	return &ChiContext{
		w:   w,
		r:   r,
		ctx: r.Context(),
	}
}

// Get implements adapter.RequestContext Get 实现 adapter.RequestContext 接口
func (c *ChiContext) Get(key string) (interface{}, bool) {
	value := c.ctx.Value(key)
	return value, value != nil
}

// GetHeaders implements adapter.RequestContext GetHeaders 实现 adapter.RequestContext 接口
func (c *ChiContext) GetHeaders() map[string][]string {
	return c.r.Header
}

// GetHeader implements adapter.RequestContext GetHeader 实现 adapter.RequestContext 接口
func (c *ChiContext) GetHeader(key string) string {
	return c.r.Header.Get(key)
}

// GetQuery implements adapter.RequestContext GetQuery 实现 adapter.RequestContext 接口
func (c *ChiContext) GetQuery(key string) string {
	return c.r.URL.Query().Get(key)
}

// GetQueryAll implements adapter.RequestContext GetQueryAll 实现 adapter.RequestContext 接口
func (c *ChiContext) GetQueryAll() map[string][]string {
	return c.r.URL.Query()
}

// GetPostForm implements adapter.RequestContext GetPostForm 实现 adapter.RequestContext 接口
func (c *ChiContext) GetPostForm(key string) string {
	return c.r.FormValue(key)
}

// GetCookie implements adapter.RequestContext GetCookie 实现 adapter.RequestContext 接口
func (c *ChiContext) GetCookie(key string) string {
	cookie, err := c.r.Cookie(key)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// GetBody implements adapter.RequestContext GetBody 实现 adapter.RequestContext 接口
func (c *ChiContext) GetBody() ([]byte, error) {
	body, err := io.ReadAll(c.r.Body)
	if err != nil {
		return nil, err
	}
	c.r.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

// GetClientIP implements adapter.RequestContext GetClientIP 实现 adapter.RequestContext 接口
func (c *ChiContext) GetClientIP() string {
	if ip := strings.TrimSpace(c.r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}
	if forwarded := strings.TrimSpace(c.r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(c.r.RemoteAddr)
	if err == nil {
		return host
	}
	return c.r.RemoteAddr
}

// GetMethod implements adapter.RequestContext GetMethod 实现 adapter.RequestContext 接口
func (c *ChiContext) GetMethod() string {
	return c.r.Method
}

// GetPath implements adapter.RequestContext GetPath 实现 adapter.RequestContext 接口
func (c *ChiContext) GetPath() string {
	return c.r.URL.Path
}

// GetURL implements adapter.RequestContext GetURL 实现 adapter.RequestContext 接口
func (c *ChiContext) GetURL() string {
	return c.r.URL.String()
}

// GetUserAgent implements adapter.RequestContext GetUserAgent 实现 adapter.RequestContext 接口
func (c *ChiContext) GetUserAgent() string {
	return c.r.UserAgent()
}

// IsTLS implements adapter.RequestContext IsTLS 实现 adapter.RequestContext 接口
func (c *ChiContext) IsTLS() bool {
	return c.r.TLS != nil
}

// SetStatusCode implements adapter.RequestContext SetStatusCode 实现 adapter.RequestContext 接口
func (c *ChiContext) SetStatusCode(code int) {
	c.w.WriteHeader(code)
}

// SetHeader implements adapter.RequestContext SetHeader 实现 adapter.RequestContext 接口
func (c *ChiContext) SetHeader(key, value string) {
	c.w.Header().Set(key, value)
}

// Write implements adapter.RequestContext Write 实现 adapter.RequestContext 接口
func (c *ChiContext) Write(data []byte) (int, error) {
	return c.w.Write(data)
}

// SetCookie implements adapter.RequestContext SetCookie 实现 adapter.RequestContext 接口
func (c *ChiContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	http.SetCookie(c.w, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: http.SameSiteLaxMode,
	})
}

// SetCookieWithOptions implements adapter.RequestContext SetCookieWithOptions 实现 adapter.RequestContext 接口
func (c *ChiContext) SetCookieWithOptions(options *adapter.CookieOptions) {
	cookie := &http.Cookie{
		Name:     options.Name,
		Value:    options.Value,
		MaxAge:   options.MaxAge,
		Path:     options.Path,
		Domain:   options.Domain,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
		SameSite: http.SameSiteLaxMode,
	}

	switch options.SameSite {
	case "Strict":
		cookie.SameSite = http.SameSiteStrictMode
	case "Lax":
		cookie.SameSite = http.SameSiteLaxMode
	case "None":
		cookie.SameSite = http.SameSiteNoneMode
	}

	http.SetCookie(c.w, cookie)
}

// Set implements adapter.RequestContext Set 实现 adapter.RequestContext 接口
func (c *ChiContext) Set(key string, value interface{}) {
	c.ctx = context.WithValue(c.ctx, key, value)
	c.r = c.r.WithContext(c.ctx)
}

// GetString implements adapter.RequestContext GetString 实现 adapter.RequestContext 接口
func (c *ChiContext) GetString(key string) string {
	value := c.ctx.Value(key)
	if value == nil {
		return ""
	}
	str, ok := value.(string)
	if !ok {
		return ""
	}
	return str
}

// MustGet implements adapter.RequestContext MustGet 实现 adapter.RequestContext 接口
func (c *ChiContext) MustGet(key string) any {
	value := c.ctx.Value(key)
	if value == nil {
		panic("key not found: " + key)
	}
	return value
}

// Abort implements adapter.RequestContext Abort 实现 adapter.RequestContext 接口
func (c *ChiContext) Abort() {
	c.aborted = true
}

// IsAborted implements adapter.RequestContext IsAborted 实现 adapter.RequestContext 接口
func (c *ChiContext) IsAborted() bool {
	return c.aborted
}

// JSON implements adapter.RequestContextExt JSON 实现 adapter.RequestContextExt 接口
func (c *ChiContext) JSON(code int, v any) error {
	c.w.Header().Set("Content-Type", "application/json")
	c.w.WriteHeader(code)
	return json.NewEncoder(c.w).Encode(v)
}

// GetRawRequest implements adapter.RequestContextExt GetRawRequest 实现 adapter.RequestContextExt 接口
func (c *ChiContext) GetRawRequest() *http.Request {
	return c.r
}

// GetRawResponseWriter implements adapter.RequestContextExt GetRawResponseWriter 实现 adapter.RequestContextExt 接口
func (c *ChiContext) GetRawResponseWriter() http.ResponseWriter {
	return c.w
}
