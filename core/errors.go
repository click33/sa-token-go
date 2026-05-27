package core

import (
	"errors"
	"fmt"

	"github.com/sa-tokens/sa-token-go/core/errs"
)

// Re-export sentinels from errs (manager/oauth2/sso import errs; callers may use core.*) | 从 errs 再导出
var (
	ErrNotLogin       = errs.ErrNotLogin
	ErrTokenInvalid   = errs.ErrTokenInvalid
	ErrTokenExpired   = errs.ErrTokenExpired
	ErrInvalidLoginID = errs.ErrInvalidLoginID
	ErrInvalidDevice  = errs.ErrInvalidDevice

	ErrPermissionDenied = errs.ErrPermissionDenied
	ErrRoleDenied       = errs.ErrRoleDenied

	ErrAccountDisabled = errs.ErrAccountDisabled
	ErrAccountNotFound = errs.ErrAccountNotFound

	ErrSessionNotFound    = errs.ErrSessionNotFound
	ErrInvalidSessionData = errs.ErrInvalidSessionData
	ErrKickedOut          = errs.ErrKickedOut
	ErrTokenReplaced      = errs.ErrTokenReplaced
	ErrActiveTimeout      = errs.ErrActiveTimeout
	ErrMaxLoginCount      = errs.ErrMaxLoginCount

	ErrTokenSessionTokenEmpty   = errs.ErrTokenSessionTokenEmpty
	ErrTokenSessionInvalidToken = errs.ErrTokenSessionInvalidToken
	ErrInvalidTokenData         = errs.ErrInvalidTokenData
	ErrTokenNotFound            = errs.ErrTokenNotFound
	ErrInvalidValueType         = errs.ErrInvalidValueType
	ErrFeatureNotSupported      = errs.ErrFeatureNotSupported

	ErrSafeTimeInvalid   = errs.ErrSafeTimeInvalid
	ErrNotPassedSafeAuth = errs.ErrNotPassedSafeAuth

	ErrLoginIDEmpty         = errs.ErrLoginIDEmpty
	ErrInvalidDisableLevel  = errs.ErrInvalidDisableLevel
	ErrDisableLevelExceeded = errs.ErrDisableLevelExceeded
	ErrDisableServiceEmpty  = errs.ErrDisableServiceEmpty

	ErrPathAuthRequired = errs.ErrPathAuthRequired
	ErrPathNotAllowed   = errs.ErrPathNotAllowed

	ErrOAuth2InvalidClientID       = errs.ErrOAuth2InvalidClientID
	ErrOAuth2InvalidClientSecret   = errs.ErrOAuth2InvalidClientSecret
	ErrOAuth2InvalidRedirectURI    = errs.ErrOAuth2InvalidRedirectURI
	ErrOAuth2RedirectURIContainsAt = errs.ErrOAuth2RedirectURIContainsAt
	ErrOAuth2IllegalRedirectURI    = errs.ErrOAuth2IllegalRedirectURI
	ErrOAuth2ScopeNotContracted    = errs.ErrOAuth2ScopeNotContracted
	ErrOAuth2InvalidGrantType      = errs.ErrOAuth2InvalidGrantType
	ErrOAuth2InvalidResponseType   = errs.ErrOAuth2InvalidResponseType
	ErrOAuth2InvalidCode           = errs.ErrOAuth2InvalidCode
	ErrOAuth2CodeUsed              = errs.ErrOAuth2CodeUsed
	ErrOAuth2CodeExpired           = errs.ErrOAuth2CodeExpired
	ErrOAuth2RedirectMismatch      = errs.ErrOAuth2RedirectMismatch
	ErrOAuth2ClientMismatch        = errs.ErrOAuth2ClientMismatch
	ErrOAuth2InvalidAccessToken    = errs.ErrOAuth2InvalidAccessToken
	ErrOAuth2InvalidRefreshToken   = errs.ErrOAuth2InvalidRefreshToken
	ErrOAuth2InvalidUserCredential = errs.ErrOAuth2InvalidUserCredential
	ErrOAuth2GrantTypeNotAllowed   = errs.ErrOAuth2GrantTypeNotAllowed
	ErrOAuth2RequiredParamMissing  = errs.ErrOAuth2RequiredParamMissing

	ErrSsoInvalidTicket        = errs.ErrSsoInvalidTicket
	ErrSsoTicketClientMismatch = errs.ErrSsoTicketClientMismatch
	ErrSsoInvalidRedirect      = errs.ErrSsoInvalidRedirect
	ErrSsoCallbackFailed       = errs.ErrSsoCallbackFailed
	ErrSsoSignatureInvalid     = errs.ErrSsoSignatureInvalid
	ErrSsoServerUnreachable    = errs.ErrSsoServerUnreachable
	ErrSsoClientNotRegistered  = errs.ErrSsoClientNotRegistered

	ErrStorageUnavailable = errs.ErrStorageUnavailable
	ErrInvalidConfig      = errs.ErrInvalidConfig

	ErrSameTokenInvalid = errs.ErrSameTokenInvalid

	ErrApiKeyNotFound  = errs.ErrApiKeyNotFound
	ErrApiKeyExpired   = errs.ErrApiKeyExpired
	ErrApiKeyDisabled  = errs.ErrApiKeyDisabled

	ErrSignatureInvalid = errs.ErrSignatureInvalid
	ErrSignatureExpired = errs.ErrSignatureExpired
	ErrNonceAlreadyUsed = errs.ErrNonceAlreadyUsed

	ErrTempTokenNotFound = errs.ErrTempTokenNotFound
	ErrTempTokenExpired  = errs.ErrTempTokenExpired
	ErrTempTokenUsed     = errs.ErrTempTokenUsed
)

