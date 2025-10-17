[English](nonce.md) | 中文文档

# Nonce 防重放攻击

## 什么是重放攻击？

重放攻击是指攻击者截获并重复发送合法请求，以达到非法目的。例如：
- 攻击者截获转账请求，重复发送导致多次扣款
- 攻击者截获登录请求，重复登录获取多个 Token
- 攻击者截获操作请求，重复执行敏感操作

## Nonce 防重放原理

**Nonce** (Number used once) 是一次性随机数：
1. 服务器生成唯一的 Nonce
2. 客户端在请求中携带 Nonce
3. 服务器验证 Nonce 并立即删除
4. 相同的 Nonce 无法再次使用

## 快速开始

### 基本使用

```go
import (
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/storage/memory"
)

func init() {
    stputil.SetManager(
        core.NewBuilder().
            Storage(memory.NewStorage()).
            Build(),
    )
}

func main() {
    // 1. 生成 Nonce
    nonce, err := stputil.GenerateNonce()
    if err != nil {
        panic(err)
    }
    fmt.Println("Nonce:", nonce)
    // 输出: 64字符十六进制字符串
    
    // 2. 首次验证（成功）
    valid := stputil.VerifyNonce(nonce)
    fmt.Println("First verify:", valid)  // true
    
    // 3. 再次验证（失败 - 防重放）
    valid = stputil.VerifyNonce(nonce)
    fmt.Println("Second verify:", valid)  // false
}
```

## 完整流程

### 1. API 端点保护

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/click33/sa-token-go/stputil"
)

func main() {
    r := gin.Default()
    
    // 生成 Nonce
    r.GET("/nonce", func(c *gin.Context) {
        nonce, err := stputil.GenerateNonce()
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        c.JSON(200, gin.H{"nonce": nonce})
    })
    
    // 需要 Nonce 保护的 API
    r.POST("/transfer", func(c *gin.Context) {
        nonce := c.GetHeader("X-Nonce")
        
        // 验证 Nonce
        if !stputil.VerifyNonce(nonce) {
            c.JSON(401, gin.H{"error": "Invalid or expired nonce"})
            return
        }
        
        // 执行转账逻辑
        amount := c.PostForm("amount")
        c.JSON(200, gin.H{
            "message": "Transfer successful",
            "amount":  amount,
        })
    })
    
    r.Run(":8080")
}
```

### 2. 客户端使用

```go
// Step 1: 获取 Nonce
resp1, _ := http.Get("http://localhost:8080/nonce")
var result map[string]string
json.NewDecoder(resp1.Body).Decode(&result)
nonce := result["nonce"]

// Step 2: 使用 Nonce 发起请求
req, _ := http.NewRequest("POST", "http://localhost:8080/transfer", nil)
req.Header.Set("X-Nonce", nonce)
req.PostForm = url.Values{
    "amount": []string{"100"},
}

resp2, _ := http.DefaultClient.Do(req)
// 转账成功

// Step 3: 重复请求（会失败）
resp3, _ := http.DefaultClient.Do(req)
// 失败：Invalid or expired nonce
```

## Nonce 配置

### 自定义过期时间

```go
import "time"

// 方式 1：通过 Manager 创建
manager := core.NewBuilder().
    Storage(storage).
    Build()

