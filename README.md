# Sa-Token-Go

**English** | **[中文](README_zh.md)**

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.25-blue)](https://img.shields.io)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](https://opensource.org/licenses/Apache-2.0)

A lightweight, high-performance Go authentication and authorization framework, inspired by [sa-token](https://github.com/dromara/sa-token).

## ✨ Core Features

- 🔐 **Authentication** - Supports multi-device login, token management, and login-state validation
- 🛡️ **Authorization** - Fine-grained permission control with wildcard support (`*`, `user:*`, `user:*:view`)
- 👥 **Role Management** - Flexible role assignment, removal, and combined checks
- 🚫 **Account Disable** - Supports disable, restore, disable info query, and service-level disable
- 👢 **Kickout / Replace** - Supports kickout or replace by token, account, or device dimension
- 💾 **Session Management** - Supports reading session data by login ID or token
- ⏰ **Activity Detection** - Supports `ActiveTimeout` and auto-renew settings
- 🔄 **Auto Renewal** - Built-in activity renewal and renew-pool support
- 🎨 **Annotation Support** - Integration packages provide `CheckLoginMiddleware`, `CheckRoleMiddleware`, `CheckPermissionMiddleware`, and more
- 🎧 **Event Listener** - Built-in login, logout, renew, disable, permission-check, and role-check events with priority, filter, and stats support
- 📦 **Modular Design** - Clean layering with `core`, `stputil`, `com/*`, `integrations/*`, and `examples/*`
- 🔒 **Nonce Anti-Replay** - Supports nonce generation, verification, and one-time consumption
- 🔄 **OAuth2 / Refresh Token** - Includes OAuth2 authorization code flow and refresh token support
- 🧩 **Replaceable Components** - Codec, Generator, Log, Pool, and Storage can all be swapped

## 🚀 Quick Start

### 📥 Installation

#### Option 1: Simplified Import (Recommended) ✨

**Import only one framework integration package, then add a storage module as needed. The integration package already depends on `core` and `stputil`.**

```bash
# Import only the framework integration package
go get github.com/click33/sa-token-go/integrations/gin@latest # Gin integration, suitable for direct use in Gin projects
# or
go get github.com/click33/sa-token-go/integrations/echo@latest # Echo integration, suitable for direct use in Echo projects
# or
go get github.com/click33/sa-token-go/integrations/fiber@latest # Fiber integration, suitable for direct use in Fiber projects
# or
go get github.com/click33/sa-token-go/integrations/chi@latest # Chi integration, suitable for direct use in Chi projects
# or
go get github.com/click33/sa-token-go/integrations/gf@latest # GoFrame integration, suitable for direct use in GoFrame projects
# or
go get github.com/click33/sa-token-go/integrations/hertz@latest # Hertz integration, suitable for direct use in Hertz projects
# or
go get github.com/click33/sa-token-go/integrations/kratos@latest # Kratos integration, suitable for direct use in Kratos projects

# Storage module (choose one)
go get github.com/click33/sa-token-go/com/storage/memory@latest # In-memory storage, suitable for local development and simple testing
go get github.com/click33/sa-token-go/com/storage/redis@latest # Redis storage, suitable for production or multi-instance deployment
```

#### Option 2: Separate Import

```bash
# Core modules
go get github.com/click33/sa-token-go/core@latest     # Core capabilities, Builder, Manager, config, and more
go get github.com/click33/sa-token-go/stputil@latest  # Global authentication utility entry, usually used directly in business code

# Component modules
go get github.com/click33/sa-token-go/com/storage/memory@latest # In-memory storage, suitable for local development and simple testing
go get github.com/click33/sa-token-go/com/storage/redis@latest  # Redis storage, suitable for production or multi-instance deployment

# Framework integrations (optional)
go get github.com/click33/sa-token-go/integrations/gin@latest    # Gin integration
go get github.com/click33/sa-token-go/integrations/echo@latest   # Echo integration
go get github.com/click33/sa-token-go/integrations/fiber@latest  # Fiber integration
go get github.com/click33/sa-token-go/integrations/chi@latest    # Chi integration
go get github.com/click33/sa-token-go/integrations/gf@latest     # GoFrame integration
go get github.com/click33/sa-token-go/integrations/hertz@latest  # Hertz integration
go get github.com/click33/sa-token-go/integrations/kratos@latest # Kratos integration
```

### ⚡ Minimal Usage (One-Line Initialization)

```go
package main

import (
    "context"

    "github.com/click33/sa-token-go/com/storage/memory"
    "github.com/click33/sa-token-go/core/adapter"
    "github.com/click33/sa-token-go/core/builder"
    "github.com/click33/sa-token-go/stputil"
)

var ctx = context.Background()

func init() {
    stputil.SetManager(
        builder.NewBuilder().
            SetStorage(memory.NewStorage()).
            TokenName("Authorization").
            Timeout(86400).
            TokenStyle(adapter.TokenStyleRandom64).
            IsPrintBanner(true).
            Build(),
    )
}
```

**A startup banner will be printed by default:**

```text
   _____         ______      __                  ______     
  / ___/____ _  /_  __/___  / /_____  ____      / ____/____ 
  \__ \/ __  |   / / / __ \/ //_/ _ \/ __ \_____/ / __/ __ \
 ___/ / /_/ /   / / / /_/ / ,< /  __/ / / /_____/ /_/ / /_/ /
/____/\__,_/   /_/  \____/_/|_|\___/_/ /_/      \____/\____/ 

:: Sa-Token-Go ::                            (v0.1.5)
:: Go Version ::                             go1.25.0
:: GOOS/GOARCH ::                            linux/amd64

========================================
         Configuration Summary
========================================
AuthType         : default
TokenName        : Authorization
TokenStyle       : Random-64
AutoRenew        : Enabled
ActiveTimeout    : Disabled
========================================
```

```go
func main() {
    token, _ := stputil.Login(ctx, "1000")
    println("login success, token:", token)

    _ = stputil.AddPermissions(ctx, "1000", []string{"user:read", "user:write"})

    if stputil.HasPermission(ctx, "1000", "user:read") {
        println("has permission")
    }

    _ = stputil.Logout(ctx, token)
}
```

If you want to see more complete startup and routing styles, you can directly refer to these examples in the current repository:

- `examples/quick_start`
- `examples/gin`
- `examples/gf`
- `examples/echo`
- `examples/fiber`
- `examples/chi`
- `examples/hertz`
- `examples/kratos`

## 🔧 Core API

### 🔑 Authentication

```go
// Login
token, _ := stputil.Login(ctx, "1000")
token, _ := stputil.Login(ctx, "user123")
token, _ := stputil.Login(ctx, "1000", "mobile") // specify device

// Login with custom timeout
tempToken, _ := stputil.LoginWithTimeout(ctx, "1000", 2*time.Hour, "web")

// Renew login based on an existing token
_ = stputil.LoginByToken(ctx, token)

// Check login
isLogin := stputil.IsLogin(ctx, token)

// Get login ID
loginID, _ := stputil.GetLoginID(ctx, token)

// Get token info
tokenInfo, _ := stputil.GetTokenInfo(ctx, token)

// Get device, device ID, create time, and TTL
device, _ := stputil.GetDevice(ctx, token)
deviceID, _ := stputil.GetDeviceId(ctx, token)
createTime, _ := stputil.GetTokenCreateTime(ctx, token)
ttl, _ := stputil.GetTokenTTL(ctx, token)

// Logout
_ = stputil.Logout(ctx, token)
_ = stputil.LogoutByDevice(ctx, "1000", "mobile")
_ = stputil.LogoutByLoginID(ctx, "1000")
_ = stputil.LogoutByDeviceAndDeviceId(ctx, "1000", "mobile", "device-001")

// Kickout / Replace
_ = stputil.Kickout(ctx, token)
_ = stputil.KickoutByLoginID(ctx, "1000")
_ = stputil.KickoutByDevice(ctx, "1000", "web")
_ = stputil.ReplaceByLoginID(ctx, "1000")
_ = stputil.ReplaceByDevice(ctx, "1000", "app")
_ = stputil.RenewTimeout(ctx, token, 24*time.Hour)

_ = isLogin
_ = loginID
_ = tokenInfo
_ = tempToken
_ = device
_ = deviceID
_ = createTime
_ = ttl
```

### 🛡️ Permission Management

```go
// Add permissions
_ = stputil.AddPermissions(ctx, "1000", []string{
    "user:read",
    "user:write",
    "admin:*",
})

// Check permissions
hasPermission := stputil.HasPermission(ctx, "1000", "user:read")
hasPermission = stputil.HasPermission(ctx, "1000", "admin:delete")

// Multiple permission checks
hasAll := stputil.HasPermissionsAnd(ctx, "1000", []string{"user:read", "user:write"})
hasAny := stputil.HasPermissionsOr(ctx, "1000", []string{"admin:*", "super:*"})

// Query / remove permissions
perms, _ := stputil.GetPermissions(ctx, "1000")
_ = stputil.RemovePermissions(ctx, "1000", []string{"user:write"})

// Check by token
hasByToken := stputil.HasPermissionByToken(ctx, token, "user:read")

_ = hasPermission
_ = hasAll
_ = hasAny
_ = perms
_ = hasByToken
```

### 👥 Role Management

```go
// Add roles
_ = stputil.AddRoles(ctx, "1000", []string{"admin", "manager"})

// Check roles
hasRole := stputil.HasRole(ctx, "1000", "admin")

// Multiple role checks
hasAll := stputil.HasRolesAnd(ctx, "1000", []string{"admin", "manager"})
hasAny := stputil.HasRolesOr(ctx, "1000", []string{"admin", "super-admin"})

// Query / remove roles
roles, _ := stputil.GetRoles(ctx, "1000")
_ = stputil.RemoveRoles(ctx, "1000", []string{"manager"})

// Check by token
hasRoleByToken := stputil.HasRoleByToken(ctx, token, "admin")

_ = hasRole
_ = hasAll
_ = hasAny
_ = roles
_ = hasRoleByToken
```

### 💾 Session Management

```go
// Get session
sess, _ := stputil.GetSession(ctx, "1000")

// Or get session by token
sessByToken, _ := stputil.GetSessionByToken(ctx, token)

// Set data
sess.Set("nickname", "John")
sess.Set("age", 25)

// Read data
nickname := sess.GetString("nickname")
age := sess.GetInt("age")

// Delete field
sess.Delete("nickname")

// Get token list / terminal list of current account
tokenList, _ := stputil.GetTokenValueListByLoginID(ctx, "1000", true)
terminalList, _ := stputil.GetTerminalListByLoginID(ctx, "1000")

// Search token / session
tokenKeys, _ := stputil.SearchTokenValue(ctx, "1000", 0, 20)
sessionKeys, _ := stputil.SearchSessionId(ctx, "1000", 0, 20)

_ = age
_ = nickname
_ = sessByToken
_ = tokenList
_ = terminalList
_ = tokenKeys
_ = sessionKeys
```

### 🚫 Account Disable

```go
// Disable for 1 hour
_ = stputil.Disable(ctx, "1000", 1*time.Hour, "manual disable")

// Restore
_ = stputil.Untie(ctx, "1000")

// Check whether disabled
isDisabled := stputil.IsDisable(ctx, "1000")

// Get disable info and TTL
disableInfo, _ := stputil.GetDisableInfo(ctx, "1000")
ttl, _ := stputil.GetDisableTTL(ctx, "1000")

// Service-level disable
_ = stputil.DisableService(ctx, "1000", "comment", 30*time.Minute)
_ = stputil.DisableServiceLevelWithReason(ctx, "1000", "post", 2, time.Hour, "risk control")
serviceInfo, _ := stputil.GetDisableServiceInfo(ctx, "1000", "post")
serviceTTL, _ := stputil.GetDisableServiceTTL(ctx, "1000", "post")

_ = isDisabled
_ = disableInfo
_ = ttl
_ = serviceInfo
_ = serviceTTL
```

## 🌐 Framework Integration

### 🌟 Gin Integration (Single Import)

**The recommended approach now is to use `integrations/gin` directly and access builder, middleware, and context capability through the `satoken` alias.**

```go
import (
    "context"

    "github.com/click33/sa-token-go/com/storage/memory"
    satoken "github.com/click33/sa-token-go/integrations/gin"
    "github.com/gin-gonic/gin"
)

func main() {
    ctx := context.Background()

    mgr := satoken.NewDefaultBuilder().
        SetStorage(memory.NewStorage()).
        Timeout(7200).
        ActiveTimeout(1800).
        MaxLoginCount(3).
        Build()

    satoken.SetManager(mgr)

    r := gin.Default()
    r.Use(satoken.RegisterSaTokenContextMiddleware(ctx))

    r.POST("/login", func(c *gin.Context) {
        token, _ := satoken.Login(c.Request.Context(), "1000")
        c.JSON(200, gin.H{"token": token})
    })

    user := r.Group("/user")
    user.Use(satoken.AuthMiddleware(ctx))
    user.GET("/info", func(c *gin.Context) {
        saCtx, ok := satoken.GetSaTokenContext(c)
        if !ok {
            c.JSON(500, gin.H{"message": "failed to get context"})
            return
        }

        loginID, _ := saCtx.GetLoginID(c.Request.Context())
        c.JSON(200, gin.H{"loginId": loginID})
    })

    r.Run(":8080")
}
```

### 🎯 Annotation-Style Middleware

All integration packages now provide **annotation-style checking middleware** with a unified naming convention:

| Capability | Function |
|---|---|
| Ignore auth | `IgnoreMiddleware(...)` |
| Check login | `CheckLoginMiddleware(...)` |
| Check role | `CheckRoleMiddleware(...)` |
| Check permission | `CheckPermissionMiddleware(...)` |
| Check disable | `CheckDisableMiddleware(...)` |
| Combined check | `CheckAllMiddleware(...)` |

**Gin example:**

```go
annotation := r.Group("/api/annotation")

annotation.GET("/profile",
    satoken.CheckLoginMiddleware(ctx, handleProfile, handleAuthFail))

annotation.GET("/admin-data",
    satoken.CheckRoleMiddleware(ctx, []string{"admin"}, handleAdminData, handleAuthFail))

annotation.GET("/sensitive",
    satoken.CheckPermissionMiddleware(ctx, []string{"data:read"}, handleSensitiveData, handleAuthFail))

annotation.GET("/super",
    satoken.CheckAllMiddleware(ctx, []string{"super-admin"}, []string{"all:access"}, handleSuperData, handleAuthFail))
```

### 🌟 GoFrame Integration (Single Import)

```go
import (
    "context"

    "github.com/click33/sa-token-go/com/storage/memory"
    satoken "github.com/click33/sa-token-go/integrations/gf"
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/net/ghttp"
)

func main() {
    ctx := context.Background()

    mgr := satoken.NewDefaultBuilder().
        SetStorage(memory.NewStorage()).
        Timeout(7200).
        Build()

    satoken.SetManager(mgr)

    s := g.Server()
    s.Use(satoken.RegisterSaTokenContextMiddleware(ctx))

    s.Group("/api/user", func(group *ghttp.RouterGroup) {
        group.Middleware(satoken.AuthMiddleware(ctx))
        group.GET("/info", handleUserInfo)
    })

    s.Group("/api/annotation", func(group *ghttp.RouterGroup) {
        group.GET("/profile", satoken.CheckLoginMiddleware(ctx, handleProfile, handleAuthFail))
    })

    s.Run()
}
```

### 🔌 Other Framework Integrations

The current repository already provides these integration packages and matching examples:

- `integrations/echo` with `examples/echo`
- `integrations/fiber` with `examples/fiber`
- `integrations/chi` with `examples/chi`
- `integrations/gin` with `examples/gin`
- `integrations/gf` with `examples/gf`
- `integrations/hertz` with `examples/hertz`
- `integrations/kratos` with `examples/kratos`

These integration packages have been aligned to the same `SaToken` naming style, including:

- `RegisterSaTokenContextMiddleware`
- `AuthMiddleware`
- `RoleMiddleware`
- `PermissionMiddleware`
- `CheckLoginMiddleware`
- `CheckRoleMiddleware`
- `CheckPermissionMiddleware`
- `CheckAllMiddleware`

Common frameworks now follow a largely unified usage style, for example:

```go
// Echo
e.GET("/profile", saecho.CheckLoginMiddleware(ctx, handleProfile, handleAuthFail))

// Fiber
app.Get("/profile", safiber.CheckLoginMiddleware(ctx, handleProfile, handleAuthFail))

// Chi
r.With(sachi.AuthMiddleware()).Get("/profile", handleProfile)

// Hertz
h.GET("/profile", sahertz.CheckLoginMiddleware(ctx, handleProfile, handleAuthFail))

// Kratos
srv := http.NewServer(
    http.Middleware(
        sakratos.RegisterSaTokenContextMiddleware(),
        sakratos.CheckLoginMiddleware(handleAuthFail),
    ),
)
```

## 🚀 Advanced Features

### 🎨 Token Styles

The current code supports 9 token styles:

| Style | Format Example | Length | Use Case |
|---|---|---|---|
| `TokenStyleUUID` | `550e8400-e29b-41d4-...` | 36 | General purpose |
| `TokenStyleSimple` | `aB3dE5fG7hI9jK1l` | 16 | Compact token |
| `TokenStyleRandom32` | Random string | 32 | Higher security |
| `TokenStyleRandom64` | Random string | 64 | Recommended default |
| `TokenStyleRandom128` | Random string | 128 | High-security scenarios |
| `TokenStyleJWT` | `eyJhbGciOiJIUzI1...` | Variable | Stateless authentication |
| `TokenStyleHash` | `a3f5d8b2c1e4f6a9...` | 64 | SHA256 hash |
| `TokenStyleTimestamp` | `1700000000123_user1000_...` | Variable | Easier creation-time tracing |
| `TokenStyleTik` | `7Kx9mN2pQr4` | 11 | Short token scenarios |

**JWT token example:**

```go
mgr := builder.NewBuilder().
    SetStorage(memory.NewStorage()).
    TokenStyle(adapter.TokenStyleJWT).
    JwtSecretKey("your-256-bit-secret").
    Timeout(3600).
    Build()

stputil.SetManager(mgr)
token, _ := stputil.Login(ctx, "1000")
```

If you want to quickly experience usage styles under different frameworks, you can directly refer to the complete examples under `examples/*`.

### 🔒 Security Features

#### 🔐 Nonce Anti-Replay

```go
nonce, _ := stputil.GenerateNonce(ctx)
valid := stputil.VerifyNonce(ctx, nonce) // true
valid = stputil.VerifyNonce(ctx, nonce)  // false
_ = valid
```

#### 🔄 OAuth2 / Refresh Token

```go
import "github.com/click33/sa-token-go/core/oauth2"

_ = stputil.RegisterOAuth2Client(&oauth2.Client{
    ClientID:     "webapp",
    ClientSecret: "secret123",
    RedirectURIs: []string{"http://localhost:8080/callback"},
    GrantTypes:   []oauth2.GrantType{oauth2.GrantTypeAuthorizationCode, oauth2.GrantTypeRefreshToken},
    Scopes:       []string{"read", "write"},
})

authCode, _ := stputil.GenerateOAuth2AuthorizationCode(
    ctx,
    "webapp",
    "1000",
    "http://localhost:8080/callback",
    []string{"read"},
)

accessToken, _ := stputil.ExchangeOAuth2CodeForToken(
    ctx,
    authCode.Code,
    "webapp",
    "secret123",
    "http://localhost:8080/callback",
)

newToken, _ := stputil.RefreshOAuth2AccessToken(
    ctx,
    "webapp",
    accessToken.RefreshToken,
    "secret123",
)

_ = newToken
```

### 🎧 Event Listener

The current codebase already includes event-listener capability in `core/listener` and `manager`, with support for:

- Registering listeners by event type
- Priority control
- Async / sync execution
- Global filters
- Trigger stats and panic handling

Built-in events include:

- `EventLogin`
- `EventLogout`
- `EventKickout`
- `EventReplace`
- `EventDisable`
- `EventUntie`
- `EventRenew`
- `EventCreateSession`
- `EventDestroySession`
- `EventPermissionCheck`
- `EventRoleCheck`
- `EventDisableService`
- `EventUntieService`
- `EventAll`

For detailed registration and listener configuration, see:

- [Event Listener Guide](docs/guide/listener.md)

## 🏗️ Architecture Overview

The overall structure of `sa-token-go` can be summarized in 4 layers:

- `com/*`: replaceable component layer for storage, codec, logging, pool, token generator, and similar foundational implementations
- `core/*`: core capability layer for config, context, manager, event listener, nonce, OAuth2, and other core logic
- `stputil`: global utility entry that exposes common authentication, permission, role, session, and disable capabilities
- `integrations/*`: framework integration layer for Gin, GoFrame, Echo, Fiber, Chi, Hertz, Kratos, and more

For more detailed design notes, see:

- [Architecture Design](docs/design/architecture.md)

## 📁 Project Structure

```text
sa-token-go/
├── com/                    # Replaceable component modules
│   ├── codec/              # Codec implementations
│   │   ├── base64/         # Base64 codec
│   │   ├── json/           # Standard JSON codec
│   │   ├── jsonv2/         # JSON v2 codec
│   │   └── msgpack/        # MsgPack codec
│   ├── generator/          # Token generator implementations
│   │   └── sgenerator/     # Default token generator
│   ├── log/                # Logging implementations
│   │   ├── gf/             # GoFrame log adapter
│   │   ├── nop/            # No-op logger
│   │   └── slog/           # slog adapter
│   ├── pool/               # Goroutine pool implementations
│   │   └── ants/           # ants pool adapter
│   └── storage/            # Storage implementations
│       ├── memory/         # In-memory storage
│       └── redis/          # Redis storage
├── core/                   # Core capability modules
│   ├── adapter/            # Core adapter interfaces and constants
│   ├── banner/             # Startup banner printing
│   ├── builder/            # Builder
│   ├── config/             # Config definitions
│   ├── context/            # SaTokenContext wrappers
│   ├── listener/           # Event listener system
│   ├── manager/            # Core auth manager implementation
│   ├── nonce/              # Nonce anti-replay implementation
│   ├── oauth2/             # OAuth2 implementation
│   ├── serror/             # Error definitions
│   └── utils/              # Common utilities
├── stputil/                # Global authentication utility entry
├── integrations/           # Web framework integration layer
│   ├── chi/                # Chi integration
│   ├── echo/               # Echo integration
│   ├── fiber/              # Fiber integration
│   ├── gf/                 # GoFrame integration
│   ├── gin/                # Gin integration
│   ├── hertz/              # Hertz integration
│   └── kratos/             # Kratos integration
├── examples/               # Example projects
│   ├── chi/                # Chi example
│   ├── echo/               # Echo example
│   ├── fiber/              # Fiber example
│   ├── gf/                 # GoFrame example
│   ├── gin/                # Gin example
│   ├── hertz/              # Hertz example
│   ├── kratos/             # Kratos example
│   └── quick_start/        # Quick Start example
└── docs/                   # Documentation directory
```

## 📚 Documentation & Examples

### 📖 Guides

- [Quick Start](docs/tutorial/quick-start.md) - Initialize in 5 minutes
- [Authentication](docs/guide/authentication.md) - Login, logout, kickout, and replace
- [Permission](docs/guide/permission.md) - Permissions, roles, and combined checks
- [Annotation](docs/guide/annotation.md) - Annotation-style middleware across frameworks
- [Event Listener](docs/guide/listener.md) - Event system and listener capability
- [JWT Usage](docs/guide/jwt.md) - JWT-style token details
- [Redis Storage](docs/guide/redis-storage.md) - Redis storage and production configuration
- [Nonce Anti-Replay](docs/guide/nonce.md) - Nonce generation and verification
- [Refresh Token](docs/guide/refresh-token.md) - Refresh token overview
- [OAuth2](docs/guide/oauth2.md) - OAuth2 authorization code flow
- [Single Import](docs/guide/single-import.md) - `integrations/*` single-package import style

### 📘 API Documentation

- [StpUtil API](docs/api/stputil.md) - Complete global utility API reference

### 🧠 Design Documents

- [Architecture Design](docs/design/architecture.md) - Module relationships and main call flow
- [Auto Renew Design](docs/design/auto-renew.md) - Auto-renew mechanism
- [Modular Design](docs/design/modular.md) - Multi-module split and dependency strategy

### 🧪 Example Projects

| Example | Description | Path |
|---|---|---|
| Quick Start | Basic quick start example | [examples/quick_start/](examples/quick_start/) |
| Gin | Gin integration example | [examples/gin/](examples/gin/) |
| GoFrame | GoFrame integration example | [examples/gf/](examples/gf/) |
| Echo | Echo integration example | [examples/echo/](examples/echo/) |
| Fiber | Fiber integration example | [examples/fiber/](examples/fiber/) |
| Chi | Chi integration example | [examples/chi/](examples/chi/) |
| Hertz | Hertz integration example | [examples/hertz/](examples/hertz/) |
| Kratos | Kratos integration example | [examples/kratos/](examples/kratos/) |

### 💾 Storage Options

- [Memory Storage](com/storage/memory/)
- [Redis Storage](com/storage/redis/)

## 📄 License

Apache License 2.0

## 🙏 Acknowledgments

Implemented with inspiration from [sa-token](https://github.com/dromara/sa-token).

### Contributors

Special thanks to the following contributors:

- [@qprodn](https://github.com/qprodn)
- [@Zany2](https://github.com/Zany2)
- [@zyw](https://github.com/zyw)
- [@nuanxinqing123](https://github.com/nuanxinqing123)
- [@vera-byte](https://github.com/vera-byte)
- [@MoLing-Dong](https://github.com/MoLing-Dong)

## 📞 Support

- 💬 Issues: [GitHub Issues](https://github.com/click33/sa-token-go/issues)
- 📖 Docs entry: [docs/](docs/)

### WeChat Group

![sa-token-go WeChat group QR code](docs/wechat.JPG)
