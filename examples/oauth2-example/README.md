English | [中文文档](README_zh.md)

# OAuth2 Authorization Code Flow Example

Complete OAuth2 authorization code flow implementation example.

## Features

- **Authorization Code Grant** - Standard OAuth2 authorization code flow
- **Token Refresh** - Refresh access tokens using refresh tokens
- **Token Validation** - Validate access tokens
- **Token Revocation** - Revoke access tokens
- **Multiple Clients** - Support for multiple OAuth2 clients
- **Scope Management** - Fine-grained permission control

## Quick Start

### 1. Run the Server

```bash
cd examples/oauth2-example
go run main.go
```

Server runs on `http://localhost:8080`

### 2. OAuth2 Flow

#### Step 1: Authorization Request

```bash
curl "http://localhost:8080/oauth/authorize?client_id=webapp&redirect_uri=http://localhost:8080/callback&response_type=code&state=xyz123"
```

Response:
```json
{
  "message": "Authorization code generated",
  "code": "a3f5d8b2c1e4f6a9...",
  "redirect_url": "http://localhost:8080/callback?code=...&state=xyz123",
  "user_id": "user123",
  "scopes": ["read", "write"]
}
```

#### Step 2: Exchange Code for Token

```bash
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=authorization_code" \
  -d "code=a3f5d8b2c1e4f6a9..." \
  -d "client_id=webapp" \
  -d "client_secret=secret123" \
  -d "redirect_uri=http://localhost:8080/callback"
```

Response:
```json
{
  "access_token": "b4f6d9c3d2e5f7b0...",
  "token_type": "Bearer",
  "expires_in": 7200,
  "refresh_token": "c5f7e0d4e3f6e8c1...",
  "scope": ["read", "write"]
}
```

#### Step 3: Use Access Token

```bash
curl http://localhost:8080/oauth/userinfo \
  -H "Authorization: Bearer b4f6d9c3d2e5f7b0..."
```

Response:
```json
{
  "user_id": "user123",
  "client_id": "webapp",
  "scopes": ["read", "write"],
  "expires_in": 7200,
  "issued_at": 1700000000
}
```

#### Step 4: Refresh Access Token

```bash
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=refresh_token" \
  -d "refresh_token=c5f7e0d4e3f6e8c1..." \
  -d "client_id=webapp" \
  -d "client_secret=secret123"
```

#### Step 5: Revoke Token

```bash
curl -X POST http://localhost:8080/oauth/revoke \
  -d "token=b4f6d9c3d2e5f7b0..."
```

## Registered Clients

### Web Application

```
Client ID: webapp
Client Secret: secret123
Redirect URIs:
  - http://localhost:8080/callback
  - http://localhost:3000/callback
Scopes: read, write, profile
```

### Mobile Application

```
Client ID: mobile-app
Client Secret: mobile-secret-456
Redirect URIs:
  - myapp://oauth/callback
Scopes: read, write
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/oauth/authorize` | GET | Authorization endpoint |
| `/oauth/token` | POST | Token endpoint |
| `/oauth/userinfo` | GET | User info endpoint |
| `/oauth/revoke` | POST | Token revocation endpoint |

## Authorization Request Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `client_id` | Yes | Client identifier |
| `redirect_uri` | Yes | Callback URI |
| `response_type` | Yes | Must be "code" |
| `state` | Recommended | CSRF protection |
| `scope` | Optional | Requested scopes |

## Token Request Parameters

### Authorization Code Grant

| Parameter | Required | Description |
|-----------|----------|-------------|
| `grant_type` | Yes | "authorization_code" |
| `code` | Yes | Authorization code |
| `client_id` | Yes | Client identifier |
| `client_secret` | Yes | Client secret |
| `redirect_uri` | Yes | Must match authorization request |

### Refresh Token Grant

| Parameter | Required | Description |
|-----------|----------|-------------|
| `grant_type` | Yes | "refresh_token" |
| `refresh_token` | Yes | Refresh token |
| `client_id` | Yes | Client identifier |
| `client_secret` | Yes | Client secret |

## Security Features

1. **Client Authentication** - Client secret verification
2. **Redirect URI Validation** - Prevent open redirect attacks
3. **State Parameter** - CSRF protection
4. **Code Expiration** - Authorization codes expire in 10 minutes
5. **Token Expiration** - Access tokens expire in 2 hours
6. **One-time Use** - Authorization codes can only be used once
7. **Scope Validation** - Requested scopes must be allowed

## Integration Example

```go
package main

import (
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/storage/memory"
)

func main() {
    storage := memory.NewStorage()
    oauth2Server := core.NewOAuth2Server(storage)

    // Register client
    oauth2Server.RegisterClient(&core.OAuth2Client{
        ClientID:     "my-app",
        ClientSecret: "my-secret",
        RedirectURIs: []string{"http://localhost:3000/callback"},
        GrantTypes:   []core.OAuth2GrantType{core.GrantTypeAuthorizationCode},
        Scopes:       []string{"read", "write"},
    })

    // Generate authorization code
    authCode, _ := oauth2Server.GenerateAuthorizationCode(
        "my-app",
        "http://localhost:3000/callback",
        "user123",
        []string{"read"},
    )

    // Exchange code for token
    token, _ := oauth2Server.ExchangeCodeForToken(
        authCode.Code,
        "my-app",
        "my-secret",
        "http://localhost:3000/callback",
    )

    // Validate token
    validated, _ := oauth2Server.ValidateAccessToken(token.Token)
    
    // Refresh token
    newToken, _ := oauth2Server.RefreshAccessToken(
        token.RefreshToken,
        "my-app",
        "my-secret",
    )
}
```

## Next Steps

- [Security Features Example](../security-features/)
- [Refresh Token Guide](../../docs/guide/refresh-token.md)
- [OAuth2 Documentation](../../docs/guide/oauth2.md)

