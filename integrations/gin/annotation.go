package gin

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/click33/sa-token-go/stputil"
	ginfw "github.com/gin-gonic/gin"
)

// Annotation constants | 注解常量
const (
	TagSaCheckLogin      = "sa_check_login"
	TagSaCheckRole       = "sa_check_role"
	TagSaCheckPermission = "sa_check_permission"
	TagSaCheckDisable    = "sa_check_disable"
	TagSaIgnore          = "sa_ignore"
)

// Annotation annotation structure | 注解结构体
type Annotation struct {
	CheckLogin      bool     `json:"checkLogin"`
	CheckRole       []string `json:"checkRole"`
	CheckPermission []string `json:"checkPermission"`
	CheckDisable    bool     `json:"checkDisable"`
	Ignore          bool     `json:"ignore"`
}

// ParseTag parses struct tags | 解析结构体标签
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

// Validate validates if annotation is valid | 验证注解是否有效
func (a *Annotation) Validate() bool {
	if a.Ignore {
		return true // When ignore is true, other checks are invalid | 忽略认证时，其他检查无效
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

	// At most one check type allowed | 最多只能有一个检查类型
	return count <= 1
}

// GetHandler gets handler with annotations | 获取带注解的处理器
func GetHandler(handler interface{}, annotations ...*Annotation) ginfw.HandlerFunc {
	return func(c *ginfw.Context) {
		// Check if authentication should be ignored | 检查是否忽略认证
		if len(annotations) > 0 && annotations[0].Ignore {
			if handler != nil {
				handler.(func(*ginfw.Context))(c)
			} else {
				c.Next()
			}
			return
		}

		// Get token | 获取Token
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.GetHeader("satoken")
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ginfw.H{
				"code":    401,
				"message": "未登录",
			})
			return
		}

		// Check login | 检查登录
		if !stputil.IsLogin(token) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ginfw.H{
				"code":    401,
				"message": "未登录",
			})
			return
		}

		// Get login ID | 获取登录ID
		loginID, err := stputil.GetLoginID(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ginfw.H{
				"code":    401,
				"message": "登录状态无效",
			})
			return
		}

		// Check if account is disabled | 检查是否被封禁
		if len(annotations) > 0 && annotations[0].CheckDisable {
			if stputil.IsDisable(loginID) {
				c.AbortWithStatusJSON(http.StatusForbidden, ginfw.H{
					"code":    403,
					"message": "账号已被封禁",
				})
				return
			}
		}

		// Check permission | 检查权限
		if len(annotations) > 0 && len(annotations[0].CheckPermission) > 0 {
			hasPermission := false
			for _, perm := range annotations[0].CheckPermission {
				if stputil.HasPermission(loginID, strings.TrimSpace(perm)) {
					hasPermission = true
					break
				}
			}
			if !hasPermission {
				c.AbortWithStatusJSON(http.StatusForbidden, ginfw.H{
					"code":    403,
					"message": "权限不足",
				})
				return
			}
		}

		// Check role | 检查角色
		if len(annotations) > 0 && len(annotations[0].CheckRole) > 0 {
			hasRole := false
			for _, role := range annotations[0].CheckRole {
				if stputil.HasRole(loginID, strings.TrimSpace(role)) {
					hasRole = true
					break
				}
			}
			if !hasRole {
				c.AbortWithStatusJSON(http.StatusForbidden, ginfw.H{
					"code":    403,
					"message": "角色不足",
				})
				return
			}
		}

		// All checks passed, execute original handler or continue | 所有检查通过，执行原函数或继续
		if handler != nil {
			handler.(func(*ginfw.Context))(c)
		} else {
			c.Next()
		}
	}
}

// Decorator functions | 装饰器函数

// CheckLogin decorator for login checking | 检查登录装饰器
func CheckLogin() ginfw.HandlerFunc {
	return GetHandler(nil, &Annotation{CheckLogin: true})
}

// CheckRole decorator for role checking | 检查角色装饰器
func CheckRole(roles ...string) ginfw.HandlerFunc {
	return GetHandler(nil, &Annotation{CheckRole: roles})
}

// CheckPermission decorator for permission checking | 检查权限装饰器
func CheckPermission(perms ...string) ginfw.HandlerFunc {
	return GetHandler(nil, &Annotation{CheckPermission: perms})
}

