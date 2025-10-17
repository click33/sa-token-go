# 登录认证

[English](authentication.md) | 中文文档

## 基本登录

### 简单登录

```go
// 登录用户（支持多种类型）
token, err := stputil.Login(1000)           // int
token, err := stputil.Login("user123")      // string
token, err := stputil.Login(int64(1000))    // int64
```

### 多设备登录

```go
// 指定设备类型
token, _ := stputil.Login(1000, "web")
token, _ := stputil.Login(1000, "mobile")
token, _ := stputil.Login(1000, "app")
```

## 检查登录状态

```go
// 检查是否登录
isLogin := stputil.IsLogin(token)

// 检查登录（未登录抛出错误）
err := stputil.CheckLogin(token)
```

## 获取登录信息

```go
// 获取登录ID
loginID, err := stputil.GetLoginID(token)

// 获取Token信息
info, err := stputil.GetTokenInfo(token)
fmt.Printf("登录ID: %s\n", info.LoginID)
fmt.Printf("设备: %s\n", info.Device)
fmt.Printf("创建时间: %d\n", info.CreateTime)
fmt.Printf("活跃时间: %d\n", info.ActiveTime)
```

## 登出

```go
// 根据登录ID登出
stputil.Logout(1000)
stputil.Logout(1000, "mobile")  // 指定设备

// 根据Token登出
stputil.LogoutByToken(token)
```

## 踢人下线

```go
// 踢掉指定账号
stputil.Kickout(1000)
stputil.Kickout(1000, "mobile")  // 踢掉指定设备

// 配置互斥登录
core.NewBuilder().
    IsConcurrent(false).  // 不允许并发登录
    Build()
// 新登录会自动踢掉旧登录
```

## 自动续签

### 工作原理

每次调用 `IsLogin()` 时，如果开启了自动续签，会**异步**延长Token过期时间。

### 配置

```go
core.NewBuilder().
    Timeout(86400).      // 24小时超时
    AutoRenew(true).     // 开启自动续签（异步）
    Build()
```

### 效果

```go
// 用户持续活跃，Token永不过期
for {
    stputil.IsLogin(token)  // 每次检查都会异步续签
    // 用户继续使用...
}
```

## 活跃检测

### 配置活跃超时

```go
core.NewBuilder().
    ActiveTimeout(1800).  // 30分钟无操作强制登出
    Build()
```

### 工作流程

1. 用户登录，记录活跃时间
2. 每次`IsLogin()`检查时，对比当前时间和上次活跃时间
3. 如果超过`ActiveTimeout`，强制登出
4. 否则，更新活跃时间并继续

## 完整配置示例

```go
stputil.SetManager(
    core.NewBuilder().
        Storage(memory.NewStorage()).
        TokenName("Authorization").
        Timeout(86400).                // 24小时绝对超时
        ActiveTimeout(1800).           // 30分钟活跃超时
        IsConcurrent(false).           // 不允许并发登录
        IsShare(false).                // 不共享Token
        TokenStyle(core.TokenStyleRandom64).
        AutoRenew(true).               // 异步自动续签
        Build(),
)
```

## 下一步

- [权限验证](permission.md)
- [角色管理](role.md)
- [Session管理](session.md)

