package adapter

import (
	"net/http"
)

// CookieOptions defines cookie option fields CookieOptions 定义 Cookie 设置选项
type CookieOptions struct {
	// Name stores cookie name Name 存储 Cookie 名称
	Name string
	// Value stores cookie value Value 存储 Cookie 值
	Value string
	// MaxAge stores expiration in seconds MaxAge 存储过期时间秒数
	MaxAge int
	// Path stores cookie path Path 存储路径
	Path string
	// Domain stores cookie domain Domain 存储域名
	Domain string
	// Secure indicates HTTPS only Secure 表示是否仅在 HTTPS 下生效
	Secure bool
	// HttpOnly indicates JS access disabled HttpOnly 表示是否禁止 JS 访问
	HttpOnly bool
	// SameSite stores SameSite attribute SameSite 存储 SameSite 属性
	SameSite string
}

// RequestContext defines request context abstraction RequestContext 定义请求上下文抽象接口
type RequestContext interface {
	// GetHeader gets request header GetHeader 获取请求头
	GetHeader(key string) string
	// GetHeaders gets all request headers GetHeaders 获取所有请求头
	GetHeaders() map[string][]string
	// GetQuery gets query parameter GetQuery 获取查询参数
	GetQuery(key string) string
	// GetQueryAll gets all query parameters GetQueryAll 获取所有查询参数
	GetQueryAll() map[string][]string
	// GetPostForm gets POST form parameter GetPostForm 获取 POST 表单参数
	GetPostForm(key string) string
	// GetCookie gets cookie value GetCookie 获取 Cookie
	GetCookie(key string) string
	// GetBody gets raw request body GetBody 获取请求体字节数据
	GetBody() ([]byte, error)
	// GetClientIP gets client IP GetClientIP 获取客户端 IP 地址
	GetClientIP() string
	// GetMethod gets request method GetMethod 获取请求方法
	GetMethod() string
	// GetPath gets request path GetPath 获取请求路径
	GetPath() string
	// GetURL gets full request URL GetURL 获取完整请求 URL
	GetURL() string
	// GetUserAgent gets user agent GetUserAgent 获取 User-Agent
	GetUserAgent() string
	// IsTLS checks whether request uses HTTPS IsTLS 检查请求是否通过 HTTPS 发起
	IsTLS() bool

	// SetStatusCode sets HTTP status code SetStatusCode 设置 HTTP 响应状态码
	SetStatusCode(code int)
	// SetHeader sets response header SetHeader 设置响应头
	SetHeader(key, value string)
	// Write writes bytes to response body Write 直接写入响应体字节数据
	Write(data []byte) (int, error)
	// SetCookie sets cookie using legacy arguments SetCookie 使用兼容旧版参数设置 Cookie
	SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool)
	// SetCookieWithOptions sets cookie using options SetCookieWithOptions 使用选项设置 Cookie
	SetCookieWithOptions(options *CookieOptions)

	// Set stores context value Set 设置上下文值
	Set(key string, value any)
	// Get gets context value Get 获取上下文值
	Get(key string) (any, bool)
	// GetString gets string value from context GetString 从上下文获取字符串值
	GetString(key string) string
	// MustGet gets value or panics MustGet 获取上下文值且不存在时 panic
	MustGet(key string) any

	// Abort aborts request handling Abort 中止请求处理
	Abort()
	// IsAborted checks whether request aborted IsAborted 检查请求是否已中止
	IsAborted() bool
}

// RequestContextExt defines extended request context interface RequestContextExt 定义扩展请求上下文接口
type RequestContextExt interface {
	RequestContext

	// JSON writes JSON response JSON 返回 JSON 格式响应
	JSON(code int, v any) error

	// GetRawRequest gets raw http request GetRawRequest 获取原始 *http.Request 对象
	GetRawRequest() *http.Request
	// GetRawResponseWriter gets raw response writer GetRawResponseWriter 获取原始 http.ResponseWriter 对象
	GetRawResponseWriter() http.ResponseWriter
}
