# Fiber 框架集成示例

本示例演示如何在 Fiber 框架中使用 Sa-Token-Go。

## 快速开始

### 安装依赖

```bash
go mod download
```

### 运行示例

```bash
go run cmd/main.go
```

服务器将在 `http://localhost:8080` 启动。

## 使用方式

### 方式一：使用 Manager 实例（本示例采用）

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/click33/sa-token-go/core"
    safiber "github.com/click33/sa-token-go/integrations/fiber"
    "github.com/click33/sa-token-go/storage/memory"
)

func main() {
    // 创建 Manager
    manager := core.NewBuilder().
        Storage(memory.NewStorage()).
        Build()
    
    // 创建插件
    plugin := safiber.NewPlugin(manager)
    
    // 创建 Fiber 应用
    app := fiber.New()
    app.Post("/login", plugin.LoginHandler)
    
    // 受保护路由组
    api := app.Group("/api")
    api.Use(plugin.AuthMiddleware())
    api.Get("/user/info", userInfoHandler)
    
    app.Listen(":8080")
}
```

### 方式二：使用 StpUtil 全局工具类

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/storage/memory"
)

func init() {
    stputil.SetManager(
        core.NewBuilder().
            Storage(memory.NewStorage()).
            Build(),
    )
}

func main() {
    app := fiber.New()
    
    // 登录接口
    app.Post("/login", func(c *fiber.Ctx) error {
        var req struct {
            UserID int `json:"userId"`
        }
        c.BodyParser(&req)
        
        token, _ := stputil.Login(req.UserID)
        
        return c.JSON(fiber.Map{
            "token": token,
        })
    })
    
    // 受保护路由组
    api := app.Group("/api")
    api.Use(func(c *fiber.Ctx) error {
        token := c.Get("Authorization")
        if !stputil.IsLogin(token) {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "未登录",
            })
        }
        return c.Next()
    })
    
    api.Get("/user/info", func(c *fiber.Ctx) error {
        token := c.Get("Authorization")
        loginID, _ := stputil.GetLoginID(token)
        
        return c.JSON(fiber.Map{
            "loginId": loginID,
        })
    })
    
    app.Listen(":8080")
}
```

## API 端点

- `POST /login` - 用户登录
- `GET /public` - 公开访问
- `GET /api/user/info` - 获取用户信息（需要登录）

## 测试

```bash
# 登录
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'

# 访问受保护接口
curl http://localhost:8080/api/user/info \
  -H "Authorization: YOUR_TOKEN"
```

## 性能优势

Fiber 是基于 Fasthttp 构建的高性能 Web 框架，非常适合对性能有严格要求的应用场景。

## 特性

- ✅ Fiber 中间件集成
- ✅ 高性能请求处理
- ✅ 登录认证
- ✅ 权限验证
- ✅ 角色管理

