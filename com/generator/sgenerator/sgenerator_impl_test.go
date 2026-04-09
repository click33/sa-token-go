// @Author daixk 2026/4/8 10:20:00
package sgenerator

import (
	"encoding/hex"
	"strconv"
	"strings"
	"testing"

	"github.com/click33/sa-token-go/core/adapter"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TestNewGenerator tests generator constructor behavior 测试生成器构造函数行为
func TestNewGenerator(t *testing.T) {
	g := NewGenerator(3600, "my-secret", adapter.TokenStyleJWT)

	if g.timeout != 3600 {
		t.Errorf("timeout = %d, want %d", g.timeout, 3600)
	}
	if g.jwtSecretKey != "my-secret" {
		t.Errorf("jwtSecretKey = %q, want %q", g.jwtSecretKey, "my-secret")
	}
	if g.tokenStyle != adapter.TokenStyleJWT {
		t.Errorf("tokenStyle = %q, want %q", g.tokenStyle, adapter.TokenStyleJWT)
	}
}

// TestNewDefaultGenerator tests default generator behavior 测试默认生成器行为
func TestNewDefaultGenerator(t *testing.T) {
	g := NewDefaultGenerator()

	if g.timeout != DefaultTimeout {
		t.Errorf("timeout = %d, want %d", g.timeout, DefaultTimeout)
	}
	if g.jwtSecretKey != DefaultJWTSecret {
		t.Errorf("jwtSecretKey = %q, want %q", g.jwtSecretKey, DefaultJWTSecret)
	}
	if g.tokenStyle != adapter.TokenStyleUUID {
		t.Errorf("tokenStyle = %q, want %q", g.tokenStyle, adapter.TokenStyleUUID)
	}
}

// TestGenerator_Generate tests token generation behavior 测试 Token 生成行为
func TestGenerator_Generate(t *testing.T) {
	tests := []struct {
		name       string
		generator  *Generator
		loginID    string
		device     string
		deviceID   string
		checkToken func(t *testing.T, token string)
		wantErr    bool
	}{
		{
			name:      "empty login id",
			generator: NewDefaultGenerator(),
			wantErr:   true,
		},
		{
			name:      "uuid style",
			generator: NewGenerator(3600, "", adapter.TokenStyleUUID),
			loginID:   "user-1001",
			device:    "pc",
			deviceID:  "pc-001",
			checkToken: func(t *testing.T, token string) {
				if _, err := uuid.Parse(token); err != nil {
					t.Errorf("token is not valid uuid: %v", err)
				}
			},
		},
		{
			name:      "simple style",
			generator: NewGenerator(3600, "", adapter.TokenStyleSimple),
			loginID:   "user-1001",
			device:    "pc",
			deviceID:  "pc-001",
			checkToken: func(t *testing.T, token string) {
				if len(token) != DefaultSimpleLength {
					t.Errorf("token length = %d, want %d", len(token), DefaultSimpleLength)
				}
				if !containsOnlyCharset(token, TikCharset) {
					t.Errorf("token contains unexpected char: %q", token)
				}
			},
		},
		{
			name:      "random32 style",
			generator: NewGenerator(3600, "", adapter.TokenStyleRandom32),
			loginID:   "user-1001",
			device:    "pc",
			deviceID:  "pc-001",
			checkToken: func(t *testing.T, token string) {
				if len(token) != 32 {
					t.Errorf("token length = %d, want %d", len(token), 32)
				}
				if !containsOnlyCharset(token, TikCharset) {
					t.Errorf("token contains unexpected char: %q", token)
				}
			},
		},
		{
			name:      "random64 style",
			generator: NewGenerator(3600, "", adapter.TokenStyleRandom64),
			loginID:   "user-1001",
			device:    "pc",
			deviceID:  "pc-001",
			checkToken: func(t *testing.T, token string) {
				if len(token) != 64 {
					t.Errorf("token length = %d, want %d", len(token), 64)
				}
				if !containsOnlyCharset(token, TikCharset) {
					t.Errorf("token contains unexpected char: %q", token)
				}
			},
		},
		{
			name:      "random128 style",
			generator: NewGenerator(3600, "", adapter.TokenStyleRandom128),
			loginID:   "user-1001",
			device:    "pc",
			deviceID:  "pc-001",
			checkToken: func(t *testing.T, token string) {
				if len(token) != 128 {
					t.Errorf("token length = %d, want %d", len(token), 128)
				}
				if !containsOnlyCharset(token, TikCharset) {
					t.Errorf("token contains unexpected char: %q", token)
				}
			},
		},
		{
			name:      "jwt style",
			generator: NewGenerator(3600, "jwt-secret", adapter.TokenStyleJWT),
			loginID:   "user-1001",
			device:    "pc",
			deviceID:  "pc-001",
			checkToken: func(t *testing.T, token string) {
				claims, err := NewGenerator(3600, "jwt-secret", adapter.TokenStyleJWT).ParseJWT(token)
				if err != nil {
					t.Fatalf("ParseJWT() error = %v", err)
				}
				if claims["loginId"] != "user-1001" {
					t.Errorf("claims[loginId] = %v, want %q", claims["loginId"], "user-1001")
				}
				if claims["device"] != "pc" {
					t.Errorf("claims[device] = %v, want %q", claims["device"], "pc")
				}
				if claims["deviceId"] != "pc-001" {
					t.Errorf("claims[deviceId] = %v, want %q", claims["deviceId"], "pc-001")
				}
				if _, ok := claims["iat"]; !ok {
					t.Errorf("claims[iat] is missing")
				}
				if _, ok := claims["exp"]; !ok {
					t.Errorf("claims[exp] is missing")
				}
			},
		},
		{
			name:      "hash style",
			generator: NewGenerator(3600, "", adapter.TokenStyleHash),
			loginID:   "user-1001",
			device:    "pc",
			deviceID:  "pc-001",
			checkToken: func(t *testing.T, token string) {
				if len(token) != 64 {
					t.Errorf("token length = %d, want %d", len(token), 64)
				}
				if _, err := hex.DecodeString(token); err != nil {
					t.Errorf("token is not valid hex: %v", err)
				}
			},
		},
		{
			name:      "timestamp style",
			generator: NewGenerator(3600, "", adapter.TokenStyleTimestamp),
			loginID:   "user-1001",
			device:    "pc",
			deviceID:  "pc-001",
			checkToken: func(t *testing.T, token string) {
				parts := strings.Split(token, "_")
				if len(parts) != 3 {
					t.Fatalf("token parts = %d, want %d", len(parts), 3)
				}
				if _, err := strconv.ParseInt(parts[0], 10, 64); err != nil {
					t.Errorf("timestamp part is invalid: %v", err)
				}
				if parts[1] != "user-1001" {
					t.Errorf("loginID part = %q, want %q", parts[1], "user-1001")
				}
				if len(parts[2]) != TimestampRandomLen*2 {
					t.Errorf("random part length = %d, want %d", len(parts[2]), TimestampRandomLen*2)
				}
				if _, err := hex.DecodeString(parts[2]); err != nil {
					t.Errorf("random part is not valid hex: %v", err)
				}
			},
		},
		{
			name:      "tik style",
			generator: NewGenerator(3600, "", adapter.TokenStyleTik),
			loginID:   "user-1001",
			device:    "pc",
			deviceID:  "pc-001",
			checkToken: func(t *testing.T, token string) {
				if len(token) != TikTokenLength {
					t.Errorf("token length = %d, want %d", len(token), TikTokenLength)
				}
				if !containsOnlyCharset(token, TikCharset) {
					t.Errorf("token contains unexpected char: %q", token)
				}
			},
		},
		{
			name:      "invalid style fallback",
			generator: NewGenerator(3600, "", adapter.TokenStyle("invalid-style")),
			loginID:   "user-1001",
			device:    "pc",
			deviceID:  "pc-001",
			checkToken: func(t *testing.T, token string) {
				if _, err := uuid.Parse(token); err != nil {
					t.Errorf("fallback token is not valid uuid: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := tt.generator.Generate(tt.loginID, tt.device, tt.deviceID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			tt.checkToken(t, token)
		})
	}
}

// TestGenerator_GenerateJWTWithoutExpiration tests JWT expiration behavior 测试 JWT 过期时间行为
func TestGenerator_GenerateJWTWithoutExpiration(t *testing.T) {
	g := NewGenerator(0, "jwt-secret", adapter.TokenStyleJWT)

	token, err := g.Generate("user-1001", "pc", "pc-001")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	claims, err := g.ParseJWT(token)
	if err != nil {
		t.Fatalf("ParseJWT() error = %v", err)
	}
	if _, ok := claims["exp"]; ok {
		t.Errorf("claims[exp] should be absent when timeout is 0")
	}
}

// TestGenerator_ParseJWT tests JWT parse behavior 测试 JWT 解析行为
func TestGenerator_ParseJWT(t *testing.T) {
	g := NewGenerator(3600, "jwt-secret", adapter.TokenStyleJWT)

	token, err := g.Generate("user-1001", "mobile", "mobile-001")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	claims, err := g.ParseJWT(token)
	if err != nil {
		t.Fatalf("ParseJWT() error = %v", err)
	}

	if claims["loginId"] != "user-1001" {
		t.Errorf("claims[loginId] = %v, want %q", claims["loginId"], "user-1001")
	}
	if claims["device"] != "mobile" {
		t.Errorf("claims[device] = %v, want %q", claims["device"], "mobile")
	}
	if claims["deviceId"] != "mobile-001" {
		t.Errorf("claims[deviceId] = %v, want %q", claims["deviceId"], "mobile-001")
	}
}

// TestGenerator_ParseJWTError tests JWT parse error behavior 测试 JWT 解析错误行为
func TestGenerator_ParseJWTError(t *testing.T) {
	tests := []struct {
		name    string
		prepare func() (string, *Generator)
		wantErr bool
	}{
		{
			name: "empty token",
			prepare: func() (string, *Generator) {
				return "", NewDefaultGenerator()
			},
			wantErr: true,
		},
		{
			name: "invalid token",
			prepare: func() (string, *Generator) {
				return "invalid.token.string", NewDefaultGenerator()
			},
			wantErr: true,
		},
		{
			name: "wrong secret",
			prepare: func() (string, *Generator) {
				g1 := NewGenerator(3600, "secret-1", adapter.TokenStyleJWT)
				token, _ := g1.Generate("user-1001", "pc", "pc-001")
				return token, NewGenerator(3600, "secret-2", adapter.TokenStyleJWT)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, generator := tt.prepare()
			_, err := generator.ParseJWT(token)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseJWT() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGenerator_ValidateJWT tests JWT validate behavior 测试 JWT 校验行为
func TestGenerator_ValidateJWT(t *testing.T) {
	g := NewGenerator(3600, "jwt-secret", adapter.TokenStyleJWT)

	token, err := g.Generate("user-1001", "pc", "pc-001")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if err = g.ValidateJWT(token); err != nil {
		t.Errorf("ValidateJWT() error = %v", err)
	}
	if err = g.ValidateJWT("invalid.token"); err == nil {
		t.Errorf("ValidateJWT() error = nil, want non-nil")
	}
}

// TestGenerator_GetLoginIDFromJWT tests extracting loginID behavior 测试提取 loginID 行为
func TestGenerator_GetLoginIDFromJWT(t *testing.T) {
	g := NewGenerator(3600, "jwt-secret", adapter.TokenStyleJWT)

	token, err := g.Generate("user-2002", "pc", "pc-002")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	loginID, err := g.GetLoginIDFromJWT(token)
	if err != nil {
		t.Fatalf("GetLoginIDFromJWT() error = %v", err)
	}
	if loginID != "user-2002" {
		t.Errorf("loginID = %q, want %q", loginID, "user-2002")
	}
}

// TestGenerator_GetLoginIDFromJWTError tests extracting loginID error behavior 测试提取 loginID 错误行为
func TestGenerator_GetLoginIDFromJWTError(t *testing.T) {
	g := NewGenerator(3600, "jwt-secret", adapter.TokenStyleJWT)

	if _, err := g.GetLoginIDFromJWT("invalid.token"); err == nil {
		t.Errorf("GetLoginIDFromJWT() error = nil, want non-nil")
	}

	// Build a token with non-string loginId 构造 loginId 不是字符串的 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"loginId":  123,
		"device":   "pc",
		"deviceId": "pc-001",
	})
	tokenStr, err := token.SignedString([]byte("jwt-secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err = g.GetLoginIDFromJWT(tokenStr); err == nil {
		t.Errorf("GetLoginIDFromJWT() error = nil, want non-nil")
	}
}

// TestGenerator_GetJWTSecret tests JWT secret fallback behavior 测试 JWT 密钥兜底行为
func TestGenerator_GetJWTSecret(t *testing.T) {
	if got := NewGenerator(3600, "custom-secret", adapter.TokenStyleJWT).getJWTSecret(); got != "custom-secret" {
		t.Errorf("getJWTSecret() = %q, want %q", got, "custom-secret")
	}
	if got := NewGenerator(3600, "", adapter.TokenStyleJWT).getJWTSecret(); got != DefaultJWTSecret {
		t.Errorf("getJWTSecret() = %q, want %q", got, DefaultJWTSecret)
	}
}

// TestRandomStringFromCharset tests random string helper behavior 测试随机字符串辅助方法行为
func TestRandomStringFromCharset(t *testing.T) {
	token, err := randomStringFromCharset("abc123", 20)
	if err != nil {
		t.Fatalf("randomStringFromCharset() error = %v", err)
	}
	if len(token) != 20 {
		t.Errorf("token length = %d, want %d", len(token), 20)
	}
	if !containsOnlyCharset(token, "abc123") {
		t.Errorf("token contains unexpected char: %q", token)
	}

	if _, err = randomStringFromCharset("", 20); err == nil {
		t.Errorf("randomStringFromCharset() error = nil, want non-nil")
	}
	if _, err = randomStringFromCharset("abc123", 0); err == nil {
		t.Errorf("randomStringFromCharset() error = nil, want non-nil")
	}
}

// containsOnlyCharset checks whether token only uses charset 检查 Token 是否只使用指定字符集
func containsOnlyCharset(token, charset string) bool {
	for _, ch := range token {
		if !strings.ContainsRune(charset, ch) {
			return false
		}
	}
	return true
}
