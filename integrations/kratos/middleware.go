package kratos

import (
	"context"
	stderrors "errors"
	"net/http"

	corecontext "github.com/click33/sa-token-go/core/context"
	"github.com/click33/sa-token-go/core/manager"
	"github.com/click33/sa-token-go/core/serror"
	"github.com/click33/sa-token-go/stputil"
	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
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

// FailFunc defines auth failure callback FailFunc 定义认证失败回调
type FailFunc func(ctx context.Context, err error) error

// AuthOption defines auth option setter AuthOption 定义认证选项设置器
type AuthOption func(*AuthOptions)

// AuthOptions defines middleware auth options AuthOptions 定义中间件认证选项
type AuthOptions struct {
	AuthType  string
	LogicType LogicType
	FailFunc  FailFunc
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
func WithFailFunc(fn FailFunc) AuthOption {
	return func(o *AuthOptions) {
		o.FailFunc = fn
	}
}

// RegisterSaTokenContextMiddleware registers SaToken context middleware RegisterSaTokenContextMiddleware 注册 SaToken 上下文中间件
func RegisterSaTokenContextMiddleware(opts ...AuthOption) middleware.Middleware {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			mgr, err := stputil.GetManager(options.AuthType)
			if err != nil {
				return nil, dispatchFail(ctx, options.FailFunc, err)
			}

			_, ctx = getSaTokenContext(ctx, mgr)
			return next(ctx, req)
		}
	}
}

// AuthMiddleware checks login status AuthMiddleware 校验登录状态
func AuthMiddleware(opts ...AuthOption) middleware.Middleware {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			mgr, err := stputil.GetManager(options.AuthType)
			if err != nil {
				return nil, dispatchFail(ctx, options.FailFunc, err)
			}

			saCtx, ctx := getSaTokenContext(ctx, mgr)
			if !mgr.IsLogin(ctx, saCtx.GetTokenValue()) {
				return nil, dispatchFail(ctx, options.FailFunc, serror.ErrNotLogin)
			}

			return next(ctx, req)
		}
	}
}

// PermissionMiddleware checks permissions PermissionMiddleware 校验权限
func PermissionMiddleware(permissions []string, opts ...AuthOption) middleware.Middleware {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			if len(permissions) == 0 {
				return next(ctx, req)
			}

			mgr, err := stputil.GetManager(options.AuthType)
			if err != nil {
				return nil, dispatchFail(ctx, options.FailFunc, err)
			}

			saCtx, ctx := getSaTokenContext(ctx, mgr)
			tokenValue := saCtx.GetTokenValue()

			var ok bool
			if options.LogicType == LogicAnd {
				ok = mgr.HasPermissionsAndByToken(ctx, tokenValue, permissions)
			} else {
				ok = mgr.HasPermissionsOrByToken(ctx, tokenValue, permissions)
			}

			if !ok {
				return nil, dispatchFail(ctx, options.FailFunc, serror.ErrPermissionDenied)
			}

			return next(ctx, req)
		}
	}
}

// PermissionPathMiddleware checks path permissions PermissionPathMiddleware 基于路径校验权限
func PermissionPathMiddleware(permissions []string, opts ...AuthOption) middleware.Middleware {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			reqPermissions := append([]string{}, permissions...)
			if path := NewKratosContext(ctx).GetPath(); path != "" {
				reqPermissions = append(reqPermissions, path)
			}

			if len(reqPermissions) == 0 {
				return next(ctx, req)
			}

			mgr, err := stputil.GetManager(options.AuthType)
			if err != nil {
				return nil, dispatchFail(ctx, options.FailFunc, err)
			}

			saCtx, ctx := getSaTokenContext(ctx, mgr)
			tokenValue := saCtx.GetTokenValue()

			var ok bool
			if options.LogicType == LogicAnd {
				ok = mgr.HasPermissionsAndByToken(ctx, tokenValue, reqPermissions)
			} else {
				ok = mgr.HasPermissionsOrByToken(ctx, tokenValue, reqPermissions)
			}

			if !ok {
				return nil, dispatchFail(ctx, options.FailFunc, serror.ErrPermissionDenied)
			}

			return next(ctx, req)
		}
	}
}

// RoleMiddleware checks roles RoleMiddleware 校验角色
func RoleMiddleware(roles []string, opts ...AuthOption) middleware.Middleware {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			if len(roles) == 0 {
				return next(ctx, req)
			}

			mgr, err := stputil.GetManager(options.AuthType)
			if err != nil {
				return nil, dispatchFail(ctx, options.FailFunc, err)
			}

			saCtx, ctx := getSaTokenContext(ctx, mgr)
			tokenValue := saCtx.GetTokenValue()

			var ok bool
			if options.LogicType == LogicAnd {
				ok = mgr.HasRolesAndByToken(ctx, tokenValue, roles)
			} else {
				ok = mgr.HasRolesOrByToken(ctx, tokenValue, roles)
			}

			if !ok {
				return nil, dispatchFail(ctx, options.FailFunc, serror.ErrRoleDenied)
			}

			return next(ctx, req)
		}
	}
}

