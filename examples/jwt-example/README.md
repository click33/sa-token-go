# JWT Token 示例

本示例演示如何在 Sa-Token-Go 中使用 JWT（JSON Web Token）。

## JWT 简介

JWT 是一种无状态的 Token 方案，Token 本身包含了用户信息和过期时间，适合分布式系统。

### JWT 优势

- ✅ **无状态**：不需要服务端存储 Session
- ✅ **分布式友好**：多个服务可以独立验证
- ✅ **信息自包含**：Token 包含用户信息
- ✅ **跨域支持**：可以跨不同域使用

### JWT 结构

JWT 由三部分组成，用 `.` 分隔：

```
Header.Payload.Signature
```

- **Header**：Token 类型和加密算法
- **Payload**：用户数据（loginId, device, exp等）
- **Signature**：签名（使用密钥加密）

## 运行示例

```bash
go run main.go
```

## 基本使用

### 1. 配置 JWT

```go
import (
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/storage/memory"
)

func init() {
    stputil.SetManager(
        core.NewBuilder().
            Storage(memory.NewStorage()).
            TokenStyle(core.TokenStyleJWT).                    // 使用 JWT
            JwtSecretKey("your-256-bit-secret-key-here").    // 设置密钥
            Timeout(3600).                                     // 过期时间
            Build(),
    )
}
```

### 2. 登录获取 JWT Token

```go
token, _ := stputil.Login(1000)
// 返回类似：eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 3. 验证 JWT Token

```go
// 验证 Token 是否有效
if stputil.IsLogin(token) {
    fmt.Println("Token 有效")
}

// 获取登录 ID
loginID, _ := stputil.GetLoginID(token)
```

### 4. 解析 JWT

你可以使用 [jwt.io](https://jwt.io) 在线解析 JWT Token 查看内容。

**Payload 示例：**

```json
{
  "loginId": "1000",
  "device": "",
  "iat": 1697234567,
  "exp": 1697238167
}
```

## JWT 配置选项

```go
core.NewBuilder().
    TokenStyle(core.TokenStyleJWT).           // 启用 JWT
    JwtSecretKey("your-secret-key").         // 密钥（必需）
    Timeout(3600).                            // Token 过期时间（秒）
    IsPrintBanner(true).                      // 显示启动 Banner
    Build()
```

## 安全建议

### 1. 使用强密钥

```go
// ❌ 弱密钥
JwtSecretKey("secret")

// ✅ 强密钥（建议至少 32 字节）
JwtSecretKey("a-very-long-and-random-secret-key-at-least-256-bits")
```

### 2. 设置合理的过期时间

```go
// 短期 Token（推荐）
Timeout(3600)  // 1小时

// 长期 Token（需要配合刷新机制）
Timeout(86400) // 24小时
```

### 3. 在生产环境中保护密钥

```go
// ✅ 从环境变量读取
import "os"

JwtSecretKey(os.Getenv("JWT_SECRET_KEY"))
```

## JWT vs 普通 Token

| 特性 | JWT | UUID/Random |
|------|-----|-------------|
| 状态 | 无状态 | 有状态 |
| 服务端存储 | 不需要 | 需要 |
| Token 大小 | 较大 | 较小 |
| 可撤销性 | 困难 | 容易 |
| 分布式 | 优秀 | 需要共享存储 |
| 性能 | 高（不查数据库） | 中等（需查数据库） |

## 使用场景

### 适合 JWT 的场景

- ✅ 微服务架构
- ✅ 无状态 API
- ✅ 跨域认证
- ✅ 短期访问令牌

### 不适合 JWT 的场景

- ❌ 需要立即撤销 Token
- ❌ Token 包含敏感信息
- ❌ 需要频繁更新权限

## 完整示例

```go
package main

import (
    "fmt"
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/storage/memory"
)

func main() {
    // 初始化 JWT
    stputil.SetManager(
        core.NewBuilder().
            Storage(memory.NewStorage()).
            TokenStyle(core.TokenStyleJWT).
            JwtSecretKey("your-256-bit-secret").
            Timeout(3600).
            Build(),
    )

    // 登录
    token, _ := stputil.Login(1000)
    fmt.Println("Token:", token)

    // 验证
    if stputil.IsLogin(token) {
        loginID, _ := stputil.GetLoginID(token)
        fmt.Println("登录ID:", loginID)
    }

    // 权限管理
    stputil.SetPermissions(1000, []string{"admin:*"})
    if stputil.HasPermission(1000, "admin:read") {
        fmt.Println("有权限")
    }
}
```

## 相关文档

- [Authentication Guide](../../docs/guide/authentication.md)
- [Token Configuration](../../docs/guide/configuration.md)
- [Quick Start](../../docs/tutorial/quick-start.md)

## 在线工具

- [JWT.io](https://jwt.io) - JWT 调试工具
- [JWT Inspector](https://jwt-inspector.netlify.app/) - JWT 检查器

