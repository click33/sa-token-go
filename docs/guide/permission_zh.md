# 权限验证

[English](permission.md) | 中文文档

## 设置权限

```go
// 设置用户权限
stputil.SetPermissions(1000, []string{
    "user:read",
    "user:write",
    "user:delete",
})

// 设置管理员权限（使用通配符）
stputil.SetPermissions(2000, []string{
    "admin:*",      // 所有admin权限
    "user:*",       // 所有user权限
})

// 设置超级管理员
stputil.SetPermissions(3000, []string{
    "*",            // 所有权限
})
```

## 检查权限

### 单个权限检查

```go
hasPermission := stputil.HasPermission(1000, "user:read")

if hasPermission {
    // 有权限，执行操作
}
```

### 多权限检查（AND逻辑）

```go
// 需要同时拥有多个权限
hasAll := stputil.HasPermissionsAnd(1000, []string{
    "user:read",
    "user:write",
})
```

### 多权限检查（OR逻辑）

```go
// 拥有其中任一权限即可
hasAny := stputil.HasPermissionsOr(1000, []string{
    "admin:delete",
    "super:delete",
})
```

## 权限通配符

### 基础通配符

| 模式 | 说明 | 匹配示例 |
|------|------|----------|
| `*` | 匹配所有权限 | 任何权限 |
| `user:*` | 匹配user开头的所有权限 | `user:read`, `user:write`, `user:delete` |
| `admin:*` | 匹配admin开头的所有权限 | `admin:read`, `admin:write` |

### 高级通配符

| 模式 | 说明 | 匹配示例 |
|------|------|----------|
| `user:*:view` | 三段式通配符 | `user:profile:view`, `user:settings:view` |
| `*:read` | 所有读权限 | `user:read`, `admin:read`, `article:read` |

### 匹配规则

```go
// 用户拥有权限：["admin:*"]

// 检查权限
stputil.HasPermission(1000, "admin:read")    // true
stputil.HasPermission(1000, "admin:write")   // true
stputil.HasPermission(1000, "admin:delete")  // true
stputil.HasPermission(1000, "user:read")     // false
```

## 在Gin中使用

### 装饰器模式

```go
import sagin "github.com/click33/sa-token-go/integrations/gin"

// 需要user:read权限
r.GET("/users", sagin.CheckPermission("user:read"), listUsersHandler)

// 需要admin:*权限
r.POST("/admin", sagin.CheckPermission("admin:*"), adminHandler)

// 需要任一权限（OR逻辑）
r.GET("/dashboard",
    sagin.CheckPermission("user:read", "admin:read"),
    dashboardHandler)
```

### 手动检查

```go
func handler(c *gin.Context) {
    token := c.GetHeader("Authorization")
    loginID, _ := stputil.GetLoginID(token)
    
    // 手动检查权限
    if !stputil.HasPermission(loginID, "admin:write") {
        c.JSON(403, gin.H{"error": "权限不足"})
        return
    }
    
    // 执行操作
    c.JSON(200, gin.H{"message": "success"})
}
```

## 权限最佳实践

### 1. 权限命名规范

```go
// 推荐格式：<资源>:<操作>
"user:read"       // 读取用户
"user:write"      // 写入用户
"user:delete"     // 删除用户
"article:publish" // 发布文章
"order:cancel"    // 取消订单

// 三段式：<模块>:<资源>:<操作>
"admin:user:read"
"admin:article:delete"
```

### 2. 使用通配符

```go
// 部门管理员：拥有本部门所有权限
stputil.SetPermissions(deptAdmin, []string{
    "dept:user:*",
    "dept:article:*",
})

// 超级管理员：拥有所有权限
stputil.SetPermissions(superAdmin, []string{"*"})
```

### 3. 权限与角色结合

```go
// 设置角色和权限
stputil.SetRoles(1000, []string{"editor"})
stputil.SetPermissions(1000, []string{
    "article:read",
    "article:write",
    "article:publish",
})

// 检查时可以同时检查角色和权限
if stputil.HasRole(1000, "editor") && 
   stputil.HasPermission(1000, "article:publish") {
    // 执行发布操作
}
```

## 下一步

- [角色管理](role.md)
- [注解使用](annotation.md)
- [配置说明](configuration.md)

