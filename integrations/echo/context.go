package echo

import (
	"bytes"
	"io"
	"net/http"

	"github.com/click33/sa-token-go/core/adapter"
	echo4 "github.com/labstack/echo/v4"
)

// EchoContext adapts Echo request context to SaToken request context Echo 请求上下文适配到 SaToken 请求上下文
type EchoContext struct {
	c       echo4.Context
	aborted bool
}

// NewEchoContext creates Echo request context adapter 创建 Echo 请求上下文适配器
func NewEchoContext(c echo4.Context) adapter.RequestContext {
	return &EchoContext{c: c}
}

// Get implements adapter.RequestContext Get 实现 adapter.RequestContext 接口
func (e *EchoContext) Get(key string) (interface{}, bool) {
	value := e.c.Get(key)
	return value, value != nil
}

// GetHeaders implements adapter.RequestContext GetHeaders 实现 adapter.RequestContext 接口
func (e *EchoContext) GetHeaders() map[string][]string {
	return e.c.Request().Header
}

// GetHeader implements adapter.RequestContext GetHeader 实现 adapter.RequestContext 接口
func (e *EchoContext) GetHeader(key string) string {
	return e.c.Request().Header.Get(key)
}

// GetQuery implements adapter.RequestContext GetQuery 实现 adapter.RequestContext 接口
func (e *EchoContext) GetQuery(key string) string {
	return e.c.QueryParam(key)
}

// GetQueryAll implements adapter.RequestContext GetQueryAll 实现 adapter.RequestContext 接口
func (e *EchoContext) GetQueryAll() map[string][]string {
	return e.c.QueryParams()
}

// GetPostForm implements adapter.RequestContext GetPostForm 实现 adapter.RequestContext 接口
func (e *EchoContext) GetPostForm(key string) string {
	return e.c.FormValue(key)
}

// GetCookie implements adapter.RequestContext GetCookie 实现 adapter.RequestContext 接口
func (e *EchoContext) GetCookie(key string) string {
	cookie, err := e.c.Cookie(key)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// GetBody implements adapter.RequestContext GetBody 实现 adapter.RequestContext 接口
func (e *EchoContext) GetBody() ([]byte, error) {
	body, err := io.ReadAll(e.c.Request().Body)
	if err != nil {
		return nil, err
	}
	e.c.Request().Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

// GetClientIP implements adapter.RequestContext GetClientIP 实现 adapter.RequestContext 接口
func (e *EchoContext) GetClientIP() string {
	return e.c.RealIP()
}

// GetMethod implements adapter.RequestContext GetMethod 实现 adapter.RequestContext 接口
func (e *EchoContext) GetMethod() string {
	return e.c.Request().Method
}

// GetPath implements adapter.RequestContext GetPath 实现 adapter.RequestContext 接口
func (e *EchoContext) GetPath() string {
	return e.c.Request().URL.Path
}

// GetURL implements adapter.RequestContext GetURL 实现 adapter.RequestContext 接口
func (e *EchoContext) GetURL() string {
	return e.c.Request().URL.String()
}

// GetUserAgent implements adapter.RequestContext GetUserAgent 实现 adapter.RequestContext 接口
func (e *EchoContext) GetUserAgent() string {
	return e.c.Request().UserAgent()
}

// IsTLS implements adapter.RequestContext IsTLS 实现 adapter.RequestContext 接口
func (e *EchoContext) IsTLS() bool {
	return e.c.Request().TLS != nil
}

// SetStatusCode implements adapter.RequestContext SetStatusCode 实现 adapter.RequestContext 接口
func (e *EchoContext) SetStatusCode(code int) {
	e.c.Response().WriteHeader(code)
}

// SetHeader implements adapter.RequestContext SetHeader 实现 adapter.RequestContext 接口
func (e *EchoContext) SetHeader(key, value string) {
	e.c.Response().Header().Set(key, value)
}

// Write implements adapter.RequestContext Write 实现 adapter.RequestContext 接口
func (e *EchoContext) Write(data []byte) (int, error) {
	return e.c.Response().Write(data)
}

// SetCookie implements adapter.RequestContext SetCookie 实现 adapter.RequestContext 接口
func (e *EchoContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	e.c.SetCookie(&http.Cookie{
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
func (e *EchoContext) SetCookieWithOptions(options *adapter.CookieOptions) {
	cookie := &http.Cookie{
		Name:     options.Name,
		Value:    options.Value,
		MaxAge:   options.MaxAge,
		Path:     options.Path,
		Domain:   options.Domain,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
	}

	switch options.SameSite {
	case "Strict":
		cookie.SameSite = http.SameSiteStrictMode
	case "Lax":
		cookie.SameSite = http.SameSiteLaxMode
	case "None":
		cookie.SameSite = http.SameSiteNoneMode
	default:
		cookie.SameSite = http.SameSiteLaxMode
	}

	e.c.SetCookie(cookie)
}

// Set implements adapter.RequestContext Set 实现 adapter.RequestContext 接口
func (e *EchoContext) Set(key string, value interface{}) {
	e.c.Set(key, value)
}

// GetString implements adapter.RequestContext GetString 实现 adapter.RequestContext 接口
func (e *EchoContext) GetString(key string) string {
	value, ok := e.c.Get(key).(string)
	if !ok {
		return ""
	}
	return value
}

// MustGet implements adapter.RequestContext MustGet 实现 adapter.RequestContext 接口
func (e *EchoContext) MustGet(key string) any {
	value := e.c.Get(key)
	if value == nil {
		panic("key not found: " + key)
	}
	return value
}

// Abort implements adapter.RequestContext Abort 实现 adapter.RequestContext 接口
func (e *EchoContext) Abort() {
	e.aborted = true
}

// IsAborted implements adapter.RequestContext IsAborted 实现 adapter.RequestContext 接口
func (e *EchoContext) IsAborted() bool {
	return e.aborted
}
