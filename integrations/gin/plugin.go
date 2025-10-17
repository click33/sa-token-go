package gin

import (
	"net/http"

	"github.com/click33/sa-token-go/core"
	"github.com/gin-gonic/gin"
)

// Plugin Gin plugin for Sa-Token | Gin插件
type Plugin struct {
	manager *core.Manager
}

// NewPlugin creates a Gin plugin | 创建Gin插件
func NewPlugin(manager *core.Manager) *Plugin {
	return &Plugin{
		manager: manager,
	}
}

// AuthMiddleware authentication middleware | 认证中间件
func (p *Plugin) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := NewGinContext(c)
		saCtx := core.NewContext(ctx, p.manager)

		// Check login | 检查登录
		if err := saCtx.CheckLogin(); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未登录",
			})
			c.Abort()
			return
		}

		// Store Sa-Token context in Gin context | 将Sa-Token上下文存储到Gin上下文
		c.Set("satoken", saCtx)
		c.Next()
	}
}

// PermissionRequired permission validation middleware | 权限验证中间件
func (p *Plugin) PermissionRequired(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := NewGinContext(c)
		saCtx := core.NewContext(ctx, p.manager)

		// Check login | 检查登录
		if err := saCtx.CheckLogin(); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未登录",
			})
			c.Abort()
			return
		}

		// Check permission | 检查权限
		if !saCtx.HasPermission(permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		c.Set("satoken", saCtx)
		c.Next()
	}
}

// RoleRequired role validation middleware | 角色验证中间件
func (p *Plugin) RoleRequired(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := NewGinContext(c)
		saCtx := core.NewContext(ctx, p.manager)

		// Check login | 检查登录
		if err := saCtx.CheckLogin(); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未登录",
			})
			c.Abort()
			return
		}

		// Check role | 检查角色
		if !saCtx.HasRole(role) {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		c.Set("satoken", saCtx)
		c.Next()
	}
}

// LoginHandler login handler example | 登录处理器示例
func (p *Plugin) LoginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Device   string `json:"device"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// TODO: Validate username and password (should call your user service) | 验证用户名密码（这里应该调用你的用户服务）
	// if !validateUser(req.Username, req.Password) { ... }

	// Login | 登录
	device := req.Device
	if device == "" {
		device = "default"
	}

	token, err := p.manager.Login(req.Username, device)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "登录失败",
		})
		return
	}

	// Set cookie (optional) | 设置Cookie（可选）
	cfg := p.manager.GetConfig()
	if cfg.IsReadCookie {
		maxAge := int(cfg.Timeout)
		if maxAge < 0 {
			maxAge = 0
		}
		c.SetCookie(
			cfg.TokenName,
			token,
			maxAge,
			cfg.CookieConfig.Path,
			cfg.CookieConfig.Domain,
			cfg.CookieConfig.Secure,
			cfg.CookieConfig.HttpOnly,
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": gin.H{
			"token": token,
		},
	})
}

// LogoutHandler logout handler | 登出处理器
func (p *Plugin) LogoutHandler(c *gin.Context) {
	ctx := NewGinContext(c)
	saCtx := core.NewContext(ctx, p.manager)

	loginID, err := saCtx.GetLoginID()
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	if err := p.manager.Logout(loginID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "登出失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登出成功",
	})
}

// UserInfoHandler user info handler example | 获取用户信息处理器示例
func (p *Plugin) UserInfoHandler(c *gin.Context) {
	ctx := NewGinContext(c)
	saCtx := core.NewContext(ctx, p.manager)

	loginID, err := saCtx.GetLoginID()
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	// Get user permissions and roles | 获取用户权限和角色
	permissions, _ := p.manager.GetPermissions(loginID)
	roles, _ := p.manager.GetRoles(loginID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"loginId":     loginID,
			"permissions": permissions,
			"roles":       roles,
		},
	})
}

// GetSaToken gets Sa-Token context from Gin context | 从Gin上下文获取Sa-Token上下文
func GetSaToken(c *gin.Context) (*core.SaTokenContext, bool) {
	satoken, exists := c.Get("satoken")
	if !exists {
		return nil, false
	}
	ctx, ok := satoken.(*core.SaTokenContext)
	return ctx, ok
}
