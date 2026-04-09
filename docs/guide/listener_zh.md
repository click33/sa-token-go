# 事件监听指南

[English](listener.md) | 中文文档

## 当前状态

当前项目内部已经实现了完整的事件系统，底层代码位于 `core/listener`。

它会在这些场景触发事件：

- 登录
- 登出
- 踢下线
- 顶下线
- 账号封禁 / 解封
- 服务封禁 / 解封
- 自动续期
- Session 创建 / 销毁
- 权限校验
- 角色校验

## 重要说明

虽然事件系统已经存在，但当前版本还**没有**从 `stputil` 或 `manager.Manager` 对外暴露统一的监听器注册入口。

也就是说，下面这种历史写法在当前代码里并不存在：

```go
// 当前版本没有这类公开入口
// manager.RegisterFunc(...)
// manager.GetEventManager(...)
```

因此这篇文档会分成两部分：

1. 当前项目内部已经有哪些事件
2. `core/listener` 包本身有哪些公开能力

## 内部事件类型

当前定义在 `core/listener/consts.go` 中的事件有：

| 事件 | 说明 |
|------|------|
| `EventLogin` | 登录事件 |
| `EventLogout` | 登出事件 |
| `EventKickout` | 踢下线事件 |
| `EventReplace` | 顶下线事件 |
| `EventDisable` | 账号封禁事件 |
| `EventUntie` | 账号解封事件 |
| `EventRenew` | Token 续期事件 |
| `EventCreateSession` | Session 创建事件 |
| `EventDestroySession` | Session 销毁事件 |
| `EventPermissionCheck` | 权限校验事件 |
| `EventRoleCheck` | 角色校验事件 |
| `EventDisableService` | 服务封禁事件 |
| `EventUntieService` | 服务解封事件 |
| `EventAll` | 通配事件 |

## EventData 结构

事件载荷结构在 `core/listener/listener.go` 中定义：

```go
type EventData struct {
    Event     Event
    AuthType  string
    LoginID   string
    Device    string
    DeviceId  string
    Token     string
    Extra     map[string]any
    Timestamp int64
}
```

其中：

- `AuthType`：当前认证体系
- `LoginID`：账号 ID
- `Device` / `DeviceId`：终端信息
- `Token`：相关 token
- `Extra`：额外信息，比如权限校验结果、服务封禁等级等

## Extra 字段常量

当前额外字段常量包括：

- `ExtraKeyPermission`
- `ExtraKeyPermissions`
- `ExtraKeyRole`
- `ExtraKeyRoles`
- `ExtraKeyLogic`
- `ExtraKeyResult`
- `ExtraKeyService`
- `ExtraKeyLevel`

## core/listener 包能力

虽然还没接到 `stputil` 的公开入口上，但 `core/listener` 包本身是可用的。

### 创建监听管理器

```go
import (
    "github.com/click33/sa-token-go/core/listener"
)

eventMgr := listener.NewManager()
```

### 注册监听器

```go
id := eventMgr.RegisterFunc(listener.EventLogin, func(data *listener.EventData) {
    println("login:", data.LoginID)
})

_ = id
```

### 使用配置注册

```go
id := eventMgr.RegisterFuncWithConfig(
    listener.EventLogin,
    func(data *listener.EventData) {
        println("login:", data.LoginID)
    },
    listener.ListenerConfig{
        Async:    true,
        Priority: 100,
        ID:       "login-audit",
    },
)
```

### 注销监听器

```go
ok := eventMgr.Unregister("login-audit")
_ = ok
```

## 高级能力

### 全局过滤器

```go
eventMgr.AddFilter(func(data *listener.EventData) bool {
    return data.AuthType == "satoken:"
})
```

过滤器返回 `false` 时，事件不会继续分发。

### 统计信息

```go
eventMgr.EnableStats(true)

stats := eventMgr.GetStats()
println(stats.TotalTriggered)
```

### Panic 处理

```go
eventMgr.SetPanicHandler(func(event listener.Event, data *listener.EventData, recovered any) {
    println("listener panic:", event, recovered)
})
```

### 启用 / 禁用事件

```go
eventMgr.DisableEvent(listener.EventRenew, listener.EventPermissionCheck)
eventMgr.EnableEvent(listener.EventLogin, listener.EventLogout)
eventMgr.EnableEvent() // 不传参表示启用全部
```

### 等待异步监听器完成

```go
eventMgr.Wait()
```

## 逻辑常量

```go
listener.LogicAnd
listener.LogicOr
```

它们通常出现在权限 / 角色校验事件的 `Extra` 字段里。

## 现阶段建议

如果你只是使用 `stputil` 现成的登录能力：

1. 可以先把这套事件文档当成“内部机制说明”
2. 如果后续项目把监听器注册入口导出，再直接接业务审计、告警、统计

如果你准备二次开发：

1. 可以直接参考 `core/listener` 包
2. 再结合 `manager.triggerEvent(...)` 这条内部链路做扩展

## 相关文档

- [登录认证](authentication_zh.md)
- [权限管理](permission_zh.md)
- [OAuth2 指南](oauth2_zh.md)
