package iris

import (
	"net"
	"net/http"
	"strings"

	irisfw "github.com/kataras/iris/v12"
	"github.com/sa-tokens/sa-token-go/core/adapter"
)

// IrisContext adapts iris.Context to sa-token core's RequestContext.
// It only holds the underlying Iris context and a local aborted flag; no business state.
// IrisContext 将 iris.Context 适配为 sa-token core 的 RequestContext，
// 仅保存原始 Iris 上下文与本地的 aborted 标志，不持有任何业务状态。
type IrisContext struct {
	c       irisfw.Context
	aborted bool
}

// NewIrisContext constructs an IrisContext as adapter.RequestContext,
// so core code can read/write the request without depending on Iris types.
// NewIrisContext 构造 IrisContext 并以 adapter.RequestContext 形式返回，
// 让 core 在不依赖 Iris 具体类型的前提下读写请求信息。
func NewIrisContext(c irisfw.Context) adapter.RequestContext {
	return &IrisContext{c: c}
}

// GetHeader returns a request header value. | GetHeader 读取请求头。
func (i *IrisContext) GetHeader(key string) string { return i.c.GetHeader(key) }

// GetQuery returns a URL query parameter. | GetQuery 读取 URL Query 参数。
func (i *IrisContext) GetQuery(key string) string { return i.c.URLParam(key) }

// GetCookie returns a request cookie value. | GetCookie 读取请求 Cookie 值。
func (i *IrisContext) GetCookie(key string) string { return i.c.GetCookie(key) }

// SetHeader writes a response header. | SetHeader 写响应头。
func (i *IrisContext) SetHeader(key, value string) { i.c.Header(key, value) }

// SetCookie writes a response cookie with SameSite=Lax by default,
// which balances CSRF mitigation and browser compatibility.
// SetCookie 写响应 Cookie，默认 SameSite=Lax，兼顾 CSRF 防护与浏览器兼容性。
func (i *IrisContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	i.c.SetCookie(buildCookie(name, value, maxAge, path, domain, secure, httpOnly, http.SameSiteLaxMode))
}

// GetClientIP returns the real client IP.
// Priority: first segment of X-Forwarded-For -> X-Real-IP -> RemoteAddr stripped of port.
// The order matches chi/go-zero adapters so rate-limit / audit / IP-based bans behave
// consistently across frameworks.
// GetClientIP 获取真实客户端 IP；优先级：X-Forwarded-For 首段 -> X-Real-IP -> RemoteAddr 去端口。
// 该顺序与 chi/go-zero 适配器一致，确保限流、审计、按 IP 封禁等核心能力跨框架行为一致。
func (i *IrisContext) GetClientIP() string {
	if xff := i.c.GetHeader("X-Forwarded-For"); xff != "" {
		// XFF may be "client, proxy1, proxy2"; take first non-empty segment.
		// XFF 可能是 "client, proxy1, proxy2"，取第一个非空段。
		if idx := strings.IndexByte(xff, ','); idx >= 0 {
			xff = xff[:idx]
		}
		if ip := strings.TrimSpace(xff); ip != "" {
			return ip
		}
	}
	if xri := strings.TrimSpace(i.c.GetHeader("X-Real-IP")); xri != "" {
		return xri
	}
	// RemoteAddr looks like "ip:port" or bare ip; strip port safely.
	// RemoteAddr 形如 "ip:port" 或纯 ip，需安全去端口。
	remote := i.c.RemoteAddr()
	if host, _, err := net.SplitHostPort(remote); err == nil {
		return host
	}
	return remote
}

// GetMethod returns the HTTP method (e.g. GET/POST). | GetMethod 返回请求方法（如 GET/POST）。
func (i *IrisContext) GetMethod() string { return i.c.Method() }

// GetPath returns the request path, equivalent to c.Request().URL.Path via Iris API.
// GetPath 返回请求路径，等价于 c.Request().URL.Path，但走 Iris 接口。
func (i *IrisContext) GetPath() string { return i.c.Path() }

// Set stores a per-request key/value for downstream handlers/middlewares.
// Set 写入请求级上下文键值，供后续 handler/中间件读取。
func (i *IrisContext) Set(key string, value interface{}) { i.c.Values().Set(key, value) }

