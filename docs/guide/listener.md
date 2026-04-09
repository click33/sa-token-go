# Event Listener Guide

[中文文档](listener_zh.md) | English

## Current Status

The project already contains a complete internal event system under `core/listener`.

It is used for:

- login
- logout
- kickout
- replace
- account disable / untie
- service disable / untie
- auto renew
- session create / destroy
- permission checks
- role checks

## Important Note

Even though the event system exists internally, the current version does **not** expose a unified listener-registration API from `stputil` or `manager.Manager`.

So historical examples such as these are not valid in the current codebase:

```go
// Not available as public APIs in the current version
// manager.RegisterFunc(...)
// manager.GetEventManager(...)
```

This guide therefore focuses on:

1. which internal events already exist
2. what the standalone `core/listener` package currently exposes

## Internal Event Types

The events defined in `core/listener/consts.go` are:

| Event | Description |
|------|------|
| `EventLogin` | login |
| `EventLogout` | logout |
| `EventKickout` | kickout |
| `EventReplace` | replace |
| `EventDisable` | account disable |
| `EventUntie` | account untie |
| `EventRenew` | token renew |
| `EventCreateSession` | session create |
| `EventDestroySession` | session destroy |
| `EventPermissionCheck` | permission check |
| `EventRoleCheck` | role check |
| `EventDisableService` | service disable |
| `EventUntieService` | service untie |
| `EventAll` | wildcard |

## EventData Structure

The payload is defined in `core/listener/listener.go`:

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

## Extra Field Constants

Current extra keys include:

- `ExtraKeyPermission`
- `ExtraKeyPermissions`
- `ExtraKeyRole`
- `ExtraKeyRoles`
- `ExtraKeyLogic`
- `ExtraKeyResult`
- `ExtraKeyService`
- `ExtraKeyLevel`

## Standalone core/listener APIs

The package itself is public and usable on its own.

### Create a Manager

```go
import (
    "github.com/click33/sa-token-go/core/listener"
)

eventMgr := listener.NewManager()
```

### Register a Listener

```go
id := eventMgr.RegisterFunc(listener.EventLogin, func(data *listener.EventData) {
    println("login:", data.LoginID)
})

_ = id
```

### Register With Config

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

### Unregister

```go
ok := eventMgr.Unregister("login-audit")
_ = ok
```

## Advanced Features

### Global Filters

```go
eventMgr.AddFilter(func(data *listener.EventData) bool {
    return data.AuthType == "satoken:"
})
```

If the filter returns `false`, the event will not be dispatched any further.

### Statistics

```go
eventMgr.EnableStats(true)

stats := eventMgr.GetStats()
println(stats.TotalTriggered)
```

### Panic Handling

```go
eventMgr.SetPanicHandler(func(event listener.Event, data *listener.EventData, recovered any) {
    println("listener panic:", event, recovered)
})
```

### Enable or Disable Events

```go
eventMgr.DisableEvent(listener.EventRenew, listener.EventPermissionCheck)
eventMgr.EnableEvent(listener.EventLogin, listener.EventLogout)
eventMgr.EnableEvent() // no args means enable all
```

### Wait For Async Listeners

```go
eventMgr.Wait()
```

## Logic Constants

```go
listener.LogicAnd
listener.LogicOr
```

These usually appear in the `Extra` payload of permission and role check events.

## Practical Recommendation

If you are only using `stputil` as-is:

1. treat this guide mainly as documentation of the internal event model
2. connect business audit and monitoring later if the project exposes public registration APIs

If you are extending the framework:

1. you can work directly with `core/listener`
2. and follow the internal `manager.triggerEvent(...)` path for integration work

## Related Documentation

- [Authentication Guide](authentication.md)
- [Permission Management](permission.md)
- [OAuth2 Guide](oauth2.md)
