English | [中文文档](README_zh.md)

# Token Styles Example

This example demonstrates all available token generation styles in Sa-Token-Go.

## Available Token Styles

### 1. UUID Style (`uuid`)
```
e.g., 550e8400-e29b-41d4-a716-446655440000
```
- Standard UUID v4 format
- 36 characters (including hyphens)
- Globally unique

### 2. Simple Style (`simple`)
```
e.g., aB3dE5fG7hI9jK1l
```
- 16-character random string
- Base64 URL-safe encoding
- Compact and simple

### 3. Random32 Style (`random32`)
```
e.g., aB3dE5fG7hI9jK1lMnO2pQ4rS6tU8vW0
```
- 32-character random string
- High randomness
- Secure and unique

### 4. Random64 Style (`random64`)
```
e.g., aB3dE5fG7hI9jK1lMnO2pQ4rS6tU8vW0xY1zA2bC3dD4eE5fF6gG7hH8iI9jJ0kK1l
```
- 64-character random string
- Maximum randomness
- Extra secure

### 5. Random128 Style (`random128`)
```
e.g., aB3dE5fG7hI9jK1lMnO2pQ4rS6tU8vW0xY1zA2bC3dD4eE5fF6gG7hH8iI9jJ0kK1lMmN2nO3oP4pQ5qR6rS7sT8tU9uV0vW1wX2xY3yZ4zA5aB6bC7cD8dE9eF0fG1gH2hI3iJ4jK5kL6lM7mN8nO9oP0
```
- 128-character random string
- Extremely secure
- For high-security scenarios

### 6. JWT Style (`jwt`)
```
e.g., eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkZXZpY2UiOiJkZWZhdWx0IiwiaWF0IjoxNzAwMDAwMDAwLCJsb2dpbklkIjoidXNlcjEwMDAifQ.xxx
```
- Standard JWT format
- Contains claims (loginId, device, iat, exp)
- Self-contained and verifiable
- Requires `JwtSecretKey` configuration

### 7. Hash Style (`hash`) 🆕
```
e.g., a3f5d8b2c1e4f6a9d7b8c5e2f1a4d6b9c8e5f2a7d4b1c9e6f3a8d5b2c1e7f4a6
```
- SHA256 hash-based token
- Combines loginID, device, timestamp, and random data
- 64-character hexadecimal
- High security and unpredictability

### 8. Timestamp Style (`timestamp`) 🆕
```
e.g., 1700000000123_user1000_a3f5d8b2c1e4f6a9
```
- Format: `timestamp_loginID_random`
- Millisecond precision timestamp
- Easily traceable creation time
- Good for debugging and logging

### 9. Tik Style (`tik`) 🆕
```
e.g., 7Kx9mN2pQr4
```
- Short ID format (11 characters)
- Similar to TikTok/Douyin style
- Alphanumeric characters (0-9, A-Z, a-z)
- Perfect for URL shortening and sharing

## Quick Start

### Installation

```bash
go get github.com/click33/sa-token-go/core
go get github.com/click33/sa-token-go/stputil
go get github.com/click33/sa-token-go/storage/memory
```

### Run the Example

```bash
cd examples/token-styles
go run main.go
```

### Output

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

## Usage in Your Project

### Using Hash Style

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
            TokenStyle(core.TokenStyleHash).  // SHA256 hash style
            Timeout(86400).
            Build(),
    )
}

func main() {
    token, _ := stputil.Login(1000)
    // token: a3f5d8b2c1e4f6a9d7b8c5e2f1a4d6b9c8e5f2a7d4b1c9e6f3a8d5b2c1e7f4a6
}
```

### Using Timestamp Style

```go
stputil.SetManager(
    core.NewBuilder().
        Storage(memory.NewStorage()).
        TokenStyle(core.TokenStyleTimestamp).  // Timestamp style
        Timeout(86400).
        Build(),
)

token, _ := stputil.Login(1000)
// token: 1700000000123_1000_a3f5d8b2c1e4f6a9
```

### Using Tik Style

```go
stputil.SetManager(
    core.NewBuilder().
        Storage(memory.NewStorage()).
        TokenStyle(core.TokenStyleTik).  // Short ID style
        Timeout(86400).
        Build(),
)

token, _ := stputil.Login(1000)
// token: 7Kx9mN2pQr4
```

## Use Cases

| Style | Best For | Pros | Cons |
|-------|----------|------|------|
| **UUID** | General purpose | Standard, widely supported | Longer |
| **Simple** | Internal APIs | Compact | Less entropy |
| **Random32/64/128** | High security | Very random | Longer strings |
| **JWT** | Stateless auth | Self-contained | Larger size |
| **Hash** 🆕 | Secure tracking | High security, deterministic | 64 chars |
| **Timestamp** 🆕 | Debugging, auditing | Time-traceable | Exposes creation time |
| **Tik** 🆕 | URL sharing, short links | Very short, user-friendly | Lower entropy |

## Next Steps

- [Quick Start Guide](../quick-start/)
- [JWT Example](../jwt-example/)
- [Full Documentation](../../docs/)

