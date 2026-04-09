# Chi SaToken Example

这是一个基于当前 `SaToken` 项目的 `Chi` 示例，结构和 `examples/gf`、`examples/gin` 保持一致，主要演示登录、角色、权限以及注解式校验的基本用法。

## 目录说明

```text
examples/chi/
├── main.go
├── README.md
├── go.mod
└── go.sum
```

## 当前使用的包

- `github.com/click33/sa-token-go/integrations/chi`
- `github.com/click33/sa-token-go/com/storage/redis`
- `github.com/go-chi/chi/v5`

示例里统一使用：

```go
import (
    "github.com/click33/sa-token-go/com/storage/redis"
    satoken "github.com/click33/sa-token-go/integrations/chi"
)
```

## 功能概览

- 注册全局 `SaTokenContext` 中间件
- 公开接口访问
- 登录校验中间件
- 角色校验中间件
- 权限校验中间件
- 注解式登录、角色、权限组合校验

## 启动方式

### 1. 进入目录

```bash
cd examples/chi
```

### 2. 配置 Redis

请先确认 Redis 可用，并按实际环境修改 [main.go](/g:/code/go/my_project/sa-token-go/examples/chi/main.go) 里的连接地址。

当前初始化位置：

```go
func initManager(ctx context.Context)
```

### 3. 运行示例

```bash
go run main.go
```

默认访问地址：

```text
http://localhost:8080
```

## 中间件说明

示例中使用了这些 `Chi` 集成能力：

```go
r.Use(satoken.RegisterSaTokenContextMiddleware())
user.Use(satoken.AuthMiddleware())
admin.Use(satoken.RoleMiddleware([]string{"admin"}))
resource.Use(satoken.PermissionMiddleware([]string{"resource:read"}))
```

需要直接读取上下文时，可以使用：

```go
saCtx, ok := satoken.GetSaTokenContextByCtx(r.Context())
```

## 说明

- 当前示例主要用于展示 `Chi` 下的 `SaToken` 基础集成方式
- 示例入口已经调整为和 `gf`、`gin` 一致的根目录 `main.go`
- 示例逻辑保留为简单 quick demo，方便继续扩展
