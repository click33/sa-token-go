package main

import (
	"log"

	"github.com/sa-tokens/sa-token-go/core"
	safiber "github.com/sa-tokens/sa-token-go/integrations/fiber"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// 创建存储
	storage := memory.NewStorage()

	// 创建配置
	config := core.DefaultConfig()
	config.TokenName = "Authorization"
	config.Timeout = 7200 // 2小时

	// 创建管理器
	manager := core.NewManager(storage, config)

	// 创建Fiber插件
	plugin := safiber.NewPlugin(manager)

	// 创建Fiber应用
	app := fiber.New()
	app.Use(logger.New())
	app.Use(recover.New())

	// 公开路由
	app.Post("/login", plugin.LoginHandler)
	app.Get("/public", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "公开访问",
		})
	})

	// 受保护路由
	api := app.Group("/api")
	// 先执行 TokenInterceptor，再执行 AuthMiddleware，避免业务层重复写 token 提取逻辑
	api.Use(plugin.TokenInterceptor())
	api.Use(plugin.AuthMiddleware())
	{
		api.Get("/token", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"tokenFromCtx": safiber.GetTokenFromCtx(c),
			})
		})
		api.Get("/user/info", func(c *fiber.Ctx) error {
			saCtx, _ := safiber.GetSaToken(c)
			loginID, _ := saCtx.GetLoginID()
			permissions, _ := manager.GetPermissions(loginID)
			roles, _ := manager.GetRoles(loginID)

			return c.JSON(fiber.Map{
				"code":    200,
				"message": "success",
				"data": fiber.Map{
					"loginId":     loginID,
					"permissions": permissions,
					"roles":       roles,
				},
			})
		})
	}

	// 启动服务器
	log.Println("服务器启动在端口: 8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
