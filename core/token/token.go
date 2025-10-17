package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/click33/sa-token-go/core/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Generator Token generator | Token生成器
type Generator struct {
	config *config.Config
}

// NewGenerator creates a new token generator | 创建新的Token生成器
func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{
		config: cfg,
	}
}

// Generate generates token based on configured style | 根据配置的风格生成Token
func (g *Generator) Generate(loginID string, device string) (string, error) {
	switch g.config.TokenStyle {
	case config.TokenStyleUUID:
		return g.generateUUID()
	case config.TokenStyleSimple:
		return g.generateSimple(16)
	case config.TokenStyleRandom32:
		return g.generateSimple(32)
	case config.TokenStyleRandom64:
		return g.generateSimple(64)
	case config.TokenStyleRandom128:
		return g.generateSimple(128)
	case config.TokenStyleJWT:
		return g.generateJWT(loginID, device)
	case config.TokenStyleHash:
		return g.generateHash(loginID, device)
	case config.TokenStyleTimestamp:
		return g.generateTimestamp(loginID, device)
	case config.TokenStyleTik:
		return g.generateTik()
	default:
		return g.generateUUID()
	}
}

// generateUUID generates UUID token | 生成UUID Token
func (g *Generator) generateUUID() (string, error) {
	return uuid.New().String(), nil
}

// generateSimple generates simple random string token | 生成简单随机字符串Token
func (g *Generator) generateSimple(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// generateJWT generates JWT token | 生成JWT Token
func (g *Generator) generateJWT(loginID string, device string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"loginId": loginID,
		"device":  device,
		"iat":     now.Unix(),
	}

	if g.config.Timeout > 0 {
		claims["exp"] = now.Add(time.Duration(g.config.Timeout) * time.Second).Unix()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secretKey := g.config.JwtSecretKey
	if secretKey == "" {
		secretKey = "default-secret-key"
	}

	return token.SignedString([]byte(secretKey))
}

// ParseJWT parses JWT token and returns claims | 解析JWT Token并返回声明
func (g *Generator) ParseJWT(tokenStr string) (jwt.MapClaims, error) {
	secretKey := g.config.JwtSecretKey
	if secretKey == "" {
		secretKey = "default-secret-key"
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateJWT validates JWT token | 验证JWT Token
func (g *Generator) ValidateJWT(tokenStr string) error {
	_, err := g.ParseJWT(tokenStr)
	return err
}

// generateHash generates SHA256 hash-based token | 生成SHA256哈希风格Token
func (g *Generator) generateHash(loginID string, device string) (string, error) {
	// Combine loginID, device, timestamp and random bytes
	// 组合 loginID、device、时间戳和随机字节
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	data := fmt.Sprintf("%s:%s:%d:%s", loginID, device, time.Now().UnixNano(), hex.EncodeToString(randomBytes))
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:]), nil
}

// generateTimestamp generates timestamp-based token | 生成时间戳风格Token
func (g *Generator) generateTimestamp(loginID string, device string) (string, error) {
	// Format: timestamp_loginID_random
	// 格式：时间戳_loginID_随机数
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	timestamp := time.Now().UnixMilli()
	random := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%d_%s_%s", timestamp, loginID, random), nil
}

// generateTik generates short ID style token (like TikTok) | 生成Tik风格短ID Token（类似抖音）
func (g *Generator) generateTik() (string, error) {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	const length = 11 // TikTok-style short ID length | 抖音风格短ID长度

	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}
