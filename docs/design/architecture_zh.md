[English](architecture.md) | 中文文档

# 架构设计

## 总体架构

```
┌─────────────────────────────────────────┐
│         应用层 (Your Application)        │
└──────────────┬──────────────────────────┘
               │
       ┌───────┴────────┐
       │                │
       ↓                ↓
┌─────────────┐  ┌─────────────┐
│ 框架集成层   │  │  全局工具类  │
│ Gin/Echo/   │  │  StpUtil    │
│ Fiber/Chi   │  │             │
└──────┬──────┘  └──────┬──────┘
       │                │
       └───────┬────────┘
               ↓
┌──────────────────────────────┐
│        核心层 (Core)          │
│  - Manager (认证管理器)       │
│  - Session (会话管理)         │
│  - Token (Token生成)          │
│  - Builder (构建器)           │
└──────────────┬───────────────┘
               │
       ┌───────┴────────┐
       │                │
       ↓                ↓
┌─────────────┐  ┌─────────────┐
│  存储层      │  │  适配器层    │
│  Memory/    │  │  接口定义    │
│  Redis      │  │             │
└─────────────┘  └─────────────┘
```

## 模块划分

### 1. 核心层 (core/)

**职责**：提供认证授权的核心功能

**主要组件**：
- `Manager` - 认证管理器
- `Session` - 会话管理
- `Token` - Token生成器
- `Builder` - 构建器
- `StpUtil` - 全局工具类
- `Listener` - 事件监听器

**依赖**：
- 仅依赖标准库和少量工具库（jwt, uuid）
- 不依赖任何Web框架
- 不依赖具体存储实现

### 2. 存储层 (storage/)

**职责**：提供数据存储实现

**实现**：
- `Memory` - 内存存储（开发环境）
- `Redis` - Redis存储（生产环境）

**接口**：
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

### 3. 框架集成层 (integrations/)

**职责**：提供Web框架集成

**实现**：
- `Gin` - Gin框架集成（包含注解）
- `Echo` - Echo框架集成
- `Fiber` - Fiber框架集成
- `Chi` - Chi框架集成

**功能**：
- 中间件适配
- 上下文适配
- 注解装饰器（Gin）

## 设计模式

### 1. Builder模式

```go
manager := core.NewBuilder().
    Storage(storage).
    TokenName("Authorization").
    Timeout(86400).
    Build()
```

**优势**：
- 链式调用，代码简洁
- 参数可选，灵活配置
- 类型安全

### 2. 适配器模式

```go
// Storage适配器
type Storage interface {
    Set(key string, value interface{}, expiration time.Duration) error
    Get(key string) (interface{}, error)
    // ...
}

// 不同实现
type MemoryStorage struct { ... }
type RedisStorage struct { ... }
```

**优势**：
- 解耦存储实现
- 易于扩展新存储
- 统一接口

### 3. 装饰器模式

```go
r.GET("/admin", sagin.CheckPermission("admin"), handler)
```

**优势**：
- 代码清晰直观
- 易于组合
- 类似Java注解

### 4. 单例模式

```go
// StpUtil全局工具类
var StpUtil = struct {
    Login    func(interface{}, ...string) (string, error)
    Logout   func(interface{}, ...string) error
    // ...
}
```

**优势**：
- 全局可用
- 无需传递
- 简化API

## 数据流转

### 登录流程

```
用户请求
  ↓
Login(loginID)
  ↓
1. 检查账号是否被封禁
  ↓
2. 互斥登录检查（如果配置）
  ↓
3. 生成Token
  ↓
4. 保存Token信息到Storage
  ↓
5. 创建Session
  ↓
6. 返回Token
```

### Token验证流程

```
用户请求
  ↓
IsLogin(token)
  ↓
1. 检查Token存在
  ↓
2. 检查活跃超时（如果配置）
  ↓
3. 异步续签（如果开启AutoRenew）
  ├─→ goroutine
  │     ↓
  │   延长过期时间
  │     ↓
  │   更新活跃时间
  │
  └─→ 立即返回true
```

### 权限验证流程

```
请求
  ↓
CheckPermission装饰器
  ↓
1. 获取Token
  ↓
2. 检查登录（调用IsLogin）
  ↓
3. 获取登录ID
  ↓
4. 获取权限列表
  ↓
5. 匹配权限（支持通配符）
  ├─→ 有权限：继续
  └─→ 无权限：返回403
```

## 异步续签设计

### 核心思想

在`IsLogin()`方法中，将续签操作异步执行，避免阻塞主流程。

### 实现

```go
if m.config.AutoRenew && m.config.Timeout > 0 {
    go func() {
        // 异步执行
        m.storage.Expire(tokenKey, expiration)
        
        info, _ := m.getTokenInfo(tokenValue)
        if info != nil {
            info.ActiveTime = time.Now().Unix()
            m.saveTokenInfo(tokenValue, info, expiration)
        }
    }()
}
return true  // 立即返回
```

### 优势

- 响应速度提升 400%
- QPS从 2000 → 10000
- 用户体验更流畅
- 无阻塞延迟

详见：[自动续签设计](auto-renew.md)

## 数据存储结构

### Storage键结构

```
satoken:token:{tokenValue}      → TokenInfo (JSON)
satoken:account:{loginID}:{device} → tokenValue
satoken:session:{loginID}       → Session (JSON)
satoken:disable:{loginID}       → "1"
```

### TokenInfo结构

```go
type TokenInfo struct {
    LoginID    string  // 登录ID
    Device     string  // 设备类型
    CreateTime int64   // 创建时间
    ActiveTime int64   // 最后活跃时间
    Tag        string  // Token标签
}
```

### Session结构

```go
type Session struct {
    ID         string                   // Session ID
    CreateTime int64                    // 创建时间
    Data       map[string]interface{}   // 数据
}
```

## 下一步

- [自动续签设计](auto-renew.md)
- [模块化设计](modular.md)
- [性能优化](performance.md)

