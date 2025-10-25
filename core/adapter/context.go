package adapter

// CookieOptions Cookie setting options | Cookie设置选项
type CookieOptions struct {
	// Name Cookie name | Cookie名称
	Name string
	// Value Cookie value | Cookie值
	Value string
	// MaxAge Cookie expiration time in seconds, 0 means delete cookie, -1 means session cookie | 过期时间（秒），0表示删除cookie，-1表示会话cookie
	MaxAge int
	// Path Cookie path | 路径
	Path string
	// Domain Cookie domain | 域名
	Domain string
	// Secure Only effective under HTTPS | 是否只在HTTPS下生效
	Secure bool
	// HttpOnly Prevent JavaScript access | 是否禁止JS访问
	HttpOnly bool
	// SameSite SameSite attribute (Strict, Lax, None) | SameSite属性
	SameSite string
}

// RequestContext defines request context interface for abstracting different web frameworks | 定义请求上下文接口，用于抽象不同Web框架的请求/响应
type RequestContext interface {
	// ============== Request Methods | 请求方法 ==============

	// GetHeader gets request header | 获取请求头
	GetHeader(key string) string

	// GetHeaders gets all request headers | 获取所有请求头
	GetHeaders() map[string][]string

	// GetQuery gets query parameter | 获取查询参数
	GetQuery(key string) string

	// GetQueryAll gets all query parameters | 获取所有查询参数
	GetQueryAll() map[string][]string

	// GetPostForm gets POST form parameter | 获取POST表单参数
	GetPostForm(key string) string

	// GetCookie gets cookie | 获取Cookie
	GetCookie(key string) string

	// GetBody gets request body as bytes | 获取请求体字节数据
	GetBody() ([]byte, error)

	// GetClientIP gets client IP address | 获取客户端IP地址
	GetClientIP() string

	// GetMethod gets request method (GET, POST, etc.) | 获取请求方法（GET、POST等）
	GetMethod() string

	// GetPath gets request path | 获取请求路径
	GetPath() string

	// GetURL gets full request URL | 获取完整请求URL
	GetURL() string

	// GetUserAgent gets User-Agent header | 获取User-Agent
	GetUserAgent() string

	// ============== Response Methods | 响应方法 ==============

	// SetHeader sets response header | 设置响应头
	SetHeader(key, value string)

	// SetCookie sets cookie (legacy method for backward compatibility) | 设置Cookie（兼容旧版本的方法）
	SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool)

	// SetCookieWithOptions sets cookie with options | 使用选项设置Cookie
	SetCookieWithOptions(options *CookieOptions)

	// ============== Context Storage Methods | 上下文存储方法 ==============

	// Set sets context value | 设置上下文值
	Set(key string, value any)

	// Get gets context value | 获取上下文值
	Get(key string) (any, bool)

	// GetString gets string value from context | 从上下文获取字符串值
	GetString(key string) string

	// MustGet gets context value, panics if not exists | 获取上下文值，不存在则panic
	MustGet(key string) any

	// ============== Utility Methods | 工具方法 ==============

	// Abort aborts the request processing | 中止请求处理
	Abort()

	// IsAborted checks if the request is aborted | 检查请求是否已中止
	IsAborted() bool
}
