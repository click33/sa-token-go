[English](README.md) | 中文文档

# OAuth2 授权码模式示例

完整的 OAuth2 授权码流程实现示例。

## 功能特性

- **授权码模式** - 标准的 OAuth2 授权码流程
- **令牌刷新** - 使用刷新令牌刷新访问令牌
- **令牌验证** - 验证访问令牌
- **令牌撤销** - 撤销访问令牌
- **多客户端** - 支持多个 OAuth2 客户端
- **权限管理** - 细粒度权限控制

## 快速开始

### 1. 运行服务器

```bash
cd examples/oauth2-example
go run main.go
```

服务器运行在 `http://localhost:8080`

### 2. OAuth2 流程

#### 步骤 1: 授权请求

```bash
curl "http://localhost:8080/oauth/authorize?client_id=webapp&redirect_uri=http://localhost:8080/callback&response_type=code&state=xyz123"
```

响应:
```json
{
  "message": "Authorization code generated",
  "code": "a3f5d8b2c1e4f6a9...",
  "redirect_url": "http://localhost:8080/callback?code=...&state=xyz123",
  "user_id": "user123",
  "scopes": ["read", "write"]
}
```

#### 步骤 2: 用授权码换取令牌

```bash
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=authorization_code" \
  -d "code=a3f5d8b2c1e4f6a9..." \
  -d "client_id=webapp" \
  -d "client_secret=secret123" \
  -d "redirect_uri=http://localhost:8080/callback"
```

响应:
```json
{
  "access_token": "b4f6d9c3d2e5f7b0...",
  "token_type": "Bearer",
  "expires_in": 7200,
  "refresh_token": "c5f7e0d4e3f6e8c1...",
  "scope": ["read", "write"]
}
```

#### 步骤 3: 使用访问令牌

```bash
curl http://localhost:8080/oauth/userinfo \
  -H "Authorization: Bearer b4f6d9c3d2e5f7b0..."
```

响应:
```json
{
  "user_id": "user123",
  "client_id": "webapp",
  "scopes": ["read", "write"],
  "expires_in": 7200,
  "issued_at": 1700000000
}
```

#### 步骤 4: 刷新访问令牌

```bash
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=refresh_token" \
  -d "refresh_token=c5f7e0d4e3f6e8c1..." \
  -d "client_id=webapp" \
  -d "client_secret=secret123"
```

#### 步骤 5: 撤销令牌

```bash
curl -X POST http://localhost:8080/oauth/revoke \
  -d "token=b4f6d9c3d2e5f7b0..."
```

## 已注册的客户端

### Web 应用

```
Client ID: webapp
Client Secret: secret123
回调 URI:
  - http://localhost:8080/callback
  - http://localhost:3000/callback
权限范围: read, write, profile
```

### 移动应用

```
Client ID: mobile-app
Client Secret: mobile-secret-456
回调 URI:
  - myapp://oauth/callback
权限范围: read, write
```

## API 端点

| 端点 | 方法 | 说明 |
|------|------|------|
| `/oauth/authorize` | GET | 授权端点 |
| `/oauth/token` | POST | 令牌端点 |
| `/oauth/userinfo` | GET | 用户信息端点 |
| `/oauth/revoke` | POST | 令牌撤销端点 |

## 授权请求参数

| 参数 | 必需 | 说明 |
|------|------|------|
| `client_id` | 是 | 客户端标识符 |
| `redirect_uri` | 是 | 回调 URI |
| `response_type` | 是 | 必须是 "code" |
| `state` | 推荐 | CSRF 保护 |
| `scope` | 可选 | 请求的权限范围 |

## 令牌请求参数

### 授权码模式

| 参数 | 必需 | 说明 |
|------|------|------|
| `grant_type` | 是 | "authorization_code" |
| `code` | 是 | 授权码 |
| `client_id` | 是 | 客户端标识符 |
| `client_secret` | 是 | 客户端密钥 |
| `redirect_uri` | 是 | 必须与授权请求匹配 |

### 刷新令牌模式

| 参数 | 必需 | 说明 |
|------|------|------|
| `grant_type` | 是 | "refresh_token" |
| `refresh_token` | 是 | 刷新令牌 |
| `client_id` | 是 | 客户端标识符 |
| `client_secret` | 是 | 客户端密钥 |

## 安全特性

1. **客户端认证** - 客户端密钥验证
2. **回调 URI 验证** - 防止开放重定向攻击
3. **State 参数** - CSRF 保护
4. **授权码过期** - 授权码 10 分钟后过期
5. **令牌过期** - 访问令牌 2 小时后过期
6. **一次性使用** - 授权码只能使用一次
7. **权限验证** - 请求的权限必须被允许

## 集成示例

```go
package main

import (
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/storage/memory"
)

func main() {
    storage := memory.NewStorage()
    oauth2Server := core.NewOAuth2Server(storage)

    // 注册客户端
    oauth2Server.RegisterClient(&core.OAuth2Client{
        ClientID:     "my-app",
        ClientSecret: "my-secret",
        RedirectURIs: []string{"http://localhost:3000/callback"},
        GrantTypes:   []core.OAuth2GrantType{core.GrantTypeAuthorizationCode},
        Scopes:       []string{"read", "write"},
    })

    // 生成授权码
    authCode, _ := oauth2Server.GenerateAuthorizationCode(
        "my-app",
        "http://localhost:3000/callback",
        "user123",
        []string{"read"},
    )

    // 用授权码换取令牌
    token, _ := oauth2Server.ExchangeCodeForToken(
        authCode.Code,
        "my-app",
        "my-secret",
        "http://localhost:3000/callback",
    )

    // 验证令牌
    validated, _ := oauth2Server.ValidateAccessToken(token.Token)
    
    // 刷新令牌
    newToken, _ := oauth2Server.RefreshAccessToken(
        token.RefreshToken,
        "my-app",
        "my-secret",
    )
}
```

## 下一步

- [安全特性示例](../security-features/)
- [刷新令牌指南](../../docs/guide/refresh-token_zh.md)
- [OAuth2 文档](../../docs/guide/oauth2_zh.md)

