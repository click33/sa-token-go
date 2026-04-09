package echo

import (
	"context"
	"errors"
	"net/http"

	corecontext "github.com/click33/sa-token-go/core/context"
	"github.com/click33/sa-token-go/core/manager"
	"github.com/click33/sa-token-go/core/serror"
	"github.com/click33/sa-token-go/stputil"
	echo4 "github.com/labstack/echo/v4"
)

// LogicType defines permission and role check logic LogicType 定义权限与角色校验逻辑
type LogicType string

const (
	// SaTokenCtxKey stores request scoped SaToken context SaTokenCtxKey 存储请求级 SaToken 上下文
	SaTokenCtxKey = "SaTokenCtx"

	// LogicOr uses OR logic for checks LogicOr 使用 OR 逻辑校验
	LogicOr LogicType = "OR"
	// LogicAnd uses AND logic for checks LogicAnd 使用 AND 逻辑校验
	LogicAnd LogicType = "AND"
)

// AuthOption defines auth option setter AuthOption 定义认证选项设置器
type AuthOption func(*AuthOptions)

// AuthOptions carries middleware auth options AuthOptions 保存中间件认证选项
type AuthOptions struct {
	AuthType  string
	LogicType LogicType
	FailFunc  func(c echo4.Context, err error) error
}

// defaultAuthOptions returns default middleware options defaultAuthOptions 返回默认中间件选项
func defaultAuthOptions() *AuthOptions {
	return &AuthOptions{LogicType: LogicAnd}
}

// WithAuthType sets middleware auth type WithAuthType 设置中间件认证类型
func WithAuthType(authType string) AuthOption {
	return func(o *AuthOptions) {
		o.AuthType = authType
	}
}

// WithLogicType sets middleware logic type WithLogicType 设置中间件逻辑类型
func WithLogicType(logicType LogicType) AuthOption {
	return func(o *AuthOptions) {
		o.LogicType = logicType
	}
}

// WithFailFunc sets auth failure callback WithFailFunc 设置认证失败回调
func WithFailFunc(fn func(c echo4.Context, err error) error) AuthOption {
	return func(o *AuthOptions) {
		o.FailFunc = fn
	}
}

// RegisterSaTokenContextMiddleware initializes SaToken context per request RegisterSaTokenContextMiddleware 初始化每个请求的 SaToken 上下文
func RegisterSaTokenContextMiddleware(ctx context.Context, opts ...AuthOption) echo4.MiddlewareFunc {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(next echo4.HandlerFunc) echo4.HandlerFunc {
		return func(c echo4.Context) error {
			mgr, err := stputil.GetManager(options.AuthType)
			if err != nil {
				if options.FailFunc != nil {
					return options.FailFunc(c, err)
				}
				return writeErrorResponse(c, err)
			}

			_ = getSaTokenContext(c, mgr)
			return next(c)
		}
	}
}

// AuthMiddleware checks whether current request is authenticated AuthMiddleware 校验当前请求是否已认证
func AuthMiddleware(ctx context.Context, opts ...AuthOption) echo4.MiddlewareFunc {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(next echo4.HandlerFunc) echo4.HandlerFunc {
		return func(c echo4.Context) error {
			mgr, err := stputil.GetManager(options.AuthType)
			if err != nil {
				if options.FailFunc != nil {
					return options.FailFunc(c, err)
				}
				return writeErrorResponse(c, err)
			}

			saCtx := getSaTokenContext(c, mgr)
			tokenValue := saCtx.GetTokenValue()
			if !mgr.IsLogin(ctx, tokenValue) {
				if options.FailFunc != nil {
					return options.FailFunc(c, serror.ErrTokenExpired)
				}
				return writeErrorResponse(c, serror.ErrTokenExpired)
			}

			return next(c)
		}
	}
}

