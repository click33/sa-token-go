English | [中文文档](stputil_zh.md)

# StpUtil API Documentation

## Overview

`stputil` is the global utility entry of Sa-Token-Go. It wraps common capabilities such as authentication, permission checks, roles, account disable, session access, nonce, and OAuth2.

In the current version, `stputil` uses `context.Context` as the unified entry, and different auth systems can be selected through the optional `authType` parameter.

## Initialization

```go
import (
    "context"

    "github.com/click33/sa-token-go/com/storage/memory"
    "github.com/click33/sa-token-go/core/builder"
    "github.com/click33/sa-token-go/stputil"
)

var ctx = context.Background()

func init() {
    stputil.SetManager(
        builder.NewBuilder().
            SetStorage(memory.NewStorage()).
            Build(),
    )
}
```

## Authentication API

### Login

Login and return a token.

**Signature**:
```go
func Login(ctx context.Context, loginID string, params ...string) (string, error)
```

**Parameters**:
- `ctx` - Request context
- `loginID` - Login ID
- `params` - Optional parameters in order: `[device, deviceId, authType]`

**Returns**:
- `string` - Token value
- `error` - Error information

**Example**:
```go
token, _ := stputil.Login(ctx, "1000")
token, _ := stputil.Login(ctx, "1000", "mobile")
token, _ := stputil.Login(ctx, "1000", "mobile", "device-001", "user")
```

### IsLogin

Check whether a token is logged in.

**Signature**:
```go
func IsLogin(ctx context.Context, tokenValue string, authType ...string) bool
```

**Example**:
```go
if stputil.IsLogin(ctx, token) {
    // logged in
}
```

### GetLoginID

Get the login ID by token.

**Signature**:
```go
func GetLoginID(ctx context.Context, tokenValue string, authType ...string) (string, error)
```

**Example**:
```go
loginID, err := stputil.GetLoginID(ctx, token)
```

### Logout

Logout by token.

**Signature**:
```go
func Logout(ctx context.Context, tokenValue string, authType ...string) error
```

**Example**:
```go
_ = stputil.Logout(ctx, token)
```

### Kickout

Kick a user offline by token.

**Signature**:
```go
func Kickout(ctx context.Context, tokenValue string, authType ...string) error
```

**Example**:
```go
_ = stputil.Kickout(ctx, token)
```

### Common Extended Methods

```go
func LoginWithTimeout(ctx context.Context, loginID string, timeout time.Duration, params ...string) (string, error)
func LoginByToken(ctx context.Context, tokenValue string, authType ...string) error
func CheckLogin(ctx context.Context, tokenValue string, authType ...string) error
func LogoutByDevice(ctx context.Context, loginID string, device string, authType ...string) error
func LogoutByDeviceAndDeviceId(ctx context.Context, loginID string, params ...string) error
func LogoutByLoginID(ctx context.Context, loginID string, authType ...string) error
func KickoutByDevice(ctx context.Context, loginID string, device string, authType ...string) error
func KickoutByDeviceAndDeviceId(ctx context.Context, loginID string, params ...string) error
func KickoutByLoginID(ctx context.Context, loginID string, authType ...string) error
func Replace(ctx context.Context, tokenValue string, authType ...string) error
func ReplaceByDevice(ctx context.Context, loginID string, device string, authType ...string) error
func ReplaceByDeviceAndDeviceId(ctx context.Context, loginID string, params ...string) error
func ReplaceByLoginID(ctx context.Context, loginID string, authType ...string) error
func GetTokenInfo(ctx context.Context, tokenValue string, authType ...string) (*manager.TokenInfo, error)
func GetDevice(ctx context.Context, tokenValue string, authType ...string) (string, error)
func GetDeviceId(ctx context.Context, tokenValue string, authType ...string) (string, error)
func GetTokenCreateTime(ctx context.Context, tokenValue string, authType ...string) (int64, error)
func GetTokenTTL(ctx context.Context, tokenValue string, authType ...string) (int64, error)
func RenewTimeout(ctx context.Context, tokenValue string, timeout time.Duration, authType ...string) error
```

