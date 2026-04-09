# 注解鉴权指南

[English](annotation.md) | 中文文档

## 概览

当前项目里的“注解式鉴权”本质上是各框架封装出来的装饰器 / 中间件组合，不是 Go 原生注解。

现在每个集成包都提供了统一思路：

- `RegisterSaTokenContextMiddleware(...)`
- `GetHandler(...)`
- `CheckLoginMiddleware(...)`
- `CheckRoleMiddleware(...)`
- `CheckPermissionMiddleware(...)`
- `CheckDisableMiddleware(...)`
- `IgnoreMiddleware(...)`

## Annotation 结构

以 `integrations/gin` 为例，当前 `Annotation` 支持这些字段：

| 字段 | 说明 |
|------|------|
| `AuthType` | 指定认证体系 |
| `CheckLogin` | 是否校验登录 |
| `CheckRole` | 角色列表 |
| `CheckPermission` | 权限列表 |
| `CheckDisable` | 是否校验账号封禁 |
| `Ignore` | 是否忽略鉴权 |
| `LogicType` | 多角色 / 多权限逻辑，支持 `OR`、`AND` |

## Gin 中使用

### 初始化上下文中间件

```go
ctx := context.Background()

r := gin.Default()
r.Use(sagin.RegisterSaTokenContextMiddleware(ctx))
```

后续注解中间件会复用请求里缓存的 `SaTokenContext`。

## 基础用法

### 忽略认证

```go
r.GET("/public",
    sagin.IgnoreMiddleware(ctx, nil, nil),
    func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "public"})
    },
)
```

### 检查登录

```go
r.GET("/user/info",
    sagin.CheckLoginMiddleware(ctx, nil, nil),
    func(c *gin.Context) {
        saCtx, _ := sagin.GetSaTokenContext(c)
        loginID, _ := saCtx.GetLoginID(ctx)
        c.JSON(200, gin.H{"loginId": loginID})
    },
)
```

### 检查权限

```go
r.GET("/admin/users",
    sagin.CheckPermissionMiddleware(ctx, []string{"admin:user:read"}, nil, nil),
    listUsersHandler,
)
```

### 检查角色

```go
r.GET("/admin/dashboard",
    sagin.CheckRoleMiddleware(ctx, []string{"admin"}, nil, nil),
    dashboardHandler,
)
```

### 检查封禁

```go
r.GET("/comment/send",
    sagin.CheckDisableMiddleware(ctx, nil, nil),
    sendCommentHandler,
)
```

## 多角色 / 多权限逻辑

### OR 逻辑

默认是 `OR` 逻辑，只要满足任一角色或任一权限即可。

```go
r.GET("/reports",
    sagin.GetHandler(ctx, nil, nil, &sagin.Annotation{
        CheckLogin:      true,
        CheckRole:       []string{"admin", "manager"},
        CheckPermission: []string{"report:read", "report:*"},
        LogicType:       sagin.LogicOr,
    }),
    reportsHandler,
)
```

### AND 逻辑

```go
r.POST("/admin/publish",
    sagin.GetHandler(ctx, nil, nil, &sagin.Annotation{
        CheckLogin:      true,
        CheckRole:       []string{"admin", "editor"},
        CheckPermission: []string{"article:write", "article:publish"},
        LogicType:       sagin.LogicAnd,
    }),
    publishHandler,
)
```

## 组合使用

```go
r.POST("/super-admin",
    sagin.CheckLoginMiddleware(ctx, nil, nil),
    sagin.CheckRoleMiddleware(ctx, []string{"admin"}, nil, nil),
    sagin.CheckPermissionMiddleware(ctx, []string{"super:*"}, nil, nil),
    superAdminHandler,
)
```

## 路由组使用

```go
ctx := context.Background()

api := r.Group("/api")

userGroup := sagin.AuthGroup(ctx, api.Group("/user"), nil, nil)
userGroup.GET("/profile", profileHandler)

adminGroup := sagin.RoleGroup(ctx, api.Group("/admin"), []string{"admin"}, nil, nil)
adminGroup.GET("/dashboard", dashboardHandler)

permGroup := sagin.PermissionGroup(ctx, api.Group("/report"), []string{"report:read"}, nil, nil)
permGroup.GET("/list", reportListHandler)
```

## 自定义失败处理

```go
failFunc := func(c *gin.Context, err error) {
    c.JSON(200, gin.H{
        "code": 50001,
        "msg":  err.Error(),
    })
}

r.GET("/admin",
    sagin.CheckRoleMiddleware(ctx, []string{"admin"}, nil, failFunc),
    adminHandler,
)
```

## 其他框架

`gf`、`echo`、`fiber`、`chi`、`hertz`、`kratos` 都已经有同风格封装，只是函数签名会贴合各自框架。

核心思路不变：先注册 `SaTokenContext`，再挂注解式中间件。

## 相关文档

- [登录认证](authentication_zh.md)
- [权限管理](permission_zh.md)
- [单包导入](single-import_zh.md)