// PermissionMiddleware checks whether current token has required permissions PermissionMiddleware 校验当前 token 是否具备所需权限
func PermissionMiddleware(ctx context.Context, permissions []string, opts ...AuthOption) echo4.MiddlewareFunc {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(next echo4.HandlerFunc) echo4.HandlerFunc {
		return func(c echo4.Context) error {
			if len(permissions) == 0 {
				return next(c)
			}

			mgr, err := stputil.GetManager(options.AuthType)
			if err != nil {
				if options.FailFunc != nil {
					return options.FailFunc(c, err)
				}
				return writeErrorResponse(c, err)
			}

			saCtx := getSaTokenContext(c, mgr)
			tokenValue := saCtx.GetTokenValue()

			var ok bool
			if options.LogicType == LogicAnd {
				ok = mgr.HasPermissionsAndByToken(ctx, tokenValue, permissions)
			} else {
				ok = mgr.HasPermissionsOrByToken(ctx, tokenValue, permissions)
			}

			if !ok {
				if options.FailFunc != nil {
					return options.FailFunc(c, serror.ErrPermissionDenied)
				}
				return writeErrorResponse(c, serror.ErrPermissionDenied)
			}

			return next(c)
		}
	}
}

// RoleMiddleware checks whether current token has required roles RoleMiddleware 校验当前 token 是否具备所需角色
func RoleMiddleware(ctx context.Context, roles []string, opts ...AuthOption) echo4.MiddlewareFunc {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(next echo4.HandlerFunc) echo4.HandlerFunc {
		return func(c echo4.Context) error {
			if len(roles) == 0 {
				return next(c)
			}

			mgr, err := stputil.GetManager(options.AuthType)
			if err != nil {
				if options.FailFunc != nil {
					return options.FailFunc(c, err)
				}
				return writeErrorResponse(c, err)
			}

			saCtx := getSaTokenContext(c, mgr)
			tokenValue := saCtx.GetTokenValue()

			var ok bool
			if options.LogicType == LogicAnd {
				ok = mgr.HasRolesAndByToken(ctx, tokenValue, roles)
			} else {
				ok = mgr.HasRolesOrByToken(ctx, tokenValue, roles)
			}

			if !ok {
				if options.FailFunc != nil {
					return options.FailFunc(c, serror.ErrRoleDenied)
				}
				return writeErrorResponse(c, serror.ErrRoleDenied)
			}

			return next(c)
		}
	}
}

// GetSaTokenContext gets cached SaToken context from Echo request GetSaTokenContext 从 Echo 请求中获取缓存的 SaToken 上下文
func GetSaTokenContext(c echo4.Context) (*corecontext.SaTokenContext, bool) {
	value := c.Get(SaTokenCtxKey)
	if value == nil {
		return nil, false
	}

	saCtx, ok := value.(*corecontext.SaTokenContext)
	return saCtx, ok
}

// getSaTokenContext gets or creates sa-token context getSaTokenContext 获取或创建 SaToken 上下文
func getSaTokenContext(c echo4.Context, mgr *manager.Manager) *corecontext.SaTokenContext {
	if value := c.Get(SaTokenCtxKey); value != nil {
		if saCtx, ok := value.(*corecontext.SaTokenContext); ok {
			return saCtx
		}
	}

	saCtx := corecontext.NewContext(NewEchoContext(c), mgr)
	c.Set(SaTokenCtxKey, saCtx)
	return saCtx
}

// writeErrorResponse writes standard error response writeErrorResponse 写入标准错误响应
func writeErrorResponse(c echo4.Context, err error) error {
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

	return c.JSON(httpStatus, echo4.Map{
		"code":    code,
		"message": message,
		"data":    err.Error(),
	})
}

// writeSuccessResponse writes standard success response writeSuccessResponse 写入标准成功响应
func writeSuccessResponse(c echo4.Context, data interface{}) error {
	return c.JSON(http.StatusOK, echo4.Map{
		"code":    serror.CodeSuccess,
		"message": "success",
		"data":    data,
	})
}

// getHTTPStatusFromCode maps SaToken code to HTTP status getHTTPStatusFromCode 映射 SaToken 错误码到 HTTP 状态码
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
