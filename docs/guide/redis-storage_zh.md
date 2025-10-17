# Redis 存储配置指南

[English](redis-storage.md) | 中文文档

## 概述

Redis 存储是生产环境推荐的存储后端。它提供高性能、数据持久化，并支持分布式部署。

## 安装

```bash
# 安装 Redis 存储模块
go get github.com/click33/sa-token-go/storage/redis

# 安装 Redis 客户端
go get github.com/redis/go-redis/v9
```

## 基本使用

### 1. 简单配置

```go
package main

import (
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/storage/redis"
    goredis "github.com/redis/go-redis/v9"
)

func main() {
    // 创建 Redis 客户端
    rdb := goredis.NewClient(&goredis.Options{
        Addr:     "localhost:6379",
        Password: "", // 无密码
        DB:       0,  // 默认 DB
    })

    // 使用 Redis 存储初始化 Sa-Token
    stputil.SetManager(
        core.NewBuilder().
            Storage(redis.NewStorage(rdb)).
            TokenName("Authorization").
            Timeout(86400). // 24小时
            Build(),
    )

    // 现在可以使用 Sa-Token 了
    token, _ := stputil.Login(1000)
    println("登录成功，Token:", token)
}
```

### 2. 带密码认证

```go
rdb := goredis.NewClient(&goredis.Options{
    Addr:     "localhost:6379",
    Password: "your-redis-password", // 设置密码
    DB:       0,
})

stputil.SetManager(
    core.NewBuilder().
        Storage(redis.NewStorage(rdb)).
        Build(),
)
```

### 3. 使用 Redis 集群

```go
rdb := goredis.NewClusterClient(&goredis.ClusterOptions{
    Addrs: []string{
        "localhost:7000",
        "localhost:7001",
        "localhost:7002",
    },
    Password: "your-password",
})

stputil.SetManager(
    core.NewBuilder().
        Storage(redis.NewStorage(rdb)).
        Build(),
)
```

### 4. 使用 Redis 哨兵

```go
rdb := goredis.NewFailoverClient(&goredis.FailoverOptions{
    MasterName:    "mymaster",
    SentinelAddrs: []string{
        "localhost:26379",
        "localhost:26380",
        "localhost:26381",
    },
    Password: "your-password",
    DB:       0,
})

stputil.SetManager(
    core.NewBuilder().
        Storage(redis.NewStorage(rdb)).
        Build(),
)
```

## 高级配置

### 完整配置示例

```go
package main

import (
    "time"
    
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/storage/redis"
    goredis "github.com/redis/go-redis/v9"
)

func main() {
    // Redis 客户端完整选项
    rdb := goredis.NewClient(&goredis.Options{
        Addr:         "localhost:6379",
        Password:     "",
        DB:           0,
        PoolSize:     10,              // 连接池大小
        MinIdleConns: 5,               // 最小空闲连接数
        MaxRetries:   3,               // 最大重试次数
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
        PoolTimeout:  4 * time.Second,
    })

    // 使用 Redis 初始化 Sa-Token
    stputil.SetManager(
        core.NewBuilder().
            Storage(redis.NewStorage(rdb)).
            TokenName("Authorization").
            TokenStyle(core.TokenStyleJWT).
            JwtSecretKey("your-secret-key").
            Timeout(7200).              // 2小时
            ActiveTimeout(1800).        // 30分钟
            IsConcurrent(true).
            IsShare(false).             // 每次登录获得唯一Token
            MaxLoginCount(5).           // 最多5个并发登录
            AutoRenew(true).
            IsReadHeader(true).
            IsPrintBanner(true).
            Build(),
    )

    // 使用 Sa-Token
    token, _ := stputil.Login(1000)
    println("Token:", token)
}
```

### 连接池配置

```go
rdb := goredis.NewClient(&goredis.Options{
    Addr:     "localhost:6379",
    
    // 连接池设置
    PoolSize:     100,              // 最大连接数
    MinIdleConns: 10,               // 最小空闲连接数
    MaxIdleConns: 50,               // 最大空闲连接数
    
    // 超时设置
    DialTimeout:  5 * time.Second,  // 连接超时
    ReadTimeout:  3 * time.Second,  // 读取超时
    WriteTimeout: 3 * time.Second,  // 写入超时
    PoolTimeout:  4 * time.Second,  // 连接池获取超时
    
    // 重试设置
    MaxRetries:      3,              // 最大重试次数
    MinRetryBackoff: 8 * time.Millisecond,
    MaxRetryBackoff: 512 * time.Millisecond,
})
```

## 环境变量

### 使用环境变量

```go
package main

import (
    "os"
    "strconv"
    
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/storage/redis"
    goredis "github.com/redis/go-redis/v9"
)

func main() {
    // 从环境变量读取配置
    redisAddr := os.Getenv("REDIS_ADDR")
    if redisAddr == "" {
        redisAddr = "localhost:6379"
    }
    
    redisPassword := os.Getenv("REDIS_PASSWORD")
    
    redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
    
    rdb := goredis.NewClient(&goredis.Options{
        Addr:     redisAddr,
        Password: redisPassword,
        DB:       redisDB,
    })

    stputil.SetManager(
        core.NewBuilder().
            Storage(redis.NewStorage(rdb)).
            JwtSecretKey(os.Getenv("JWT_SECRET_KEY")).
            Build(),
    )
}
```

