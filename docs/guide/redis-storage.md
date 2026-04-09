# Redis Storage Guide

[中文文档](redis-storage_zh.md) | English

## Overview

The current Redis storage implementation is located at:

- `com/storage/redis`

There are only 3 public constructors:

1. `redis.NewStorage(url string)`
2. `redis.NewStorageFromConfig(cfg *redis.Config)`
3. `redis.NewStorageFromClient(client *redis.Client)`

## Installation

```bash
go get github.com/click33/sa-token-go/com/storage/redis
go get github.com/redis/go-redis/v9
```

## Usage

### Option 1: Redis URL

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

### Option 2: Structured Config

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

### Option 3: Reuse an Existing go-redis Client

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

## Config Fields

Current `redis.Config` fields:

| Field | Description |
|------|------|
| `Host` | Redis host |
| `Port` | Redis port |
| `Password` | Redis password |
| `Database` | DB index |
| `PoolSize` | pool size |
| `DialTimeout` | dial timeout |
| `ReadTimeout` | read timeout |
| `WriteTimeout` | write timeout |
| `PoolTimeout` | pool acquisition timeout |
| `OperationTimeout` | reserved field; not applied per operation by the current storage implementation |

## Using It With Sa-Token

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

Login state, session data, permissions, roles, nonce data, and OAuth2 tokens will then all use Redis.

## Current Storage Capabilities

The Redis adapter currently implements:

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

## Important Notes

### No Redis Builder API

There is **no** `redis.NewBuilder()` API in the current package. Older documentation that mentions it is outdated.

### NewStorageFromClient Only Accepts *redis.Client

The current signature is:

```go
func NewStorageFromClient(client *redis.Client) *Storage
```

That means:

1. a standard standalone `*redis.Client` works directly
2. Redis Cluster and Sentinel are not exposed through a ready-made adapter entry in this package
3. if you need another client shape, you should add your own `adapter.Storage` implementation

### Missing Keys Are Not Treated As Errors

Current behavior:

- `Get()` returns `nil, nil` when the key is missing
- `GetAndDelete()` also returns `nil, nil` when the key is missing
- `Expire()` returns `ErrKeyNotFound` when the key does not exist

## Quick Health Check

```go
ctx := context.Background()

if err := storage.Ping(ctx); err != nil {
    panic(err)
}

client := storage.GetClient()
_ = client
```

## Best Practices

1. start with memory storage in development and switch to Redis for integration or production
2. configure pool size and timeouts explicitly in production
3. JWT style still uses server-side storage, so Redis remains useful
4. close the underlying client on graceful shutdown when you own the client lifecycle

## Related Documentation

- [Authentication Guide](authentication.md)
- [JWT Guide](jwt.md)
- [OAuth2 Guide](oauth2.md)
