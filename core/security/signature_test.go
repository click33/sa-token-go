package security

import (
	"testing"
	"time"

	"github.com/sa-tokens/sa-token-go/core/errs"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignTemplate_Sign(t *testing.T) {
	storage := memory.NewStorage()
	st := NewSignTemplate(storage, "test:", 5*time.Minute)

	params := map[string]string{
		"name": "test",
		"age":  "25",
	}

	sig := st.Sign(params, "mysecret")
	assert.NotEmpty(t, sig)
	assert.Len(t, sig, 64) // SHA256 hex = 64 chars

	// Same params + secret = same signature
	sig2 := st.Sign(params, "mysecret")
	assert.Equal(t, sig, sig2)

	// Different secret = different signature
	sig3 := st.Sign(params, "othersecret")
	assert.NotEqual(t, sig, sig3)
}

func TestSignTemplate_VerifySign_Valid(t *testing.T) {
	storage := memory.NewStorage()
	st := NewSignTemplate(storage, "test:", 5*time.Minute)

	secret := "mysecret"
	timestamp := "1700000000"
	nonce := "abc123"

	params := map[string]string{
		"name": "test",
		"age":  "25",
	}

	// Build params with timestamp and nonce for signing
	signParams := map[string]string{
		"name":      "test",
		"age":       "25",
		"timestamp": timestamp,
		"nonce":     nonce,
	}
	sig := st.Sign(signParams, secret)

	err := st.VerifySign(params, secret, timestamp, nonce, sig, 0)
	require.NoError(t, err)
}

func TestSignTemplate_VerifySign_InvalidSignature(t *testing.T) {
	storage := memory.NewStorage()
	st := NewSignTemplate(storage, "test:", 5*time.Minute)

	params := map[string]string{"name": "test"}

	err := st.VerifySign(params, "secret", "1700000000", "nonce123", "badsignature", 0)
	assert.ErrorIs(t, err, errs.ErrSignatureInvalid)
}

func TestSignTemplate_VerifySign_ExpiredTimestamp(t *testing.T) {
	storage := memory.NewStorage()
	st := NewSignTemplate(storage, "test:", 5*time.Minute)

	secret := "mysecret"
	oldTimestamp := "1000000000" // very old
	nonce := "nonce1"

	params := map[string]string{"name": "test"}
	signParams := map[string]string{
		"name":      "test",
		"timestamp": oldTimestamp,
		"nonce":     nonce,
	}
	sig := st.Sign(signParams, secret)

	err := st.VerifySign(params, secret, oldTimestamp, nonce, sig, 60)
	assert.ErrorIs(t, err, errs.ErrSignatureExpired)
}

func TestSignTemplate_VerifySign_NonceReplay(t *testing.T) {
	storage := memory.NewStorage()
	st := NewSignTemplate(storage, "test:", 5*time.Minute)

	secret := "mysecret"
	timestamp := "1700000000"
	nonce := "unique-nonce"

	params := map[string]string{"name": "test"}
	signParams := map[string]string{
		"name":      "test",
		"timestamp": timestamp,
		"nonce":     nonce,
	}
	sig := st.Sign(signParams, secret)

	// First verification should succeed
	err := st.VerifySign(params, secret, timestamp, nonce, sig, 0)
	require.NoError(t, err)

	// Second verification with same nonce should fail (replay)
	err = st.VerifySign(params, secret, timestamp, nonce, sig, 0)
	assert.ErrorIs(t, err, errs.ErrNonceAlreadyUsed)
}

func TestSignTemplate_Sign_Deterministic(t *testing.T) {
	storage := memory.NewStorage()
	st := NewSignTemplate(storage, "test:", 5*time.Minute)

	// Params in different order should produce same signature
	params1 := map[string]string{"a": "1", "b": "2", "c": "3"}
	params2 := map[string]string{"c": "3", "a": "1", "b": "2"}

	sig1 := st.Sign(params1, "secret")
	sig2 := st.Sign(params2, "secret")
	assert.Equal(t, sig1, sig2)
}