## Permission API

### AddPermissions

Add permissions for a specified account.

**Signature**:
```go
func AddPermissions(ctx context.Context, loginID string, permissions []string, authType ...string) error
```

**Example**:
```go
_ = stputil.AddPermissions(ctx, "1000", []string{
    "user:read",
    "user:write",
    "admin:*",
})
```

### HasPermission

Check whether the account has a specified permission.

**Signature**:
```go
func HasPermission(ctx context.Context, loginID string, permission string, authType ...string) bool
```

**Example**:
```go
if stputil.HasPermission(ctx, "1000", "user:read") {
    // has permission
}
```

### HasPermissionsAnd

Check whether the account has all permissions (AND logic).

**Signature**:
```go
func HasPermissionsAnd(ctx context.Context, loginID string, permissions []string, authType ...string) bool
```

### HasPermissionsOr

Check whether the account has any permission (OR logic).

**Signature**:
```go
func HasPermissionsOr(ctx context.Context, loginID string, permissions []string, authType ...string) bool
```

### Common Extended Methods

```go
func AddPermissionsByToken(ctx context.Context, tokenValue string, permissions []string, authType ...string) error
func RemovePermissions(ctx context.Context, loginID string, permissions []string, authType ...string) error
func RemovePermissionsByToken(ctx context.Context, tokenValue string, permissions []string, authType ...string) error
func GetPermissions(ctx context.Context, loginID string, authType ...string) ([]string, error)
func GetPermissionsByToken(ctx context.Context, tokenValue string, authType ...string) ([]string, error)
func HasPermissionByToken(ctx context.Context, tokenValue string, permission string, authType ...string) bool
func HasPermissionsAndByToken(ctx context.Context, tokenValue string, permissions []string, authType ...string) bool
func HasPermissionsOrByToken(ctx context.Context, tokenValue string, permissions []string, authType ...string) bool
func CheckPermission(ctx context.Context, loginID string, permission string, authType ...string) error
func CheckPermissionAnd(ctx context.Context, loginID string, permissions []string, authType ...string) error
func CheckPermissionOr(ctx context.Context, loginID string, permissions []string, authType ...string) error
```

## Role Management API

### AddRoles

Add roles for a specified account.

**Signature**:
```go
func AddRoles(ctx context.Context, loginID string, roles []string, authType ...string) error
```

**Example**:
```go
_ = stputil.AddRoles(ctx, "1000", []string{"admin", "manager"})
```

### HasRole

Check whether the account has a specified role.

**Signature**:
```go
func HasRole(ctx context.Context, loginID string, role string, authType ...string) bool
```

**Example**:
```go
if stputil.HasRole(ctx, "1000", "admin") {
    // has admin role
}
```

### HasRolesAnd / HasRolesOr

Multiple role checks.

**Example**:
```go
stputil.HasRolesAnd(ctx, "1000", []string{"admin", "manager"})
stputil.HasRolesOr(ctx, "1000", []string{"admin", "super-admin"})
```

### Common Extended Methods

```go
func AddRolesByToken(ctx context.Context, tokenValue string, roles []string, authType ...string) error
func RemoveRoles(ctx context.Context, loginID string, roles []string, authType ...string) error
func RemoveRolesByToken(ctx context.Context, tokenValue string, roles []string, authType ...string) error
func GetRoles(ctx context.Context, loginID string, authType ...string) ([]string, error)
func GetRolesByToken(ctx context.Context, tokenValue string, authType ...string) ([]string, error)
func HasRoleByToken(ctx context.Context, tokenValue string, role string, authType ...string) bool
func HasRolesAndByToken(ctx context.Context, tokenValue string, roles []string, authType ...string) bool
func HasRolesOrByToken(ctx context.Context, tokenValue string, roles []string, authType ...string) bool
func CheckRole(ctx context.Context, loginID string, role string, authType ...string) error
func CheckRoleAnd(ctx context.Context, loginID string, roles []string, authType ...string) error
func CheckRoleOr(ctx context.Context, loginID string, roles []string, authType ...string) error
```

