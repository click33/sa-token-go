package hertz

import (
	"context"
	"errors"
	"net/http"

	corecontext "github.com/click33/sa-token-go/core/context"
	"github.com/click33/sa-token-go/core/manager"
	"github.com/click33/sa-token-go/core/serror"
	"github.com/click33/sa-token-go/stputil"
	hertzapp "github.com/cloudwego/hertz/pkg/app"
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
	FailFunc  func(c context.Context, ctx *hertzapp.RequestContext, err error)
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
func WithFailFunc(fn func(c context.Context, ctx *hertzapp.RequestContext, err error)) AuthOption {
	return func(o *AuthOptions) {
		o.FailFunc = fn
	}
}

// RegisterSaTokenContextMiddleware registers SaToken context middleware RegisterSaTokenContextMiddleware 注册 SaToken 上下文中间件
func RegisterSaTokenContextMiddleware(ctx context.Context, opts ...AuthOption) hertzapp.HandlerFunc {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(c context.Context, reqCtx *hertzapp.RequestContext) {
		mgr, err := stputil.GetManager(options.AuthType)
		if err != nil {
			if options.FailFunc != nil {
				options.FailFunc(c, reqCtx, err)
			} else {
				writeErrorResponse(reqCtx, err)
			}
			reqCtx.Abort()
			return
		}

		_ = getSaTokenContext(reqCtx, mgr)
		reqCtx.Next(c)
	}
}

// AuthMiddleware checks login status AuthMiddleware 校验登录状态
func AuthMiddleware(ctx context.Context, opts ...AuthOption) hertzapp.HandlerFunc {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(c context.Context, reqCtx *hertzapp.RequestContext) {
		mgr, err := stputil.GetManager(options.AuthType)
		if err != nil {
			if options.FailFunc != nil {
				options.FailFunc(c, reqCtx, err)
			} else {
				writeErrorResponse(reqCtx, err)
			}
			reqCtx.Abort()
			return
		}

		saCtx := getSaTokenContext(reqCtx, mgr)
		tokenValue := saCtx.GetTokenValue()

		if !mgr.IsLogin(ctx, tokenValue) {
			if options.FailFunc != nil {
				options.FailFunc(c, reqCtx, serror.ErrTokenExpired)
			} else {
				writeErrorResponse(reqCtx, serror.ErrTokenExpired)
			}
			reqCtx.Abort()
			return
		}

		reqCtx.Next(c)
	}
}

// PermissionMiddleware checks permissions PermissionMiddleware 校验权限
func PermissionMiddleware(
	ctx context.Context,
	permissions []string,
	opts ...AuthOption,
) hertzapp.HandlerFunc {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(c context.Context, reqCtx *hertzapp.RequestContext) {
		if len(permissions) == 0 {
			reqCtx.Next(c)
			return
		}

		mgr, err := stputil.GetManager(options.AuthType)
		if err != nil {
			if options.FailFunc != nil {
				options.FailFunc(c, reqCtx, err)
			} else {
				writeErrorResponse(reqCtx, err)
			}
			reqCtx.Abort()
			return
		}

		saCtx := getSaTokenContext(reqCtx, mgr)
		tokenValue := saCtx.GetTokenValue()

		var ok bool
		if options.LogicType == LogicAnd {
			ok = mgr.HasPermissionsAndByToken(ctx, tokenValue, permissions)
		} else {
			ok = mgr.HasPermissionsOrByToken(ctx, tokenValue, permissions)
		}

		if !ok {
			if options.FailFunc != nil {
				options.FailFunc(c, reqCtx, serror.ErrPermissionDenied)
			} else {
				writeErrorResponse(reqCtx, serror.ErrPermissionDenied)
			}
			reqCtx.Abort()
			return
		}

		reqCtx.Next(c)
	}
}

// RoleMiddleware checks roles RoleMiddleware 校验角色
func RoleMiddleware(
	ctx context.Context,
	roles []string,
	opts ...AuthOption,
) hertzapp.HandlerFunc {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(c context.Context, reqCtx *hertzapp.RequestContext) {
		if len(roles) == 0 {
			reqCtx.Next(c)
			return
		}

		mgr, err := stputil.GetManager(options.AuthType)
		if err != nil {
			if options.FailFunc != nil {
				options.FailFunc(c, reqCtx, err)
			} else {
				writeErrorResponse(reqCtx, err)
			}
			reqCtx.Abort()
			return
		}

		saCtx := getSaTokenContext(reqCtx, mgr)
		tokenValue := saCtx.GetTokenValue()

		var ok bool
		if options.LogicType == LogicAnd {
			ok = mgr.HasRolesAndByToken(ctx, tokenValue, roles)
		} else {
			ok = mgr.HasRolesOrByToken(ctx, tokenValue, roles)
		}

		if !ok {
			if options.FailFunc != nil {
				options.FailFunc(c, reqCtx, serror.ErrRoleDenied)
			} else {
				writeErrorResponse(reqCtx, serror.ErrRoleDenied)
			}
			reqCtx.Abort()
			return
		}

		reqCtx.Next(c)
	}
}

// GetSaTokenContext gets cached SaToken context GetSaTokenContext 获取缓存的 SaToken 上下文
func GetSaTokenContext(ctx *hertzapp.RequestContext) (*corecontext.SaTokenContext, bool) {
	value, exists := ctx.Get(SaTokenCtxKey)
	if !exists {
		return nil, false
	}

	saCtx, ok := value.(*corecontext.SaTokenContext)
	return saCtx, ok
}

// getSaTokenContext gets or creates sa-token context getSaTokenContext 获取或创建 SaToken 上下文
func getSaTokenContext(ctx *hertzapp.RequestContext, mgr *manager.Manager) *corecontext.SaTokenContext {
	if value, exists := ctx.Get(SaTokenCtxKey); exists {
		if saCtx, ok := value.(*corecontext.SaTokenContext); ok {
			return saCtx
		}
	}

	saCtx := corecontext.NewContext(NewHertzContext(ctx), mgr)
	ctx.Set(SaTokenCtxKey, saCtx)
	return saCtx
}

// writeErrorResponse writes error response writeErrorResponse 写入错误响应
func writeErrorResponse(ctx *hertzapp.RequestContext, err error) {
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

	ctx.JSON(httpStatus, map[string]interface{}{
		"code":    code,
		"message": message,
		"data":    err.Error(),
	})
}

// writeSuccessResponse writes success response writeSuccessResponse 写入成功响应
func writeSuccessResponse(ctx *hertzapp.RequestContext, data interface{}) {
	ctx.JSON(http.StatusOK, map[string]interface{}{
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
