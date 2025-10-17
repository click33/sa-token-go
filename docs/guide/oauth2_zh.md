[English](oauth2.md) | 中文文档

# OAuth2 授权码模式

## 什么是 OAuth2？

OAuth2 是一个授权框架，允许第三方应用在用户授权下访问用户资源，而无需获取用户密码。

### OAuth2 角色

- **Resource Owner（资源所有者）** - 用户
- **Client（客户端）** - 第三方应用
- **Authorization Server（授权服务器）** - Sa-Token-Go
- **Resource Server（资源服务器）** - API 服务器

## 授权码模式流程

```
┌─────────┐                                  ┌──────────┐
│  用户    │                                  │  客户端  │
└────┬────┘                                  └─────┬────┘
     │                                             │
     │  1. 访问应用                                 │
     ├────────────────────────────────────────────>│
     │                                             │
     │  2. 重定向到授权页面                         │
     │<────────────────────────────────────────────┤
     │                                             │
┌────┴────┐                                  ┌─────┴────┐
│授权服务器│                                  │  客户端  │
└────┬────┘                                  └─────┬────┘
     │                                             │
     │  3. 用户授权                                 │
     ├────────────────────────────────────────────>│
     │                                             │
     │  4. 返回授权码                               │
     │<────────────────────────────────────────────┤
     │                                             │
     │  5. 用授权码换取令牌                          │
     ├────────────────────────────────────────────>│
     │                                             │
     │  6. 返回访问令牌                             │
     │<────────────────────────────────────────────┤
     │                                             │
```

## 快速开始

### 1. 初始化 OAuth2 服务器

```go
import (
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/storage/memory"
)

func init() {
    storage := memory.NewStorage()
    manager := core.NewBuilder().
        Storage(storage).
        Timeout(7200).
        Build()
    
    stputil.SetManager(manager)
}

func main() {
    // 获取 OAuth2 服务器
    oauth2Server := stputil.GetOAuth2Server()
    
    // 注册客户端
    oauth2Server.RegisterClient(&core.OAuth2Client{
        ClientID:     "my-app",
        ClientSecret: "my-secret",
        RedirectURIs: []string{"http://localhost:3000/callback"},
        GrantTypes:   []core.OAuth2GrantType{
            core.GrantTypeAuthorizationCode,
            core.GrantTypeRefreshToken,
        },
        Scopes: []string{"read", "write", "profile"},
    })
}
```

### 2. 授权端点

```go
// GET /oauth/authorize
func authorizeHandler(c *gin.Context) {
    clientID := c.Query("client_id")
    redirectURI := c.Query("redirect_uri")
    state := c.Query("state")
    scope := c.Query("scope")
    
    // 验证用户登录状态
    userID := getCurrentUserID(c)
    if userID == "" {
        c.Redirect(302, "/login")
        return
    }
    
    // 生成授权码
    scopes := strings.Split(scope, " ")
    authCode, err := oauth2Server.GenerateAuthorizationCode(
        clientID,
        redirectURI,
        userID,
        scopes,
    )
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // 重定向回客户端
    redirectURL := fmt.Sprintf("%s?code=%s&state=%s", 
        redirectURI, authCode.Code, state)
    c.Redirect(302, redirectURL)
}
```

### 3. 令牌端点

```go
// POST /oauth/token
func tokenHandler(c *gin.Context) {
    grantType := c.PostForm("grant_type")
    
    switch grantType {
    case "authorization_code":
        handleAuthorizationCodeGrant(c)
    case "refresh_token":
        handleRefreshTokenGrant(c)
    default:
        c.JSON(400, gin.H{"error": "unsupported_grant_type"})
    }
}

func handleAuthorizationCodeGrant(c *gin.Context) {
    code := c.PostForm("code")
    clientID := c.PostForm("client_id")
    clientSecret := c.PostForm("client_secret")
    redirectURI := c.PostForm("redirect_uri")
    
    // 用授权码换取令牌
    token, err := oauth2Server.ExchangeCodeForToken(
        code, clientID, clientSecret, redirectURI,
    )
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{
        "access_token":  token.Token,
        "token_type":    token.TokenType,
        "expires_in":    token.ExpiresIn,
        "refresh_token": token.RefreshToken,
        "scope":         strings.Join(token.Scopes, " "),
    })
}
```

### 4. 资源访问保护

```go
// GET /api/userinfo
func userinfoHandler(c *gin.Context) {
    authHeader := c.GetHeader("Authorization")
    
    var token string
    fmt.Sscanf(authHeader, "Bearer %s", &token)
    
    // 验证访问令牌
    accessToken, err := oauth2Server.ValidateAccessToken(token)
    if err != nil {
        c.JSON(401, gin.H{"error": "Invalid access token"})
        return
    }
    
    // 返回用户信息
    c.JSON(200, gin.H{
        "user_id": accessToken.UserID,
        "scopes":  accessToken.Scopes,
    })
}
```

## 支持的授权类型

### 1. 授权码模式（Authorization Code）

最安全的授权模式，适用于有后端的应用。

```go
// 注册客户端时指定
GrantTypes: []core.OAuth2GrantType{
    core.GrantTypeAuthorizationCode,
}
```

