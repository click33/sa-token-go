package gin

import (
	"reflect"
	"strings"

	ginfw "github.com/gin-gonic/gin"
	"github.com/sa-tokens/sa-token-go/core"
	"github.com/sa-tokens/sa-token-go/stputil"
)

// Annotation tag / field name constants | 注解常量
const (
	TagSaCheckLogin      = "sa_check_login"
	TagSaCheckRole       = "sa_check_role"
	TagSaCheckPermission = "sa_check_permission"
	TagSaCheckDisable    = "sa_check_disable"
	TagSaIgnore          = "sa_ignore"
	// TagSaMode is the struct-tag key for set combine: mode=AND|OR (alias combine=)
	// TagSaMode 权限组与角色组组合运算标签键：mode=AND|OR（亦支持 combine=）
	TagSaMode = "mode"
)

// Annotation describes decorator checks for a Gin handler.
// CheckPermission / CheckRole slices use OR within the list;
// PermissionAndRoleOperation combines the two set-results (empty = independent mode).
// Annotation 注解结构体：组内为 OR；PermissionAndRoleOperation 描述两组结果之间的关系（空串=独立模式）。
type Annotation struct {
	CheckLogin                 bool                            `json:"checkLogin"`
	CheckRole                  []string                        `json:"checkRole"`
	CheckPermission            []string                        `json:"checkPermission"`
	CheckDisable               bool                            `json:"checkDisable"`
	Ignore                     bool                            `json:"ignore"`
	PermissionAndRoleOperation core.PermissionAndRoleOperation `json:"permissionAndRoleOperation"`
}

// ParseTag parses a comma-separated struct tag; supports mode=AND|OR and combine=AND|OR.
// ParseTag 解析结构体标签；支持 mode=AND|OR 与 combine=AND|OR。
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
		case strings.HasPrefix(part, TagSaMode+"=") || strings.HasPrefix(part, "combine="):
			// Write combine op; invalid values are rejected by Validate | 组间运算写入 Operation；非法值留给 Validate
			mode := strings.TrimPrefix(part, TagSaMode+"=")
			mode = strings.TrimPrefix(mode, "combine=")
			ann.PermissionAndRoleOperation = core.PermissionAndRoleOperation(strings.ToUpper(strings.TrimSpace(mode)))
		}
	}
	return ann
}

// Validate checks whether the annotation config is self-consistent.
// Validate 校验注解配置是否自洽。
// Rules | 规则:
//   - Ignore alone => valid;
//   - AND/OR: both permission and role lists required; must not mix with Disable;
//   - Independent (empty op): Permission+Role together is allowed (both must pass at runtime);
//     Login / Disable / auth-set count as categories; at most one category.
func (a *Annotation) Validate() bool {
	if a == nil {
		return false
	}
	if a.Ignore {
		return true
	}
	op := a.PermissionAndRoleOperation
	if op != "" {
		if !op.IsValid() {
			return false
		}
		if len(trimNonEmpty(a.CheckPermission)) == 0 || len(trimNonEmpty(a.CheckRole)) == 0 {
			return false
		}
		if a.CheckDisable {
			return false
		}
		return true
	}
	// Independent mode: auth-set counts as 1; Login / Disable each count as 1 | 独立模式计数
	count := 0
	if a.CheckLogin {
		count++
	}
	if len(trimNonEmpty(a.CheckPermission)) > 0 || len(trimNonEmpty(a.CheckRole)) > 0 {
		count++
	}
	if a.CheckDisable {
		count++
	}
	return count <= 1
}

// firstAnnotation returns the first annotation safely | 安全取第一个注解
func firstAnnotation(annotations []*Annotation) *Annotation {
	if len(annotations) == 0 {
		return nil
	}
	return annotations[0]
}

