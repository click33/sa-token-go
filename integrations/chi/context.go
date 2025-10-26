package chi

import (
	"context"
	"io"
	"net/http"

	"github.com/click33/sa-token-go/core/adapter"
)

// ChiContext Chi request context adapter | Chi请求上下文适配器
type ChiContext struct {
	w       http.ResponseWriter
	r       *http.Request
	ctx     context.Context
	aborted bool
}

// NewChiContext creates a Chi context adapter | 创建Chi上下文适配器
func NewChiContext(w http.ResponseWriter, r *http.Request) adapter.RequestContext {
	return &ChiContext{
		w:   w,
		r:   r,
		ctx: r.Context(),
	}
}

// GetHeader gets request header | 获取请求头
func (c *ChiContext) GetHeader(key string) string {
	return c.r.Header.Get(key)
}

// GetQuery gets query parameter | 获取查询参数
func (c *ChiContext) GetQuery(key string) string {
	return c.r.URL.Query().Get(key)
}

// GetCookie gets cookie | 获取Cookie
func (c *ChiContext) GetCookie(key string) string {
	cookie, err := c.r.Cookie(key)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// SetHeader sets response header | 设置响应头
func (c *ChiContext) SetHeader(key, value string) {
	c.w.Header().Set(key, value)
}

// SetCookie sets cookie | 设置Cookie
func (c *ChiContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(c.w, cookie)
}

// GetClientIP gets client IP address | 获取客户端IP地址
func (c *ChiContext) GetClientIP() string {
	// Try to get from common proxy headers | 尝试从常见的代理头获取
	ip := c.r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = c.r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = c.r.RemoteAddr
	}
	return ip
}

// GetMethod gets request method | 获取请求方法
func (c *ChiContext) GetMethod() string {
	return c.r.Method
}

// GetPath gets request path | 获取请求路径
func (c *ChiContext) GetPath() string {
	return c.r.URL.Path
}

// Set sets context value | 设置上下文值
func (c *ChiContext) Set(key string, value interface{}) {
	c.ctx = context.WithValue(c.ctx, key, value)
	c.r = c.r.WithContext(c.ctx)
}

// Get gets context value | 获取上下文值
func (c *ChiContext) Get(key string) (interface{}, bool) {
	value := c.ctx.Value(key)
	return value, value != nil
}

// ============ Additional Required Methods | 额外必需的方法 ============

// GetHeaders implements adapter.RequestContext.
func (c *ChiContext) GetHeaders() map[string][]string {
	headers := make(map[string][]string)
	for key, values := range c.r.Header {
		headers[key] = values
	}
	return headers
}

// GetQueryAll implements adapter.RequestContext.
func (c *ChiContext) GetQueryAll() map[string][]string {
	query := c.r.URL.Query()
	params := make(map[string][]string)
	for key, values := range query {
		params[key] = values
	}
	return params
}

// GetPostForm implements adapter.RequestContext.
func (c *ChiContext) GetPostForm(key string) string {
	return c.r.FormValue(key)
}

// GetBody implements adapter.RequestContext.
func (c *ChiContext) GetBody() ([]byte, error) {
	return io.ReadAll(c.r.Body)
}

// GetURL implements adapter.RequestContext.
func (c *ChiContext) GetURL() string {
	return c.r.URL.String()
}

// GetUserAgent implements adapter.RequestContext.
func (c *ChiContext) GetUserAgent() string {
	return c.r.UserAgent()
}

// SetCookieWithOptions implements adapter.RequestContext.
func (c *ChiContext) SetCookieWithOptions(options *adapter.CookieOptions) {
	cookie := &http.Cookie{
		Name:     options.Name,
		Value:    options.Value,
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
		SameSite: http.SameSiteLaxMode, // Default to Lax
	}
	
	// Set SameSite attribute
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

// GetString implements adapter.RequestContext.
func (c *ChiContext) GetString(key string) string {
	value := c.ctx.Value(key)
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// MustGet implements adapter.RequestContext.
func (c *ChiContext) MustGet(key string) any {
	value := c.ctx.Value(key)
	if value == nil {
		panic("key not found: " + key)
	}
	return value
}

// Abort implements adapter.RequestContext.
func (c *ChiContext) Abort() {
	c.aborted = true
}

// IsAborted implements adapter.RequestContext.
func (c *ChiContext) IsAborted() bool {
	return c.aborted
}
