package errs

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSentinelErrors_NotNil(t *testing.T) {
	sentinels := []error{
		ErrNotLogin,
		ErrTokenNotFound,
		ErrTokenExpired,
		ErrKickedOut,
		ErrTokenReplaced,
		ErrAccountDisabled,
		ErrPermissionDenied,
		ErrRoleDenied,
		ErrMaxLoginCount,
		ErrInvalidTokenData,
		ErrInvalidSessionData,
		ErrSessionNotFound,
		ErrStorageUnavailable,
		ErrInvalidConfig,
	}
	for i, err := range sentinels {
		assert.NotNil(t, err, "sentinel at index %d should not be nil", i)
	}
}

func TestSentinelErrors_ErrorString(t *testing.T) {
	assert.NotEmpty(t, ErrNotLogin.Error())
	assert.NotEmpty(t, ErrTokenNotFound.Error())
	assert.NotEmpty(t, ErrKickedOut.Error())
	assert.NotEmpty(t, ErrPermissionDenied.Error())
}

func TestErrorsIs_Wrapped(t *testing.T) {
	wrapped := ErrTokenNotFoundForLogin("user1")
	assert.True(t, errors.Is(wrapped, ErrTokenNotFound))

	wrapped2 := ErrKickedOutWithToken("tok123")
	assert.True(t, errors.Is(wrapped2, ErrKickedOut))
}

func TestErrStorageWrap(t *testing.T) {
	cause := errors.New("connection refused")
	wrapped := ErrStorageWrap(cause)
	assert.True(t, errors.Is(wrapped, ErrStorageUnavailable))
	assert.Contains(t, wrapped.Error(), "connection refused")
}

func TestErrMarshalTokenInfo(t *testing.T) {
	cause := errors.New("json error")
	wrapped := ErrMarshalTokenInfo(cause)
	assert.True(t, errors.Is(wrapped, ErrInvalidTokenData))
	assert.Contains(t, wrapped.Error(), "json error")
}

func TestErrInvalidTokenDataWrap(t *testing.T) {
	cause := errors.New("bad data")
	wrapped := ErrInvalidTokenDataWrap(cause)
	assert.True(t, errors.Is(wrapped, ErrInvalidTokenData))
}

func TestErrKickedOutWithToken(t *testing.T) {
	err := ErrKickedOutWithToken("abc123")
	assert.True(t, errors.Is(err, ErrKickedOut))
	assert.Contains(t, err.Error(), "abc123")
}

func TestErrTokenReplacedWithToken(t *testing.T) {
	err := ErrTokenReplacedWithToken("xyz789")
	assert.True(t, errors.Is(err, ErrTokenReplaced))
	assert.Contains(t, err.Error(), "xyz789")
}

func TestErrFeatureNotSupportedNamed(t *testing.T) {
	err := ErrFeatureNotSupportedNamed("temp-token")
	assert.True(t, errors.Is(err, ErrFeatureNotSupported))
	assert.Contains(t, err.Error(), "temp-token")
}

func TestErrOAuth2ParamMissing(t *testing.T) {
	err := ErrOAuth2ParamMissing("redirect_uri")
	assert.True(t, errors.Is(err, ErrOAuth2RequiredParamMissing))
	assert.Contains(t, err.Error(), "redirect_uri")
}

func TestErrDisableLevelExceededWithContext(t *testing.T) {
	err := ErrDisableLevelExceededWithContext("user1", "spam", 3)
	assert.True(t, errors.Is(err, ErrDisableLevelExceeded))
	assert.Contains(t, err.Error(), "user1")
	assert.Contains(t, err.Error(), "spam")
}
