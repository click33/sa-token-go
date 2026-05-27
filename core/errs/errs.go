// Package errs holds sentinel errors for subpackages that must not import the root core package
// (avoids import cycle: core/satoken → manager → core).
// Package errs 存放哨兵错误，供不能导入根 core 的子包使用（避免 core/satoken → manager → core 循环依赖）
package errs

import "fmt"

// Authentication | 认证
var (
	ErrNotLogin       = fmt.Errorf("authentication required: user not logged in")
	ErrTokenInvalid   = fmt.Errorf("invalid token: the token is malformed or corrupted")
	ErrTokenExpired   = fmt.Errorf("token expired: please login again to get a new token")
	ErrInvalidLoginID = fmt.Errorf("invalid login ID: the login identifier cannot be empty")
	ErrInvalidDevice  = fmt.Errorf("invalid device: the device identifier is not valid")
)

// Authorization | 授权
var (
	ErrPermissionDenied = fmt.Errorf("permission denied: you don't have the required permission")
	ErrRoleDenied       = fmt.Errorf("role denied: you don't have the required role")
)

// Account | 账号
var (
	ErrAccountDisabled = fmt.Errorf("account disabled: this account has been temporarily or permanently disabled")
	ErrAccountNotFound = fmt.Errorf("account not found: no account associated with this identifier")
)

// Session / token state | 会话与 Token
var (
	ErrSessionNotFound    = fmt.Errorf("session not found: the session may have expired or been deleted")
	ErrInvalidSessionData = fmt.Errorf("invalid session data: stored payload is malformed")
	ErrKickedOut          = fmt.Errorf("kicked out: this session has been forcibly terminated")
	ErrTokenReplaced      = fmt.Errorf("token replaced: this session has been overrun by a newer login")
	ErrActiveTimeout      = fmt.Errorf("session inactive: the session has exceeded the inactivity timeout")
	ErrMaxLoginCount      = fmt.Errorf("max login limit: maximum number of concurrent logins reached")

	ErrTokenSessionTokenEmpty   = fmt.Errorf("token-session lookup failed: token is empty")
	ErrTokenSessionInvalidToken = fmt.Errorf("token-session lookup failed: token is invalid")

	ErrInvalidTokenData    = fmt.Errorf("invalid token data: stored payload is malformed")
	ErrTokenNotFound       = fmt.Errorf("token not found: no token associated with this account/device")
	ErrInvalidValueType    = fmt.Errorf("invalid value type: storage returned unexpected type")
	ErrFeatureNotSupported = fmt.Errorf("feature not supported in current version")
)

// Path | 路径
var (
	ErrPathAuthRequired = fmt.Errorf("path authentication required: this path requires authentication")
	ErrPathNotAllowed   = fmt.Errorf("path not allowed: access to this path is forbidden")
)

// Safe-auth | 二级认证
var (
	ErrSafeTimeInvalid   = fmt.Errorf("safe time must be greater than zero")
	ErrNotPassedSafeAuth = fmt.Errorf("second-level authentication required")
)

// Tiered disable | 分级封禁
var (
	ErrLoginIDEmpty         = fmt.Errorf("loginID cannot be empty")
	ErrInvalidDisableLevel  = fmt.Errorf("invalid disable level: must be greater than or equal to 1")
	ErrDisableLevelExceeded = fmt.Errorf("account is disabled at the required level")
	ErrDisableServiceEmpty  = fmt.Errorf("service list cannot be empty for untie")
)

// OAuth2
var (
	ErrOAuth2InvalidClientID       = fmt.Errorf("invalid client_id")
	ErrOAuth2InvalidClientSecret   = fmt.Errorf("invalid client_secret")
	ErrOAuth2InvalidRedirectURI    = fmt.Errorf("invalid redirect_uri")
	ErrOAuth2RedirectURIContainsAt = fmt.Errorf("redirect_uri must not contain '@' character")
	ErrOAuth2IllegalRedirectURI    = fmt.Errorf("illegal redirect_uri: not in allow list")
	ErrOAuth2ScopeNotContracted    = fmt.Errorf("client has not contracted the requested scope")
	ErrOAuth2InvalidGrantType      = fmt.Errorf("invalid grant_type")
	ErrOAuth2InvalidResponseType   = fmt.Errorf("invalid response_type")
	ErrOAuth2InvalidCode           = fmt.Errorf("invalid authorization code")
	ErrOAuth2CodeUsed              = fmt.Errorf("authorization code already used")
	ErrOAuth2CodeExpired           = fmt.Errorf("authorization code expired")
	ErrOAuth2RedirectMismatch      = fmt.Errorf("redirect_uri mismatch")
	ErrOAuth2ClientMismatch        = fmt.Errorf("client mismatch")
	ErrOAuth2InvalidAccessToken    = fmt.Errorf("invalid access_token")
	ErrOAuth2InvalidRefreshToken   = fmt.Errorf("invalid refresh_token")
	ErrOAuth2InvalidUserCredential = fmt.Errorf("invalid user credential")
	ErrOAuth2GrantTypeNotAllowed   = fmt.Errorf("grant_type not allowed for this client")
	ErrOAuth2RequiredParamMissing  = fmt.Errorf("required parameter is missing")
)

