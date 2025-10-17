English | [中文文档](refresh-token_zh.md)

# Refresh Token Mechanism

## What is a Refresh Token?

A Refresh Token is a long-lived token used to obtain new access tokens without requiring user re-authentication.

### Token Comparison

| Feature | Access Token | Refresh Token |
|---------|--------------|---------------|
| Validity | Short (2 hours) | Long (30 days) |
| Purpose | Access API resources | Refresh Access Token |
| Storage | Memory/Local storage | Secure storage |
| Frequency | Every request | Only when refreshing |
| Security | Medium | High |

## Why Do We Need Refresh Tokens?

### Problem: Short-lived Token Hassle

```go
// Access token expires after 2 hours
token, _ := stputil.Login(1000)
time.Sleep(2 * time.Hour)

// User needs to login again
isLogin := stputil.IsLogin(token)  // false
// Poor user experience!
```

### Solution: Refresh Token

```go
// Use Refresh Token mechanism
tokenInfo, _ := stputil.LoginWithRefreshToken(1000, "web")
// tokenInfo.AccessToken  - expires in 2 hours
// tokenInfo.RefreshToken - expires in 30 days

// After 2 hours...
newInfo, _ := stputil.RefreshAccessToken(tokenInfo.RefreshToken)
// Get new Access Token, seamless for user!
```

## Quick Start

### Basic Usage

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
            Timeout(7200).  // Access Token 2 hours
            Build(),
    )
}

func main() {
    // 1. Login and get token pair
    tokenInfo, err := stputil.LoginWithRefreshToken(1000, "web")
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Access Token:", tokenInfo.AccessToken)
    fmt.Println("Refresh Token:", tokenInfo.RefreshToken)
    fmt.Println("Expires at:", time.Unix(tokenInfo.ExpireTime, 0))
    
    // 2. Use Access Token
    // ... API requests ...
    
    // 3. When Access Token expires, refresh it
    newInfo, err := stputil.RefreshAccessToken(tokenInfo.RefreshToken)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("New Access Token:", newInfo.AccessToken)
    // Refresh Token remains the same
}
```

## Complete Workflow

### 1. Login Endpoint

```go
r.POST("/login", func(c *gin.Context) {
    username := c.PostForm("username")
    password := c.PostForm("password")
    
    // Validate credentials
    userID := validateUser(username, password)
    if userID == "" {
        c.JSON(401, gin.H{"error": "Invalid credentials"})
        return
    }
    
    // Generate token pair
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

### 2. Refresh Endpoint

```go
r.POST("/refresh", func(c *gin.Context) {
    refreshToken := c.PostForm("refresh_token")
    
    // Refresh access token
    newInfo, err := stputil.RefreshAccessToken(refreshToken)
    if err != nil {
        c.JSON(401, gin.H{"error": "Invalid refresh token"})
        return
    }
    
    c.JSON(200, gin.H{
        "access_token":  newInfo.AccessToken,
        "refresh_token": newInfo.RefreshToken,  // Optional: rotate refresh token
        "token_type":    "Bearer",
        "expires_in":    7200,
    })
})
```

### 3. API Access Protection

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
    
    // Return user info
    c.JSON(200, gin.H{"user": "..."})
})
```

### 4. Auto-Refresh Client

```go
type TokenManager struct {
    accessToken  string
    refreshToken string
    expiresAt    time.Time
}

