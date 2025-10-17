package chi

import (
	"encoding/json"
	"net/http"

	"github.com/click33/sa-token-go/core"
)

// Plugin Chi plugin for Sa-Token | Chi插件
type Plugin struct {
	manager *core.Manager
}

// NewPlugin creates a Chi plugin | 创建Chi插件
func NewPlugin(manager *core.Manager) *Plugin {
	return &Plugin{
		manager: manager,
	}
}

// AuthMiddleware authentication middleware | 认证中间件
func (p *Plugin) AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := NewChiContext(w, r)
			saCtx := core.NewContext(ctx, p.manager)

			if err := saCtx.CheckLogin(); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"code":    401,
					"message": "未登录",
				})
				return
			}

			// Store Sa-Token context | 存储Sa-Token上下文
			ctx.Set("satoken", saCtx)
			next.ServeHTTP(w, r)
		})
	}
}

// PermissionRequired permission validation middleware | 权限验证中间件
func (p *Plugin) PermissionRequired(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := NewChiContext(w, r)
			saCtx := core.NewContext(ctx, p.manager)

			if err := saCtx.CheckLogin(); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"code":    401,
					"message": "未登录",
				})
				return
			}

			if !saCtx.HasPermission(permission) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"code":    403,
					"message": "权限不足",
				})
				return
			}

			ctx.Set("satoken", saCtx)
			next.ServeHTTP(w, r)
		})
	}
}

// RoleRequired role validation middleware | 角色验证中间件
func (p *Plugin) RoleRequired(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := NewChiContext(w, r)
			saCtx := core.NewContext(ctx, p.manager)

			if err := saCtx.CheckLogin(); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"code":    401,
					"message": "未登录",
				})
				return
			}

			if !saCtx.HasRole(role) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"code":    403,
					"message": "权限不足",
				})
				return
			}

			ctx.Set("satoken", saCtx)
			next.ServeHTTP(w, r)
		})
	}
}

// LoginHandler 登录处理器
func (p *Plugin) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Device   string `json:"device"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	device := req.Device
	if device == "" {
		device = "default"
	}

	token, err := p.manager.Login(req.Username, device)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    500,
			"message": "登录失败",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    200,
		"message": "登录成功",
		"data": map[string]interface{}{
			"token": token,
		},
	})
}

// GetSaToken 从请求上下文获取Sa-Token上下文
func GetSaToken(r *http.Request) (*core.SaTokenContext, bool) {
	satoken := r.Context().Value("satoken")
	if satoken == nil {
		return nil, false
	}
	ctx, ok := satoken.(*core.SaTokenContext)
	return ctx, ok
}
