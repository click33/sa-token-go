# 注解装饰器示例

本示例演示如何在 Gin 框架中使用 Sa-Token-Go 的注解装饰器（类似 Java 的 `@SaCheckLogin`、`@SaCheckRole` 等）。

## 运行示例

```bash
go run main.go
```

服务器将在 `http://localhost:8080` 启动。

## 注解装饰器

Sa-Token-Go 提供了类似 Java 注解的装饰器函数：

### CheckLogin - 检查登录

```go
r.GET("/user/info", sagin.CheckLogin(), handler.GetUserInfo)
```

### CheckRole - 检查角色

```go
r.GET("/manager", sagin.CheckRole("admin"), handler.GetManagerData)
```

### CheckPermission - 检查权限

```go
// 单个权限
r.GET("/admin", sagin.CheckPermission("admin:*"), handler.GetAdminData)

// 多个权限（OR 逻辑）
r.GET("/user-or-admin", 
    sagin.CheckPermission("user:read", "admin:*"), 
    handler.GetUserOrAdmin)
```

### CheckDisable - 检查是否被封禁

```go
r.GET("/sensitive", sagin.CheckDisable(), handler.GetSensitiveData)
```

### Ignore - 忽略认证

```go
r.GET("/public", sagin.Ignore(), handler.GetPublic)
```

## 完整示例

```go
package main

import (
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/click33/sa-token-go/core"
    sagin "github.com/click33/sa-token-go/integrations/gin"
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
    r := gin.Default()
    
    // 登录接口（公开）
    r.POST("/login", loginHandler)
    
    // 使用注解装饰器
    handler := &UserHandler{}
    
    // 公开访问
    r.GET("/public", sagin.Ignore(), handler.GetPublic)
    
    // 需要登录
    r.GET("/user/info", sagin.CheckLogin(), handler.GetUserInfo)
    
    // 需要管理员权限
    r.GET("/admin", sagin.CheckPermission("admin:*"), handler.GetAdminData)
    
    // 需要管理员角色
    r.GET("/manager", sagin.CheckRole("admin"), handler.GetManagerData)
    
    // 检查账号是否被封禁
    r.GET("/sensitive", sagin.CheckDisable(), handler.GetSensitiveData)
    
    r.Run(":8080")
}

func loginHandler(c *gin.Context) {
    var req struct {
        UserID int `json:"userId"`
    }
    c.ShouldBindJSON(&req)
    
    // 登录
    token, _ := stputil.Login(req.UserID)
    
    // 设置权限和角色
    stputil.SetPermissions(req.UserID, []string{"user:read", "admin:*"})
    stputil.SetRoles(req.UserID, []string{"admin"})
    
    c.JSON(http.StatusOK, gin.H{
        "token": token,
        "message": "登录成功",
    })
}
```

## API 测试

### 1. 登录

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"userId": 1000}'
```

响应：
```json
{
  "token": "YOUR_TOKEN",
  "message": "登录成功"
}
```

### 2. 访问公开接口（无需登录）

```bash
curl http://localhost:8080/public
```

### 3. 访问需要登录的接口

```bash
curl http://localhost:8080/user/info \
  -H "Authorization: YOUR_TOKEN"
```

### 4. 访问需要权限的接口

```bash
curl http://localhost:8080/admin \
  -H "Authorization: YOUR_TOKEN"
```

### 5. 封禁账号

```bash
curl -X POST http://localhost:8080/disable \
  -H "Content-Type: application/json" \
  -H "Authorization: YOUR_TOKEN" \
  -d '{"userId": 1000}'
```

## 注解对比

### Java (Sa-Token)

```java
@SaCheckLogin
@GetMapping("/user/info")
public Result getUserInfo() {
    return Result.success();
}

@SaCheckRole("admin")
@GetMapping("/admin")
public Result getAdminData() {
    return Result.success();
}
```

### Go (Sa-Token-Go)

```go
r.GET("/user/info", sagin.CheckLogin(), handler.GetUserInfo)

r.GET("/admin", sagin.CheckRole("admin"), handler.GetAdminData)
```

## 优势

- ✅ **声明式编程** - 代码更简洁、可读性更强
- ✅ **统一验证** - 自动处理认证和授权逻辑
- ✅ **错误处理** - 自动返回标准错误响应
- ✅ **灵活组合** - 可以组合使用多个装饰器

## 更多示例

- [快速开始](../../quick-start/simple-example) - 学习基础用法
- [Gin 集成](../../gin/gin-example) - 完整的 Gin 集成示例

