# 快速开始示例

这是一个最简单的 Sa-Token-Go 使用示例，展示了如何使用 `StpUtil` 全局工具类快速实现认证和授权功能。

## 运行示例

```bash
go run main.go
```

## 示例说明

本示例展示了以下功能：

1. **一行初始化** - 使用 Builder 模式快速配置
2. **登录认证** - 支持多种类型的用户 ID
3. **检查登录** - 验证用户登录状态
4. **权限管理** - 设置和检查用户权限
5. **角色管理** - 设置和检查用户角色
6. **Session 管理** - 存储和读取会话数据
7. **账号封禁** - 临时封禁用户
8. **Token 信息** - 查看 Token 详细信息
9. **登出** - 清除用户登录状态

## 核心代码

```go
import (
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/storage/memory"
)

func init() {
    // 🎯 一行初始化！
    stputil.SetManager(
        core.NewBuilder().
            Storage(memory.NewStorage()).
            TokenName("Authorization").
            Timeout(86400).  // 24小时
            TokenStyle(core.TokenStyleRandom64).
            Build(),
    )
}

func main() {
    // 登录
    token, _ := stputil.Login(1000)
    
    // 设置权限
    stputil.SetPermissions(1000, []string{"user:read", "user:write"})
    
    // 检查权限
    hasPermission := stputil.HasPermission(1000, "user:read")
    
    // 登出
    stputil.Logout(1000)
}
```

## 输出示例

```
=== Sa-Token-Go 简洁使用示例 ===

1. 登录测试
   用户1000登录成功，Token: xxx
   用户user123登录成功，Token: yyy

2. 检查登录
   Token1是否登录: true
   Token2是否登录: true

3. 获取登录ID
   Token1的登录ID: 1000
   Token2的登录ID: user123

4. 权限管理
   已设置权限: user:read, user:write, admin:*
   是否有user:read权限: true
   是否有user:delete权限: false
   是否有admin:delete权限(通配符): true

...
```

## 扩展学习

- [Gin 集成示例](../../gin/gin-example) - 学习如何在 Gin 框架中使用
- [注解装饰器示例](../../annotation/annotation-example) - 学习注解式编程
- [完整文档](../../../docs) - 查看详细的 API 文档

