package adapter

// RequestContext defines request context interface for abstracting different web frameworks | 定义请求上下文接口，用于抽象不同Web框架的请求/响应
type RequestContext interface {
	// GetHeader gets request header | 获取请求头
	GetHeader(key string) string

	// GetQuery gets query parameter | 获取查询参数
	GetQuery(key string) string

	// GetCookie gets cookie | 获取Cookie
	GetCookie(key string) string

	// SetHeader sets response header | 设置响应头
	SetHeader(key, value string)

	// SetCookie sets cookie | 设置Cookie
	SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool)

	// GetClientIP gets client IP address | 获取客户端IP地址
	GetClientIP() string

	// GetMethod gets request method | 获取请求方法
	GetMethod() string

	// GetPath gets request path | 获取请求路径
	GetPath() string

	// Set sets context value | 设置上下文值
	Set(key string, value interface{})

	// Get gets context value | 获取上下文值
	Get(key string) (interface{}, bool)
}
