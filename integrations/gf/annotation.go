package gf

import (
	"context"

	"github.com/click33/sa-token-go/core/serror"
	"github.com/click33/sa-token-go/stputil"
	"github.com/gogf/gf/v2/net/ghttp"
)

// Annotation defines annotation config Annotation 定义注解配置
type Annotation struct {
	// AuthType specifies auth type AuthType 指定认证类型
	AuthType string `json:"authType"`
	// CheckLogin indicates login check CheckLogin 表示是否检查登录
	CheckLogin bool `json:"checkLogin"`
	// CheckRole lists required roles CheckRole 列出需要校验的角色
	CheckRole []string `json:"checkRole"`
	// CheckPermission lists required permissions CheckPermission 列出需要校验的权限
	CheckPermission []string `json:"checkPermission"`
	// CheckDisable indicates disable check CheckDisable 表示是否检查封禁
	CheckDisable bool `json:"checkDisable"`
	// Ignore bypasses authentication Ignore 表示是否忽略认证
	Ignore bool `json:"ignore"`
	// LogicType sets logic mode LogicType 指定逻辑类型
	LogicType LogicType `json:"logicType"`
}

// GetHandler gets annotation handler GetHandler 获取注解处理器
func GetHandler(ctx context.Context, handler ghttp.HandlerFunc, failFunc func(r *ghttp.Request, err error), annotations ...*Annotation) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		if len(annotations) > 0 && annotations[0].Ignore {
			if handler != nil {
				handler(r)
			} else {
				r.Middleware.Next()
			}
			return
		}

		ann := &Annotation{}
		if len(annotations) > 0 {
			ann = annotations[0]
		}

		needAuth := ann.CheckLogin || ann.CheckDisable || len(ann.CheckPermission) > 0 || len(ann.CheckRole) > 0
		if !needAuth {
			if handler != nil {
				handler(r)
			} else {
				r.Middleware.Next()
			}
			return
		}

		mgr, err := stputil.GetManager(ann.AuthType)
		if err != nil {
			if failFunc != nil {
				failFunc(r, err)
			} else {
				writeErrorResponse(r, err)
			}
			return
		}

		// Get SaTokenContext (reuse cached context) 获取 SaTokenContext（复用缓存上下文）
		saCtx := getSaTokenContext(r, mgr)
		token := saCtx.GetTokenValue()

		if !stputil.IsLogin(ctx, token) {
			if failFunc != nil {
				failFunc(r, serror.ErrNotLogin)
			} else {
				writeErrorResponse(r, serror.ErrNotLogin)
			}
			return
		}

		var loginID string
		if ann.CheckDisable || len(ann.CheckPermission) > 0 || len(ann.CheckRole) > 0 {
			loginID, err = mgr.GetLoginID(ctx, token)
			if err != nil {
				if failFunc != nil {
					failFunc(r, err)
				} else {
					writeErrorResponse(r, err)
				}
				return
			}
		}

		if ann.CheckDisable {
			if stputil.IsDisable(ctx, loginID) {
				if failFunc != nil {
					failFunc(r, serror.ErrAccountDisabled)
				} else {
					writeErrorResponse(r, serror.ErrAccountDisabled)
				}
				return
			}
		}

		if len(ann.CheckPermission) > 0 {
			var ok bool
			if ann.LogicType == LogicAnd {
				ok = stputil.HasPermissionsAnd(ctx, loginID, ann.CheckPermission)
			} else {
				ok = stputil.HasPermissionsOr(ctx, loginID, ann.CheckPermission)
			}
			if !ok {
				if failFunc != nil {
					failFunc(r, serror.ErrPermissionDenied)
				} else {
					writeErrorResponse(r, serror.ErrPermissionDenied)
				}
				return
			}
		}

		if len(ann.CheckRole) > 0 {
			var ok bool
			if ann.LogicType == LogicAnd {
				ok = stputil.HasRolesAnd(ctx, loginID, ann.CheckRole)
			} else {
				ok = stputil.HasRolesOr(ctx, loginID, ann.CheckRole)
			}
			if !ok {
				if failFunc != nil {
					failFunc(r, serror.ErrRoleDenied)
				} else {
					writeErrorResponse(r, serror.ErrRoleDenied)
				}
				return
			}
		}

		if handler != nil {
			handler(r)
		} else {
			r.Middleware.Next()
		}
	}
}

