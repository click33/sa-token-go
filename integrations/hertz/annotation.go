package hertz

import (
	"context"

	"github.com/click33/sa-token-go/core/serror"
	"github.com/click33/sa-token-go/stputil"
	hertzapp "github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
)

// Annotation defines annotation config Annotation 定义注解配置
type Annotation struct {
	AuthType        string
	CheckLogin      bool
	CheckRole       []string
	CheckPermission []string
	CheckDisable    bool
	Ignore          bool
	LogicType       LogicType
}

// GetHandler gets annotation handler GetHandler 获取注解处理器
func GetHandler(
	ctx context.Context,
	handler hertzapp.HandlerFunc,
	failFunc func(c context.Context, reqCtx *hertzapp.RequestContext, err error),
	annotations ...*Annotation,
) hertzapp.HandlerFunc {
	return func(c context.Context, reqCtx *hertzapp.RequestContext) {
		if len(annotations) > 0 && annotations[0].Ignore {
			if handler != nil {
				handler(c, reqCtx)
			} else {
				reqCtx.Next(c)
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
				handler(c, reqCtx)
			} else {
				reqCtx.Next(c)
			}
			return
		}

		mgr, err := stputil.GetManager(ann.AuthType)
		if err != nil {
			if failFunc != nil {
				failFunc(c, reqCtx, err)
			} else {
				writeErrorResponse(reqCtx, err)
			}
			reqCtx.Abort()
			return
		}

		// Get SaTokenContext Get SaTokenContext 获取缓存上下文
		saCtx := getSaTokenContext(reqCtx, mgr)
		token := saCtx.GetTokenValue()

		if !mgr.IsLogin(ctx, token) {
			if failFunc != nil {
				failFunc(c, reqCtx, serror.ErrNotLogin)
			} else {
				writeErrorResponse(reqCtx, serror.ErrNotLogin)
			}
			reqCtx.Abort()
			return
		}

		var loginID string
		if ann.CheckDisable || len(ann.CheckPermission) > 0 || len(ann.CheckRole) > 0 {
			loginID, err = mgr.GetLoginID(ctx, token)
			if err != nil {
				if failFunc != nil {
					failFunc(c, reqCtx, err)
				} else {
					writeErrorResponse(reqCtx, err)
				}
				reqCtx.Abort()
				return
			}
		}

		if ann.CheckDisable && mgr.IsDisable(ctx, loginID) {
			if failFunc != nil {
				failFunc(c, reqCtx, serror.ErrAccountDisabled)
			} else {
				writeErrorResponse(reqCtx, serror.ErrAccountDisabled)
			}
			reqCtx.Abort()
			return
		}

		if len(ann.CheckPermission) > 0 {
			var ok bool
			if ann.LogicType == LogicAnd {
				ok = mgr.HasPermissionsAnd(ctx, loginID, ann.CheckPermission)
			} else {
				ok = mgr.HasPermissionsOr(ctx, loginID, ann.CheckPermission)
			}
			if !ok {
				if failFunc != nil {
					failFunc(c, reqCtx, serror.ErrPermissionDenied)
				} else {
					writeErrorResponse(reqCtx, serror.ErrPermissionDenied)
				}
				reqCtx.Abort()
				return
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
				if failFunc != nil {
					failFunc(c, reqCtx, serror.ErrRoleDenied)
				} else {
					writeErrorResponse(reqCtx, serror.ErrRoleDenied)
				}
				reqCtx.Abort()
				return
			}
		}

		if handler != nil {
			handler(c, reqCtx)
		} else {
			reqCtx.Next(c)
		}
	}
}

