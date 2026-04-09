package chi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	SaContext "github.com/click33/sa-token-go/core/context"
	"github.com/click33/sa-token-go/core/manager"
	"github.com/click33/sa-token-go/core/serror"
	"github.com/click33/sa-token-go/stputil"
)

// LogicType defines middleware logic type LogicType 定义中间件逻辑类型
type LogicType string

const (
	// SaTokenCtxKey stores request scoped SaToken context SaTokenCtxKey 存储请求级 SaToken 上下文
	SaTokenCtxKey = "SaTokenCtx"

	// LogicOr represents OR logic LogicOr 表示或逻辑
	LogicOr LogicType = "OR"
	// LogicAnd represents AND logic LogicAnd 表示与逻辑
	LogicAnd LogicType = "AND"
)

// AuthOption defines auth option setter AuthOption 定义认证选项设置器
type AuthOption func(*AuthOptions)

// AuthOptions defines middleware auth options AuthOptions 定义中间件认证选项
type AuthOptions struct {
	AuthType  string
	LogicType LogicType
	FailFunc  func(w http.ResponseWriter, r *http.Request, err error)
}

// defaultAuthOptions returns default auth options defaultAuthOptions 返回默认认证选项
func defaultAuthOptions() *AuthOptions {
	return &AuthOptions{LogicType: LogicAnd}
}

// WithAuthType sets auth type WithAuthType 设置认证类型
func WithAuthType(authType string) AuthOption {
	return func(o *AuthOptions) {
		o.AuthType = authType
	}
}

// WithLogicType sets logic type WithLogicType 设置逻辑类型
func WithLogicType(logicType LogicType) AuthOption {
	return func(o *AuthOptions) {
		o.LogicType = logicType
	}
}

// WithFailFunc sets auth failure callback WithFailFunc 设置认证失败回调
func WithFailFunc(fn func(w http.ResponseWriter, r *http.Request, err error)) AuthOption {
	return func(o *AuthOptions) {
		o.FailFunc = fn
	}
}

// RegisterSaTokenContextMiddleware registers SaToken context middleware RegisterSaTokenContextMiddleware 注册 SaToken 上下文中间件
func RegisterSaTokenContextMiddleware(opts ...AuthOption) func(http.Handler) http.Handler {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mgr, err := stputil.GetManager(options.AuthType)
			if err != nil {
				if options.FailFunc != nil {
					options.FailFunc(w, r, err)
				} else {
					writeErrorResponse(w, err)
				}
				return
			}

			chiCtx := NewChiContext(w, r).(*ChiContext)
			_ = getSaTokenContext(chiCtx, mgr)
			next.ServeHTTP(w, chiCtx.r)
		})
	}
}

// AuthMiddleware checks login status AuthMiddleware 校验登录状态
func AuthMiddleware(opts ...AuthOption) func(http.Handler) http.Handler {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mgr, err := stputil.GetManager(options.AuthType)
			if err != nil {
				if options.FailFunc != nil {
					options.FailFunc(w, r, err)
				} else {
					writeErrorResponse(w, err)
				}
				return
			}

			chiCtx := NewChiContext(w, r).(*ChiContext)
			saCtx := getSaTokenContext(chiCtx, mgr)
			tokenValue := saCtx.GetTokenValue()

			if !mgr.IsLogin(chiCtx.r.Context(), tokenValue) {
				if options.FailFunc != nil {
					options.FailFunc(w, chiCtx.r, serror.ErrTokenExpired)
				} else {
					writeErrorResponse(w, serror.ErrTokenExpired)
				}
				return
			}

			next.ServeHTTP(w, chiCtx.r)
		})
	}
}

// PermissionMiddleware checks permissions PermissionMiddleware 校验权限
func PermissionMiddleware(permissions []string, opts ...AuthOption) func(http.Handler) http.Handler {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(permissions) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			mgr, err := stputil.GetManager(options.AuthType)
			if err != nil {
				if options.FailFunc != nil {
					options.FailFunc(w, r, err)
				} else {
					writeErrorResponse(w, err)
				}
				return
			}

			chiCtx := NewChiContext(w, r).(*ChiContext)
			saCtx := getSaTokenContext(chiCtx, mgr)
			tokenValue := saCtx.GetTokenValue()

			var ok bool
			if options.LogicType == LogicAnd {
				ok = mgr.HasPermissionsAndByToken(chiCtx.r.Context(), tokenValue, permissions)
			} else {
				ok = mgr.HasPermissionsOrByToken(chiCtx.r.Context(), tokenValue, permissions)
			}

			if !ok {
				if options.FailFunc != nil {
					options.FailFunc(w, chiCtx.r, serror.ErrPermissionDenied)
				} else {
					writeErrorResponse(w, serror.ErrPermissionDenied)
				}
				return
			}

			next.ServeHTTP(w, chiCtx.r)
		})
	}
}