// SSO
var (
	ErrSsoInvalidTicket        = fmt.Errorf("invalid ticket")
	ErrSsoTicketClientMismatch = fmt.Errorf("ticket does not belong to the specified client")
	ErrSsoInvalidRedirect      = fmt.Errorf("invalid sso redirect url")
	ErrSsoCallbackFailed       = fmt.Errorf("sso single-logout callback failed")
	ErrSsoSignatureInvalid     = fmt.Errorf("sso signature verification failed")
	ErrSsoServerUnreachable    = fmt.Errorf("sso server is unreachable")
	ErrSsoClientNotRegistered  = fmt.Errorf("sso client is not registered")
)

// API Key
var (
	ErrApiKeyNotFound = fmt.Errorf("api key not found")
	ErrApiKeyExpired  = fmt.Errorf("api key expired")
	ErrApiKeyDisabled = fmt.Errorf("api key disabled")
)

// Signature
var (
	ErrSignatureInvalid = fmt.Errorf("signature invalid")
	ErrSignatureExpired = fmt.Errorf("signature expired: timestamp too old")
	ErrNonceAlreadyUsed = fmt.Errorf("nonce already used: replay detected")
)

// Temp Token
var (
	ErrTempTokenNotFound = fmt.Errorf("temp token not found")
	ErrTempTokenExpired  = fmt.Errorf("temp token expired")
	ErrTempTokenUsed     = fmt.Errorf("temp token already used")
)

// Same-Token | 服务间调用令牌
var (
	ErrSameTokenInvalid = fmt.Errorf("same-token invalid: service-to-service authentication failed")
)

// System | 系统
var (
	ErrStorageUnavailable = fmt.Errorf("storage unavailable: unable to connect to storage backend")
	ErrInvalidConfig      = fmt.Errorf("invalid configuration")
)

// --- Wrapped helpers (fmt lives here so manager avoids ad-hoc strings) | 包装错误辅助函数 ---

// ErrTokenNotFoundForLogin token missing for account/device | 账号下无 token
func ErrTokenNotFoundForLogin(loginID string) error {
	return fmt.Errorf("%w: loginId=%s", ErrTokenNotFound, loginID)
}

// ErrInvalidStorageValueType unexpected type at storage key | 存储值类型异常
func ErrInvalidStorageValueType(key string) error {
	return fmt.Errorf("%w: key=%s", ErrInvalidValueType, key)
}

// ErrFeatureNotSupportedNamed unsupported feature | 功能不支持
func ErrFeatureNotSupportedNamed(feature string) error {
	return fmt.Errorf("%w: %s", ErrFeatureNotSupported, feature)
}

// ErrKickedOutWithToken kicked out marker with token context | 踢下线
func ErrKickedOutWithToken(token string) error {
	return fmt.Errorf("%w: token=%s", ErrKickedOut, token)
}

// ErrTokenReplacedWithToken replaced marker with token context | 顶号下线
func ErrTokenReplacedWithToken(token string) error {
	return fmt.Errorf("%w: token=%s", ErrTokenReplaced, token)
}

// ErrActiveTimeoutWithToken active timeout with token | 活跃超时
func ErrActiveTimeoutWithToken(token string) error {
	return fmt.Errorf("%w: token=%s", ErrActiveTimeout, token)
}

// ErrNotPassedSafeAuthWithService second-level auth not passed | 二级认证未通过
func ErrNotPassedSafeAuthWithService(service string) error {
	return fmt.Errorf("%w: service=%s", ErrNotPassedSafeAuth, service)
}

// ErrDisableLevelExceededWithContext tiered disable hit | 分级封禁触发
func ErrDisableLevelExceededWithContext(loginID, service string, level int) error {
	return fmt.Errorf("%w: loginId=%s service=%s level=%d", ErrDisableLevelExceeded, loginID, service, level)
}

// ErrMarshalTokenInfo marshal failure | TokenInfo 序列化失败
func ErrMarshalTokenInfo(cause error) error {
	return fmt.Errorf("%w: %v", ErrInvalidTokenData, cause)
}

// ErrMarshalSession marshal failure | Session 序列化失败
func ErrMarshalSession(cause error) error {
	return fmt.Errorf("%w: %v", ErrInvalidSessionData, cause)
}

// ErrStorageWrap wraps storage failures | 存储失败包装
func ErrStorageWrap(cause error) error {
	return fmt.Errorf("%w: %v", ErrStorageUnavailable, cause)
}

// ErrInvalidTokenDataWrap wraps decode failures | Token 数据解析失败
func ErrInvalidTokenDataWrap(cause error) error {
	return fmt.Errorf("%w: %v", ErrInvalidTokenData, cause)
}

// ErrOAuth2ParamMissing required OAuth2 parameter missing | OAuth2 缺少必填参数
func ErrOAuth2ParamMissing(param string) error {
	return fmt.Errorf("%w: %s", ErrOAuth2RequiredParamMissing, param)
}
