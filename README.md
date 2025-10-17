# Sa-Token-Go

**[中文文档](README_zh.md)** | **English**

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.21-blue)]()
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](https://opensource.org/licenses/Apache-2.0)

A lightweight, high-performance authentication and authorization framework for Go, inspired by [sa-token](https://github.com/dromara/sa-token).

## ✨ Features

- 🔐 **Authentication** - Multi-device login, Token management
- 🛡️ **Permission** - Fine-grained permission control, wildcard support
- 👥 **Role** - Flexible role-based authorization
- 🚫 **Disable** - Temporary/permanent account disabling
- 👢 **Kickout** - Force user logout
- 💾 **Session** - Complete session management
- 🎨 **Annotations** - @SaCheckLogin, @SaCheckRole, @SaCheckPermission
- 🎧 **Events** - Powerful event system with priority and async support
- 📦 **Modular** - Import only what you need
- 🔒 **Nonce Anti-Replay** - Prevent replay attacks with one-time tokens
- 🔄 **Refresh Token** - Token refresh mechanism for seamless renewal
- 🔐 **OAuth2** - Complete OAuth2 authorization code flow implementation

## 🎨 Token Styles

Sa-Token-Go supports 9 token generation styles:

| Style | Format | Length | Use Case |
|-------|--------|--------|----------|
| **UUID** | `550e8400-e29b-41d4-...` | 36 | General purpose |
| **Simple** | `aB3dE5fG7hI9jK1l` | 16 | Compact tokens |
| **Random32/64/128** | Random string | 32/64/128 | High security |
| **JWT** | `eyJhbGciOiJIUzI1...` | Variable | Stateless auth |
| **Hash** 🆕 | `a3f5d8b2c1e4f6a9...` | 64 | SHA256 hash-based |
| **Timestamp** 🆕 | `1700000000123_user1000_...` | Variable | Time-traceable |
| **Tik** 🆕 | `7Kx9mN2pQr4` | 11 | Short ID (like TikTok) |

[👉 View Token Styles Example](examples/token-styles/)

## 🔒 Security Features

### Nonce Anti-Replay Attack

```go
// Generate nonce
nonce, _ := stputil.GenerateNonce()

// Verify nonce (one-time use)
valid := stputil.VerifyNonce(nonce)  // true
valid = stputil.VerifyNonce(nonce)   // false (replay prevented)
```

### Refresh Token Mechanism

```go
// Login with refresh token
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

// Exchange code for token
accessToken, _ := oauth2Server.ExchangeCodeForToken(
    authCode.Code, "webapp", "secret123", "http://localhost:8080/callback",
)
```

[👉 View Complete OAuth2 Example](examples/oauth2-example/)

## 🚀 快速开始

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

### 最简使用（一行初始化）

```go
package main

import (
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/storage/memory"
)

func init() {
    // 一行初始化！显示启动 Banner
    stputil.SetManager(
        core.NewBuilder().
            Storage(memory.NewStorage()).
            TokenName("Authorization").
            Timeout(86400).                      // 24小时
            TokenStyle(core.TokenStyleRandom64). // Token风格
            IsPrintBanner(true).                 // 显示启动Banner
            Build(),
    )
}

// 启动时会显示 Banner：
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
    // 直接使用 StpUtil
    token, _ := stputil.Login(1000)
    stputil.SetPermissions(1000, []string{"user:read"})
    hasPermission := stputil.HasPermission(1000, "user:read")
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
│   ├── errors.go                  # Error definitions
│   └── satoken.go                 # Core exports
│
├── stputil/                       # 🔧 Global utility (recommended)
│   └── stputil.go                # StpUtil.Login(), StpUtil.Logout()...
│
├── storage/                       # 💾 Storage backends
│   ├── memory/                    # Memory storage (development)
│   │   └── memory.go
│   └── redis/                     # Redis storage (production)
│       └── redis.go
│
├── integrations/                  # 🌐 Framework integrations
│   ├── gin/                       # Gin framework (with annotations)
│   │   ├── context.go
│   │   ├── plugin.go
│   │   └── annotation.go
│   ├── echo/                      # Echo framework
│   │   ├── context.go
│   │   └── plugin.go
│   ├── fiber/                     # Fiber framework
│   │   ├── context.go
│   │   └── plugin.go
│   └── chi/                       # Chi framework
│       ├── context.go
│       └── plugin.go
│
├── examples/                      # 📚 Example projects
│   ├── quick-start/
│   │   └── simple-example/       # ⚡ Quick start
│   ├── annotation/
│   │   └── annotation-example/   # 🎨 Annotation usage
│   ├── jwt-example/              # 🔑 JWT token example
│   ├── redis-example/            # 💾 Redis storage example
│   ├── listener-example/         # 🎧 Event listener example
│   ├── gin/gin-example/          # Gin integration
│   ├── echo/echo-example/        # Echo integration
│   ├── fiber/fiber-example/      # Fiber integration
│   └── chi/chi-example/          # Chi integration
│
├── docs/                          # 📖 Documentation
│   ├── tutorial/                  # Tutorials
│   │   └── quick-start.md
│   ├── guide/                     # Guides
│   │   ├── authentication.md
│   │   ├── permission.md
│   │   ├── annotation.md
│   │   ├── listener.md
│   │   ├── jwt.md
│   │   ├── redis-storage.md      # English
│   │   └── redis-storage_zh.md   # Chinese
│   ├── api/                       # API docs
│   └── design/                    # Design docs
│
├── go.work                        # Go workspace
├── README.md                      # English README
└── README_zh.md                   # Chinese README
```

## ⚙️ 配置选项

### Token 读取位置

默认只从 **Header** 读取 Token（推荐）：

```go
core.NewBuilder().
    IsReadHeader(true).   // 从 Header 读取（默认：true，推荐）
    IsReadCookie(false).  // 从 Cookie 读取（默认：false）
    IsReadBody(false).    // 从 Body 读取（默认：false）
    Build()
```

**Token 读取优先级：** Header > Cookie > Body

**推荐配置：** 只启用 `IsReadHeader`，Token 放在 HTTP Header 中：
```
Authorization: your-token-here
```

### JWT Token 支持

```go
// 使用 JWT Token
stputil.SetManager(
    core.NewBuilder().
        Storage(memory.NewStorage()).
        TokenStyle(core.TokenStyleJWT).              // 使用 JWT
        JwtSecretKey("your-256-bit-secret").       // JWT 密钥
        Timeout(3600).                               // 1小时过期
        Build(),
)

// 登录后获得 JWT Token
token, _ := stputil.Login(1000)
// 返回格式：eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

// JWT Token 包含用户信息，可在 https://jwt.io 解析
```

**支持的 Token 风格：**
- `TokenStyleUUID` - UUID（默认）
- `TokenStyleSimple` - 简单随机字符串
- `TokenStyleRandom32/64/128` - 指定长度随机串
- `TokenStyleJWT` - JWT Token（推荐用于分布式）

### 启动 Banner

```go
core.NewBuilder().
    IsPrintBanner(true).  // 显示启动 Banner（默认：true）
    Build()
```

关闭 Banner：
```go
core.NewBuilder().
    IsPrintBanner(false).  // 不显示 Banner
    Build()
```

## 📚 Documentation

### Language
- [中文文档 (Chinese)](README_zh.md)
- [English Documentation](README.md)

### Tutorials & Guides
- [Quick Start](docs/tutorial/quick-start.md) - Get started in 5 minutes
- [Authentication Guide](docs/guide/authentication.md) - Login, logout, and session management
- [Permission Management](docs/guide/permission.md) - Fine-grained permission control
- [Annotation Usage](docs/guide/annotation.md) - Decorator pattern for route protection
- [Event Listener](docs/guide/listener.md) - Event system for audit and monitoring
- [JWT Guide](docs/guide/jwt.md) - JWT token configuration and usage
- [Redis Storage](docs/guide/redis-storage.md) - Production-ready Redis backend

### API Documentation
- [StpUtil API](docs/api/stputil.md) - Complete global utility API reference

### Design Documentation
- [Architecture Design](docs/design/architecture.md) - System architecture and data flow
- [Auto-Renewal Design](docs/design/auto-renew.md) - Asynchronous renewal mechanism
- [Modular Design](docs/design/modular.md) - Module organization strategy

### Storage
- [Memory Storage](storage/memory/) - For development
- [Redis Storage](storage/redis/) - For production

## 🔧 核心API

```go
// 登录认证
stputil.Login(loginID)
stputil.Logout(loginID)
stputil.IsLogin(token)
stputil.GetLoginID(token)

// 权限验证
stputil.SetPermissions(loginID, []string{"user:read"})
stputil.HasPermission(loginID, "user:read")

// 角色管理
stputil.SetRoles(loginID, []string{"admin"})
stputil.HasRole(loginID, "admin")

// 账号封禁
stputil.Disable(loginID, time.Hour)
stputil.IsDisable(loginID)

// Session管理
sess, _ := stputil.GetSession(loginID)
sess.Set("key", "value")
```

## 📖 Examples

Check out the [examples](examples/) directory:

| Example | Description | Path |
|---------|-------------|------|
| ⚡ Quick Start | Minimal setup with Builder & StpUtil | [examples/quick-start/](examples/quick-start/) |
| 🎨 Token Styles | All 9 token generation styles | [examples/token-styles/](examples/token-styles/) |
| 🔒 Security Features | Nonce/RefreshToken/OAuth2 | [examples/security-features/](examples/security-features/) |
| 🔐 OAuth2 Example | Complete OAuth2 authorization flow | [examples/oauth2-example/](examples/oauth2-example/) |
| 📝 Annotations | Decorator pattern usage | [examples/annotation/](examples/annotation/) |
| 🔑 JWT Example | JWT token configuration | [examples/jwt-example/](examples/jwt-example/) |
| 💾 Redis Example | Redis storage setup | [examples/redis-example/](examples/redis-example/) |
| 🎧 Event Listener | Event system usage | [examples/listener-example/](examples/listener-example/) |
| 🌐 Gin Integration | Gin framework integration | [examples/gin/](examples/gin/) |
| 🌐 Echo Integration | Echo framework integration | [examples/echo/](examples/echo/) |
| 🌐 Fiber Integration | Fiber framework integration | [examples/fiber/](examples/fiber/) |
| 🌐 Chi Integration | Chi framework integration | [examples/chi/](examples/chi/) |

## 📄 许可证

Apache License 2.0

## 🙏 致谢

参考 [sa-token](https://github.com/dromara/sa-token) 设计

---

**Sa-Token-Go v0.1.0**

