package adapter

// Generator defines token generator interface Generator 定义 Token 生成接口
type Generator interface {
	// Generate creates a token Generate 生成 Token
	Generate(loginId, device, deviceId string) (string, error)
}

// TokenStyle defines token generation style TokenStyle 定义 Token 生成风格
type TokenStyle string

const (
	// TokenStyleUUID uses UUID style TokenStyleUUID 使用 UUID 风格
	TokenStyleUUID TokenStyle = "uuid"
	// TokenStyleSimple uses simple random string style TokenStyleSimple 使用简单随机字符串风格
	TokenStyleSimple TokenStyle = "simple"
	// TokenStyleRandom32 uses 32-char random string style TokenStyleRandom32 使用 32 位随机字符串风格
	TokenStyleRandom32 TokenStyle = "random32"
	// TokenStyleRandom64 uses 64-char random string style TokenStyleRandom64 使用 64 位随机字符串风格
	TokenStyleRandom64 TokenStyle = "random64"
	// TokenStyleRandom128 uses 128-char random string style TokenStyleRandom128 使用 128 位随机字符串风格
	TokenStyleRandom128 TokenStyle = "random128"
	// TokenStyleJWT uses JWT style TokenStyleJWT 使用 JWT 风格
	TokenStyleJWT TokenStyle = "jwt"
	// TokenStyleHash uses SHA256 hash style TokenStyleHash 使用 SHA256 哈希风格
	TokenStyleHash TokenStyle = "hash"
	// TokenStyleTimestamp uses timestamp style TokenStyleTimestamp 使用时间戳风格
	TokenStyleTimestamp TokenStyle = "timestamp"
	// TokenStyleTik uses Tik short ID style TokenStyleTik 使用 Tik 短 ID 风格
	TokenStyleTik TokenStyle = "tik"
)
