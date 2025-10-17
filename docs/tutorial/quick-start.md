# Quick Start

[中文文档](quick-start_zh.md) | English

## Get Started with Sa-Token-Go in 5 Minutes

### Step 1: Installation

```bash
go get github.com/click33/sa-token-go/core
go get github.com/click33/sa-token-go/stputil
go get github.com/click33/sa-token-go/storage/memory
```

### Step 2: Initialize

```go
import (
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/storage/memory"
)

func init() {
    // One-line initialization!
    stputil.SetManager(
        core.NewBuilder().
            Storage(memory.NewStorage()).
            Build(),
    )
}
```

### Step 3: Use

```go
func main() {
    // Login
    token, _ := stputil.Login(1000)
    println("Token:", token)

    // Check login
    if stputil.IsLogin(token) {
        println("User is logged in")
    }

    // Set permissions
    stputil.SetPermissions(1000, []string{"user:read", "user:write"})

    // Check permission
    if stputil.HasPermission(1000, "user:read") {
        println("Has permission")
    }

    // Logout
    stputil.Logout(1000)
}
```

## Next Steps

- [Authentication Guide](../guide/authentication.md)
- [Permission Management](../guide/permission.md)
- [Annotation Usage](../guide/annotation.md)
- [Event Listener](../guide/listener.md)
- [JWT Guide](../guide/jwt.md)
- [Redis Storage](../guide/redis-storage.md)