// Get reads a per-request key/value; returns (nil, false) when absent.
// Get 读取请求级上下文键值，键不存在返回 (nil, false)。
func (i *IrisContext) Get(key string) (interface{}, bool) {
	v := i.c.Values().Get(key)
	return v, v != nil
}

// GetHeaders returns a shallow copy of all request headers to prevent external mutation.
// GetHeaders 返回所有请求头的浅拷贝，防止外部修改原始 header map。
func (i *IrisContext) GetHeaders() map[string][]string {
	src := i.c.Request().Header
	out := make(map[string][]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

// GetQueryAll returns all URL query parameters. | GetQueryAll 返回所有 URL Query 参数。
func (i *IrisContext) GetQueryAll() map[string][]string {
	src := i.c.Request().URL.Query()
	out := make(map[string][]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

// GetPostForm returns an application/x-www-form-urlencoded form field.
// GetPostForm 读取 application/x-www-form-urlencoded 表单字段。
func (i *IrisContext) GetPostForm(key string) string { return i.c.PostValue(key) }

// GetBody returns the raw request body. Iris caches it internally, so repeated calls are safe.
// GetBody 读取请求体原始字节；Iris 内部已做缓存，重复调用安全。
func (i *IrisContext) GetBody() ([]byte, error) { return i.c.GetBody() }

// GetURL returns the full request URL (including query) as a string.
// GetURL 返回完整请求 URL（含 Query）字符串。
func (i *IrisContext) GetURL() string { return i.c.Request().URL.String() }

// GetUserAgent returns the User-Agent header. | GetUserAgent 返回 User-Agent 头。
func (i *IrisContext) GetUserAgent() string { return i.c.GetHeader("User-Agent") }

// SetCookieWithOptions writes a cookie according to adapter.CookieOptions.
// SameSite string is recognized as Strict/Lax/None, falling back to Lax when unspecified.
// SetCookieWithOptions 按 adapter.CookieOptions 全量配置写 Cookie，
// SameSite 字符串识别 Strict/Lax/None，未指定时回退 Lax。
func (i *IrisContext) SetCookieWithOptions(options *adapter.CookieOptions) {
	sameSite := http.SameSiteLaxMode
	switch options.SameSite {
	case "Strict":
		sameSite = http.SameSiteStrictMode
	case "Lax":
		sameSite = http.SameSiteLaxMode
	case "None":
		sameSite = http.SameSiteNoneMode
	}
	i.c.SetCookie(buildCookie(
		options.Name, options.Value, options.MaxAge,
		options.Path, options.Domain,
		options.Secure, options.HttpOnly, sameSite,
	))
}

// GetString reads a context value as string; returns "" when absent or type-mismatched.
// GetString 以字符串方式读取上下文键值，键不存在或类型不符返回空串。
func (i *IrisContext) GetString(key string) string { return i.c.Values().GetString(key) }

// MustGet reads a context value; panics when absent. Callers must ensure the key is set
// by an upstream middleware.
// MustGet 读取上下文键值，不存在则 panic；调用方应确保 key 已被中间件写入。
func (i *IrisContext) MustGet(key string) any {
	v := i.c.Values().Get(key)
	if v == nil {
		panic("iris adapter: key not found: " + key)
	}
	return v
}

// Abort marks the current request as aborted and stops Iris pipeline execution.
// Abort 标记本次请求已中止，并调用 Iris 的 StopExecution 阻断后续 handler。
func (i *IrisContext) Abort() {
	i.aborted = true
	i.c.StopExecution()
}

// IsAborted reports whether Abort has been called. | IsAborted 返回当前是否已通过 Abort 中止。
func (i *IrisContext) IsAborted() bool { return i.aborted }

// buildCookie assembles an *http.Cookie in one place, centralizing the SameSite default.
// buildCookie 统一组装 *http.Cookie，集中处理 SameSite 默认值，避免散落各处。
func buildCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool, sameSite http.SameSite) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Domain:   domain,
		MaxAge:   maxAge,
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: sameSite,
	}
}