## Account Disable API

### Disable

Disable an account.

**Signature**:
```go
func Disable(ctx context.Context, loginID string, duration time.Duration, reason string, authType ...string) error
```

**Parameters**:
- `loginID` - Login ID
- `duration` - Disable duration
- `reason` - Disable reason

**Example**:
```go
_ = stputil.Disable(ctx, "1000", 1*time.Hour, "manual disable")
```

### IsDisable

Check whether an account is disabled.

**Signature**:
```go
func IsDisable(ctx context.Context, loginID string, authType ...string) bool
```

### Untie

Restore a disabled account.

**Signature**:
```go
func Untie(ctx context.Context, loginID string, authType ...string) error
```

### GetDisableTTL

Get the remaining disable time.

**Signature**:
```go
func GetDisableTTL(ctx context.Context, loginID string, authType ...string) (int64, error)
```

### Common Extended Methods

```go
func GetDisableInfo(ctx context.Context, loginID string, authType ...string) (*manager.DisableInfo, error)
func CheckDisable(ctx context.Context, loginID string, authType ...string) error
func DisableService(ctx context.Context, loginID, service string, duration time.Duration, authType ...string) error
func DisableServiceWithReason(ctx context.Context, loginID, service string, duration time.Duration, reason string, authType ...string) error
func DisableServiceLevel(ctx context.Context, loginID, service string, level int, duration time.Duration, authType ...string) error
func DisableServiceLevelWithReason(ctx context.Context, loginID, service string, level int, duration time.Duration, reason string, authType ...string) error
func UntieService(ctx context.Context, loginID, service string, authType ...string) error
func IsDisableService(ctx context.Context, loginID, service string, authType ...string) bool
func IsDisableServiceLevel(ctx context.Context, loginID, service string, level int, authType ...string) bool
func CheckDisableService(ctx context.Context, loginID string, services []string, authType ...string) error
func CheckDisableServiceLevel(ctx context.Context, loginID, service string, level int, authType ...string) error
func GetDisableServiceInfo(ctx context.Context, loginID, service string, authType ...string) (*manager.ServiceDisableInfo, error)
func GetDisableServiceTTL(ctx context.Context, loginID, service string, authType ...string) (int64, error)
```

## Session Management API

### GetSession

Get session by login ID.

**Signature**:
```go
func GetSession(ctx context.Context, loginID string, authType ...string) (*manager.Session, error)
```

### GetSessionByToken

Get session by token.

**Signature**:
```go
func GetSessionByToken(ctx context.Context, tokenValue string, authType ...string) (*manager.Session, error)
```

**Example**:
```go
sess, _ := stputil.GetSession(ctx, "1000")
sessByToken, _ := stputil.GetSessionByToken(ctx, token)

sess.Set("nickname", "John")
nickname := sess.GetString("nickname")

_ = sessByToken
_ = nickname
```

### Common Extended Methods

