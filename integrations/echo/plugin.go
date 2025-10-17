package echo

import (
	"net/http"

	"github.com/click33/sa-token-go/core"
	"github.com/labstack/echo/v4"
)

// Plugin Echo plugin for Sa-Token | Echo插件
type Plugin struct {
	manager *core.Manager
}

// NewPlugin creates an Echo plugin | 创建Echo插件
func NewPlugin(manager *core.Manager) *Plugin {
	return &Plugin{
		manager: manager,
	}
}

// AuthMiddleware authentication middleware | 认证中间件
func (p *Plugin) AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := NewEchoContext(c)
			saCtx := core.NewContext(ctx, p.manager)

			if err := saCtx.CheckLogin(); err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"code":    401,
					"message": "未登录",
				})
			}

			c.Set("satoken", saCtx)
			return next(c)
		}
	}
}

// PermissionRequired permission validation middleware | 权限验证中间件
func (p *Plugin) PermissionRequired(permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := NewEchoContext(c)
			saCtx := core.NewContext(ctx, p.manager)

			if err := saCtx.CheckLogin(); err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"code":    401,
					"message": "未登录",
				})
			}

			if !saCtx.HasPermission(permission) {
				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"code":    403,
					"message": "权限不足",
				})
			}

			c.Set("satoken", saCtx)
			return next(c)
		}
	}
}

// RoleRequired role validation middleware | 角色验证中间件
func (p *Plugin) RoleRequired(role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := NewEchoContext(c)
			saCtx := core.NewContext(ctx, p.manager)

			if err := saCtx.CheckLogin(); err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"code":    401,
					"message": "未登录",
				})
			}

			if !saCtx.HasRole(role) {
				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"code":    403,
					"message": "权限不足",
				})
			}

			c.Set("satoken", saCtx)
			return next(c)
		}
	}
}

// LoginHandler 登录处理器
func (p *Plugin) LoginHandler(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Device   string `json:"device"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
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
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"code":    500,
			"message": "登录失败",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"code":    200,
		"message": "登录成功",
		"data": map[string]interface{}{
			"token": token,
		},
	})
}

// GetSaToken 从Echo上下文获取Sa-Token上下文
func GetSaToken(c echo.Context) (*core.SaTokenContext, bool) {
	satoken := c.Get("satoken")
	if satoken == nil {
		return nil, false
	}
	ctx, ok := satoken.(*core.SaTokenContext)
	return ctx, ok
}
