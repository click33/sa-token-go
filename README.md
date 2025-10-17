# Sa-Token-Go

**English** | **[中文](README_zh.md)**

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.21-blue)]()
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](https://opensource.org/licenses/Apache-2.0)

A lightweight, high-performance Go authentication and authorization framework, inspired by [sa-token](https://github.com/dromara/sa-token).

## ✨ Core Features

- 🔐 **Authentication** - Multi-device login, Token management
- 🛡️ **Authorization** - Fine-grained permission control, wildcard support (`*`, `user:*`, `user:*:view`)
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

## 🎨 Token Styles

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

[👉 View Token Style Examples](examples/token-styles/)

## 🔒 Security Features

### Nonce Anti-Replay Attack

```go
// Generate nonce
nonce, _ := stputil.GenerateNonce()

// Verify nonce (one-time use)
valid := stputil.VerifyNonce(nonce)  // true
valid = stputil.VerifyNonce(nonce)   // false (prevents replay)
```

### Refresh Token Mechanism

```go
// Login to get access token and refresh token
tokenInfo, _ := stputil.LoginWithRefreshToken(1000, "web")
fmt.Println("Access Token:", tokenInfo.AccessToken)
fmt.Println("Refresh Token:", tokenInfo.RefreshToken)

// Refresh access token
newInfo, _ := stputil.RefreshAccessToken(tokenInfo.RefreshToken)
```

### OAuth2 Authorization Code Flow

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

## 🚀 Quick Start

### Installation

#### Option 1: Simplified Import (Recommended) ✨

**Import only one framework integration package, which automatically includes core and stputil!**

```bash
# Import only the framework integration (includes core + stputil automatically)
go get github.com/click33/sa-token-go/integrations/gin@v0.1.0    # Gin framework
# or
go get github.com/click33/sa-token-go/integrations/echo@v0.1.0   # Echo framework
# or
go get github.com/click33/sa-token-go/integrations/fiber@v0.1.0  # Fiber framework
# or
go get github.com/click33/sa-token-go/integrations/chi@v0.1.0    # Chi framework

# Storage module (choose one)
go get github.com/click33/sa-token-go/storage/memory@v0.1.0  # Memory storage (dev)
go get github.com/click33/sa-token-go/storage/redis@v0.1.0   # Redis storage (prod)
```

#### Option 2: Separate Import

```bash
# Core modules
go get github.com/click33/sa-token-go/core@v0.1.0
go get github.com/click33/sa-token-go/stputil@v0.1.0

# Storage module (choose one)
go get github.com/click33/sa-token-go/storage/memory@v0.1.0  # Memory storage (dev)
go get github.com/click33/sa-token-go/storage/redis@v0.1.0   # Redis storage (prod)

# Framework integration (optional)
go get github.com/click33/sa-token-go/integrations/gin@v0.1.0    # Gin framework
go get github.com/click33/sa-token-go/integrations/echo@v0.1.0   # Echo framework
go get github.com/click33/sa-token-go/integrations/fiber@v0.1.0  # Fiber framework
go get github.com/click33/sa-token-go/integrations/chi@v0.1.0    # Chi framework
```

### Minimal Usage (One-line Initialization)

```go
package main

import (
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/storage/memory"
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

// Startup banner will be displayed:
//    _____         ______      __                  ______     
//   / ___/____ _  /_  __/___  / /_____  ____      / ____/____ 
//   \__ \/ __  |   / / / __ \/ //_/ _ \/ __ \_____/ / __/ __ \
//  ___/ / /_/ /   / / / /_/ / ,< /  __/ / / /_____/ /_/ / /_/ /
// /____/\__,_/   /_/  \____/_/|_|\___/_/ /_/      \____/\____/ 
//                                                              
// :: Sa-Token-Go ::                                    (v0.1.0)
// :: Go Version ::                                     go1.21.0
// :: GOOS/GOARCH ::                                    darwin/arm64
//
// ┌─────────────────────────────────────────────────────────┐
// │ Token Style     : random64                              │
// │ Token Timeout   : 86400                      seconds    │
// │ Auto Renew      : true                                  │
// └─────────────────────────────────────────────────────────┘

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

### Gin Framework Integration (Single Import) ✨

**New way: Import only `integrations/gin` to use all features!**

```go
import (
    "github.com/gin-gonic/gin"
    sagin "github.com/click33/sa-token-go/integrations/gin"  // Only this import needed!
    "github.com/click33/sa-token-go/storage/memory"
)

func main() {
    // Initialize (all features in sagin package)
    storage := memory.NewStorage()
    config := sagin.DefaultConfig()  // Use sagin.DefaultConfig
    manager := sagin.NewManager(storage, config)  // Use sagin.NewManager
    sagin.SetManager(manager)  // Use sagin.SetManager
    
    r := gin.Default()
    
    // Login endpoint
    r.POST("/login", func(c *gin.Context) {
        userID := c.PostForm("user_id")
        token, _ := sagin.Login(userID)  // Use sagin.Login
        c.JSON(200, gin.H{"token": token})
    })
    
    // Use annotation-style decorators (like Java)
    r.GET("/public", sagin.Ignore(), publicHandler)                  // Public access
    r.GET("/user", sagin.CheckLogin(), userHandler)                  // Login required
    r.GET("/admin", sagin.CheckPermission("admin:*"), adminHandler)  // Permission required
    r.GET("/manager", sagin.CheckRole("manager"), managerHandler)    // Role required
    r.GET("/sensitive", sagin.CheckDisable(), sensitiveHandler)      // Check if disabled
    
    r.Run(":8080")
}
```

## 📦 Project Structure

```
sa-token-go/
├── core/                          # 🔴 Core module (required)
│   ├── adapter/                   # Adapter interfaces
│   │   ├── storage.go            # Storage interface
│   │   └── context.go            # Request context interface
│   ├── manager/                   # Authentication manager
│   ├── builder/                   # Builder pattern
│   ├── session/                   # Session management
│   ├── token/                     # Token generator (JWT support)
│   ├── listener/                  # Event listener system
│   ├── banner/                    # Startup banner
│   ├── config/                    # Configuration
│   ├── context/                   # Sa-Token context
│   ├── utils/                     # Utility functions
│   ├── oauth2/                    # OAuth2 implementation
│   ├── security/                  # Security features
│   │   ├── nonce.go              # Nonce anti-replay
│   │   └── refresh_token.go      # Refresh token mechanism
│   ├── errors.go                  # Error definitions
│   └── satoken.go                 # Core exports
│
├── stputil/                       # 🔧 Global utility (recommended)
│   └── stputil.go                # StpUtil.Login(), StpUtil.Logout()...
│
├── integrations/                  # 🌐 Framework integrations (optional)
│   ├── gin/                      # Gin integration
│   │   ├── export.go             # Re-export core + stputil
│   │   ├── plugin.go             # Gin plugin
│   │   ├── context.go            # Gin context adapter
│   │   └── annotation.go         # Annotation decorators
│   ├── echo/                     # Echo integration
│   │   ├── export.go             # Re-export core + stputil
│   │   ├── plugin.go             # Echo plugin
│   │   └── context.go            # Echo context adapter
│   ├── fiber/                    # Fiber integration
│   │   ├── export.go             # Re-export core + stputil
│   │   ├── plugin.go             # Fiber plugin
│   │   └── context.go            # Fiber context adapter
│   └── chi/                      # Chi integration
│       ├── export.go             # Re-export core + stputil
│       ├── plugin.go             # Chi plugin
│       └── context.go            # Chi context adapter
│
├── storage/                       # 💾 Storage implementations
│   ├── memory/                   # Memory storage (development)
│   └── redis/                    # Redis storage (production)
│
├── examples/                      # 📚 Example projects
│   ├── quick-start/              # Quick start example
│   ├── gin/                      # Gin examples
│   │   ├── gin-example/          # Complete Gin example
│   │   └── gin-simple/           # Simple Gin example (single import)
│   ├── echo/                     # Echo examples
│   ├── fiber/                    # Fiber examples
│   ├── chi/                      # Chi examples
│   ├── token-styles/             # Token style examples
│   ├── security-features/        # Security feature examples
│   ├── oauth2-example/           # OAuth2 complete example
│   ├── jwt-example/              # JWT example
│   ├── redis-example/            # Redis storage example
│   ├── listener-example/         # Event listener example
│   └── annotation/               # Annotation example
│
├── docs/                          # 📖 Documentation
│   ├── guide/                    # Usage guides
│   │   ├── single-import.md      # Single import guide
│   │   ├── authentication.md     # Authentication guide
│   │   ├── permission.md         # Permission guide
│   │   ├── annotation.md         # Annotation guide
│   │   ├── jwt.md                # JWT guide
│   │   ├── listener.md           # Event listener guide
│   │   ├── nonce.md              # Nonce guide
│   │   ├── refresh-token.md      # Refresh token guide
│   │   ├── oauth2.md             # OAuth2 guide
│   │   └── redis-storage.md      # Redis storage guide
│   ├── api/                      # API documentation
│   ├── design/                   # Design documents
│   └── tutorial/                 # Tutorials
│
├── go.work                        # Go workspace file
├── LICENSE                        # Apache 2.0 License
├── README.md                      # This file
└── README_zh.md                   # Chinese documentation
```

## 🔧 Configuration

### Basic Configuration

```go
config := &core.Config{
    TokenName:        "Authorization",     // Token name in header/cookie
    Timeout:          7200,               // Token timeout (seconds)
    ActiveTimeout:    1800,               // Active timeout (seconds)
    IsConcurrent:     true,               // Allow concurrent login
    IsShare:          true,               // Share session
    TokenStyle:       core.TokenStyleRandom64, // Token generation style
    IsLog:            true,               // Enable logging
    IsPrintBanner:    true,               // Print startup banner
    IsReadHeader:     true,               // Read token from header
    IsReadCookie:     false,              // Read token from cookie
    IsReadBody:       false,              // Read token from body
    CookieConfig: core.CookieConfig{
        Domain:       "",                 // Cookie domain
        Path:         "/",                // Cookie path
        Secure:       false,              // HTTPS only
        HttpOnly:     true,               // HTTP only
        SameSite:     "",                 // SameSite policy
    },
}
```

### Builder Pattern

```go
manager := core.NewBuilder().
    Storage(memory.NewStorage()).
    TokenName("satoken").
    Timeout(86400).
    TokenStyle(core.TokenStyleJWT).
    IsPrintBanner(true).
    Build()
```

## 🎯 Usage Examples

### Authentication

```go
// Login
token, err := stputil.Login(1000, "web")

// Check login status
isLogin := stputil.IsLogin(token)

// Get login ID
loginID, err := stputil.GetLoginID(token)

// Logout
stputil.Logout(1000, "web")
```

### Permission Management

```go
// Set permissions
stputil.SetPermissions(1000, []string{"user:read", "user:write", "admin:*"})

// Check single permission
hasPermission := stputil.HasPermission(1000, "user:read")

// Check multiple permissions (AND)
err := stputil.CheckPermissionAnd(1000, "user:read", "user:write")

// Check multiple permissions (OR)
err := stputil.CheckPermissionOr(1000, "admin:*", "user:write")
```

### Role Management

```go
// Set roles
stputil.SetRoles(1000, []string{"user", "admin"})

// Check role
hasRole := stputil.HasRole(1000, "admin")

// Check multiple roles
err := stputil.CheckRoleAnd(1000, "user", "admin")
```

### Account Management

```go
// Disable account for 1 hour
stputil.Disable(1000, time.Hour)

// Check if account is disabled
isDisabled := stputil.IsDisable(1000)

// Kick user out
stputil.Kickout(1000, "web")

// Unlock account
stputil.Untie(1000)
```

## 🌐 Framework Integration

### Gin

```go
import sagin "github.com/click33/sa-token-go/integrations/gin"

r.GET("/user", sagin.CheckLogin(), userHandler)
r.GET("/admin", sagin.CheckPermission("admin:*"), adminHandler)
r.GET("/public", sagin.Ignore(), publicHandler)
```

### Echo

```go
import saecho "github.com/click33/sa-token-go/integrations/echo"

e.GET("/user", saecho.CheckLogin(), userHandler)
e.GET("/admin", saecho.CheckPermission("admin:*"), adminHandler)
```

### Fiber

```go
import safiber "github.com/click33/sa-token-go/integrations/fiber"

app.Get("/user", safiber.CheckLogin(), userHandler)
app.Get("/admin", safiber.CheckPermission("admin:*"), adminHandler)
```

### Chi

```go
import sachi "github.com/click33/sa-token-go/integrations/chi"

r.Get("/user", sachi.CheckLogin(), userHandler)
r.Get("/admin", sachi.CheckPermission("admin:*"), adminHandler)
```

## 📊 Performance

- **QPS**: 10,000+ requests per second
- **Memory**: Low memory footprint
- **Concurrent**: Thread-safe design
- **Redis**: Support for Redis cluster

## 🛠️ Development

### Build

```bash
# Build all modules
go build ./...

# Run tests
go test ./...

# Run benchmarks
go test -bench=. ./...
```

### Examples

```bash
# Run Gin example
cd examples/gin/gin-simple
go run main.go

# Run OAuth2 example
cd examples/oauth2-example
go run main.go
```

## 📚 Documentation

- [📖 Documentation Center](docs/README.md)
- [🚀 Quick Start Guide](docs/tutorial/quick-start.md)
- [🔧 Single Import Guide](docs/guide/single-import.md)
- [🔐 Authentication Guide](docs/guide/authentication.md)
- [🛡️ Permission Guide](docs/guide/permission.md)
- [🎨 Annotation Guide](docs/guide/annotation.md)
- [🔒 Security Features](docs/guide/nonce.md)
- [🔄 OAuth2 Guide](docs/guide/oauth2.md)
- [📊 API Reference](docs/api/stputil.md)

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by [sa-token](https://github.com/dromara/sa-token) - A powerful Java authentication framework
- Built with ❤️ using Go

## 📞 Support

- 📧 Email: support@sa-token-go.dev
- 💬 Issues: [GitHub Issues](https://github.com/click33/sa-token-go/issues)
- 📖 Documentation: [docs/](docs/)

---

**Made with ❤️ for the Go community**