```go
func GetOnlineTerminalCount(ctx context.Context, loginID string, authType ...string) (int, error)
func GetOnlineTerminalCountByDevice(ctx context.Context, loginID string, device string, authType ...string) (int, error)
func GetOnlineTerminalCountByDeviceAndDeviceId(ctx context.Context, loginID string, device string, deviceId string, authType ...string) (int, error)
func ForEachTerminal(ctx context.Context, loginID string, visitor manager.TerminalVisitor, authType ...string) error
func ForEachTerminalByDevice(ctx context.Context, loginID, device string, visitor manager.TerminalVisitor, authType ...string) error
func GetTokenValueListByLoginID(ctx context.Context, loginID string, checkAlive bool, authType ...string) ([]string, error)
func GetTokenValueListByDeviceAndDeviceId(ctx context.Context, loginID string, device string, deviceId string, checkAlive bool, authType ...string) ([]string, error)
func GetTokenValueListByDevice(ctx context.Context, loginID string, device string, checkAlive bool, authType ...string) ([]string, error)
func GetTerminalListByLoginID(ctx context.Context, loginID string, authType ...string) ([]manager.TerminalInfo, error)
func GetTerminalListByLoginIDAndDevice(ctx context.Context, loginID string, device string, authType ...string) ([]manager.TerminalInfo, error)
func GetTerminalInfoByToken(ctx context.Context, tokenValue string, authType ...string) (*manager.TerminalInfo, error)
func GetTokenValueByLoginID(ctx context.Context, loginID string, authType ...string) (string, error)
func GetTokenValueByLoginIDAndDevice(ctx context.Context, loginID string, device string, authType ...string) (string, error)
func SearchTokenValue(ctx context.Context, keyword string, start, size int, authType ...string) ([]string, error)
func SearchSessionId(ctx context.Context, keyword string, start, size int, authType ...string) ([]string, error)
```

## Advanced API

### GetTokenInfo

Get detailed token information.

**Signature**:
```go
func GetTokenInfo(ctx context.Context, tokenValue string, authType ...string) (*manager.TokenInfo, error)
```

**Returns**:
```go
type TokenInfo struct {
    AuthType   string `json:"authType"`
    LoginID    string `json:"loginId"`
    Device     string `json:"device"`
    DeviceId   string `json:"deviceId"`
    CreateTime int64  `json:"createTime"`
}
```

### Nonce API

```go
func GenerateNonce(ctx context.Context, authType ...string) (string, error)
func GenerateNonceWithTimeout(ctx context.Context, timeout time.Duration, authType ...string) (string, error)
func VerifyNonce(ctx context.Context, nonce string, authType ...string) bool
func VerifyAndConsumeNonce(ctx context.Context, nonce string, authType ...string) error
func IsNonceValid(ctx context.Context, nonce string, authType ...string) bool
func GetNonceTTL(ctx context.Context, nonce string, authType ...string) (int64, error)
```

### OAuth2 API

```go
func RegisterOAuth2Client(client *oauth2.Client, authType ...string) error
func UnregisterOAuth2Client(clientID string, authType ...string) error
func GetOAuth2Client(clientID string, authType ...string) (*oauth2.Client, error)
func OAuth2Token(ctx context.Context, req *oauth2.TokenRequest, validateUser oauth2.UserValidator, authType ...string) (*oauth2.AccessToken, error)
func GenerateOAuth2AuthorizationCode(ctx context.Context, clientID, userID, redirectURI string, scopes []string, authType ...string) (*oauth2.AuthorizationCode, error)
func ExchangeOAuth2CodeForToken(ctx context.Context, code, clientID, clientSecret, redirectURI string, authType ...string) (*oauth2.AccessToken, error)
func OAuth2ClientCredentialsToken(ctx context.Context, clientID, clientSecret string, scopes []string, authType ...string) (*oauth2.AccessToken, error)
func OAuth2PasswordGrantToken(ctx context.Context, clientID, clientSecret, username, password string, scopes []string, validateUser oauth2.UserValidator, authType ...string) (*oauth2.AccessToken, error)
func RefreshOAuth2AccessToken(ctx context.Context, clientID, refreshToken, clientSecret string, authType ...string) (*oauth2.AccessToken, error)
func ValidateOAuth2AccessToken(ctx context.Context, accessToken string, authType ...string) bool
func ValidateOAuth2AccessTokenAndGetInfo(ctx context.Context, accessToken string, authType ...string) (*oauth2.AccessToken, error)
func RevokeOAuth2Token(ctx context.Context, accessToken string, authType ...string) error
```

### Manager Lifecycle API

```go
func SetManager(mgr *manager.Manager)
func GetManager(authType ...string) (*manager.Manager, error)
func DeleteManager(authType ...string) error
func DeleteAllManager()
```

## Complete Method List

