package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandomString(t *testing.T) {
	s := RandomString(32)
	assert.Len(t, s, 32)
	assert.NotEmpty(t, s)

	s2 := RandomString(32)
	assert.NotEqual(t, s, s2)
}

func TestRandomString_ZeroLength(t *testing.T) {
	assert.Equal(t, "", RandomString(0))
	assert.Equal(t, "", RandomString(-1))
}

func TestRandomNumericString(t *testing.T) {
	s := RandomNumericString(16)
	assert.Len(t, s, 16)
	for _, c := range s {
		assert.True(t, c >= '0' && c <= '9')
	}
}

func TestRandomAlphanumeric(t *testing.T) {
	s := RandomAlphanumeric(20)
	assert.Len(t, s, 20)
	assert.True(t, IsAlphanumeric(s))
}

func TestMatchPattern_SimpleWildcard(t *testing.T) {
	assert.True(t, MatchPattern("user:*", "user:add"))
	assert.True(t, MatchPattern("user:*", "user:delete"))
	assert.False(t, MatchPattern("user:*", "admin:add"))
}

func TestMatchPattern_ExactMatch(t *testing.T) {
	assert.True(t, MatchPattern("user:add", "user:add"))
	assert.False(t, MatchPattern("user:add", "user:delete"))
}

func TestMatchPattern_MultipleWildcards(t *testing.T) {
	assert.True(t, MatchPattern("user:*:view", "user:123:view"))
	assert.False(t, MatchPattern("user:*:view", "user:123:edit"))
}

func TestMatchPattern_WildcardOnly(t *testing.T) {
	assert.True(t, MatchPattern("*", "anything"))
	assert.True(t, MatchPattern("*", ""))
}

func TestParseDuration(t *testing.T) {
	assert.Equal(t, int64(3600), ParseDuration("1h"))
	assert.Equal(t, int64(1800), ParseDuration("30m"))
	assert.Equal(t, int64(86400), ParseDuration("1d"))
	assert.Equal(t, int64(604800), ParseDuration("1w"))
	assert.Equal(t, int64(3600), ParseDuration("3600"))
	assert.Equal(t, int64(86400), ParseDuration("1天"))
}

func TestFormatDuration(t *testing.T) {
	assert.Equal(t, "30秒", FormatDuration(30))
	assert.Equal(t, "5分钟", FormatDuration(300))
	assert.Equal(t, "2小时", FormatDuration(7200))
	assert.Equal(t, "3天", FormatDuration(259200))
	assert.Equal(t, "永久", FormatDuration(-1))
}

func TestSHA256Hash(t *testing.T) {
	h1 := SHA256Hash("hello")
	h2 := SHA256Hash("hello")
	assert.Equal(t, h1, h2)
	assert.Len(t, h1, 64) // hex encoded

	h3 := SHA256Hash("world")
	assert.NotEqual(t, h1, h3)
}

