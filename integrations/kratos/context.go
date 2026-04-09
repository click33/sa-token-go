package kratos

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/click33/sa-token-go/core/adapter"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// KratosContext adapts request context KratosContext 适配 Kratos 请求上下文
type KratosContext struct {
	ctx     context.Context
	values  map[string]any
	mu      sync.RWMutex
	aborted bool
}

// NewKratosContext creates request context adapter NewKratosContext 创建 Kratos 请求上下文适配器
func NewKratosContext(ctx context.Context) adapter.RequestContext {
	return &KratosContext{ctx: ctx}
}

// Get implements adapter.RequestContext Get 实现 adapter.RequestContext 接口
func (k *KratosContext) Get(key string) (any, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.values == nil {
		return nil, false
	}

	value, exists := k.values[key]
	return value, exists
}

// GetHeaders implements adapter.RequestContext GetHeaders 实现 adapter.RequestContext 接口
func (k *KratosContext) GetHeaders() map[string][]string {
	headers := make(map[string][]string)
	if request := k.GetRawRequest(); request != nil {
		for key, values := range request.Header {
			headers[key] = append([]string(nil), values...)
		}
		return headers
	}

	if tr, ok := transport.FromServerContext(k.ctx); ok {
		for _, key := range tr.RequestHeader().Keys() {
			headers[key] = append([]string(nil), tr.RequestHeader().Values(key)...)
		}
	}

	return headers
}

// GetHeader implements adapter.RequestContext GetHeader 实现 adapter.RequestContext 接口
func (k *KratosContext) GetHeader(key string) string {
	if tr, ok := transport.FromServerContext(k.ctx); ok {
		return tr.RequestHeader().Get(key)
	}
	return ""
}

// GetQuery implements adapter.RequestContext GetQuery 实现 adapter.RequestContext 接口
func (k *KratosContext) GetQuery(key string) string {
	if request := k.GetRawRequest(); request != nil {
		return request.URL.Query().Get(key)
	}
	return ""
}

// GetQueryAll implements adapter.RequestContext GetQueryAll 实现 adapter.RequestContext 接口
func (k *KratosContext) GetQueryAll() map[string][]string {
	query := make(map[string][]string)
	if request := k.GetRawRequest(); request != nil {
		for key, values := range request.URL.Query() {
			query[key] = append([]string(nil), values...)
		}
	}
	return query
}

// GetPostForm implements adapter.RequestContext GetPostForm 实现 adapter.RequestContext 接口
func (k *KratosContext) GetPostForm(key string) string {
	request := k.GetRawRequest()
	if request == nil {
		return ""
	}

	if err := request.ParseForm(); err != nil {
		return ""
	}

	return request.PostFormValue(key)
}

// GetCookie implements adapter.RequestContext GetCookie 实现 adapter.RequestContext 接口
func (k *KratosContext) GetCookie(key string) string {
	request := k.GetRawRequest()
	if request == nil {
		return ""
	}

	cookie, err := request.Cookie(key)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// GetBody implements adapter.RequestContext GetBody 实现 adapter.RequestContext 接口
func (k *KratosContext) GetBody() ([]byte, error) {
	request := k.GetRawRequest()
	if request == nil || request.Body == nil {
		return nil, nil
	}

	// Reset request body after reading 重置请求体避免后续读取失败
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}

	_ = request.Body.Close()
	request.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

// GetClientIP implements adapter.RequestContext GetClientIP 实现 adapter.RequestContext 接口
func (k *KratosContext) GetClientIP() string {
	request := k.GetRawRequest()
	if request == nil {
		return ""
	}

	// Prefer forwarded headers first 优先读取代理转发头
	if forwarded := strings.TrimSpace(request.Header.Get("X-Forwarded-For")); forwarded != "" {
		if index := strings.Index(forwarded, ","); index >= 0 {
			return strings.TrimSpace(forwarded[:index])
		}
		return forwarded
	}

	if realIP := strings.TrimSpace(request.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(request.RemoteAddr))
	if err == nil {
		return host
	}

	return strings.TrimSpace(request.RemoteAddr)
}

// GetMethod implements adapter.RequestContext GetMethod 实现 adapter.RequestContext 接口
func (k *KratosContext) GetMethod() string {
	if request := k.GetRawRequest(); request != nil {
		return request.Method
	}
	return ""
}

// GetPath implements adapter.RequestContext GetPath 实现 adapter.RequestContext 接口
func (k *KratosContext) GetPath() string {
	if request := k.GetRawRequest(); request != nil && request.URL != nil {
		return request.URL.Path
	}

	if tr, ok := transport.FromServerContext(k.ctx); ok {
		return tr.Operation()
	}
	return ""
}

// GetURL implements adapter.RequestContext GetURL 实现 adapter.RequestContext 接口
func (k *KratosContext) GetURL() string {
	if request := k.GetRawRequest(); request != nil && request.URL != nil {
		return request.URL.String()
	}
	return ""
}

// GetUserAgent implements adapter.RequestContext GetUserAgent 实现 adapter.RequestContext 接口
func (k *KratosContext) GetUserAgent() string {
	return k.GetHeader("User-Agent")
}

// IsTLS implements adapter.RequestContext IsTLS 实现 adapter.RequestContext 接口
func (k *KratosContext) IsTLS() bool {
	if request := k.GetRawRequest(); request != nil {
		return request.TLS != nil
	}
	return false
}

// SetStatusCode implements adapter.RequestContext SetStatusCode 实现 adapter.RequestContext 接口
func (k *KratosContext) SetStatusCode(code int) {
	if writer := k.GetRawResponseWriter(); writer != nil {
		writer.WriteHeader(code)
	}
}

// SetHeader implements adapter.RequestContext SetHeader 实现 adapter.RequestContext 接口
func (k *KratosContext) SetHeader(key, value string) {
	if tr, ok := transport.FromServerContext(k.ctx); ok {
		tr.ReplyHeader().Set(key, value)
	}
}

// Write implements adapter.RequestContext Write 实现 adapter.RequestContext 接口
func (k *KratosContext) Write(data []byte) (int, error) {
	writer := k.GetRawResponseWriter()
	if writer == nil {
		return 0, http.ErrNotSupported
	}
	return writer.Write(data)
}

// SetCookie implements adapter.RequestContext SetCookie 实现 adapter.RequestContext 接口
func (k *KratosContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	khttp.SetCookie(k.ctx, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: http.SameSiteLaxMode,
	})
}

