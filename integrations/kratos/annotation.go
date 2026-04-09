package kratos

import (
	"context"

	"github.com/click33/sa-token-go/core/serror"
	"github.com/click33/sa-token-go/stputil"
	"github.com/go-kratos/kratos/v2/middleware"
)

// Annotation defines annotation config Annotation 定义注解配置
type Annotation struct {
	AuthType        string    `json:"authType"`
	CheckLogin      bool      `json:"checkLogin"`
	CheckRole       []string  `json:"checkRole"`
	CheckPermission []string  `json:"checkPermission"`
	CheckDisable    bool      `json:"checkDisable"`
	Ignore          bool      `json:"ignore"`
	LogicType       LogicType `json:"logicType"`
}

// GetHandler gets annotation middleware GetHandler 获取注解中间件
func GetHandler(failFunc FailFunc, annotations ...*Annotation) middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			if len(annotations) > 0 && annotations[0].Ignore {
				return next(ctx, req)
			}

			ann := &Annotation{LogicType: LogicAnd}
			if len(annotations) > 0 && annotations[0] != nil {
				ann = annotations[0]
				if ann.LogicType == "" {
					ann.LogicType = LogicAnd
				}
			}

			needAuth := ann.CheckLogin || ann.CheckDisable || len(ann.CheckPermission) > 0 || len(ann.CheckRole) > 0
			if !needAuth {
				return next(ctx, req)
			}

			mgr, err := stputil.GetManager(ann.AuthType)
			if err != nil {
				return nil, dispatchFail(ctx, failFunc, err)
			}

			saCtx, ctx := getSaTokenContext(ctx, mgr)
			tokenValue := saCtx.GetTokenValue()
			if !mgr.IsLogin(ctx, tokenValue) {
				return nil, dispatchFail(ctx, failFunc, serror.ErrNotLogin)
			}

			var loginID string
			if ann.CheckDisable || len(ann.CheckPermission) > 0 || len(ann.CheckRole) > 0 {
				loginID, err = mgr.GetLoginID(ctx, tokenValue)
				if err != nil {
					return nil, dispatchFail(ctx, failFunc, err)
				}
			}

			if ann.CheckDisable && mgr.IsDisable(ctx, loginID) {
				return nil, dispatchFail(ctx, failFunc, serror.ErrAccountDisabled)
			}

			if len(ann.CheckPermission) > 0 {
				var ok bool
				if ann.LogicType == LogicAnd {
					ok = mgr.HasPermissionsAnd(ctx, loginID, ann.CheckPermission)
				} else {
					ok = mgr.HasPermissionsOr(ctx, loginID, ann.CheckPermission)
				}
				if !ok {
					return nil, dispatchFail(ctx, failFunc, serror.ErrPermissionDenied)
				}
			}

			if len(ann.CheckRole) > 0 {
				var ok bool
				if ann.LogicType == LogicAnd {
					ok = mgr.HasRolesAnd(ctx, loginID, ann.CheckRole)
				} else {
					ok = mgr.HasRolesOr(ctx, loginID, ann.CheckRole)
				}
				if !ok {
					return nil, dispatchFail(ctx, failFunc, serror.ErrRoleDenied)
				}
			}

			return next(ctx, req)
		}
	}
}

// CheckLoginMiddleware creates login middleware CheckLoginMiddleware 创建登录校验中间件
func CheckLoginMiddleware(failFunc FailFunc, authType ...string) middleware.Middleware {
	ann := &Annotation{CheckLogin: true}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(failFunc, ann)
}

// CheckRoleMiddleware creates role middleware CheckRoleMiddleware 创建角色校验中间件
func CheckRoleMiddleware(roles []string, failFunc FailFunc, authType ...string) middleware.Middleware {
	ann := &Annotation{CheckRole: roles, LogicType: LogicAnd}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(failFunc, ann)
}

// CheckPermissionMiddleware creates permission middleware CheckPermissionMiddleware 创建权限校验中间件
func CheckPermissionMiddleware(perms []string, failFunc FailFunc, authType ...string) middleware.Middleware {
	ann := &Annotation{CheckPermission: perms, LogicType: LogicAnd}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(failFunc, ann)
}

// CheckDisableMiddleware creates disable middleware CheckDisableMiddleware 创建封禁校验中间件
func CheckDisableMiddleware(failFunc FailFunc, authType ...string) middleware.Middleware {
	ann := &Annotation{CheckDisable: true}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(failFunc, ann)
}

// IgnoreMiddleware creates ignore middleware IgnoreMiddleware 创建忽略认证中间件
func IgnoreMiddleware(failFunc FailFunc) middleware.Middleware {
	return GetHandler(failFunc, &Annotation{Ignore: true})
}

// CheckLoginAndRoleMiddleware creates login and role middleware CheckLoginAndRoleMiddleware 创建登录与角色校验中间件
func CheckLoginAndRoleMiddleware(roles []string, failFunc FailFunc, authType ...string) middleware.Middleware {
	ann := &Annotation{CheckLogin: true, CheckRole: roles, LogicType: LogicAnd}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(failFunc, ann)
}

// CheckLoginAndPermissionMiddleware creates login and permission middleware CheckLoginAndPermissionMiddleware 创建登录与权限校验中间件
func CheckLoginAndPermissionMiddleware(perms []string, failFunc FailFunc, authType ...string) middleware.Middleware {
	ann := &Annotation{CheckLogin: true, CheckPermission: perms, LogicType: LogicAnd}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(failFunc, ann)
}

// CheckAllMiddleware creates combined middleware CheckAllMiddleware 创建组合校验中间件
func CheckAllMiddleware(roles []string, perms []string, failFunc FailFunc, authType ...string) middleware.Middleware {
	ann := &Annotation{
		CheckLogin:      true,
		CheckRole:       roles,
		CheckPermission: perms,
		CheckDisable:    true,
		LogicType:       LogicAnd,
	}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(failFunc, ann)
}

// GetLoginIDFromRequest gets login ID from request context GetLoginIDFromRequest 从请求上下文获取登录 ID
func GetLoginIDFromRequest(ctx context.Context, authType ...string) (string, error) {
	return GetLoginIDByCtx(ctx, authType...)
}

// IsLoginFromRequest checks login state from request context IsLoginFromRequest 从请求上下文检查登录状态
func IsLoginFromRequest(ctx context.Context, authType ...string) bool {
	mgr, err := stputil.GetManager(authType...)
	if err != nil {
		return false
	}

	saCtx, ctx := getSaTokenContext(ctx, mgr)
	return mgr.IsLogin(ctx, saCtx.GetTokenValue())
}

// GetTokenFromRequest gets token from request context GetTokenFromRequest 从请求上下文获取 Token
func GetTokenFromRequest(ctx context.Context, authType ...string) string {
	mgr, err := stputil.GetManager(authType...)
	if err != nil {
		return ""
	}

	saCtx, _ := getSaTokenContext(ctx, mgr)
	return saCtx.GetTokenValue()
}

// WithContext returns request context WithContext 返回请求上下文
func WithContext(ctx context.Context, authType ...string) context.Context {
	return ctx
}
