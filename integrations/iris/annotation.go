package iris

import (
	"reflect"
	"strings"

	"github.com/sa-tokens/sa-token-go/core"
	"github.com/sa-tokens/sa-token-go/stputil"
	irisfw "github.com/kataras/iris/v12"
)

// Annotation tag constants, kept in sync with gin/hertz for cross-framework migration.
// 注解标签常量；与 gin/hertz 保持一致，便于跨框架迁移。
const (
	TagSaCheckLogin      = "sa_check_login"
	TagSaCheckRole       = "sa_check_role"
	TagSaCheckPermission = "sa_check_permission"
	TagSaCheckDisable    = "sa_check_disable"
	TagSaIgnore          = "sa_ignore"
)

// Annotation describes the authentication requirement of a single handler.
// A single instance must carry at most one check kind (enforced by Validate()).
// Ignore has highest priority: when true, all other fields are ignored.
// Annotation 描述一个 handler 的鉴权诉求。
// 同一实例最多承载一种检查类型，由 Validate() 强制；Ignore 优先级最高，置位时其余字段被忽略。
type Annotation struct {
	CheckLogin      bool     `json:"checkLogin"`
	CheckRole       []string `json:"checkRole"`
	CheckPermission []string `json:"checkPermission"`
	CheckDisable    bool     `json:"checkDisable"`
	Ignore          bool     `json:"ignore"`
}

// ParseTag parses a comma-separated annotation tag string, tolerating whitespace
// and accepting both canonical names (e.g. sa_check_role) and short aliases (e.g. role).
// ParseTag 解析逗号分隔的注解标签字符串，宽容空白与多种别名。
func ParseTag(tag string) *Annotation {
	ann := &Annotation{}
	if tag == "" {
		return ann
	}
	for _, part := range strings.Split(tag, ",") {
		part = strings.TrimSpace(part)
		switch {
		case part == TagSaCheckLogin || part == "login":
			ann.CheckLogin = true
		case strings.HasPrefix(part, TagSaCheckRole+"=") || strings.HasPrefix(part, "role="):
			v := strings.TrimPrefix(strings.TrimPrefix(part, TagSaCheckRole+"="), "role=")
			if v != "" {
				ann.CheckRole = strings.Split(v, "|")
			}
		case strings.HasPrefix(part, TagSaCheckPermission+"=") || strings.HasPrefix(part, "permission="):
			v := strings.TrimPrefix(strings.TrimPrefix(part, TagSaCheckPermission+"="), "permission=")
			if v != "" {
				ann.CheckPermission = strings.Split(v, "|")
			}
		case part == TagSaCheckDisable || part == "disable":
			ann.CheckDisable = true
		case part == TagSaIgnore || part == "ignore":
			ann.Ignore = true
		}
	}
	return ann
}

