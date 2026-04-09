# Quick Start

[中文文档](quick-start_zh.md) | English

## Get Started with Sa-Token-Go in 5 Minutes

This page is based on the current version of the codebase and walks through the minimum steps for installation, initialization, and basic usage.

## Step 1: Install

### Option 1: Core Modules + Memory Storage

```bash
go get github.com/click33/sa-token-go/core
go get github.com/click33/sa-token-go/stputil
go get github.com/click33/sa-token-go/com/storage/memory
```

### Option 2: Use a Framework Integration Package Directly

If you are already using a web framework, you can also import an integration package directly:

```bash
go get github.com/click33/sa-token-go/integrations/gin
go get github.com/click33/sa-token-go/com/storage/memory
```

## Step 2: Initialize

```go
package main

import (
    "context"

    "github.com/click33/sa-token-go/com/storage/memory"
    "github.com/click33/sa-token-go/core/builder"
    "github.com/click33/sa-token-go/stputil"
)

var ctx = context.Background()

func init() {
    stputil.SetManager(
        builder.NewBuilder().
            SetStorage(memory.NewStorage()).
            TokenName("Authorization").
            Timeout(86400).
            Build(),
    )
}
```

### Initialization Notes

- `builder.NewBuilder()` is used to build a `Manager`
- `SetStorage(...)` specifies the storage implementation
- `stputil.SetManager(...)` registers the global `Manager`
- after that, application code can call capability directly through `stputil`

## Step 3: Use It

```go
func main() {
    // Login
    token, _ := stputil.Login(ctx, "1000")

    // Check login state
    isLogin := stputil.IsLogin(ctx, token)

    // Add permissions
    _ = stputil.AddPermissions(ctx, "1000", []string{"user:read"})

    // Check permission
    hasPermission := stputil.HasPermission(ctx, "1000", "user:read")

    // Logout
    _ = stputil.Logout(ctx, token)

    _, _, _ = token, isLogin, hasPermission
}
```

## Step 4: Common Configuration

You can continue adjusting common settings through the Builder:

```go
mgr := builder.NewBuilder().
    SetStorage(memory.NewStorage()).
    TokenName("token").
    Timeout(7200).
    ActiveTimeout(1800).
    AutoRenew(true).
    IsReadHeader(true).
    IsPrintBanner(true).
    Build()

stputil.SetManager(mgr)
```

Common config meanings:

- `TokenName` - token name
- `Timeout` - absolute token timeout
- `ActiveTimeout` - maximum inactive duration
- `AutoRenew` - whether auto-renew is enabled
- `IsReadHeader` - whether to read token from HTTP header
- `IsPrintBanner` - whether to print the startup banner

## Step 5: Read Complete Examples

If you want to see more complete examples, you can directly refer to:

- [Quick Start example](../../examples/quick_start/)
- [Gin example](../../examples/gin/)
- [GoFrame example](../../examples/gf/)
- [Echo example](../../examples/echo/)
- [Fiber example](../../examples/fiber/)
- [Chi example](../../examples/chi/)
- [Hertz example](../../examples/hertz/)
- [Kratos example](../../examples/kratos/)

After these steps, you already understand the most basic usage of the current Sa-Token-Go version.

## Next Steps

- [Authentication Guide](../guide/authentication.md)
- [Permission Guide](../guide/permission.md)
- [Annotation Guide](../guide/annotation.md)
- [Single Import Guide](../guide/single-import.md)
