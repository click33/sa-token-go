package core

import "fmt"

// Common error definitions for better error handling and internationalization support
// 常见错误定义，用于更好的错误处理和国际化支持

var (
	// ErrNotLogin indicates the user is not logged in | 用户未登录错误
	ErrNotLogin = fmt.Errorf("authentication required: user not logged in")

	// ErrTokenInvalid indicates the provided token is invalid or malformed | Token无效或格式错误
	ErrTokenInvalid = fmt.Errorf("invalid token: the token is malformed or corrupted")

	// ErrTokenExpired indicates the token has expired | Token已过期
	ErrTokenExpired = fmt.Errorf("token expired: please login again to get a new token")

	// ErrAccountDisabled indicates the account has been disabled or banned | 账号已被禁用
	ErrAccountDisabled = fmt.Errorf("account disabled: this account has been temporarily or permanently disabled")

	// ErrPermissionDenied indicates insufficient permissions | 权限不足
	ErrPermissionDenied = fmt.Errorf("permission denied: you don't have the required permission")

	// ErrRoleDenied indicates insufficient role | 角色权限不足
	ErrRoleDenied = fmt.Errorf("role denied: you don't have the required role")

	// ErrSessionNotFound indicates the session doesn't exist | Session不存在
	ErrSessionNotFound = fmt.Errorf("session not found: the session may have expired or been deleted")

	// ErrAccountNotFound indicates the account doesn't exist | 账号不存在
	ErrAccountNotFound = fmt.Errorf("account not found: no account associated with this identifier")

	// ErrKickedOut indicates the user has been kicked out | 用户已被踢下线
	ErrKickedOut = fmt.Errorf("kicked out: this session has been forcibly terminated")

	// ErrActiveTimeout indicates the session has been inactive for too long | Session活跃超时
	ErrActiveTimeout = fmt.Errorf("session inactive: the session has exceeded the inactivity timeout")

	// ErrMaxLoginCount indicates maximum concurrent login limit reached | 达到最大登录数量限制
	ErrMaxLoginCount = fmt.Errorf("max login limit: maximum number of concurrent logins reached")

	// ErrStorageUnavailable indicates the storage backend is unavailable | 存储后端不可用
	ErrStorageUnavailable = fmt.Errorf("storage unavailable: unable to connect to storage backend")

	// ErrInvalidLoginID indicates the login ID is invalid | 登录ID无效
	ErrInvalidLoginID = fmt.Errorf("invalid login ID: the login identifier cannot be empty")

	// ErrInvalidDevice indicates the device identifier is invalid | 设备标识无效
	ErrInvalidDevice = fmt.Errorf("invalid device: the device identifier is not valid")
)

// SaTokenError represents a custom error with error code and context | 自定义错误类型，包含错误码和上下文信息
type SaTokenError struct {
	Code    int                    // Error code for programmatic handling | 错误码，用于程序化处理
	Message string                 // Human-readable error message | 可读的错误消息
	Err     error                  // Underlying error (if any) | 底层错误（如果有）
	Context map[string]interface{} // Additional context information | 额外的上下文信息
}

// Error implements the error interface | 实现 error 接口
func (e *SaTokenError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s (code: %d): %v", e.Message, e.Code, e.Err)
	}
	return fmt.Sprintf("%s (code: %d)", e.Message, e.Code)
}

// Unwrap implements the unwrap interface for error chains | 实现 unwrap 接口，支持错误链
func (e *SaTokenError) Unwrap() error {
	return e.Err
}

// WithContext adds context information to the error | 为错误添加上下文信息
func (e *SaTokenError) WithContext(key string, value interface{}) *SaTokenError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// NewError creates a new Sa-Token error | 创建新的 Sa-Token 错误
func NewError(code int, message string, err error) *SaTokenError {
	return &SaTokenError{
		Code:    code,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// NewErrorWithContext creates a new Sa-Token error with context | 创建带上下文的 Sa-Token 错误
func NewErrorWithContext(code int, message string, err error, context map[string]interface{}) *SaTokenError {
	return &SaTokenError{
		Code:    code,
		Message: message,
		Err:     err,
		Context: context,
	}
}

// Error code definitions | 错误码定义
const (
	// Standard HTTP status codes | 标准 HTTP 状态码
	CodeSuccess          = 200 // Request successful | 请求成功
	CodeBadRequest       = 400 // Bad request | 错误的请求
	CodeNotLogin         = 401 // Not authenticated | 未认证
	CodePermissionDenied = 403 // Permission denied | 权限不足
	CodeNotFound         = 404 // Resource not found | 资源未找到
	CodeServerError      = 500 // Internal server error | 服务器内部错误

	// Sa-Token specific error codes (10000-19999) | Sa-Token 特定错误码 (10000-19999)
	CodeTokenInvalid     = 10001 // Token is invalid or malformed | Token无效或格式错误
	CodeTokenExpired     = 10002 // Token has expired | Token已过期
	CodeAccountDisabled  = 10003 // Account is disabled | 账号已被禁用
	CodeKickedOut        = 10004 // User has been kicked out | 用户已被踢下线
	CodeActiveTimeout    = 10005 // Session inactive timeout | Session活跃超时
	CodeMaxLoginCount    = 10006 // Maximum login count reached | 达到最大登录数量
	CodeStorageError     = 10007 // Storage backend error | 存储后端错误
	CodeInvalidParameter = 10008 // Invalid parameter | 无效参数
	CodeSessionError     = 10009 // Session operation error | Session操作错误
)
