[English](stputil.md) | 中文文档

# StpUtil API 文档

## 概述

`stputil` 是 Sa-Token-Go 的全局工具入口，封装了登录认证、权限、角色、封禁、Session、Nonce、OAuth2 等常用能力。

当前版本的 `stputil` 以 `context.Context` 为统一入口，认证体系可通过可选 `authType` 参数区分。

## 初始化

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

## 登录认证 API

### Login

登录并返回 Token。

**签名**：
```go
func Login(ctx context.Context, loginID string, params ...string) (string, error)
```

**参数**：
- `ctx` - 请求上下文
- `loginID` - 登录 ID
- `params` - 可选参数，顺序为 `[device, deviceId, authType]`

**返回**：
- `string` - Token 值
- `error` - 错误信息

**示例**：
```go
token, _ := stputil.Login(ctx, "1000")
token, _ := stputil.Login(ctx, "1000", "mobile")
token, _ := stputil.Login(ctx, "1000", "mobile", "device-001", "user")
```

### IsLogin

检查 Token 是否已登录。

**签名**：
```go
func IsLogin(ctx context.Context, tokenValue string, authType ...string) bool
```

**示例**：
```go
if stputil.IsLogin(ctx, token) {
    // 已登录
}
```

### GetLoginID

根据 Token 获取登录 ID。

**签名**：
```go
func GetLoginID(ctx context.Context, tokenValue string, authType ...string) (string, error)
```

**示例**：
```go
loginID, err := stputil.GetLoginID(ctx, token)
```

### Logout

根据 Token 登出。

**签名**：
```go
func Logout(ctx context.Context, tokenValue string, authType ...string) error
```

**示例**：
```go
_ = stputil.Logout(ctx, token)
```

### Kickout

根据 Token 踢人下线。

**签名**：
```go
func Kickout(ctx context.Context, tokenValue string, authType ...string) error
```

**示例**：
```go
_ = stputil.Kickout(ctx, token)
```

### 常用扩展方法

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

## 权限验证 API

### AddPermissions

为指定账号添加权限。

**签名**：
```go
func AddPermissions(ctx context.Context, loginID string, permissions []string, authType ...string) error
```

**示例**：
```go
_ = stputil.AddPermissions(ctx, "1000", []string{
    "user:read",
    "user:write",
    "admin:*",
})
```

### HasPermission

检查是否拥有指定权限。

**签名**：
```go
func HasPermission(ctx context.Context, loginID string, permission string, authType ...string) bool
```

**示例**：
```go
if stputil.HasPermission(ctx, "1000", "user:read") {
    // 有权限
}
```

### HasPermissionsAnd

检查是否拥有所有权限（AND 逻辑）。

**签名**：
```go
func HasPermissionsAnd(ctx context.Context, loginID string, permissions []string, authType ...string) bool
```

### HasPermissionsOr

检查是否拥有任一权限（OR 逻辑）。

**签名**：
```go
func HasPermissionsOr(ctx context.Context, loginID string, permissions []string, authType ...string) bool
```

### 常用扩展方法

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

## 角色管理 API

### AddRoles

为指定账号添加角色。

**签名**：
```go
func AddRoles(ctx context.Context, loginID string, roles []string, authType ...string) error
```

**示例**：
```go
_ = stputil.AddRoles(ctx, "1000", []string{"admin", "manager"})
```

### HasRole

检查是否拥有指定角色。

**签名**：
```go
func HasRole(ctx context.Context, loginID string, role string, authType ...string) bool
```

**示例**：
```go
if stputil.HasRole(ctx, "1000", "admin") {
    // 有 admin 角色
}
```

### HasRolesAnd / HasRolesOr

多角色检查。

**示例**：
```go
stputil.HasRolesAnd(ctx, "1000", []string{"admin", "manager"})
stputil.HasRolesOr(ctx, "1000", []string{"admin", "super-admin"})
```

### 常用扩展方法

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

## 账号封禁 API

### Disable

封禁账号。

**签名**：
```go
func Disable(ctx context.Context, loginID string, duration time.Duration, reason string, authType ...string) error
```

**参数**：
- `loginID` - 登录 ID
- `duration` - 封禁时长
- `reason` - 封禁原因

**示例**：
```go
_ = stputil.Disable(ctx, "1000", 1*time.Hour, "manual disable")
```

### IsDisable

检查账号是否被封禁。

**签名**：
```go
func IsDisable(ctx context.Context, loginID string, authType ...string) bool
```

### Untie

解封账号。

**签名**：
```go
func Untie(ctx context.Context, loginID string, authType ...string) error
```

### GetDisableTTL

获取剩余封禁时间。

**签名**：
```go
func GetDisableTTL(ctx context.Context, loginID string, authType ...string) (int64, error)
```

### 常用扩展方法

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

## Session 管理 API

### GetSession

根据登录 ID 获取 Session。

**签名**：
```go
func GetSession(ctx context.Context, loginID string, authType ...string) (*manager.Session, error)
```

### GetSessionByToken

根据 Token 获取 Session。

**签名**：
```go
func GetSessionByToken(ctx context.Context, tokenValue string, authType ...string) (*manager.Session, error)
```

**示例**：
```go
sess, _ := stputil.GetSession(ctx, "1000")
sessByToken, _ := stputil.GetSessionByToken(ctx, token)

sess.Set("nickname", "张三")
nickname := sess.GetString("nickname")

_ = sessByToken
_ = nickname
```

### 常用扩展方法

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

## 高级 API

### GetTokenInfo

获取 Token 详细信息。

**签名**：
```go
func GetTokenInfo(ctx context.Context, tokenValue string, authType ...string) (*manager.TokenInfo, error)
```

**返回**：
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

### Manager 生命周期 API

```go
func SetManager(mgr *manager.Manager)
func GetManager(authType ...string) (*manager.Manager, error)
func DeleteManager(authType ...string) error
func DeleteAllManager()
```

## 完整方法列表

### Manager 生命周期
- `SetManager`
- `GetManager`
- `DeleteManager`
- `DeleteAllManager`

### 登录认证
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

### 在线终端 / Session
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

### 账号封禁
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

### 权限验证
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

### 角色管理
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

## 下一步

- [登录认证指南](../guide/authentication_zh.md)
- [权限验证指南](../guide/permission_zh.md)
- [OAuth2 指南](../guide/oauth2_zh.md)