// Validate reports whether the annotation combination is legal: Ignore is always legal;
// otherwise at most one of the other check kinds may be set.
// Validate 校验注解组合是否合法：Ignore 总是合法；其余字段同时最多置位一种。
func (a *Annotation) Validate() bool {
	if a.Ignore {
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

// runAnnotationChecks runs the annotation pipeline.
// Returns true when the request should proceed (caller should run handler or Next);
// returns false when an error response has been written and StopExecution has been called,
// in which case the caller MUST return immediately.
//
// Rationale: extract the duplicated auth logic from GetHandler and Middleware so that
// bugfixes and extensions only need to land in one place.
//
// runAnnotationChecks 执行注解链路检查。
// 返回 true 表示请求应当被放行（调用方应继续 Next 或执行 handler）；
// 返回 false 表示已写入错误响应并 StopExecution，调用方必须立即返回。
// 设计动机：抽取 GetHandler 与 Middleware 中的重复鉴权逻辑，集中修复 bug 与扩展能力。
func runAnnotationChecks(c irisfw.Context, ann *Annotation) bool {
	if ann != nil && ann.Ignore {
		return true
	}

	saCtx := core.NewContext(NewIrisContext(c), stputil.GetManager())
	token := saCtx.GetTokenValue()
	if token == "" || !stputil.IsLogin(token) {
		writeErrorResponse(c, core.NewNotLoginError())
		c.StopExecution()
		return false
	}

	loginID, err := stputil.GetLoginID(token)
	if err != nil {
		writeErrorResponse(c, err)
		c.StopExecution()
		return false
	}

	if ann == nil {
		return true
	}

	// CheckDisable: reject if the account has been disabled.
	// CheckDisable：账号封禁拦截。
	if ann.CheckDisable && stputil.IsDisable(loginID) {
		writeErrorResponse(c, core.NewAccountDisabledError(loginID))
		c.StopExecution()
		return false
	}

	// CheckPermission: OR semantics — pass if ANY of the listed permissions matches.
	// CheckPermission：OR 语义，命中任一即通过。
	if len(ann.CheckPermission) > 0 {
		hit := false
		for _, perm := range ann.CheckPermission {
			if stputil.HasPermission(loginID, strings.TrimSpace(perm)) {
				hit = true
				break
			}
		}
		if !hit {
			writeErrorResponse(c, core.NewPermissionDeniedError(strings.Join(ann.CheckPermission, ",")))
			c.StopExecution()
			return false
		}
	}

	// CheckRole: OR semantics. | CheckRole：OR 语义。
	if len(ann.CheckRole) > 0 {
		hit := false
		for _, role := range ann.CheckRole {
			if stputil.HasRole(loginID, strings.TrimSpace(role)) {
				hit = true
				break
			}
		}
		if !hit {
			writeErrorResponse(c, core.NewRoleDeniedError(strings.Join(ann.CheckRole, ",")))
			c.StopExecution()
			return false
		}
	}

	// CheckLogin: explicit branch. IsLogin has already been verified above; this is a
	// semantic placeholder reserved for future extensions (e.g. session renewal, risk checks).
	// CheckLogin 单独分支：上方已通过 IsLogin 校验，这里无需额外动作，
	// 显式 case 留作语义占位，便于未来扩展（如登录态续期、风险检测等）。
	if ann.CheckLogin {
		_ = loginID
	}

	return true
}

// GetHandler wraps a raw handler into an annotation-checked Iris Handler.
// handler may be either func(irisfw.Context) or any function reflect-callable with
// a single argument assignable to irisfw.Context.
// GetHandler 将原始 handler 包裹为带注解鉴权的 Iris Handler。
// handler 可为 func(irisfw.Context) 或通过反射可调用、入参兼容 irisfw.Context 的函数。
func GetHandler(handler interface{}, annotations ...*Annotation) irisfw.Handler {
	var ann *Annotation
	if len(annotations) > 0 {
		ann = annotations[0]
	}
	// Validate combinations at registration time so misuse fails fast instead of in production.
	// 注册期校验非法组合，把错误前置；线上路由不应混用多种检查。
	if ann != nil && !ann.Validate() {
		panic("sa-token iris: invalid annotation combination, see Annotation.Validate()")
	}

	return func(c irisfw.Context) {
		if ann != nil && ann.Ignore {
			if callHandler(handler, c) {
				return
			}
			c.Next()
			return
		}
		if !runAnnotationChecks(c, ann) {
			return
		}
		if callHandler(handler, c) {
			return
		}
		c.Next()
	}
}

// callHandler supports both a direct function value and a reflect-callable function whose
// single argument is assignable to irisfw.Context.
// Returns true when executed; false when handler is not callable, in which case callers
// should fall back to c.Next().
// callHandler 兼容直接函数与反射可调用入参的两种形态。
// 返回 true 表示已执行；false 表示 handler 不可调用，调用方应走 c.Next()。
func callHandler(handler interface{}, c irisfw.Context) bool {
	if handler == nil {
		return false
	}
	if h, ok := handler.(func(irisfw.Context)); ok {
		if h == nil {
			return false
		}
		h(c)
		return true
	}

	hv := reflect.ValueOf(handler)
	if hv.Kind() != reflect.Func || hv.IsNil() || hv.Type().NumIn() != 1 {
		return false
	}
	argType := hv.Type().In(0)
	if !argType.AssignableTo(reflect.TypeOf((*irisfw.Context)(nil)).Elem()) {
		return false
	}
	hv.Call([]reflect.Value{reflect.ValueOf(c)})
	return true
}

// ===== Decorators: single-responsibility shortcuts | 装饰器：单一职责的快捷工厂 =====

// CheckLogin only checks login state. | CheckLogin 仅校验登录态。
func CheckLogin() irisfw.Handler { return GetHandler(nil, &Annotation{CheckLogin: true}) }

// CheckRole checks roles with OR semantics. | CheckRole OR 语义校验角色。
func CheckRole(roles ...string) irisfw.Handler {
	return GetHandler(nil, &Annotation{CheckRole: roles})
}

// CheckPermission checks permissions with OR semantics. | CheckPermission OR 语义校验权限。
func CheckPermission(perms ...string) irisfw.Handler {
	return GetHandler(nil, &Annotation{CheckPermission: perms})
}

// CheckDisable checks whether the account is disabled. | CheckDisable 校验账号是否被封禁。
func CheckDisable() irisfw.Handler { return GetHandler(nil, &Annotation{CheckDisable: true}) }

// Ignore explicitly skips authentication, typically used for whitelisted routes.
// Ignore 显式跳过鉴权，常用于白名单路由。
func Ignore() irisfw.Handler { return GetHandler(nil, &Annotation{Ignore: true}) }

// WithAnnotation uses a custom annotation. | WithAnnotation 使用自定义注解。
func WithAnnotation(ann *Annotation) irisfw.Handler { return GetHandler(nil, ann) }

// Middleware is a check-only middleware that does not invoke a handler.
// Suitable for Party-level mounting.
// Middleware 仅做鉴权、不执行 handler 的中间件形态，适合 Party 级别挂载。
func Middleware(annotations ...*Annotation) irisfw.Handler {
	var ann *Annotation
	if len(annotations) > 0 {
		ann = annotations[0]
	}
	if ann != nil && !ann.Validate() {
		panic("sa-token iris: invalid annotation combination, see Annotation.Validate()")
	}
	return func(c irisfw.Context) {
		if ann != nil && ann.Ignore {
			c.Next()
			return
		}
		if !runAnnotationChecks(c, ann) {
			return
		}
		c.Next()
	}
}

// HandlerWithAnnotations bundles an annotation set with a handler so that route tables
// can construct them uniformly.
// HandlerWithAnnotations 注解 + handler 的封装，便于路由表统一构造。
type HandlerWithAnnotations struct {
	Handler     interface{}
	Annotations []*Annotation
}

// NewHandlerWithAnnotations constructs a HandlerWithAnnotations.
// NewHandlerWithAnnotations 构造 HandlerWithAnnotations。
func NewHandlerWithAnnotations(handler interface{}, annotations ...*Annotation) *HandlerWithAnnotations {
	return &HandlerWithAnnotations{Handler: handler, Annotations: annotations}
}

// ToIrisHandler converts the bundle into an Iris Handler.
// ToIrisHandler 转换为 Iris Handler。
func (h *HandlerWithAnnotations) ToIrisHandler() irisfw.Handler {
	return GetHandler(h.Handler, h.Annotations...)
}

// Deprecated: ProcessStructAnnotations is currently a stub that does NOT actually
// parse method-level annotation tags. Its behavior matches the same-named API in
// gin/hertz; kept only for compile-time compatibility. Production code should use
// GetHandler / WithAnnotation directly.
//
// Deprecated: ProcessStructAnnotations 当前为占位实现，未真正解析结构体方法上的注解。
// 与 gin/hertz 同名 API 行为一致，保留仅为编译期兼容；
// 生产代码请直接使用 GetHandler / WithAnnotation。
func ProcessStructAnnotations(handler interface{}) irisfw.Handler {
	handlerValue := reflect.ValueOf(handler)
	handlerType := reflect.TypeOf(handler)
	if handlerType != nil && handlerType.Kind() == reflect.Ptr {
		handlerType = handlerType.Elem()
	}
	ann := parseMethodAnnotation(handlerType)
	return GetHandler(func(c irisfw.Context) {
		handlerValue.MethodByName("ServeHTTP").Call([]reflect.Value{reflect.ValueOf(c)})
	}, ann)
}

// Deprecated: parseMethodAnnotation is a stub that always returns an empty Annotation.
// Deprecated: parseMethodAnnotation 是占位实现，始终返回空 Annotation。
func parseMethodAnnotation(_ reflect.Type) *Annotation { return &Annotation{} }
