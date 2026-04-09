English | [中文文档](auto-renew_zh.md)

# Auto-Renew Design

## Design Goals

The goals of auto-renew are:

- keep active users logged in without frequent re-authentication
- avoid heavy synchronous work on every login check
- control renew frequency with threshold and throttling
- improve stability under high concurrency with a worker pool

## Core Design

### Asynchronous Renew Strategy

In the current version, auto-renew happens inside the login validation flow, with `Manager.checkLoginInternal()` as the key entry.

Compared with the old "plain goroutine on every `IsLogin()`" idea, the current implementation:

1. checks the token TTL first
2. only considers renew when TTL is less than or equal to `RenewMaxRefresh`
3. uses a renew marker key to throttle repeated renews when `RenewInterval` is configured
4. prefers submitting renew work to a worker pool
5. updates the active timestamp asynchronously as well

### Implementation Idea

```go
if m.config.AutoRenew && m.config.Timeout > 0 {
    if ttl <= RenewMaxRefresh && renewInterval condition passes {
        renewFunc := func() {
            m.renewFunc(context.Background(), tokenValue, tokenInfo.LoginID)
        }

        if m.pool != nil {
            _ = m.pool.Submit(renewFunc)
        } else {
            go renewFunc()
        }
    }
}

if m.config.ActiveTimeout > 0 {
    activeFunc := func() {
        _ = m.storage.Set(ctx, m.getActiveKey(tokenValue), time.Now().Unix(), m.getExpiration())
    }
    // also prefers the worker pool
}
```

## Workflow

### Synchronous Part

```text
1. Load TokenInfo
   ├─ failed -> return not-login or token-state error
   └─ success -> continue

2. Check account disable state
   ├─ disabled -> return disable error
   └─ not disabled -> continue

3. Check ActiveTimeout
   ├─ timeout -> kick out and return error
   └─ not timeout -> continue

4. Return login validation success
```

### Asynchronous Part

```text
Async renew task
  ↓
1. Extend token expiration
  ↓
2. Extend session expiration
  ↓
3. Write renew throttle marker (if enabled)
  ↓
4. Trigger renew event

Async active task
  ↓
1. Update active:{token} key
```

## Renew Trigger Conditions

In the current implementation, auto-renew typically requires all of the following:

- `AutoRenew = true`
- `Timeout > 0`
- current token TTL is greater than 0
- `TTL <= RenewMaxRefresh`, or `RenewMaxRefresh <= 0`
- the renew-throttle condition of `RenewInterval` is not hit

This means:

- renew does not happen on every request
- renew is more likely when the token is closer to expiry
- frequent requests do not endlessly refresh the same token

## Trigger Timing

Any scenario that goes through login validation may trigger auto-renew:

### 1. Middleware Authentication

```go
r.Use(satoken.AuthMiddleware(ctx))
```

### 2. Annotation-Style Login Check

```go
annotation.GET("/profile",
    satoken.CheckLoginMiddleware(ctx, handleProfile, handleAuthFail))
```

### 3. Manual Login Check

```go
stputil.IsLogin(ctx, token)
stputil.CheckLogin(ctx, token)
```

### 4. Fetching Login Information

```go
stputil.GetLoginID(ctx, token)
stputil.GetTokenInfo(ctx, token)
```

## Configuration Options

### Enable Auto-Renew

```go
builder.NewBuilder().
    SetStorage(memory.NewStorage()).
    Timeout(86400).
    AutoRenew(true).
    Build()
```

### Set Renew Trigger Threshold

```go
builder.NewBuilder().
    SetStorage(memory.NewStorage()).
    Timeout(86400).
    AutoRenew(true).
    RenewMaxRefresh(3600). // only renew within the last hour
    Build()
```

### Set Minimum Renew Interval

```go
builder.NewBuilder().
    SetStorage(memory.NewStorage()).
    Timeout(86400).
    AutoRenew(true).
    RenewMaxRefresh(3600).
    RenewInterval(300). // at most once every 5 minutes for the same token
    Build()
```

### Combine with Active Timeout

```go
builder.NewBuilder().
    SetStorage(memory.NewStorage()).
    Timeout(86400).
    ActiveTimeout(1800).
    AutoRenew(true).
    RenewMaxRefresh(3600).
    RenewInterval(300).
    Build()
```

**Effect**:
- active users are auto-renewed when close to expiry
- inactive users are kicked out after `ActiveTimeout`
- renew work is not executed without limit under high request frequency

## Concurrency Safety

### Storage Interface Is Context-Aware

```go
type Storage interface {
    Set(ctx context.Context, key string, value any, expiration time.Duration) error
    Get(ctx context.Context, key string) (any, error)
    Delete(ctx context.Context, keys ...string) error
    Exists(ctx context.Context, key string) bool
    TTL(ctx context.Context, key string) (time.Duration, error)
}
```

### Worker Pool Support

When auto-renew is enabled, renew work prefers a renew task pool:

- `com/pool/ants`
- `adapter.Pool`
- `Builder.SetPool(...)`

If no pool is explicitly provided, a default renew pool may still be created internally in suitable cases.

## Renew Failure Handling

### Strategy

Async renew failure does not directly block the current request:

1. the current login validation result has already been returned
2. renew failure only affects the extension of future lifetime
3. renew can still be retried on the next eligible request

### Impact Scope

- it does not directly fail the current request
- it may leave the token unextended
- if renew keeps failing, the token will eventually expire by its original TTL

## Best Practices

### Recommended Production Configuration

```go
builder.NewBuilder().
    SetStorage(redisStorage).
    Timeout(86400).
    ActiveTimeout(1800).
    AutoRenew(true).
    RenewMaxRefresh(3600).
    RenewInterval(300).
    Build()
```

### Recommended Development Configuration

```go
builder.NewBuilder().
    SetStorage(memory.NewStorage()).
    Timeout(7200).
    AutoRenew(true).
    Build()
```

### Security-First Configuration

```go
builder.NewBuilder().
    SetStorage(redisStorage).
    Timeout(1800).
    AutoRenew(false).
    Build()
```

## Next Steps

- [Architecture Design](architecture.md)
- [Modular Design](modular.md)
- [StpUtil API Documentation](../api/stputil.md)