**流程：**
1. 用户授权 → 获取授权码
2. 用授权码 + 客户端凭证 → 换取令牌

### 2. 刷新令牌模式（Refresh Token）

用于刷新过期的访问令牌。

```go
GrantTypes: []core.OAuth2GrantType{
    core.GrantTypeRefreshToken,
}
```

### 3. 客户端凭证模式（Client Credentials）

适用于服务间通信。

```go
GrantTypes: []core.OAuth2GrantType{
    core.GrantTypeClientCredentials,
}
```

### 4. 密码模式（Password）

用户直接提供用户名密码（不推荐）。

```go
GrantTypes: []core.OAuth2GrantType{
    core.GrantTypePassword,
}
```

## Scope 权限管理

### 定义 Scope

```go
oauth2Server.RegisterClient(&core.OAuth2Client{
    ClientID: "webapp",
    Scopes:   []string{
        "read",        // 读取权限
        "write",       // 写入权限
        "profile",     // 个人资料
        "email",       // 邮箱
        "admin",       // 管理员
    },
})
```

### 请求特定 Scope

```go
// 用户授权时指定需要的权限
authCode, _ := oauth2Server.GenerateAuthorizationCode(
    "webapp",
    "http://localhost:3000/callback",
    "user123",
    []string{"read", "profile"},  // 仅请求这两个权限
)
```

### 验证 Scope

```go
func requireScope(requiredScope string) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := getAccessToken(c)
        
        accessToken, _ := oauth2Server.ValidateAccessToken(token)
        
        hasScope := false
        for _, scope := range accessToken.Scopes {
            if scope == requiredScope {
                hasScope = true
                break
            }
        }
        
        if !hasScope {
            c.JSON(403, gin.H{"error": "insufficient_scope"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// 使用
r.GET("/api/profile", requireScope("profile"), profileHandler)
r.POST("/api/data", requireScope("write"), dataHandler)
```

## 完整示例

查看完整的 OAuth2 服务器实现：[examples/oauth2-example](../../examples/oauth2-example/)

## 客户端集成示例

### cURL 测试

```bash
# 1. 获取授权码
curl "http://localhost:8080/oauth/authorize?client_id=webapp&redirect_uri=http://localhost:8080/callback&response_type=code&state=xyz"

# 2. 换取令牌
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=authorization_code" \
  -d "code=授权码" \
  -d "client_id=webapp" \
  -d "client_secret=secret123" \
  -d "redirect_uri=http://localhost:8080/callback"

# 3. 访问资源
curl http://localhost:8080/oauth/userinfo \
  -H "Authorization: Bearer 访问令牌"

# 4. 刷新令牌
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=refresh_token" \
  -d "refresh_token=刷新令牌" \
  -d "client_id=webapp" \
  -d "client_secret=secret123"
```

### Go 客户端

```go
type OAuth2Client struct {
    ClientID     string
    ClientSecret string
    RedirectURI  string
    AuthEndpoint string
    TokenEndpoint string
}

func (c *OAuth2Client) GetAuthorizationURL(state string, scopes []string) string {
    return fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&state=%s&scope=%s",
        c.AuthEndpoint,
        c.ClientID,
        c.RedirectURI,
        state,
        strings.Join(scopes, " "),
    )
}

func (c *OAuth2Client) ExchangeCode(code string) (*TokenResponse, error) {
    data := url.Values{
        "grant_type":    []string{"authorization_code"},
        "code":          []string{code},
        "client_id":     []string{c.ClientID},
        "client_secret": []string{c.ClientSecret},
        "redirect_uri":  []string{c.RedirectURI},
    }
    
    resp, err := http.PostForm(c.TokenEndpoint, data)
    if err != nil {
        return nil, err
    }
    
    var result TokenResponse
    json.NewDecoder(resp.Body).Decode(&result)
    return &result, nil
}
```

## 安全最佳实践

### 1. State 参数（CSRF 保护）

```go
// 生成随机 state
state := generateRandomString(32)
session.Set("oauth_state", state)

// 验证 state
callbackState := c.Query("state")
sessionState := session.Get("oauth_state")

if callbackState != sessionState {
    c.JSON(400, gin.H{"error": "Invalid state"})
    return
}
```

### 2. PKCE（增强安全性）

```go
// 生成 code_verifier 和 code_challenge
codeVerifier := generateRandomString(64)
codeChallenge := base64.URLEncoding.EncodeToString(
    sha256.Sum256([]byte(codeVerifier)),
)

// 授权时发送 code_challenge
// 换取令牌时发送 code_verifier
```

### 3. 客户端凭证加密存储

```go
// ❌ 明文存储
ClientSecret: "secret123"

// ✅ 加密存储
ClientSecret: hashPassword("secret123")
```

### 4. Redirect URI 白名单

```go
oauth2Server.RegisterClient(&core.OAuth2Client{
    RedirectURIs: []string{
        "https://app.example.com/callback",  // ✅ HTTPS
        "http://localhost:3000/callback",     // ✅ 开发环境
    },
})

// ❌ 不允许
"http://example.com/callback"     // HTTP（生产）
"https://evil.com/callback"        // 未注册
```

