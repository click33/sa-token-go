package gf

import (
	"context"
	"errors"
	"net/http"

	"github.com/click33/sa-token-go/core/manager"
	"github.com/click33/sa-token-go/core/serror"

	corecontext "github.com/click33/sa-token-go/core/context"
	"github.com/click33/sa-token-go/stputil"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
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
	FailFunc  func(r *ghttp.Request, err error)
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
func WithFailFunc(fn func(r *ghttp.Request, err error)) AuthOption {
	return func(o *AuthOptions) {
		o.FailFunc = fn
	}
}

// RegisterSaTokenContextMiddleware registers SaToken context middleware RegisterSaTokenContextMiddleware 注册 SaToken 上下文中间件
func RegisterSaTokenContextMiddleware(ctx context.Context, opts ...AuthOption) ghttp.HandlerFunc {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(r *ghttp.Request) {
		mgr, err := stputil.GetManager(options.AuthType)
		if err != nil {
			if options.FailFunc != nil {
				options.FailFunc(r, err)
			} else {
				writeErrorResponse(r, err)
			}
			return
		}

		_ = getSaTokenContext(r, mgr)
		r.Middleware.Next()
	}
}

// AuthMiddleware checks login status AuthMiddleware 校验登录状态
func AuthMiddleware(ctx context.Context, opts ...AuthOption) ghttp.HandlerFunc {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(r *ghttp.Request) {
		mgr, err := stputil.GetManager(options.AuthType)
		if err != nil {
			if options.FailFunc != nil {
				options.FailFunc(r, err)
			} else {
				writeErrorResponse(r, err)
			}
			return
		}

		saCtx := getSaTokenContext(r, mgr)
		tokenValue := saCtx.GetTokenValue()

		if !stputil.IsLogin(ctx, tokenValue) {
			if options.FailFunc != nil {
				options.FailFunc(r, serror.ErrTokenExpired)
			} else {
				writeErrorResponse(r, serror.ErrTokenExpired)
			}
			return
		}

		r.Middleware.Next()
	}
}

// PermissionMiddleware checks permissions PermissionMiddleware 校验权限
func PermissionMiddleware(
	ctx context.Context,
	permissions []string,
	opts ...AuthOption,
) ghttp.HandlerFunc {

	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(r *ghttp.Request) {
		if len(permissions) == 0 {
			r.Middleware.Next()
			return
		}

		mgr, err := stputil.GetManager(options.AuthType)
		if err != nil {
			if options.FailFunc != nil {
				options.FailFunc(r, err)
			} else {
				writeErrorResponse(r, err)
			}
			return
		}

		saCtx := getSaTokenContext(r, mgr)
		tokenValue := saCtx.GetTokenValue()

		var ok bool
		if options.LogicType == LogicAnd {
			ok = mgr.HasPermissionsAndByToken(ctx, tokenValue, permissions)
		} else {
			ok = mgr.HasPermissionsOrByToken(ctx, tokenValue, permissions)
		}

		if !ok {
			if options.FailFunc != nil {
				options.FailFunc(r, serror.ErrPermissionDenied)
			} else {
				writeErrorResponse(r, serror.ErrPermissionDenied)
			}
			return
		}

		r.Middleware.Next()
	}
}

// PermissionPathMiddleware checks path permissions PermissionPathMiddleware 基于路径校验权限
func PermissionPathMiddleware(
	ctx context.Context,
	permissions []string,
	opts ...AuthOption,
) ghttp.HandlerFunc {

	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(r *ghttp.Request) {
		reqPermissions := append([]string{}, permissions...)
		reqPermissions = append(reqPermissions, r.URL.Path)

		if len(reqPermissions) == 0 {
			r.Middleware.Next()
			return
		}

		mgr, err := stputil.GetManager(options.AuthType)
		if err != nil {
			if options.FailFunc != nil {
				options.FailFunc(r, err)
			} else {
				writeErrorResponse(r, err)
			}
			return
		}

		saCtx := getSaTokenContext(r, mgr)
		tokenValue := saCtx.GetTokenValue()

		var ok bool
		if options.LogicType == LogicAnd {
			ok = mgr.HasPermissionsAndByToken(ctx, tokenValue, reqPermissions)
		} else {
			ok = mgr.HasPermissionsOrByToken(ctx, tokenValue, reqPermissions)
		}

		if !ok {
			if options.FailFunc != nil {
				options.FailFunc(r, serror.ErrPermissionDenied)
			} else {
				writeErrorResponse(r, serror.ErrPermissionDenied)
			}
			return
		}

		r.Middleware.Next()
	}
}

// RoleMiddleware checks roles RoleMiddleware 校验角色
func RoleMiddleware(
	ctx context.Context,
	roles []string,
	opts ...AuthOption,
) ghttp.HandlerFunc {

	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(r *ghttp.Request) {
		if len(roles) == 0 {
			r.Middleware.Next()
			return
		}

		mgr, err := stputil.GetManager(options.AuthType)
		if err != nil {
			if options.FailFunc != nil {
				options.FailFunc(r, err)
			} else {
				writeErrorResponse(r, err)
			}
			return
		}

		saCtx := getSaTokenContext(r, mgr)
		tokenValue := saCtx.GetTokenValue()

		var ok bool
		if options.LogicType == LogicAnd {
			ok = mgr.HasRolesAndByToken(ctx, tokenValue, roles)
		} else {
			ok = mgr.HasRolesOrByToken(ctx, tokenValue, roles)
		}

		if !ok {
			if options.FailFunc != nil {
				options.FailFunc(r, serror.ErrRoleDenied)
			} else {
				writeErrorResponse(r, serror.ErrRoleDenied)
			}
			return
		}

		r.Middleware.Next()
	}
}

// GetSaTokenContext gets cached SaToken context GetSaTokenContext 获取缓存的 SaToken 上下文
func GetSaTokenContext(r *ghttp.Request) (*corecontext.SaTokenContext, bool) {
	v := r.GetCtxVar(SaTokenCtxKey)
	if v == nil {
		return nil, false
	}

	ctx, ok := v.Val().(*corecontext.SaTokenContext)
	return ctx, ok
}

// GetSaTokenContextByCtx gets SaToken context by context GetSaTokenContextByCtx 从上下文获取 SaToken 上下文
func GetSaTokenContextByCtx(ctx context.Context) (*corecontext.SaTokenContext, bool) {
	request := g.RequestFromCtx(ctx)
	ctxVar := request.GetCtxVar(SaTokenCtxKey)
	if ctxVar == nil {
		return nil, false
	}

	tokenContext, ok := ctxVar.Val().(*corecontext.SaTokenContext)
	return tokenContext, ok
}

// GetLoginIDByCtx gets login ID by context GetLoginIDByCtx 从上下文获取登录 ID
func GetLoginIDByCtx(ctx context.Context, authType ...string) (string, error) {
	mgr, err := stputil.GetManager(authType...)
	if err != nil {
		return "", err
	}

	tokenValue := getSaTokenContext(g.RequestFromCtx(ctx), mgr).GetTokenValue()
	return mgr.GetLoginID(ctx, tokenValue)
}

// GetTokenInfoByCtx gets token info by context GetTokenInfoByCtx 从上下文获取 Token 信息
func GetTokenInfoByCtx(ctx context.Context, authType ...string) (*manager.TokenInfo, error) {
	mgr, err := stputil.GetManager(authType...)
	if err != nil {
		return nil, err
	}

	return mgr.GetTokenInfo(ctx, getSaTokenContext(g.RequestFromCtx(ctx), mgr).GetTokenValue())
}

// getSaTokenContext gets or creates sa-token context getSaTokenContext 获取或创建 SaToken 上下文
func getSaTokenContext(r *ghttp.Request, mgr *manager.Manager) *corecontext.SaTokenContext {
	if v := r.GetCtxVar(SaTokenCtxKey); v != nil {
		if saCtx, ok := v.Val().(*corecontext.SaTokenContext); ok {
			return saCtx
		}
	}

	saCtx := corecontext.NewContext(NewGFContext(r), mgr)
	r.SetCtxVar(SaTokenCtxKey, saCtx)

	return saCtx
}

// writeErrorResponse writes error response writeErrorResponse 写入错误响应
func writeErrorResponse(r *ghttp.Request, err error) {
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

	r.Response.WriteStatusExit(httpStatus, g.Map{
		"code":    code,
		"message": message,
		"data":    err.Error(),
	})
}

// writeSuccessResponse writes success response writeSuccessResponse 写入成功响应
func writeSuccessResponse(r *ghttp.Request, data interface{}) {
	r.Response.WriteStatusExit(http.StatusOK, g.Map{
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
