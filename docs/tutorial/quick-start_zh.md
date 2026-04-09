# 快速开始

[English](quick-start.md) | 中文文档

## 5分钟上手 Sa-Token-Go

这一节基于当前版本代码，带你用最小步骤完成安装、初始化和基础调用。

## 步骤1：安装

### 方式一：核心模块 + 内存存储

```bash
go get github.com/click33/sa-token-go/core
go get github.com/click33/sa-token-go/stputil
go get github.com/click33/sa-token-go/com/storage/memory
```

### 方式二：直接使用框架集成包

如果你本身就在使用 Web 框架，也可以直接导入集成包：

```bash
go get github.com/click33/sa-token-go/integrations/gin
go get github.com/click33/sa-token-go/com/storage/memory
```

## 步骤2：初始化

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

### 初始化说明

- `builder.NewBuilder()` 用来构建 `Manager`
- `SetStorage(...)` 用来指定存储实现
- `stputil.SetManager(...)` 用来注册全局 `Manager`
- 后续业务中就可以直接通过 `stputil` 调用能力

## 步骤3：使用

```go
func main() {
    // 登录
    token, _ := stputil.Login(ctx, "1000")

    // 检查登录
    isLogin := stputil.IsLogin(ctx, token)

    // 添加权限
    _ = stputil.AddPermissions(ctx, "1000", []string{"user:read"})

    // 检查权限
    hasPermission := stputil.HasPermission(ctx, "1000", "user:read")

    // 登出
    _ = stputil.Logout(ctx, token)

    _, _, _ = token, isLogin, hasPermission
}
```

## 步骤4：常见配置

你可以继续通过 Builder 调整常见配置：

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

常见配置含义：

- `TokenName` - Token 名称
- `Timeout` - Token 绝对超时时间
- `ActiveTimeout` - 最大不活跃时长
- `AutoRenew` - 是否自动续签
- `IsReadHeader` - 是否从 Header 中读取 Token
- `IsPrintBanner` - 是否打印启动 Banner

## 步骤5：查看完整示例

如果你想看更完整的示例，可以直接参考：

- [Quick Start 示例](../../examples/quick_start/)
- [Gin 示例](../../examples/gin/)
- [GoFrame 示例](../../examples/gf/)
- [Echo 示例](../../examples/echo/)
- [Fiber 示例](../../examples/fiber/)
- [Chi 示例](../../examples/chi/)
- [Hertz 示例](../../examples/hertz/)
- [Kratos 示例](../../examples/kratos/)

完成以上步骤后，你已经掌握了当前版本 Sa-Token-Go 的最基础使用方式。

## 下一步

- [登录认证详解](../guide/authentication_zh.md)
- [权限验证详解](../guide/permission_zh.md)
- [注解使用](../guide/annotation_zh.md)
- [单包导入](../guide/single-import_zh.md)