## 错误处理

### 标准 OAuth2 错误

```go
type OAuth2Error struct {
    Error            string `json:"error"`
    ErrorDescription string `json:"error_description,omitempty"`
    ErrorURI         string `json:"error_uri,omitempty"`
}

// 常见错误
errors := map[string]string{
    "invalid_request":          "请求参数无效",
    "invalid_client":           "客户端认证失败",
    "invalid_grant":            "授权码无效或已过期",
    "unauthorized_client":      "客户端未授权此操作",
    "unsupported_grant_type":   "不支持的授权类型",
    "invalid_scope":            "请求的权限无效",
}
```

## 生产环境配置

### Redis 存储

```go
import (
    "github.com/click33/sa-token-go/storage/redis"
)

func init() {
    redisStorage, _ := redis.NewStorage(&redis.Config{
        Addr: "localhost:6379",
    })
    
    manager := core.NewBuilder().
        Storage(redisStorage).
        Timeout(7200).  // 2小时
        Build()
    
    stputil.SetManager(manager)
}
```

### 客户端管理

```go
// 从数据库加载客户端
func loadClientsFromDB() {
    clients := queryClientsFromDB()
    
    oauth2Server := stputil.GetOAuth2Server()
    for _, client := range clients {
        oauth2Server.RegisterClient(&core.OAuth2Client{
            ClientID:     client.ID,
            ClientSecret: client.Secret,
            RedirectURIs: client.RedirectURIs,
            GrantTypes:   client.GrantTypes,
            Scopes:       client.Scopes,
        })
    }
}
```

## 监控和审计

### 记录授权事件

```go
type OAuth2Log struct {
    EventType   string    // authorize, token, revoke
    ClientID    string
    UserID      string
    Scopes      []string
    Timestamp   time.Time
    ClientIP    string
    UserAgent   string
    Success     bool
    ErrorMsg    string
}

func logOAuth2Event(event OAuth2Log) {
    // 保存到数据库
    db.Create(&event)
    
    // 异常检测
    if isAbnormalPattern(event) {
        alertSecurity(event)
    }
}
```

### 使用统计

```go
type OAuth2Stats struct {
    ClientID         string
    TotalAuthorizes  int64
    TotalTokens      int64
    TotalRefreshes   int64
    ActiveUsers      int64
    LastUsed         time.Time
}

func trackOAuth2Usage(clientID string) {
    stats := getStats(clientID)
    stats.TotalTokens++
    stats.LastUsed = time.Now()
    saveStats(stats)
}
```

## 高级特性

### 1. 动态 Scope 授权

```go
// 用户可选择授权哪些权限
r.POST("/oauth/authorize", func(c *gin.Context) {
    requestedScopes := c.PostFormArray("scopes")  // ["read", "write", "profile"]
    grantedScopes := c.PostFormArray("granted")   // ["read", "write"]  // 用户只授权了这两个
    
    authCode, _ := oauth2Server.GenerateAuthorizationCode(
        clientID, redirectURI, userID, grantedScopes,
    )
})
```

### 2. 令牌撤销

```go
// POST /oauth/revoke
func revokeHandler(c *gin.Context) {
    token := c.PostForm("token")
    tokenTypeHint := c.PostForm("token_type_hint")  // "access_token" or "refresh_token"
    
    if err := oauth2Server.RevokeToken(token); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"message": "revoked"})
}
```

### 3. 令牌内省

```go
// POST /oauth/introspect
func introspectHandler(c *gin.Context) {
    token := c.PostForm("token")
    
    accessToken, err := oauth2Server.ValidateAccessToken(token)
    if err != nil {
        c.JSON(200, gin.H{"active": false})
        return
    }
    
    c.JSON(200, gin.H{
        "active":    true,
        "client_id": accessToken.ClientID,
        "user_id":   accessToken.UserID,
        "scope":     strings.Join(accessToken.Scopes, " "),
        "exp":       time.Now().Unix() + accessToken.ExpiresIn,
    })
}
```

## 常见问题

### Q: 授权码和访问令牌有什么区别？

A: 
- 授权码：临时凭证（10分钟），仅用于换取令牌
- 访问令牌：用于访问资源（2小时）

### Q: 如何实现第三方登录（微信、GitHub）？

A: Sa-Token-Go 作为授权服务器，你的应用作为客户端，接入微信/GitHub 的 OAuth2。

### Q: 如何撤销所有令牌？

A: 遍历用户的所有 refresh token 并撤销。

### Q: 支持 PKCE 吗？

A: 当前版本支持标准授权码模式，PKCE 可以在应用层实现。

## 性能优化

### 1. 授权码缓存

```
单次生成: ~200ns
并发安全: ✅
过期清理: 自动
```

### 2. 令牌验证

```
内存验证: ~50ns
Redis验证: ~1ms
```

## 下一步

- [Nonce 防重放](nonce_zh.md)
- [Refresh Token](refresh-token_zh.md)
- [OAuth2 完整示例](../../examples/oauth2-example/)
- [安全特性示例](../../examples/security-features/)

