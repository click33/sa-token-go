[English](refresh-token.md) | 中文文档

# Refresh Token 刷新令牌机制

## 什么是 Refresh Token？

Refresh Token 是一种长期有效的令牌，用于获取新的访问令牌（Access Token），无需用户重新登录。

### 令牌对比

| 特性 | Access Token | Refresh Token |
|------|--------------|---------------|
| 有效期 | 短期（2小时） | 长期（30天） |
| 用途 | 访问API资源 | 刷新Access Token |
| 存储位置 | 内存/本地存储 | 安全存储 |
| 传输频率 | 每次请求 | 仅刷新时 |
| 安全性 | 中等 | 高 |

## 为什么需要 Refresh Token？

### 问题：短期 Token 的困扰

```go
// Access Token 2小时后过期
token, _ := stputil.Login(1000)
time.Sleep(2 * time.Hour)

// 用户需要重新登录
isLogin := stputil.IsLogin(token)  // false
// 影响用户体验！
```

### 解决方案：Refresh Token

```go
// 使用 Refresh Token 机制
tokenInfo, _ := stputil.LoginWithRefreshToken(1000, "web")
// tokenInfo.AccessToken  - 2小时过期
// tokenInfo.RefreshToken - 30天过期

// 2小时后...
newInfo, _ := stputil.RefreshAccessToken(tokenInfo.RefreshToken)
// 获得新的 Access Token，用户无感知！
```

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
            Timeout(7200).  // Access Token 2小时
            Build(),
    )
}

func main() {
    // 1. 登录获取令牌对
    tokenInfo, err := stputil.LoginWithRefreshToken(1000, "web")
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Access Token:", tokenInfo.AccessToken)
    fmt.Println("Refresh Token:", tokenInfo.RefreshToken)
    fmt.Println("Expires at:", time.Unix(tokenInfo.ExpireTime, 0))
    
    // 2. 使用 Access Token
    // ... API 请求 ...
    
    // 3. Access Token 即将过期，刷新它
    newInfo, err := stputil.RefreshAccessToken(tokenInfo.RefreshToken)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("New Access Token:", newInfo.AccessToken)
    // Refresh Token 保持不变
}
```

## 完整流程

### 1. 登录端点

```go
r.POST("/login", func(c *gin.Context) {
    username := c.PostForm("username")
    password := c.PostForm("password")
    
    // 验证用户名密码
    userID := validateUser(username, password)
    if userID == "" {
        c.JSON(401, gin.H{"error": "Invalid credentials"})
        return
    }
    
    // 生成令牌对
    tokenInfo, err := stputil.LoginWithRefreshToken(userID, "web")
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{
        "access_token":  tokenInfo.AccessToken,
        "refresh_token": tokenInfo.RefreshToken,
        "token_type":    "Bearer",
        "expires_in":    7200,  // 2 hours
    })
})
```

### 2. 刷新端点

```go
r.POST("/refresh", func(c *gin.Context) {
    refreshToken := c.PostForm("refresh_token")
    
    // 刷新访问令牌
    newInfo, err := stputil.RefreshAccessToken(refreshToken)
    if err != nil {
        c.JSON(401, gin.H{"error": "Invalid refresh token"})
        return
    }
    
    c.JSON(200, gin.H{
        "access_token":  newInfo.AccessToken,
        "refresh_token": newInfo.RefreshToken,  // 可选：轮换刷新令牌
        "token_type":    "Bearer",
        "expires_in":    7200,
    })
})
```

### 3. API 访问保护

```go
r.GET("/api/user/profile", func(c *gin.Context) {
    token := c.GetHeader("Authorization")
    
    if !stputil.IsLogin(token) {
        c.JSON(401, gin.H{
            "error": "access_token_expired",
            "message": "Please refresh your token",
        })
        return
    }
    
    // 返回用户信息
    c.JSON(200, gin.H{"user": "..."})
})
```

### 4. 客户端自动刷新

```go
type TokenManager struct {
    accessToken  string
    refreshToken string
    expiresAt    time.Time
}

func (tm *TokenManager) GetValidToken() (string, error) {
    // 检查是否即将过期（提前5分钟刷新）
    if time.Now().Add(5 * time.Minute).After(tm.expiresAt) {
        // 自动刷新
        newInfo, err := refreshAccessToken(tm.refreshToken)
        if err != nil {
            return "", err
        }
        
        tm.accessToken = newInfo.AccessToken
        tm.refreshToken = newInfo.RefreshToken
        tm.expiresAt = time.Unix(newInfo.ExpireTime, 0)
    }
    
    return tm.accessToken, nil
}
```

## 高级用法

### 1. 刷新令牌轮换

```go
// 每次刷新时生成新的 Refresh Token
func (rtm *RefreshTokenManager) RefreshWithRotation(refreshToken string) (*RefreshTokenInfo, error) {
    // 验证旧的 Refresh Token
    oldInfo, err := rtm.RefreshAccessToken(refreshToken)
    if err != nil {
        return nil, err
    }
    
    // 撤销旧的 Refresh Token
    rtm.RevokeRefreshToken(refreshToken)
    
    // 生成新的令牌对
    return rtm.GenerateTokenPair(oldInfo.LoginID, oldInfo.Device)
}
```

### 2. 设备绑定

```go
// 为不同设备生成不同的令牌对
webTokens, _ := stputil.LoginWithRefreshToken(1000, "web")
mobileTokens, _ := stputil.LoginWithRefreshToken(1000, "mobile")