// CheckLoginMiddleware creates login check middleware CheckLoginMiddleware 生成登录检查中间件
func CheckLoginMiddleware(
	ctx context.Context,
	handler ghttp.HandlerFunc,
	failFunc func(r *ghttp.Request, err error),
	authType ...string,
) ghttp.HandlerFunc {
	ann := &Annotation{CheckLogin: true}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// CheckRoleMiddleware creates role check middleware CheckRoleMiddleware 生成角色检查中间件
func CheckRoleMiddleware(
	ctx context.Context,
	roles []string,
	handler ghttp.HandlerFunc,
	failFunc func(r *ghttp.Request, err error),
	authType ...string,
) ghttp.HandlerFunc {
	ann := &Annotation{CheckRole: roles}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// CheckPermissionMiddleware creates permission check middleware CheckPermissionMiddleware 生成权限检查中间件
func CheckPermissionMiddleware(
	ctx context.Context,
	perms []string,
	handler ghttp.HandlerFunc,
	failFunc func(r *ghttp.Request, err error),
	authType ...string,
) ghttp.HandlerFunc {
	ann := &Annotation{CheckPermission: perms}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// CheckDisableMiddleware creates disable check middleware CheckDisableMiddleware 生成封禁检查中间件
func CheckDisableMiddleware(
	ctx context.Context,
	handler ghttp.HandlerFunc,
	failFunc func(r *ghttp.Request, err error),
	authType ...string,
) ghttp.HandlerFunc {
	ann := &Annotation{CheckDisable: true}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// IgnoreMiddleware creates ignore auth middleware IgnoreMiddleware 生成忽略认证中间件
func IgnoreMiddleware(
	ctx context.Context,
	handler ghttp.HandlerFunc,
	failFunc func(r *ghttp.Request, err error),
) ghttp.HandlerFunc {
	ann := &Annotation{Ignore: true}
	return GetHandler(ctx, handler, failFunc, ann)
}

// CheckLoginAndRoleMiddleware creates login and role middleware CheckLoginAndRoleMiddleware 生成登录与角色检查中间件
func CheckLoginAndRoleMiddleware(
	ctx context.Context,
	roles []string,
	handler ghttp.HandlerFunc,
	failFunc func(r *ghttp.Request, err error),
	authType ...string,
) ghttp.HandlerFunc {
	ann := &Annotation{CheckLogin: true, CheckRole: roles}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// CheckLoginAndPermissionMiddleware creates login and permission middleware CheckLoginAndPermissionMiddleware 生成登录与权限检查中间件
func CheckLoginAndPermissionMiddleware(
	ctx context.Context,
	perms []string,
	handler ghttp.HandlerFunc,
	failFunc func(r *ghttp.Request, err error),
	authType ...string,
) ghttp.HandlerFunc {
	ann := &Annotation{CheckLogin: true, CheckPermission: perms}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// CheckAllMiddleware creates combined auth middleware CheckAllMiddleware 生成全部检查中间件
func CheckAllMiddleware(
	ctx context.Context,
	roles []string,
	perms []string,
	handler ghttp.HandlerFunc,
	failFunc func(r *ghttp.Request, err error),
	authType ...string,
) ghttp.HandlerFunc {
	ann := &Annotation{CheckLogin: true, CheckRole: roles, CheckPermission: perms}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// AuthGroup creates auth route group AuthGroup 创建认证路由组
func AuthGroup(
	ctx context.Context,
	group *ghttp.RouterGroup,
	handler ghttp.HandlerFunc,
	failFunc func(r *ghttp.Request, err error),
	authType ...string,
) *ghttp.RouterGroup {
	group.Middleware(CheckLoginMiddleware(ctx, handler, failFunc, authType...))
	return group
}

// RoleGroup creates role route group RoleGroup 创建角色路由组
func RoleGroup(
	ctx context.Context,
	group *ghttp.RouterGroup,
	roles []string,
	handler ghttp.HandlerFunc,
	failFunc func(r *ghttp.Request, err error),
	authType ...string,
) *ghttp.RouterGroup {
	group.Middleware(CheckLoginAndRoleMiddleware(ctx, roles, handler, failFunc, authType...))
	return group
}

// PermissionGroup creates permission route group PermissionGroup 创建权限路由组
func PermissionGroup(
	ctx context.Context,
	group *ghttp.RouterGroup,
	perms []string,
	handler ghttp.HandlerFunc,
	failFunc func(r *ghttp.Request, err error),
	authType ...string,
) *ghttp.RouterGroup {
	group.Middleware(CheckLoginAndPermissionMiddleware(ctx, perms, handler, failFunc, authType...))
	return group
}

// RoleAndPermissionGroup creates role and permission route group RoleAndPermissionGroup 创建角色与权限路由组
func RoleAndPermissionGroup(
	ctx context.Context,
	group *ghttp.RouterGroup,
	roles []string,
	perms []string,
	handler ghttp.HandlerFunc,
	failFunc func(r *ghttp.Request, err error),
	authType ...string,
) *ghttp.RouterGroup {
	group.Middleware(CheckAllMiddleware(ctx, roles, perms, handler, failFunc, authType...))
	return group
}
