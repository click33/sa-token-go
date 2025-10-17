[English](README.md) | 中文文档

# Token 风格示例

本示例演示 Sa-Token-Go 中所有可用的 Token 生成风格。

## 可用的 Token 风格

### 1. UUID 风格 (`uuid`)
```
例如：550e8400-e29b-41d4-a716-446655440000
```
- 标准 UUID v4 格式
- 36 个字符（包含连字符）
- 全局唯一

### 2. 简单风格 (`simple`)
```
例如：aB3dE5fG7hI9jK1l
```
- 16 字符随机字符串
- Base64 URL 安全编码
- 紧凑简单

### 3. Random32 风格 (`random32`)
```
例如：aB3dE5fG7hI9jK1lMnO2pQ4rS6tU8vW0
```
- 32 字符随机字符串
- 高随机性
- 安全且唯一

### 4. Random64 风格 (`random64`)
```
例如：aB3dE5fG7hI9jK1lMnO2pQ4rS6tU8vW0xY1zA2bC3dD4eE5fF6gG7hH8iI9jJ0kK1l
```
- 64 字符随机字符串
- 最大随机性
- 超级安全

### 5. Random128 风格 (`random128`)
```
例如：aB3dE5fG7hI9jK1lMnO2pQ4rS6tU8vW0xY1zA2bC3dD4eE5fF6gG7hH8iI9jJ0kK1lMmN2nO3oP4pQ5qR6rS7sT8tU9uV0vW1wX2xY3yZ4zA5aB6bC7cD8dE9eF0fG1gH2hI3iJ4jK5kL6lM7mN8nO9oP0
```
- 128 字符随机字符串
- 极度安全
- 用于高安全性场景

### 6. JWT 风格 (`jwt`)
```
例如：eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkZXZpY2UiOiJkZWZhdWx0IiwiaWF0IjoxNzAwMDAwMDAwLCJsb2dpbklkIjoidXNlcjEwMDAifQ.xxx
```
- 标准 JWT 格式
- 包含声明（loginId, device, iat, exp）
- 自包含且可验证
- 需要配置 `JwtSecretKey`

### 7. 哈希风格 (`hash`) 🆕
```
例如：a3f5d8b2c1e4f6a9d7b8c5e2f1a4d6b9c8e5f2a7d4b1c9e6f3a8d5b2c1e7f4a6
```
- 基于 SHA256 哈希的 Token
- 组合 loginID、device、时间戳和随机数据
- 64 字符十六进制
- 高安全性和不可预测性

### 8. 时间戳风格 (`timestamp`) 🆕
```
例如：1700000000123_user1000_a3f5d8b2c1e4f6a9
```
- 格式：`时间戳_loginID_随机数`
- 毫秒精度时间戳
- 易于追溯创建时间
- 便于调试和日志记录

### 9. Tik 风格 (`tik`) 🆕
```
例如：7Kx9mN2pQr4
```
- 短 ID 格式（11 字符）
- 类似抖音/TikTok 风格
- 字母数字字符（0-9, A-Z, a-z）
- 适合 URL 缩短和分享

## 快速开始

### 安装

```bash
go get github.com/click33/sa-token-go/core
go get github.com/click33/sa-token-go/stputil
go get github.com/click33/sa-token-go/storage/memory
```

### 运行示例

```bash
cd examples/token-styles
go run main.go
```

### 输出

```
Sa-Token-Go Token Styles Demo
========================================

📌 UUID Style (uuid)
----------------------------------------
  1. Token for user1001:
     550e8400-e29b-41d4-a716-446655440000
  2. Token for user1002:
     f47ac10b-58cc-4372-a567-0e02b2c3d479
  3. Token for user1003:
     7c9e6679-7425-40de-944b-e07fc1f90ae7

📌 Hash Style (SHA256) (hash)
----------------------------------------
  1. Token for user1001:
     a3f5d8b2c1e4f6a9d7b8c5e2f1a4d6b9c8e5f2a7d4b1c9e6f3a8d5b2c1e7f4a6
  2. Token for user1002:
     b4f6d9c3d2e5f7b0e8c9d6f3e2b5d7c0d9f6e3b8d5c2e0f7d4b9c3e8f6b3d2f5
  3. Token for user1003:
     c5f7e0d4e3f6e8c1f9d0e7f4e3c6e8d1e0f7f4c9e6d3f1e8e5c0e9f7c4e3f6e7

📌 Timestamp Style (timestamp)
----------------------------------------
  1. Token for user1001:
     1700000000123_user1001_a3f5d8b2c1e4f6a9
  2. Token for user1002:
     1700000000456_user1002_b4f6d9c3d2e5f7b0
  3. Token for user1003:
     1700000000789_user1003_c5f7e0d4e3f6e8c1

📌 Tik Style (Short ID) (tik)
----------------------------------------
  1. Token for user1001:
     7Kx9mN2pQr4
  2. Token for user1002:
     8Ly0oO3qRs5
  3. Token for user1003:
     9Mz1pP4rSt6

========================================
✅ All token styles demonstrated!
```

## 在项目中使用

### 使用哈希风格

```go
import (
    "github.com/click33/sa-token-go/core"
    "github.com/click33/sa-token-go/stputil"
    "github.com/click33/sa-token-go/storage/memory"
)

func init() {
    stputil.SetManager(
        core.NewBuilder().
            Storage(memory.NewStorage()).
            TokenStyle(core.TokenStyleHash).  // SHA256 哈希风格
            Timeout(86400).
            Build(),
    )
}

func main() {
    token, _ := stputil.Login(1000)
    // token: a3f5d8b2c1e4f6a9d7b8c5e2f1a4d6b9c8e5f2a7d4b1c9e6f3a8d5b2c1e7f4a6
}
```

### 使用时间戳风格

```go
stputil.SetManager(
    core.NewBuilder().
        Storage(memory.NewStorage()).
        TokenStyle(core.TokenStyleTimestamp).  // 时间戳风格
        Timeout(86400).
        Build(),
)

token, _ := stputil.Login(1000)
// token: 1700000000123_1000_a3f5d8b2c1e4f6a9
```

### 使用 Tik 风格

```go
stputil.SetManager(
    core.NewBuilder().
        Storage(memory.NewStorage()).
        TokenStyle(core.TokenStyleTik).  // 短 ID 风格
        Timeout(86400).
        Build(),
)

token, _ := stputil.Login(1000)
// token: 7Kx9mN2pQr4
```

## 使用场景

| 风格 | 最适用于 | 优点 | 缺点 |
|------|----------|------|------|
| **UUID** | 通用场景 | 标准、广泛支持 | 较长 |
| **Simple** | 内部 API | 紧凑 | 熵值较低 |
| **Random32/64/128** | 高安全性 | 随机性强 | 字符串较长 |
| **JWT** | 无状态认证 | 自包含 | 体积较大 |
| **Hash** 🆕 | 安全追踪 | 高安全性、确定性 | 64 字符 |
| **Timestamp** 🆕 | 调试、审计 | 可追溯时间 | 暴露创建时间 |
| **Tik** 🆕 | URL 分享、短链接 | 很短、用户友好 | 熵值较低 |

## 下一步

- [快速开始指南](../quick-start/)
- [JWT 示例](../jwt-example/)
- [完整文档](../../docs/)

