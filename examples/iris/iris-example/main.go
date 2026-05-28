package main

import (
	"log"

	"github.com/kataras/iris/v12"
	sairis "github.com/sa-tokens/sa-token-go/integrations/iris"
	"github.com/sa-tokens/sa-token-go/storage/memory"
)

func main() {
	// 初始化存储 | Initialize storage
	storage := memory.NewStorage()

	// 创建配置（只需要 sairis 包!） | Create config (only need sairis package!)
	config := sairis.DefaultConfig()
	config.TokenName = "token"
	config.Timeout = 7200
	config.IsPrintBanner = true

	// 创建管理器 | Create manager
	manager := sairis.NewManager(storage, config)

	// 设置全局管理器 | Set global manager
	sairis.SetManager(manager)

	// 创建 Iris 插件 | Create Iris plugin
	plugin := sairis.NewPlugin(manager)

	// 创建应用 | Create app
	app := iris.New()

	// 登录接口 | Login endpoint
	app.Post("/login", func(c iris.Context) {
		// SECURITY NOTICE: this demo endpoint does NOT validate username/password.
		// You MUST replace it with real authentication in production, otherwise it
		// behaves as passwordless login.
		// 安全提示：本示例为演示用途，未做用户名/密码校验。
		// 生产环境必须替换为真实的用户身份验证逻辑，否则等同无密码登录。
		userID := c.PostValue("user_id")
		if userID == "" {
			c.StatusCode(iris.StatusBadRequest)
			_ = c.JSON(iris.Map{"error": "user_id is required"})
			return
		}

		token, err := sairis.Login(userID)
		if err != nil {
			c.StatusCode(iris.StatusInternalServerError)
			_ = c.JSON(iris.Map{"error": err.Error()})
			return
		}

		_ = c.JSON(iris.Map{
			"message": "登录成功",
			"token":   token,
		})
	})

	// 登出接口 | Logout endpoint
	app.Post("/logout", func(c iris.Context) {
		token := c.GetHeader("token")
		if token == "" {
			c.StatusCode(iris.StatusBadRequest)
			_ = c.JSON(iris.Map{"error": "token is required"})
			return
		}

		if err := sairis.LogoutByToken(token); err != nil {
			c.StatusCode(iris.StatusInternalServerError)
			_ = c.JSON(iris.Map{"error": err.Error()})
			return
		}

		_ = c.JSON(iris.Map{"message": "登出成功"})
	})

	// 检查登录状态 | Check login status
	app.Get("/check", func(c iris.Context) {
		token := c.GetHeader("token")
		if token == "" {
			c.StatusCode(iris.StatusBadRequest)
			_ = c.JSON(iris.Map{"error": "token is required"})
			return
		}

		if !sairis.IsLogin(token) {
			c.StatusCode(iris.StatusUnauthorized)
			_ = c.JSON(iris.Map{"error": "未登录"})
			return
		}

		loginID, _ := sairis.GetLoginID(token)
		_ = c.JSON(iris.Map{
			"message":  "已登录",
			"login_id": loginID,
		})
	})

	// 公共路由 | Public route - skip authentication
	app.Get("/public", sairis.Ignore(), func(c iris.Context) {
		_ = c.JSON(iris.Map{"message": "public data"})
	})

	// 需要登录 | Login required
	app.Get("/user", sairis.CheckLogin(), func(c iris.Context) {
		_ = c.JSON(iris.Map{"message": "user data"})
	})

	// 需要 admin 角色 | Admin role required
	app.Get("/admin", sairis.CheckRole("admin"), func(c iris.Context) {
		_ = c.JSON(iris.Map{"message": "admin area"})
	})

	// 需要 admin:* 权限 | Permission required
	app.Get("/manage", sairis.CheckPermission("admin:*"), func(c iris.Context) {
		_ = c.JSON(iris.Map{"message": "management area"})
	})

	// 受保护路由组：先 TokenInterceptor，再 AuthMiddleware
	// Protected group: TokenInterceptor first, then AuthMiddleware
	protected := app.Party("/api", plugin.TokenInterceptor(), plugin.AuthMiddleware())
	protected.Get("/token", func(c iris.Context) {
		token := sairis.GetTokenFromCtx(c)
		_ = c.JSON(iris.Map{"token": token})
	})

	log.Println("服务器启动在端口: 8080")
	log.Println("示例: curl -X POST http://localhost:8080/login -d 'user_id=1000'")
	if err := app.Listen(":8080"); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