// SaTokenError custom error with code and context | 自定义错误类型
type SaTokenError struct {
	Code    int
	Message string
	Err     error
	Context map[string]any
}

// Error implements error | 实现 error
func (e *SaTokenError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s (code: %d): %v", e.Message, e.Code, e.Err)
	}
	return fmt.Sprintf("%s (code: %d)", e.Message, e.Code)
}

// Unwrap unwraps underlying error | 解包底层错误
func (e *SaTokenError) Unwrap() error {
	return e.Err
}

// WithContext attaches context | 附加上下文
func (e *SaTokenError) WithContext(key string, value any) *SaTokenError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// GetContext gets context value | 获取上下文
func (e *SaTokenError) GetContext(key string) (any, bool) {
	if e.Context == nil {
		return nil, false
	}
	val, exists := e.Context[key]
	return val, exists
}

// Is implements errors.Is | 实现 errors.Is
func (e *SaTokenError) Is(target error) bool {
	if t, ok := target.(*SaTokenError); ok {
		return e.Code == t.Code
	}
	return errors.Is(e.Err, target)
}

// NewError creates SaTokenError | 创建错误
func NewError(code int, message string, err error) *SaTokenError {
	return &SaTokenError{
		Code:    code,
		Message: message,
		Err:     err,
		Context: make(map[string]any),
	}
}

// NewErrorWithContext creates SaTokenError with context | 带上下文创建错误
func NewErrorWithContext(code int, message string, err error, context map[string]any) *SaTokenError {
	return &SaTokenError{
		Code:    code,
		Message: message,
		Err:     err,
		Context: context,
	}
}

// NewNotLoginError | 未登录错误
func NewNotLoginError() *SaTokenError {
	return NewError(CodeNotLogin, "user not logged in", ErrNotLogin)
}

// NewPermissionDeniedError | 权限拒绝
func NewPermissionDeniedError(permission string) *SaTokenError {
	return NewError(CodePermissionDenied, "permission denied", ErrPermissionDenied).
		WithContext("permission", permission)
}

// NewPermissionDeniedListError | 多权限拒绝
func NewPermissionDeniedListError(permissions []string) *SaTokenError {
	return NewError(CodePermissionDenied, "permission denied", ErrPermissionDenied).
		WithContext("permissions", permissions)
}

// NewRoleDeniedError | 角色拒绝
func NewRoleDeniedError(role string) *SaTokenError {
	return NewError(CodePermissionDenied, "role denied", ErrRoleDenied).
		WithContext("role", role)
}

// NewRoleDeniedListError | 多角色拒绝
func NewRoleDeniedListError(roles []string) *SaTokenError {
	return NewError(CodePermissionDenied, "role denied", ErrRoleDenied).
		WithContext("roles", roles)
}

// NewAccountDisabledError | 账号禁用
func NewAccountDisabledError(loginID string) *SaTokenError {
	return NewError(CodeAccountDisabled, "account disabled", ErrAccountDisabled).
		WithContext("loginID", loginID)
}

// NewPathAuthRequiredError | 路径需鉴权
func NewPathAuthRequiredError(path string) *SaTokenError {
	return NewError(CodePathAuthRequired, "path authentication required", ErrPathAuthRequired).
		WithContext("path", path)
}

// NewPathNotAllowedError | 路径禁止
func NewPathNotAllowedError(path string) *SaTokenError {
	return NewError(CodePathNotAllowed, "path not allowed", ErrPathNotAllowed).
		WithContext("path", path)
}

