package gozero

import (
	"context"
	"io"
	"net/http"

	"github.com/sa-tokens/sa-token-go/core/adapter"
)

// GoZeroContext go-zero request context adapter | go-zero请求上下文适配器
type GoZeroContext struct {
	w       http.ResponseWriter
	r       *http.Request
	ctx     context.Context
	aborted bool
}

// NewGoZeroContext creates a go-zero context adapter | 创建go-zero上下文适配器
func NewGoZeroContext(w http.ResponseWriter, r *http.Request) adapter.RequestContext {
	return &GoZeroContext{
		w:   w,
		r:   r,
		ctx: r.Context(),
	}
}

// Request returns the underlying request with updated context | 返回携带更新后上下文的请求
func (c *GoZeroContext) Request() *http.Request {
	return c.r
}

// GetHeader gets request header | 获取请求头
func (c *GoZeroContext) GetHeader(key string) string {
	return c.r.Header.Get(key)
}

// GetHeaders gets all request headers | 获取所有请求头
func (c *GoZeroContext) GetHeaders() map[string][]string {
	headers := make(map[string][]string)
	for key, values := range c.r.Header {
		headers[key] = values
	}
	return headers
}

// GetQuery gets query parameter | 获取查询参数
func (c *GoZeroContext) GetQuery(key string) string {
	return c.r.URL.Query().Get(key)
}

// GetQueryAll gets all query parameters | 获取所有查询参数
func (c *GoZeroContext) GetQueryAll() map[string][]string {
	query := c.r.URL.Query()
	params := make(map[string][]string)
	for key, values := range query {
		params[key] = values
	}
	return params
}

// GetPostForm gets form parameter | 获取表单参数
func (c *GoZeroContext) GetPostForm(key string) string {
	return c.r.FormValue(key)
}

// GetCookie gets cookie | 获取Cookie
func (c *GoZeroContext) GetCookie(key string) string {
	cookie, err := c.r.Cookie(key)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// GetBody gets request body | 获取请求体
func (c *GoZeroContext) GetBody() ([]byte, error) {
	return io.ReadAll(c.r.Body)
}

// GetClientIP gets client IP address | 获取客户端IP地址
func (c *GoZeroContext) GetClientIP() string {
	ip := c.r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = c.r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = c.r.RemoteAddr
	}
	return ip
}

// GetMethod gets request method | 获取请求方法
func (c *GoZeroContext) GetMethod() string {
	return c.r.Method
}

// GetPath gets request path | 获取请求路径
func (c *GoZeroContext) GetPath() string {
	return c.r.URL.Path
}

// GetURL gets request URL | 获取请求URL
func (c *GoZeroContext) GetURL() string {
	return c.r.URL.String()
}

// GetUserAgent gets User-Agent | 获取User-Agent
func (c *GoZeroContext) GetUserAgent() string {
	return c.r.UserAgent()
}

// SetHeader sets response header | 设置响应头
func (c *GoZeroContext) SetHeader(key, value string) {
	c.w.Header().Set(key, value)
}

// SetCookie sets cookie | 设置Cookie
func (c *GoZeroContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
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

// SetCookieWithOptions sets cookie with options | 使用选项设置Cookie
func (c *GoZeroContext) SetCookieWithOptions(options *adapter.CookieOptions) {
	cookie := &http.Cookie{
		Name:     options.Name,
		Value:    options.Value,
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
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

// Set sets context value | 设置上下文值
func (c *GoZeroContext) Set(key string, value interface{}) {
	c.ctx = context.WithValue(c.ctx, key, value)
	c.r = c.r.WithContext(c.ctx)
}

// SetKey stores a value using typed ctxKey (preferred for Sa-Token internal keys).
// SetKey 使用类型安全的 ctxKey 存值（Sa-Token 内部键推荐使用）。
func (c *GoZeroContext) SetKey(key ctxKey, value interface{}) {
	c.ctx = context.WithValue(c.ctx, key, value)
	c.r = c.r.WithContext(c.ctx)
}

// GetKey gets a value by typed ctxKey.
// GetKey 按类型安全的 ctxKey 取值。
func (c *GoZeroContext) GetKey(key ctxKey) (interface{}, bool) {
	v := c.ctx.Value(key)
	return v, v != nil
}

// Get gets context value | 获取上下文值
func (c *GoZeroContext) Get(key string) (interface{}, bool) {
	value := c.ctx.Value(key)
	return value, value != nil
}

// GetString gets string from context | 从上下文获取字符串值
func (c *GoZeroContext) GetString(key string) string {
	value := c.ctx.Value(key)
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// MustGet must get value from context (panics if not found) | 必须获取上下文值（未找到则panic）
func (c *GoZeroContext) MustGet(key string) any {
	value := c.ctx.Value(key)
	if value == nil {
		panic("key not found: " + key)
	}
	return value
}

// Abort marks adapter state only; go-zero middleware must return to stop the chain.
// Abort 仅标记状态；go-zero 须在中间件内 return 终止，无法像 Gin 一样 Abort 后续节点。
func (c *GoZeroContext) Abort() {
	c.aborted = true
}

// IsAborted checks if request is aborted | 检查请求是否已中断
func (c *GoZeroContext) IsAborted() bool {
	return c.aborted
}
