package echo

import (
	"context"

	"github.com/click33/sa-token-go/core/serror"
	"github.com/click33/sa-token-go/stputil"
	echo4 "github.com/labstack/echo/v4"
)

// Annotation describes declarative auth requirements Annotation 描述声明式认证要求
type Annotation struct {
	AuthType        string
	CheckLogin      bool
	CheckRole       []string
	CheckPermission []string
	CheckDisable    bool
	Ignore          bool
	LogicType       LogicType
}

// GetHandler wraps Echo handler with annotation checks GetHandler 使用注解检查包装 Echo 处理器
func GetHandler(ctx context.Context, handler echo4.HandlerFunc, failFunc func(c echo4.Context, err error) error, annotations ...*Annotation) echo4.HandlerFunc {
	return func(c echo4.Context) error {
		if len(annotations) > 0 && annotations[0].Ignore {
			if handler != nil {
				return handler(c)
			}
			return nil
		}

		ann := &Annotation{}
		if len(annotations) > 0 {
			ann = annotations[0]
		}

		needAuth := ann.CheckLogin || ann.CheckDisable || len(ann.CheckPermission) > 0 || len(ann.CheckRole) > 0
		if !needAuth {
			if handler != nil {
				return handler(c)
			}
			return nil
		}

		mgr, err := stputil.GetManager(ann.AuthType)
		if err != nil {
			if failFunc != nil {
				return failFunc(c, err)
			}
			return writeErrorResponse(c, err)
		}

		// Get SaTokenContext (reuse cached context) 获取 SaTokenContext（复用缓存上下文）
		saCtx := getSaTokenContext(c, mgr)
		token := saCtx.GetTokenValue()
		if !mgr.IsLogin(ctx, token) {
			if failFunc != nil {
				return failFunc(c, serror.ErrNotLogin)
			}
			return writeErrorResponse(c, serror.ErrNotLogin)
		}

		var loginID string
		if ann.CheckDisable || len(ann.CheckPermission) > 0 || len(ann.CheckRole) > 0 {
			loginID, err = mgr.GetLoginID(ctx, token)
			if err != nil {
				if failFunc != nil {
					return failFunc(c, err)
				}
				return writeErrorResponse(c, err)
			}
		}

		if ann.CheckDisable && mgr.IsDisable(ctx, loginID) {
			if failFunc != nil {
				return failFunc(c, serror.ErrAccountDisabled)
			}
			return writeErrorResponse(c, serror.ErrAccountDisabled)
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
					return failFunc(c, serror.ErrPermissionDenied)
				}
				return writeErrorResponse(c, serror.ErrPermissionDenied)
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
					return failFunc(c, serror.ErrRoleDenied)
				}
				return writeErrorResponse(c, serror.ErrRoleDenied)
			}
		}

		if handler != nil {
			return handler(c)
		}
		return nil
	}
}

