# Sa-Token-Go

**中文文档** | **[English](README.md)**

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.25-blue)](https://img.shields.io)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](https://opensource.org/licenses/Apache-2.0)

一个轻量级、高性能的 Go 权限认证框架，参考 [sa-token](https://github.com/dromara/sa-token) 设计。

## ✨ 核心特性

- 🔐 **登录认证** - 支持多设备登录、Token 管理、登录态校验
- 🛡️ **权限验证** - 细粒度权限控制、通配符支持（`*`、`user:*`、`user:*:view`）
- 👥 **角色管理** - 灵活的角色授予、移除与组合校验
- 🚫 **账号封禁** - 支持封禁、解封、封禁信息查询与服务级封禁
- 👢 **踢人下线** - 支持按 Token、账号、设备维度踢人下线或顶号
- 💾 **Session 会话** - 支持按账号或 Token 读取 Session 数据
- ⏰ **活跃检测** - 支持 `ActiveTimeout` 与自动续期配置
- 🔄 **自动续期** - 内置活跃续签与续签池能力，兼顾性能与体验
- 🎨 **注解支持** - 在各集成包中提供 `CheckLoginMiddleware`、`CheckRoleMiddleware`、`CheckPermissionMiddleware` 等校验中间件
- 🎧 **事件监听** - 内置登录、登出、续期、封禁、权限校验、角色校验等事件，支持优先级、过滤器、统计
- 📦 **模块化设计** - `core`、`stputil`、`com/*`、`integrations/*`、`examples/*` 分层清晰
- 🔒 **Nonce 防重放** - 支持生成、校验与一次性消费
- 🔄 **OAuth2 / Refresh Token** - 已包含 OAuth2 授权码、刷新令牌等实现
- 🧩 **组件可替换** - Codec、Generator、Log、Pool、Storage 都可以按需替换

## 🚀 快速开始

### 📥 安装

#### 方式一：简化导入（推荐）✨

**只需导入一个框架集成包，再按需引入存储模块即可。集成包本身已经依赖 `core` 与 `stputil`。**

```bash
# 只导入框架集成包
go get github.com/click33/sa-token-go/integrations/gin@latest # Gin 集成，适合 Gin 项目直接接入
# 或
go get github.com/click33/sa-token-go/integrations/echo@latest # Echo 集成，适合 Echo 项目直接接入
# 或
go get github.com/click33/sa-token-go/integrations/fiber@latest # Fiber 集成，适合 Fiber 项目直接接入
# 或
go get github.com/click33/sa-token-go/integrations/chi@latest # Chi 集成，适合 Chi 项目直接接入
# 或
go get github.com/click33/sa-token-go/integrations/gf@latest # GoFrame 集成，适合 GoFrame 项目直接接入
# 或
go get github.com/click33/sa-token-go/integrations/hertz@latest # Hertz 集成，适合 Hertz 项目直接接入
# 或
go get github.com/click33/sa-token-go/integrations/kratos@latest # Kratos 集成，适合 Kratos 项目直接接入

# 存储模块（选一个）
go get github.com/click33/sa-token-go/com/storage/memory@latest # 内存存储，适合本地开发和简单测试
go get github.com/click33/sa-token-go/com/storage/redis@latest # Redis 存储，适合生产环境或多实例部署
```

#### 方式二：分开导入

```bash
# 核心模块
go get github.com/click33/sa-token-go/core@latest     # 核心能力、Builder、Manager、配置等
go get github.com/click33/sa-token-go/stputil@latest  # 全局认证工具入口，业务层通常直接使用它

# 组件模块
go get github.com/click33/sa-token-go/com/storage/memory@latest # 内存存储，适合本地开发和简单测试
go get github.com/click33/sa-token-go/com/storage/redis@latest  # Redis 存储，适合生产环境或多实例部署

# 框架集成（可选）
go get github.com/click33/sa-token-go/integrations/gin@latest    # Gin 集成
go get github.com/click33/sa-token-go/integrations/echo@latest   # Echo 集成
go get github.com/click33/sa-token-go/integrations/fiber@latest  # Fiber 集成
go get github.com/click33/sa-token-go/integrations/chi@latest    # Chi 集成
go get github.com/click33/sa-token-go/integrations/gf@latest     # GoFrame 集成
go get github.com/click33/sa-token-go/integrations/hertz@latest  # Hertz 集成
go get github.com/click33/sa-token-go/integrations/kratos@latest # Kratos 集成
```

### ⚡ 超简洁使用（一行初始化）

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

**启动时默认会打印 Banner：**

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
    println("登录成功，Token:", token)

    _ = stputil.AddPermissions(ctx, "1000", []string{"user:read", "user:write"})

    if stputil.HasPermission(ctx, "1000", "user:read") {
        println("有权限！")
    }

    _ = stputil.Logout(ctx, token)
}
```

如果你想看更完整的启动方式和路由组织，可以直接参考当前仓库中的这些示例：

- `examples/quick_start`
- `examples/gin`
- `examples/gf`
- `examples/echo`
- `examples/fiber`
- `examples/chi`
- `examples/hertz`
- `examples/kratos`

## 🔧 核心 API

### 🔑 登录认证

```go
// 登录
token, _ := stputil.Login(ctx, "1000")
token, _ := stputil.Login(ctx, "user123")
token, _ := stputil.Login(ctx, "1000", "mobile") // 指定设备

// 指定超时时间登录
tempToken, _ := stputil.LoginWithTimeout(ctx, "1000", 2*time.Hour, "web")

// 基于已有 Token 续期登录
_ = stputil.LoginByToken(ctx, token)

// 检查登录
isLogin := stputil.IsLogin(ctx, token)

// 获取登录 ID
loginID, _ := stputil.GetLoginID(ctx, token)

// 获取 Token 信息
tokenInfo, _ := stputil.GetTokenInfo(ctx, token)

// 获取设备、设备 ID、创建时间、剩余有效期
device, _ := stputil.GetDevice(ctx, token)
deviceID, _ := stputil.GetDeviceId(ctx, token)
createTime, _ := stputil.GetTokenCreateTime(ctx, token)
ttl, _ := stputil.GetTokenTTL(ctx, token)

// 登出
_ = stputil.Logout(ctx, token)
_ = stputil.LogoutByDevice(ctx, "1000", "mobile")
_ = stputil.LogoutByLoginID(ctx, "1000")
_ = stputil.LogoutByDeviceAndDeviceId(ctx, "1000", "mobile", "device-001")

// 踢人下线 / 顶号
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

### 🛡️ 权限验证

```go
// 添加权限
_ = stputil.AddPermissions(ctx, "1000", []string{
    "user:read",
    "user:write",
    "admin:*",
})

// 检查权限
hasPermission := stputil.HasPermission(ctx, "1000", "user:read")
hasPermission = stputil.HasPermission(ctx, "1000", "admin:delete")

// 多权限检查
hasAll := stputil.HasPermissionsAnd(ctx, "1000", []string{"user:read", "user:write"})
hasAny := stputil.HasPermissionsOr(ctx, "1000", []string{"admin:*", "super:*"})

// 查询 / 删除权限
perms, _ := stputil.GetPermissions(ctx, "1000")
_ = stputil.RemovePermissions(ctx, "1000", []string{"user:write"})

// 按 Token 维度校验
hasByToken := stputil.HasPermissionByToken(ctx, token, "user:read")

_ = hasPermission
_ = hasAll
_ = hasAny
_ = perms
_ = hasByToken
```

### 👥 角色管理

```go
// 添加角色
_ = stputil.AddRoles(ctx, "1000", []string{"admin", "manager"})

// 检查角色
hasRole := stputil.HasRole(ctx, "1000", "admin")

// 多角色检查
hasAll := stputil.HasRolesAnd(ctx, "1000", []string{"admin", "manager"})
hasAny := stputil.HasRolesOr(ctx, "1000", []string{"admin", "super-admin"})

// 查询 / 删除角色
roles, _ := stputil.GetRoles(ctx, "1000")
_ = stputil.RemoveRoles(ctx, "1000", []string{"manager"})

// 按 Token 维度校验
hasRoleByToken := stputil.HasRoleByToken(ctx, token, "admin")

_ = hasRole
_ = hasAll
_ = hasAny
_ = roles
_ = hasRoleByToken
```

### 💾 Session 管理

```go
// 获取 Session
sess, _ := stputil.GetSession(ctx, "1000")

// 或通过 Token 获取 Session
sessByToken, _ := stputil.GetSessionByToken(ctx, token)

// 设置数据
sess.Set("nickname", "张三")
sess.Set("age", 25)

// 读取数据
nickname := sess.GetString("nickname")
age := sess.GetInt("age")

// 删除字段
sess.Delete("nickname")

// 获取当前账号的 Token / 终端列表
tokenList, _ := stputil.GetTokenValueListByLoginID(ctx, "1000", true)
terminalList, _ := stputil.GetTerminalListByLoginID(ctx, "1000")

// 搜索 Token / Session
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

### 🚫 账号封禁

```go
// 封禁 1 小时
_ = stputil.Disable(ctx, "1000", 1*time.Hour, "manual disable")

// 解封
_ = stputil.Untie(ctx, "1000")

// 检查是否被封禁
isDisabled := stputil.IsDisable(ctx, "1000")

// 获取封禁信息与剩余时间
disableInfo, _ := stputil.GetDisableInfo(ctx, "1000")
ttl, _ := stputil.GetDisableTTL(ctx, "1000")

// 服务级封禁
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

## 🌐 框架集成

### 🌟 Gin 集成（单一导入）

**当前推荐直接使用 `integrations/gin`，统一通过 `satoken` 别名访问构建器、中间件和上下文能力。**

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

### 🎯 注解装饰器支持

当前各集成包统一提供的是**注解式校验中间件**，命名风格如下：

| 能力 | 当前函数 |
|---|---|
| 忽略认证 | `IgnoreMiddleware(...)` |
| 检查登录 | `CheckLoginMiddleware(...)` |
| 检查角色 | `CheckRoleMiddleware(...)` |
| 检查权限 | `CheckPermissionMiddleware(...)` |
| 检查封禁 | `CheckDisableMiddleware(...)` |
| 组合校验 | `CheckAllMiddleware(...)` |

**Gin 使用示例：**

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

### 🌟 GoFrame 集成（单一导入）

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

### 🔌 其他框架集成

当前仓库已经提供这些集成包与对应示例：

- `integrations/echo` 对应 `examples/echo`
- `integrations/fiber` 对应 `examples/fiber`
- `integrations/chi` 对应 `examples/chi`
- `integrations/gin` 对应 `examples/gin`
- `integrations/gf` 对应 `examples/gf`
- `integrations/hertz` 对应 `examples/hertz`
- `integrations/kratos` 对应 `examples/kratos`

这些集成包整体都已经对齐到同一套 `SaToken` 命名，包括：

- `RegisterSaTokenContextMiddleware`
- `AuthMiddleware`
- `RoleMiddleware`
- `PermissionMiddleware`
- `CheckLoginMiddleware`
- `CheckRoleMiddleware`
- `CheckPermissionMiddleware`
- `CheckAllMiddleware`

常见框架的使用风格也已经基本统一，例如：

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

## 🚀 高级特性

### 🎨 Token 风格

当前代码支持 9 种 Token 风格：

| 风格 | 格式示例 | 长度 | 适用场景 |
|---|---|---|---|
| `TokenStyleUUID` | `550e8400-e29b-41d4-...` | 36 | 通用场景 |
| `TokenStyleSimple` | `aB3dE5fG7hI9jK1l` | 16 | 紧凑型 Token |
| `TokenStyleRandom32` | 随机字符串 | 32 | 较高安全性 |
| `TokenStyleRandom64` | 随机字符串 | 64 | 默认推荐 |
| `TokenStyleRandom128` | 随机字符串 | 128 | 高安全场景 |
| `TokenStyleJWT` | `eyJhbGciOiJIUzI1...` | 可变 | 无状态认证 |
| `TokenStyleHash` | `a3f5d8b2c1e4f6a9...` | 64 | SHA256 哈希 |
| `TokenStyleTimestamp` | `1700000000123_user1000_...` | 可变 | 便于追踪创建时间 |
| `TokenStyleTik` | `7Kx9mN2pQr4` | 11 | 短 Token 场景 |

**JWT Token 示例：**

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

如果你想快速体验不同框架下的使用方式，可以直接参考 `examples/*` 下的完整示例。

### 🔒 安全特性

#### 🔐 Nonce 防重放

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

### 🎧 事件监听

当前代码中事件监听能力已经内置在 `core/listener` 与 `manager` 中，支持以下特性：

- 支持按事件类型注册监听器
- 支持优先级控制
- 支持异步 / 同步执行
- 支持全局过滤器
- 支持触发统计与 panic 处理

当前内置事件包括：

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

详细注册方式与监听器配置请参考：

- [事件监听中文文档](docs/guide/listener_zh.md)

## 🏗️ 架构概览

当前 `sa-token-go` 的整体结构可以概括为 4 层：

- `com/*`：可替换组件层，负责存储、编解码、日志、协程池、Token 生成器等基础实现
- `core/*`：核心能力层，负责配置、上下文、Manager、事件监听、Nonce、OAuth2 等核心逻辑
- `stputil`：全局工具入口，对外提供常用认证、权限、角色、Session、封禁等能力
- `integrations/*`：框架集成层，对接 Gin、GoFrame、Echo、Fiber、Chi、Hertz、Kratos 等 Web 框架

详细设计说明可以参考：

- [架构设计文档](docs/design/architecture_zh.md)

## 📁 项目结构

```text
sa-token-go/
├── com/                    # 可替换组件模块
│   ├── codec/              # 编解码实现
│   │   ├── base64/         # Base64 编解码
│   │   ├── json/           # 标准 JSON 编解码
│   │   ├── jsonv2/         # JSON v2 编解码
│   │   └── msgpack/        # MsgPack 编解码
│   ├── generator/          # Token 生成器实现
│   │   └── sgenerator/     # 默认 Token 生成器
│   ├── log/                # 日志实现
│   │   ├── gf/             # GoFrame 日志适配
│   │   ├── nop/            # 空日志实现
│   │   └── slog/           # slog 日志适配
│   ├── pool/               # 协程池实现
│   │   └── ants/           # ants 协程池适配
│   └── storage/            # 存储实现
│       ├── memory/         # 内存存储
│       └── redis/          # Redis 存储
├── core/                   # 核心能力模块
│   ├── adapter/            # 核心适配器接口与常量
│   ├── banner/             # 启动 Banner 打印
│   ├── builder/            # Builder 构建器
│   ├── config/             # 配置定义
│   ├── context/            # SaTokenContext 上下文封装
│   ├── listener/           # 事件监听系统
│   ├── manager/            # 认证管理器核心实现
│   ├── nonce/              # Nonce 防重放实现
│   ├── oauth2/             # OAuth2 能力实现
│   ├── serror/             # 错误定义
│   └── utils/              # 通用工具函数
├── stputil/                # 全局认证工具入口
├── integrations/           # Web 框架集成层
│   ├── chi/                # Chi 集成
│   ├── echo/               # Echo 集成
│   ├── fiber/              # Fiber 集成
│   ├── gf/                 # GoFrame 集成
│   ├── gin/                # Gin 集成
│   ├── hertz/              # Hertz 集成
│   └── kratos/             # Kratos 集成
├── examples/               # 示例工程
│   ├── chi/                # Chi 示例
│   ├── echo/               # Echo 示例
│   ├── fiber/              # Fiber 示例
│   ├── gf/                 # GoFrame 示例
│   ├── gin/                # Gin 示例
│   ├── hertz/              # Hertz 示例
│   ├── kratos/             # Kratos 示例
│   └── quick_start/        # Quick Start 示例
└── docs/                   # 文档目录
```

## 📚 文档与示例

### 📖 详细文档

- [快速开始](docs/tutorial/quick-start_zh.md) - 5 分钟完成初始化
- [登录认证](docs/guide/authentication_zh.md) - 登录、登出、踢人下线、顶号
- [权限验证](docs/guide/permission_zh.md) - 权限、角色、组合校验
- [注解使用](docs/guide/annotation_zh.md) - 各框架注解式中间件说明
- [事件监听](docs/guide/listener_zh.md) - 事件系统与监听器能力
- [JWT 使用](docs/guide/jwt_zh.md) - JWT 风格 Token 说明
- [Redis 存储](docs/guide/redis-storage_zh.md) - Redis 存储与生产配置
- [Nonce 防重放](docs/guide/nonce_zh.md) - Nonce 生成与校验
- [Refresh Token](docs/guide/refresh-token_zh.md) - 刷新令牌说明
- [OAuth2](docs/guide/oauth2_zh.md) - OAuth2 授权码模式
- [单包导入](docs/guide/single-import_zh.md) - integrations 单包导入方式

### 📘 API 文档

- [StpUtil API](docs/api/stputil_zh.md) - 全局工具类完整 API 说明

### 🧠 设计文档

- [架构设计](docs/design/architecture_zh.md) - 系统模块关系与调用链路
- [自动续签设计](docs/design/auto-renew_zh.md) - 自动续签机制说明
- [模块化设计](docs/design/modular_zh.md) - 多模块拆分与依赖策略

### 🧪 示例项目

| 示例 | 说明 | 路径 |
|---|---|---|
| Quick Start | 基础快速开始示例 | [examples/quick_start/](examples/quick_start/) |
| Gin | Gin 集成示例 | [examples/gin/](examples/gin/) |
| GoFrame | GoFrame 集成示例 | [examples/gf/](examples/gf/) |
| Echo | Echo 集成示例 | [examples/echo/](examples/echo/) |
| Fiber | Fiber 集成示例 | [examples/fiber/](examples/fiber/) |
| Chi | Chi 集成示例 | [examples/chi/](examples/chi/) |
| Hertz | Hertz 集成示例 | [examples/hertz/](examples/hertz/) |
| Kratos | Kratos 集成示例 | [examples/kratos/](examples/kratos/) |

### 💾 存储方案

- [Memory 存储](com/storage/memory/)
- [Redis 存储](com/storage/redis/)

## 📄 许可证

Apache License 2.0

## 🙏 致谢

参考 [sa-token](https://github.com/dromara/sa-token) 的设计思路实现。

### 贡献者

特别感谢以下贡献者的宝贵贡献：

- [@qprodn](https://github.com/qprodn)
- [@Zany2](https://github.com/Zany2)
- [@zyw](https://github.com/zyw)
- [@nuanxinqing123](https://github.com/nuanxinqing123)
- [@vera-byte](https://github.com/vera-byte)
- [@MoLing-Dong](https://github.com/MoLing-Dong)

## 📞 支持

- 💬 问题反馈: [GitHub Issues](https://github.com/click33/sa-token-go/issues)
- 📖 文档入口: [docs/](docs/)

### 微信交流群

![sa-token-go 微信交流群二维码](docs/wechat.JPG)
