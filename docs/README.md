English | [中文文档](README_zh.md)

# Sa-Token-Go Documentation Center

## 📚 Documentation Navigation

### 🚀 Quick Start

- [5-Minute Quick Start](tutorial/quick-start.md) - Fastest way to get started

### 📖 User Guides

- [Authentication](guide/authentication.md) - Login, logout, token management
- [Permission Verification](guide/permission.md) - Permission system, wildcard usage
- [Path-Based Auth](guide/path-auth.md) - Path auth and ant-style route matching
- [Annotations](guide/annotation.md) - Decorator pattern guide
- [Event Listener](guide/listener.md) - Event system usage guide
- [JWT Integration](guide/jwt.md) - JWT token configuration and usage
- [Redis Storage](guide/redis-storage.md) - Redis storage configuration guide

### 🔒 Security Features

- [Nonce Anti-Replay](guide/nonce.md) - Prevent replay attacks
- [Refresh Token](guide/refresh-token.md) - Token refresh mechanism
- [OAuth2 Authorization](guide/oauth2.md) - OAuth2 authorization code flow

### 🔧 API Documentation

- [StpUtil API](api/stputil.md) - Complete global utility API reference

### Engineering

- [Style variance & PR checklist](engineering/style-variance.md) - Semantic-safe refactor review

### 🏗️ Design Documentation

- [Architecture Design](design/architecture.md) - System architecture and data flow
- [Auto-Renewal Design](design/auto-renew.md) - Asynchronous renewal mechanism
- [Modular Design](design/modular.md) - Module organization strategy

## 📖 Example Projects

- [quick-start](../examples/quick-start/) - Quick start example
- [token-styles](../examples/token-styles/) - 9 token styles demonstration
- [security-features](../examples/security-features/) - Security features demo
- [oauth2-example](../examples/oauth2-example/) - Complete OAuth2 implementation
- [annotation](../examples/annotation/) - Annotation usage example
- [jwt-example](../examples/jwt-example/) - JWT usage example
- [redis-example](../examples/redis-example/) - Redis storage example
- [listener-example](../examples/listener-example/) - Event listener example
- [gin-simple](../examples/gin/gin-simple/) - Minimal gin integration example
- [gin-example](../examples/gin/gin-example/) - Gin integration with config file
- [echo-example](../examples/echo/echo-example/) - Echo integration example
- [fiber-example](../examples/fiber/fiber-example/) - Fiber integration example
- [chi-example](../examples/chi/chi-example/) - Chi integration example
- [gf-example](../examples/gf/) - GoFrame integration example
- [kratos-example](../examples/kratos/kratos-example/) - Kratos integration example
- [hertz-example](../examples/hertz/hertz-example/) - Hertz integration example

### 🔄 Integration Upgrade Notes

- All integration plugins now support `TokenInterceptor()`:
  - Unified token extraction order: `Header -> Cookie -> Query(apikey)`.
  - Respects `TokenPrefix` via `CutTokenPrefix`.
  - Falls back to `Authorization` header when `TokenName` is blank.
- All integration plugins provide `GetTokenFromCtx(...)`:
  - Read the parsed token from framework context instead of manually parsing headers in handlers.

## 🔗 External Resources

- [GitHub Repository](https://github.com/click33/sa-token-go)
- [Java sa-token](https://github.com/dromara/sa-token)

---

**Sa-Token-Go v0.2.1**
