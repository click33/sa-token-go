package iris

import (
	"net/http"

	"github.com/click33/sa-token-go/core/adapter"
	irisfw "github.com/kataras/iris/v12"
)

// IrisContext Iris request context adapter | Iris请求上下文适配器
type IrisContext struct {
	c       irisfw.Context
	aborted bool
}

// NewIrisContext creates an Iris context adapter | 创建Iris上下文适配器
func NewIrisContext(c irisfw.Context) adapter.RequestContext {
	return &IrisContext{c: c}
}

// GetHeader gets request header | 获取请求头
func (i *IrisContext) GetHeader(key string) string {
	return i.c.GetHeader(key)
}

// GetQuery gets query parameter | 获取查询参数
func (i *IrisContext) GetQuery(key string) string {
	return i.c.URLParam(key)
}

// GetCookie gets cookie | 获取Cookie
func (i *IrisContext) GetCookie(key string) string {
	return i.c.GetCookie(key)
}

// SetHeader sets response header | 设置响应头
func (i *IrisContext) SetHeader(key, value string) {
	i.c.Header(key, value)
}

// SetCookie sets cookie | 设置Cookie
func (i *IrisContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	i.c.SetCookie(&http.Cookie{
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

// GetClientIP gets client IP address | 获取客户端IP地址
func (i *IrisContext) GetClientIP() string {
	return i.c.RemoteAddr()
}

// GetMethod gets request method | 获取请求方法
func (i *IrisContext) GetMethod() string {
	return i.c.Method()
}

// GetPath gets request path | 获取请求路径
func (i *IrisContext) GetPath() string {
	return i.c.Path()
}

// Set sets context value | 设置上下文值
func (i *IrisContext) Set(key string, value interface{}) {
	i.c.Values().Set(key, value)
}

// Get gets context value | 获取上下文值
func (i *IrisContext) Get(key string) (interface{}, bool) {
	v := i.c.Values().Get(key)
	return v, v != nil
}

// ============ Additional Required Methods | 额外必需的方法 ============

// GetHeaders implements adapter.RequestContext.
func (i *IrisContext) GetHeaders() map[string][]string {
	headers := make(map[string][]string)
	for key, values := range i.c.Request().Header {
		headers[key] = values
	}
	return headers
}

// GetQueryAll implements adapter.RequestContext.
func (i *IrisContext) GetQueryAll() map[string][]string {
	query := i.c.Request().URL.Query()
	params := make(map[string][]string)
	for key, values := range query {
		params[key] = values
	}
	return params
}

// GetPostForm implements adapter.RequestContext.
func (i *IrisContext) GetPostForm(key string) string {
	return i.c.PostValue(key)
}

// GetBody implements adapter.RequestContext.
func (i *IrisContext) GetBody() ([]byte, error) {
	return i.c.GetBody()
}

// GetURL implements adapter.RequestContext.
func (i *IrisContext) GetURL() string {
	return i.c.Request().URL.String()
}

// GetUserAgent implements adapter.RequestContext.
func (i *IrisContext) GetUserAgent() string {
	return i.c.GetHeader("User-Agent")
}

// SetCookieWithOptions implements adapter.RequestContext.
func (i *IrisContext) SetCookieWithOptions(options *adapter.CookieOptions) {
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
	i.c.SetCookie(cookie)
}

// GetString implements adapter.RequestContext.
func (i *IrisContext) GetString(key string) string {
	return i.c.Values().GetString(key)
}

// MustGet implements adapter.RequestContext.
func (i *IrisContext) MustGet(key string) any {
	v := i.c.Values().Get(key)
	if v == nil {
		panic("key not found: " + key)
	}
	return v
}

// Abort implements adapter.RequestContext.
func (i *IrisContext) Abort() {
	i.aborted = true
	i.c.StopExecution()
}

// IsAborted implements adapter.RequestContext.
func (i *IrisContext) IsAborted() bool {
	return i.aborted
}
