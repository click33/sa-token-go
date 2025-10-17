# Redis Storage Configuration Guide

[中文文档](redis-storage_zh.md) | English

## Overview

Redis storage is the recommended storage backend for production environments. It provides high performance, data persistence, and supports distributed deployments.

## Installation

```bash
# Install Redis storage module
go get github.com/click33/sa-token-go/storage/redis

# Install Redis client
go get github.com/redis/go-redis/v9
```

## Basic Usage

### 1. Simple Configuration

```go
package main

import (
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/storage/redis"
    goredis "github.com/redis/go-redis/v9"
)

func main() {
    // Create Redis client
    rdb := goredis.NewClient(&goredis.Options{
        Addr:     "localhost:6379",
        Password: "", // No password
        DB:       0,  // Default DB
    })

    // Initialize Sa-Token with Redis storage
    stputil.SetManager(
        core.NewBuilder().
            Storage(redis.NewStorage(rdb)).
            TokenName("Authorization").
            Timeout(86400). // 24 hours
            Build(),
    )

    // Now you can use Sa-Token
    token, _ := stputil.Login(1000)
    println("Login successful, Token:", token)
}
```

### 2. With Password Authentication

```go
rdb := goredis.NewClient(&goredis.Options{
    Addr:     "localhost:6379",
    Password: "your-redis-password", // Set password
    DB:       0,
})

stputil.SetManager(
    core.NewBuilder().
        Storage(redis.NewStorage(rdb)).
        Build(),
)
```

### 3. Using Redis Cluster

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

### 4. Using Redis Sentinel

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

## Advanced Configuration

### Complete Configuration Example

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
    // Redis client with full options
    rdb := goredis.NewClient(&goredis.Options{
        Addr:         "localhost:6379",
        Password:     "",
        DB:           0,
        PoolSize:     10,              // Connection pool size
        MinIdleConns: 5,               // Minimum idle connections
        MaxRetries:   3,               // Maximum retry attempts
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
        PoolTimeout:  4 * time.Second,
    })

    // Initialize Sa-Token with Redis
    stputil.SetManager(
        core.NewBuilder().
            Storage(redis.NewStorage(rdb)).
            TokenName("Authorization").
            TokenStyle(core.TokenStyleJWT).
            JwtSecretKey("your-secret-key").
            Timeout(7200).              // 2 hours
            ActiveTimeout(1800).        // 30 minutes
            IsConcurrent(true).
            IsShare(false).             // Each login gets unique token
            MaxLoginCount(5).           // Max 5 concurrent logins
            AutoRenew(true).
            IsReadHeader(true).
            IsPrintBanner(true).
            Build(),
    )

    // Use Sa-Token
    token, _ := stputil.Login(1000)
    println("Token:", token)
}
```

### Connection Pool Configuration

```go
rdb := goredis.NewClient(&goredis.Options{
    Addr:     "localhost:6379",
    
    // Connection pool settings
    PoolSize:     100,              // Maximum connections
    MinIdleConns: 10,               // Minimum idle connections
    MaxIdleConns: 50,               // Maximum idle connections
    
    // Timeout settings
    DialTimeout:  5 * time.Second,  // Connection timeout
    ReadTimeout:  3 * time.Second,  // Read timeout
    WriteTimeout: 3 * time.Second,  // Write timeout
    PoolTimeout:  4 * time.Second,  // Pool get timeout
    
    // Retry settings
    MaxRetries:      3,              // Maximum retries
    MinRetryBackoff: 8 * time.Millisecond,
    MaxRetryBackoff: 512 * time.Millisecond,
})
```

## Environment Variables

### Using Environment Variables

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
    // Read from environment variables
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

### .env File Example

```bash
# Redis Configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=your-password
REDIS_DB=0

# Sa-Token Configuration
JWT_SECRET_KEY=your-256-bit-secret-key
TOKEN_TIMEOUT=7200
```

## Redis Key Structure

Sa-Token-Go uses the following key patterns in Redis:

```
satoken:login:token:{tokenValue}        # Token -> LoginID mapping
satoken:login:session:{loginID}:token   # LoginID -> Token list
satoken:session:{loginID}               # User session data
satoken:permission:{loginID}            # User permissions
satoken:role:{loginID}                  # User roles
satoken:disable:{loginID}               # Account disable status
```

### View Keys in Redis CLI

```bash
# Connect to Redis
redis-cli

# List all Sa-Token keys
KEYS satoken:*

# View token info
GET satoken:login:token:your-token-value

# View user session
GET satoken:session:1000

# View user permissions
SMEMBERS satoken:permission:1000

# View user roles
SMEMBERS satoken:role:1000
```

## Production Best Practices

### 1. Connection Pool

```go
rdb := goredis.NewClient(&goredis.Options{
    Addr:         "localhost:6379",
    PoolSize:     100,  // Adjust based on your load
    MinIdleConns: 10,   // Keep some connections alive
})
```

### 2. Error Handling

```go
rdb := goredis.NewClient(&goredis.Options{
    Addr:     "localhost:6379",
    Password: os.Getenv("REDIS_PASSWORD"),
})

