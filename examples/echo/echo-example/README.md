# Echo 框架集成示例

本示例演示如何在 Echo 框架中使用 Sa-Token-Go。

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
    "github.com/labstack/echo/v4"
    "github.com/click33/sa-token-go/core"
    saecho "github.com/click33/sa-token-go/integrations/echo"
    "github.com/click33/sa-token-go/storage/memory"
)

func main() {
    // 创建 Manager
    manager := core.NewBuilder().
        Storage(memory.NewStorage()).
        Build()
    
    // 创建插件
    plugin := saecho.NewPlugin(manager)
    
    // 创建 Echo 实例
    e := echo.New()
    e.POST("/login", plugin.LoginHandler)
    
    // 受保护路由组
    api := e.Group("/api")
    api.Use(plugin.AuthMiddleware())
    api.GET("/user/info", userInfoHandler)
    
    e.Start(":8080")
}
```

### 方式二：使用 StpUtil 全局工具类

```go
package main

import (
    "net/http"
    
    "github.com/labstack/echo/v4"
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
    e := echo.New()
    
    // 登录接口
    e.POST("/login", func(c echo.Context) error {
        var req struct {
            UserID int `json:"userId"`
        }
        c.Bind(&req)
        
        token, _ := stputil.Login(req.UserID)
        
        return c.JSON(http.StatusOK, map[string]interface{}{
            "token": token,
        })
    })
    
    // 受保护路由组
    api := e.Group("/api")
    api.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            token := c.Request().Header.Get("Authorization")
            if !stputil.IsLogin(token) {
                return c.JSON(http.StatusUnauthorized, map[string]interface{}{
                    "error": "未登录",
                })
            }
            return next(c)
        }
    })
    
    api.GET("/user/info", func(c echo.Context) error {
        token := c.Request().Header.Get("Authorization")
        loginID, _ := stputil.GetLoginID(token)
        
        return c.JSON(http.StatusOK, map[string]interface{}{
            "loginId": loginID,
        })
    })
    
    e.Start(":8080")
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

## 特性

- ✅ Echo 中间件集成
- ✅ 请求上下文适配
- ✅ 登录认证
- ✅ 权限验证
- ✅ 角色管理

