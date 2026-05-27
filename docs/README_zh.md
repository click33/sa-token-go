[English](README.md) | 中文文档

# Sa-Token-Go 文档中心

## 📚 文档导航

### 🚀 快速上手

- [5分钟快速开始](tutorial/quick-start_zh.md) - 最快上手方式

### 📖 使用指南

- [登录认证](guide/authentication_zh.md) - 登录、登出、Token管理
- [权限验证](guide/permission_zh.md) - 权限系统详解、通配符使用
- [路径鉴权](guide/path-auth_zh.md) - 路径鉴权与 Ant 风格匹配
- [注解使用](guide/annotation_zh.md) - 装饰器模式详解
- [事件监听](guide/listener_zh.md) - 事件系统使用指南
- [JWT集成](guide/jwt_zh.md) - JWT Token配置和使用
- [Redis存储](guide/redis-storage_zh.md) - Redis存储配置详解

### 🔒 安全特性

- [Nonce防重放](guide/nonce_zh.md) - 防止请求重放攻击
- [Refresh Token](guide/refresh-token_zh.md) - 刷新令牌机制
- [OAuth2授权](guide/oauth2_zh.md) - OAuth2授权码模式

### 🔧 API文档

- [StpUtil API](api/stputil_zh.md) - 全局工具类完整API

### 工程维护

- [风格改造 PR 审查清单](engineering/style-variance.md) - 语义不变类改动的合并门槛

### 🏗️ 设计文档

- [架构设计](design/architecture_zh.md) - 系统架构、数据流转
- [自动续签设计](design/auto-renew_zh.md) - 异步续签原理
- [模块化设计](design/modular_zh.md) - 模块划分策略

## 📖 示例项目

- [quick-start](../examples/quick-start/) - 快速开始示例
- [token-styles](../examples/token-styles/) - 9种Token风格演示
- [security-features](../examples/security-features/) - 安全特性综合示例
- [oauth2-example](../examples/oauth2-example/) - OAuth2完整实现
- [annotation](../examples/annotation/) - 注解使用示例
- [jwt-example](../examples/jwt-example/) - JWT使用示例
- [redis-example](../examples/redis-example/) - Redis存储示例
- [listener-example](../examples/listener-example/) - 事件监听示例
- [gin-simple](../examples/gin/gin-simple/) - Gin 最简集成示例
- [gin-example](../examples/gin/gin-example/) - Gin 配置化集成示例
- [echo-example](../examples/echo/echo-example/) - Echo 集成示例
- [fiber-example](../examples/fiber/fiber-example/) - Fiber 集成示例
- [chi-example](../examples/chi/chi-example/) - Chi 集成示例
- [gf-example](../examples/gf/) - GoFrame 集成示例
- [kratos-example](../examples/kratos/kratos-example/) - Kratos 集成示例
- [hertz-example](../examples/hertz/hertz-example/) - Hertz 集成示例

### 🔄 集成能力更新说明

- 所有集成插件已支持 `TokenInterceptor()`：
  - Token 统一提取顺序：`Header -> Cookie -> Query(apikey)`。
  - 自动走 `CutTokenPrefix`，支持 `TokenPrefix`。
  - 当 `TokenName` 为空白时，自动回退读取 `Authorization`。
- 所有集成插件均提供 `GetTokenFromCtx(...)`：
  - 业务处理器中可直接读取拦截器解析后的 token，避免重复解析 Header/Cookie。

## 🔗 外部资源

- [GitHub仓库](https://github.com/click33/sa-token-go)
- [Java版sa-token](https://github.com/dromara/sa-token)

---

**Sa-Token-Go v0.1.0**

