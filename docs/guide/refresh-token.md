English | [中文文档](refresh-token_zh.md)

# Refresh Token Guide

## Current Status

This topic needs a very explicit clarification in the current codebase:

The project does **not** provide a standalone “business login refresh token module”, and it does not contain these old APIs:

- `LoginWithRefreshToken(...)`
- `RefreshAccessToken(...)`
- `RevokeRefreshToken(...)`

Those APIs do not exist in the current version.

## Where Refresh Token Exists Today

Refresh token support currently exists only inside the **OAuth2** module.

Relevant APIs:

- `RefreshOAuth2AccessToken(...)`
- `RevokeOAuth2Token(...)`

So refresh tokens are currently part of the OAuth2 access-token system, not a separate extension of the normal `stputil.Login(...)` login flow.

## Refresh Token In OAuth2

When you obtain an OAuth2 token through authorization code, password, or client credentials related flows, the returned `AccessToken` structure includes:

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

## Refresh Flow

```go
newToken, err := stputil.RefreshOAuth2AccessToken(
    ctx,
    "web-app",
    oldToken.RefreshToken,
    "secret",
)
```

The current implementation:

1. validates the refresh token
2. validates the client identity
3. deletes the old access token
4. deletes the old refresh token
5. issues a fresh token pair

## Revoke Flow

```go
err := stputil.RevokeOAuth2Token(ctx, accessToken)
```

Revoking an access token also clears its related refresh token.

## Default TTL

According to the current OAuth2 constants:

- access token: `2` hours
- refresh token: `30` days

## If You Need Dual-Token Login Outside OAuth2

The project does not currently ship a built-in “normal login dual-token” mechanism.  
If you want an app/web access-token + refresh-token login model, you generally have two choices:

1. use the existing OAuth2 capabilities directly
2. build your own refresh-token storage and rotation logic on top of `stputil.Login(...)`

## Practical Conclusion

The correct interpretation of this guide in the current repository is:

1. there is no standalone refresh-token module
2. existing refresh-token support belongs only to OAuth2
3. for implementation details and examples, the OAuth2 guide is the primary reference

## Related Documentation

- [OAuth2 Guide](oauth2.md)
- [Authentication Guide](authentication.md)
- [Nonce Anti-Replay](nonce.md)