// CheckDisable decorator for checking if account is disabled | 检查是否被封禁装饰器
func CheckDisable() ginfw.HandlerFunc {
	return GetHandler(nil, &Annotation{CheckDisable: true})
}

// Ignore decorator to ignore authentication | 忽略认证装饰器
func Ignore() ginfw.HandlerFunc {
	return GetHandler(nil, &Annotation{Ignore: true})
}

// WithAnnotation decorator with custom annotation | 使用自定义注解装饰器
func WithAnnotation(ann *Annotation) ginfw.HandlerFunc {
	return GetHandler(nil, ann)
}

// ProcessStructAnnotations processes annotations on struct tags | 处理结构体上的注解标签
func ProcessStructAnnotations(handler interface{}) ginfw.HandlerFunc {
	handlerValue := reflect.ValueOf(handler)
	handlerType := reflect.TypeOf(handler)

	// Find method name, usually the last path segment | 查找方法名，通常是最后一个路径段
	methodName := "unknown"
	if handlerType.Kind() == reflect.Ptr {
		handlerType = handlerType.Elem()
	}
	if handlerType.Kind() == reflect.Struct {
		methodName = handlerType.Name()
	}

	// Parse method annotations | 解析方法上的注解标签
	ann := parseMethodAnnotation(handlerType, methodName)

	return GetHandler(func(c *ginfw.Context) {
		handlerValue.MethodByName("ServeHTTP").Call([]reflect.Value{reflect.ValueOf(c)})
	}, ann)
}

// parseMethodAnnotation parses method annotations | 解析方法注解
func parseMethodAnnotation(t reflect.Type, methodName string) *Annotation {
	// Simplified implementation, returns empty annotation | 简化实现，直接返回空注解
	return &Annotation{}
}

// HandlerWithAnnotations 带注解的处理器包装器
type HandlerWithAnnotations struct {
	Handler     interface{}
	Annotations []*Annotation
}

// NewHandlerWithAnnotations 创建带注解的处理器
func NewHandlerWithAnnotations(handler interface{}, annotations ...*Annotation) *HandlerWithAnnotations {
	return &HandlerWithAnnotations{
		Handler:     handler,
		Annotations: annotations,
	}
}

// ToGinHandler 转换为Gin处理器
func (h *HandlerWithAnnotations) ToGinHandler() ginfw.HandlerFunc {
	return GetHandler(h.Handler, h.Annotations...)
}

// Middleware 创建中间件版本
func Middleware(annotations ...*Annotation) ginfw.HandlerFunc {
	return func(c *ginfw.Context) {

		// 检查是否忽略认证
		if len(annotations) > 0 && annotations[0].Ignore {
			c.Next()
			return
		}

		// 获取Token
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.GetHeader("satoken")
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ginfw.H{
				"code":    401,
				"message": "未登录",
			})
			return
		}

		// 检查登录
		if !stputil.IsLogin(token) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ginfw.H{
				"code":    401,
				"message": "未登录",
			})
			return
		}

		// 获取登录ID
		loginID, err := stputil.GetLoginID(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ginfw.H{
				"code":    401,
				"message": "登录状态无效",
			})
			return
		}

		// 检查是否被封禁
		if len(annotations) > 0 && annotations[0].CheckDisable {
			if stputil.IsDisable(loginID) {
				c.AbortWithStatusJSON(http.StatusForbidden, ginfw.H{
					"code":    403,
					"message": "账号已被封禁",
				})
				return
			}
		}

		// 检查权限
		if len(annotations) > 0 && len(annotations[0].CheckPermission) > 0 {
			hasPermission := false
			for _, perm := range annotations[0].CheckPermission {
				if stputil.HasPermission(loginID, strings.TrimSpace(perm)) {
					hasPermission = true
					break
				}
			}
			if !hasPermission {
				c.AbortWithStatusJSON(http.StatusForbidden, ginfw.H{
					"code":    403,
					"message": "权限不足",
				})
				return
			}
		}

		// 检查角色
		if len(annotations) > 0 && len(annotations[0].CheckRole) > 0 {
			hasRole := false
			for _, role := range annotations[0].CheckRole {
				if stputil.HasRole(loginID, strings.TrimSpace(role)) {
					hasRole = true
					break
				}
			}
			if !hasRole {
				c.AbortWithStatusJSON(http.StatusForbidden, ginfw.H{
					"code":    403,
					"message": "角色不足",
				})
				return
			}
		}

		// 所有检查通过，继续下一个处理器
		c.Next()
	}
}
