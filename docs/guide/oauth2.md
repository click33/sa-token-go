English | [中文文档](oauth2_zh.md)

# OAuth2 Authorization Code Flow

## What is OAuth2?

OAuth2 is an authorization framework that allows third-party applications to access user resources with user authorization, without obtaining user passwords.

### OAuth2 Roles

- **Resource Owner** - User
- **Client** - Third-party application
- **Authorization Server** - Sa-Token-Go
- **Resource Server** - API server

## Authorization Code Flow

```
┌─────────┐                                  ┌──────────┐
│  User    │                                  │  Client  │
└────┬────┘                                  └─────┬────┘
     │                                             │
     │  1. Access application                      │
     ├────────────────────────────────────────────>│
     │                                             │
     │  2. Redirect to authorization page          │
     │<────────────────────────────────────────────┤
     │                                             │
┌────┴────┐                                  ┌─────┴────┐
│  Auth    │                                  │  Client  │
│ Server   │                                  │          │
└────┬────┘                                  └─────┬────┘
     │                                             │
     │  3. User authorizes                         │
     ├────────────────────────────────────────────>│
     │                                             │
     │  4. Return authorization code               │
     │<────────────────────────────────────────────┤
     │                                             │
     │  5. Exchange code for token                 │
     ├────────────────────────────────────────────>│
     │                                             │
     │  6. Return access token                     │
     │<────────────────────────────────────────────┤
     │                                             │
```

## Quick Start

### 1. Initialize OAuth2 Server

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
    // Get OAuth2 server
    oauth2Server := stputil.GetOAuth2Server()
    
    // Register client
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

### 2. Authorization Endpoint

```go
// GET /oauth/authorize
func authorizeHandler(c *gin.Context) {
    clientID := c.Query("client_id")
    redirectURI := c.Query("redirect_uri")
    state := c.Query("state")
    scope := c.Query("scope")
    
    // Verify user login status
    userID := getCurrentUserID(c)
    if userID == "" {
        c.Redirect(302, "/login")
        return
    }
    
    // Generate authorization code
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
    
    // Redirect back to client
    redirectURL := fmt.Sprintf("%s?code=%s&state=%s", 
        redirectURI, authCode.Code, state)
    c.Redirect(302, redirectURL)
}
```

### 3. Token Endpoint

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
    
    // Exchange code for token
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

### 4. Resource Access Protection

```go
// GET /api/userinfo
func userinfoHandler(c *gin.Context) {
    authHeader := c.GetHeader("Authorization")
    
    var token string
    fmt.Sscanf(authHeader, "Bearer %s", &token)
    
    // Validate access token
    accessToken, err := oauth2Server.ValidateAccessToken(token)
    if err != nil {
        c.JSON(401, gin.H{"error": "Invalid access token"})
        return
    }
    
    // Return user info
    c.JSON(200, gin.H{
        "user_id": accessToken.UserID,
        "scopes":  accessToken.Scopes,
    })
}
```

## Supported Grant Types

### 1. Authorization Code

Most secure grant type, suitable for server-side applications.

```go
// Specify when registering client
GrantTypes: []core.OAuth2GrantType{
    core.GrantTypeAuthorizationCode,
}
```

**Flow:**
1. User authorizes → Get authorization code
2. Code + client credentials → Exchange for token

### 2. Refresh Token

Used to refresh expired access tokens.

```go
GrantTypes: []core.OAuth2GrantType{
    core.GrantTypeRefreshToken,
}
```

### 3. Client Credentials

Suitable for service-to-service communication.

```go
GrantTypes: []core.OAuth2GrantType{
    core.GrantTypeClientCredentials,
}
```

### 4. Password

User provides username and password directly (not recommended).

```go
GrantTypes: []core.OAuth2GrantType{
    core.GrantTypePassword,
}
```

## Scope Management

### Define Scopes

```go
oauth2Server.RegisterClient(&core.OAuth2Client{
    ClientID: "webapp",
    Scopes:   []string{
        "read",        // Read permission
        "write",       // Write permission
        "profile",     // Profile access
        "email",       // Email access
        "admin",       // Admin access
    },
})
```

### Request Specific Scopes

```go
// Specify required scopes during authorization
authCode, _ := oauth2Server.GenerateAuthorizationCode(
    "webapp",
    "http://localhost:3000/callback",
    "user123",
    []string{"read", "profile"},  // Only request these two scopes
)
```

### Validate Scopes

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

// Usage
r.GET("/api/profile", requireScope("profile"), profileHandler)
r.POST("/api/data", requireScope("write"), dataHandler)
```

## Complete Example

View complete OAuth2 server implementation: [examples/oauth2-example](../../examples/oauth2-example/)

## Client Integration Example

### cURL Testing

```bash
# 1. Get authorization code
curl "http://localhost:8080/oauth/authorize?client_id=webapp&redirect_uri=http://localhost:8080/callback&response_type=code&state=xyz"

