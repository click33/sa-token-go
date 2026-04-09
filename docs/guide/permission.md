# Permission Management

[中文文档](permission_zh.md) | English

## Overview

In the current version, permissions and roles come from two places:

1. `Session` fields: `Permissions` and `Roles`
2. Custom callbacks configured through `builder.NewBuilder()`

Priority rules:

- By `loginID`: `CustomPermissionListFunc` / `CustomRoleListFunc` override session data
- By `token`: `CustomPermissionListExtFunc` / `CustomRoleListExtFunc` override normal callbacks, which override session data

## Initialization

```go
package main

import (
    "context"

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
```

## Permission APIs

### Add Permissions

```go
ctx := context.Background()

token, _ := stputil.Login(ctx, "10001")

_ = stputil.AddPermissions(ctx, "10001", []string{
    "user:read",
    "user:write",
    "admin:*",
})

_ = stputil.AddPermissionsByToken(ctx, token, []string{
    "article:publish",
})
```

### Remove Permissions

```go
ctx := context.Background()

_ = stputil.RemovePermissions(ctx, "10001", []string{"user:write"})
_ = stputil.RemovePermissionsByToken(ctx, token, []string{"article:publish"})
```

### Query Permissions

```go
ctx := context.Background()

permissions, err := stputil.GetPermissions(ctx, "10001")
permissionsByToken, err := stputil.GetPermissionsByToken(ctx, token)
```

## Permission Checks

### Single Permission

```go
ctx := context.Background()

hasPermission := stputil.HasPermission(ctx, "10001", "user:read")
hasPermissionByToken := stputil.HasPermissionByToken(ctx, token, "user:read")

err := stputil.CheckPermission(ctx, "10001", "user:read")
```

### AND Logic

```go
ctx := context.Background()

hasAll := stputil.HasPermissionsAnd(ctx, "10001", []string{
    "user:read",
    "user:write",
})

err := stputil.CheckPermissionAnd(ctx, "10001", []string{
    "user:read",
    "user:write",
})
```

### OR Logic

```go
ctx := context.Background()

hasAny := stputil.HasPermissionsOr(ctx, "10001", []string{
    "admin:read",
    "report:read",
})

err := stputil.CheckPermissionOr(ctx, "10001", []string{
    "admin:read",
    "report:read",
})
```

## Wildcard Matching

Wildcard matching supports `*` and works segment by segment:

| Pattern | Description | Example |
|------|------|------|
| `*` | matches all permissions | any permission |
| `user:*` | matches two-part `user` permissions | `user:read`, `user:write` |
| `user:*:view` | matches three-part permissions | `user:profile:view` |
| `admin/*` | `/` is also supported as a separator | `admin/read` |

Two important implementation details:

1. The separator is detected automatically, defaulting to `:` and switching to `/` when the pattern contains `/`
2. Segment counts must be equal, which prevents overly broad matches

```go
ctx := context.Background()

_ = stputil.AddPermissions(ctx, "10001", []string{
    "admin:*",
    "user:*:view",
})

stputil.HasPermission(ctx, "10001", "admin:read")        // true
stputil.HasPermission(ctx, "10001", "admin:delete")      // true
stputil.HasPermission(ctx, "10001", "user:profile:view") // true
stputil.HasPermission(ctx, "10001", "user:view")         // false
```

## Role APIs

```go
ctx := context.Background()

_ = stputil.AddRoles(ctx, "10001", []string{"admin", "editor"})
_ = stputil.AddRolesByToken(ctx, token, []string{"reviewer"})

_ = stputil.RemoveRoles(ctx, "10001", []string{"editor"})
_ = stputil.RemoveRolesByToken(ctx, token, []string{"reviewer"})

roles, err := stputil.GetRoles(ctx, "10001")
rolesByToken, err := stputil.GetRolesByToken(ctx, token)

hasRole := stputil.HasRole(ctx, "10001", "admin")
hasAnyRole := stputil.HasRolesOr(ctx, "10001", []string{"admin", "manager"})
hasAllRoles := stputil.HasRolesAnd(ctx, "10001", []string{"admin", "editor"})

err = stputil.CheckRole(ctx, "10001", "admin")
err = stputil.CheckRoleOr(ctx, "10001", []string{"admin", "manager"})
err = stputil.CheckRoleAnd(ctx, "10001", []string{"admin", "editor"})
```