func TestBase64EncodeDecode(t *testing.T) {
	original := "hello world"
	encoded := Base64Encode(original)
	decoded, err := Base64Decode(encoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestBase64URLEncodeDecode(t *testing.T) {
	original := "hello world?&="
	encoded := Base64URLEncode(original)
	decoded, err := Base64URLDecode(encoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestContainsString(t *testing.T) {
	slice := []string{"a", "b", "c"}
	assert.True(t, ContainsString(slice, "a"))
	assert.True(t, ContainsString(slice, "b"))
	assert.False(t, ContainsString(slice, "d"))
}

func TestUniqueStrings(t *testing.T) {
	result := UniqueStrings([]string{"a", "b", "a", "c", "b"})
	assert.Len(t, result, 3)
	assert.Contains(t, result, "a")
	assert.Contains(t, result, "b")
	assert.Contains(t, result, "c")
}

func TestSplitAndTrim(t *testing.T) {
	result := SplitAndTrim("a , b, c , d", ",")
	assert.Equal(t, []string{"a", "b", "c", "d"}, result)

	assert.Empty(t, SplitAndTrim("", ","))
}

func TestIsAlphanumeric(t *testing.T) {
	assert.True(t, IsAlphanumeric("abc123"))
	assert.True(t, IsAlphanumeric("ABC"))
	assert.False(t, IsAlphanumeric("abc-123"))
	assert.False(t, IsAlphanumeric(""))
}

func TestIsNumeric(t *testing.T) {
	assert.True(t, IsNumeric("12345"))
	assert.False(t, IsNumeric("12a45"))
	assert.False(t, IsNumeric(""))
}

func TestInSlice(t *testing.T) {
	s := []int{1, 2, 3}
	assert.True(t, InSlice(s, 1))
	assert.True(t, InSlice(s, 3))
	assert.False(t, InSlice(s, 4))
}

func TestUniqueSlice(t *testing.T) {
	result := UniqueSlice([]int{1, 2, 1, 3, 2})
	assert.Len(t, result, 3)
}

func TestToInt(t *testing.T) {
	v, err := ToInt(42)
	require.NoError(t, err)
	assert.Equal(t, 42, v)

	v, err = ToInt("123")
	require.NoError(t, err)
	assert.Equal(t, 123, v)

	v, err = ToInt(int64(100))
	require.NoError(t, err)
	assert.Equal(t, 100, v)

	_, err = ToInt("not-a-number")
	assert.Error(t, err)
}

func TestToInt64(t *testing.T) {
	v, err := ToInt64(42)
	require.NoError(t, err)
	assert.Equal(t, int64(42), v)

	v, err = ToInt64("123")
	require.NoError(t, err)
	assert.Equal(t, int64(123), v)
}

func TestToString(t *testing.T) {
	assert.Equal(t, "hello", ToString("hello"))
	assert.Equal(t, "42", ToString(42))
	assert.Equal(t, "true", ToString(true))
	assert.Equal(t, "", ToString(nil))
}

func TestToBool(t *testing.T) {
	v, err := ToBool(true)
	require.NoError(t, err)
	assert.True(t, v)

	v, err = ToBool("true")
	require.NoError(t, err)
	assert.True(t, v)

	v, err = ToBool(1)
	require.NoError(t, err)
	assert.True(t, v)

	v, err = ToBool(0)
	require.NoError(t, err)
	assert.False(t, v)
}

func TestToBytes(t *testing.T) {
	b, err := ToBytes("hello")
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), b)

	b, err = ToBytes([]byte("world"))
	require.NoError(t, err)
	assert.Equal(t, []byte("world"), b)
}

func TestIsEmpty(t *testing.T) {
	assert.True(t, IsEmpty(""))
	assert.True(t, IsEmpty("  "))
	assert.False(t, IsEmpty("a"))
}

func TestDefaultString(t *testing.T) {
	assert.Equal(t, "default", DefaultString("", "default"))
	assert.Equal(t, "default", DefaultString("  ", "default"))
	assert.Equal(t, "value", DefaultString("value", "default"))
}

func TestRemoveString(t *testing.T) {
	result := RemoveString([]string{"a", "b", "c"}, "b")
	assert.Equal(t, []string{"a", "c"}, result)
}

func TestFilterStrings(t *testing.T) {
	result := FilterStrings([]string{"a", "bb", "ccc", "dddd"}, func(s string) bool {
		return len(s) > 2
	})
	assert.Equal(t, []string{"ccc", "dddd"}, result)
}

func TestMapStrings(t *testing.T) {
	result := MapStrings([]string{"a", "b", "c"}, func(s string) string {
		return s + "!"
	})
	assert.Equal(t, []string{"a!", "b!", "c!"}, result)
}

func TestMergeStrings(t *testing.T) {
	result := MergeStrings([]string{"a", "b"}, []string{"b", "c"}, []string{"c", "d"})
	assert.Len(t, result, 4)
}

func TestJoinNonEmpty(t *testing.T) {
	assert.Equal(t, "a,b,c", JoinNonEmpty(",", "a", "", "b", "", "c"))
	assert.Equal(t, "", JoinNonEmpty(",", "", ""))
}

func TestParsePermissionTag(t *testing.T) {
	assert.Equal(t, []string{"user:read", "user:write"}, ParsePermissionTag("perm:user:read,user:write"))
	assert.Empty(t, ParsePermissionTag(""))
}

func TestParseRoleTag(t *testing.T) {
	assert.Equal(t, []string{"admin", "manager"}, ParseRoleTag("role:admin,manager"))
	assert.Empty(t, ParseRoleTag(""))
}
