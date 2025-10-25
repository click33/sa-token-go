package gf

import (
	"net/http"

	"github.com/click33/sa-token-go/core"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// Plugin GoFrame plugin for Sa-Token | GoFrame插件
type Plugin struct {
	manager *core.Manager
}

// NewPlugin creates an GoFrame plugin | 创建GoFrame插件
func NewPlugin(manager *core.Manager) *Plugin {
	return &Plugin{
		manager: manager,
	}
}

// AuthMiddleware authentication middleware | 认证中间件
func (p *Plugin) AuthMiddleware() ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		ctx := NewGFContext(r)
		saCtx := core.NewContext(ctx, p.manager)
		// Check login | 检查登录
		if err := saCtx.CheckLogin(); err != nil {
			r.Response.WriteStatusExit(http.StatusUnauthorized, g.Map{
				"code":    401,
				"message": "未登录",
			})
			return
		}
		// Store Sa-Token context in GoFrame context | 将Sa-Token上下文存储到GoFrame上下文
		r.SetCtxVar("satoken", saCtx)

		r.Middleware.Next()
	}

}

// PermissionRequired permission validation middleware | 权限验证中间件
func (p *Plugin) PermissionRequired(permission string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		ctx := NewGFContext(r)
		saCtx := core.NewContext(ctx, p.manager)

		if err := saCtx.CheckLogin(); err != nil {
			r.Response.WriteStatusExit(http.StatusUnauthorized, g.Map{
				"code":    401,
				"message": "未登录",
			})
		}
		if !saCtx.HasPermission(permission) {
			r.Response.WriteStatusExit(http.StatusForbidden, g.Map{
				"code":    403,
				"message": "权限不足",
			})
		}
		r.SetCtxVar("satoken", saCtx)
		r.Middleware.Next()
	}

}

// RoleRequired role validation middleware | 角色验证中间件
func (p *Plugin) RoleRequired(role string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		ctx := NewGFContext(r)
		saCtx := core.NewContext(ctx, p.manager)

		if err := saCtx.CheckLogin(); err != nil {
			r.Response.WriteStatusExit(http.StatusUnauthorized, g.Map{
				"code":    401,
				"message": "未登录",
			})
		}

		if !saCtx.HasRole(role) {
			r.Response.WriteStatusExit(http.StatusForbidden, g.Map{
				"code":    403,
				"message": "权限不足",
			})
		}

		r.SetCtxVar("satoken", saCtx)
		r.Middleware.Next()
	}
}

// LoginHandler 登录处理器
func (p *Plugin) LoginHandler(r *ghttp.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Device   string `json:"device"`
	}

	if err := r.Parse(&req); err != nil {
		r.Response.WriteStatusExit(http.StatusBadRequest, g.Map{
			"code":    400,
			"message": "参数错误",
		})
	}

	device := req.Device
	if device == "" {
		device = "default"
	}

	token, err := p.manager.Login(req.Username, device)
	if err != nil {
		r.Response.WriteStatusExit(http.StatusInternalServerError, g.Map{
			"code":    500,
			"message": "登录失败",
		})
	}

	r.Response.WriteStatusExit(http.StatusOK, g.Map{
		"code":    200,
		"message": "登录成功",
		"data": g.Map{
			"token": token,
		},
	})
}

// UserInfoHandler user info handler example | 获取用户信息处理器示例
func (p *Plugin) UserInfoHandler(r *ghttp.Request) {
	ctx := NewGFContext(r)
	saCtx := core.NewContext(ctx, p.manager)

	loginID, err := saCtx.GetLoginID()
	if err != nil {
		r.Response.WriteStatusExit(http.StatusUnauthorized, g.Map{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	// Get user permissions and roles | 获取用户权限和角色
	permissions, _ := p.manager.GetPermissions(loginID)
	roles, _ := p.manager.GetRoles(loginID)

	r.Response.WriteStatusExit(http.StatusOK, g.Map{
		"code":    200,
		"message": "success",
		"data": g.Map{
			"loginId":     loginID,
			"permissions": permissions,
			"roles":       roles,
		},
	})
}

// GetSaToken 从GoFrame上下文获取Sa-Token上下文
func GetSaToken(r *ghttp.Request) (*core.SaTokenContext, bool) {
	satoken := r.GetCtx().Value("satoken")
	if satoken == nil {
		return nil, false
	}
	ctx, ok := satoken.(*core.SaTokenContext)
	return ctx, ok
}
