English | [中文文档](nonce_zh.md)

# Nonce Anti-Replay

## Overview

The project already includes built-in nonce support to prevent replayed requests.

Public APIs:

- `GenerateNonce`
- `GenerateNonceWithTimeout`
- `VerifyNonce`
- `VerifyAndConsumeNonce`
- `IsNonceValid`
- `GetNonceTTL`

## How It Works

The current implementation works like this:

1. generate a random nonce
2. store it with a TTL
3. verify it through an atomic `GetAndDelete`
4. allow it to succeed only once

## Default Behavior

Based on the current `core/nonce` implementation:

- the raw nonce uses `32` random bytes
- the output is a `64`-character hexadecimal string
- the default TTL is `5` minutes

## Basic Usage

```go
package main

import (
    "context"
    "fmt"

    "github.com/click33/sa-token-go/com/storage/memory"
    "github.com/click33/sa-token-go/core/builder"
    "github.com/click33/sa-token-go/stputil"
)

func initSaToken() {
    stputil.SetManager(
        builder.NewBuilder().
            SetStorage(memory.NewStorage()).
            Build(),
    )
}

func main() {
    ctx := context.Background()

    nonce, _ := stputil.GenerateNonce(ctx)
    fmt.Println(nonce)

    ok := stputil.VerifyNonce(ctx, nonce)
    fmt.Println(ok) // true

    ok = stputil.VerifyNonce(ctx, nonce)
    fmt.Println(ok) // false
}
```

## Custom TTL

```go
ctx := context.Background()

nonce, err := stputil.GenerateNonceWithTimeout(ctx, 30*time.Second)
_ = nonce
_ = err
```

If the timeout is less than or equal to `0`, the implementation falls back to the default TTL.

## Non-Consuming Validation

```go
ctx := context.Background()

nonce, _ := stputil.GenerateNonce(ctx)

valid := stputil.IsNonceValid(ctx, nonce) // validate only
err := stputil.VerifyAndConsumeNonce(ctx, nonce)
```

Difference:

- `IsNonceValid`: validate only
- `VerifyNonce`: validate and consume, returning `bool`
- `VerifyAndConsumeNonce`: validate and consume, returning `ErrInvalidNonce` on failure

## TTL Query

```go
ctx := context.Background()

ttl, err := stputil.GetNonceTTL(ctx, nonce)
```

Return values:

- `-2`: nonce does not exist
- `-1`: no expiration
- `>=0`: remaining seconds

## HTTP Example

```go
r.GET("/nonce", func(c *gin.Context) {
    nonce, err := stputil.GenerateNonce(ctx)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"nonce": nonce})
})

r.POST("/transfer", func(c *gin.Context) {
    nonce := c.GetHeader("X-Nonce")

    if err := stputil.VerifyAndConsumeNonce(ctx, nonce); err != nil {
        c.JSON(401, gin.H{"error": "invalid_nonce"})
        return
    }

    c.JSON(200, gin.H{"message": "ok"})
})
```

## Best Practices

1. Use nonce for sensitive write operations such as payment, transfer, password change, and delete
2. Combine nonce checks with normal login checks instead of using nonce alone
3. A TTL around 5 minutes is usually enough for form submissions and one-time confirmations
4. If the client needs a pre-check, use `IsNonceValid` first and only consume on final submit

## Related Documentation

- [OAuth2 Guide](oauth2.md)
- [Authentication Guide](authentication.md)
- [Refresh Token Guide](refresh-token.md)
