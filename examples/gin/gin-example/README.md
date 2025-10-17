# Gin 框架集成示例

本示例演示如何在 Gin 框架中使用 Sa-Token-Go。

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

### 方式一：使用 Manager 实例（推荐用于复杂场景）

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/click33/sa-token-go/core"
    sagin "github.com/click33/sa-token-go/integrations/gin"
    "github.com/click33/sa-token-go/storage/memory"
)

func main() {
    // 创建 Manager
    manager := core.NewBuilder().
        Storage(memory.NewStorage()).
        TokenName("Authorization").
        Build()
    
    // 创建插件
    plugin := sagin.NewPlugin(manager)
    
    // 设置路由
    r := gin.Default()
    r.POST("/login", plugin.LoginHandler)
    r.GET("/user", plugin.AuthMiddleware(), plugin.UserInfoHandler)
    
    r.Run(":8080")
}
```

### 方式二：使用 StpUtil 全局工具类（推荐用于简单场景）

```go
package main

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    sagin "github.com/click33/sa-token-go/integrations/gin"
    "github.com/click33/sa-token-go/storage/memory"
)

func init() {
    // 初始化 StpUtil
    stputil.SetManager(
        core.NewBuilder().
            Storage(memory.NewStorage()).
            Build(),
    )
}

func main() {
    r := gin.Default()
    
    // 登录接口
    r.POST("/login", func(c *gin.Context) {
        var req struct {
            UserID int `json:"userId"`
        }
        c.ShouldBindJSON(&req)
        
        token, _ := stputil.Login(req.UserID)
        c.JSON(http.StatusOK, gin.H{"token": token})
    })
    
    // 使用注解装饰器
    r.GET("/user", sagin.CheckLogin(), func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        loginID, _ := stputil.GetLoginID(token)
        
        c.JSON(http.StatusOK, gin.H{
            "loginId": loginID,
            "message": "用户信息",
        })
    })
    
    // 需要权限
    r.GET("/admin", sagin.CheckPermission("admin:*"), func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "管理员数据"})
    })
    
    r.Run(":8080")
}
```

## API 端点

### 公开接口

- `POST /login` - 用户登录
  ```bash
  curl -X POST http://localhost:8080/login \
    -H "Content-Type: application/json" \
    -d '{"username":"test","password":"123456"}'
  ```

- `GET /public` - 公开访问
  ```bash
  curl http://localhost:8080/public
  ```

### 受保护接口

- `GET /api/user` - 获取用户信息（需要登录）
  ```bash
  curl http://localhost:8080/api/user \
    -H "Authorization: YOUR_TOKEN"
  ```

- `GET /api/admin` - 管理员接口（需要管理员权限）
  ```bash
  curl http://localhost:8080/api/admin \
    -H "Authorization: YOUR_TOKEN"
  ```

## 配置文件

配置文件位于 `configs/config.yaml`：

```yaml
token:
  timeout: 7200        # Token超时时间（秒）
  active_timeout: 1800 # 活跃超时时间（秒）

server:
  port: 8080           # 服务器端口
```

## 更多示例

查看 [注解示例](../../annotation/annotation-example) 了解更多注解装饰器的用法。

