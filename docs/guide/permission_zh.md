# 权限管理

[English](permission.md) | 中文文档

## 概览

当前版本的权限与角色数据主要有两种来源：

1. `Session` 中维护的 `Permissions`、`Roles`
2. 通过 `builder.NewBuilder()` 注入的自定义回调

优先级方面：

- 按 `loginID` 查询时：`CustomPermissionListFunc` / `CustomRoleListFunc` 优先于 Session
- 按 `token` 查询时：`CustomPermissionListExtFunc` / `CustomRoleListExtFunc` 优先，其次普通回调，最后才是 Session

## 初始化

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

## 权限管理

### 添加权限

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

### 删除权限

```go
ctx := context.Background()

_ = stputil.RemovePermissions(ctx, "10001", []string{"user:write"})
_ = stputil.RemovePermissionsByToken(ctx, token, []string{"article:publish"})
```

### 查询权限

```go
ctx := context.Background()

permissions, err := stputil.GetPermissions(ctx, "10001")
permissionsByToken, err := stputil.GetPermissionsByToken(ctx, token)
```

## 权限校验

### 单个权限

```go
ctx := context.Background()

hasPermission := stputil.HasPermission(ctx, "10001", "user:read")
hasPermissionByToken := stputil.HasPermissionByToken(ctx, token, "user:read")

err := stputil.CheckPermission(ctx, "10001", "user:read")
```

### 多权限 AND

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

### 多权限 OR

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

## 权限通配符

当前权限匹配支持 `*` 通配符，并按分段匹配：

| 模式 | 说明 | 示例 |
|------|------|------|
| `*` | 匹配全部权限 | 任意权限 |
| `user:*` | 匹配两段式 `user` 权限 | `user:read`、`user:write` |
| `user:*:view` | 匹配三段式权限 | `user:profile:view` |
| `admin/*` | 也支持 `/` 作为分隔符 | `admin/read` |

当前实现的两个关键点：

1. 分隔符会自动根据权限模式识别，优先使用 `:`，若模式里包含 `/` 则使用 `/`
2. 分段数量必须一致，避免因为通配过宽造成越权

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

## 角色管理

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

## 封禁管理

### 账号封禁

```go
ctx := context.Background()

_ = stputil.Disable(ctx, "10001", 2*time.Hour, "abuse")

disabled := stputil.IsDisable(ctx, "10001")
err := stputil.CheckDisable(ctx, "10001")

info, err := stputil.GetDisableInfo(ctx, "10001")
ttl, err := stputil.GetDisableTTL(ctx, "10001")

_ = stputil.Untie(ctx, "10001")
```

`GetDisableTTL()` 返回值约定：

- `-2`：未封禁
- `-1`：永久封禁
- `>0`：剩余秒数

### 服务级封禁

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

## 自定义权限与角色回调

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

扩展回调还支持拿到 `device`、`deviceId`：

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

## Gin 路由中使用

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

## 最佳实践

1. 先登录，再维护角色和权限，避免因为 Session 不存在导致写入失败
2. 业务权限变化频繁时，优先使用自定义回调而不是长期缓存到 Session
3. 对外接口建议使用 `Check*`，内部逻辑判断建议使用 `Has*`
4. 服务级封禁适合评论、发帖、支付等细粒度能力，不必总是封禁整个账号

## 相关文档

- [登录认证](authentication_zh.md)
- [注解鉴权](annotation_zh.md)
- [JWT 指南](jwt_zh.md)