// trimNonEmpty drops blank entries so "|" splits do not create false permissions.
// trimNonEmpty 去掉空白与空串，避免 "|" 切出空权限误判。
func trimNonEmpty(items []string) []string {
	out := make([]string, 0, len(items))
	for _, s := range items {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// ensureLoggedIn resolves token and confirms login; on failure writes response and Abort.
// ensureLoggedIn 解析 token 并确认登录；失败时已写响应并 Abort。
func ensureLoggedIn(c *ginfw.Context) (loginID string, ok bool) {
	ctx := NewGinContext(c)
	saCtx := core.NewContext(ctx, stputil.GetManager())
	token := saCtx.GetTokenValue()
	if token == "" || !stputil.IsLogin(token) {
		writeErrorResponse(c, core.NewNotLoginError())
		c.Abort()
		return "", false
	}
	id, err := stputil.GetLoginID(token)
	if err != nil {
		writeErrorResponse(c, err)
		c.Abort()
		return "", false
	}
	return id, true
}

// runAnnotationChecks runs disable / permission / role / combine checks.
// Returns false if an error was already written and the request aborted.
// runAnnotationChecks 执行封禁 / 权限 / 角色 / 组间组合校验；返回 false 表示已写错误并 Abort。
//
// Flow | 流程: Disable? → permOK/roleOK (OR within list) → Operation branch → pass
func runAnnotationChecks(c *ginfw.Context, loginID string, ann *Annotation) bool {
	if ann == nil {
		return true
	}
	if ann.CheckDisable && stputil.IsDisable(loginID) {
		writeErrorResponse(c, core.NewAccountDisabledError(loginID))
		c.Abort()
		return false
	}

	perms := trimNonEmpty(ann.CheckPermission)
	roles := trimNonEmpty(ann.CheckRole)
	permOK := true
	roleOK := true
	if len(perms) > 0 {
		permOK = stputil.HasPermissionsOr(loginID, perms)
	}
	if len(roles) > 0 {
		roleOK = stputil.HasRolesOr(loginID, roles)
	}

	op := ann.PermissionAndRoleOperation
	switch {
	case op == core.PermissionAndRoleOperationAND:
		// Both lists required; both set-results must be true | 双边非空且双真
		if len(perms) == 0 || len(roles) == 0 {
			writeErrorResponse(c, core.NewInvalidAnnotationError("AND requires both permissions and roles"))
			c.Abort()
			return false
		}
		if !(permOK && roleOK) {
			writeErrorResponse(c, core.NewPermissionRoleCombineError("AND", perms, roles))
			c.Abort()
			return false
		}
	case op == core.PermissionAndRoleOperationOR:
		// Both lists required; either set-result may be true | 双边非空且一真
		if len(perms) == 0 || len(roles) == 0 {
			writeErrorResponse(c, core.NewInvalidAnnotationError("OR requires both permissions and roles"))
			c.Abort()
			return false
		}
		if !(permOK || roleOK) {
			writeErrorResponse(c, core.NewPermissionRoleCombineError("OR", perms, roles))
			c.Abort()
			return false
		}
	default:
		// Independent mode (invalid op falls here): each configured list must pass
		// 独立模式（含非法 op 兜底）：有列表则必须通过
		if len(perms) > 0 && !permOK {
			writeErrorResponse(c, core.NewPermissionDeniedError(strings.Join(perms, ",")))
			c.Abort()
			return false
		}
		if len(roles) > 0 && !roleOK {
			writeErrorResponse(c, core.NewRoleDeniedError(strings.Join(roles, ",")))
			c.Abort()
			return false
		}
	}
	return true
}

// GetHandler wraps a handler with annotation checks (Ignore / login / auth).
// GetHandler 获取带注解的处理器。
func GetHandler(handler interface{}, annotations ...*Annotation) ginfw.HandlerFunc {
	return func(c *ginfw.Context) {
		ann := firstAnnotation(annotations)
		if ann != nil && ann.Ignore {
			if callHandler(handler, c) {
				return
			}
			c.Next()
			return
		}
		loginID, ok := ensureLoggedIn(c)
		if !ok {
			return
		}
		if !runAnnotationChecks(c, loginID, ann) {
			return
		}
		if callHandler(handler, c) {
			return
		}
		c.Next()
	}
}

// callHandler invokes handler if it is a Gin handler func; returns whether it ran.
// callHandler 若 handler 可调用则执行并返回 true。
func callHandler(handler interface{}, c *ginfw.Context) bool {
	if handler == nil {
		return false
	}

	switch h := handler.(type) {
	case func(*ginfw.Context):
		if h == nil {
			return false
		}
		h(c)
		return true
	case ginfw.HandlerFunc:
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
	if !argType.AssignableTo(reflect.TypeOf(c)) {
		return false
	}

	hv.Call([]reflect.Value{reflect.ValueOf(c)})
	return true
}

// Decorator functions | 装饰器函数

// CheckLogin requires login (non-Ignore paths already enforce login).
// CheckLogin 检查登录装饰器（非 Ignore 路径本就会强制登录）。
func CheckLogin() ginfw.HandlerFunc {
	return GetHandler(nil, &Annotation{CheckLogin: true})
}

// CheckRole requires any of the given roles (OR within list).
// CheckRole 检查角色（变参 OR）。
func CheckRole(roles ...string) ginfw.HandlerFunc {
	return GetHandler(nil, &Annotation{CheckRole: roles})
}

// CheckPermission requires any of the given permissions (OR within list).
// CheckPermission 检查权限（变参 OR）。
func CheckPermission(perms ...string) ginfw.HandlerFunc {
	return GetHandler(nil, &Annotation{CheckPermission: perms})
}

// CheckPermissionRoleAnd requires both permission-set and role-set to pass (OR within each set).
// CheckPermissionRoleAnd 权限组与角色组都必须通过（组内仍为 OR）。
func CheckPermissionRoleAnd(permissions, roles []string) ginfw.HandlerFunc {
	return GetHandler(nil, &Annotation{
		CheckPermission:            permissions,
		CheckRole:                  roles,
		PermissionAndRoleOperation: core.PermissionAndRoleOperationAND,
	})
}

// CheckPermissionRoleOr requires either permission-set or role-set to pass (OR within each set).
// CheckPermissionRoleOr 权限组或角色组任一通过即可（组内仍为 OR）。
func CheckPermissionRoleOr(permissions, roles []string) ginfw.HandlerFunc {
	return GetHandler(nil, &Annotation{
		CheckPermission:            permissions,
		CheckRole:                  roles,
		PermissionAndRoleOperation: core.PermissionAndRoleOperationOR,
	})
}

// CheckDisable rejects disabled accounts | 检查是否被封禁
func CheckDisable() ginfw.HandlerFunc {
	return GetHandler(nil, &Annotation{CheckDisable: true})
}

// Ignore skips authentication | 忽略认证
func Ignore() ginfw.HandlerFunc {
	return GetHandler(nil, &Annotation{Ignore: true})
}

// WithAnnotation applies a custom annotation | 使用自定义注解
func WithAnnotation(ann *Annotation) ginfw.HandlerFunc {
	return GetHandler(nil, ann)
}

// ProcessStructAnnotations processes annotations on struct tags (legacy stub behavior kept).
// ProcessStructAnnotations 处理结构体注解（保持原行为）。
func ProcessStructAnnotations(handler interface{}) ginfw.HandlerFunc {
	handlerValue := reflect.ValueOf(handler)
	handlerType := reflect.TypeOf(handler)

	methodName := "unknown"
	if handlerType.Kind() == reflect.Ptr {
		handlerType = handlerType.Elem()
	}
	if handlerType.Kind() == reflect.Struct {
		methodName = handlerType.Name()
	}

	ann := parseMethodAnnotation(handlerType, methodName)

	return GetHandler(func(c *ginfw.Context) {
		handlerValue.MethodByName("ServeHTTP").Call([]reflect.Value{reflect.ValueOf(c)})
	}, ann)
}

func parseMethodAnnotation(t reflect.Type, methodName string) *Annotation {
	return &Annotation{}
}

// HandlerWithAnnotations wraps a handler with annotations | 带注解的处理器包装器
type HandlerWithAnnotations struct {
	Handler     interface{}
	Annotations []*Annotation
}

// NewHandlerWithAnnotations creates a wrapped handler | 创建带注解的处理器
func NewHandlerWithAnnotations(handler interface{}, annotations ...*Annotation) *HandlerWithAnnotations {
	return &HandlerWithAnnotations{
		Handler:     handler,
		Annotations: annotations,
	}
}

// ToGinHandler converts to gin.HandlerFunc | 转换为 Gin 处理器
func (h *HandlerWithAnnotations) ToGinHandler() ginfw.HandlerFunc {
	return GetHandler(h.Handler, h.Annotations...)
}

// Middleware is the middleware form sharing ensureLoggedIn + runAnnotationChecks with GetHandler.
// Middleware 中间件版本（与 GetHandler 共用校验逻辑，仅末尾 Next）。
func Middleware(annotations ...*Annotation) ginfw.HandlerFunc {
	return func(c *ginfw.Context) {
		ann := firstAnnotation(annotations)
		if ann != nil && ann.Ignore {
			c.Next()
			return
		}
		loginID, ok := ensureLoggedIn(c)
		if !ok {
			return
		}
		if !runAnnotationChecks(c, loginID, ann) {
			return
		}
		c.Next()
	}
}
