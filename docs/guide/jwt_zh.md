# JWT 指南

[English](jwt.md) | 中文文档

## 概览

当前项目支持把 Token 生成风格切换为 `JWT`，入口是：

- `builder.NewBuilder().TokenStyle(TokenStyleJWT)`
- `builder.NewBuilder().JwtSecretKey(...)`
- `builder.NewBuilder().JwtSecret(...)`

## 重要说明

当前版本里的 JWT 只是“Token 字符串格式”改成 JWT，不代表系统完全无状态。

登录之后这些数据仍然会进入存储层：

- `TokenInfo`
- `Session`
- 自动续期标记
- 活跃时间标记

所以：

1. `Kickout`、`Replace`、`Logout` 依然可用
2. 权限、角色、封禁、Session 能力依然可用
3. Redis / Memory 这类存储仍然有价值

## JWT Claim

当前生成器里写入的 claim 主要有：

```json
{
  "loginId": "10001",
  "device": "web",
  "deviceId": "chrome-mac",
  "iat": 1710000000,
  "exp": 1710003600
}
```

其中：

- `loginId`
- `device`
- `deviceId`
- `iat`
- `exp`：只有配置了超时时才会写入

## 基本使用

```go
package main

import (
    "context"

    "github.com/click33/sa-token-go/com/storage/memory"
    "github.com/click33/sa-token-go/core/adapter"
    "github.com/click33/sa-token-go/core/builder"
    "github.com/click33/sa-token-go/stputil"
)

func initSaToken() {
    stputil.SetManager(
        builder.NewBuilder().
            SetStorage(memory.NewStorage()).
            TokenStyle(adapter.TokenStyleJWT).
            JwtSecretKey("your-very-strong-secret-key").
            Timeout(2 * 60 * 60).
            Build(),
    )
}

func main() {
    ctx := context.Background()
    token, _ := stputil.Login(ctx, "10001", "web", "chrome-mac")
    _ = token
}
```

### 使用快捷方式

```go
builder.NewBuilder().
    SetStorage(memory.NewStorage()).
    JwtSecret("your-very-strong-secret-key").
    Timeout(2 * 60 * 60).
    Build()
```

`JwtSecret(...)` 会同时把 `TokenStyle` 切换成 `JWT` 并设置密钥。

## 登录与校验

```go
ctx := context.Background()

token, err := stputil.Login(ctx, "10001", "web")

isLogin := stputil.IsLogin(ctx, token)
loginID, err := stputil.GetLoginID(ctx, token)
info, err := stputil.GetTokenInfo(ctx, token)
ttl, err := stputil.GetTokenTTL(ctx, token)
```

这些 API 和普通 UUID Token 的用法一致。

## 生成器能力

```go
generator := sgenerator.NewGenerator(7200, "your-secret", adapter.TokenStyleJWT)

token, err := generator.Generate("10001", "web", "chrome-mac")
claims, err := generator.ParseJWT(token)
err = generator.ValidateJWT(token)
loginID, err := generator.GetLoginIDFromJWT(token)
```

## 配置项

| 配置项 | 说明 |
|------|------|
| `TokenStyle(TokenStyleJWT)` | 开启 JWT 风格 |
| `JwtSecretKey(key)` | 设置 JWT 密钥 |
| `JwtSecret(key)` | 一步开启 JWT 并设置密钥 |
| `Timeout(seconds)` | 控制 `exp` 与服务端存储 TTL |
| `AutoRenew(true)` | 控制服务端续期逻辑 |

## 安全建议

### 使用强密钥

`TokenStyleJWT` 开启时，`JwtSecretKey` 不能为空。建议使用足够长、足够随机的密钥。

### 不要误以为它是纯无状态方案

如果你需要真正完全无状态的认证方案，当前项目并不是这个设计目标。当前 JWT 更适合：

1. 统一 token 表现形式
2. 方便网关或调试工具查看 claim
3. 与现有登录态、Session、权限体系共存

### 配合 HTTPS

JWT 本身只是签名，不是加密，生产环境里仍然应该通过 HTTPS 传输。

## JWT 与普通 Token 对比

| 项目 | JWT 风格 | UUID / Random 风格 |
|------|------|------|
| Token 可读性 | 更高，可解析 claim | 更低 |
| Token 长度 | 更长 | 更短 |
| 服务端存储 | 仍然需要 | 需要 |
| 踢下线 / 顶下线 | 支持 | 支持 |
| 权限 / Session | 支持 | 支持 |

## 相关文档

- [登录认证](authentication_zh.md)
- [Redis 存储](redis-storage_zh.md)
- [单包导入](single-import_zh.md)
