# Permission Management

[中文文档](permission_zh.md) | English

## Set Permissions

```go
// Set user permissions
stputil.SetPermissions(1000, []string{
    "user:read",
    "user:write",
    "admin:*",      // Wildcard: matches all admin permissions
})
```

## Check Permissions

### Single Permission

```go
// Check if has permission
hasPermission := stputil.HasPermission(1000, "user:read")
```

### Multiple Permissions (AND)

```go
// Check if has all permissions
hasAll := stputil.HasPermissionsAnd(1000, []string{
    "user:read",
    "user:write",
})
```

### Multiple Permissions (OR)

```go
// Check if has any permission
hasAny := stputil.HasPermissionsOr(1000, []string{
    "admin:read",
    "admin:write",
})
```

## Wildcard Support

```go
// Set wildcard permissions
stputil.SetPermissions(1000, []string{
    "admin:*",          // All admin permissions
    "user:*:view",      // All user view permissions
    "*",                // All permissions
})

// Wildcard matching
stputil.HasPermission(1000, "admin:read")    // ✅ Match admin:*
stputil.HasPermission(1000, "admin:delete")  // ✅ Match admin:*
stputil.HasPermission(1000, "user:1:view")   // ✅ Match user:*:view
```

## Get Permissions

```go
// Get user permissions list
permissions, err := stputil.GetPermissions(1000)
for _, perm := range permissions {
    fmt.Println(perm)
}
```

## Permission Patterns

### Resource-based Permissions

```go
"user:read"         // Read user
"user:write"        // Write user
"user:delete"       // Delete user
"user:*"            // All user operations
```

### Hierarchical Permissions

```go
"system:user:read"          // Read system users
"system:user:*"             // All system user operations
"system:*"                  // All system operations
```

### Action-based Permissions

```go
"create:post"       // Create post
"edit:post"         // Edit post
"delete:post"       // Delete post
"*:post"            // All post operations
```

## Complete Example

```go
package main

import (
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/storage/memory"
)

func main() {
    // Initialize
    stputil.SetManager(
        core.NewBuilder().
            Storage(memory.NewStorage()).
            Build(),
    )

    // Login
    token, _ := stputil.Login(1000)

    // Set permissions
    stputil.SetPermissions(1000, []string{
        "user:read",
        "user:write",
        "post:*",
        "admin:*",
    })

    // Check permissions
    if stputil.HasPermission(1000, "user:read") {
        println("✅ Can read user")
    }

    if stputil.HasPermission(1000, "post:delete") {
        println("✅ Can delete post (wildcard match)")
    }

    // Check multiple permissions
    if stputil.HasPermissionsAnd(1000, []string{"user:read", "user:write"}) {
        println("✅ Can read and write user")
    }
}
```

## Related Documentation

- [Quick Start](../tutorial/quick-start.md)
- [Authentication Guide](authentication.md)
- [Annotation Usage](annotation.md)
