# JWT Guide

[中文文档](jwt_zh.md) | English

## Overview

The current project supports switching the token generation style to `JWT` through:

- `builder.NewBuilder().TokenStyle(TokenStyleJWT)`
- `builder.NewBuilder().JwtSecretKey(...)`
- `builder.NewBuilder().JwtSecret(...)`

## Important Note

In the current implementation, JWT only changes the token format. It does not make the whole system fully stateless.

These items are still stored on the server side after login:

- `TokenInfo`
- `Session`
- auto-renew markers
- activity timeout markers

That means:

1. `Kickout`, `Replace`, and `Logout` still work
2. permissions, roles, disable checks, and session features still work
3. storage backends such as Memory or Redis are still meaningful

## JWT Claims

The current generator writes claims like:

```json
{
  "loginId": "10001",
  "device": "web",
  "deviceId": "chrome-mac",
  "iat": 1710000000,
  "exp": 1710003600
}
```

## Basic Usage

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

### Shortcut

```go
builder.NewBuilder().
    SetStorage(memory.NewStorage()).
    JwtSecret("your-very-strong-secret-key").
    Timeout(2 * 60 * 60).
    Build()
```

`JwtSecret(...)` both enables JWT style and sets the secret.

## Login and Validation

```go
ctx := context.Background()

token, err := stputil.Login(ctx, "10001", "web")

isLogin := stputil.IsLogin(ctx, token)
loginID, err := stputil.GetLoginID(ctx, token)
info, err := stputil.GetTokenInfo(ctx, token)
ttl, err := stputil.GetTokenTTL(ctx, token)
```

The usage is the same as with UUID or random token styles.

## Generator APIs

```go
generator := sgenerator.NewGenerator(7200, "your-secret", adapter.TokenStyleJWT)

token, err := generator.Generate("10001", "web", "chrome-mac")
claims, err := generator.ParseJWT(token)
err = generator.ValidateJWT(token)
loginID, err := generator.GetLoginIDFromJWT(token)
```

## Configuration

| Option | Description |
|------|------|
| `TokenStyle(TokenStyleJWT)` | enable JWT style |
| `JwtSecretKey(key)` | set the JWT secret |
| `JwtSecret(key)` | enable JWT and set the secret in one step |
| `Timeout(seconds)` | controls both `exp` and server-side TTL |
| `AutoRenew(true)` | controls server-side renew behavior |

## Security Notes

### Use a Strong Secret

When `TokenStyleJWT` is enabled, `JwtSecretKey` must not be empty. Use a long, random secret in production.

### Do Not Treat It As Fully Stateless

If you need a truly stateless authentication architecture, that is not the current design goal of this project. Here, JWT is better understood as:

1. a token format option
2. a convenient way to inspect basic claims
3. something that coexists with session, permission, and login-state features

### Use HTTPS

JWT is signed, not encrypted. It should still be transported over HTTPS in production.

## JWT vs Regular Token

| Item | JWT Style | UUID / Random Style |
|------|------|------|
| Token readability | higher | lower |
| Token length | longer | shorter |
| Server-side storage | still required | required |
| Kickout / Replace | supported | supported |
| Permissions / Session | supported | supported |

## Related Documentation

- [Authentication Guide](authentication.md)
- [Redis Storage](redis-storage.md)
- [Single Import Guide](single-import.md)
