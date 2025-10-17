package echo

import (
	"github.com/click33/sa-token-go/core/adapter"
	"github.com/labstack/echo/v4"
	"net/http"
)

// EchoContext Echo request context adapter | Echo请求上下文适配器
type EchoContext struct {
	c echo.Context
}

// NewEchoContext creates an Echo context adapter | 创建Echo上下文适配器
func NewEchoContext(c echo.Context) adapter.RequestContext {
	return &EchoContext{c: c}
}

// GetHeader gets request header | 获取请求头
func (e *EchoContext) GetHeader(key string) string {
	return e.c.Request().Header.Get(key)
}

// GetQuery gets query parameter | 获取查询参数
func (e *EchoContext) GetQuery(key string) string {
	return e.c.QueryParam(key)
}

// GetCookie gets cookie | 获取Cookie
func (e *EchoContext) GetCookie(key string) string {
	cookie, err := e.c.Cookie(key)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// SetHeader sets response header | 设置响应头
func (e *EchoContext) SetHeader(key, value string) {
	e.c.Response().Header().Set(key, value)
}

// SetCookie sets cookie | 设置Cookie
func (e *EchoContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	cookie := new(http.Cookie)
	cookie.Name = name
	cookie.Value = value
	cookie.MaxAge = maxAge
	cookie.Path = path
	cookie.Domain = domain
	cookie.Secure = secure
	cookie.HttpOnly = httpOnly
	cookie.SameSite = http.SameSiteLaxMode
	e.c.SetCookie(cookie)
}

// GetClientIP gets client IP address | 获取客户端IP地址
func (e *EchoContext) GetClientIP() string {
	return e.c.RealIP()
}

// GetMethod gets request method | 获取请求方法
func (e *EchoContext) GetMethod() string {
	return e.c.Request().Method
}

// GetPath gets request path | 获取请求路径
func (e *EchoContext) GetPath() string {
	return e.c.Request().URL.Path
}

// Set sets context value | 设置上下文值
func (e *EchoContext) Set(key string, value interface{}) {
	e.c.Set(key, value)
}

// Get gets context value | 获取上下文值
func (e *EchoContext) Get(key string) (interface{}, bool) {
	value := e.c.Get(key)
	return value, value != nil
}
