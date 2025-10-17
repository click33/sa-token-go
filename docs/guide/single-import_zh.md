# 单一导入使用指南

**[English](single-import.md)**

## 概述

从 v0.1.0 开始，Sa-Token-Go 支持**单一导入模式** - 您只需要导入一个框架集成包，就能访问 core 和 stputil 的所有功能。

## 优势

✅ **更简单的依赖** - 只需导入一个包  
✅ **更清晰的代码** - 更少的导入语句  
✅ **更好的 IDE 支持** - 所有函数在一个命名空间  
✅ **向后兼容** - 旧的导入方式仍然有效  

## 传统方式 vs. 新方式

### 传统方式（多个导入）

```go
import (
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/integrations/gin"
    "github.com/click33/sa-token-go/storage/memory"
)

func main() {
    // 使用 core
    config := core.DefaultConfig()
    storage := memory.NewStorage()
    manager := core.NewManager(storage, config)
    
    // 使用 stputil
    stputil.SetManager(manager)
    token, _ := stputil.Login(1000)
    
    // 使用 gin 集成
    plugin := gin.NewPlugin(manager)
}
```

### 新方式（单一导入）✨

```go
import (
    sagin "github.com/click33/sa-token-go/integrations/gin"
    "github.com/click33/sa-token-go/storage/memory"
)

func main() {
    // 所有功能都在 sagin 包中！
    config := sagin.DefaultConfig()
    storage := memory.NewStorage()
    manager := sagin.NewManager(storage, config)
    
    sagin.SetManager(manager)
    token, _ := sagin.Login(1000)
    
    plugin := sagin.NewPlugin(manager)
}
```

## 安装

### Gin 框架

```bash
go get github.com/click33/sa-token-go/integrations/gin@v0.1.0
go get github.com/click33/sa-token-go/storage/memory@v0.1.0
```

### Echo 框架

```bash
go get github.com/click33/sa-token-go/integrations/echo@v0.1.0
go get github.com/click33/sa-token-go/storage/memory@v0.1.0
```

### Fiber 框架

```bash
go get github.com/click33/sa-token-go/integrations/fiber@v0.1.0
go get github.com/click33/sa-token-go/storage/memory@v0.1.0
```

### Chi 框架

```bash
go get github.com/click33/sa-token-go/integrations/chi@v0.1.0
go get github.com/click33/sa-token-go/storage/memory@v0.1.0
```

## 完整示例（Gin）

