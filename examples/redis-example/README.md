# Redis Storage Example

[中文说明](README_zh.md) | English

This example demonstrates how to use Redis as the storage backend for Sa-Token-Go.

## Prerequisites

- Redis server running on `localhost:6379` (or set `REDIS_ADDR` environment variable)
- Go 1.21 or higher

## Install Redis

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

## Run Example

```bash
# Without password
go run main.go

# With password
REDIS_PASSWORD=your-password go run main.go

# Custom Redis address
REDIS_ADDR=redis.example.com:6379 go run main.go
```

## Key Features Demonstrated

1. ✅ **Redis Connection** - Connect to Redis with go-redis
2. ✅ **Authentication** - Login/Logout with Redis storage
3. ✅ **Permission Management** - Store permissions in Redis
4. ✅ **Role Management** - Store roles in Redis
5. ✅ **Session Management** - Persistent session data
6. ✅ **Data Persistence** - Data survives application restarts

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_ADDR` | Redis server address | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password | (empty) |
| `REDIS_DB` | Redis database number | `0` |

## View Data in Redis

```bash
# Connect to Redis CLI
redis-cli

# List all Sa-Token keys
KEYS satoken:*

# View token info
GET satoken:login:token:{your-token}

# View session data
GET satoken:session:1000

# View permissions
SMEMBERS satoken:permission:1000

# View roles
SMEMBERS satoken:role:1000
```

## Production Deployment

See [Redis Storage Guide](../../docs/guide/redis-storage.md) for:
- Connection pool configuration
- High availability (Sentinel)
- Cluster mode
- TLS/SSL support
- Docker/Kubernetes deployment

## Related Documentation

- [Redis Storage Guide](../../docs/guide/redis-storage.md)
- [Quick Start](../../docs/tutorial/quick-start.md)
- [Authentication Guide](../../docs/guide/authentication.md)

