English | [中文文档](oauth2_zh.md)

# OAuth2 Guide

## Overview

The current project includes a lightweight OAuth2 authorization server implementation under:

- `core/oauth2`
- `core/manager/manager_oauth2_func.go`
- the OAuth2 wrapper functions in `stputil`

## Supported Grant Types

The current implementation supports 4 grant types:

- `authorization_code`
- `refresh_token`
- `client_credentials`
- `password`

Matching constants:

```go
oauth2.GrantTypeAuthorizationCode
oauth2.GrantTypeRefreshToken
oauth2.GrantTypeClientCredentials
oauth2.GrantTypePassword
```

## Default Expiration

According to the current code:

- authorization code: `10` minutes
- access token: `2` hours
- refresh token: `30` days

## Client Model

```go
client := &oauth2.Client{
    ClientID:     "web-app",
    ClientSecret: "secret",
    RedirectURIs: []string{"http://localhost:3000/callback"},
    GrantTypes: []oauth2.GrantType{
        oauth2.GrantTypeAuthorizationCode,
        oauth2.GrantTypeRefreshToken,
    },
    Scopes: []string{"read", "write", "profile"},
}
```

## Register Client

```go
err := stputil.RegisterOAuth2Client(&oauth2.Client{
    ClientID:     "web-app",
    ClientSecret: "secret",
    RedirectURIs: []string{"http://localhost:3000/callback"},
    GrantTypes: []oauth2.GrantType{
        oauth2.GrantTypeAuthorizationCode,
        oauth2.GrantTypeRefreshToken,
    },
    Scopes: []string{"read", "write", "profile"},
})
```

## Unified Token Entry

`stputil` provides a unified token entry:

```go
token, err := stputil.OAuth2Token(ctx, &oauth2.TokenRequest{
    GrantType:    oauth2.GrantTypeClientCredentials,
    ClientID:     "server-app",
    ClientSecret: "secret",
    Scopes:       []string{"read"},
}, nil)
```

## Authorization Code Flow

### Generate Authorization Code

```go
authCode, err := stputil.GenerateOAuth2AuthorizationCode(
    ctx,
    "web-app",
    "10001",
    "http://localhost:3000/callback",
    []string{"read", "profile"},
)
```

This validates:

1. whether the `clientID` exists
2. whether the `redirectURI` is whitelisted
3. whether scopes are allowed
4. whether `userID` is empty

### Exchange For Token

```go
token, err := stputil.ExchangeOAuth2CodeForToken(
    ctx,
    authCode.Code,
    "web-app",
    "secret",
    "http://localhost:3000/callback",
)
```

Returned `AccessToken`:

```go
type AccessToken struct {
    Token        string
    TokenType    string
    ExpiresIn    int64
    RefreshToken string
    Scopes       []string
    UserID       string
    ClientID     string
}
```

## Client Credentials Flow

```go
token, err := stputil.OAuth2ClientCredentialsToken(
    ctx,
    "server-app",
    "secret",
    []string{"read"},
)
```

This is meant for service-to-service calls without a user context.

## Password Flow

```go
validateUser := func(username, password string) (string, error) {
    if username == "admin" && password == "123456" {
        return "10001", nil
    }
    return "", fmt.Errorf("invalid credentials")
}

token, err := stputil.OAuth2PasswordGrantToken(
    ctx,
    "native-app",
    "secret",
    "admin",
    "123456",
    []string{"read", "write"},
    validateUser,
)
```

The current implementation requires `validateUser`; otherwise it returns an error.

## Refresh Access Token

```go
newToken, err := stputil.RefreshOAuth2AccessToken(
    ctx,
    "web-app",
    token.RefreshToken,
    "secret",
)
```

The current implementation:

1. validates the refresh token
2. validates `clientID` and `clientSecret`
3. deletes the old access token
4. deletes the old refresh token
5. issues a brand new token pair

## Validate Access Token

```go
valid := stputil.ValidateOAuth2AccessToken(ctx, token.Token)

info, err := stputil.ValidateOAuth2AccessTokenAndGetInfo(ctx, token.Token)
```

## Revoke Token

```go
err := stputil.RevokeOAuth2Token(ctx, token.Token)
```

Revocation clears both:

- the access token
- its matching refresh token

## Scope Validation

The current implementation checks whether requested scopes belong to the client allowlist before issuing a token.  
If the client `Scopes` field is empty, scope restriction is treated as open.

## Recommended Integration Model

### Authorization Endpoint

1. the user logs into your own system first
2. your system generates an authorization code
3. your system redirects back to the client callback URL

### Token Endpoint

1. the client sends `grant_type`
2. the server calls `OAuth2Token` or one of the grant-specific helpers
3. the server returns an `AccessToken` payload

### Resource Endpoint

1. extract the token from `Authorization: Bearer xxx`
2. call `ValidateOAuth2AccessToken` or `ValidateOAuth2AccessTokenAndGetInfo`
3. apply resource authorization based on `Scopes`

## Current Boundaries

The OAuth2 implementation is already usable for common authorization flows, but there are a few important limits:

1. there is no `GetOAuth2Server()` public entry in the current API
2. PKCE is not built in and should be implemented at the application layer
3. no ready-made HTTP introspection endpoint is provided; you write the route yourself

## Related Documentation

- [Nonce Anti-Replay](nonce.md)
- [Refresh Token Guide](refresh-token.md)
- [Authentication Guide](authentication.md)
