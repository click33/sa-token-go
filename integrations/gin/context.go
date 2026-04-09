package gin

import (
	"net/http"

	"github.com/click33/sa-token-go/core/adapter"
	"github.com/gin-gonic/gin"
)

// GinContext adapts request context GinContext 适配 Gin 请求上下文
type GinContext struct {
	c       *gin.Context
	aborted bool
}

// NewGinContext creates request context adapter 创建请求上下文适配器
func NewGinContext(c *gin.Context) adapter.RequestContext {
	return &GinContext{
		c: c,
	}
}

// Get implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) Get(key string) (interface{}, bool) {
	return g.c.Get(key)
}

// GetClientIP implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) GetClientIP() string {
	return g.c.ClientIP()
}

// GetCookie implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) GetCookie(key string) string {
	cookie, _ := g.c.Cookie(key)
	return cookie
}

// GetHeader implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) GetHeader(key string) string {
	return g.c.GetHeader(key)
}

// GetMethod implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) GetMethod() string {
	return g.c.Request.Method
}

// GetPath implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) GetPath() string {
	return g.c.Request.URL.Path
}

// GetQuery implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) GetQuery(key string) string {
	return g.c.Query(key)
}

// Set implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) Set(key string, value interface{}) {
	g.c.Set(key, value)
}

// SetCookie implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) SetCookie(name string, value string, maxAge int, path string, domain string, secure bool, httpOnly bool) {
	g.c.SetCookie(name, value, maxAge, path, domain, secure, httpOnly)
	g.c.SetSameSite(http.SameSiteLaxMode)
}

// SetHeader implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) SetHeader(key string, value string) {
	g.c.Header(key, value)
}

// GetHeaders implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) GetHeaders() map[string][]string {
	return g.c.Request.Header
}

// GetQueryAll implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) GetQueryAll() map[string][]string {
	return g.c.Request.URL.Query()
}

// GetPostForm implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) GetPostForm(key string) string {
	return g.c.PostForm(key)
}

// GetBody implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) GetBody() ([]byte, error) {
	return g.c.GetRawData()
}

// GetURL implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) GetURL() string {
	return g.c.Request.URL.String()
}

// GetUserAgent implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) GetUserAgent() string {
	return g.c.GetHeader("User-Agent")
}

// SetCookieWithOptions implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) SetCookieWithOptions(options *adapter.CookieOptions) {
	g.c.SetCookie(
		options.Name,
		options.Value,
		options.MaxAge,
		options.Path,
		options.Domain,
		options.Secure,
		options.HttpOnly,
	)

	switch options.SameSite {
	case "Strict":
		g.c.SetSameSite(http.SameSiteStrictMode)
	case "Lax":
		g.c.SetSameSite(http.SameSiteLaxMode)
	case "None":
		g.c.SetSameSite(http.SameSiteNoneMode)
	}
}

// GetString implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) GetString(key string) string {
	v := g.c.GetString(key)
	return v
}

// MustGet implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) MustGet(key string) any {
	v, exists := g.c.Get(key)
	if !exists {
		panic("key not found: " + key)
	}
	return v
}

// Abort implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) Abort() {
	g.aborted = true
	g.c.Abort()
}

// IsAborted implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) IsAborted() bool {
	return g.aborted
}

// IsTLS implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) IsTLS() bool {
	return g.c.Request.TLS != nil
}

// SetStatusCode implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) SetStatusCode(code int) {
	g.c.Status(code)
}

// Write implements adapter.RequestContext 实现 adapter.RequestContext 接口
func (g *GinContext) Write(data []byte) (int, error) {
	return g.c.Writer.Write(data)
}
