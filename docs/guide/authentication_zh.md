# 登录认证指南

[English](authentication.md) | 中文文档

## 基本登录

### 初始化

```go
package main

import (
    "context"

    "github.com/click33/sa-token-go/com/storage/memory"
    "github.com/click33/sa-token-go/core/builder"
    "github.com/click33/sa-token-go/stputil"
)

func initSaToken() {
    stputil.SetManager(
        builder.NewBuilder().
            SetStorage(memory.NewStorage()).
            Build(),
    )
}

func main() {
    ctx := context.Background()
    token, _ := stputil.Login(ctx, "10001")
    _ = token
}
```

### 简单登录

```go
ctx := context.Background()

// 默认登录
token, err := stputil.Login(ctx, "10001")

// 指定设备类型
token, err = stputil.Login(ctx, "10001", "web")

// 指定设备类型和设备 ID
token, err = stputil.Login(ctx, "10001", "app", "ios-iphone-15")
```

### 指定过期时间登录

```go
ctx := context.Background()

// 当前 token 单独设置 2 小时过期
token, err := stputil.LoginWithTimeout(ctx, "10001", 2*time.Hour)

// 同时指定设备类型和设备 ID
token, err = stputil.LoginWithTimeout(ctx, "10001", 2*time.Hour, "web", "chrome-mac")
```

### 基于旧 Token 续期登录

```go
ctx := context.Background()

err := stputil.LoginByToken(ctx, token)
```

`LoginByToken()` 会在当前 token 仍然有效时，异步续期 token、session、活跃时间等相关信息。

## 检查登录状态

```go
ctx := context.Background()

// 返回布尔值
isLogin := stputil.IsLogin(ctx, token)

// 未登录时返回错误
err := stputil.CheckLogin(ctx, token)
```

## 获取登录信息

### 获取登录 ID

```go
ctx := context.Background()

loginID, err := stputil.GetLoginID(ctx, token)
```

### 获取 Token 信息

```go
ctx := context.Background()

info, err := stputil.GetTokenInfo(ctx, token)
if err == nil {
    fmt.Println("认证体系:", info.AuthType)
    fmt.Println("登录 ID:", info.LoginID)
    fmt.Println("设备类型:", info.Device)
    fmt.Println("设备 ID:", info.DeviceId)
    fmt.Println("创建时间:", info.CreateTime)
}
```

当前 `TokenInfo` 只包含：

- `AuthType`
- `LoginID`
- `Device`
- `DeviceId`
- `CreateTime`

### 获取 Token 其他信息

```go
ctx := context.Background()

device, err := stputil.GetDevice(ctx, token)
deviceId, err := stputil.GetDeviceId(ctx, token)
createTime, err := stputil.GetTokenCreateTime(ctx, token)
ttl, err := stputil.GetTokenTTL(ctx, token)
```

## 登出

### 根据 Token 登出

```go
ctx := context.Background()

err := stputil.Logout(ctx, token)
```

### 根据账号维度登出

```go
ctx := context.Background()

// 登出指定账号的所有终端
err := stputil.LogoutByLoginID(ctx, "10001")

// 登出指定账号下某个设备类型的所有终端
err = stputil.LogoutByDevice(ctx, "10001", "web")

// 登出指定账号下某个设备类型 + 设备 ID 的终端
err = stputil.LogoutByDeviceAndDeviceId(ctx, "10001", "app", "ios-iphone-15")
```

## 踢人下线

### 根据 Token 踢人

```go
ctx := context.Background()

err := stputil.Kickout(ctx, token)
```

### 根据账号维度踢人

```go
ctx := context.Background()

err := stputil.KickoutByLoginID(ctx, "10001")
err = stputil.KickoutByDevice(ctx, "10001", "web")
err = stputil.KickoutByDeviceAndDeviceId(ctx, "10001", "app", "ios-iphone-15")
```

被踢下线的 token 不会立即从存储中删除，而是会被标记为 `kickout` 状态，后续校验时会返回对应错误。

## 顶人下线

### 根据 Token 顶人