// CheckLoginMiddleware decorates handler with login checks CheckLoginMiddleware 为处理器增加登录校验
func CheckLoginMiddleware(ctx context.Context, handler echo4.HandlerFunc, failFunc func(c echo4.Context, err error) error, authType ...string) echo4.HandlerFunc {
	ann := &Annotation{CheckLogin: true}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// CheckRoleMiddleware decorates handler with role checks CheckRoleMiddleware 为处理器增加角色校验
func CheckRoleMiddleware(ctx context.Context, roles []string, handler echo4.HandlerFunc, failFunc func(c echo4.Context, err error) error, authType ...string) echo4.HandlerFunc {
	ann := &Annotation{CheckRole: roles}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// CheckPermissionMiddleware decorates handler with permission checks CheckPermissionMiddleware 为处理器增加权限校验
func CheckPermissionMiddleware(ctx context.Context, perms []string, handler echo4.HandlerFunc, failFunc func(c echo4.Context, err error) error, authType ...string) echo4.HandlerFunc {
	ann := &Annotation{CheckPermission: perms}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// CheckDisableMiddleware decorates handler with disable checks CheckDisableMiddleware 为处理器增加封禁校验
func CheckDisableMiddleware(ctx context.Context, handler echo4.HandlerFunc, failFunc func(c echo4.Context, err error) error, authType ...string) echo4.HandlerFunc {
	ann := &Annotation{CheckDisable: true}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// IgnoreMiddleware skips SaToken checks for wrapped handler IgnoreMiddleware 为处理器跳过 SaToken 校验
func IgnoreMiddleware(ctx context.Context, handler echo4.HandlerFunc, failFunc func(c echo4.Context, err error) error) echo4.HandlerFunc {
	return GetHandler(ctx, handler, failFunc, &Annotation{Ignore: true})
}

// CheckLoginAndRoleMiddleware decorates handler with login and role checks CheckLoginAndRoleMiddleware 为处理器增加登录与角色校验
func CheckLoginAndRoleMiddleware(ctx context.Context, roles []string, handler echo4.HandlerFunc, failFunc func(c echo4.Context, err error) error, authType ...string) echo4.HandlerFunc {
	ann := &Annotation{CheckLogin: true, CheckRole: roles}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// CheckLoginAndPermissionMiddleware decorates handler with login and permission checks CheckLoginAndPermissionMiddleware 为处理器增加登录与权限校验
func CheckLoginAndPermissionMiddleware(ctx context.Context, perms []string, handler echo4.HandlerFunc, failFunc func(c echo4.Context, err error) error, authType ...string) echo4.HandlerFunc {
	ann := &Annotation{CheckLogin: true, CheckPermission: perms}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// CheckAllMiddleware decorates handler with login role and permission checks CheckAllMiddleware 为处理器增加登录角色权限校验
func CheckAllMiddleware(ctx context.Context, roles []string, perms []string, handler echo4.HandlerFunc, failFunc func(c echo4.Context, err error) error, authType ...string) echo4.HandlerFunc {
	ann := &Annotation{CheckLogin: true, CheckRole: roles, CheckPermission: perms}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// AuthGroup attaches login middleware to Echo group AuthGroup 为 Echo 路由组挂载登录校验
func AuthGroup(ctx context.Context, group *echo4.Group, failFunc func(c echo4.Context, err error) error, authType ...string) *echo4.Group {
	options := []AuthOption{WithFailFunc(failFunc)}
	if len(authType) > 0 {
		options = append(options, WithAuthType(authType[0]))
	}
	group.Use(AuthMiddleware(ctx, options...))
	return group
}

// RoleGroup attaches login and role middleware to Echo group RoleGroup 为 Echo 路由组挂载登录与角色校验
func RoleGroup(ctx context.Context, group *echo4.Group, roles []string, failFunc func(c echo4.Context, err error) error, authType ...string) *echo4.Group {
	options := []AuthOption{WithFailFunc(failFunc)}
	if len(authType) > 0 {
		options = append(options, WithAuthType(authType[0]))
	}
	group.Use(AuthMiddleware(ctx, options...))
	group.Use(RoleMiddleware(ctx, roles, options...))
	return group
}

// PermissionGroup attaches login and permission middleware to Echo group PermissionGroup 为 Echo 路由组挂载登录与权限校验
func PermissionGroup(ctx context.Context, group *echo4.Group, perms []string, failFunc func(c echo4.Context, err error) error, authType ...string) *echo4.Group {
	options := []AuthOption{WithFailFunc(failFunc)}
	if len(authType) > 0 {
		options = append(options, WithAuthType(authType[0]))
	}
	group.Use(AuthMiddleware(ctx, options...))
	group.Use(PermissionMiddleware(ctx, perms, options...))
	return group
}

// RoleAndPermissionGroup attaches login role and permission middleware to Echo group RoleAndPermissionGroup 为 Echo 路由组挂载登录角色权限校验
func RoleAndPermissionGroup(ctx context.Context, group *echo4.Group, roles []string, perms []string, failFunc func(c echo4.Context, err error) error, authType ...string) *echo4.Group {
	options := []AuthOption{WithFailFunc(failFunc)}
	if len(authType) > 0 {
		options = append(options, WithAuthType(authType[0]))
	}
	group.Use(AuthMiddleware(ctx, options...))
	group.Use(RoleMiddleware(ctx, roles, options...))
	group.Use(PermissionMiddleware(ctx, perms, options...))
	return group
}