### .env 文件示例

```bash
# Redis 配置
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=your-password
REDIS_DB=0

# Sa-Token 配置
JWT_SECRET_KEY=your-256-bit-secret-key
TOKEN_TIMEOUT=7200
```

## Redis 键结构

Sa-Token-Go 在 Redis 中使用以下键模式：

```
satoken:login:token:{tokenValue}        # Token -> LoginID 映射
satoken:login:session:{loginID}:token   # LoginID -> Token 列表
satoken:session:{loginID}               # 用户 Session 数据
satoken:permission:{loginID}            # 用户权限
satoken:role:{loginID}                  # 用户角色
satoken:disable:{loginID}               # 账号禁用状态
```

### 在 Redis CLI 中查看键

```bash
# 连接到 Redis
redis-cli

# 列出所有 Sa-Token 键
KEYS satoken:*

# 查看 Token 信息
GET satoken:login:token:your-token-value

# 查看用户 Session
GET satoken:session:1000

# 查看用户权限
SMEMBERS satoken:permission:1000

# 查看用户角色
SMEMBERS satoken:role:1000
```

## 生产环境最佳实践

### 1. 连接池配置

```go
rdb := goredis.NewClient(&goredis.Options{
    Addr:         "localhost:6379",
    PoolSize:     100,  // 根据负载调整
    MinIdleConns: 10,   // 保持一些连接活跃
})
```

### 2. 错误处理

```go
rdb := goredis.NewClient(&goredis.Options{
    Addr:     "localhost:6379",
    Password: os.Getenv("REDIS_PASSWORD"),
})

// 测试连接
ctx := context.Background()
if err := rdb.Ping(ctx).Err(); err != nil {
    log.Fatalf("无法连接到 Redis: %v", err)
}
```

### 3. 高可用（哨兵模式）

```go
rdb := goredis.NewFailoverClient(&goredis.FailoverOptions{
    MasterName:    "mymaster",
    SentinelAddrs: []string{
        "sentinel1:26379",
        "sentinel2:26379",
        "sentinel3:26379",
    },
    Password: os.Getenv("REDIS_PASSWORD"),
    DB:       0,
    
    // 哨兵选项
    SentinelPassword: os.Getenv("SENTINEL_PASSWORD"),
    
    // 连接池
    PoolSize:     100,
    MinIdleConns: 10,
})
```

### 4. TLS/SSL 支持

```go
import "crypto/tls"

rdb := goredis.NewClient(&goredis.Options{
    Addr:     "localhost:6379",
    Password: os.Getenv("REDIS_PASSWORD"),
    
    // 启用 TLS
    TLSConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
    },
})
```

### 5. 优雅关闭

```go
func main() {
    rdb := goredis.NewClient(&goredis.Options{
        Addr: "localhost:6379",
    })
    
    stputil.SetManager(
        core.NewBuilder().
            Storage(redis.NewStorage(rdb)).
            Build(),
    )

    // ... 你的应用代码 ...

    // 优雅关闭
    defer func() {
        if err := rdb.Close(); err != nil {
            log.Printf("关闭 Redis 时出错: %v", err)
        }
    }()
}
```

## 性能优化

### 1. 使用管道

Sa-Token-Go 的 Redis 存储自动为批量操作使用管道。

### 2. 键过期时间

Sa-Token 会根据你的 `Timeout` 配置自动设置键的过期时间：

```go
core.NewBuilder().
    Timeout(3600).  // 键将在1小时后过期
    Build()
```

### 3. 连接复用

Redis 客户端维护连接池以获得最佳性能：

```go
rdb := goredis.NewClient(&goredis.Options{
    PoolSize:     100,  // 复用最多100个连接
    MinIdleConns: 10,   // 始终保持10个热连接
})
```

## 监控

### 检查 Redis 状态

```go
import "context"

ctx := context.Background()

// Ping
pong, err := rdb.Ping(ctx).Err()
if err != nil {
    log.Printf("Redis ping 失败: %v", err)
}

// 获取信息
info, err := rdb.Info(ctx).Result()
if err != nil {
    log.Printf("获取 Redis 信息失败: %v", err)
}
println(info)
```

### 监控键数量

```bash
# 在 Redis CLI 中
INFO keyspace

# 输出示例：
# db0:keys=1234,expires=567,avg_ttl=3600000
```

## 故障排查

### 连接被拒绝

```go
// 问题：无法连接到 Redis
// 解决方案：检查 Redis 是否运行
// 命令：redis-cli ping
```

### 认证失败