### Manager Lifecycle
- `SetManager`
- `GetManager`
- `DeleteManager`
- `DeleteAllManager`

### Authentication
- `Login`
- `LoginWithTimeout`
- `LoginByToken`
- `Logout`
- `LogoutByDeviceAndDeviceId`
- `LogoutByDevice`
- `LogoutByLoginID`
- `Kickout`
- `Replace`
- `KickoutByDeviceAndDeviceId`
- `KickoutByDevice`
- `KickoutByLoginID`
- `ReplaceByDeviceAndDeviceId`
- `ReplaceByDevice`
- `ReplaceByLoginID`
- `IsLogin`
- `CheckLogin`
- `GetLoginID`
- `GetTokenInfo`
- `GetDevice`
- `GetDeviceId`
- `GetTokenCreateTime`
- `GetTokenTTL`
- `RenewTimeout`

### Online Terminals / Session
- `GetOnlineTerminalCount`
- `GetOnlineTerminalCountByDevice`
- `GetOnlineTerminalCountByDeviceAndDeviceId`
- `ForEachTerminal`
- `ForEachTerminalByDevice`
- `GetSession`
- `GetSessionByToken`
- `GetTokenValueListByLoginID`
- `GetTokenValueListByDeviceAndDeviceId`
- `GetTokenValueListByDevice`
- `GetTerminalListByLoginID`
- `GetTerminalListByLoginIDAndDevice`
- `GetTerminalInfoByToken`
- `GetTokenValueByLoginID`
- `GetTokenValueByLoginIDAndDevice`
- `SearchTokenValue`
- `SearchSessionId`

### Account Disable
- `Disable`
- `Untie`
- `IsDisable`
- `GetDisableInfo`
- `GetDisableTTL`
- `DisableService`
- `DisableServiceWithReason`
- `DisableServiceLevel`
- `DisableServiceLevelWithReason`
- `UntieService`
- `IsDisableService`
- `IsDisableServiceLevel`
- `CheckDisableService`
- `CheckDisableServiceLevel`
- `GetDisableServiceInfo`
- `GetDisableServiceTTL`
- `CheckDisable`

### Permission
- `AddPermissions`
- `AddPermissionsByToken`
- `RemovePermissions`
- `RemovePermissionsByToken`
- `GetPermissions`
- `GetPermissionsByToken`
- `HasPermission`
- `HasPermissionByToken`
- `HasPermissionsAnd`
- `HasPermissionsAndByToken`
- `HasPermissionsOr`
- `HasPermissionsOrByToken`
- `CheckPermission`
- `CheckPermissionAnd`
- `CheckPermissionOr`

### Role Management
- `AddRoles`
- `AddRolesByToken`
- `RemoveRoles`
- `RemoveRolesByToken`
- `GetRoles`
- `GetRolesByToken`
- `HasRole`
- `HasRoleByToken`
- `HasRolesAnd`
- `HasRolesAndByToken`
- `HasRolesOr`
- `HasRolesOrByToken`
- `CheckRole`
- `CheckRoleAnd`
- `CheckRoleOr`

### Nonce
- `GenerateNonce`
- `GenerateNonceWithTimeout`
- `VerifyNonce`
- `VerifyAndConsumeNonce`
- `IsNonceValid`
- `GetNonceTTL`

### OAuth2
- `RegisterOAuth2Client`
- `UnregisterOAuth2Client`
- `GetOAuth2Client`
- `OAuth2Token`
- `GenerateOAuth2AuthorizationCode`
- `ExchangeOAuth2CodeForToken`
- `OAuth2ClientCredentialsToken`
- `OAuth2PasswordGrantToken`
- `RefreshOAuth2AccessToken`
- `ValidateOAuth2AccessToken`
- `ValidateOAuth2AccessTokenAndGetInfo`
- `RevokeOAuth2Token`

## Next Steps

- [Authentication Guide](../guide/authentication.md)
- [Permission Guide](../guide/permission.md)
- [OAuth2 Guide](../guide/oauth2.md)
