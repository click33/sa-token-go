package gin

import (
	"context"
	"errors"
	"net/http"

	corecontext "github.com/click33/sa-token-go/core/context"
	"github.com/click33/sa-token-go/core/manager"
	"github.com/click33/sa-token-go/core/serror"
	"github.com/click33/sa-token-go/stputil"
	"github.com/gin-gonic/gin"
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
	FailFunc  func(c *gin.Context, err error)
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
func WithFailFunc(fn func(c *gin.Context, err error)) AuthOption {
	return func(o *AuthOptions) {
		o.FailFunc = fn
	}
}

// RegisterSaTokenContextMiddleware registers SaToken context middleware RegisterSaTokenContextMiddleware 注册 SaToken 上下文中间件
func RegisterSaTokenContextMiddleware(ctx context.Context, opts ...AuthOption) gin.HandlerFunc {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(c *gin.Context) {
		mgr, err := stputil.GetManager(options.AuthType)
		if err != nil {
			if options.FailFunc != nil {
				options.FailFunc(c, err)
			} else {
				writeErrorResponse(c, err)
			}
			return
		}

		_ = getSaTokenContext(c, mgr)
	}
}

// AuthMiddleware checks login status AuthMiddleware 校验登录状态
func AuthMiddleware(ctx context.Context, opts ...AuthOption) gin.HandlerFunc {
	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(c *gin.Context) {
		mgr, err := stputil.GetManager(options.AuthType)
		if err != nil {
			if options.FailFunc != nil {
				options.FailFunc(c, err)
			} else {
				writeErrorResponse(c, err)
			}
			c.Abort()
			return
		}

		saCtx := getSaTokenContext(c, mgr)
		tokenValue := saCtx.GetTokenValue()

		if !mgr.IsLogin(ctx, tokenValue) {
			if options.FailFunc != nil {
				options.FailFunc(c, serror.ErrTokenExpired)
			} else {
				writeErrorResponse(c, serror.ErrTokenExpired)
			}
			c.Abort()
			return
		}

		c.Next()
	}
}

// PermissionMiddleware checks permissions PermissionMiddleware 校验权限
func PermissionMiddleware(
	ctx context.Context,
	permissions []string,
	opts ...AuthOption,
) gin.HandlerFunc {

	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(c *gin.Context) {
		if len(permissions) == 0 {
			c.Next()
			return
		}

		mgr, err := stputil.GetManager(options.AuthType)
		if err != nil {
			if options.FailFunc != nil {
				options.FailFunc(c, err)
			} else {
				writeErrorResponse(c, err)
			}
			c.Abort()
			return
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
			} else {
				writeErrorResponse(c, serror.ErrPermissionDenied)
			}
			c.Abort()
			return
		}

		c.Next()
	}
}

// RoleMiddleware checks roles RoleMiddleware 校验角色
func RoleMiddleware(
	ctx context.Context,
	roles []string,
	opts ...AuthOption,
) gin.HandlerFunc {

	options := defaultAuthOptions()
	for _, opt := range opts {
		opt(options)
	}

	return func(c *gin.Context) {
		if len(roles) == 0 {
			c.Next()
			return
		}

		mgr, err := stputil.GetManager(options.AuthType)
		if err != nil {
			if options.FailFunc != nil {
				options.FailFunc(c, err)
			} else {
				writeErrorResponse(c, err)
			}
			c.Abort()
			return
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
			} else {
				writeErrorResponse(c, serror.ErrRoleDenied)
			}
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetSaTokenContext gets cached SaToken context GetSaTokenContext 获取缓存的 SaToken 上下文
func GetSaTokenContext(c *gin.Context) (*corecontext.SaTokenContext, bool) {
	v, exists := c.Get(SaTokenCtxKey)
	if !exists {
		return nil, false
	}

	ctx, ok := v.(*corecontext.SaTokenContext)
	return ctx, ok
}

// getSaTokenContext gets or creates sa-token context getSaTokenContext 获取或创建 SaToken 上下文
func getSaTokenContext(c *gin.Context, mgr *manager.Manager) *corecontext.SaTokenContext {
	if v, exists := c.Get(SaTokenCtxKey); exists {
		if saCtx, ok := v.(*corecontext.SaTokenContext); ok {
			return saCtx
		}
	}

	saCtx := corecontext.NewContext(NewGinContext(c), mgr)
	c.Set(SaTokenCtxKey, saCtx)

	return saCtx
}

// writeErrorResponse writes error response writeErrorResponse 写入错误响应
func writeErrorResponse(c *gin.Context, err error) {
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

	c.JSON(httpStatus, gin.H{
		"code":    code,
		"message": message,
		"data":    err.Error(),
	})
}

// writeSuccessResponse writes success response writeSuccessResponse 写入成功响应
func writeSuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
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
