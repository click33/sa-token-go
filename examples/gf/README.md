# GoFrame SaToken Example

这是一个基于当前 `SaToken` 项目的 `GoFrame` 示例，保留了现有的路由结构和中间件用法，主要用于演示 `GoFrame` 集成下的登录、权限、角色和注解式校验能力。

## 目录说明

```text
examples/gf/
├── main.go
├── README.md
├── go.mod
└── go.sum
```

## 当前使用的包

- `github.com/click33/sa-token-go/integrations/gf`
- `github.com/click33/sa-token-go/com/storage/redis`
- `github.com/gogf/gf/v2`

示例中统一使用：

```go
import (
    "github.com/click33/sa-token-go/com/storage/redis"
    satoken "github.com/click33/sa-token-go/integrations/gf"
)
```

## 功能概览

当前示例保留了这些能力：

- 全局 `SaToken` 上下文中间件注册
- 公共接口访问
- 登录态校验中间件
- 角色校验中间件
- 权限校验中间件
- 注解式登录、角色、权限、组合校验

## 启动方式

### 1. 进入示例目录

```bash
cd examples/gf
```

### 2. 准备 Redis

请先确保 Redis 可用，并根据你的实际环境修改 `main.go` 里的 Redis 连接地址。

当前示例里初始化位置在：

```go
func initManager(ctx context.Context)
```

其中使用的是：

```go
storage, err := redis.NewStorage("redis://:root@192.168.19.104:6379/0?dial_timeout=3&read_timeout=10s&max_retries=2")
```

### 3. 运行示例

```bash
go run main.go
```

默认监听地址：

```text
http://localhost:8080
```

## 初始化说明

示例通过 `satoken.NewDefaultBuilder()` 创建管理器，并注册到全局：

```go
builder := satoken.NewDefaultBuilder()
mgr := builder.
    SetStorage(storage).
    Timeout(3600).
    ActiveTimeout(1800).
    MaxLoginCount(3).
    Build()

satoken.SetManager(mgr)
```

## 中间件说明

示例中使用了这些 `GoFrame` 集成能力：

```go
s.Use(satoken.RegisterSaTokenContextMiddleware(ctx))
group.Middleware(satoken.AuthMiddleware(ctx))
group.Middleware(satoken.RoleMiddleware(ctx, []string{"admin"}))
group.Middleware(satoken.PermissionMiddleware(ctx, []string{"resource:read"}))
```

在需要直接读取上下文时，使用：

```go
saCtx, ok := satoken.GetSaTokenContextByCtx(r.Context())
```

## 说明

- 这个示例当前主要用于展示 `GoFrame` 与 `SaToken` 的基础集成方式
- 代码里仍然保留了你现有的路由组织和处理逻辑
- 这次调整重点是同步当前 `SaToken` 项目的包路径、别名命名、module 和文档描述
