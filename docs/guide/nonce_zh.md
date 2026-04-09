[English](nonce.md) | 中文文档

# Nonce 防重放

## 概览

当前项目已经内置 `Nonce` 管理能力，用来防止请求重放。

公开入口包括：

- `GenerateNonce`
- `GenerateNonceWithTimeout`
- `VerifyNonce`
- `VerifyAndConsumeNonce`
- `IsNonceValid`
- `GetNonceTTL`

## 工作机制

当前实现里的 Nonce 规则很清晰：

1. 生成一个随机 nonce
2. 存入存储层并带 TTL
3. 校验时通过 `GetAndDelete` 原子消费
4. 同一个 nonce 只能成功一次

## 默认行为

根据 `core/nonce` 当前实现：

- nonce 原始长度为 `32` 字节随机数
- 输出后是 `64` 位十六进制字符串
- 默认 TTL 为 `5` 分钟

## 基本使用

```go
package main

import (
    "context"
    "fmt"

    "github.com/click33/sa-token-go/com/storage/memory"
    "github.com/click33/sa-token-go/core/builder"
    "github.com/click33/sa-token-go/stputil"
)

func initSaToken() {
    stputil.SetManager(
        builder.NewBuilder().
            SetStorage(memory.NewStorage()).
            Build(),
    )
}

func main() {
    ctx := context.Background()

    nonce, _ := stputil.GenerateNonce(ctx)
    fmt.Println(nonce)

    ok := stputil.VerifyNonce(ctx, nonce)
    fmt.Println(ok) // true

    ok = stputil.VerifyNonce(ctx, nonce)
    fmt.Println(ok) // false
}
```

## 自定义有效期

```go
ctx := context.Background()

nonce, err := stputil.GenerateNonceWithTimeout(ctx, 30*time.Second)
_ = nonce
_ = err
```

如果传入的超时时间小于等于 `0`，底层会退回默认 TTL。

## 非消费式校验

```go
ctx := context.Background()

nonce, _ := stputil.GenerateNonce(ctx)

valid := stputil.IsNonceValid(ctx, nonce) // 只校验，不消费
err := stputil.VerifyAndConsumeNonce(ctx, nonce)
```

区别是：

- `IsNonceValid`：只检查
- `VerifyNonce`：检查并消费，返回 `bool`
- `VerifyAndConsumeNonce`：检查并消费，失败时返回 `ErrInvalidNonce`

## 查看 TTL

```go
ctx := context.Background()

ttl, err := stputil.GetNonceTTL(ctx, nonce)
```

返回值约定：

- `-2`：nonce 不存在
- `-1`：永久有效
- `>=0`：剩余秒数

## HTTP 场景示例

```go
r.GET("/nonce", func(c *gin.Context) {
    nonce, err := stputil.GenerateNonce(ctx)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"nonce": nonce})
})

r.POST("/transfer", func(c *gin.Context) {
    nonce := c.GetHeader("X-Nonce")

    if err := stputil.VerifyAndConsumeNonce(ctx, nonce); err != nil {
        c.JSON(401, gin.H{"error": "invalid_nonce"})
        return
    }

    c.JSON(200, gin.H{"message": "ok"})
})
```

## 最佳实践

1. 只给敏感写操作加 nonce，比如支付、转账、修改密码、删除操作
2. 建议和登录态一起使用，不要只校验 nonce 不校验用户身份
3. 对表单提交或一次性确认操作，5 分钟左右的 TTL 通常足够
4. 如果客户端需要先预校验，可先调用 `IsNonceValid`，真正提交时再消费

## 相关文档

- [OAuth2 指南](oauth2_zh.md)
- [登录认证](authentication_zh.md)
- [Refresh Token 指南](refresh-token_zh.md)
