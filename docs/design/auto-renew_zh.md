[English](auto-renew.md) | 中文文档

# 自动续签设计

## 设计目标

自动续签的目标是：

- 让活跃用户尽量不因 Token 到期而频繁重新登录
- 避免每次登录校验都做重型同步操作
- 通过阈值和节流机制控制续签频率
- 通过协程池提升高并发下的稳定性

## 核心设计

### 异步续签策略

当前版本的自动续签发生在登录校验流程内部，核心入口是 `Manager.checkLoginInternal()`。

与旧版“每次 `IsLogin()` 直接裸 `goroutine` 续期”不同，当前实现会：

1. 先读取 Token 的剩余 TTL
2. 只有剩余 TTL 小于等于 `RenewMaxRefresh` 时才考虑续期
3. 如果配置了 `RenewInterval`，则通过续期标记键限制频繁续期
4. 优先通过协程池提交续期任务
5. 同时异步更新活跃时间键

### 实现思路

```go
if m.config.AutoRenew && m.config.Timeout > 0 {
    if ttl <= RenewMaxRefresh && renewInterval 条件满足 {
        renewFunc := func() {
            m.renewFunc(context.Background(), tokenValue, tokenInfo.LoginID)
        }

        if m.pool != nil {
            _ = m.pool.Submit(renewFunc)
        } else {
            go renewFunc()
        }
    }
}

if m.config.ActiveTimeout > 0 {
    activeFunc := func() {
        _ = m.storage.Set(ctx, m.getActiveKey(tokenValue), time.Now().Unix(), m.getExpiration())
    }
    // 同样优先走协程池
}
```

## 工作流程

### 同步部分

```text
1. 读取 TokenInfo
   ├─ 失败 -> 返回未登录或 Token 状态错误
   └─ 成功 -> 继续

2. 检查账号封禁状态
   ├─ 已封禁 -> 返回封禁错误
   └─ 未封禁 -> 继续

3. 检查 ActiveTimeout
   ├─ 超时 -> 执行踢出并返回错误
   └─ 未超时 -> 继续

4. 返回登录校验成功
```

### 异步部分

```text
异步续期任务
  ↓
1. 延长 Token 过期时间
  ↓
2. 延长 Session 过期时间
  ↓
3. 写入续期间隔标记（如果启用）
  ↓
4. 触发续期事件

异步活跃任务
  ↓
1. 更新 active:{token} 键
```

## 续签触发条件

当前实现中，自动续签通常需要同时满足这些条件：

- `AutoRenew = true`
- `Timeout > 0`
- Token 当前 TTL 大于 0
- `TTL <= RenewMaxRefresh`，或者 `RenewMaxRefresh <= 0`
- 未命中 `RenewInterval` 的节流限制

这意味着：

- 不是每次请求都会续期
- 越接近过期，越有可能触发续期
- 可以避免高频请求不断刷新同一 Token

## 触发时机

凡是内部走登录校验流程的场景，都可能触发自动续签：

### 1. 中间件认证

```go
r.Use(satoken.AuthMiddleware(ctx))
```

### 2. 注解式登录校验

```go
annotation.GET("/profile",
    satoken.CheckLoginMiddleware(ctx, handleProfile, handleAuthFail))
```

### 3. 手动检查登录

```go
stputil.IsLogin(ctx, token)
stputil.CheckLogin(ctx, token)
```

### 4. 获取登录信息

```go
stputil.GetLoginID(ctx, token)
stputil.GetTokenInfo(ctx, token)
```

## 配置选项

### 启用自动续签

```go
builder.NewBuilder().
    SetStorage(memory.NewStorage()).
    Timeout(86400).
    AutoRenew(true).
    Build()
```

### 指定续签触发阈值

```go
builder.NewBuilder().
    SetStorage(memory.NewStorage()).
    Timeout(86400).
    AutoRenew(true).
    RenewMaxRefresh(3600). // 剩余 1 小时内才触发续签
    Build()
```

### 指定最小续期间隔

```go
builder.NewBuilder().
    SetStorage(memory.NewStorage()).
    Timeout(86400).
    AutoRenew(true).
    RenewMaxRefresh(3600).
    RenewInterval(300). // 同一 Token 至少 5 分钟才续一次
    Build()
```

### 结合最大不活跃时长

```go
builder.NewBuilder().
    SetStorage(memory.NewStorage()).
    Timeout(86400).
    ActiveTimeout(1800).
    AutoRenew(true).
    RenewMaxRefresh(3600).
    RenewInterval(300).
    Build()
```

**效果**：
- 活跃用户在接近过期时自动续签
- 超过 `ActiveTimeout` 未活跃会被踢出
- 续签不会因高频请求无限制执行

## 并发安全

### Storage 接口是上下文化的

```go
type Storage interface {
    Set(ctx context.Context, key string, value any, expiration time.Duration) error
    Get(ctx context.Context, key string) (any, error)
    Delete(ctx context.Context, keys ...string) error
    Exists(ctx context.Context, key string) bool
    TTL(ctx context.Context, key string) (time.Duration, error)
}
```

### 协程池支持

默认情况下，开启自动续签时会优先使用续期协程池：

- `com/pool/ants`
- `adapter.Pool`
- `Builder.SetPool(...)`

如果未显式传入池，内部也会在合适场景创建默认续期池。

## 续签失败处理

### 处理策略

异步续签失败不会直接阻塞当前请求：

1. 当前登录校验结果已返回
2. 本次续期失败只影响后续有效期延长
3. 下一次满足条件时仍可重试续期

### 影响范围

- 不会直接让当前请求失败
- 可能导致 Token 没有被成功延长
- 若连续失败，Token 最终会按原始 TTL 到期

## 最佳实践

### 生产环境建议

```go
builder.NewBuilder().
    SetStorage(redisStorage).
    Timeout(86400).
    ActiveTimeout(1800).
    AutoRenew(true).
    RenewMaxRefresh(3600).
    RenewInterval(300).
    Build()
```

### 开发环境建议

```go
builder.NewBuilder().
    SetStorage(memory.NewStorage()).
    Timeout(7200).
    AutoRenew(true).
    Build()
```

### 安全优先配置

```go
builder.NewBuilder().
    SetStorage(redisStorage).
    Timeout(1800).
    AutoRenew(false).
    Build()
```

## 下一步

- [架构设计](architecture_zh.md)
- [模块化设计](modular_zh.md)
- [StpUtil API 文档](../api/stputil_zh.md)