```go
package main

import (
    "log"

    "github.com/gin-gonic/gin"
    sagin "github.com/click33/sa-token-go/integrations/gin"
    "github.com/click33/sa-token-go/storage/memory"
)

func main() {
    // 1. 初始化存储
    storage := memory.NewStorage()

    // 2. 创建配置（来自 sagin 包）
    config := sagin.DefaultConfig()
    config.TokenName = "token"
    config.Timeout = 7200  // 2小时
    config.IsPrint = true

    // 3. 创建管理器（来自 sagin 包）
    manager := sagin.NewManager(storage, config)

    // 4. 设置全局管理器（来自 sagin 包）
    sagin.SetManager(manager)

    // 5. 创建 Gin 路由器
    r := gin.Default()

    // 6. 登录接口
    r.POST("/login", func(c *gin.Context) {
        userID := c.PostForm("user_id")
        if userID == "" {
            c.JSON(400, gin.H{"error": "需要 user_id"})
            return
        }

        // 使用 sagin.Login（来自 sagin 包）
        token, err := sagin.Login(userID)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, gin.H{
            "message": "登录成功",
            "token":   token,
        })
    })

    // 7. 登出接口
    r.POST("/logout", func(c *gin.Context) {
        token := c.GetHeader("token")
        if token == "" {
            c.JSON(400, gin.H{"error": "需要 token"})
            return
        }

        // 使用 sagin.LogoutByToken（来自 sagin 包）
        if err := sagin.LogoutByToken(token); err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, gin.H{"message": "登出成功"})
    })

    // 8. 检查登录状态
    r.GET("/check", func(c *gin.Context) {
        token := c.GetHeader("token")
        if token == "" {
            c.JSON(400, gin.H{"error": "需要 token"})
            return
        }

        // 使用 sagin.IsLogin（来自 sagin 包）
        isLogin := sagin.IsLogin(token)
        if !isLogin {
            c.JSON(401, gin.H{"error": "未登录"})
            return
        }

        // 使用 sagin.GetLoginID（来自 sagin 包）
        loginID, _ := sagin.GetLoginID(token)

        c.JSON(200, gin.H{
            "message":  "已登录",
            "login_id": loginID,
        })
    })

    // 9. 使用注解的受保护路由
    plugin := sagin.NewPlugin(manager)
    protected := r.Group("/api")
    protected.Use(plugin.AuthMiddleware())
    {
        protected.GET("/user", func(c *gin.Context) {
            token := c.GetHeader("token")
            loginID, _ := sagin.GetLoginID(token)

            c.JSON(200, gin.H{
                "user_id": loginID,
                "name":    "用户 " + loginID,
            })
        })

        // 踢人下线
        protected.POST("/kickout/:user_id", func(c *gin.Context) {
            userID := c.Param("user_id")

            // 使用 sagin.Kickout（来自 sagin 包）
            if err := sagin.Kickout(userID); err != nil {
                c.JSON(500, gin.H{"error": err.Error()})
                return
            }

            c.JSON(200, gin.H{"message": "踢人成功"})
        })
    }

    // 10. 启动服务器
    log.Println("服务器启动在端口: 8080")
    log.Println("示例: curl -X POST http://localhost:8080/login -d 'user_id=1000'")
    if err := r.Run(":8080"); err != nil {
        log.Fatal("服务器启动失败:", err)
    }
}
```

## 可用函数

所有来自 `core` 和 `stputil` 的函数都在框架集成包中重新导出：

### 配置和初始化

```go
config := sagin.DefaultConfig()           // 创建默认配置
manager := sagin.NewManager(storage, cfg) // 创建管理器
builder := sagin.NewBuilder()             // 创建构建器
sagin.SetManager(manager)                 // 设置全局管理器
manager := sagin.GetManager()             // 获取全局管理器
```

### 认证

```go
token, _ := sagin.Login(loginID, device...)
sagin.LoginByToken(loginID, token, device...)
sagin.Logout(loginID, device...)
sagin.LogoutByToken(token)
isLogin := sagin.IsLogin(token)
sagin.CheckLogin(token)
loginID, _ := sagin.GetLoginID(token)
tokenValue, _ := sagin.GetTokenValue(loginID, device...)
tokenInfo, _ := sagin.GetTokenInfo(token)
```

### 踢人下线 & 封禁

```go
sagin.Kickout(loginID, device...)
sagin.Disable(loginID, duration)
isDisabled := sagin.IsDisable(loginID)
sagin.CheckDisable(loginID)
remainTime, _ := sagin.GetDisableTime(loginID)
sagin.Untie(loginID)
```

### 权限 & 角色

```go
sagin.CheckPermission(loginID, permission)
hasPermission := sagin.HasPermission(loginID, permission)
sagin.CheckPermissionAnd(loginID, perms...)
sagin.CheckPermissionOr(loginID, perms...)
permissions := sagin.GetPermissionList(loginID)

sagin.CheckRole(loginID, role)
hasRole := sagin.HasRole(loginID, role)
sagin.CheckRoleAnd(loginID, roles...)
sagin.CheckRoleOr(loginID, roles...)
roles := sagin.GetRoleList(loginID)
```

### Session 管理

```go
session, _ := sagin.GetSession(loginID)
session, _ := sagin.GetSessionByToken(token)
tokenSession, _ := sagin.GetTokenSession(token)
```

### 安全特性

```go
nonce, _ := sagin.GenerateNonce()
sagin.VerifyNonce(nonce)
accessToken, refreshToken, _ := sagin.LoginWithRefreshToken(loginID, device...)
newAccessToken, newRefreshToken, _ := sagin.RefreshAccessToken(refreshToken)
sagin.RevokeRefreshToken(refreshToken)
oauth2Server := sagin.GetOAuth2Server()
```