func (tm *TokenManager) GetValidToken() (string, error) {
    // Check if about to expire (refresh 5 minutes early)
    if time.Now().Add(5 * time.Minute).After(tm.expiresAt) {
        // Auto refresh
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

## Advanced Usage

### 1. Refresh Token Rotation

```go
// Generate new Refresh Token on each refresh
func (rtm *RefreshTokenManager) RefreshWithRotation(refreshToken string) (*RefreshTokenInfo, error) {
    // Validate old Refresh Token
    oldInfo, err := rtm.RefreshAccessToken(refreshToken)
    if err != nil {
        return nil, err
    }
    
    // Revoke old Refresh Token
    rtm.RevokeRefreshToken(refreshToken)
    
    // Generate new token pair
    return rtm.GenerateTokenPair(oldInfo.LoginID, oldInfo.Device)
}
```

### 2. Device Binding

```go
// Generate separate token pairs for different devices
webTokens, _ := stputil.LoginWithRefreshToken(1000, "web")
mobileTokens, _ := stputil.LoginWithRefreshToken(1000, "mobile")

// Refresh independently
webNewTokens, _ := stputil.RefreshAccessToken(webTokens.RefreshToken)
mobileNewTokens, _ := stputil.RefreshAccessToken(mobileTokens.RefreshToken)
```

### 3. Revoke All Tokens

```go
// Revoke all tokens on logout
func logout(userID string, refreshToken string) error {
    // Revoke Refresh Token
    stputil.RevokeRefreshToken(refreshToken)
    
    // Revoke all Access Tokens
    stputil.Logout(userID)
    
    return nil
}
```

## Storage Key Structure

```
satoken:refresh:{refresh_token} → RefreshTokenInfo (TTL: 30 days)

RefreshTokenInfo {
    RefreshToken: "c5f7e0d4..."
    AccessToken:  "b4f6d9c3..."
    LoginID:      "user123"
    Device:       "web"
    CreateTime:   1700000000
    ExpireTime:   1702592000
}
```

## Security Best Practices

### 1. Secure Storage

```javascript
// ❌ Insecure: localStorage
localStorage.setItem('refresh_token', token)

// ✅ Secure: httpOnly Cookie
document.cookie = `refresh_token=${token}; httpOnly; secure; sameSite=strict`

// ✅ More secure: Backend session
session.set('refresh_token', token)
```

### 2. HTTPS Transport

```
❌ HTTP  - Token can be intercepted
✅ HTTPS - Token encrypted in transit
```

### 3. Regular Rotation

```go
// Rotate Refresh Token on each refresh
const REFRESH_TOKEN_ROTATION = true

if REFRESH_TOKEN_ROTATION {
    // Generate new pair, revoke old
    newPair, _ := manager.GenerateNewPairAndRevoke(oldRefreshToken)
}
```

### 4. Anomaly Detection

```go
// Log refresh events
eventMgr.RegisterFunc(core.EventRefresh, func(data *core.EventData) {
    // Detect abnormal refresh patterns
    if isAbnormalRefreshPattern(data.LoginID) {
        alert("Possible token leak")
    }
})
```

## Frontend Complete Example

### React Hook

```javascript
import { useState, useEffect } from 'react';

function useAuth() {
    const [accessToken, setAccessToken] = useState('');
    const [refreshToken, setRefreshToken] = useState('');
    
    // Login
    const login = async (username, password) => {
        const resp = await fetch('/login', {
            method: 'POST',
            body: JSON.stringify({ username, password }),
        });
        const data = await resp.json();
        
        setAccessToken(data.access_token);
        setRefreshToken(data.refresh_token);
        
        // Start auto-refresh
        startAutoRefresh(data.expires_in);
    };
    
    // Auto refresh
    const startAutoRefresh = (expiresIn) => {
        // Refresh 5 minutes early
        const refreshTime = (expiresIn - 300) * 1000;
        
        setTimeout(async () => {
            const resp = await fetch('/refresh', {
                method: 'POST',
                body: JSON.stringify({ refresh_token: refreshToken }),
            });
            const data = await resp.json();
            
            setAccessToken(data.access_token);
            setRefreshToken(data.refresh_token);
            
            // Continue auto-refresh
            startAutoRefresh(data.expires_in);
        }, refreshTime);
    };
    
    return { accessToken, refreshToken, login };
}
```

## Monitoring and Auditing

### Log Refresh Events

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

## FAQ

### Q: What happens when Refresh Token expires?

A: User needs to login again. Recommend setting a sufficiently long TTL (e.g., 30 days).

### Q: How to revoke a Refresh Token?

A: Call `stputil.RevokeRefreshToken(refreshToken)`.

### Q: How to configure TTL for Access and Refresh Tokens?

A: Access Token via `Timeout()`, Refresh Token is fixed at 30 days.

### Q: Can multiple Refresh Tokens be generated for one user?

A: Yes, each device can have its own Refresh Token.

## Next Steps

- [Nonce Anti-Replay](nonce.md)
- [OAuth2 Authorization](oauth2.md)
- [Security Features Example](../../examples/security-features/)

