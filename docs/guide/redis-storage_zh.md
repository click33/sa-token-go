# Redis 存储指南

[English](redis-storage.md) | 中文文档

## 概览

当前 Redis 存储实现位于：

- `com/storage/redis`

公开构造方式只有 3 种：

1. `redis.NewStorage(url string)`
2. `redis.NewStorageFromConfig(cfg *redis.Config)`
3. `redis.NewStorageFromClient(client *redis.Client)`

## 安装

```bash
go get github.com/click33/sa-token-go/com/storage/redis
go get github.com/redis/go-redis/v9
```

## 使用方式

### 方式一：Redis URL

```go
package main

import (
    "github.com/click33/sa-token-go/com/storage/redis"
    "github.com/click33/sa-token-go/core/builder"
    "github.com/click33/sa-token-go/stputil"
)

func initSaToken() {
    storage, err := redis.NewStorage("redis://localhost:6379/0")
    if err != nil {
        panic(err)
    }

    stputil.SetManager(
        builder.NewBuilder().
            SetStorage(storage).
            Build(),
    )
}
```

### 方式二：结构化配置

```go
storage, err := redis.NewStorageFromConfig(&redis.Config{
    Host:         "127.0.0.1",
    Port:         6379,
    Password:     "",
    Database:     0,
    PoolSize:     20,
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
    PoolTimeout:  4 * time.Second,
})
if err != nil {
    panic(err)
}

stputil.SetManager(
    builder.NewBuilder().
        SetStorage(storage).
        Build(),
)
```

### 方式三：复用现有 go-redis Client

```go
rdb := goredis.NewClient(&goredis.Options{
    Addr:     "127.0.0.1:6379",
    Password: "",
    DB:       0,
})

storage := redis.NewStorageFromClient(rdb)

stputil.SetManager(
    builder.NewBuilder().
        SetStorage(storage).
        Build(),
)
```

## Config 字段

当前 `redis.Config` 支持：

| 字段 | 说明 |
|------|------|
| `Host` | 主机 |
| `Port` | 端口 |
| `Password` | 密码 |
| `Database` | 库索引 |
| `PoolSize` | 连接池大小 |
| `DialTimeout` | 建连超时 |
| `ReadTimeout` | 读超时 |
| `WriteTimeout` | 写超时 |
| `PoolTimeout` | 取连接超时 |
| `OperationTimeout` | 预留字段，当前存储实现未在各操作中单独套用 |

## 和 Sa-Token 搭配使用

```go
storage, _ := redis.NewStorage("redis://localhost:6379/0")

stputil.SetManager(
    builder.NewBuilder().
        SetStorage(storage).
        TokenName("satoken").
        Timeout(2 * 60 * 60).
        ActiveTimeout(30 * 60).
        AutoRenew(true).
        Build(),
)
```

这时登录态、Session、权限、角色、Nonce、OAuth2 Token 等数据都会走 Redis。

## 当前存储能力

当前 Redis 适配器实现了这些基础操作：

- `Set`
- `Get`
- `GetAndDelete`
- `Delete`
- `Exists`
- `Keys`
- `Expire`
- `TTL`
- `Clear`
- `Ping`
- `Close`
- `GetClient`

## 注意事项

### 不存在 Redis Builder

当前包里**没有** `redis.NewBuilder()` 这一套 API，旧文档里这部分已经过时。

### NewStorageFromClient 只接收 *redis.Client

当前 `NewStorageFromClient` 的签名是：

```go
func NewStorageFromClient(client *redis.Client) *Storage
```

这意味着：

1. 直接传入标准单机 `*redis.Client` 没问题
2. Redis Cluster / Sentinel 目前没有现成的同名适配入口
3. 如果你要接其他客户端形态，需要自己补一层 `adapter.Storage`

### 不存在 Key 不算错误

当前实现里：

- `Get()` 取不到 key 时返回 `nil, nil`
- `GetAndDelete()` 取不到 key 时也返回 `nil, nil`
- `Expire()` 如果 key 不存在，则返回 `ErrKeyNotFound`

## 简单排查

```go
ctx := context.Background()

if err := storage.Ping(ctx); err != nil {
    panic(err)
}

client := storage.GetClient()
_ = client
```

## 最佳实践

1. 开发环境可以先用内存存储，联调或生产再切 Redis
2. 生产环境建议显式配置连接池大小和超时
3. JWT 风格并不会绕开存储层，配 Redis 仍然有意义
4. 用完自定义 client 后记得在应用退出时关闭连接

## 相关文档

- [登录认证](authentication_zh.md)
- [JWT 指南](jwt_zh.md)
- [OAuth2 指南](oauth2_zh.md)
