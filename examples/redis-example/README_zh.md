# Redis 存储示例

[English](README.md) | 中文说明

本示例演示如何使用 Redis 作为 Sa-Token-Go 的存储后端。

## 前置要求

- Redis 服务器运行在 `localhost:6379`（或设置 `REDIS_ADDR` 环境变量）
- Go 1.21 或更高版本

## 安装 Redis

### macOS

```bash
brew install redis
brew services start redis
```

### Linux (Ubuntu/Debian)

```bash
sudo apt-get install redis-server
sudo systemctl start redis
```

### Docker

```bash
docker run -d -p 6379:6379 redis:7-alpine
```

## 运行示例

```bash
# 无密码
go run main.go

# 带密码
REDIS_PASSWORD=your-password go run main.go

# 自定义 Redis 地址
REDIS_ADDR=redis.example.com:6379 go run main.go
```

## 演示的核心功能

1. ✅ **Redis 连接** - 使用 go-redis 连接 Redis
2. ✅ **认证功能** - 使用 Redis 存储进行登录/登出
3. ✅ **权限管理** - 在 Redis 中存储权限
4. ✅ **角色管理** - 在 Redis 中存储角色
5. ✅ **Session 管理** - 持久化的 Session 数据
6. ✅ **数据持久化** - 数据在应用重启后仍然存在

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `REDIS_ADDR` | Redis 服务器地址 | `localhost:6379` |
| `REDIS_PASSWORD` | Redis 密码 | （空） |
| `REDIS_DB` | Redis 数据库编号 | `0` |

## 在 Redis 中查看数据

```bash
# 连接到 Redis CLI
redis-cli

# 列出所有 Sa-Token 键
KEYS satoken:*

# 查看 Token 信息
GET satoken:login:token:{your-token}

# 查看 Session 数据
GET satoken:session:1000

# 查看权限
SMEMBERS satoken:permission:1000

# 查看角色
SMEMBERS satoken:role:1000
```

## 生产环境部署

查看 [Redis 存储指南](../../docs/guide/redis-storage_zh.md) 了解：

- 连接池配置
- 高可用（哨兵模式）
- 集群模式
- TLS/SSL 支持
- Docker/Kubernetes 部署

## 相关文档

- [Redis 存储指南](../../docs/guide/redis-storage_zh.md)
- [快速开始](../../docs/tutorial/quick-start.md)
- [认证指南](../../docs/guide/authentication.md)
