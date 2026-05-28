package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMD5Hash(t *testing.T) {
	h1 := MD5Hash("hello")
	h2 := MD5Hash("hello")
	assert.Equal(t, h1, h2)
	assert.Len(t, h1, 32)
	assert.NotEqual(t, h1, MD5Hash("world"))
}

func TestSHA1Hash(t *testing.T) {
	h1 := SHA1Hash("hello")
	h2 := SHA1Hash("hello")
	assert.Equal(t, h1, h2)
	assert.Len(t, h1, 40)
	assert.NotEqual(t, h1, SHA1Hash("world"))
}

func TestAESEncryptDecrypt(t *testing.T) {
	key := []byte("0123456789abcdef") // 16 bytes for AES-128
	plaintext := "hello world, this is a test message!"

	encrypted, err := AESEncrypt(plaintext, key)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := AESDecrypt(encrypted, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestAESEncryptDecrypt_32ByteKey(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef") // 32 bytes for AES-256
	plaintext := "test with 32-byte key"

	encrypted, err := AESEncrypt(plaintext, key)
	require.NoError(t, err)

	decrypted, err := AESDecrypt(encrypted, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestAESDecrypt_InvalidKey(t *testing.T) {
	key := []byte("0123456789abcdef")
	wrongKey := []byte("fedcba9876543210")

	encrypted, err := AESEncrypt("test", key)
	require.NoError(t, err)

	_, err = AESDecrypt(encrypted, wrongKey)
	assert.Error(t, err)
}

func TestAESDecrypt_InvalidHex(t *testing.T) {
	_, err := AESDecrypt("not-hex", []byte("0123456789abcdef"))
	assert.Error(t, err)
}

func TestPasswordHash(t *testing.T) {
	h1 := PasswordHash("mypassword", "salt123")
	h2 := PasswordHash("mypassword", "salt123")
	assert.Equal(t, h1, h2)
	assert.Len(t, h1, 64)

	// Different salt produces different hash
	h3 := PasswordHash("mypassword", "differentsalt")
	assert.NotEqual(t, h1, h3)

	// Different password produces different hash
	h4 := PasswordHash("otherpassword", "salt123")
	assert.NotEqual(t, h1, h4)
}

func TestGenerateSalt(t *testing.T) {
	s1 := GenerateSalt(16)
	s2 := GenerateSalt(16)
	assert.Len(t, s1, 32) // hex encoded = 2x byte length
	assert.NotEqual(t, s1, s2)
}