// 获取 NonceManager 并配置
nonceManager := core.NewNonceManager(storage, 300) // 5分钟（秒）
```

### 默认配置

```go
// 默认过期时间：5分钟
// 默认长度：64字符（32字节十六进制）
```

## 高级用法

### 1. 中间件保护

```go
func NonceMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 跳过GET请求
        if c.Request.Method == "GET" {
            c.Next()
            return
        }
        
        nonce := c.GetHeader("X-Nonce")
        
        if !stputil.VerifyNonce(nonce) {
            c.JSON(401, gin.H{"error": "Invalid or expired nonce"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// 使用
r.Use(NonceMiddleware())
```

### 2. 敏感操作保护

```go
// 仅保护敏感操作
r.POST("/delete-account", NonceMiddleware(), deleteAccountHandler)
r.POST("/transfer-money", NonceMiddleware(), transferHandler)
r.POST("/change-password", NonceMiddleware(), changePasswordHandler)
```

### 3. 批量验证

```go
func verifyMultipleNonces(nonces []string) bool {
    for _, nonce := range nonces {
        if !stputil.VerifyNonce(nonce) {
            return false
        }
    }
    return true
}
```

## 最佳实践

### 1. 仅保护敏感操作

```go
// ✅ 需要 Nonce
POST /transfer       // 转账
POST /delete         // 删除
POST /change-email   // 修改邮箱

// ❌ 不需要 Nonce
GET  /list           // 查询
GET  /detail         // 详情
POST /search         // 搜索
```

### 2. 设置合理的过期时间

```go
// 快速操作（1分钟）
core.NewNonceManager(storage, 60)

// 表单提交（5分钟，默认）
core.NewNonceManager(storage, 300)

// 长流程操作（10分钟）
core.NewNonceManager(storage, 600)
```

### 3. 清晰的错误提示

```go
if !stputil.VerifyNonce(nonce) {
    c.JSON(401, gin.H{
        "error": "invalid_nonce",
        "message": "请求已过期或重复，请刷新页面重试",
        "code": 1001,
    })
    return
}
```

### 4. 前端集成

```javascript
// 前端示例（Vue/React）
async function protectedRequest(url, data) {
    // 1. 获取 Nonce
    const nonceResp = await fetch('/nonce');
    const { nonce } = await nonceResp.json();
    
    // 2. 发起请求
    const resp = await fetch(url, {
        method: 'POST',
        headers: {
            'X-Nonce': nonce,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    });
    
    return resp.json();
}

// 使用
protectedRequest('/transfer', { amount: 100 })
    .then(result => console.log('Success:', result))
    .catch(err => console.error('Error:', err));
```

## 存储键结构

```
satoken:nonce:{nonce_value} → timestamp (TTL: 5分钟)
```

## 性能优化

### 1. Nonce 生成性能

```
单次生成: ~100ns
10000次: ~1ms
并发安全: ✅
```

### 2. 验证性能

```
单次验证: ~50ns (内存)
单次验证: ~1ms (Redis)
```

### 3. 内存占用

```
单个 Nonce: ~100 bytes
10000个 Nonce: ~1MB
过期自动清理: ✅
```

## 安全建议

### 1. HTTPS 传输

```
❌ HTTP  - Nonce可被截获
✅ HTTPS - Nonce加密传输
```

### 2. 结合 Token 认证

```go
// 同时验证 Token 和 Nonce
token := c.GetHeader("Authorization")
nonce := c.GetHeader("X-Nonce")

if !stputil.IsLogin(token) {
    c.JSON(401, gin.H{"error": "未登录"})
    return
}

if !stputil.VerifyNonce(nonce) {
    c.JSON(401, gin.H{"error": "请求重放"})
    return
}
```

### 3. 限流配合

```go
// Nonce + 限流双重保护
r.Use(RateLimitMiddleware())
r.Use(NonceMiddleware())
```

## 常见问题

### Q: Nonce 过期了怎么办？

A: 客户端重新请求 `/nonce` 端点获取新的 Nonce。

### Q: 如何避免 Nonce 被截获？

A: 必须使用 HTTPS，配合 Token 认证双重保护。

### Q: Nonce 会占用大量存储吗？

A: 不会，Nonce 自动过期清理，内存占用很小。

### Q: 所有 API 都需要 Nonce 吗？

A: 不需要，只保护敏感的写操作（POST/PUT/DELETE）。

## 下一步

- [Refresh Token 指南](refresh-token_zh.md)
- [OAuth2 指南](oauth2_zh.md)
- [安全特性示例](../../examples/security-features/)

