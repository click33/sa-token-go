# Gin 简单示例 - 只需导入一个包

此示例演示如何通过 **只导入 `integrations/gin` 包** 来使用 Sa-Token-Go 与 Gin。

## 特性

✅ **单一导入** - 只需要 `github.com/click33/sa-token-go/integrations/gin`  
✅ **完整功能** - 访问所有 core 和 stputil 的功能  
✅ **简洁 API** - 干净易用  

## 快速开始

### 1. 安装依赖

```bash
go get github.com/click33/sa-token-go/integrations/gin@v0.1.0
go get github.com/click33/sa-token-go/storage/memory@v0.1.0
go get github.com/gin-gonic/gin
```

### 2. 运行示例

```bash
cd examples/gin/gin-simple
go run main.go
```

### 3. 测试 API

**登录：**
```bash
curl -X POST http://localhost:8080/login -d 'user_id=1000'
# 响应: {"message":"登录成功","token":"xxx"}
```

**检查登录状态：**
```bash
curl -H "token: YOUR_TOKEN" http://localhost:8080/check
# 响应: {"login_id":"1000","message":"已登录"}
```

**访问受保护的 API：**
```bash
curl -H "token: YOUR_TOKEN" http://localhost:8080/api/user
# 响应: {"name":"User 1000","user_id":"1000"}
```

**登出：**
```bash
curl -X POST -H "token: YOUR_TOKEN" http://localhost:8080/logout
# 响应: {"message":"登出成功"}
```

**踢人下线：**
```bash
curl -X POST -H "token: YOUR_TOKEN" http://localhost:8080/api/kickout/1000
# 响应: {"message":"踢人成功"}
```

## 代码亮点

### 旧方式（多个导入）

```go
import (
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/integrations/gin"
)

config := core.DefaultConfig()
manager := core.NewManager(storage, config)
stputil.SetManager(manager)
token, _ := stputil.Login(userID)
```

### 新方式（单一导入）✨

```go
import (
    sagin "github.com/click33/sa-token-go/integrations/gin"
)

config := sagin.DefaultConfig()
manager := sagin.NewManager(storage, config)
sagin.SetManager(manager)
token, _ := sagin.Login(userID)
```

## 可用函数

所有 `core` 和 `stputil` 的函数都在 `sagin` 中重新导出：

### 认证相关
- `sagin.Login(loginID, device...)`
- `sagin.Logout(loginID, device...)`
- `sagin.IsLogin(token)`
- `sagin.CheckLogin(token)`
- `sagin.GetLoginID(token)`

### 踢人下线 & 封禁
- `sagin.Kickout(loginID, device...)`
- `sagin.Disable(loginID, duration)`
- `sagin.IsDisable(loginID)`
- `sagin.Untie(loginID)`

### 权限 & 角色
- `sagin.CheckPermission(loginID, permission)`
- `sagin.CheckRole(loginID, role)`
- `sagin.HasPermission(loginID, permission)`
- `sagin.HasRole(loginID, role)`

### Session 管理
- `sagin.GetSession(loginID)`
- `sagin.GetSessionByToken(token)`

### 安全特性
- `sagin.GenerateNonce()`
- `sagin.VerifyNonce(nonce)`
- `sagin.LoginWithRefreshToken(loginID, device...)`
- `sagin.RefreshAccessToken(refreshToken)`
- `sagin.GetOAuth2Server()`

### Builder & Config
- `sagin.DefaultConfig()`
- `sagin.NewManager(storage, config)`
- `sagin.NewBuilder()`
- `sagin.SetManager(manager)`

## 优势

1. **更简单的依赖** - 只需要一个导入
2. **更清晰的代码** - 更少的导入语句
3. **框架专用** - 为 Gin 优化
4. **向后兼容** - 旧方式仍然有效

## 了解更多

- [主文档](../../../README_zh.md)
- [其他示例](../../)
- [API 参考](../../../docs/api/api_zh.md)

