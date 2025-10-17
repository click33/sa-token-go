English | [中文文档](modular_zh.md)

# Modular Design

## Design Goals

Split the project into multiple independent modules to achieve:
- On-demand imports
- Minimal dependencies
- Independent version management
- Clear responsibility separation

## Module Division

### Core Module (core)

```
github.com/click33/sa-token-go/core
```

**Responsibilities**:
- Token generation and validation
- Session management
- Permission and role verification
- Builder pattern
- StpUtil global utility class

**Dependencies**:
- `github.com/golang-jwt/jwt/v5` - JWT support
- `github.com/google/uuid` - UUID generation

**Features**:
- ✅ No web framework dependencies
- ✅ No specific storage dependencies
- ✅ Minimal dependency tree
- ✅ Can be used independently

### StpUtil Module (stputil)

```
github.com/click33/sa-token-go/stputil
```

**Responsibilities**:
- Global utility class
- Convenient access to core functions
- Direct method calls like `stputil.Login(1000)`

**Dependencies**:
- `core` module

**Features**:
- ✅ Top-level module for easy import
- ✅ Idiomatic Go usage
- ✅ No circular dependencies

### Storage Modules

#### Memory Storage

```
github.com/click33/sa-token-go/storage/memory
```

**Dependencies**:
- `core` module

**Features**:
- ✅ Zero external dependencies
- ✅ High performance
- ✅ Suitable for development environment

#### Redis Storage

```
github.com/click33/sa-token-go/storage/redis
```

**Dependencies**:
- `core` module
- `github.com/redis/go-redis/v9`

**Features**:
- ✅ Production-ready
- ✅ Distributed support
- ✅ Data persistence

### Framework Integration Modules

#### Gin Integration

```
github.com/click33/sa-token-go/integrations/gin
```

**Dependencies**:
- `core` module
- `stputil` module
- `github.com/gin-gonic/gin`

**Provides**:
- Middleware
- Context adapter
- Annotation decorators
- Built-in handlers

#### Echo/Fiber/Chi Integration

Similar to Gin, each framework is an independent module.

## Dependency Relationships

```
Application Code
  ↓
Framework Integration (gin/echo/fiber/chi)
  ↓
StpUtil Module (stputil)
  ↓
Core Module (core)
  ↓
Storage Implementation (memory/redis)
```

## On-Demand Imports

### Scenario 1: Core Functionality Only

```bash
go get github.com/click33/sa-token-go/core
go get github.com/click33/sa-token-go/stputil
go get github.com/click33/sa-token-go/storage/memory
```

**Dependency Tree**:
```
core (jwt, uuid)
stputil (core)
storage/memory (core)
```

**Total**: ~5 dependency packages

### Scenario 2: Using Gin Framework

```bash
go get github.com/click33/sa-token-go/core
go get github.com/click33/sa-token-go/stputil
go get github.com/click33/sa-token-go/storage/redis
go get github.com/click33/sa-token-go/integrations/gin
```

**Dependency Tree**:
```
core (jwt, uuid)
stputil (core)
storage/redis (core, go-redis)
integrations/gin (core, stputil, gin)
```

**Total**: ~18 dependency packages

**Comparison**: A monolithic design would pull in all framework dependencies (~50 packages)

## Module Independence

### Each Module Has Its Own go.mod

```
core/go.mod
stputil/go.mod
storage/memory/go.mod
storage/redis/go.mod
integrations/gin/go.mod
integrations/echo/go.mod
...
```

### Replace for Local Development

```go
// storage/memory/go.mod
require github.com/click33/sa-token-go/core v0.1.0

replace github.com/click33/sa-token-go/core => ../../core
```

**Advantages**:
- No need to publish for local development
- Easier testing
- Easier debugging

## Go Workspace

Use Go Workspace to manage all modules:

```go
// go.work
go 1.21

use (
    ./core
    ./stputil
    ./storage/memory
    ./storage/redis
    ./integrations/gin
    ./integrations/echo
    ./integrations/fiber
    ./integrations/chi
    ./examples/...
)
```

**Advantages**:
- Unified management of all modules
- Seamless local development
- Automatic dependency resolution

## Version Management

### Version Synchronization

All modules maintain synchronized major version numbers:

```
core                 v0.1.0
stputil              v0.1.0
storage/memory       v0.1.0
storage/redis        v0.1.0
integrations/gin     v0.1.0
...
```

### Compatibility Guarantee

- Same major version ensures compatibility
- Core interface changes require synchronous updates to all modules
- Follow semantic versioning

## Extending New Modules

### Adding New Storage

1. Create directory: `storage/mysql/`
2. Create go.mod: `module github.com/click33/sa-token-go/storage/mysql`
3. Implement Storage interface
4. Add to go.work
5. Write documentation and examples

### Adding New Framework Integration

1. Create directory: `integrations/iris/`
2. Create go.mod
3. Implement RequestContext adapter
4. Create middleware and plugin
5. Add to go.work
6. Write documentation and examples

## Advantages Summary

| Feature | Monolithic | Modular Design | Benefit |
|---------|-----------|---------------|---------|
| Dependency count | ~50 | ~18 | ↓ 64% |
| Compile time | ~15s | ~8s | ↑ 46% |
| Project size | ~45MB | ~18MB | ↓ 60% |
| Maintainability | Low | High | ⭐⭐⭐⭐⭐ |
| Extensibility | Low | High | ⭐⭐⭐⭐⭐ |

## Next Steps

- [Architecture Design](architecture.md)
- [Performance Optimization](performance.md)
- [API Documentation](../api/)
