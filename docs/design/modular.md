English | [中文文档](modular_zh.md)

# Modular Design

## Design Goals

Split the project into multiple independent modules in order to achieve:

- on-demand imports
- minimal dependencies
- independent multi-module maintenance
- clear separation of responsibilities
- easier extension for new components and framework integrations

## Module Division

### Core Module (core)

```text
github.com/click33/sa-token-go/core
```

**Responsibilities**:
- config, context, manager, builder
- core logic for permissions, roles, disable, session
- listener, nonce, and OAuth2 capability

**Features**:
- no web framework dependency
- no concrete storage dependency
- component abstraction through `adapter`

### Global Utility Module (stputil)

```text
github.com/click33/sa-token-go/stputil
```

**Responsibilities**:
- expose global convenience APIs
- manage the global `Manager`
- provide unified `context.Context`-based calling style

### Storage Modules

#### Memory Storage

```text
github.com/click33/sa-token-go/com/storage/memory
```

**Features**:
- zero external dependencies
- suitable for development and testing

#### Redis Storage

```text
github.com/click33/sa-token-go/com/storage/redis
```

**Features**:
- suitable for production
- supports distributed deployment
- depends on `github.com/redis/go-redis/v9`

### Component Modules

Besides storage, the current repository also splits out these replaceable components:

- `com/codec/base64`
- `com/codec/json`
- `com/codec/jsonv2`
- `com/codec/msgpack`
- `com/generator/sgenerator`
- `com/log/gf`
- `com/log/nop`
- `com/log/slog`
- `com/pool/ants`

### Framework Integration Modules

The current repository provides these independent integrations:

- `integrations/gin`
- `integrations/gf`
- `integrations/echo`
- `integrations/fiber`
- `integrations/chi`
- `integrations/hertz`
- `integrations/kratos`

Each integration module is responsible for:

- request context adaptation
- `SaTokenContext` injection
- middleware wrappers
- annotation-style check wrappers
- unified export layer in `export.go`

## Dependency Relationships

```text
Application Code
  ↓
Framework Integration (integrations/*)    or    stputil
  ↓
core
  ↓
com/storage/* / com/codec/* / com/log/* / com/pool/* / com/generator/*
```

## On-Demand Imports

### Scenario 1: Core Capability Only

```bash
go get github.com/click33/sa-token-go/core
go get github.com/click33/sa-token-go/stputil
go get github.com/click33/sa-token-go/com/storage/memory
```

**Suitable for**:
- non-web usage
- custom request/context integration
- full control over initialization

### Scenario 2: Using a Framework Integration Package

```bash
go get github.com/click33/sa-token-go/integrations/gin
go get github.com/click33/sa-token-go/com/storage/redis
```

**Suitable for**:
- direct integration with Gin / Echo / Fiber / Chi / GoFrame / Hertz / Kratos
- using the integration package as the unified export surface

### Scenario 3: Replacing Components

```bash
go get github.com/click33/sa-token-go/core
go get github.com/click33/sa-token-go/stputil
go get github.com/click33/sa-token-go/com/storage/redis
go get github.com/click33/sa-token-go/com/log/slog
go get github.com/click33/sa-token-go/com/pool/ants
```

**Suitable for**:
- production environments
- replacing logging, storage, pool, or other components

## Module Independence

### Each Module Has Its Own go.mod

Core modules, component modules, integration modules, and example modules all have independent `go.mod` files.

For example:

```text
core/go.mod
stputil/go.mod
com/storage/memory/go.mod
com/storage/redis/go.mod
com/codec/json/go.mod
integrations/gin/go.mod
integrations/gf/go.mod
examples/quick_start/go.mod
```

### Local Development Is Unified by go.work

The current repository uses `go.work` to connect all modules.

```go
go 1.25.0

use (
    ./com/codec/base64
    ./com/codec/json
    ./com/codec/jsonv2
    ./com/codec/msgpack
    ./com/generator/sgenerator
    ./com/log/slog
    ./com/log/gf
    ./com/log/nop
    ./com/pool/ants
    ./com/storage/memory
    ./com/storage/redis
    ./core
    ./stputil
    ./examples/chi
    ./examples/echo
    ./examples/fiber
    ./examples/gf
    ./examples/gin
    ./examples/hertz
    ./examples/kratos
    ./examples/quick_start
    ./integrations/chi
    ./integrations/fiber
    ./integrations/echo
    ./integrations/gf
    ./integrations/gin
    ./integrations/hertz
    ./integrations/kratos
)
```

**Advantages**:
- seamless local debugging
- direct linkage between modules
- no need to publish versions for local integration testing

## Version Management

### Version Synchronization Principle

For a multi-module repository, public versions should stay aligned as much as possible across:

- `core`
- `stputil`
- `com/storage/*`
- `integrations/*`

Using the same released version helps reduce cross-module compatibility issues.

### Compatibility Guarantee

- core interface changes should be propagated to related modules
- integration exports should stay aligned with `stputil` and `core`
- docs, examples, and export layers should be updated together

## Extending New Modules

### Add a New Storage Module

1. Create directory: `com/storage/mysql/`
2. Create `go.mod`
3. Implement `adapter.Storage`
4. Add it to `go.work`
5. Write documentation and examples

### Add a New Framework Integration

1. Create directory: `integrations/iris/`
2. Create `go.mod`
3. Implement context adaptation
4. Add middleware, annotation wrappers, and `export.go`
5. Add it to `go.work`
6. Write examples and documentation

### Add a New Component

1. Create directory: `com/log/xxx/` or `com/codec/xxx/`
2. Implement the corresponding `adapter` interface
3. Add it to `go.work`
4. Write tests and docs

## Advantages Summary

| Feature | Monolithic Style | Current Modular Design | Benefit |
|---------|------------------|------------------------|---------|
| Dependency control | Weak | Strong | On-demand import |
| Framework isolation | Weak | Strong | Integrations do not affect each other |
| Component replacement | Average | Strong | Storage / Codec / Log / Pool / Generator are replaceable |
| Local integration | Average | Strong | Unified by `go.work` |
| Extensibility | Average | Strong | Easier to add components and integrations |

## Next Steps

- [Architecture Design](architecture.md)
- [Auto-Renew Design](auto-renew.md)
- [StpUtil API Documentation](../api/stputil.md)