// PermissionPathMiddleware checks path permissions PermissionPathMiddleware 基于路径校验权限
func PermissionPathMiddleware(permissions []string, opts ...AuthOption) func(http.Handler) http.Handler {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqPermissions := append([]string{}, permissions...)
			reqPermissions = append(reqPermissions, r.URL.Path)

			if len(reqPermissions) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			mgr, err := stputil.GetManager(options.AuthType)
			if err != nil {
				if options.FailFunc != nil {
					options.FailFunc(w, r, err)
				} else {
					writeErrorResponse(w, err)
				}
				return
			}

			chiCtx := NewChiContext(w, r).(*ChiContext)
			saCtx := getSaTokenContext(chiCtx, mgr)
			tokenValue := saCtx.GetTokenValue()

			var ok bool
			if options.LogicType == LogicAnd {
				ok = mgr.HasPermissionsAndByToken(chiCtx.r.Context(), tokenValue, reqPermissions)
			} else {
				ok = mgr.HasPermissionsOrByToken(chiCtx.r.Context(), tokenValue, reqPermissions)
			}

			if !ok {
				if options.FailFunc != nil {
					options.FailFunc(w, chiCtx.r, serror.ErrPermissionDenied)
				} else {
					writeErrorResponse(w, serror.ErrPermissionDenied)
				}
				return
			}

			next.ServeHTTP(w, chiCtx.r)
		})
	}
}

// RoleMiddleware checks roles RoleMiddleware 校验角色
func RoleMiddleware(roles []string, opts ...AuthOption) func(http.Handler) http.Handler {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(roles) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			mgr, err := stputil.GetManager(options.AuthType)
			if err != nil {
				if options.FailFunc != nil {
					options.FailFunc(w, r, err)
				} else {
					writeErrorResponse(w, err)
				}
				return
			}

			chiCtx := NewChiContext(w, r).(*ChiContext)
			saCtx := getSaTokenContext(chiCtx, mgr)
			tokenValue := saCtx.GetTokenValue()

			var ok bool
			if options.LogicType == LogicAnd {
				ok = mgr.HasRolesAndByToken(chiCtx.r.Context(), tokenValue, roles)
			} else {
				ok = mgr.HasRolesOrByToken(chiCtx.r.Context(), tokenValue, roles)
			}

			if !ok {
				if options.FailFunc != nil {
					options.FailFunc(w, chiCtx.r, serror.ErrRoleDenied)
				} else {
					writeErrorResponse(w, serror.ErrRoleDenied)
				}
				return
			}

			next.ServeHTTP(w, chiCtx.r)
		})
	}
}

// GetSaTokenContext gets cached SaToken context GetSaTokenContext 获取缓存的 SaToken 上下文
func GetSaTokenContext(r *http.Request) (*SaContext.SaTokenContext, bool) {
	v := r.Context().Value(SaTokenCtxKey)
	if v == nil {
		return nil, false
	}

	saCtx, ok := v.(*SaContext.SaTokenContext)
	return saCtx, ok
}

// GetSaTokenContextByCtx gets SaToken context by context GetSaTokenContextByCtx 从上下文获取 SaToken 上下文
func GetSaTokenContextByCtx(ctx context.Context) (*SaContext.SaTokenContext, bool) {
	v := ctx.Value(SaTokenCtxKey)
	if v == nil {
		return nil, false
	}

	saCtx, ok := v.(*SaContext.SaTokenContext)
	return saCtx, ok
}

// GetLoginIDByCtx gets login ID by context GetLoginIDByCtx 从上下文获取登录 ID
func GetLoginIDByCtx(ctx context.Context) (string, error) {
	saCtx, ok := GetSaTokenContextByCtx(ctx)
	if !ok {
		return "", serror.ErrNotLogin
	}
	return saCtx.GetLoginID(ctx)
}

// GetTokenInfoByCtx gets token info by context GetTokenInfoByCtx 从上下文获取 Token 信息
func GetTokenInfoByCtx(ctx context.Context) (*manager.TokenInfo, error) {
	saCtx, ok := GetSaTokenContextByCtx(ctx)
	if !ok {
		return nil, serror.ErrNotLogin
	}
	return saCtx.GetTokenInfo(ctx)
}

// getSaTokenContext gets or creates sa-token context getSaTokenContext 获取或创建 SaToken 上下文
func getSaTokenContext(chiCtx *ChiContext, mgr *manager.Manager) *SaContext.SaTokenContext {
	if v := chiCtx.r.Context().Value(SaTokenCtxKey); v != nil {
		if saCtx, ok := v.(*SaContext.SaTokenContext); ok {
			return saCtx
		}
	}

	saCtx := SaContext.NewContext(chiCtx, mgr)
	chiCtx.Set(SaTokenCtxKey, saCtx)
	return saCtx
}

// writeErrorResponse writes error response writeErrorResponse 写入错误响应
func writeErrorResponse(w http.ResponseWriter, err error) {
	var saErr *serror.SaTokenError
	var code int
	var message string
	var httpStatus int

	if errors.As(err, &saErr) {
		code = saErr.Code
		message = saErr.Message
		httpStatus = getHTTPStatusFromCode(code)
	} else {
		code = serror.CodeServerError
		message = err.Error()
		httpStatus = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    code,
		"message": message,
		"data":    err.Error(),
	})
}

// writeSuccessResponse writes success response writeSuccessResponse 写入成功响应
func writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    serror.CodeSuccess,
		"message": "success",
		"data":    data,
	})
}

// getHTTPStatusFromCode maps error code to HTTP status getHTTPStatusFromCode 映射错误码到 HTTP 状态码
func getHTTPStatusFromCode(code int) int {
	switch code {
	case serror.CodeNotLogin:
		return http.StatusUnauthorized
	case serror.CodePermissionDenied:
		return http.StatusForbidden
	case serror.CodeBadRequest:
		return http.StatusBadRequest
	case serror.CodeNotFound:
		return http.StatusNotFound
	case serror.CodeServerError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