// 各自独立刷新
webNewTokens, _ := stputil.RefreshAccessToken(webTokens.RefreshToken)
mobileNewTokens, _ := stputil.RefreshAccessToken(mobileTokens.RefreshToken)
```

### 3. 撤销所有令牌

```go
// 登出时撤销所有令牌
func logout(userID string, refreshToken string) error {
    // 撤销 Refresh Token
    stputil.RevokeRefreshToken(refreshToken)
    
    // 撤销所有 Access Token
    stputil.Logout(userID)
    
    return nil
}
```

## 存储键结构

```
satoken:refresh:{refresh_token} → RefreshTokenInfo (TTL: 30天)

RefreshTokenInfo {
    RefreshToken: "c5f7e0d4..."
    AccessToken:  "b4f6d9c3..."
    LoginID:      "user123"
    Device:       "web"
    CreateTime:   1700000000
    ExpireTime:   1702592000
}
```

## 安全最佳实践

### 1. 安全存储

```javascript
// ❌ 不安全：存储在 localStorage
localStorage.setItem('refresh_token', token)

// ✅ 安全：存储在 httpOnly Cookie
document.cookie = `refresh_token=${token}; httpOnly; secure; sameSite=strict`

// ✅ 更安全：后端 Session
session.set('refresh_token', token)
```

### 2. HTTPS 传输

```
❌ HTTP  - 令牌可被截获
✅ HTTPS - 令牌加密传输
```

### 3. 定期轮换

```go
// 每次刷新时轮换 Refresh Token
const REFRESH_TOKEN_ROTATION = true

if REFRESH_TOKEN_ROTATION {
    // 生成新的令牌对，撤销旧的
    newPair, _ := manager.GenerateNewPairAndRevoke(oldRefreshToken)
}
```

### 4. 异常检测

```go
// 记录刷新事件
manager.RegisterFunc(core.EventRefresh, func(data *core.EventData) {
    // 检测异常刷新模式
    if isAbnormalRefreshPattern(data.LoginID) {
        alert("可能的令牌泄露")
    }
})
```

## 前端完整示例

### React Hook

```javascript
import { useState, useEffect } from 'react';

function useAuth() {
    const [accessToken, setAccessToken] = useState('');
    const [refreshToken, setRefreshToken] = useState('');
    
    // 登录
    const login = async (username, password) => {
        const resp = await fetch('/login', {
            method: 'POST',
            body: JSON.stringify({ username, password }),
        });
        const data = await resp.json();
        
        setAccessToken(data.access_token);
        setRefreshToken(data.refresh_token);
        
        // 启动自动刷新
        startAutoRefresh(data.expires_in);
    };
    
    // 自动刷新
    const startAutoRefresh = (expiresIn) => {
        // 提前5分钟刷新
        const refreshTime = (expiresIn - 300) * 1000;
        
        setTimeout(async () => {
            const resp = await fetch('/refresh', {
                method: 'POST',
                body: JSON.stringify({ refresh_token: refreshToken }),
            });
            const data = await resp.json();
            
            setAccessToken(data.access_token);
            setRefreshToken(data.refresh_token);
            
            // 继续自动刷新
            startAutoRefresh(data.expires_in);
        }, refreshTime);
    };
    
    return { accessToken, refreshToken, login };
}
```

## 监控和审计

### 记录刷新事件

```go
type RefreshLog struct {
    UserID       string
    Device       string
    RefreshTime  time.Time
    ClientIP     string
    UserAgent    string
}

func logRefreshEvent(info *core.RefreshTokenInfo, c *gin.Context) {
    log := RefreshLog{
        UserID:      info.LoginID,
        Device:      info.Device,
        RefreshTime: time.Now(),
        ClientIP:    c.ClientIP(),
        UserAgent:   c.GetHeader("User-Agent"),
    }
    
    saveToDatabase(log)
}
```

## 常见问题

### Q: Refresh Token 过期了怎么办？

A: 用户需要重新登录。建议设置足够长的有效期（如30天）。

### Q: 如何撤销 Refresh Token？

A: 调用 `stputil.RevokeRefreshToken(refreshToken)`。

### Q: Access Token 和 Refresh Token 的有效期如何配置？

A: Access Token 通过 `Timeout()` 配置，Refresh Token 固定30天。

### Q: 可以为一个用户生成多个 Refresh Token 吗？

A: 可以，每个设备可以有独立的 Refresh Token。

## 下一步

- [Nonce 防重放](nonce_zh.md)
- [OAuth2 授权](oauth2_zh.md)
- [安全特性示例](../../examples/security-features/)

