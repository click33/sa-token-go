package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/sa-tokens/sa-token-go/core"
	safiberv3 "github.com/sa-tokens/sa-token-go/integrations/fiberv3"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/sa-tokens/sa-token-go/stputil"
)

func init() {
	// One-line init with in-memory storage for the demo | 一行初始化；示例环境使用内存存储
	stputil.SetManager(
		core.NewBuilder().
			Storage(memory.NewStorage()).
			TokenName("Authorization").
			Timeout(86400).
			TokenStyle(core.TokenStyleRandom64).
			IsPrintBanner(true).
			AutoRenew(true).
			Build(),
	)
}

func main() {
	plugin := safiberv3.NewPlugin(stputil.GetManager())
	app := fiber.New()

	api := app.Group("/api")
	// Parse token into Locals only (no login check) | 仅解析 Token 写入上下文，不做登录校验
	api.Use(plugin.TokenInterceptor())

	// DEMO ONLY: LoginHandler does not verify password — implement real auth in production.
	// 登录仅作演示：不校验密码，生产必须自行校验。
	api.Post("/user/login", plugin.LoginHandler)

	api.Get("/user/profile", plugin.AuthMiddleware(), func(c fiber.Ctx) error {
		if saCtx, ok := safiberv3.GetSaToken(c); ok {
			log.Debug(saCtx.GetTokenValue())
		}
		return c.SendString("Hello, World!")
	})

	// Fiber c.Path() is the full path; Exclude must be /api/user/login, not /user/login.
	// Fiber c.Path() 含完整路径；Exclude 必须写 /api/user/login 而非 /user/login。
	api.Use(plugin.PathAuthMiddleware(&core.PathAuthConfig{
		Include: []string{"/**"},
		Exclude: []string{"/api/user/login"},
	}))

	log.Fatal(app.Listen(":3000"))
}
