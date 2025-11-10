# Event Listener Example

This example demonstrates the event listener system in Sa-Token-Go.

## Features Demonstrated

1. **Simple Function Listeners** - Quick way to register event handlers
2. **Wildcard Listeners** - Listen to all events
3. **Priority-Based Execution** - Control listener execution order
4. **Synchronous vs Asynchronous** - Choose blocking or non-blocking listeners
5. **Panic Recovery** - Handle listener errors gracefully
6. **Dynamic Registration** - Add/remove listeners at runtime
7. **Event Enable/Disable** - Control which events are active

## Running the Example

```bash
go run main.go
```

## Expected Output

```
=== Sa-Token-Go Event Listener Example ===

--- Triggering Events ---

[AUDIT] Login audit - User: 1000, Time: 1697234567
[ALL EVENTS] Event{type=login, loginID=1000, device=, timestamp=1697234567}
[LOGIN] User 1000 logged in with token abc123def456...

[AUDIT] Login audit - User: 2000, Time: 1697234568
[ALL EVENTS] Event{type=login, loginID=2000, device=, timestamp=1697234568}
[LOGIN] User 2000 logged in with token xyz789...

[ALL EVENTS] Event{type=logout, loginID=1000, device=, timestamp=1697234569}
[LOGOUT] User 1000 logged out

[ALL EVENTS] Event{type=kickout, loginID=2000, device=, timestamp=1697234570}
[KICKOUT] User 2000 was forcibly logged out

--- Listener Statistics ---
Total listeners: 5
Login listeners: 2
Logout listeners: 1

--- Unregistering audit logger ---
Audit logger unregistered successfully
Remaining listeners: 4

--- Disabling kickout events ---

--- Testing event disable (this should not trigger kickout listener) ---
[ALL EVENTS] Event{type=login, loginID=3000, device=, timestamp=1697234571}
[LOGIN] User 3000 logged in with token...

=== Example Complete ===
```

## Key Concepts

In this example the authentication manager automatically owns an internal event manager:

```go
manager := core.NewBuilder().
    Storage(memory.NewStorage()).
    Build()

eventMgr := manager.GetEventManager() // Advanced controls (stats, enable/disable, panic handler, ...)
```

### Function Listeners

The simplest way to register an event handler:

```go
manager.RegisterFunc(core.EventLogin, func(data *core.EventData) {
    fmt.Printf("User %s logged in\n", data.LoginID)
})
```

### Priority-Based Listeners

Control execution order with priorities:

```go
manager.RegisterWithConfig(core.EventLogin,
    myListener,
    core.ListenerConfig{
        Priority: 100,  // Higher = executes first
        Async:    false, // Synchronous execution
    },
)
```

### Wildcard Listeners

Listen to all events:

```go
manager.RegisterFunc(core.EventAll, func(data *core.EventData) {
    // This will be called for every event
})
```

### Dynamic Listener Management

Add and remove listeners at runtime:

```go
// Register with custom ID
id := manager.RegisterWithConfig(event, listener, core.ListenerConfig{
    ID: "my-listener",
})

// Later, unregister
manager.Unregister(id)
```

## Use Cases

### 1. Audit Logging

```go
manager.RegisterFunc(core.EventAll, func(data *core.EventData) {
    auditLog.Write(fmt.Sprintf("[%s] %s - %s", 
        data.Event, data.LoginID, time.Unix(data.Timestamp, 0)))
})
```

### 2. Security Monitoring

```go
manager.RegisterFunc(core.EventKickout, func(data *core.EventData) {
    alertSystem.Send(fmt.Sprintf("User %s was kicked out", data.LoginID))
})
```

### 3. Analytics

```go
manager.RegisterFunc(core.EventLogin, func(data *core.EventData) {
    analytics.Track("user_login", map[string]interface{}{
        "user_id": data.LoginID,
        "device":  data.Device,
    })
})
```

### 4. Cache Invalidation

```go
manager.RegisterFunc(core.EventLogout, func(data *core.EventData) {
    cache.Delete("user:" + data.LoginID)
})
```

## Related Documentation

- [Listener Guide](../../docs/guide/listener.md) - Complete listener documentation
- [Authentication Guide](../../docs/guide/authentication.md) - Authentication basics
- [API Reference](../../docs/api/) - API documentation