```go
ctx := context.Background()

err := stputil.Replace(ctx, token)
```

### 根据账号维度顶人

```go
ctx := context.Background()

err := stputil.ReplaceByLoginID(ctx, "10001")
err = stputil.ReplaceByDevice(ctx, "10001", "web")
err = stputil.ReplaceByDeviceAndDeviceId(ctx, "10001", "app", "ios-iphone-15")
```

被顶下线的 token 会被标记为 `replaced` 状态，适合用在“新登录挤掉旧登录”的场景中。

## 在线终端统计

```go
ctx := context.Background()

count, err := stputil.GetOnlineTerminalCount(ctx, "10001")
webCount, err := stputil.GetOnlineTerminalCountByDevice(ctx, "10001", "web")
singleCount, err := stputil.GetOnlineTerminalCountByDeviceAndDeviceId(ctx, "10001", "app", "ios-iphone-15")
```

## 登录配置

### 并发登录

```go
stputil.SetManager(
    builder.NewBuilder().
        SetStorage(memory.NewStorage()).
        IsConcurrent(true).
        Build(),
)
```

`IsConcurrent(true)` 表示允许同账号多端并发登录。  
`IsConcurrent(false)` 表示新登录时会处理旧终端。

### 共享 Token

```go
stputil.SetManager(
    builder.NewBuilder().
        SetStorage(memory.NewStorage()).
        IsConcurrent(true).
        IsShare(true).
        Build(),
)
```

`IsShare(true)` 表示并发登录时可复用已有 token。  
如果同时设置 `IsConcurrent(true)` 且 `IsShare(false)`，则每次登录都会生成新 token。

### 最大登录数

```go
stputil.SetManager(
    builder.NewBuilder().
        SetStorage(memory.NewStorage()).
        IsConcurrent(true).
        IsShare(false).
        MaxLoginCount(3).
        Build(),
)
```

当 `MaxLoginCount > 0` 时，超过上限的旧终端会被自动清理。

## 自动续期

### 配置方式

```go
stputil.SetManager(
    builder.NewBuilder().
        SetStorage(memory.NewStorage()).
        Timeout(24 * 60 * 60).
        AutoRenew(true).
        RenewMaxRefresh(30 * 60).
        RenewInterval(60).
        Build(),
)
```

### 工作机制

当前版本的自动续期并不是每次 `IsLogin()` 都无条件刷新，而是：

1. 校验 token 是否有效
2. 检查账号是否被封禁
3. 检查 `ActiveTimeout`
4. 当 `AutoRenew=true` 且 token 还有剩余 TTL 时，根据 `RenewMaxRefresh` 与 `RenewInterval` 判断是否需要异步续期

这套机制可以减少高频访问时的重复续期。

## 活跃超时

### 配置方式

```go
stputil.SetManager(
    builder.NewBuilder().
        SetStorage(memory.NewStorage()).
        Timeout(24 * 60 * 60).
        ActiveTimeout(30 * 60).
        Build(),
)
```

### 工作机制

开启 `ActiveTimeout` 后，系统会为 token 维护活跃时间：

1. 登录成功时初始化活跃时间
2. 鉴权校验时检查是否已超过活跃超时
3. 未超时则异步刷新活跃时间
4. 已超时则返回未登录相关错误

`Timeout` 表示绝对过期时间，`ActiveTimeout` 表示不活跃超时，两者可以同时存在。

## 完整配置示例

```go
stputil.SetManager(
    builder.NewBuilder().
        SetStorage(memory.NewStorage()).
        TokenName("satoken").
        Timeout(24 * 60 * 60).
        ActiveTimeout(30 * 60).
        AutoRenew(true).
        RenewMaxRefresh(30 * 60).
        RenewInterval(60).
        IsConcurrent(true).
        IsShare(false).
        MaxLoginCount(3).
        IsReadHeader(true).
        Build(),
)
```

## 相关文档

- [快速开始](../tutorial/quick-start_zh.md)
- [权限管理](permission_zh.md)
- [注解鉴权](annotation_zh.md)
- [JWT 指南](jwt_zh.md)