### Token & 工具函数

```go
sagin.RenewTimeout(token)
randomStr := sagin.RandomString(16)
isEmpty := sagin.IsEmpty(str)
matched := sagin.MatchPattern(pattern, str)
```

## 类型定义

所有来自 `core` 的类型也都被导出：

```go
type Config = sagin.Config
type Manager = sagin.Manager
type Session = sagin.Session
type TokenInfo = sagin.TokenInfo
type Storage = sagin.Storage
type EventListener = sagin.EventListener
// ... 更多
```

## 常量

所有常量都可用：

```go
sagin.TokenStyleUUID
sagin.TokenStyleJWT
sagin.TokenStyleHash
// ... 更多

sagin.EventLogin
sagin.EventLogout
sagin.EventKickout
// ... 更多
```

## 框架特定示例

### Echo 示例

```go
import (
    saecho "github.com/click33/sa-token-go/integrations/echo"
    "github.com/labstack/echo/v4"
)

func main() {
    config := saecho.DefaultConfig()
    manager := saecho.NewManager(storage, config)
    saecho.SetManager(manager)
    
    e := echo.New()
    
    e.POST("/login", func(c echo.Context) error {
        token, _ := saecho.Login(userID)
        return c.JSON(200, map[string]string{"token": token})
    })
    
    e.Start(":8080")
}
```

### Fiber 示例

```go
import (
    safiber "github.com/click33/sa-token-go/integrations/fiber"
    "github.com/gofiber/fiber/v2"
)

func main() {
    config := safiber.DefaultConfig()
    manager := safiber.NewManager(storage, config)
    safiber.SetManager(manager)
    
    app := fiber.New()
    
    app.Post("/login", func(c *fiber.Ctx) error {
        token, _ := safiber.Login(userID)
        return c.JSON(fiber.Map{"token": token})
    })
    
    app.Listen(":8080")
}
```

### Chi 示例

```go
import (
    sachi "github.com/click33/sa-token-go/integrations/chi"
    "github.com/go-chi/chi/v5"
)

func main() {
    config := sachi.DefaultConfig()
    manager := sachi.NewManager(storage, config)
    sachi.SetManager(manager)
    
    r := chi.NewRouter()
    
    r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
        token, _ := sachi.Login(userID)
        json.NewEncoder(w).Encode(map[string]string{"token": token})
    })
    
    http.ListenAndServe(":8080", r)
}
```

## 从旧的导入方式迁移

如果您有使用旧导入方式的现有代码，可以逐步迁移：

### 步骤 1: 添加新的导入（保留旧的）

```go
import (
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    sagin "github.com/click33/sa-token-go/integrations/gin"  // 添加这个
    "github.com/click33/sa-token-go/storage/memory"
)
```

### 步骤 2: 替换函数调用

```go
// 旧
config := core.DefaultConfig()

// 新
config := sagin.DefaultConfig()
```

### 步骤 3: 删除旧的导入

```go
import (
    // 删除这些
    // "github.com/click33/sa-token-go/core"
    // "github.com/click33/sa-token-go/stputil"
    
    sagin "github.com/click33/sa-token-go/integrations/gin"
    "github.com/click33/sa-token-go/storage/memory"
)
```

## 常见问题

**Q: 我还需要导入 core 或 stputil 吗？**  
A: 不需要，框架集成包已经包含了它们。

**Q: 我可以混合使用旧的和新的导入方式吗？**  
A: 可以，但不推荐。为了保持一致性，请选择一种方式。

**Q: 这对所有框架都有效吗？**  
A: 是的，Gin、Echo、Fiber 和 Chi 都支持单一导入。

**Q: 这会有性能影响吗？**  
A: 没有，这只是重新导出。没有额外的开销。

**Q: 如果我不使用任何 Web 框架呢？**  
A: 那么使用传统的导入方式，导入 `core` 和 `stputil`。

## 了解更多

- [完整示例](../../examples/gin/gin-simple/)
- [主文档](../../README_zh.md)
- [API 参考](../api/api_zh.md)

