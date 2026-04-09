package fiber

import (
	"context"

	"github.com/click33/sa-token-go/core/serror"
	"github.com/click33/sa-token-go/stputil"
	gofiber "github.com/gofiber/fiber/v2"
)

// Annotation describes declarative auth requirements for a handler Annotation 描述处理器的声明式认证要求。
type Annotation struct {
	AuthType        string
	CheckLogin      bool
	CheckRole       []string
	CheckPermission []string
	CheckDisable    bool
	Ignore          bool
	LogicType       LogicType
}

// GetHandler wraps Fiber handler with annotation-based auth checks GetHandler 为 Fiber 处理器包裹基于注解的认证校验。
func GetHandler(ctx context.Context, handler gofiber.Handler, failFunc func(c *gofiber.Ctx, err error), annotations ...*Annotation) gofiber.Handler {
	return func(c *gofiber.Ctx) error {
		if len(annotations) > 0 && annotations[0].Ignore {
			if handler != nil {
				return handler(c)
			}
			return c.Next()
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
			return c.Next()
		}

		mgr, err := stputil.GetManager(ann.AuthType)
		if err != nil {
			if failFunc != nil {
				failFunc(c, err)
				return nil
			}
			return writeErrorResponse(c, err)
		}

		// Get SaTokenContext (reuse cached context) 获取 SaTokenContext（复用缓存上下文）
		saCtx := getSaTokenContext(c, mgr)
		token := saCtx.GetTokenValue()
		if !mgr.IsLogin(ctx, token) {
			if failFunc != nil {
				failFunc(c, serror.ErrNotLogin)
				return nil
			}
			return writeErrorResponse(c, serror.ErrNotLogin)
		}

		var loginID string
		if ann.CheckDisable || len(ann.CheckPermission) > 0 || len(ann.CheckRole) > 0 {
			loginID, err = mgr.GetLoginID(ctx, token)
			if err != nil {
				if failFunc != nil {
					failFunc(c, err)
					return nil
				}
				return writeErrorResponse(c, err)
			}
		}

		if ann.CheckDisable && mgr.IsDisable(ctx, loginID) {
			if failFunc != nil {
				failFunc(c, serror.ErrAccountDisabled)
				return nil
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
					failFunc(c, serror.ErrPermissionDenied)
					return nil
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
					failFunc(c, serror.ErrRoleDenied)
					return nil
				}
				return writeErrorResponse(c, serror.ErrRoleDenied)
			}
		}

		if handler != nil {
			return handler(c)
		}
		return c.Next()
	}
}

