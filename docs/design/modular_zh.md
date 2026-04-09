[English](modular.md) | 中文文档

# 模块化设计

## 设计目标

将项目拆分为多个独立模块，实现：

- 按需导入
- 最小依赖
- 多模块独立维护
- 清晰的职责划分
- 方便扩展新的组件和框架集成

## 模块划分

### 核心模块 (core)

```text
github.com/click33/sa-token-go/core
```

**职责**：
- 配置、上下文、Manager、Builder
- 权限、角色、封禁、Session 核心逻辑
- Listener、Nonce、OAuth2 等能力

**特点**：
- 不依赖 Web 框架
- 不依赖具体存储实现
- 通过 `adapter` 接口解耦组件

### 全局工具模块 (stputil)

```text
github.com/click33/sa-token-go/stputil
```

**职责**：
- 对外暴露全局便捷 API
- 管理全局 `Manager`
- 提供统一的 `context.Context` 风格调用方式

### 存储模块

#### Memory 存储

```text
github.com/click33/sa-token-go/com/storage/memory
```

**特点**：
- 零外部依赖
- 适合开发和测试

#### Redis 存储

```text
github.com/click33/sa-token-go/com/storage/redis
```

**特点**：
- 适合生产环境
- 支持分布式部署
- 依赖 `github.com/redis/go-redis/v9`

### 组件模块

当前除了存储外，还拆分了这些可替换组件：

- `com/codec/base64`
- `com/codec/json`
- `com/codec/jsonv2`
- `com/codec/msgpack`
- `com/generator/sgenerator`
- `com/log/gf`
- `com/log/nop`
- `com/log/slog`
- `com/pool/ants`

### 框架集成模块

当前已提供这些独立集成：

- `integrations/gin`
- `integrations/gf`
- `integrations/echo`
- `integrations/fiber`
- `integrations/chi`
- `integrations/hertz`
- `integrations/kratos`

每个集成模块都负责：

- 请求上下文适配
- `SaTokenContext` 注入
- 中间件封装
- 注解式校验封装
- `export.go` 统一导出层

## 依赖关系

```text
应用代码
  ↓
框架集成 (integrations/*)    或    stputil
  ↓
core
  ↓
com/storage/* / com/codec/* / com/log/* / com/pool/* / com/generator/*
```

## 按需导入

### 场景1：只使用核心能力

```bash
go get github.com/click33/sa-token-go/core
go get github.com/click33/sa-token-go/stputil
go get github.com/click33/sa-token-go/com/storage/memory
```

**适合场景**：
- 非 Web 场景
- 自己封装请求上下文
- 希望完全掌控初始化过程

### 场景2：使用框架集成包

```bash
go get github.com/click33/sa-token-go/integrations/gin
go get github.com/click33/sa-token-go/com/storage/redis
```

**适合场景**：
- 直接接入 Gin / Echo / Fiber / Chi / GoFrame / Hertz / Kratos
- 希望使用集成包统一导出的构建器、中间件和工具函数

### 场景3：替换组件实现

```bash
go get github.com/click33/sa-token-go/core
go get github.com/click33/sa-token-go/stputil
go get github.com/click33/sa-token-go/com/storage/redis
go get github.com/click33/sa-token-go/com/log/slog
go get github.com/click33/sa-token-go/com/pool/ants
```

**适合场景**：
- 生产环境
- 需要替换日志、存储、协程池等组件

## 模块独立性

### 每个模块都有独立 go.mod

当前仓库中核心模块、组件模块、集成模块、示例模块都具备独立 `go.mod`。

例如：

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

### 本地开发通过 go.work 统一管理

当前仓库使用 `go.work` 将全部模块串联起来。

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

**优势**：
- 本地开发无缝调试
- 模块之间直接联动
- 无需频繁发布版本做本地联调

## 版本管理

### 版本同步原则

多模块仓库建议保持对外发布版本同步，例如：

- `core`
- `stputil`
- `com/storage/*`
- `integrations/*`

尽量使用相同发布版本，减少跨模块组合时的兼容性问题。

### 兼容性保证

- 核心接口变更要同步更新相关模块
- 集成层导出能力应与 `stputil` / `core` 保持一致
- 文档、示例、导出层需要一起更新

## 扩展新模块

### 添加新存储

1. 创建目录：`com/storage/mysql/`
2. 创建 `go.mod`
3. 实现 `adapter.Storage`
4. 添加到 `go.work`
5. 编写文档和示例

### 添加新框架集成

1. 创建目录：`integrations/iris/`
2. 创建 `go.mod`
3. 实现上下文适配
4. 增加中间件、注解封装、`export.go`
5. 添加到 `go.work`
6. 编写示例和文档

### 添加新组件

1. 创建目录：`com/log/xxx/` 或 `com/codec/xxx/`
2. 实现对应 `adapter` 接口
3. 添加到 `go.work`
4. 编写测试与文档

## 优势总结

| 特性 | 单体模式 | 当前模块化设计 | 优势 |
|------|----------|----------------|------|
| 依赖控制 | 弱 | 强 | 按需导入 |
| 框架隔离 | 弱 | 强 | 集成互不影响 |
| 组件替换 | 一般 | 强 | Storage / Codec / Log / Pool / Generator 可替换 |
| 本地联调 | 一般 | 强 | `go.work` 统一管理 |
| 扩展性 | 一般 | 强 | 易于新增组件与集成 |

## 下一步

- [架构设计](architecture_zh.md)
- [自动续签设计](auto-renew_zh.md)
- [StpUtil API 文档](../api/stputil_zh.md)