# 2. Exchange for token
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=authorization_code" \
  -d "code=AUTH_CODE" \
  -d "client_id=webapp" \
  -d "client_secret=secret123" \
  -d "redirect_uri=http://localhost:8080/callback"

# 3. Access resource
curl http://localhost:8080/oauth/userinfo \
  -H "Authorization: Bearer ACCESS_TOKEN"

# 4. Refresh token
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=refresh_token" \
  -d "refresh_token=REFRESH_TOKEN" \
  -d "client_id=webapp" \
  -d "client_secret=secret123"
```

### Go Client

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

## Security Best Practices

### 1. State Parameter (CSRF Protection)

```go
// Generate random state
state := generateRandomString(32)
session.Set("oauth_state", state)

// Validate state
callbackState := c.Query("state")
sessionState := session.Get("oauth_state")

if callbackState != sessionState {
    c.JSON(400, gin.H{"error": "Invalid state"})
    return
}
```

### 2. PKCE (Enhanced Security)

```go
// Generate code_verifier and code_challenge
codeVerifier := generateRandomString(64)
codeChallenge := base64.URLEncoding.EncodeToString(
    sha256.Sum256([]byte(codeVerifier)),
)

// Send code_challenge during authorization
// Send code_verifier when exchanging token
```

### 3. Encrypt Client Credentials

```go
// ❌ Plain text storage
ClientSecret: "secret123"

// ✅ Encrypted storage
ClientSecret: hashPassword("secret123")
```

### 4. Redirect URI Whitelist

```go
oauth2Server.RegisterClient(&core.OAuth2Client{
    RedirectURIs: []string{
        "https://app.example.com/callback",  // ✅ HTTPS
        "http://localhost:3000/callback",     // ✅ Development
    },
})

// ❌ Not allowed
"http://example.com/callback"     // HTTP (production)
"https://evil.com/callback"        // Not registered
```

## Error Handling

### Standard OAuth2 Errors

```go
type OAuth2Error struct {
    Error            string `json:"error"`
    ErrorDescription string `json:"error_description,omitempty"`
    ErrorURI         string `json:"error_uri,omitempty"`
}

// Common errors
errors := map[string]string{
    "invalid_request":          "Invalid request parameters",
    "invalid_client":           "Client authentication failed",
    "invalid_grant":            "Invalid or expired authorization code",
    "unauthorized_client":      "Client not authorized for this operation",
    "unsupported_grant_type":   "Unsupported grant type",
    "invalid_scope":            "Invalid requested scope",
}
```

## Production Configuration

### Redis Storage

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
        Timeout(7200).  // 2 hours
        Build()
    
    stputil.SetManager(manager)
}
```

### Client Management

```go
// Load clients from database
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

## Monitoring and Auditing

### Log Authorization Events

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
    // Save to database
    db.Create(&event)
    
    // Anomaly detection
    if isAbnormalPattern(event) {
        alertSecurity(event)
    }
}
```

### Usage Statistics

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

## Advanced Features

### 1. Dynamic Scope Authorization

```go
// User can choose which scopes to grant
r.POST("/oauth/authorize", func(c *gin.Context) {
    requestedScopes := c.PostFormArray("scopes")  // ["read", "write", "profile"]
    grantedScopes := c.PostFormArray("granted")   // ["read", "write"]  // User only granted these
    
    authCode, _ := oauth2Server.GenerateAuthorizationCode(
        clientID, redirectURI, userID, grantedScopes,
    )
})
```

### 2. Token Revocation

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

### 3. Token Introspection

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

## FAQ

### Q: What's the difference between authorization code and access token?

A: 
- Authorization code: Temporary credential (10 minutes), only for exchanging token
- Access token: For accessing resources (2 hours)

### Q: How to implement third-party login (WeChat, GitHub)?

A: Use Sa-Token-Go as authorization server, your app as client, integrate with WeChat/GitHub OAuth2.

### Q: How to revoke all tokens?

A: Iterate through all user's refresh tokens and revoke them.

### Q: Does it support PKCE?

A: Current version supports standard authorization code flow, PKCE can be implemented at application layer.

## Performance Optimization

### 1. Authorization Code Caching

```
Single generation: ~200ns
Concurrent safe: ✅
Auto cleanup: Yes
```

### 2. Token Validation

```
Memory validation: ~50ns
Redis validation: ~1ms
```

## Next Steps

- [Nonce Anti-Replay](nonce.md)
- [Refresh Token](refresh-token.md)
- [Complete OAuth2 Example](../../examples/oauth2-example/)
- [Security Features Example](../../examples/security-features/)

