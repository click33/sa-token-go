package oauth2

import (
	"strings"

	"github.com/sa-tokens/sa-token-go/core/adapter"
)

// firstNonEmptyParam prefers query then POST form | 优先 query 再取 form
func firstNonEmptyParam(ctx adapter.RequestContext, key string) string {
	if v := strings.TrimSpace(ctx.GetQuery(key)); v != "" {
		return v
	}
	return strings.TrimSpace(ctx.GetPostForm(key))
}

// ParseScopes splits comma/space-separated scope string | 解析 scope 字符串
func ParseScopes(raw string) []string {
	if raw == "" {
		return nil
	}
	sep := func(r rune) bool { return r == ',' || r == ' ' }
	parts := strings.FieldsFunc(raw, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