// CheckLoginMiddleware creates login check middleware CheckLoginMiddleware 生成登录检查中间件
func CheckLoginMiddleware(
	ctx context.Context,
	handler hertzapp.HandlerFunc,
	failFunc func(c context.Context, reqCtx *hertzapp.RequestContext, err error),
	authType ...string,
) hertzapp.HandlerFunc {
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
	handler hertzapp.HandlerFunc,
	failFunc func(c context.Context, reqCtx *hertzapp.RequestContext, err error),
	authType ...string,
) hertzapp.HandlerFunc {
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
	handler hertzapp.HandlerFunc,
	failFunc func(c context.Context, reqCtx *hertzapp.RequestContext, err error),
	authType ...string,
) hertzapp.HandlerFunc {
	ann := &Annotation{CheckPermission: perms}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// CheckDisableMiddleware creates disable check middleware CheckDisableMiddleware 生成封禁检查中间件
func CheckDisableMiddleware(
	ctx context.Context,
	handler hertzapp.HandlerFunc,
	failFunc func(c context.Context, reqCtx *hertzapp.RequestContext, err error),
	authType ...string,
) hertzapp.HandlerFunc {
	ann := &Annotation{CheckDisable: true}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// IgnoreMiddleware creates ignore auth middleware IgnoreMiddleware 生成忽略认证中间件
func IgnoreMiddleware(
	ctx context.Context,
	handler hertzapp.HandlerFunc,
	failFunc func(c context.Context, reqCtx *hertzapp.RequestContext, err error),
) hertzapp.HandlerFunc {
	return GetHandler(ctx, handler, failFunc, &Annotation{Ignore: true})
}

// CheckLoginAndRoleMiddleware creates login and role middleware CheckLoginAndRoleMiddleware 生成登录与角色检查中间件
func CheckLoginAndRoleMiddleware(
	ctx context.Context,
	roles []string,
	handler hertzapp.HandlerFunc,
	failFunc func(c context.Context, reqCtx *hertzapp.RequestContext, err error),
	authType ...string,
) hertzapp.HandlerFunc {
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
	handler hertzapp.HandlerFunc,
	failFunc func(c context.Context, reqCtx *hertzapp.RequestContext, err error),
	authType ...string,
) hertzapp.HandlerFunc {
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
	handler hertzapp.HandlerFunc,
	failFunc func(c context.Context, reqCtx *hertzapp.RequestContext, err error),
	authType ...string,
) hertzapp.HandlerFunc {
	ann := &Annotation{CheckLogin: true, CheckRole: roles, CheckPermission: perms}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// AuthGroup creates auth route group AuthGroup 创建认证路由组
func AuthGroup(
	ctx context.Context,
	group *route.RouterGroup,
	handler hertzapp.HandlerFunc,
	failFunc func(c context.Context, reqCtx *hertzapp.RequestContext, err error),
	authType ...string,
) *route.RouterGroup {
	group.Use(CheckLoginMiddleware(ctx, handler, failFunc, authType...))
	return group
}

// RoleGroup creates role route group RoleGroup 创建角色路由组
func RoleGroup(
	ctx context.Context,
	group *route.RouterGroup,
	roles []string,
	handler hertzapp.HandlerFunc,
	failFunc func(c context.Context, reqCtx *hertzapp.RequestContext, err error),
	authType ...string,
) *route.RouterGroup {
	group.Use(CheckLoginAndRoleMiddleware(ctx, roles, handler, failFunc, authType...))
	return group
}

// PermissionGroup creates permission route group PermissionGroup 创建权限路由组
func PermissionGroup(
	ctx context.Context,
	group *route.RouterGroup,
	perms []string,
	handler hertzapp.HandlerFunc,
	failFunc func(c context.Context, reqCtx *hertzapp.RequestContext, err error),
	authType ...string,
) *route.RouterGroup {
	group.Use(CheckLoginAndPermissionMiddleware(ctx, perms, handler, failFunc, authType...))
	return group
}

// RoleAndPermissionGroup creates role and permission route group RoleAndPermissionGroup 创建角色与权限路由组
func RoleAndPermissionGroup(
	ctx context.Context,
	group *route.RouterGroup,
	roles []string,
	perms []string,
	handler hertzapp.HandlerFunc,
	failFunc func(c context.Context, reqCtx *hertzapp.RequestContext, err error),
	authType ...string,
) *route.RouterGroup {
	group.Use(CheckAllMiddleware(ctx, roles, perms, handler, failFunc, authType...))
	return group
}
