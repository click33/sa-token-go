# Hertz SaToken Example

这是一个基于当前 `SaToken` 项目的 `Hertz` 示例，结构和 `gf/gin` 示例保持一致，包含登录、登录态校验、角色校验、权限校验和注解式校验。

## 目录说明

```text
examples/hertz/
├── main.go
├── README.md
└── go.mod
```

## 当前使用的包

- `github.com/click33/sa-token-go/integrations/hertz`
- `github.com/click33/sa-token-go/com/storage/redis`
- `github.com/cloudwego/hertz`

## 启动方式

### 1. 进入目录

```bash
cd examples/hertz
```

### 2. 配置 Redis

请先确认 Redis 可用，并按实际环境修改 `main.go` 里的连接地址。

### 3. 运行示例

```bash
go run main.go
```

默认访问地址：

```text
http://localhost:8080
```

## 示例接口

- `POST /api/login`
- `GET /api/public`
- `GET /api/user/info`
- `POST /api/user/logout`
- `GET /api/admin/users`
- `POST /api/admin/disable`
- `POST /api/admin/enable`
- `GET /api/resource/list`
- `GET /api/annotation/profile`
- `GET /api/annotation/admin-data`
- `GET /api/annotation/sensitive`
- `GET /api/annotation/super`

## 说明

- `POST /api/login` 默认账号密码为 `admin / 123456`
- 登录成功后会初始化示例角色和权限数据
- 示例里演示了 `AuthMiddleware`、`RoleMiddleware`、`PermissionMiddleware` 和 `CheckAllMiddleware`
- 当前版本仍然保持单文件入口，方便快速体验 `Hertz + SaToken`
