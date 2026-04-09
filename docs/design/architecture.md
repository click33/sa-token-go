English | [中文文档](architecture_zh.md)

# Architecture Design

## Overall Architecture

```text
┌──────────────────────────────────────────────┐
│              Application Layer               │
└──────────────────────┬───────────────────────┘
                       │
        ┌──────────────┴──────────────┐
        │                             │
        ↓                             ↓
┌──────────────────┐         ┌──────────────────┐
│ Framework Layer   │         │ Global Utility   │
│ integrations/*    │         │ stputil          │
└────────┬─────────┘         └────────┬─────────┘
         │                             │
         └──────────────┬──────────────┘
                        ↓
┌──────────────────────────────────────────────┐
│                 Core Layer (core/*)          │
│ Manager / Context / Config / Listener /      │
│ Nonce / OAuth2 / Builder                     │
└──────────────────────┬───────────────────────┘
                       │
        ┌──────────────┴──────────────┐
        │                             │
        ↓                             ↓
┌──────────────────┐         ┌──────────────────┐
│ Component Layer   │         │ Docs & Examples  │
│ com/*             │         │ docs / examples  │
└──────────────────┘         └──────────────────┘
```

## Module Division

### 1. Core Layer (core/)

**Responsibilities**: provide core authentication and authorization capability.

**Main components**:
- `manager` - core authentication manager implementation
- `context` - `SaTokenContext` wrapper
- `config` - config definitions and defaults
- `builder` - builder
- `listener` - event listener system
- `nonce` - nonce anti-replay capability
- `oauth2` - OAuth2 implementation
- `serror` - error definitions

**Dependency characteristics**:
- no dependency on any web framework
- no dependency on any concrete storage implementation
- Storage, Codec, Log, Pool, and Generator are abstracted through `adapter`

### 2. Global Utility Layer (stputil/)

**Responsibilities**: expose a unified global entry for application code.

**Main capabilities**:
- login, logout, kickout, replace
- permissions, roles, disable, session operations
- nonce and OAuth2 support
- multi-auth-system support via `authType`

**Features**:
- uses `context.Context` as the unified entry
- manages different auth systems through the global manager registry

### 3. Component Layer (com/)

**Responsibilities**: provide replaceable component implementations.

**Current modules**:
- `com/storage/*` - storage implementations
- `com/codec/*` - codec implementations
- `com/generator/sgenerator` - token generator
- `com/log/*` - log implementations
- `com/pool/ants` - renew task pool implementation

### 4. Framework Integration Layer (integrations/)

**Responsibilities**: provide web framework adapters.

**Current integrations**:
- `gin`
- `gf`
- `echo`
- `fiber`
- `chi`
- `hertz`
- `kratos`

**Main features**:
- request context adaptation
- `SaTokenContext` injection
- middleware wrappers
- annotation-style checking middleware wrappers

## Design Patterns

### 1. Builder Pattern

```go
mgr := builder.NewBuilder().
    SetStorage(memory.NewStorage()).
    TokenName("Authorization").
    Timeout(86400).
    Build()
```

**Advantages**:
- fluent and readable configuration
- flexible optional parameters
- easy to extend with new config items

### 2. Adapter Pattern

```go
type Storage interface {
    Set(ctx context.Context, key string, value any, expiration time.Duration) error
    Get(ctx context.Context, key string) (any, error)
    Delete(ctx context.Context, keys ...string) error
    Exists(ctx context.Context, key string) bool
    TTL(ctx context.Context, key string) (time.Duration, error)
}
```

**Advantages**:
- decouples concrete implementations
- makes storage, logging, codec, pool, and generator replaceable
- keeps the core layer implementation-agnostic

### 3. Middleware / Annotation-Style Composition

```go
annotation.GET("/profile",
    satoken.CheckLoginMiddleware(ctx, handleProfile, handleAuthFail))
```

**Advantages**:
- unified integration style across frameworks
- reusable login, permission, and role checks
- closer to real business routing style

### 4. Global Manager Registry Pattern

```go
stputil.SetManager(mgr)
mgr, err := stputil.GetManager()
```

**Advantages**:
- supports global access
- supports multiple auth systems
- avoids passing manager objects throughout business code

## Data Flow

### Login Flow

```text
Request
  ↓
stputil.Login(ctx, loginID, ...)
  ↓
1. Parse device / deviceId / authType
  ↓
2. Generate token
  ↓
3. Save TokenInfo
  ↓
4. Save Session
  ↓
5. Initialize renew / active state
  ↓
6. Return token
```

### Token Validation Flow

```text
Request
  ↓
stputil.IsLogin(ctx, token)
  ↓
1. Load TokenInfo
  ↓
2. Check account disable state
  ↓
3. Check ActiveTimeout
  ↓
4. Trigger async renew if conditions match
  ↓
5. Update active timestamp asynchronously
  ↓
6. Return validation result
```

### Permission Validation Flow

```text
Request
  ↓
CheckPermissionMiddleware / HasPermission
  ↓
1. Get token or loginID
  ↓
2. Check login state
  ↓
3. Load permission list
  ↓
4. Match permissions (with wildcard support)
  ↓
5. Trigger permission-check event
  ↓
6. Return result
```

## Auto-Renew Design

### Core Idea

The current version does not simply renew on every `IsLogin()` call. Instead it:

- checks the current token TTL first
- only renews when the token is close enough to expiry
- throttles renew operations through `RenewInterval`
- prefers submitting renew work to an async pool

### Current Implementation Highlights

- `AutoRenew` controls whether auto-renew is enabled
- `RenewMaxRefresh` controls when renew should trigger
- `RenewInterval` controls the minimum renew interval for the same token
- `ActiveTimeout` controls the maximum inactive duration
- `com/pool/ants` provides the default renew task pool

See: [Auto-Renew Design](auto-renew.md)

## Data Storage Structure

### Storage Key Structure

Current keys are composed as:

`KeyPrefix + AuthType + businessPrefix + businessID`

With default values, common key examples are:

```text
sa-token-go:auth:{tokenValue}                         -> TokenInfo
sa-token-go:auth:session:{loginID}                   -> Session
sa-token-go:auth:renew:{tokenValue}                  -> renew throttle marker
sa-token-go:auth:active:{tokenValue}                 -> last active timestamp
sa-token-go:auth:disable:{loginID}                   -> account disable info
sa-token-go:auth:disable:service:{loginID}:{service} -> service disable info
```

### TokenInfo Structure

```go
type TokenInfo struct {
    AuthType   string
    LoginID    string
    Device     string
    DeviceId   string
    CreateTime int64
}
```

### Session Structure

```go
type Session struct {
    AuthType             string
    LoginID              string
    CreateTime           int64
    TerminalInfos        []TerminalInfo
    Permissions          []string
    Roles                []string
    HistoryTerminalCount int64
}
```

## Next Steps

- [Auto-Renew Design](auto-renew.md)
- [Modular Design](modular.md)
- [StpUtil API Documentation](../api/stputil.md)
