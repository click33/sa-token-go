package fiber

import (
	"github.com/click33/sa-token-go/core/adapter"
	gofiber "github.com/gofiber/fiber/v2"
)

// FiberContext adapts Fiber request context to SaToken request context FiberContext 将 Fiber 请求上下文适配为 SaToken 请求上下文。
type FiberContext struct {
	c       *gofiber.Ctx
	aborted bool
}

// NewFiberContext creates a Fiber request context adapter NewFiberContext 创建 Fiber 请求上下文适配器。
func NewFiberContext(c *gofiber.Ctx) adapter.RequestContext {
	return &FiberContext{c: c}
}

// Get implements adapter.RequestContext Get 实现 adapter.RequestContext 接口。
func (f *FiberContext) Get(key string) (interface{}, bool) {
	value := f.c.Locals(key)
	return value, value != nil
}

// GetHeaders implements adapter.RequestContext GetHeaders 实现 adapter.RequestContext 接口。
func (f *FiberContext) GetHeaders() map[string][]string {
	return f.c.GetReqHeaders()
}

// GetHeader implements adapter.RequestContext GetHeader 实现 adapter.RequestContext 接口。
func (f *FiberContext) GetHeader(key string) string {
	return f.c.Get(key)
}

// GetQuery implements adapter.RequestContext GetQuery 实现 adapter.RequestContext 接口。
func (f *FiberContext) GetQuery(key string) string {
	return f.c.Query(key)
}

// GetQueryAll implements adapter.RequestContext GetQueryAll 实现 adapter.RequestContext 接口。
func (f *FiberContext) GetQueryAll() map[string][]string {
	values := f.c.Queries()
	result := make(map[string][]string, len(values))
	for key, value := range values {
		result[key] = []string{value}
	}
	return result
}

// GetPostForm implements adapter.RequestContext GetPostForm 实现 adapter.RequestContext 接口。
func (f *FiberContext) GetPostForm(key string) string {
	return f.c.FormValue(key)
}

// GetCookie implements adapter.RequestContext GetCookie 实现 adapter.RequestContext 接口。
func (f *FiberContext) GetCookie(key string) string {
	return f.c.Cookies(key)
}

// GetBody implements adapter.RequestContext GetBody 实现 adapter.RequestContext 接口。
func (f *FiberContext) GetBody() ([]byte, error) {
	return f.c.Body(), nil
}

// GetClientIP implements adapter.RequestContext GetClientIP 实现 adapter.RequestContext 接口。
func (f *FiberContext) GetClientIP() string {
	return f.c.IP()
}

// GetMethod implements adapter.RequestContext GetMethod 实现 adapter.RequestContext 接口。
func (f *FiberContext) GetMethod() string {
	return f.c.Method()
}

// GetPath implements adapter.RequestContext GetPath 实现 adapter.RequestContext 接口。
func (f *FiberContext) GetPath() string {
	return f.c.Path()
}

// GetURL implements adapter.RequestContext GetURL 实现 adapter.RequestContext 接口。
func (f *FiberContext) GetURL() string {
	return f.c.OriginalURL()
}

// GetUserAgent implements adapter.RequestContext GetUserAgent 实现 adapter.RequestContext 接口。
func (f *FiberContext) GetUserAgent() string {
	return f.c.Get("User-Agent")
}

// IsTLS implements adapter.RequestContext IsTLS 实现 adapter.RequestContext 接口。
func (f *FiberContext) IsTLS() bool {
	return f.c.Secure()
}

// SetStatusCode implements adapter.RequestContext SetStatusCode 实现 adapter.RequestContext 接口。
func (f *FiberContext) SetStatusCode(code int) {
	f.c.Status(code)
}

// SetHeader implements adapter.RequestContext SetHeader 实现 adapter.RequestContext 接口。
func (f *FiberContext) SetHeader(key, value string) {
	f.c.Set(key, value)
}

// Write implements adapter.RequestContext Write 实现 adapter.RequestContext 接口。
func (f *FiberContext) Write(data []byte) (int, error) {
	return f.c.Write(data)
}

// SetCookie implements adapter.RequestContext SetCookie 实现 adapter.RequestContext 接口。
func (f *FiberContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	f.c.Cookie(&gofiber.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HTTPOnly: httpOnly,
		SameSite: "Lax",
	})
}

// SetCookieWithOptions implements adapter.RequestContext SetCookieWithOptions 实现 adapter.RequestContext 接口。
func (f *FiberContext) SetCookieWithOptions(options *adapter.CookieOptions) {
	f.c.Cookie(&gofiber.Cookie{
		Name:     options.Name,
		Value:    options.Value,
		MaxAge:   options.MaxAge,
		Path:     options.Path,
		Domain:   options.Domain,
		Secure:   options.Secure,
		HTTPOnly: options.HttpOnly,
		SameSite: options.SameSite,
	})
}

// Set implements adapter.RequestContext Set 实现 adapter.RequestContext 接口。
func (f *FiberContext) Set(key string, value interface{}) {
	f.c.Locals(key, value)
}

// GetString implements adapter.RequestContext GetString 实现 adapter.RequestContext 接口。
func (f *FiberContext) GetString(key string) string {
	value, ok := f.c.Locals(key).(string)
	if !ok {
		return ""
	}
	return value
}

// MustGet implements adapter.RequestContext MustGet 实现 adapter.RequestContext 接口。
func (f *FiberContext) MustGet(key string) any {
	value := f.c.Locals(key)
	if value == nil {
		panic("key not found: " + key)
	}
	return value
}

// Abort implements adapter.RequestContext Abort 实现 adapter.RequestContext 接口。
func (f *FiberContext) Abort() {
	f.aborted = true
}

// IsAborted implements adapter.RequestContext IsAborted 实现 adapter.RequestContext 接口。
func (f *FiberContext) IsAborted() bool {
	return f.aborted
}