// NewTokenNotFoundError | Token 未找到
func NewTokenNotFoundError(loginID string) *SaTokenError {
	return NewError(CodeNotFound, "token not found", ErrTokenNotFound).WithContext("loginID", loginID)
}

// NewInvalidValueTypeError | 值类型异常
func NewInvalidValueTypeError(key string) *SaTokenError {
	return NewError(CodeServerError, "invalid value type", ErrInvalidValueType).WithContext("key", key)
}

// NewFeatureNotSupportedError | 功能不支持
func NewFeatureNotSupportedError(feature string) *SaTokenError {
	return NewError(CodeBadRequest, "feature not supported", ErrFeatureNotSupported).WithContext("feature", feature)
}

// NewActiveTimeoutError | 活跃超时
func NewActiveTimeoutError(token string) *SaTokenError {
	return NewError(CodeActiveTimeout, "session inactive timeout", ErrActiveTimeout).WithContext("token", token)
}

// NewKickedOutError | 踢下线
func NewKickedOutError(token string) *SaTokenError {
	return NewError(CodeKickedOut, "kicked out", ErrKickedOut).WithContext("token", token)
}

// NewTokenReplacedError | 顶号
func NewTokenReplacedError(token string) *SaTokenError {
	return NewError(CodeKickedOut, "token replaced", ErrTokenReplaced).WithContext("token", token)
}

// NewInvalidConfigError | 配置无效
func NewInvalidConfigError(field, reason string) *SaTokenError {
	return NewError(CodeBadRequest, "invalid configuration", ErrInvalidConfig).
		WithContext("field", field).WithContext("reason", reason)
}

// IsNotLoginError | 是否未登录
func IsNotLoginError(err error) bool {
	return errors.Is(err, ErrNotLogin)
}

// IsPermissionDeniedError | 是否权限拒绝
func IsPermissionDeniedError(err error) bool {
	return errors.Is(err, ErrPermissionDenied)
}

// IsAccountDisabledError | 是否账号禁用
func IsAccountDisabledError(err error) bool {
	return errors.Is(err, ErrAccountDisabled)
}

// IsTokenError | 是否 Token 类错误
func IsTokenError(err error) bool {
	return errors.Is(err, ErrTokenInvalid) || errors.Is(err, ErrTokenExpired)
}

// GetErrorCode extracts code | 提取错误码
func GetErrorCode(err error) int {
	var saErr *SaTokenError
	if errors.As(err, &saErr) {
		return saErr.Code
	}
	return CodeServerError
}

// Error codes | 错误码
const (
	CodeSuccess          = 200
	CodeBadRequest       = 400
	CodeNotLogin         = 401
	CodePermissionDenied = 403
	CodePathAuthRequired = 401
	CodePathNotAllowed   = 403
	CodeNotFound         = 404
	CodeServerError      = 500

	CodeTokenInvalid     = 10001
	CodeTokenExpired     = 10002
	CodeAccountDisabled  = 10003
	CodeKickedOut        = 10004
	CodeActiveTimeout    = 10005
	CodeMaxLoginCount    = 10006
	CodeStorageError     = 10007
	CodeInvalidParameter = 10008
	CodeSessionError     = 10009
)

// NewTokenSessionInvalidTokenError | Token-Session 无效
func NewTokenSessionInvalidTokenError(token string) *SaTokenError {
	return NewError(CodeTokenInvalid, "token-session lookup failed: token is invalid", ErrTokenSessionInvalidToken).
		WithContext("token", token)
}

// NewNotPassedSafeAuthError | 二级认证未通过
func NewNotPassedSafeAuthError(service string) *SaTokenError {
	return NewError(CodePermissionDenied, "second-level authentication required", ErrNotPassedSafeAuth).
		WithContext("service", service)
}

// NewDisableLevelExceededError | 分级封禁超限
func NewDisableLevelExceededError(loginID, service string, level int) *SaTokenError {
	return NewError(CodeAccountDisabled, "account is disabled at the required level", ErrDisableLevelExceeded).
		WithContext("loginID", loginID).
		WithContext("service", service).
		WithContext("level", level)
}

// NewOAuth2InvalidClientIDError | OAuth2 client 无效
func NewOAuth2InvalidClientIDError(clientID string) *SaTokenError {
	return NewError(CodeBadRequest, "invalid client_id", ErrOAuth2InvalidClientID).
		WithContext("client_id", clientID)
}

