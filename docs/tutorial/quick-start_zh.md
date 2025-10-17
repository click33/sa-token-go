# 快速开始

[English](quick-start.md) | 中文文档

## 5分钟上手 Sa-Token-Go

### 步骤1：安装

```bash
go get github.com/click33/sa-token-go/core
go get github.com/click33/sa-token-go/storage/memory
```

### 步骤2：初始化

```go
import (
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/storage/memory"
)

func init() {
    stputil.SetManager(
        core.NewBuilder().
            Storage(memory.NewStorage()).
            Build(),
    )
}
```

### 步骤3：使用

```go
// 登录
token, _ := stputil.Login(1000)

// 检查登录
isLogin := stputil.IsLogin(token)

// 设置权限
stputil.SetPermissions(1000, []string{"user:read"})

// 检查权限
hasPermission := stputil.HasPermission(1000, "user:read")

// 登出
stputil.Logout(1000)
```

完成！你已经掌握了 Sa-Token-Go 的基本用法。

## 下一步

- [登录认证详解](../guide/authentication.md)
- [权限验证详解](../guide/permission.md)
- [Gin框架集成](../guide/gin-integration.md)

