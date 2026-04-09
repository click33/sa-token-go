package serror

import (
	"errors"
	"fmt"
)

const (
	// CodeSuccess indicates success CodeSuccess 表示操作成功
	CodeSuccess = 0
	// CodeBadRequest indicates bad request CodeBadRequest 表示请求参数错误
	CodeBadRequest = 400
	// CodeNotLogin indicates not login CodeNotLogin 表示用户未登录
	CodeNotLogin = 401
	// CodePermissionDenied indicates permission denied CodePermissionDenied 表示权限不足
	CodePermissionDenied = 403
	// CodeNotFound indicates resource not found CodeNotFound 表示资源未找到
	CodeNotFound = 404
	// CodeServerError indicates server error CodeServerError 表示服务器内部错误
	CodeServerError = 500
	// CodeTokenInvalid indicates invalid token CodeTokenInvalid 表示 Token 无效
	CodeTokenInvalid = 10001
	// CodeTokenExpired indicates expired token CodeTokenExpired 表示 Token 已过期
	CodeTokenExpired = 10002
	// CodeAccountDisabled indicates disabled account CodeAccountDisabled 表示账号已被封禁
	CodeAccountDisabled = 10003
	// CodeKickedOut indicates kicked out user CodeKickedOut 表示用户已被踢下线
	CodeKickedOut = 10004
	// CodeActiveTimeout indicates active timeout CodeActiveTimeout 表示活跃超时
	CodeActiveTimeout = 10005
	// CodeMaxLoginCount indicates max login count exceeded CodeMaxLoginCount 表示超出最大登录数量
	CodeMaxLoginCount = 10006
	// CodeStorageError indicates storage error CodeStorageError 表示存储错误
	CodeStorageError = 10007
	// CodeInvalidParameter indicates invalid parameter CodeInvalidParameter 表示参数无效
	CodeInvalidParameter = 10008
)

// SaTokenError represents sa-token error SaTokenError 表示带有错误码和消息的 Sa-Token 错误
type SaTokenError struct {
	Code    int
	Message string
	Err     error
}

