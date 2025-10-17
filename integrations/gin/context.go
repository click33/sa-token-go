package gin

import (
	"net/http"

	"github.com/click33/sa-token-go/core/adapter"
	"github.com/gin-gonic/gin"
)

// GinContext Gin request context adapter | Gin请求上下文适配器
type GinContext struct {
	c *gin.Context
}

// NewGinContext creates a Gin context adapter | 创建Gin上下文适配器
func NewGinContext(c *gin.Context) adapter.RequestContext {
	return &GinContext{c: c}
}

// GetHeader gets request header | 获取请求头
func (g *GinContext) GetHeader(key string) string {
	return g.c.GetHeader(key)
}

// GetQuery gets query parameter | 获取查询参数
func (g *GinContext) GetQuery(key string) string {
	return g.c.Query(key)
}

// GetCookie gets cookie | 获取Cookie
func (g *GinContext) GetCookie(key string) string {
	cookie, _ := g.c.Cookie(key)
	return cookie
}

// SetHeader sets response header | 设置响应头
func (g *GinContext) SetHeader(key, value string) {
	g.c.Header(key, value)
}

// SetCookie sets cookie | 设置Cookie
func (g *GinContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	g.c.SetCookie(name, value, maxAge, path, domain, secure, httpOnly)
	g.c.SetSameSite(http.SameSiteLaxMode)
}

// GetClientIP gets client IP address | 获取客户端IP地址
func (g *GinContext) GetClientIP() string {
	return g.c.ClientIP()
}

// GetMethod gets request method | 获取请求方法
func (g *GinContext) GetMethod() string {
	return g.c.Request.Method
}

// GetPath gets request path | 获取请求路径
func (g *GinContext) GetPath() string {
	return g.c.Request.URL.Path
}

// Set sets context value | 设置上下文值
func (g *GinContext) Set(key string, value interface{}) {
	g.c.Set(key, value)
}

// Get gets context value | 获取上下文值
func (g *GinContext) Get(key string) (interface{}, bool) {
	return g.c.Get(key)
}