// GetSaTokenContext gets cached SaToken context GetSaTokenContext 获取缓存的 SaToken 上下文
func GetSaTokenContext(ctx context.Context) (*corecontext.SaTokenContext, bool) {
	value := ctx.Value(SaTokenCtxKey)
	if value == nil {
		return nil, false
	}

	saCtx, ok := value.(*corecontext.SaTokenContext)
	return saCtx, ok
}

// GetSaTokenContextByCtx gets cached SaToken context by ctx GetSaTokenContextByCtx 从上下文获取 SaToken 上下文
func GetSaTokenContextByCtx(ctx context.Context) (*corecontext.SaTokenContext, bool) {
	return GetSaTokenContext(ctx)
}

// GetLoginIDByCtx gets login ID by context GetLoginIDByCtx 从上下文获取登录 ID
func GetLoginIDByCtx(ctx context.Context, authType ...string) (string, error) {
	mgr, err := stputil.GetManager(authType...)
	if err != nil {
		return "", err
	}

	saCtx, ctx := getSaTokenContext(ctx, mgr)
	return mgr.GetLoginID(ctx, saCtx.GetTokenValue())
}

// GetTokenInfoByCtx gets token info by context GetTokenInfoByCtx 从上下文获取 Token 信息
func GetTokenInfoByCtx(ctx context.Context, authType ...string) (*manager.TokenInfo, error) {
	mgr, err := stputil.GetManager(authType...)
	if err != nil {
		return nil, err
	}

	saCtx, ctx := getSaTokenContext(ctx, mgr)
	return mgr.GetTokenInfo(ctx, saCtx.GetTokenValue())
}

// getSaTokenContext gets or creates sa-token context getSaTokenContext 获取或创建 SaToken 上下文
func getSaTokenContext(ctx context.Context, mgr *manager.Manager) (*corecontext.SaTokenContext, context.Context) {
	if saCtx, ok := GetSaTokenContext(ctx); ok {
		return saCtx, ctx
	}

	kratosCtx := NewKratosContext(ctx).(*KratosContext)
	saCtx := corecontext.NewContext(kratosCtx, mgr)
	ctx = context.WithValue(ctx, SaTokenCtxKey, saCtx)
	kratosCtx.ctx = ctx

	return saCtx, ctx
}

// dispatchFail dispatches auth failure dispatchFail 分发认证失败处理
func dispatchFail(ctx context.Context, failFunc FailFunc, err error) error {
	if failFunc != nil {
		return failFunc(ctx, err)
	}
	return writeErrorResponse(err)
}

// writeErrorResponse converts error to kratos error writeErrorResponse 转换为 Kratos 错误
func writeErrorResponse(err error) error {
	code, message := getErrorCodeAndMessage(err)
	return kerrors.New(getHTTPStatusFromCode(code), getReasonFromCode(code), message).WithCause(err)
}

// getErrorCodeAndMessage gets error code and message getErrorCodeAndMessage 获取错误码和错误消息
func getErrorCodeAndMessage(err error) (int, string) {
	var saErr *serror.SaTokenError
	if stderrors.As(err, &saErr) {
		return saErr.Code, saErr.Message
	}

	switch {
	case stderrors.Is(err, serror.ErrNotLogin):
		return serror.CodeNotLogin, err.Error()
	case stderrors.Is(err, serror.ErrInvalidToken):
		return serror.CodeTokenInvalid, err.Error()
	case stderrors.Is(err, serror.ErrTokenExpired), stderrors.Is(err, serror.ErrTokenKickout), stderrors.Is(err, serror.ErrTokenReplaced):
		return serror.CodeTokenExpired, err.Error()
	case stderrors.Is(err, serror.ErrPermissionDenied), stderrors.Is(err, serror.ErrRoleDenied):
		return serror.CodePermissionDenied, err.Error()
	case stderrors.Is(err, serror.ErrAccountDisabled):
		return serror.CodeAccountDisabled, err.Error()
	case stderrors.Is(err, serror.ErrInvalidParam):
		return serror.CodeBadRequest, err.Error()
	default:
		return serror.CodeServerError, err.Error()
	}
}

// getHTTPStatusFromCode maps error code to HTTP status getHTTPStatusFromCode 映射错误码到 HTTP 状态码
func getHTTPStatusFromCode(code int) int {
	switch code {
	case serror.CodeNotLogin, serror.CodeTokenInvalid, serror.CodeTokenExpired:
		return http.StatusUnauthorized
	case serror.CodePermissionDenied, serror.CodeAccountDisabled:
		return http.StatusForbidden
	case serror.CodeBadRequest:
		return http.StatusBadRequest
	case serror.CodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

// getReasonFromCode maps error code to reason getReasonFromCode 映射错误码到错误原因
func getReasonFromCode(code int) string {
	switch code {
	case serror.CodeNotLogin, serror.CodeTokenInvalid, serror.CodeTokenExpired:
		return "UNAUTHORIZED"
	case serror.CodePermissionDenied:
		return "PERMISSION_DENIED"
	case serror.CodeAccountDisabled:
		return "ACCOUNT_DISABLED"
	case serror.CodeBadRequest:
		return "BAD_REQUEST"
	case serror.CodeNotFound:
		return "NOT_FOUND"
	default:
		return "INTERNAL_SERVER_ERROR"
	}
}
