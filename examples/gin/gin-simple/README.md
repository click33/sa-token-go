# Gin Simple Example - Only Import One Package

This example demonstrates how to use Sa-Token-Go with Gin by **only importing the `integrations/gin` package**.

## Features

✅ **Single Import** - Only need `github.com/click33/sa-token-go/integrations/gin`  
✅ **All Functions** - Access to all core and stputil functions  
✅ **Simple API** - Clean and easy to use  

## Quick Start

### 1. Install dependencies

```bash
go get github.com/click33/sa-token-go/integrations/gin@v0.1.0
go get github.com/click33/sa-token-go/storage/memory@v0.1.0
go get github.com/gin-gonic/gin
```

### 2. Run the example

```bash
cd examples/gin/gin-simple
go run main.go
```

### 3. Test the API

**Login:**
```bash
curl -X POST http://localhost:8080/login -d 'user_id=1000'
# Response: {"message":"登录成功","token":"xxx"}
```

**Check Login Status:**
```bash
curl -H "token: YOUR_TOKEN" http://localhost:8080/check
# Response: {"login_id":"1000","message":"已登录"}
```

**Access Protected API:**
```bash
curl -H "token: YOUR_TOKEN" http://localhost:8080/api/user
# Response: {"name":"User 1000","user_id":"1000"}
```

**Logout:**
```bash
curl -X POST -H "token: YOUR_TOKEN" http://localhost:8080/logout
# Response: {"message":"登出成功"}
```

**Kickout User:**
```bash
curl -X POST -H "token: YOUR_TOKEN" http://localhost:8080/api/kickout/1000
# Response: {"message":"踢人成功"}
```

## Code Highlights

### Old Way (Multiple Imports)

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

### New Way (Single Import) ✨

```go
import (
    sagin "github.com/click33/sa-token-go/integrations/gin"
)

config := sagin.DefaultConfig()
manager := sagin.NewManager(storage, config)
sagin.SetManager(manager)
token, _ := sagin.Login(userID)
```

## Available Functions

All functions from `core` and `stputil` are re-exported in `sagin`:

### Authentication
- `sagin.Login(loginID, device...)`
- `sagin.Logout(loginID, device...)`
- `sagin.IsLogin(token)`
- `sagin.CheckLogin(token)`
- `sagin.GetLoginID(token)`

### Kickout & Disable
- `sagin.Kickout(loginID, device...)`
- `sagin.Disable(loginID, duration)`
- `sagin.IsDisable(loginID)`
- `sagin.Untie(loginID)`

### Permission & Role
- `sagin.CheckPermission(loginID, permission)`
- `sagin.CheckRole(loginID, role)`
- `sagin.HasPermission(loginID, permission)`
- `sagin.HasRole(loginID, role)`

### Session
- `sagin.GetSession(loginID)`
- `sagin.GetSessionByToken(token)`

### Security Features
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

## Benefits

1. **Simpler Dependencies** - Only one import needed
2. **Cleaner Code** - Less import statements
3. **Framework-Specific** - Optimized for Gin
4. **Backward Compatible** - Old way still works

## Learn More

- [Main Documentation](../../../README.md)
- [Other Examples](../../)
- [API Reference](../../../docs/api/api.md)

