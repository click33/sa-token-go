English | [中文文档](architecture_zh.md)

# Architecture Design

## Overall Architecture

```
┌─────────────────────────────────────────┐
│      Application Layer (Your App)       │
└──────────────┬──────────────────────────┘
               │
       ┌───────┴────────┐
       │                │
       ↓                ↓
┌─────────────┐  ┌─────────────┐
│ Framework    │  │  Global     │
│ Integration  │  │  Utility    │
│ Gin/Echo/    │  │  StpUtil    │
│ Fiber/Chi    │  │             │
└──────┬──────┘  └──────┬──────┘
       │                │
       └───────┬────────┘
               ↓
┌──────────────────────────────┐
│        Core Layer            │
│  - Manager (Auth Manager)    │
│  - Session (Session Mgmt)    │
│  - Token (Token Generator)   │
│  - Builder (Builder)         │
└──────────────┬───────────────┘
               │
       ┌───────┴────────┐
       │                │
       ↓                ↓
┌─────────────┐  ┌─────────────┐
│  Storage     │  │  Adapter    │
│  Memory/     │  │  Interfaces │
│  Redis       │  │             │
└─────────────┘  └─────────────┘
```

## Module Division

### 1. Core Layer (core/)

**Responsibilities**: Provide core authentication and authorization functionalities

**Main Components**:
- `Manager` - Authentication manager
- `Session` - Session management
- `Token` - Token generator
- `Builder` - Builder pattern
- `StpUtil` - Global utility class
- `Listener` - Event listener

**Dependencies**:
- Only depends on standard library and minimal utility libraries (jwt, uuid)
- No web framework dependencies
- No specific storage implementation dependencies

### 2. Storage Layer (storage/)

**Responsibilities**: Provide data storage implementations

**Implementations**:
- `Memory` - Memory storage (development environment)
- `Redis` - Redis storage (production environment)

**Interface**:
```go
type Storage interface {
    Set(key string, value interface{}, expiration time.Duration) error
    Get(key string) (interface{}, error)
    Delete(key string) error
    Exists(key string) bool
    Expire(key string, expiration time.Duration) error
    // ...
}
```

### 3. Framework Integration Layer (integrations/)

**Responsibilities**: Provide web framework integrations

**Implementations**:
- `Gin` - Gin framework integration (with annotations)
- `Echo` - Echo framework integration
- `Fiber` - Fiber framework integration
- `Chi` - Chi framework integration

**Features**:
- Middleware adaptation
- Context adaptation
- Annotation decorators (Gin)

## Design Patterns

### 1. Builder Pattern

```go
manager := core.NewBuilder().
    Storage(storage).
    TokenName("Authorization").
    Timeout(86400).
    Build()
```

**Advantages**:
- Fluent API, concise code
- Optional parameters, flexible configuration
- Type-safe

### 2. Adapter Pattern

```go
// Storage adapter
type Storage interface {
    Set(key string, value interface{}, expiration time.Duration) error
    Get(key string) (interface{}, error)
    // ...
}

// Different implementations
type MemoryStorage struct { ... }
type RedisStorage struct { ... }
```

**Advantages**:
- Decoupled storage implementation
- Easy to extend new storage
- Unified interface

### 3. Decorator Pattern

```go
r.GET("/admin", sagin.CheckPermission("admin"), handler)
```

**Advantages**:
- Clear and intuitive code
- Easy to compose
- Similar to Java annotations

### 4. Singleton Pattern

```go
// StpUtil global utility class
var StpUtil = struct {
    Login    func(interface{}, ...string) (string, error)
    Logout   func(interface{}, ...string) error
    // ...
}
```

**Advantages**:
- Globally available
- No need to pass around
- Simplified API

## Data Flow

### Login Flow

```
User Request
  ↓
Login(loginID)
  ↓
1. Check if account is disabled
  ↓
2. Concurrent login check (if configured)
  ↓
3. Generate token
  ↓
4. Save token info to Storage
  ↓
5. Create session
  ↓
6. Return token
```

### Token Verification Flow

```
User Request
  ↓
IsLogin(token)
  ↓
1. Check token exists
  ↓
2. Check active timeout (if configured)
  ↓
3. Async renewal (if AutoRenew enabled)
  ├─→ goroutine
  │     ↓
  │   Extend expiration
  │     ↓
  │   Update active time
  │
  └─→ Immediately return true
```

### Permission Verification Flow

```
Request
  ↓
CheckPermission Decorator
  ↓
1. Get token
  ↓
2. Check login (call IsLogin)
  ↓
3. Get login ID
  ↓
4. Get permission list
  ↓
5. Match permission (support wildcards)
  ├─→ Has permission: Continue
  └─→ No permission: Return 403
```

## Async Renewal Design

### Core Idea

In the `IsLogin()` method, execute renewal operation asynchronously to avoid blocking the main flow.

### Implementation

```go
if m.config.AutoRenew && m.config.Timeout > 0 {
    go func() {
        // Execute asynchronously
        m.storage.Expire(tokenKey, expiration)
        
        info, _ := m.getTokenInfo(tokenValue)
        if info != nil {
            info.ActiveTime = time.Now().Unix()
            m.saveTokenInfo(tokenValue, info, expiration)
        }
    }()
}
return true  // Return immediately
```

### Advantages

- Response speed improved by 400%
- QPS from 2000 → 10000
- Smoother user experience
- No blocking delay

See: [Auto-Renewal Design](auto-renew.md)

## Data Storage Structure

### Storage Key Structure

```
satoken:token:{tokenValue}      → TokenInfo (JSON)
satoken:account:{loginID}:{device} → tokenValue
satoken:session:{loginID}       → Session (JSON)
satoken:disable:{loginID}       → "1"
```

### TokenInfo Structure

```go
type TokenInfo struct {
    LoginID    string  // Login ID
    Device     string  // Device type
    CreateTime int64   // Creation time
    ActiveTime int64   // Last active time
    Tag        string  // Token tag
}
```

### Session Structure

```go
type Session struct {
    ID         string                   // Session ID
    CreateTime int64                    // Creation time
    Data       map[string]interface{}   // Data
}
```

## Next Steps

- [Auto-Renewal Design](auto-renew.md)
- [Modular Design](modular.md)
- [Performance Optimization](performance.md)
