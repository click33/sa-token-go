package hertz

import (
	"strings"

	"github.com/click33/sa-token-go/core/adapter"
	hertzapp "github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
)

// HertzContext adapts Hertz request context HertzContext 适配 Hertz 请求上下文
type HertzContext struct {
	ctx     *hertzapp.RequestContext
	aborted bool
}

// NewHertzContext creates request context adapter NewHertzContext 创建请求上下文适配器
func NewHertzContext(ctx *hertzapp.RequestContext) adapter.RequestContext {
	return &HertzContext{ctx: ctx}
}

// Get implements adapter.RequestContext Get 实现适配器上下文读取
func (h *HertzContext) Get(key string) (any, bool) {
	return h.ctx.Get(key)
}

// GetHeaders implements adapter.RequestContext GetHeaders 实现请求头读取
func (h *HertzContext) GetHeaders() map[string][]string {
	headers := make(map[string][]string)
	h.ctx.VisitAllHeaders(func(key, value []byte) {
		headers[string(key)] = append(headers[string(key)], string(value))
	})
	return headers
}

// GetHeader implements adapter.RequestContext GetHeader 实现单个请求头读取
func (h *HertzContext) GetHeader(key string) string {
	return string(h.ctx.GetHeader(key))
}

// GetQuery implements adapter.RequestContext GetQuery 实现查询参数读取
func (h *HertzContext) GetQuery(key string) string {
	return h.ctx.Query(key)
}

// GetQueryAll implements adapter.RequestContext GetQueryAll 实现全部查询参数读取
func (h *HertzContext) GetQueryAll() map[string][]string {
	query := make(map[string][]string)
	h.ctx.VisitAllQueryArgs(func(key, value []byte) {
		query[string(key)] = append(query[string(key)], string(value))
	})
	return query
}

// GetPostForm implements adapter.RequestContext GetPostForm 实现表单参数读取
func (h *HertzContext) GetPostForm(key string) string {
	return h.ctx.PostForm(key)
}

// GetCookie implements adapter.RequestContext GetCookie 实现 Cookie 读取
func (h *HertzContext) GetCookie(key string) string {
	return string(h.ctx.Cookie(key))
}

// GetBody implements adapter.RequestContext GetBody 实现请求体读取
func (h *HertzContext) GetBody() ([]byte, error) {
	return h.ctx.Body()
}

// GetClientIP implements adapter.RequestContext GetClientIP 实现客户端 IP 读取
func (h *HertzContext) GetClientIP() string {
	return h.ctx.ClientIP()
}

// GetMethod implements adapter.RequestContext GetMethod 实现请求方法读取
func (h *HertzContext) GetMethod() string {
	return string(h.ctx.Method())
}

// GetPath implements adapter.RequestContext GetPath 实现请求路径读取
func (h *HertzContext) GetPath() string {
	return string(h.ctx.Path())
}

// GetURL implements adapter.RequestContext GetURL 实现完整 URL 读取
func (h *HertzContext) GetURL() string {
	return h.ctx.URI().String()
}

// GetUserAgent implements adapter.RequestContext GetUserAgent 实现 User-Agent 读取
func (h *HertzContext) GetUserAgent() string {
	return string(h.ctx.UserAgent())
}

// IsTLS implements adapter.RequestContext IsTLS 实现 TLS 状态读取
func (h *HertzContext) IsTLS() bool {
	return strings.EqualFold(string(h.ctx.URI().Scheme()), "https")
}

// SetStatusCode implements adapter.RequestContext SetStatusCode 实现状态码写入
func (h *HertzContext) SetStatusCode(code int) {
	h.ctx.SetStatusCode(code)
}

// SetHeader implements adapter.RequestContext SetHeader 实现响应头写入
func (h *HertzContext) SetHeader(key, value string) {
	h.ctx.Header(key, value)
}

// Write implements adapter.RequestContext Write 实现响应体写入
func (h *HertzContext) Write(data []byte) (int, error) {
	return h.ctx.Write(data)
}

// SetCookie implements adapter.RequestContext SetCookie 实现 Cookie 写入
func (h *HertzContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	h.ctx.SetCookie(name, value, maxAge, path, domain, protocol.CookieSameSiteLaxMode, secure, httpOnly)
}

// SetCookieWithOptions implements adapter.RequestContext SetCookieWithOptions 实现带选项 Cookie 写入
func (h *HertzContext) SetCookieWithOptions(options *adapter.CookieOptions) {
	sameSite := protocol.CookieSameSiteLaxMode
	switch options.SameSite {
	case "Strict":
		sameSite = protocol.CookieSameSiteStrictMode
	case "None":
		sameSite = protocol.CookieSameSiteNoneMode
	case "":
		sameSite = protocol.CookieSameSiteLaxMode
	}

	h.ctx.SetCookie(
		options.Name,
		options.Value,
		options.MaxAge,
		options.Path,
		options.Domain,
		sameSite,
		options.Secure,
		options.HttpOnly,
	)
}

// Set implements adapter.RequestContext Set 实现上下文写入
func (h *HertzContext) Set(key string, value any) {
	h.ctx.Set(key, value)
}

// GetString implements adapter.RequestContext GetString 实现字符串上下文读取
func (h *HertzContext) GetString(key string) string {
	return h.ctx.GetString(key)
}

// MustGet implements adapter.RequestContext MustGet 实现强制上下文读取
func (h *HertzContext) MustGet(key string) any {
	return h.ctx.MustGet(key)
}

// Abort implements adapter.RequestContext Abort 实现请求中断
func (h *HertzContext) Abort() {
	h.aborted = true
	h.ctx.Abort()
}

// IsAborted implements adapter.RequestContext IsAborted 实现中断状态读取
func (h *HertzContext) IsAborted() bool {
	return h.aborted || h.ctx.IsAborted()
}
