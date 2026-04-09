# SaToken Quick Start

这是一个基于当前 `SaToken` 项目的简单 `quick_start` 示例，保留了现有的 Gin 路由结构和接口组织方式，主要用于快速体验登录、权限、角色、在线状态、Session、封禁等常见能力。

## 目录说明

```text
quick_start/
├── main.go
├── README.md
├── go.mod
└── go.sum
```

## 当前使用的包

- `github.com/click33/sa-token-go/integrations/gin`
- `github.com/click33/sa-token-go/com/storage/redis`
- `github.com/gin-gonic/gin`

示例中统一使用：

```go
import (
    "github.com/click33/sa-token-go/com/storage/redis"
    satoken "github.com/click33/sa-token-go/integrations/gin"
)
```

## 功能概览

当前示例保留了这些接口分组：

- `/api/auth`：登录、退出、登录状态、Token 信息、在线终端统计
- `/api/online`：踢下线、顶下线
- `/api/permission`：权限增删改查和权限判断
- `/api/role`：角色增删改查和角色判断
- `/api/session`：Session 和 Token 列表查询
- `/api/disable`：账号封禁、解封、封禁信息查询

## 启动方式

### 1. 进入示例目录

```bash
cd examples/quick_start
```

### 2. 准备 Redis

请先确保 Redis 可用，并根据你的实际环境修改 `main.go` 里的 Redis 连接地址。

当前示例里初始化位置在：

```go
func initSaToken() error
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
mgr := satoken.NewDefaultBuilder().
    SetStorage(storage).
    TokenName("token").
    Timeout(7200).
    RenewMaxRefresh(1800).
    IsConcurrent(true).
    MaxLoginCount(5).
    IsReadHeader(true).
    IsLog(true).
    IsPrintBanner(true).
    Build()

satoken.SetManager(mgr)
```

## 认证说明

示例中保留了一个简单的 Gin 中间件，通过 `Authorization` 请求头读取 token，并调用：

```go
satoken.CheckLogin(c.Request.Context(), token)
```

来校验当前请求是否已登录。

## 说明

- 这个示例当前是“简单 quick start”，不是完整生产模板
- 代码里仍然保留了你现有的接口组织和业务流程
- 这次调整重点是同步当前 `SaToken` 项目的包路径、别名命名、module 和文档描述