// Error returns error string Error 返回错误字符串
func (e *SaTokenError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap returns wrapped error Unwrap 返回包装的错误
func (e *SaTokenError) Unwrap() error {
	return e.Err
}

// NewSaTokenError creates sa-token error NewSaTokenError 创建新的 SaTokenError
func NewSaTokenError(code int, message string, err error) *SaTokenError {
	return &SaTokenError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

var (
	// ErrStorageUnavailable indicates storage unavailable ErrStorageUnavailable 表示存储后端不可用
	ErrStorageUnavailable = errors.New("storage unavailable: unable to connect to storage backend")
	// ErrSerializeFailed indicates serialize failed ErrSerializeFailed 表示序列化失败
	ErrSerializeFailed = errors.New("serialize failed: unable to encode data")
	// ErrTypeConvert indicates type convert failed ErrTypeConvert 表示类型转换失败
	ErrTypeConvert = errors.New("type conversion failed: unable to convert value to target type")
	// ErrManagerNotFound indicates manager not found ErrManagerNotFound 表示未找到指定认证类型的管理器
	ErrManagerNotFound = errors.New("manager not found")
	// ErrManagerInvalidType indicates invalid manager type ErrManagerInvalidType 表示全局存储中的管理器类型不正确
	ErrManagerInvalidType = errors.New("manager has invalid type")
	// ErrInvalidParam indicates invalid param ErrInvalidParam 表示参数无效
	ErrInvalidParam = errors.New("invalid parameter")
)

var (
	// ErrIDIsEmpty indicates empty id ErrIDIsEmpty 表示 ID 为空
	ErrIDIsEmpty = errors.New("ID is required and cannot be empty")
	// ErrAccountDisabled indicates disabled account ErrAccountDisabled 表示账号已被禁用
	ErrAccountDisabled = errors.New("account disabled: this account has been temporarily or permanently disabled")
	// ErrAccountNotDisabled indicates account not disabled ErrAccountNotDisabled 表示账号未被封禁
	ErrAccountNotDisabled = errors.New("account not disabled: this account is not currently disabled")
	// ErrLoginLimitExceeded indicates login limit exceeded ErrLoginLimitExceeded 表示超出最大登录数量限制
	ErrLoginLimitExceeded = errors.New("account error: login count exceeds the maximum limit")
)

var (
	// ErrNotLogin indicates not login ErrNotLogin 表示用户未登录
	ErrNotLogin = errors.New("authentication required: user is not logged in")
	// ErrInvalidToken indicates invalid token ErrInvalidToken 表示令牌无效或格式错误
	ErrInvalidToken = errors.New("invalid or malformed authentication token")
	// ErrTokenExpired indicates token expired ErrTokenExpired 表示令牌已过期
	ErrTokenExpired = errors.New("authentication required: token has expired")
	// ErrTokenKickout indicates token kicked out ErrTokenKickout 表示 Token 已被踢下线
	ErrTokenKickout = errors.New("authentication required: token has been kicked out")
	// ErrTokenReplaced indicates token replaced ErrTokenReplaced 表示 Token 已被顶下线
	ErrTokenReplaced = errors.New("authentication required: token has been replaced")
	// ErrInvalidDevice indicates invalid device ErrInvalidDevice 表示设备无效
	ErrInvalidDevice = errors.New("invalid device: device information is invalid or not recognized")
)

var (
	// ErrPermissionDenied indicates permission denied ErrPermissionDenied 表示权限不足
	ErrPermissionDenied = errors.New("permission denied: insufficient permissions to perform this action")
	// ErrRoleDenied indicates role denied ErrRoleDenied 表示角色不足
	ErrRoleDenied = errors.New("role denied: user does not have the required role")
)

var (
	// ErrServiceDisabled indicates disabled service ErrServiceDisabled 表示账号的指定服务已被封禁
	ErrServiceDisabled = errors.New("service disabled: the specified service is disabled for this account")
	// ErrServiceNotDisabled indicates service not disabled ErrServiceNotDisabled 表示该服务未被封禁
	ErrServiceNotDisabled = errors.New("service not disabled")
	// ErrDisableLevelNotReached indicates disable level not reached ErrDisableLevelNotReached 表示未达到指定封禁等级
	ErrDisableLevelNotReached = errors.New("disable level not reached")
)

var (
	// ErrClientOrClientIDEmpty indicates empty client ErrClientOrClientIDEmpty 表示客户端或客户端ID为空
	ErrClientOrClientIDEmpty = errors.New("client or client ID cannot be empty")
	// ErrClientNotFound indicates client not found ErrClientNotFound 表示客户端未找到
	ErrClientNotFound = errors.New("client not found")
	// ErrInvalidClientCredentials indicates invalid client credentials ErrInvalidClientCredentials 表示客户端凭证无效
	ErrInvalidClientCredentials = errors.New("invalid client credentials")
	// ErrInvalidGrantType indicates invalid grant type ErrInvalidGrantType 表示授权类型无效
	ErrInvalidGrantType = errors.New("invalid grant type")
	// ErrInvalidRedirectURI indicates invalid redirect uri ErrInvalidRedirectURI 表示回调 URI 无效
	ErrInvalidRedirectURI = errors.New("invalid redirect URI")
	// ErrInvalidScope indicates invalid scope ErrInvalidScope 表示权限范围无效
	ErrInvalidScope = errors.New("invalid scope")
	// ErrUserIDEmpty indicates empty user id ErrUserIDEmpty 表示用户 ID 为空
	ErrUserIDEmpty = errors.New("user ID cannot be empty")
	// ErrInvalidAuthCode indicates invalid auth code ErrInvalidAuthCode 表示授权码无效
	ErrInvalidAuthCode = errors.New("invalid authorization code")
	// ErrAuthCodeUsed indicates used auth code ErrAuthCodeUsed 表示授权码已被使用
	ErrAuthCodeUsed = errors.New("authorization code has been used")
	// ErrAuthCodeExpired indicates expired auth code ErrAuthCodeExpired 表示授权码已过期
	ErrAuthCodeExpired = errors.New("authorization code has expired")
	// ErrClientMismatch indicates client mismatch ErrClientMismatch 表示客户端不匹配
	ErrClientMismatch = errors.New("client mismatch")
	// ErrRedirectURIMismatch indicates redirect uri mismatch ErrRedirectURIMismatch 表示回调 URI 不匹配
	ErrRedirectURIMismatch = errors.New("redirect URI mismatch")
	// ErrInvalidRefreshToken indicates invalid refresh token ErrInvalidRefreshToken 表示刷新令牌无效
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	// ErrInvalidAccessToken indicates invalid access token ErrInvalidAccessToken 表示访问令牌无效
	ErrInvalidAccessToken = errors.New("invalid access token")
	// ErrInvalidUserCredentials indicates invalid user credentials ErrInvalidUserCredentials 表示用户凭证无效
	ErrInvalidUserCredentials = errors.New("invalid user credentials")
)

var (
	// ErrSessionNotFound indicates session not found ErrSessionNotFound 表示会话不存在
	ErrSessionNotFound = errors.New("session not found")
)

var (
	// ErrInvalidNonce indicates invalid nonce ErrInvalidNonce 表示 nonce 无效或已过期
	ErrInvalidNonce = errors.New("invalid or expired nonce")
)
