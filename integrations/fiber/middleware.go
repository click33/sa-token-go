package fiber

import (
	"context"
	"errors"
	"net/http"

	corecontext "github.com/click33/sa-token-go/core/context"
	"github.com/click33/sa-token-go/core/manager"
	"github.com/click33/sa-token-go/core/serror"
	"github.com/click33/sa-token-go/stputil"
	gofiber "github.com/gofiber/fiber/v2"
)

// LogicType defines permission and role check logic LogicType 定义权限与角色校验的逻辑类型。
type LogicType string

const (
	// SaTokenCtxKey stores request scoped SaToken context SaTokenCtxKey 存储请求级 SaToken 上下文
	SaTokenCtxKey = "SaTokenCtx"

	LogicOr  LogicType = "OR"
	LogicAnd LogicType = "AND"
)

// AuthOption defines auth option setter AuthOption 定义认证选项设置器
type AuthOption func(*AuthOptions)

// AuthOptions carries middleware auth options AuthOptions 保存中间件认证选项。
type AuthOptions struct {
	AuthType  string
	LogicType LogicType
	FailFunc  func(c *gofiber.Ctx, err error)
}

// defaultAuthOptions returns default middleware options defaultAuthOptions 返回默认中间件选项。
func defaultAuthOptions() *AuthOptions {
	return &AuthOptions{LogicType: LogicAnd}
}

// WithAuthType sets the auth type used by middleware WithAuthType 设置中间件使用的认证类型。
func WithAuthType(authType string) AuthOption {
	return func(o *AuthOptions) {
		o.AuthType = authType
	}
}

// WithLogicType sets the logic mode for permission and role checks WithLogicType 设置权限与角色校验的逻辑模式。
func WithLogicType(logicType LogicType) AuthOption {
	return func(o *AuthOptions) {
		o.LogicType = logicType
	}
}

// WithFailFunc sets a custom auth failure callback WithFailFunc 设置自定义认证失败回调。
func WithFailFunc(fn func(c *gofiber.Ctx, err error)) AuthOption {
	return func(o *AuthOptions) {
		o.FailFunc = fn
	}
}

// RegisterSaTokenContextMiddleware initializes SaToken context for each request RegisterSaTokenContextMiddleware 为每个请求初始化 SaToken 上下文。
func RegisterSaTokenContextMiddleware(ctx context.Context, opts ...AuthOption) gofiber.Handler {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(c *gofiber.Ctx) error {
		mgr, err := stputil.GetManager(options.AuthType)
		if err != nil {
			if options.FailFunc != nil {
				options.FailFunc(c, err)
				return nil
			}
			return writeErrorResponse(c, err)
		}

		_ = getSaTokenContext(c, mgr)
		return c.Next()
	}
}

// AuthMiddleware checks whether the current request is authenticated AuthMiddleware 检查当前请求是否已认证。
func AuthMiddleware(ctx context.Context, opts ...AuthOption) gofiber.Handler {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(c *gofiber.Ctx) error {
		mgr, err := stputil.GetManager(options.AuthType)
		if err != nil {
			if options.FailFunc != nil {
				options.FailFunc(c, err)
				return nil
			}
			return writeErrorResponse(c, err)
		}

		saCtx := getSaTokenContext(c, mgr)
		tokenValue := saCtx.GetTokenValue()
		if !mgr.IsLogin(ctx, tokenValue) {
			if options.FailFunc != nil {
				options.FailFunc(c, serror.ErrTokenExpired)
				return nil
			}
			return writeErrorResponse(c, serror.ErrTokenExpired)
		}

		return c.Next()
	}
}

// PermissionMiddleware checks whether the current token has required permissions PermissionMiddleware 检查当前 token 是否具备所需权限
func PermissionMiddleware(
	ctx context.Context,
	permissions []string,
	opts ...AuthOption,
) gofiber.Handler {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(c *gofiber.Ctx) error {
		if len(permissions) == 0 {
			return c.Next()
		}

		mgr, err := stputil.GetManager(options.AuthType)
		if err != nil {
			if options.FailFunc != nil {
				options.FailFunc(c, err)
				return nil
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
				options.FailFunc(c, serror.ErrPermissionDenied)
				return nil
			}
			return writeErrorResponse(c, serror.ErrPermissionDenied)
		}

		return c.Next()
	}
}

// RoleMiddleware checks whether the current token has required roles RoleMiddleware 检查当前 token 是否具备所需角色
func RoleMiddleware(
	ctx context.Context,
	roles []string,
	opts ...AuthOption,
) gofiber.Handler {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(c *gofiber.Ctx) error {
		if len(roles) == 0 {
			return c.Next()
		}

		mgr, err := stputil.GetManager(options.AuthType)
		if err != nil {
			if options.FailFunc != nil {
				options.FailFunc(c, err)
				return nil
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
				options.FailFunc(c, serror.ErrRoleDenied)
				return nil
			}
			return writeErrorResponse(c, serror.ErrRoleDenied)
		}

		return c.Next()
	}
}

// GetSaTokenContext gets cached SaToken context from Fiber request GetSaTokenContext 从 Fiber 请求中获取缓存的 SaToken 上下文。
func GetSaTokenContext(c *gofiber.Ctx) (*corecontext.SaTokenContext, bool) {
	value := c.Locals(SaTokenCtxKey)
	if value == nil {
		return nil, false
	}

	saCtx, ok := value.(*corecontext.SaTokenContext)
	return saCtx, ok
}

// getSaTokenContext gets or creates sa-token context getSaTokenContext 获取或创建 SaToken 上下文
func getSaTokenContext(c *gofiber.Ctx, mgr *manager.Manager) *corecontext.SaTokenContext {
	if value := c.Locals(SaTokenCtxKey); value != nil {
		if saCtx, ok := value.(*corecontext.SaTokenContext); ok {
			return saCtx
		}
	}

	saCtx := corecontext.NewContext(NewFiberContext(c), mgr)
	c.Locals(SaTokenCtxKey, saCtx)
	return saCtx
}

// writeErrorResponse writes a standard error response writeErrorResponse 写入标准错误响应。
func writeErrorResponse(c *gofiber.Ctx, err error) error {
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

	return c.Status(httpStatus).JSON(gofiber.Map{
		"code":    code,
		"message": message,
		"data":    err.Error(),
	})
}

// writeSuccessResponse writes a standard success response writeSuccessResponse 写入标准成功响应。
func writeSuccessResponse(c *gofiber.Ctx, data interface{}) error {
	return c.Status(http.StatusOK).JSON(gofiber.Map{
		"code":    serror.CodeSuccess,
		"message": "success",
		"data":    data,
	})
}

// getHTTPStatusFromCode maps SaToken error code to HTTP status getHTTPStatusFromCode 将 SaToken 错误码映射为 HTTP 状态码。
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
