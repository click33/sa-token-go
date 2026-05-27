package gozero

import (
	"context"
	"net/http"

	"github.com/sa-tokens/sa-token-go/core"
	"github.com/sa-tokens/sa-token-go/core/adapter"
)

// attachSaTokenToRequest injects Sa-Token context and optional loginID into request context.
// attachSaTokenToRequest 将 Sa-Token 上下文及可选 loginID 注入 request，并返回新 *http.Request。
func attachSaTokenToRequest(w http.ResponseWriter, r *http.Request, saCtx *core.SaTokenContext, loginID string) *http.Request {
	gz := NewGoZeroContext(w, r).(*GoZeroContext)
	gz.SetKey(ctxKeySaToken, saCtx)
	if loginID != "" {
		gz.SetKey(ctxKeyLoginID, loginID)
	}
	return gz.Request()
}

// attachTokenToRequest injects raw token into request context.
// attachTokenToRequest 将原始 token 注入 request context。
func attachTokenToRequest(r *http.Request, token string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ctxKeyToken, token))
}

// mustGoZeroRequest extracts *http.Request from GoZeroContext adapter.
// mustGoZeroRequest 从 GoZeroContext 适配器取出 *http.Request。
func mustGoZeroRequest(rc adapter.RequestContext) *http.Request {
	if gz, ok := rc.(*GoZeroContext); ok {
		return gz.Request()
	}
	return nil
}