// SetCookieWithOptions implements adapter.RequestContext SetCookieWithOptions 实现 adapter.RequestContext 接口
func (k *KratosContext) SetCookieWithOptions(options *adapter.CookieOptions) {
	cookie := &http.Cookie{
		Name:     options.Name,
		Value:    options.Value,
		MaxAge:   options.MaxAge,
		Path:     options.Path,
		Domain:   options.Domain,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
		SameSite: http.SameSiteLaxMode,
	}

	switch options.SameSite {
	case "Strict":
		cookie.SameSite = http.SameSiteStrictMode
	case "None":
		cookie.SameSite = http.SameSiteNoneMode
	}

	khttp.SetCookie(k.ctx, cookie)
}

// Set implements adapter.RequestContext Set 实现 adapter.RequestContext 接口
func (k *KratosContext) Set(key string, value any) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.values == nil {
		k.values = make(map[string]any)
	}

	k.values[key] = value
}

// GetString implements adapter.RequestContext GetString 实现 adapter.RequestContext 接口
func (k *KratosContext) GetString(key string) string {
	value, exists := k.Get(key)
	if !exists {
		return ""
	}

	str, ok := value.(string)
	if !ok {
		return ""
	}
	return str
}

// MustGet implements adapter.RequestContext MustGet 实现 adapter.RequestContext 接口
func (k *KratosContext) MustGet(key string) any {
	value, exists := k.Get(key)
	if !exists {
		panic("key not found: " + key)
	}
	return value
}

// Abort implements adapter.RequestContext Abort 实现 adapter.RequestContext 接口
func (k *KratosContext) Abort() {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.aborted = true
}

// IsAborted implements adapter.RequestContext IsAborted 实现 adapter.RequestContext 接口
func (k *KratosContext) IsAborted() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.aborted
}

// JSON implements adapter.RequestContextExt JSON 实现 adapter.RequestContextExt 接口
func (k *KratosContext) JSON(code int, value any) error {
	writer := k.GetRawResponseWriter()
	if writer == nil {
		return http.ErrNotSupported
	}

	k.SetHeader("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(code)
	return json.NewEncoder(writer).Encode(value)
}

// GetRawRequest implements adapter.RequestContextExt GetRawRequest 实现 adapter.RequestContextExt 接口
func (k *KratosContext) GetRawRequest() *http.Request {
	request, ok := khttp.RequestFromServerContext(k.ctx)
	if !ok {
		return nil
	}
	return request
}

// GetRawResponseWriter implements adapter.RequestContextExt GetRawResponseWriter 实现 adapter.RequestContextExt 接口
func (k *KratosContext) GetRawResponseWriter() http.ResponseWriter {
	writer, ok := khttp.ResponseWriterFromServerContext(k.ctx)
	if !ok {
		return nil
	}
	return writer
}
