[English](architecture.md) | 中文文档

# 架构设计

## 总体架构

```text
┌──────────────────────────────────────────────┐
│              应用层 (Your App)               │
└──────────────────────┬───────────────────────┘
                       │
        ┌──────────────┴──────────────┐
        │                             │
        ↓                             ↓
┌──────────────────┐         ┌──────────────────┐
│ 框架集成层        │         │ 全局工具入口      │
│ integrations/*   │         │ stputil          │
└────────┬─────────┘         └────────┬─────────┘
         │                             │
         └──────────────┬──────────────┘
                        ↓
┌──────────────────────────────────────────────┐
│                核心层 (core/*)               │
│ Manager / Context / Config / Listener /      │
│ Nonce / OAuth2 / Builder                     │
└──────────────────────┬───────────────────────┘
                       │
        ┌──────────────┴──────────────┐
        │                             │
        ↓                             ↓
┌──────────────────┐         ┌──────────────────┐
│ 组件实现层        │         │ 文档与示例层      │
│ com/*            │         │ docs / examples  │
└──────────────────┘         └──────────────────┘
```

## 模块划分

### 1. 核心层 (core/)

**职责**：提供认证授权的核心能力。

**主要组件**：
- `manager` - 认证管理器核心实现
- `context` - `SaTokenContext` 上下文封装
- `config` - 配置定义与默认值
- `builder` - Builder 构建器
- `listener` - 事件监听系统
- `nonce` - Nonce 防重放能力
- `oauth2` - OAuth2 能力实现
- `serror` - 错误定义

**依赖特点**：
- 不依赖任何 Web 框架
- 不依赖具体存储实现
- 通过 `adapter` 接口抽象 Storage、Codec、Log、Pool、Generator

### 2. 全局工具层 (stputil/)

**职责**：对外提供统一的全局调用入口。

**主要能力**：
- 登录、登出、踢人下线、顶号
- 权限、角色、封禁、Session 操作
- Nonce、OAuth2 等高级能力
- 多认证体系 `authType` 管理

**特点**：
- 以 `context.Context` 为统一入口
- 内部通过全局 `Manager` 映射管理不同认证体系

### 3. 组件实现层 (com/)

**职责**：提供可替换组件实现。

**当前模块**：
- `com/storage/*` - 存储实现
- `com/codec/*` - 编解码实现
- `com/generator/sgenerator` - Token 生成器
- `com/log/*` - 日志实现
- `com/pool/ants` - 续期协程池实现

### 4. 框架集成层 (integrations/)

**职责**：提供 Web 框架适配。

**当前集成**：
- `gin`
- `gf`
- `echo`
- `fiber`
- `chi`
- `hertz`
- `kratos`

**主要功能**：
- 请求上下文适配
- `SaTokenContext` 注入
- 中间件封装
- 注解式校验中间件封装

## 设计模式

### 1. Builder 模式

```go
mgr := builder.NewBuilder().
    SetStorage(memory.NewStorage()).
    TokenName("Authorization").
    Timeout(86400).
    Build()
```

**优势**：
- 链式调用，配置清晰
- 可选参数灵活组合
- 便于后续扩展配置项

### 2. 适配器模式

```go
type Storage interface {
    Set(ctx context.Context, key string, value any, expiration time.Duration) error
    Get(ctx context.Context, key string) (any, error)
    Delete(ctx context.Context, keys ...string) error
    Exists(ctx context.Context, key string) bool
    TTL(ctx context.Context, key string) (time.Duration, error)
}
```

**优势**：
- 解耦具体实现
- 便于替换存储、日志、编解码、协程池
- 所有组件都可以独立扩展

### 3. 中间件 / 注解式封装

```go
annotation.GET("/profile",
    satoken.CheckLoginMiddleware(ctx, handleProfile, handleAuthFail))
```

**优势**：
- 框架集成风格统一
- 登录、权限、角色校验可复用可组合
- 更贴近业务路由写法

### 4. 全局 Manager 注册模式

```go
stputil.SetManager(mgr)
mgr, err := stputil.GetManager()
```

**优势**：
- 支持全局统一调用
- 支持多认证体系并存
- 不需要在业务层层传递 Manager

## 数据流转

### 登录流程

```text
请求
  ↓
stputil.Login(ctx, loginID, ...)
  ↓
1. 解析 device / deviceId / authType
  ↓
2. 生成 Token
  ↓
3. 保存 TokenInfo
  ↓
4. 保存 Session
  ↓
5. 初始化续期 / 活跃状态
  ↓
6. 返回 Token
```

### Token 校验流程

```text
请求
  ↓
stputil.IsLogin(ctx, token)
  ↓
1. 读取 TokenInfo
  ↓
2. 检查账号封禁状态
  ↓
3. 检查 ActiveTimeout
  ↓
4. 满足条件时触发异步续期
  ↓
5. 异步更新活跃时间
  ↓
6. 返回校验结果
```

### 权限校验流程

```text
请求
  ↓
CheckPermissionMiddleware / HasPermission
  ↓
1. 获取 Token 或 loginID
  ↓
2. 检查登录状态
  ↓
3. 获取权限列表
  ↓
4. 执行权限匹配（支持通配符）
  ↓
5. 触发权限校验事件
  ↓
6. 返回结果
```

## 自动续签设计

### 核心思想

当前版本的自动续签不是简单的“每次 `IsLogin()` 都续期”，而是：

- 先检查 Token 当前 TTL
- 只有在接近过期时才触发续期
- 通过 `RenewInterval` 控制最小续期间隔
- 优先提交到异步协程池执行

### 当前实现要点

- `AutoRenew` 控制是否开启自动续签
- `RenewMaxRefresh` 控制何时触发续期
- `RenewInterval` 控制同一 Token 的最小续期间隔
- `ActiveTimeout` 控制最大不活跃时长
- `com/pool/ants` 提供默认续期任务协程池

详见：[自动续签设计](auto-renew_zh.md)

## 数据存储结构

### Storage 键结构

当前 Key 由：

`KeyPrefix + AuthType + 业务前缀 + 业务标识`

组成，默认前缀示例如下：

```text
sa-token-go:auth:{tokenValue}                    -> TokenInfo
sa-token-go:auth:session:{loginID}              -> Session
sa-token-go:auth:renew:{tokenValue}             -> 续期节流标记
sa-token-go:auth:active:{tokenValue}            -> 最近活跃时间
sa-token-go:auth:disable:{loginID}              -> 账号封禁信息
sa-token-go:auth:disable:service:{loginID}:{service} -> 服务封禁信息
```

### TokenInfo 结构

```go
type TokenInfo struct {
    AuthType   string
    LoginID    string
    Device     string
    DeviceId   string
    CreateTime int64
}
```

### Session 结构

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

## 下一步

- [自动续签设计](auto-renew_zh.md)
- [模块化设计](modular_zh.md)
- [StpUtil API 文档](../api/stputil_zh.md)
