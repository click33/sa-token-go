[English](oauth2.md) | 中文文档

# OAuth2 指南

## 概览

当前项目内置了一个轻量级 OAuth2 服务端实现，核心代码位于：

- `core/oauth2`
- `core/manager/manager_oauth2_func.go`
- `stputil` 中对应的 OAuth2 包装函数

## 当前支持的授权类型

当前实现支持 4 种 `GrantType`：

- `authorization_code`
- `refresh_token`
- `client_credentials`
- `password`

对应常量：

```go
oauth2.GrantTypeAuthorizationCode
oauth2.GrantTypeRefreshToken
oauth2.GrantTypeClientCredentials
oauth2.GrantTypePassword
```

## 默认有效期

根据当前代码：

- 授权码：`10` 分钟
- Access Token：`2` 小时
- Refresh Token：`30` 天

## 客户端模型

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

## 注册客户端

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

## 统一令牌入口

当前 `stputil` 提供统一入口：

```go
token, err := stputil.OAuth2Token(ctx, &oauth2.TokenRequest{
    GrantType:    oauth2.GrantTypeClientCredentials,
    ClientID:     "server-app",
    ClientSecret: "secret",
    Scopes:       []string{"read"},
}, nil)
```

## 授权码模式

### 生成授权码

```go
authCode, err := stputil.GenerateOAuth2AuthorizationCode(
    ctx,
    "web-app",
    "10001",
    "http://localhost:3000/callback",
    []string{"read", "profile"},
)
```

这里会校验：

1. `clientID` 是否存在
2. `redirectURI` 是否在白名单中
3. scope 是否合法
4. `userID` 是否为空

### 交换 Token

```go
token, err := stputil.ExchangeOAuth2CodeForToken(
    ctx,
    authCode.Code,
    "web-app",
    "secret",
    "http://localhost:3000/callback",
)
```

返回的 `AccessToken` 结构包括：

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

## 客户端凭证模式

```go
token, err := stputil.OAuth2ClientCredentialsToken(
    ctx,
    "server-app",
    "secret",
    []string{"read"},
)
```

适合服务与服务之间调用，没有用户参与。

## 密码模式

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

当前实现要求必须传入 `validateUser`，否则会返回错误。

## 刷新 Access Token

```go
newToken, err := stputil.RefreshOAuth2AccessToken(
    ctx,
    "web-app",
    token.RefreshToken,
    "secret",
)
```

当前实现会：

1. 校验 refresh token
2. 校验 `clientID` / `clientSecret`
3. 删除旧 access token
4. 删除旧 refresh token
5. 重新签发一组新的 token

## 校验 Access Token

```go
valid := stputil.ValidateOAuth2AccessToken(ctx, token.Token)

info, err := stputil.ValidateOAuth2AccessTokenAndGetInfo(ctx, token.Token)
```

## 撤销 Token

```go
err := stputil.RevokeOAuth2Token(ctx, token.Token)
```

撤销时会同时清理：

- access token
- 对应 refresh token

## Scope 校验

当前实现会在签发前校验 scope 是否属于客户端允许范围。  
如果客户端 `Scopes` 为空，则视为不限制。

```go
client := &oauth2.Client{
    ClientID:     "web-app",
    ClientSecret: "secret",
    RedirectURIs: []string{"http://localhost:3000/callback"},
    GrantTypes: []oauth2.GrantType{
        oauth2.GrantTypeAuthorizationCode,
    },
    Scopes: []string{"read", "write", "profile"},
}
```

## 推荐接入方式

### 授权端

1. 用户先登录你自己的系统
2. 你的系统生成授权码
3. 再把授权码带回 client 的回调地址

### Token 端

1. client 带 `grant_type`
2. server 调用 `OAuth2Token` 或各模式专用函数
3. 返回 `AccessToken` 结构

### 资源端

1. 从 `Authorization: Bearer xxx` 中取 token
2. 调用 `ValidateOAuth2AccessToken` 或 `ValidateOAuth2AccessTokenAndGetInfo`
3. 再按 `Scopes` 做资源权限判断

## 当前边界

当前 OAuth2 实现已经能覆盖常见授权场景，但也要注意：

1. 文档层不要再写不存在的 `GetOAuth2Server()` 这类入口
2. PKCE 目前没有内建，需要在应用层自己补
3. Token introspection 端点没有内建 HTTP 封装，需要你自己写路由

## 相关文档

- [Nonce 防重放](nonce_zh.md)
- [Refresh Token 指南](refresh-token_zh.md)
- [登录认证](authentication_zh.md)
