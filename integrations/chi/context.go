package chi

import (
	"context"
	"net/http"

	"github.com/click33/sa-token-go/core/adapter"
)

// ChiContext Chi request context adapter | Chi请求上下文适配器
type ChiContext struct {
	w   http.ResponseWriter
	r   *http.Request
	ctx context.Context
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
