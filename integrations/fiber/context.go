package fiber

import (
	"github.com/click33/sa-token-go/core/adapter"
	"github.com/gofiber/fiber/v2"
	"time"
)

// FiberContext Fiber request context adapter | Fiber请求上下文适配器
type FiberContext struct {
	c       *fiber.Ctx
	aborted bool
}

// NewFiberContext creates a Fiber context adapter | 创建Fiber上下文适配器
func NewFiberContext(c *fiber.Ctx) adapter.RequestContext {
	return &FiberContext{c: c}
}

// GetHeader gets request header | 获取请求头
func (f *FiberContext) GetHeader(key string) string {
	return f.c.Get(key)
}

// GetQuery gets query parameter | 获取查询参数
func (f *FiberContext) GetQuery(key string) string {
	return f.c.Query(key)
}

// GetCookie gets cookie | 获取Cookie
func (f *FiberContext) GetCookie(key string) string {
	return f.c.Cookies(key)
}

// SetHeader sets response header | 设置响应头
func (f *FiberContext) SetHeader(key, value string) {
	f.c.Set(key, value)
}

// SetCookie sets cookie | 设置Cookie
func (f *FiberContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	cookie := &fiber.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Domain:   domain,
		MaxAge:   maxAge,
		Secure:   secure,
		HTTPOnly: httpOnly,
		SameSite: "Lax",
	}
	if maxAge > 0 {
		cookie.Expires = time.Now().Add(time.Duration(maxAge) * time.Second)
	}
	f.c.Cookie(cookie)
}

// GetClientIP gets client IP address | 获取客户端IP地址
func (f *FiberContext) GetClientIP() string {
	return f.c.IP()
}

// GetMethod gets request method | 获取请求方法
func (f *FiberContext) GetMethod() string {
	return f.c.Method()
}

// GetPath gets request path | 获取请求路径
func (f *FiberContext) GetPath() string {
	return f.c.Path()
}

// Set sets context value | 设置上下文值
func (f *FiberContext) Set(key string, value interface{}) {
	f.c.Locals(key, value)
}

// Get gets context value | 获取上下文值
func (f *FiberContext) Get(key string) (interface{}, bool) {
	value := f.c.Locals(key)
	return value, value != nil
}

// ============ Additional Required Methods | 额外必需的方法 ============

// GetHeaders implements adapter.RequestContext.
func (f *FiberContext) GetHeaders() map[string][]string {
	headers := make(map[string][]string)
	f.c.Request().Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = []string{string(value)}
	})
	return headers
}

// GetQueryAll implements adapter.RequestContext.
func (f *FiberContext) GetQueryAll() map[string][]string {
	query := f.c.Request().URI().QueryArgs()
	params := make(map[string][]string)
	query.VisitAll(func(key, value []byte) {
		params[string(key)] = []string{string(value)}
	})
	return params
}

// GetPostForm implements adapter.RequestContext.
func (f *FiberContext) GetPostForm(key string) string {
	return f.c.FormValue(key)
}

// GetBody implements adapter.RequestContext.
func (f *FiberContext) GetBody() ([]byte, error) {
	return f.c.Body(), nil
}

// GetURL implements adapter.RequestContext.
func (f *FiberContext) GetURL() string {
	return string(f.c.Request().URI().FullURI())
}

// GetUserAgent implements adapter.RequestContext.
func (f *FiberContext) GetUserAgent() string {
	return f.c.Get("User-Agent")
}

// SetCookieWithOptions implements adapter.RequestContext.
func (f *FiberContext) SetCookieWithOptions(options *adapter.CookieOptions) {
	cookie := &fiber.Cookie{
		Name:     options.Name,
		Value:    options.Value,
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HTTPOnly: options.HttpOnly,
		SameSite: "Lax", // Default to Lax
	}
	
	// Set SameSite attribute
	switch options.SameSite {
	case "Strict":
		cookie.SameSite = "Strict"
	case "Lax":
		cookie.SameSite = "Lax"
	case "None":
		cookie.SameSite = "None"
	}
	
	if options.MaxAge > 0 {
		cookie.Expires = time.Now().Add(time.Duration(options.MaxAge) * time.Second)
	}
	
	f.c.Cookie(cookie)
}

// GetString implements adapter.RequestContext.
func (f *FiberContext) GetString(key string) string {
	value := f.c.Locals(key)
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// MustGet implements adapter.RequestContext.
func (f *FiberContext) MustGet(key string) any {
	value := f.c.Locals(key)
	if value == nil {
		panic("key not found: " + key)
	}
	return value
}

// Abort implements adapter.RequestContext.
func (f *FiberContext) Abort() {
	f.aborted = true
}

// IsAborted implements adapter.RequestContext.
func (f *FiberContext) IsAborted() bool {
	return f.aborted
}
