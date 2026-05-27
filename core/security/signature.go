package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sa-tokens/sa-token-go/core/adapter"
	"github.com/sa-tokens/sa-token-go/core/errs"
)

// SignTemplate provides API parameter signing and verification | API 参数签名模板
type SignTemplate struct {
	storage   adapter.Storage
	keyPrefix string
	nonceTTL  time.Duration
}

// NewSignTemplate creates a new signature template | 创建签名模板
func NewSignTemplate(storage adapter.Storage, prefix string, nonceTTL time.Duration) *SignTemplate {
	if nonceTTL <= 0 {
		nonceTTL = 5 * time.Minute
	}
	return &SignTemplate{
		storage:   storage,
		keyPrefix: prefix + "sign:nonce:",
		nonceTTL:  nonceTTL,
	}
}

// Sign generates HMAC-SHA256 signature from sorted params | 生成 HMAC-SHA256 签名
func (s *SignTemplate) Sign(params map[string]string, secret string) string {
	// Sort params by key
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(params[k])
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sb.String()))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySign verifies signature with timestamp and nonce replay protection | 验证签名（含时间戳+Nonce防重放）
func (s *SignTemplate) VerifySign(params map[string]string, secret, timestamp, nonce, signature string, maxAgeSeconds int64) error {
	// Check timestamp freshness
	if maxAgeSeconds > 0 {
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: invalid timestamp", errs.ErrSignatureExpired)
		}
		if time.Now().Unix()-ts > maxAgeSeconds {
			return errs.ErrSignatureExpired
		}
	}

	// Check nonce replay
	if nonce != "" {
		nonceKey := s.keyPrefix + nonce
		exists, err := s.storage.Get(nonceKey)
		if err == nil && exists != nil {
			return errs.ErrNonceAlreadyUsed
		}
		// Store nonce to prevent replay
		_ = s.storage.Set(nonceKey, "1", s.nonceTTL)
	}

	// Add timestamp and nonce to params for verification
	verifyParams := make(map[string]string, len(params)+2)
	for k, v := range params {
		verifyParams[k] = v
	}
	if timestamp != "" {
		verifyParams["timestamp"] = timestamp
	}
	if nonce != "" {
		verifyParams["nonce"] = nonce
	}

	expected := s.Sign(verifyParams, secret)
	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return errs.ErrSignatureInvalid
	}

	return nil
}