```go
// 问题：NOAUTH Authentication required
// 解决方案：设置正确的密码
rdb := goredis.NewClient(&goredis.Options{
    Addr:     "localhost:6379",
    Password: "correct-password",
})
```

### 连接数过多

```go
// 问题：ERR max number of clients reached
// 解决方案：增加 Redis 最大客户端数或减少连接池大小
// Redis 配置：maxclients 10000

rdb := goredis.NewClient(&goredis.Options{
    PoolSize: 50, // 减少连接池大小
})
```

## Docker 部署

### Docker Compose 示例

```yaml
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --requirepass your-password
    volumes:
      - redis-data:/data
    restart: unless-stopped

  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - REDIS_ADDR=redis:6379
      - REDIS_PASSWORD=your-password
      - JWT_SECRET_KEY=your-secret-key
    depends_on:
      - redis

volumes:
  redis-data:
```

### 应用代码

```go
// 在你的 Go 应用中
func main() {
    rdb := goredis.NewClient(&goredis.Options{
        Addr:     os.Getenv("REDIS_ADDR"),     // redis:6379
        Password: os.Getenv("REDIS_PASSWORD"),
        DB:       0,
    })

    stputil.SetManager(
        core.NewBuilder().
            Storage(redis.NewStorage(rdb)).
            JwtSecretKey(os.Getenv("JWT_SECRET_KEY")).
            Build(),
    )
    
    // 启动你的 Web 服务器...
}
```

## Kubernetes 部署

### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: satoken-config
data:
  REDIS_ADDR: "redis-service:6379"
  REDIS_DB: "0"
```

### Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: satoken-secret
type: Opaque
stringData:
  REDIS_PASSWORD: "your-redis-password"
  JWT_SECRET_KEY: "your-jwt-secret-key"
```

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: satoken-app
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: app
        image: your-app:latest
        env:
        - name: REDIS_ADDR
          valueFrom:
            configMapKeyRef:
              name: satoken-config
              key: REDIS_ADDR
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: satoken-secret
              key: REDIS_PASSWORD
        - name: JWT_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: satoken-secret
              key: JWT_SECRET_KEY
```

## 对比：Memory vs Redis

| 特性 | Memory | Redis |
|------|--------|-------|
| 性能 | 优秀 | 很好 |
| 持久化 | ❌ 重启丢失 | ✅ 持久化 |
| 分布式 | ❌ 不支持 | ✅ 支持 |
| 扩展性 | 有限 | 优秀 |
| 配置 | 简单 | 需要 Redis |
| 适用场景 | 开发/测试 | 生产环境 |

## 完整示例

```go
package main

import (
    "context"
    "log"
    "os"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/storage/redis"
    sagin "github.com/click33/sa-token-go/integrations/gin"
    goredis "github.com/redis/go-redis/v9"
)

func main() {
    // 初始化 Redis
    rdb := goredis.NewClient(&goredis.Options{
        Addr:     os.Getenv("REDIS_ADDR"),
        Password: os.Getenv("REDIS_PASSWORD"),
        DB:       0,
        
        PoolSize:     100,
        MinIdleConns: 10,
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
    })

    // 测试 Redis 连接
    ctx := context.Background()
    if err := rdb.Ping(ctx).Err(); err != nil {
        log.Fatalf("无法连接到 Redis: %v", err)
    }

    // 初始化 Sa-Token
    stputil.SetManager(
        core.NewBuilder().
            Storage(redis.NewStorage(rdb)).
            TokenName("Authorization").
            TokenStyle(core.TokenStyleJWT).
            JwtSecretKey(os.Getenv("JWT_SECRET_KEY")).
            Timeout(7200).
            ActiveTimeout(1800).
            IsConcurrent(true).
            IsShare(false).
            MaxLoginCount(5).
            AutoRenew(true).
            IsReadHeader(true).
            IsPrintBanner(true).
            IsLog(true).
            Build(),
    )

    // 设置 Gin
    r := gin.Default()
    r.Use(sagin.NewPlugin(stputil.GetManager()).Build())

    // 路由
    r.POST("/login", loginHandler)
    r.GET("/user/info", sagin.CheckLogin(), userInfoHandler)
    r.GET("/admin", sagin.CheckPermission("admin"), adminHandler)

    // 启动服务器
    if err := r.Run(":8080"); err != nil {
        log.Fatal(err)
    }

    // 优雅关闭
    defer rdb.Close()
}

func loginHandler(c *gin.Context) {
    // ... 登录逻辑 ...
}

func userInfoHandler(c *gin.Context) {
    // ... 用户信息逻辑 ...
}

func adminHandler(c *gin.Context) {
    // ... 管理员逻辑 ...
}
```

## 相关文档

- [快速开始](../tutorial/quick-start.md)
- [Memory 存储](../../storage/memory/)
- [认证指南](authentication.md)
- [JWT 指南](jwt.md)

## Redis 资源

- [Redis 官方网站](https://redis.io/)
- [go-redis 文档](https://redis.uptrace.dev/)
- [Redis 命令](https://redis.io/commands/)

