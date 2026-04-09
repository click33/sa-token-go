[English](refresh-token.md) | 中文文档

# Refresh Token 指南

## 当前状态

这个主题在当前项目里需要特别说明：

当前代码库**没有提供独立的“业务登录 Refresh Token 模块”**，也没有这些旧文档里提到的 API：

- `LoginWithRefreshToken(...)`
- `RefreshAccessToken(...)`
- `RevokeRefreshToken(...)`

这些接口在当前版本中并不存在。

## 现在项目里哪里有 Refresh Token

当前项目里的 Refresh Token 能力只出现在 **OAuth2** 模块里。

对应 API 是：

- `RefreshOAuth2AccessToken(...)`
- `RevokeOAuth2Token(...)`

也就是说，Refresh Token 目前是 OAuth2 访问令牌体系的一部分，而不是普通 `stputil.Login(...)` 登录体系的独立扩展。

## OAuth2 中的 Refresh Token

当你通过 OAuth2 授权码模式、密码模式、客户端凭证模式拿到 `AccessToken` 结构时，返回值里会包含：

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

其中：

- `Token`：访问令牌
- `RefreshToken`：刷新令牌

## 刷新方式

```go
newToken, err := stputil.RefreshOAuth2AccessToken(
    ctx,
    "web-app",
    oldToken.RefreshToken,
    "secret",
)
```

当前实现会：

1. 校验 refresh token
2. 校验客户端身份
3. 删除旧 access token
4. 删除旧 refresh token
5. 重新签发一组新的 token

## 撤销方式

```go
err := stputil.RevokeOAuth2Token(ctx, accessToken)
```

撤销 access token 时，会同时清掉它对应的 refresh token。

## 默认有效期

根据当前 OAuth2 常量：

- access token：`2` 小时
- refresh token：`30` 天

## 如果你想做普通登录态的双 token

当前项目没有内建“普通登录态双 token”方案。  
如果你要做 App / Web 的 access token + refresh token 登录机制，通常有两条路：

1. 直接使用项目现有 OAuth2 能力
2. 在业务层基于 `stputil.Login(...)` 自己再封装一层 refresh token 存储和刷新逻辑

## 当前文档建议

所以这篇文档的正确理解应该是：

1. 当前仓库没有独立 refresh-token 模块
2. 现有 refresh token 只属于 OAuth2
3. 相关示例和接入说明请优先参考 OAuth2 文档

## 相关文档

- [OAuth2 指南](oauth2_zh.md)
- [登录认证](authentication_zh.md)
- [Nonce 防重放](nonce_zh.md)
