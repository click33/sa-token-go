# Event Listener Guide

[中文文档](listener_zh.md) | English

## Overview

Sa-Token-Go provides a powerful event system for monitoring authentication and authorization events.

## Event Types

- `EventLogin` - User login event
- `EventLogout` - User logout event
- `EventKickout` - User kicked out event
- `EventDisable` - Account disabled event
- `EventUntie` - Account re-enabled event
- `EventRenew` - Token renewal event
- `EventCreateSession` - Session created event
- `EventDestroySession` - Session destroyed event
- `EventPermissionCheck` - Permission check event
- `EventRoleCheck` - Role check event
- `EventAll` - Wildcard (all events)

## Basic Usage

### Create Event Manager

```go
import "github.com/click33/sa-token-go/core"

eventMgr := core.NewEventManager()
```

### Register Listener (Function)

```go
// Listen to login event
eventMgr.RegisterFunc(core.EventLogin, func(data *core.EventData) {
    fmt.Printf("[LOGIN] User: %s, Token: %s\n", data.LoginID, data.Token)
})

// Listen to logout event
eventMgr.RegisterFunc(core.EventLogout, func(data *core.EventData) {
    fmt.Printf("[LOGOUT] User: %s\n", data.LoginID)
})
```

### Register Listener (Interface)

```go
type AuditLogger struct{}

func (a *AuditLogger) OnEvent(data *core.EventData) {
    // Log to database, file, etc.
    fmt.Printf("[AUDIT] Event: %s, User: %s\n", data.Event, data.LoginID)
}

eventMgr.Register(core.EventLogin, &AuditLogger{})
```

## Advanced Features

### Priority

```go
// Higher priority listeners execute first
eventMgr.RegisterWithConfig(
    core.EventLogin,
    myListener,
    core.ListenerConfig{
        Priority: 100,  // Higher = earlier execution
    },
)
```

### Synchronous Execution

```go
// Execute synchronously (blocking)
eventMgr.RegisterWithConfig(
    core.EventLogin,
    myListener,
    core.ListenerConfig{
        Async: false,  // Synchronous
    },
)
```

### Wildcard Listener

```go
// Listen to all events
eventMgr.RegisterFunc(core.EventAll, func(data *core.EventData) {
    fmt.Printf("[ALL] Event: %s, User: %s\n", data.Event, data.LoginID)
})
```

### Unregister Listener

```go
// Register and get ID
id := eventMgr.RegisterWithConfig(
    core.EventLogin,
    myListener,
    core.ListenerConfig{
        ID: "my-listener",
    },
)

// Unregister by ID
eventMgr.Unregister(id)
```

## Use Cases

### Audit Logging

```go
eventMgr.RegisterFunc(core.EventAll, func(data *core.EventData) {
    log.Printf("[AUDIT] %s - User: %s, IP: %s, Time: %d",
        data.Event, data.LoginID, data.Extra["ip"], data.Timestamp)
})
```

### Security Monitoring

```go
eventMgr.RegisterFunc(core.EventLogin, func(data *core.EventData) {
    // Check for suspicious login
    // Send alert if needed
})
```

### Session Analytics

```go
eventMgr.RegisterFunc(core.EventLogin, func(data *core.EventData) {
    // Track active users
    // Update analytics
})
```

## Related Documentation

- [Quick Start](../tutorial/quick-start.md)
- [Authentication Guide](authentication.md)
- [Event Listener Example](../../examples/listener-example/README.md)
