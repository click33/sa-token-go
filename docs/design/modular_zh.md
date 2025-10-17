[English](modular.md) | 中文文档

# 模块化设计

## 设计目标

将项目拆分为多个独立模块，实现：
- 按需导入
- 最小依赖
- 独立版本管理
- 清晰的职责划分

## 模块划分

### 核心模块 (core)

```
github.com/click33/sa-token-go/core
```

**职责**：
- Token生成和验证
- Session管理
- 权限和角色验证
- Builder构建器
- StpUtil全局工具类

**依赖**：
- `github.com/golang-jwt/jwt/v5` - JWT支持
- `github.com/google/uuid` - UUID生成

**特点**：
- ✅ 无Web框架依赖
- ✅ 无特定存储依赖
- ✅ 最小依赖树
- ✅ 可独立使用

### 存储模块

#### Memory存储

```
github.com/click33/sa-token-go/storage/memory
```

**依赖**：
- `core` 模块

**特点**：
- ✅ 零外部依赖
- ✅ 高性能
- ✅ 适合开发环境

#### Redis存储

```
github.com/click33/sa-token-go/storage/redis
```

**依赖**：
- `core` 模块
- `github.com/redis/go-redis/v9`

**特点**：
- ✅ 生产环境就绪
- ✅ 分布式支持
- ✅ 数据持久化

### 框架集成模块

#### Gin集成

```
github.com/click33/sa-token-go/integrations/gin
```

**依赖**：
- `core` 模块
- `github.com/gin-gonic/gin`

**提供**：
- 中间件
- 上下文适配器
- 注解装饰器
- 内置处理器

#### Echo/Fiber/Chi集成

类似Gin，每个框架都是独立模块。

## 依赖关系

```
应用代码
  ↓
框架集成 (gin/echo/fiber/chi)
  ↓
核心模块 (core)
  ↓
存储实现 (memory/redis)
```

## 按需导入

### 场景1：只用核心功能

```bash
go get github.com/click33/sa-token-go/core
go get github.com/click33/sa-token-go/storage/memory
```

**依赖树**：
```
core (jwt, uuid)
storage/memory (core)
```

**总计**：~5个依赖包

### 场景2：使用Gin框架

```bash
go get github.com/click33/sa-token-go/core
go get github.com/click33/sa-token-go/storage/redis
go get github.com/click33/sa-token-go/integrations/gin
```

**依赖树**：
```
core (jwt, uuid)
storage/redis (core, go-redis)
integrations/gin (core, gin)
```

**总计**：~15个依赖包

**对比**：如果是单一模块设计，会引入所有框架依赖（~50个包）

## 模块独立性

### 每个模块都有独立的go.mod

```
core/go.mod
storage/memory/go.mod
storage/redis/go.mod
integrations/gin/go.mod
integrations/echo/go.mod
...
```

### replace用于本地开发

```go
// storage/memory/go.mod
require github.com/click33/sa-token-go/core v0.1.0

replace github.com/click33/sa-token-go/core => ../../core
```

**优势**：
- 本地开发无需发布
- 测试更方便
- 调试更容易

## Go Workspace

使用Go Workspace统一管理所有模块：

```go
// go.work
go 1.21

use (
    ./core
    ./storage/memory
    ./storage/redis
    ./integrations/gin
    ./integrations/echo
    ./integrations/fiber
    ./integrations/chi
    ./examples/...
)
```

**优势**：
- 统一管理所有模块
- 本地开发无缝衔接
- 自动解析依赖

## 版本管理

### 版本同步

所有模块保持主版本号同步：

```
core                 v0.1.0
storage/memory       v0.1.0
storage/redis        v0.1.0
integrations/gin     v0.1.0
...
```

### 兼容性保证

- 主版本号相同保证兼容性
- 核心接口变更需同步更新所有模块
- 遵循语义化版本规范

## 扩展新模块

### 添加新存储

1. 创建目录：`storage/mysql/`
2. 创建go.mod：`module github.com/click33/sa-token-go/storage/mysql`
3. 实现Storage接口
4. 添加到go.work
5. 编写文档和示例

### 添加新框架集成

1. 创建目录：`integrations/iris/`
2. 创建go.mod
3. 实现RequestContext适配器
4. 创建中间件和插件
5. 添加到go.work
6. 编写文档和示例

## 优势总结

| 特性 | 单一模块 | 模块化设计 | 优势 |
|------|----------|-----------|------|
| 依赖包数 | ~50个 | ~15个 | ↓ 70% |
| 编译时间 | ~15秒 | ~8秒 | ↑ 46% |
| 项目体积 | ~45MB | ~18MB | ↓ 60% |
| 可维护性 | 低 | 高 | ⭐⭐⭐⭐⭐ |
| 扩展性 | 低 | 高 | ⭐⭐⭐⭐⭐ |

## 下一步

- [架构设计](architecture.md)
- [性能优化](performance.md)
- [API文档](../api/)