// CheckLoginMiddleware decorates handler with login checks CheckLoginMiddleware 为处理器增加登录校验。
func CheckLoginMiddleware(
	ctx context.Context,
	handler gofiber.Handler,
	failFunc func(c *gofiber.Ctx, err error),
	authType ...string,
) gofiber.Handler {
	ann := &Annotation{CheckLogin: true}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// CheckRoleMiddleware decorates handler with role checks CheckRoleMiddleware 为处理器增加角色校验。
func CheckRoleMiddleware(
	ctx context.Context,
	roles []string,
	handler gofiber.Handler,
	failFunc func(c *gofiber.Ctx, err error),
	authType ...string,
) gofiber.Handler {
	ann := &Annotation{CheckRole: roles}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// CheckPermissionMiddleware decorates handler with permission checks CheckPermissionMiddleware 为处理器增加权限校验。
func CheckPermissionMiddleware(
	ctx context.Context,
	perms []string,
	handler gofiber.Handler,
	failFunc func(c *gofiber.Ctx, err error),
	authType ...string,
) gofiber.Handler {
	ann := &Annotation{CheckPermission: perms}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// CheckDisableMiddleware decorates handler with account-disable checks CheckDisableMiddleware 为处理器增加封禁状态校验。
func CheckDisableMiddleware(
	ctx context.Context,
	handler gofiber.Handler,
	failFunc func(c *gofiber.Ctx, err error),
	authType ...string,
) gofiber.Handler {
	ann := &Annotation{CheckDisable: true}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// IgnoreMiddleware skips SaToken checks for wrapped handler IgnoreMiddleware 为处理器跳过 SaToken 校验。
func IgnoreMiddleware(
	ctx context.Context,
	handler gofiber.Handler,
	failFunc func(c *gofiber.Ctx, err error),
) gofiber.Handler {
	return GetHandler(ctx, handler, failFunc, &Annotation{Ignore: true})
}

// CheckLoginAndRoleMiddleware decorates handler with login and role checks CheckLoginAndRoleMiddleware 为处理器增加登录与角色校验。
func CheckLoginAndRoleMiddleware(
	ctx context.Context,
	roles []string,
	handler gofiber.Handler,
	failFunc func(c *gofiber.Ctx, err error),
	authType ...string,
) gofiber.Handler {
	ann := &Annotation{CheckLogin: true, CheckRole: roles}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// CheckLoginAndPermissionMiddleware decorates handler with login and permission checks CheckLoginAndPermissionMiddleware 为处理器增加登录与权限校验。
func CheckLoginAndPermissionMiddleware(
	ctx context.Context,
	perms []string,
	handler gofiber.Handler,
	failFunc func(c *gofiber.Ctx, err error),
	authType ...string,
) gofiber.Handler {
	ann := &Annotation{CheckLogin: true, CheckPermission: perms}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// CheckAllMiddleware decorates handler with login, role and permission checks CheckAllMiddleware 为处理器增加登录、角色和权限校验。
func CheckAllMiddleware(
	ctx context.Context,
	roles []string,
	perms []string,
	handler gofiber.Handler,
	failFunc func(c *gofiber.Ctx, err error),
	authType ...string,
) gofiber.Handler {
	ann := &Annotation{CheckLogin: true, CheckRole: roles, CheckPermission: perms}
	if len(authType) > 0 {
		ann.AuthType = authType[0]
	}
	return GetHandler(ctx, handler, failFunc, ann)
}

// AuthGroup attaches login checks to a Fiber router group AuthGroup 为 Fiber 路由组挂载登录校验。
func AuthGroup(
	ctx context.Context,
	group gofiber.Router,
	handler gofiber.Handler,
	failFunc func(c *gofiber.Ctx, err error),
	authType ...string,
) gofiber.Router {
	group.Use(CheckLoginMiddleware(ctx, handler, failFunc, authType...))
	return group
}

// RoleGroup attaches role checks to a Fiber router group RoleGroup 为 Fiber 路由组挂载角色校验。
func RoleGroup(
	ctx context.Context,
	group gofiber.Router,
	roles []string,
	handler gofiber.Handler,
	failFunc func(c *gofiber.Ctx, err error),
	authType ...string,
) gofiber.Router {
	group.Use(CheckLoginAndRoleMiddleware(ctx, roles, handler, failFunc, authType...))
	return group
}

// PermissionGroup attaches permission checks to a Fiber router group PermissionGroup 为 Fiber 路由组挂载权限校验。
func PermissionGroup(
	ctx context.Context,
	group gofiber.Router,
	perms []string,
	handler gofiber.Handler,
	failFunc func(c *gofiber.Ctx, err error),
	authType ...string,
) gofiber.Router {
	group.Use(CheckLoginAndPermissionMiddleware(ctx, perms, handler, failFunc, authType...))
	return group
}

// RoleAndPermissionGroup attaches role and permission checks to a Fiber router group RoleAndPermissionGroup 为 Fiber 路由组挂载角色与权限校验。
func RoleAndPermissionGroup(
	ctx context.Context,
	group gofiber.Router,
	roles []string,
	perms []string,
	handler gofiber.Handler,
	failFunc func(c *gofiber.Ctx, err error),
	authType ...string,
) gofiber.Router {
	group.Use(CheckAllMiddleware(ctx, roles, perms, handler, failFunc, authType...))
	return group
}
