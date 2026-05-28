package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchPath_ExactMatch(t *testing.T) {
	assert.True(t, MatchPath("/api/users", "/api/users"))
	assert.False(t, MatchPath("/api/users", "/api/user"))
	assert.False(t, MatchPath("/api/users/", "/api/users"))
}

func TestMatchPath_WildcardAll(t *testing.T) {
	assert.True(t, MatchPath("/", "/**"))
	assert.True(t, MatchPath("/api/users", "/**"))
	assert.True(t, MatchPath("/any/path/here", "/**"))
}

func TestMatchPath_DoubleWildcardSuffix(t *testing.T) {
	assert.True(t, MatchPath("/api/", "/api/**"))
	assert.True(t, MatchPath("/api/users", "/api/**"))
	assert.True(t, MatchPath("/api/users/123", "/api/**"))
	assert.False(t, MatchPath("/other/path", "/api/**"))
}

func TestMatchPath_SingleWildcard(t *testing.T) {
	assert.True(t, MatchPath("/api/users", "/api/*"))
	assert.False(t, MatchPath("/api/users/123", "/api/*"))
	assert.True(t, MatchPath("/api/", "/api/*"))
}

func TestMatchPath_SuffixWildcard(t *testing.T) {
	assert.True(t, MatchPath("/page.html", "*.html"))
	assert.True(t, MatchPath("/deep/page.html", "*.html"))
	assert.False(t, MatchPath("/page.css", "*.html"))
}

func TestMatchPath_NoMatch(t *testing.T) {
	assert.False(t, MatchPath("/api/users", "/other/path"))
	assert.False(t, MatchPath("", "/api"))
}

func TestMatchAny(t *testing.T) {
	patterns := []string{"/api/**", "/public/**"}
	assert.True(t, MatchAny("/api/users", patterns))
	assert.True(t, MatchAny("/public/index.html", patterns))
	assert.False(t, MatchAny("/admin/dashboard", patterns))
}

func TestNeedAuth(t *testing.T) {
	include := []string{"/api/**"}
	exclude := []string{"/api/public/**"}

	assert.True(t, NeedAuth("/api/users", include, exclude))
	assert.False(t, NeedAuth("/api/public/health", include, exclude))
	assert.False(t, NeedAuth("/other", include, exclude))
}

func TestPathAuthConfig_Check(t *testing.T) {
	cfg := NewPathAuthConfig().
		SetInclude([]string{"/api/**", "/admin/**"}).
		SetExclude([]string{"/api/public/**"})

	assert.True(t, cfg.Check("/api/users"))
	assert.True(t, cfg.Check("/admin/dashboard"))
	assert.False(t, cfg.Check("/api/public/health"))
	assert.False(t, cfg.Check("/other"))
}

func TestPathAuthConfig_ValidateLoginID(t *testing.T) {
	cfg := NewPathAuthConfig()
	assert.True(t, cfg.ValidateLoginID("anyone")) // no validator = always true

	cfg.SetValidator(func(loginID string) bool {
		return loginID == "admin"
	})
	assert.True(t, cfg.ValidateLoginID("admin"))
	assert.False(t, cfg.ValidateLoginID("user"))
}

func TestAuthResult_ShouldReject(t *testing.T) {
	// No auth needed, no token - should not reject
	r := &AuthResult{NeedAuth: false, Token: "", IsValid: false}
	assert.False(t, r.ShouldReject())

	// Auth needed, no token - should reject
	r = &AuthResult{NeedAuth: true, Token: "", IsValid: false}
	assert.True(t, r.ShouldReject())

	// Auth needed, valid token - should not reject
	r = &AuthResult{NeedAuth: true, Token: "abc", IsValid: true}
	assert.False(t, r.ShouldReject())

	// Auth needed, invalid token - should reject
	r = &AuthResult{NeedAuth: true, Token: "abc", IsValid: false}
	assert.True(t, r.ShouldReject())
}

func TestAuthResult_LoginID(t *testing.T) {
	r := &AuthResult{}
	assert.Equal(t, "", r.LoginID())
}
