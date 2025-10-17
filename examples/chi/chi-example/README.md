# Chi 框架集成示例

本示例演示如何在 Chi 框架中使用 Sa-Token-Go。

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
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/click33/sa-token-go/core"
    sachi "github.com/click33/sa-token-go/integrations/chi"
    "github.com/click33/sa-token-go/storage/memory"
)

func main() {
    // 创建 Manager
    manager := core.NewBuilder().
        Storage(memory.NewStorage()).
        TokenName("Authorization").
        Build()
    
    // 创建插件
    plugin := sachi.NewPlugin(manager)
    
    // 设置路由
    r := chi.NewRouter()
    r.Post("/login", plugin.LoginHandler)
    
    r.Group(func(r chi.Router) {
        r.Use(plugin.AuthMiddleware())
        r.Get("/api/user/info", userInfoHandler)
    })
    
    http.ListenAndServe(":8080", r)
}
```

### 方式二：使用 StpUtil 全局工具类

```go
package main

import (
    "encoding/json"
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    sachi "github.com/click33/sa-token-go/integrations/chi"
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
    r := chi.NewRouter()
    
    // 登录接口
    r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            UserID int `json:"userId"`
        }
        json.NewDecoder(r.Body).Decode(&req)
        
        token, _ := stputil.Login(req.UserID)
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "token": token,
        })
    })
    
    // 需要登录的路由组
    r.Group(func(r chi.Router) {
        // 使用认证中间件
        r.Use(func(next http.Handler) http.Handler {
            return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                token := r.Header.Get("Authorization")
                if !stputil.IsLogin(token) {
                    w.WriteHeader(http.StatusUnauthorized)
                    json.NewEncoder(w).Encode(map[string]interface{}{
                        "error": "未登录",
                    })
                    return
                }
                next.ServeHTTP(w, r)
            })
        })
        
        r.Get("/api/user/info", func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            loginID, _ := stputil.GetLoginID(token)
            
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(map[string]interface{}{
                "loginId": loginID,
            })
        })
    })
    
    http.ListenAndServe(":8080", r)
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