// Test connection
ctx := context.Background()
if err := rdb.Ping(ctx).Err(); err != nil {
    log.Fatalf("Failed to connect to Redis: %v", err)
}
```

### 3. High Availability (Sentinel)

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
    
    // Sentinel options
    SentinelPassword: os.Getenv("SENTINEL_PASSWORD"),
    
    // Connection pool
    PoolSize:     100,
    MinIdleConns: 10,
})
```

### 4. TLS/SSL Support

```go
import "crypto/tls"

rdb := goredis.NewClient(&goredis.Options{
    Addr:     "localhost:6379",
    Password: os.Getenv("REDIS_PASSWORD"),
    
    // Enable TLS
    TLSConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
    },
})
```

### 5. Graceful Shutdown

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

    // ... your application code ...

    // Graceful shutdown
    defer func() {
        if err := rdb.Close(); err != nil {
            log.Printf("Error closing Redis: %v", err)
        }
    }()
}
```

## Performance Optimization

### 1. Use Pipelining

Redis storage in Sa-Token-Go automatically uses pipelining for batch operations.

### 2. Key Expiration

Sa-Token automatically sets expiration time for keys based on your `Timeout` configuration:

```go
core.NewBuilder().
    Timeout(3600).  // Keys will expire in 1 hour
    Build()
```

### 3. Connection Reuse

The Redis client maintains a connection pool for optimal performance:

```go
rdb := goredis.NewClient(&goredis.Options{
    PoolSize:     100,  // Reuse up to 100 connections
    MinIdleConns: 10,   // Always keep 10 warm connections
})
```

## Monitoring

### Check Redis Status

```go
import "context"

ctx := context.Background()

// Ping
pong, err := rdb.Ping(ctx).Err()
if err != nil {
    log.Printf("Redis ping failed: %v", err)
}

// Get info
info, err := rdb.Info(ctx).Result()
if err != nil {
    log.Printf("Failed to get Redis info: %v", err)
}
println(info)
```

### Monitor Key Count

```bash
# In Redis CLI
INFO keyspace

# Output example:
# db0:keys=1234,expires=567,avg_ttl=3600000
```

## Troubleshooting

### Connection Refused

```go
// Problem: cannot connect to Redis
// Solution: Check if Redis is running
// Command: redis-cli ping
```

### Authentication Failed

```go
// Problem: NOAUTH Authentication required
// Solution: Set correct password
rdb := goredis.NewClient(&goredis.Options{
    Addr:     "localhost:6379",
    Password: "correct-password",
})
```

### Too Many Connections

```go
// Problem: ERR max number of clients reached
// Solution: Increase Redis max clients or reduce pool size
// Redis config: maxclients 10000

rdb := goredis.NewClient(&goredis.Options{
    PoolSize: 50, // Reduce pool size
})
```

## Docker Deployment

### Docker Compose Example

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

### Application Code

```go
// In your Go application
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
    
    // Start your web server...
}
```

## Kubernetes Deployment

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

## Comparison: Memory vs Redis

| Feature | Memory | Redis |
|---------|--------|-------|
| Performance | Excellent | Very Good |
| Persistence | ❌ Lost on restart | ✅ Persistent |
| Distributed | ❌ Not supported | ✅ Supported |
| Scalability | Limited | Excellent |
| Setup | Simple | Requires Redis |
| Use Case | Development/Testing | Production |

## Complete Example

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
    // Initialize Redis
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

    // Test Redis connection
    ctx := context.Background()
    if err := rdb.Ping(ctx).Err(); err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }

    // Initialize Sa-Token
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

    // Setup Gin
    r := gin.Default()
    r.Use(sagin.NewPlugin(stputil.GetManager()).Build())

    // Routes
    r.POST("/login", loginHandler)
    r.GET("/user/info", sagin.CheckLogin(), userInfoHandler)
    r.GET("/admin", sagin.CheckPermission("admin"), adminHandler)

    // Start server
    if err := r.Run(":8080"); err != nil {
        log.Fatal(err)
    }

    // Graceful shutdown
    defer rdb.Close()
}

func loginHandler(c *gin.Context) {
    // ... login logic ...
}

func userInfoHandler(c *gin.Context) {
    // ... user info logic ...
}

func adminHandler(c *gin.Context) {
    // ... admin logic ...
}
```

## Related Documentation

- [Quick Start](../tutorial/quick-start.md)
- [Memory Storage](../../storage/memory/)
- [Authentication Guide](authentication.md)
- [JWT Guide](jwt.md)

## Redis Resources

- [Redis Official Site](https://redis.io/)
- [go-redis Documentation](https://redis.uptrace.dev/)
- [Redis Commands](https://redis.io/commands/)