## Disable APIs

### Account Disable

```go
ctx := context.Background()

_ = stputil.Disable(ctx, "10001", 2*time.Hour, "abuse")

disabled := stputil.IsDisable(ctx, "10001")
err := stputil.CheckDisable(ctx, "10001")

info, err := stputil.GetDisableInfo(ctx, "10001")
ttl, err := stputil.GetDisableTTL(ctx, "10001")

_ = stputil.Untie(ctx, "10001")
```

`GetDisableTTL()` returns:

- `-2`: not disabled
- `-1`: permanently disabled
- `>0`: remaining seconds

### Service-Level Disable

```go
ctx := context.Background()

_ = stputil.DisableService(ctx, "10001", "comment", 30*time.Minute)
_ = stputil.DisableServiceWithReason(ctx, "10001", "comment", 30*time.Minute, "spam")
_ = stputil.DisableServiceLevel(ctx, "10001", "comment", 2, 30*time.Minute)
_ = stputil.DisableServiceLevelWithReason(ctx, "10001", "comment", 3, 30*time.Minute, "risk")

serviceDisabled := stputil.IsDisableService(ctx, "10001", "comment")
levelDisabled := stputil.IsDisableServiceLevel(ctx, "10001", "comment", 2)

err := stputil.CheckDisableService(ctx, "10001", []string{"comment", "post"})
err = stputil.CheckDisableServiceLevel(ctx, "10001", "comment", 2)

serviceInfo, err := stputil.GetDisableServiceInfo(ctx, "10001", "comment")
serviceTTL, err := stputil.GetDisableServiceTTL(ctx, "10001", "comment")

_ = stputil.UntieService(ctx, "10001", "comment")
```

## Custom Permission and Role Callbacks

```go
stputil.SetManager(
    builder.NewBuilder().
        SetStorage(memory.NewStorage()).
        SetCustomPermissionListFunc(func(loginID, authType string) ([]string, error) {
            if loginID == "10001" {
                return []string{"user:read", "user:write"}, nil
            }
            return []string{"user:read"}, nil
        }).
        SetCustomRoleListFunc(func(loginID, authType string) ([]string, error) {
            if loginID == "10001" {
                return []string{"admin"}, nil
            }
            return []string{"user"}, nil
        }).
        Build(),
)
```

Extended callbacks can also use `device` and `deviceId`:

```go
builder.NewBuilder().
    SetCustomPermissionListExtFunc(func(loginID, device, deviceId, authType string) ([]string, error) {
        if device == "app" {
            return []string{"mobile:read", "mobile:write"}, nil
        }
        return []string{"web:read"}, nil
    }).
    SetCustomRoleListExtFunc(func(loginID, device, deviceId, authType string) ([]string, error) {
        if device == "app" {
            return []string{"mobile-user"}, nil
        }
        return []string{"web-user"}, nil
    })
```

## Using With Gin Routes

```go
ctx := context.Background()

r.Use(sagin.RegisterSaTokenContextMiddleware(ctx))

r.GET("/users",
    sagin.CheckPermissionMiddleware(ctx, []string{"user:read"}, nil, nil),
    listUsersHandler,
)

r.POST("/admin",
    sagin.CheckRoleMiddleware(ctx, []string{"admin"}, nil, nil),
    adminHandler,
)
```

## Best Practices

1. Login before writing roles or permissions, otherwise the session may not exist yet
2. Use custom callbacks when permission data changes frequently
3. Use `Check*` APIs for request guards and `Has*` APIs for internal branching
4. Use service-level disable for capabilities such as comments, posts, or payment instead of disabling the full account every time

## Related Documentation

- [Authentication Guide](authentication.md)
- [Annotation Guide](annotation.md)
- [JWT Guide](jwt.md)
