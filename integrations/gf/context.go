package gf

import (
	"net/http"

	"github.com/click33/sa-token-go/core/adapter"
	"github.com/gogf/gf/v2/net/ghttp"
)

type GFContext struct {
	c *ghttp.Request
}

// Get implements adapter.RequestContext.
func (g *GFContext) Get(key string) (interface{}, bool) {
	v := g.c.Get(key)
	return v, v.IsNil()
}

// GetClientIP implements adapter.RequestContext.
func (g *GFContext) GetClientIP() string {
	return g.c.GetClientIp()
}

// GetCookie implements adapter.RequestContext.
func (g *GFContext) GetCookie(key string) string {
	return g.c.Cookie.Get(key).String()
}

// GetHeader implements adapter.RequestContext.
func (g *GFContext) GetHeader(key string) string {
	return g.c.Header.Get(key)
}

// GetMethod implements adapter.RequestContext.
func (g *GFContext) GetMethod() string {
	return g.c.Method
}

// GetPath implements adapter.RequestContext.
func (g *GFContext) GetPath() string {
	return g.c.Request.URL.Path
}

// GetQuery implements adapter.RequestContext.
func (g *GFContext) GetQuery(key string) string {
	return g.c.Request.URL.Query().Get(key)
}

// Set implements adapter.RequestContext.
func (g *GFContext) Set(key string, value interface{}) {
	g.c.SetCtxVar(key, value)
}

// SetCookie implements adapter.RequestContext.
func (g *GFContext) SetCookie(name string, value string, maxAge int, path string, domain string, secure bool, httpOnly bool) {
	g.c.Cookie.SetHttpCookie(&http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

// SetHeader implements adapter.RequestContext.
func (g *GFContext) SetHeader(key string, value string) {
	g.c.Header.Set(key, value)
}

// NewGFContext creates a GF context adapter | 创建GF上下文适配器
func NewGFContext(c *ghttp.Request) adapter.RequestContext {
	return &GFContext{
		c: c,
	}
}