// NewOAuth2InvalidRedirectURIError | redirect_uri 无效
func NewOAuth2InvalidRedirectURIError(url string) *SaTokenError {
	return NewError(CodeBadRequest, "invalid redirect_uri", ErrOAuth2InvalidRedirectURI).
		WithContext("redirect_uri", url)
}

// NewOAuth2RedirectURIContainsAtError | redirect_uri 含 @
func NewOAuth2RedirectURIContainsAtError(url string) *SaTokenError {
	return NewError(CodeBadRequest, "redirect_uri must not contain '@' character", ErrOAuth2RedirectURIContainsAt).
		WithContext("redirect_uri", url)
}

// NewOAuth2IllegalRedirectURIError | redirect_uri 不在白名单
func NewOAuth2IllegalRedirectURIError(url string) *SaTokenError {
	return NewError(CodeBadRequest, "illegal redirect_uri: not in allow list", ErrOAuth2IllegalRedirectURI).
		WithContext("redirect_uri", url)
}

// NewOAuth2ScopeNotContractedError | scope 未签约
func NewOAuth2ScopeNotContractedError(clientID, scope string) *SaTokenError {
	return NewError(CodeBadRequest, "client has not contracted the requested scope", ErrOAuth2ScopeNotContracted).
		WithContext("client_id", clientID).
		WithContext("scope", scope)
}

// NewSsoTicketClientMismatchError | ticket client 不匹配
func NewSsoTicketClientMismatchError(ticket, expectedClient, actualClient string) *SaTokenError {
	return NewError(CodeBadRequest, "ticket does not belong to the specified client", ErrSsoTicketClientMismatch).
		WithContext("ticket", ticket).
		WithContext("expected_client", expectedClient).
		WithContext("actual_client", actualClient)
}

// NewOAuth2InvalidCodeError | 无效授权码
func NewOAuth2InvalidCodeError(code string) *SaTokenError {
	return NewError(CodeBadRequest, "invalid authorization code", ErrOAuth2InvalidCode).WithContext("code", code)
}

// NewOAuth2CodeExpiredError | 授权码过期
func NewOAuth2CodeExpiredError(code string) *SaTokenError {
	return NewError(CodeBadRequest, "authorization code expired", ErrOAuth2CodeExpired).WithContext("code", code)
}

// NewOAuth2RedirectMismatchError | redirect 不匹配
func NewOAuth2RedirectMismatchError(expect, actual string) *SaTokenError {
	return NewError(CodeBadRequest, "redirect_uri mismatch", ErrOAuth2RedirectMismatch).
		WithContext("expect", expect).WithContext("actual", actual)
}

// NewOAuth2ClientMismatchError | client 不匹配
func NewOAuth2ClientMismatchError(expect, actual string) *SaTokenError {
	return NewError(CodeBadRequest, "client mismatch", ErrOAuth2ClientMismatch).
		WithContext("expect", expect).WithContext("actual", actual)
}

// NewOAuth2InvalidAccessTokenError | access_token 无效
func NewOAuth2InvalidAccessTokenError(token string) *SaTokenError {
	return NewError(CodeTokenInvalid, "invalid access_token", ErrOAuth2InvalidAccessToken).WithContext("access_token", token)
}

// NewOAuth2InvalidRefreshTokenError | refresh_token 无效
func NewOAuth2InvalidRefreshTokenError(token string) *SaTokenError {
	return NewError(CodeBadRequest, "invalid refresh_token", ErrOAuth2InvalidRefreshToken).WithContext("refresh_token", token)
}

// NewOAuth2RequiredParamError | 缺少参数
func NewOAuth2RequiredParamError(param string) *SaTokenError {
	return NewError(CodeBadRequest, "required parameter missing", ErrOAuth2RequiredParamMissing).WithContext("param", param)
}

// NewOAuth2GrantTypeNotAllowedError | grant 不允许
func NewOAuth2GrantTypeNotAllowedError(clientID, grantType string) *SaTokenError {
	return NewError(CodeBadRequest, "grant_type not allowed for this client", ErrOAuth2GrantTypeNotAllowed).
		WithContext("client_id", clientID).WithContext("grant_type", grantType)
}

// NewSsoSignatureInvalidError | SSO 签名无效
func NewSsoSignatureInvalidError(client string) *SaTokenError {
	return NewError(CodeBadRequest, "sso signature verification failed", ErrSsoSignatureInvalid).WithContext("client", client)
}

// NewSsoCallbackFailedError | SSO 回调失败
func NewSsoCallbackFailedError(client, url string) *SaTokenError {
	return NewError(CodeServerError, "sso single-logout callback failed", ErrSsoCallbackFailed).
		WithContext("client", client).WithContext("url", url)
}
