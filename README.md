# Sa-Token-Go

**English** | **[中文](README_zh.md)**

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.21-blue)](https://img.shields.io)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](https://opensource.org/licenses/Apache-2.0)

A lightweight, high-performance Go authentication and authorization framework, inspired by [sa-token](https://github.com/dromara/sa-token).

## ✨ Core Features

- 🔐 **Authentication** - Multi-device login, Token management
- 🛡️ **Authorization** - Fine-grained permission control, wildcard support (`*`, `user:*`, `user:*:view`)
- 🛣️ **Path-Based Auth** - Flexible path-based authentication with Ant-style wildcards
- 👥 **Role Management** - Flexible role authorization mechanism
- 🚫 **Account Ban** - Temporary/permanent account disabling
- 👢 **Kickout** - Force user logout, multi-device mutual exclusion
- 💾 **Session Management** - Complete Session management
- ⏰ **Active Detection** - Automatic token activity detection
- 🔄 **Auto Renewal** - Asynchronous token auto-renewal (400% performance improvement)
- 🎨 **Annotation Support** - `@SaCheckLogin`, `@SaCheckRole`, `@SaCheckPermission`
- 🎧 **Event System** - Powerful event system with priority and async execution
- 📦 **Modular Design** - Import only what you need, minimal dependencies
- 🔒 **Nonce Anti-Replay** - Prevent replay attacks with one-time tokens
- 🔄 **Refresh Token** - Refresh token mechanism with seamless refresh
- 🔐 **OAuth2** - Complete OAuth2 authorization code flow implementation

## 🚀 Quick Start

### 📥 Installation

#### Option 1: Simplified Import (Recommended) ✨

**Import only one framework integration package, which automatically includes core and stputil!**

```bash
# Import only the framework integration (includes core + stputil automatically)
go get github.com/sa-tokens/sa-token-go/integrations/gin@latest    # Gin framework
# or
go get github.com/sa-tokens/sa-token-go/integrations/echo@latest   # Echo framework
# or
go get github.com/sa-tokens/sa-token-go/integrations/fiber@latest  # Fiber framework
# or
go get github.com/sa-tokens/sa-token-go/integrations/chi@latest    # Chi framework
# or
go get github.com/sa-tokens/sa-token-go/integrations/gf@latest     # GoFrame framework
# or
go get github.com/sa-tokens/sa-token-go/integrations/kratos@latest   # Kratos framework
# or
go get github.com/sa-tokens/sa-token-go/integrations/hertz@latest  # Hertz framework
```

# Storage module (choose one)
go get github.com/sa-tokens/sa-token-go/storage/memory@latest # Memory storage (dev)
go get github.com/sa-tokens/sa-token-go/storage/redis@latest  # Redis storage (prod)
```

#### Option 2: Separate Import

```bash
# Core modules
go get github.com/sa-tokens/sa-token-go/core@vlatest 
go get github.com/sa-tokens/sa-token-go/stputil@vlatest 

# Storage module (choose one)
go get github.com/sa-tokens/sa-token-go/storage/memory@latest # Memory storage (dev)
go get github.com/sa-tokens/sa-token-go/storage/redis@latest  # Redis storage (prod)

# Framework integration (optional)
go get github.com/sa-tokens/sa-token-go/integrations/gin@latest    # Gin framework
go get github.com/sa-tokens/sa-token-go/integrations/echo@latest   # Echo framework
go get github.com/sa-tokens/sa-token-go/integrations/fiber@latest  # Fiber framework
go get github.com/sa-tokens/sa-token-go/integrations/chi@latest    # Chi framework
go get github.com/sa-tokens/sa-token-go/integrations/gf@latest     # GoFrame framework
go get github.com/sa-tokens/sa-token-go/integrations/kratos@latest # Kratos framework
go get github.com/sa-tokens/sa-token-go/integrations/hertz@latest  # Hertz framework
```

### ⚡ Minimal Usage (One-line Initialization)

```go
package main

import (
    "github.com/sa-tokens/sa-token-go/core"
    "github.com/sa-tokens/sa-token-go/stputil"
    "github.com/sa-tokens/sa-token-go/storage/memory"
)

func init() {
    // One-line initialization! Shows startup banner
    stputil.SetManager(
        core.NewBuilder().
            Storage(memory.NewStorage()).
            TokenName("Authorization").
            Timeout(86400).                      // 24 hours
            TokenStyle(core.TokenStyleRandom64). // Token style
            IsPrintBanner(true).                 // Show startup banner
            Build(),
    )
}
```

**Startup banner will be displayed:**

```
   _____         ______      __                  ______     
  / ___/____ _  /_  __/___  / /_____  ____      / ____/____ 
  \__ \/ __  |   / / / __ \/ //_/ _ \/ __ \_____/ / __/ __ \
 ___/ / /_/ /   / / / /_/ / ,< /  __/ / / /_____/ /_/ / /_/ /
/____/\__,_/   /_/  \____/_/|_|\___/_/ /_/      \____/\____/ 
                                                             
:: Sa-Token-Go ::                                    (v0.1.9)
:: Go Version ::                                     go1.21.0
:: GOOS/GOARCH ::                                    linux/amd64

┌─────────────────────────────────────────────────────────┐
│ Token Style     : random64                              │
│ Token Timeout   : 86400                      seconds    │
│ Auto Renew      : true                                  │
└─────────────────────────────────────────────────────────┘
```

```go
func main() {
    // Use StpUtil directly without passing manager
    token, _ := stputil.Login(1000)
    println("Login successful, Token:", token)
    
    // Set permissions
    stputil.SetPermissions(1000, []string{"user:read", "user:write"})
    
    // Check permissions
    if stputil.HasPermission(1000, "user:read") {
        println("Has permission!")
    }
    
    // Logout
    stputil.Logout(1000)
}
```

## 🔧 Core API

### 🔑 Authentication

```go
// Login
token, _ := stputil.Login(1000)
token, _ := stputil.Login("user123")
token, _ := stputil.Login(1000, "mobile")  // Specify device

// Check login status
isLogin := stputil.IsLogin(token)

// Get login ID
loginID, _ := stputil.GetLoginID(token)

// Logout
stputil.Logout(1000)
stputil.LogoutByToken(token)

// Kickout
stputil.Kickout(1000)
stputil.Kickout(1000, "mobile")
```

### 🛡️ Permission Management

```go
// Set permissions
stputil.SetPermissions(1000, []string{
    "user:read",
    "user:write",
    "admin:*",  // Wildcard: matches all admin permissions
})

// Check single permission
hasPermission := stputil.HasPermission(1000, "user:read")
hasPermission := stputil.HasPermission(1000, "admin:delete")  // Wildcard match

// Check multiple permissions
hasAll := stputil.HasPermissionsAnd(1000, []string{"user:read", "user:write"})  // AND logic
hasAny := stputil.HasPermissionsOr(1000, []string{"admin", "super"})           // OR logic
```

### 👥 Role Management

```go
// Set roles
stputil.SetRoles(1000, []string{"admin", "manager"})

// Check role
hasRole := stputil.HasRole(1000, "admin")

// Check multiple roles
hasAll := stputil.HasRolesAnd(1000, []string{"admin", "manager"})
hasAny := stputil.HasRolesOr(1000, []string{"admin", "super"})
```

### 💾 Session Management

```go
// Get session
sess, _ := stputil.GetSession(1000)

// Set data
sess.Set("nickname", "John")
sess.Set("age", 25)

// Get data
nickname := sess.GetString("nickname")
age := sess.GetInt("age")

// Delete data
sess.Delete("nickname")

// Delete session
stputil.DeleteSession(1000)
```

### 🚫 Account Management

```go
// Disable for 1 hour
stputil.Disable(1000, 1*time.Hour)

// Permanent disable
stputil.Disable(1000, 0)

// Enable account
stputil.Untie(1000)

// Check if disabled
isDisabled := stputil.IsDisable(1000)

// Get remaining disable time
remainingTime, _ := stputil.GetDisableTime(1000)
```

### 🧩 Plan-Based Advanced Demos

#### 1) Context Identity (Switch + Token-in-Context)

```go
ctx := context.Background()
ctx = stputil.SetTokenValueToCtx(ctx, token)

// Default: resolve loginId from token in context
loginID, err := stputil.GetLoginIDFromCtx(ctx)

// Switch identity for current call chain (higher priority than token parsing)
ctx = stputil.SwitchTo(ctx, "admin-1001")
switchedID, _ := stputil.GetLoginIDFromCtx(ctx) // -> admin-1001
```

#### 2) Safe Auth + Tiered Disable

```go
// Open second-level auth for 5 minutes
_ = stputil.OpenSafe(token, "pay", 300)
_ = stputil.CheckSafe(token, "pay")

// Tiered disable: service=comment, level=2
_ = stputil.DisableLevel(1000, "comment", 2, time.Hour)
_ = stputil.CheckDisableLevel(1000, "comment", 1) // blocked when level >= 1
```

#### 3) Replaced/Kickout + Terminal Query

```go
// Replace current device token(s) according to ReplacedRange
_ = stputil.Replaced(1000, "mobile")

// Query active terminals and trusted device flags
terminals, _ := stputil.GetTerminalListByLoginID(1000)
isTrusted := stputil.IsTrustDeviceID(1000, "ios-device-id")
_ = stputil.AddTrustDeviceID(1000, "ios-device-id")

// Search tokens/sessions
tokens, _ := stputil.SearchTokenValue("abc", 0, 20, true)
_ = tokens
```

## 🌐 Framework Integration

### Unified token extraction (`plugin-token-interceptor` plan)

Design and rollout are captured in [.cursor/plans/plugin-token-interceptor_8a06e5ac.plan.md](.cursor/plans/plugin-token-interceptor_8a06e5ac.plan.md). Implemented behavior:

1. **`ResolveTokenName` / `ReadTokenFromRequest`** live in [`core/context/context.go`](core/context/context.go). `SaTokenContext.GetTokenValue()` calls `ReadTokenFromRequest` so all code paths share the same order and prefix handling.
2. **`core/satoken.go`** re-exports `ResolveTokenName` and `ReadTokenFromRequest` as package-level identifiers so integrations can use `core.ReadTokenFromRequest` without importing `context` separately.
3. **Seven integrations** (Gin, Echo, Fiber, Hertz, Chi, GoFrame, Kratos) each provide **`TokenInterceptor()`**, which reads the token once via `core.ReadTokenFromRequest`, stores it on the framework context under **`satoken_token`**, and does **not** perform login checks. Handlers retrieve it with **`GetTokenFromCtx(...)`**. **`PathAuthMiddleware`** uses the same helper instead of ad-hoc header/cookie reads (fixes missing Query/`CutTokenPrefix`/`Authorization` fallback).
4. **Read order**: Header (including Bearer when the resolved name is `Authorization`, plus `Authorization` fallback when `TokenName` is custom) → Cookie → Query (`?tokenName=value`, api-key style). **`mgr.CutTokenPrefix`** is applied to the final raw string. Header/cookie reads respect **`IsReadHeader`** / **`IsReadCookie`**; Query is attempted when earlier steps yield nothing.

The following mirrors the production helpers in `core/context/context.go` (**English comments for this excerpt**):

```go
package context

import (
	"strings"

	"github.com/sa-tokens/sa-token-go/core/adapter"
	"github.com/sa-tokens/sa-token-go/core/config"
	"github.com/sa-tokens/sa-token-go/core/manager"
)

const bearerPrefix = "Bearer "
const AuthHeaderName = "Authorization"

// ResolveTokenName uses cfg.TokenName when non-empty; otherwise "Authorization".
func ResolveTokenName(cfg *config.Config) string {
	if cfg != nil && strings.TrimSpace(cfg.TokenName) != "" {
		return cfg.TokenName
	}
	return AuthHeaderName
}

// extractBearerToken removes a leading case-insensitive "Bearer " prefix.
func extractBearerToken(auth string) string {
	auth = strings.TrimSpace(auth)
	if auth == "" {
		return ""
	}
	if len(auth) > 7 && strings.EqualFold(auth[:7], bearerPrefix) {
		return strings.TrimSpace(auth[7:])
	}
	return auth
}

// ReadTokenFromRequest: Header → Cookie → Query; then CutTokenPrefix on the value.
func ReadTokenFromRequest(ctx adapter.RequestContext, mgr *manager.Manager) string {
	if ctx == nil || mgr == nil {
		return ""
	}
	cfg := mgr.GetConfig()
	name := ResolveTokenName(cfg)

	readHeader := cfg == nil || cfg.IsReadHeader
	readCookie := cfg == nil || cfg.IsReadCookie

	if readHeader {
		if v := strings.TrimSpace(ctx.GetHeader(name)); v != "" {
			if strings.EqualFold(name, AuthHeaderName) {
				if t := extractBearerToken(v); t != "" {
					return mgr.CutTokenPrefix(t)
				}
			}
			return mgr.CutTokenPrefix(v)
		}
		if !strings.EqualFold(name, AuthHeaderName) {
			if auth := strings.TrimSpace(ctx.GetHeader(AuthHeaderName)); auth != "" {
				if t := extractBearerToken(auth); t != "" {
					return mgr.CutTokenPrefix(t)
				}
			}
		}
	}

	if readCookie {
		if v := strings.TrimSpace(ctx.GetCookie(name)); v != "" {
			return mgr.CutTokenPrefix(v)
		}
	}

	if v := strings.TrimSpace(ctx.GetQuery(name)); v != "" {
		return mgr.CutTokenPrefix(v)
	}

	return ""
}
```

### 🌟 Gin Integration (Single Import)

**New way: Import only `integrations/gin` to use all features!**

```go
import (
    "github.com/gin-gonic/gin"
    sagin "github.com/sa-tokens/sa-token-go/integrations/gin"  // Only this import needed!
    "github.com/sa-tokens/sa-token-go/storage/memory"
)

func main() {
    // Initialize (all features in sagin package)
    storage := memory.NewStorage()
    config := sagin.DefaultConfig()
    manager := sagin.NewManager(storage, config)
    sagin.SetManager(manager)
    plugin := sagin.NewPlugin(manager)

    r := gin.Default()
    
    // Login endpoint
    r.POST("/login", func(c *gin.Context) {
        userID := c.PostForm("user_id")
        token, _ := sagin.Login(userID)
        c.JSON(200, gin.H{"token": token})
    })
    
    // Use annotation-style decorators (like Java)
    r.GET("/public", sagin.Ignore(), publicHandler)                  // Public access
    r.GET("/user", sagin.CheckLogin(), userHandler)                  // Login required
    r.GET("/admin", sagin.CheckPermission("admin:*"), adminHandler)  // Permission required
    r.GET("/manager", sagin.CheckRole("manager"), managerHandler)    // Role required
    r.GET("/sensitive", sagin.CheckDisable(), sensitiveHandler)      // Check if disabled

    // Recommended middleware order for protected routes:
    // 1) TokenInterceptor: normalize token from Header/Cookie/Query
    // 2) AuthMiddleware: validate login state
    api := r.Group("/api")
    api.Use(plugin.TokenInterceptor(), plugin.AuthMiddleware())
    api.GET("/token", func(c *gin.Context) {
        // Read parsed token directly from framework context
        c.JSON(200, gin.H{"token": sagin.GetTokenFromCtx(c)})
    })
    
    r.Run(":8080")
}
```

### 🎯 Annotation Decorators

**Supported annotations:**

| Annotation | Description | Example |
|------------|-------------|---------|
| `@SaIgnore` | Ignore authentication | `sagin.Ignore()` |
| `@SaCheckLogin` | Check login | `sagin.CheckLogin()` |
| `@SaCheckRole` | Check role | `sagin.CheckRole("admin")` |
| `@SaCheckPermission` | Check permission | `sagin.CheckPermission("admin:*")` |
| `@SaCheckDisable` | Check if disabled | `sagin.CheckDisable()` |

**Usage example:**

```go
import sagin "github.com/sa-tokens/sa-token-go/integrations/gin"

func main() {
    r := gin.Default()

    // Public access - ignore authentication
    r.GET("/public", sagin.Ignore(), publicHandler)

    // Login required
    r.GET("/user/info", sagin.CheckLogin(), userInfoHandler)

    // Admin permission required
    r.GET("/admin", sagin.CheckPermission("admin:*"), adminHandler)

    // Any of multiple permissions (OR logic)
    r.GET("/user-or-admin",
        sagin.CheckPermission("user:read", "admin:*"),
        userOrAdminHandler)

    // Admin role required
    r.GET("/manager", sagin.CheckRole("admin"), managerHandler)

    // Check if account is disabled
    r.GET("/sensitive", sagin.CheckDisable(), sensitiveHandler)

    r.Run(":8080")
}
```

### 🌟 GoFrame Integration (Single Import)

**GoFrame framework integration with full feature support!**

```go
import (
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/net/ghttp"
    sagf "github.com/sa-tokens/sa-token-go/integrations/gf"  // Only this import needed!
    "github.com/sa-tokens/sa-token-go/storage/memory"
)

func main() {
    // Initialize (all features in sagf package)
    storage := memory.NewStorage()
    config := sagf.DefaultConfig()
    manager := sagf.NewManager(storage, config)
    sagf.SetManager(manager)
    
    s := g.Server()
    
    // Login endpoint
    s.BindHandler("POST:/login", func(r *ghttp.Request) {
        userID := r.Get("user_id").String()
        token, _ := sagf.Login(userID)
        r.Response.WriteJson(g.Map{"token": token})
    })
    
    // Use annotation-style decorators (like Java)
    s.BindHandler("GET:/public", sagf.Ignore(), publicHandler)                  // Public access
    s.BindHandler("GET:/user", sagf.CheckLogin(), userHandler)                  // Login required
    s.BindHandler("GET:/admin", sagf.CheckPermission("admin:*"), adminHandler)  // Permission required
    s.BindHandler("GET:/manager", sagf.CheckRole("manager"), managerHandler)    // Role required
    s.BindHandler("GET:/sensitive", sagf.CheckDisable(), sensitiveHandler)      // Check if disabled
    
    s.SetPort(8080)
    s.Run()
}
```

### 🔌 Other Framework Integrations

**Echo / Fiber / Chi / Kratos** also support annotation decorators:

```go
// Echo
import saecho "github.com/sa-tokens/sa-token-go/integrations/echo"
e.GET("/user", saecho.CheckLogin(), handler)

// Fiber
import safiber "github.com/sa-tokens/sa-token-go/integrations/fiber"
app.Get("/user", safiber.CheckLogin(), handler)

// Chi
import sachi "github.com/sa-tokens/sa-token-go/integrations/chi"
r.Get("/user", sachi.CheckLogin(), handler)

// Kratos
import sakratos "github.com/sa-tokens/sa-token-go/integrations/kratos"
// Use Plugin.Server() as middleware

// Hertz
import sahertz "github.com/sa-tokens/sa-token-go/integrations/hertz"
h.GET("/user", sahertz.CheckLogin(), handler)
```

All integration plugins now provide:

- `TokenInterceptor()` - unified token extraction (`Header -> Cookie -> Query(apikey)`), with `TokenPrefix` trimming.
- `GetTokenFromCtx(...)` - fetch parsed token from framework context in handlers.

## 🎨 Advanced Features

### 🎨 Token Styles

Sa-Token-Go supports 9 token generation styles:

| Style | Format Example | Length | Use Case |
|-------|---------------|--------|----------|
| **UUID** | `550e8400-e29b-41d4-...` | 36 | General purpose |
| **Simple** | `aB3dE5fG7hI9jK1l` | 16 | Compact tokens |
| **Random32/64/128** | Random string | 32/64/128 | High security |
| **JWT** | `eyJhbGciOiJIUzI1...` | Variable | Stateless auth |
| **Hash** 🆕 | `a3f5d8b2c1e4f6a9...` | 64 | SHA256 hash |
| **Timestamp** 🆕 | `1700000000123_user1000_...` | Variable | Time traceable |
| **Tik** 🆕 | `7Kx9mN2pQr4` | 11 | Short ID (like TikTok) |

**JWT Token Support:**

```go
// Use JWT Token
stputil.SetManager(
    core.NewBuilder().
        Storage(memory.NewStorage()).
        TokenStyle(core.TokenStyleJWT).          // Use JWT
        JwtSecretKey("your-256-bit-secret").     // JWT secret
        Timeout(3600).                           // 1 hour expiration
        Build(),
)

// Login to get JWT Token
token, _ := stputil.Login(1000)
// Format: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

[👉 View Token Style Examples](examples/token-styles/)

### 🔒 Security Features

#### 🔐 Nonce Anti-Replay Attack

```go
// Generate nonce
nonce, _ := stputil.GenerateNonce()

// Verify nonce (one-time use)
valid := stputil.VerifyNonce(nonce)  // true
valid = stputil.VerifyNonce(nonce)   // false (prevents replay)
```

#### 🔄 Refresh Token Mechanism

```go
// Login to get access token and refresh token
tokenInfo, _ := stputil.LoginWithRefreshToken(1000, "web")
fmt.Println("Access Token:", tokenInfo.AccessToken)
fmt.Println("Refresh Token:", tokenInfo.RefreshToken)

// Refresh access token
newInfo, _ := stputil.RefreshAccessToken(tokenInfo.RefreshToken)
```

#### 🔑 OAuth2 Authorization Code Flow

```go
// Create OAuth2 server
oauth2Server := stputil.GetOAuth2Server()

// Register client
oauth2Server.RegisterClient(&core.OAuth2Client{
    ClientID:     "webapp",
    ClientSecret: "secret123",
    RedirectURIs: []string{"http://localhost:8080/callback"},
    GrantTypes:   []core.OAuth2GrantType{core.GrantTypeAuthorizationCode},
    Scopes:       []string{"read", "write"},
})

// Generate authorization code
authCode, _ := oauth2Server.GenerateAuthorizationCode(
    "webapp", "http://localhost:8080/callback", "user123", []string{"read"},
)

// Exchange authorization code for access token
accessToken, _ := oauth2Server.ExchangeCodeForToken(
    authCode.Code, "webapp", "secret123", "http://localhost:8080/callback",
)
```

[👉 View Complete OAuth2 Example](examples/oauth2-example/)

### 🎧 Event System

Listen to authentication and authorization events for audit logging, security monitoring, etc:

```go
storage := memory.NewStorage()

manager := core.NewBuilder().
    Storage(storage).
    Build()

// Listen to login events
manager.RegisterFunc(core.EventLogin, func(data *core.EventData) {
    fmt.Printf("[LOGIN] User: %s, Token: %s\n", data.LoginID, data.Token)
})

// Listen to logout events
manager.RegisterFunc(core.EventLogout, func(data *core.EventData) {
    fmt.Printf("[LOGOUT] User: %s\n", data.LoginID)
})

// Advanced: priority and sync execution
manager.RegisterWithConfig(core.EventLogin,
    core.ListenerFunc(auditLogger),
    core.ListenerConfig{
        Priority: 100,   // High priority
        Async:    false, // Sync execution
    },
)

// Listen to all events (wildcard)
manager.RegisterFunc(core.EventAll, func(data *core.EventData) {
    log.Printf("[%s] %s", data.Event, data.LoginID)
})

// Access advanced controls via the underlying EventManager
manager.GetEventManager().SetPanicHandler(customPanicHandler)

// Use the manager globally
stputil.SetManager(manager)
```

**Available events:**

- `EventLogin` - User login
- `EventLogout` - User logout  
- `EventKickout` - Force logout
- `EventDisable` - Account ban
- `EventPermissionCheck` - Permission check
- `EventRoleCheck` - Role check
- `EventAll` - All events (wildcard)

[→ View Event System Documentation](docs/guide/listener.md)

## 📦 Project Structure

```
sa-token-go/
├── core/                    # Core module
│   ├── adapter/            # Adapter interfaces
│   ├── builder/            # Builder pattern
│   ├── config/             # Configuration
│   ├── context/            # Context
│   ├── listener/           # Event listener
│   ├── manager/            # Authentication manager
│   ├── oauth2/             # OAuth2 implementation 🆕
│   ├── security/           # Security features (Nonce, RefreshToken) 🆕
│   ├── session/            # Session management
│   ├── token/              # Token generator
│   └── utils/              # Utility functions
│
├── stputil/                # Global utility
│
├── storage/                # Storage modules
│   ├── memory/             # Memory storage
│   └── redis/              # Redis storage
│
├── integrations/           # Framework integrations
│   ├── gin/                # Gin integration (with annotations)
│   ├── echo/               # Echo integration
│   ├── fiber/              # Fiber integration
│   ├── chi/                # Chi integration
│   ├── gf/                 # GoFrame integration
│   ├── kratos/             # Kratos integration
│   └── hertz/              # Hertz integration
│
├── examples/               # Example projects
│   ├── quick-start/        # Quick start
│   ├── token-styles/       # Token style demos 🆕
│   ├── security-features/  # Security features demos 🆕
│   ├── oauth2-example/     # Complete OAuth2 example 🆕
│   ├── annotation/         # Annotation usage
│   ├── jwt-example/        # JWT example
│   ├── redis-example/      # Redis example
│   ├── listener-example/   # Event listener example
│   └── gin/echo/fiber/chi/gf/kratos/hertz/ # Framework integration examples
│
└── docs/                   # Documentation
    ├── tutorial/           # Tutorials
    ├── guide/              # Usage guides
    ├── api/                # API documentation
    └── design/             # Design documents
```

## 📚 Documentation & Examples

### 📖 Documentation

- [Quick Start](docs/tutorial/quick-start.md) - Get started in 5 minutes
- [Authentication](docs/guide/authentication.md) - Authentication guide
- [Path-Based Auth](docs/guide/path-auth.md) - Path-based authentication guide
- [Permission](docs/guide/permission.md) - Permission system
- [Annotations](docs/guide/annotation.md) - Decorator pattern guide
- [Event Listener](docs/guide/listener.md) - Event system guide
- [JWT Integration](docs/guide/jwt.md) - JWT token guide
- [Redis Storage](docs/guide/redis-storage.md) - Redis storage configuration
- [Nonce Anti-Replay](docs/guide/nonce.md) - Nonce anti-replay attack
- [Refresh Token](docs/guide/refresh-token.md) - Refresh token mechanism
- [OAuth2](docs/guide/oauth2.md) - OAuth2 authorization guide

### 📋 API Reference

- [StpUtil API](docs/api/stputil.md) - Complete global utility API reference

### 🏗️ Design Documentation

- [Architecture Design](docs/design/architecture.md) - System architecture and data flow
- [Auto-Renewal Design](docs/design/auto-renew.md) - Asynchronous renewal mechanism
- [Modular Design](docs/design/modular.md) - Module organization strategy

### 💡 Example Projects

| Example | Description | Path |
|---------|-------------|------|
| ⚡ Quick Start | Builder+StpUtil minimal usage | [examples/quick-start/](examples/quick-start/) |
| 🎨 Token Styles | 9 token style demonstrations | [examples/token-styles/](examples/token-styles/) |
| 🔒 Security Features | Nonce/RefreshToken/OAuth2 | [examples/security-features/](examples/security-features/) |
| 🔐 OAuth2 Example | Complete OAuth2 implementation | [examples/oauth2-example/](examples/oauth2-example/) |
| 📝 Annotations | Annotation usage example | [examples/annotation/](examples/annotation/) |
| 🔑 JWT Example | JWT token usage | [examples/jwt-example/](examples/jwt-example/) |
| 💾 Redis Example | Redis storage example | [examples/redis-example/](examples/redis-example/) |
| 🎧 Event Listener | Event system usage | [examples/listener-example/](examples/listener-example/) |
| 🌐 Gin (Simple) | Minimal Gin integration | [examples/gin/gin-simple/](examples/gin/gin-simple/) |
| 🌐 Gin (Full) | Config-driven Gin integration | [examples/gin/gin-example/](examples/gin/gin-example/) |
| 🌐 Echo Integration | Echo framework integration | [examples/echo/echo-example/](examples/echo/echo-example/) |
| 🌐 Fiber Integration | Fiber framework integration | [examples/fiber/fiber-example/](examples/fiber/fiber-example/) |
| 🌐 Chi Integration | Chi framework integration | [examples/chi/chi-example/](examples/chi/chi-example/) |
| 🌐 GoFrame Integration | GoFrame framework integration | [examples/gf/](examples/gf/) |
| 🌐 Kratos Integration | Kratos framework integration | [examples/kratos/kratos-example/](examples/kratos/kratos-example/) |
| 🌐 Hertz Integration | Hertz framework integration | [examples/hertz/herz-example/](examples/hertz/herz-example/) |

### 💾 Storage Options

- [Memory Storage](storage/memory/) - For development environment
- [Redis Storage](storage/redis/) - For production environment

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by [sa-token](https://github.com/dromara/sa-token) - A powerful Java authentication framework
- Built with ❤️ using Go

### Contributors

Special thanks to the following contributors for their valuable contributions:

- [@qprodn](https://github.com/qprodn)
- [@Zany2](https://github.com/Zany2)
- [@zyw](https://github.com/zyw)
- [@nuanxinqing123](https://github.com/nuanxinqing123)
- [@vera-byte](https://github.com/vera-byte)
- [@MoLing-Dong](https://github.com/MoLing-Dong)

## 📞 Support

- 📧 Email: <support@sa-token-go.dev>
- 💬 Issues: [GitHub Issues](https://github.com/sa-tokens/sa-token-go/issues)
- 📖 Documentation: [docs/](docs/)

---
