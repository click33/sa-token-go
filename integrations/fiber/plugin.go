package fiber

import (
	"github.com/click33/sa-token-go/core"
	"github.com/gofiber/fiber/v2"
)

// Plugin Fiber plugin for Sa-Token | Fiber插件
type Plugin struct {
	manager *core.Manager
}

// NewPlugin creates a Fiber plugin | 创建Fiber插件
func NewPlugin(manager *core.Manager) *Plugin {
	return &Plugin{
		manager: manager,
	}
}

// AuthMiddleware authentication middleware | 认证中间件
func (p *Plugin) AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := NewFiberContext(c)
		saCtx := core.NewContext(ctx, p.manager)

		if err := saCtx.CheckLogin(); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    401,
				"message": "未登录",
			})
		}

		c.Locals("satoken", saCtx)
		return c.Next()
	}
}

// PermissionRequired permission validation middleware | 权限验证中间件
func (p *Plugin) PermissionRequired(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := NewFiberContext(c)
		saCtx := core.NewContext(ctx, p.manager)

		if err := saCtx.CheckLogin(); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    401,
				"message": "未登录",
			})
		}

		if !saCtx.HasPermission(permission) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code":    403,
				"message": "权限不足",
			})
		}

		c.Locals("satoken", saCtx)
		return c.Next()
	}
}

// RoleRequired role validation middleware | 角色验证中间件
func (p *Plugin) RoleRequired(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := NewFiberContext(c)
		saCtx := core.NewContext(ctx, p.manager)

		if err := saCtx.CheckLogin(); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    401,
				"message": "未登录",
			})
		}

		if !saCtx.HasRole(role) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code":    403,
				"message": "权限不足",
			})
		}

		c.Locals("satoken", saCtx)
		return c.Next()
	}
}

// LoginHandler 登录处理器
func (p *Plugin) LoginHandler(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Device   string `json:"device"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": "登录失败",
		})
	}

	return c.JSON(fiber.Map{
		"code":    200,
		"message": "登录成功",
		"data": fiber.Map{
			"token": token,
		},
	})
}

// GetSaToken 从Fiber上下文获取Sa-Token上下文
func GetSaToken(c *fiber.Ctx) (*core.SaTokenContext, bool) {
	satoken := c.Locals("satoken")
	if satoken == nil {
		return nil, false
	}
	ctx, ok := satoken.(*core.SaTokenContext)
	return ctx, ok
}
