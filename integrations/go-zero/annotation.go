package gozero

import (
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest"
)

// Annotation tag constants (aligned with Gin).
// 注解标签常量（与 Gin 集成一致）。
const (
	TagSaCheckLogin      = "sa_check_login"
	TagSaCheckRole       = "sa_check_role"
	TagSaCheckPermission = "sa_check_permission"
	TagSaCheckDisable    = "sa_check_disable"
	TagSaIgnore          = "sa_ignore"
)

// Annotation describes auth requirements for a route.
// Annotation 描述路由的鉴权要求。
type Annotation struct {
	CheckLogin      bool     `json:"checkLogin"`
	CheckRole       []string `json:"checkRole"`
	CheckPermission []string `json:"checkPermission"`
	CheckDisable    bool     `json:"checkDisable"`
	Ignore          bool     `json:"ignore"`
}

// ParseTag parses comma-separated sa-token annotation tag.
// ParseTag 解析逗号分隔的 sa-token 注解标签。
func ParseTag(tag string) *Annotation {
	ann := &Annotation{}
	if tag == "" {
		return ann
	}
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		switch {
		case part == TagSaCheckLogin || part == "login":
			ann.CheckLogin = true
		case strings.HasPrefix(part, TagSaCheckRole+"=") || strings.HasPrefix(part, "role="):
			roles := strings.TrimPrefix(part, TagSaCheckRole+"=")
			roles = strings.TrimPrefix(roles, "role=")
			if roles != "" {
				ann.CheckRole = strings.Split(roles, "|")
			}
		case strings.HasPrefix(part, TagSaCheckPermission+"=") || strings.HasPrefix(part, "permission="):
			perms := strings.TrimPrefix(part, TagSaCheckPermission+"=")
			perms = strings.TrimPrefix(perms, "permission=")
			if perms != "" {
				ann.CheckPermission = strings.Split(perms, "|")
			}
		case part == TagSaCheckDisable || part == "disable":
			ann.CheckDisable = true
		case part == TagSaIgnore || part == "ignore":
			ann.Ignore = true
		}
	}
	return ann
}

// Validate returns true when at most one check type is set (or Ignore).
// Validate 在至多一种检查类型（或 Ignore）时返回 true。
func (a *Annotation) Validate() bool {
	if a == nil || a.Ignore {
		return true
	}
	count := 0
	if a.CheckLogin {
		count++
	}
	if len(a.CheckRole) > 0 {
		count++
	}
	if len(a.CheckPermission) > 0 {
		count++
	}
	if a.CheckDisable {
		count++
	}
	return count <= 1
}

// GetHandler wraps http.HandlerFunc with annotation auth.
// GetHandler 用注解鉴权包装 http.HandlerFunc。
func GetHandler(handler http.HandlerFunc, annotations ...*Annotation) http.HandlerFunc {
	var ann *Annotation
	if len(annotations) > 0 {
		ann = annotations[0]
	}
	return func(w http.ResponseWriter, r *http.Request) {
		r2, ok := runAnnotationAuth(w, r, ann)
		if !ok {
			return
		}
		if handler != nil {
			handler(w, r2)
		}
	}
}

// Middleware builds rest.Middleware from annotations.
// Middleware 根据注解构建 rest.Middleware。
func Middleware(annotations ...*Annotation) rest.Middleware {
	var ann *Annotation
	if len(annotations) > 0 {
		ann = annotations[0]
	}
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			r2, ok := runAnnotationAuth(w, r, ann)
			if !ok {
				return
			}
			next(w, r2)
		}
	}
}

// WithAnnotation builds middleware from a custom annotation (panics if invalid).
// WithAnnotation 用自定义注解构建中间件（非法配置时 panic）。
func WithAnnotation(ann *Annotation) rest.Middleware {
	if ann != nil && !ann.Validate() {
		panic("gozero: invalid annotation configuration")
	}
	return Middleware(ann)
}

// HandlerWithAnnotations bundles handler and annotations.
// HandlerWithAnnotations 组合 Handler 与注解。
type HandlerWithAnnotations struct {
	Handler     http.HandlerFunc
	Annotations []*Annotation
}

// ToHandler converts to http.HandlerFunc.
// ToHandler 转为 http.HandlerFunc。
func (h *HandlerWithAnnotations) ToHandler() http.HandlerFunc {
	return GetHandler(h.Handler, h.Annotations...)
}

// ToMiddleware converts to rest.Middleware.
// ToMiddleware 转为 rest.Middleware。
func (h *HandlerWithAnnotations) ToMiddleware() rest.Middleware {
	return Middleware(h.Annotations...)
}

// CheckRole returns middleware that requires any of the roles.
// CheckRole 返回要求任一角色的中间件。
func CheckRole(roles ...string) rest.Middleware {
	return Middleware(&Annotation{CheckRole: roles})
}

// CheckPermission returns middleware that requires any of the permissions.
// CheckPermission 返回要求任一权限的中间件。
func CheckPermission(perms ...string) rest.Middleware {
	return Middleware(&Annotation{CheckPermission: perms})
}

// CheckDisable returns middleware that rejects disabled accounts.
// CheckDisable 返回拒绝封禁账号的中间件。
func CheckDisable() rest.Middleware {
	return Middleware(&Annotation{CheckDisable: true})
}

// Ignore returns middleware that skips auth checks.
// Ignore 返回跳过鉴权的中间件。
func Ignore() rest.Middleware {
	return Middleware(&Annotation{Ignore: true})
}

// CheckLoginMiddleware returns middleware that requires login.
// CheckLoginMiddleware 返回要求登录的中间件（包内不与 export.CheckLogin 重名）。
func CheckLoginMiddleware() rest.Middleware {
	return Middleware(&Annotation{CheckLogin: true})
}

// CheckRoleMiddleware is an alias of CheckRole.
// CheckRoleMiddleware 为 CheckRole 的别名。
func CheckRoleMiddleware(roles ...string) rest.Middleware { return CheckRole(roles...) }

// CheckPermissionMiddleware is an alias of CheckPermission.
// CheckPermissionMiddleware 为 CheckPermission 的别名。
func CheckPermissionMiddleware(perms ...string) rest.Middleware { return CheckPermission(perms...) }

// CheckDisableMiddleware is an alias of CheckDisable.
// CheckDisableMiddleware 为 CheckDisable 的别名。
func CheckDisableMiddleware() rest.Middleware { return CheckDisable() }

// IgnoreMiddleware is an alias of Ignore.
// IgnoreMiddleware 为 Ignore 的别名。
func IgnoreMiddleware() rest.Middleware { return Ignore() }

// callHandler invokes supported handler types.
// callHandler 调用支持的 handler 类型。
func callHandler(handler interface{}, w http.ResponseWriter, r *http.Request) bool {
	if handler == nil {
		return false
	}
	switch h := handler.(type) {
	case http.HandlerFunc:
		if h != nil {
			h(w, r)
			return true
		}
	case func(http.ResponseWriter, *http.Request):
		h(w, r)
		return true
	}
	return false
}

// ProcessStructAnnotations is a placeholder aligned with Gin (returns empty annotation).
// ProcessStructAnnotations 与 Gin 一致的占位实现（返回空注解）。
func ProcessStructAnnotations(handler interface{}) http.HandlerFunc {
	_ = handler
	return GetHandler(nil, &Annotation{})
}
