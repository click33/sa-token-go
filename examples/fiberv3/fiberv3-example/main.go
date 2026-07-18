package fiberv3_example

import (
	"github.com/gofiber/fiber/v3/log"
	"github.com/sa-tokens/sa-token-go/core"
	safiberv3 "github.com/sa-tokens/sa-token-go/integrations/fiberv3"
	"github.com/sa-tokens/sa-token-go/storage/memory"
	"github.com/sa-tokens/sa-token-go/stputil"
)

func init() {
	// 🎯 一行初始化！显示启动 Banner
	stputil.SetManager(
		core.NewBuilder().
			Storage(memory.NewStorage()).
			TokenName("Authorization").
			Timeout(86400).                      // 24小时
			TokenStyle(core.TokenStyleRandom64). // Token风格
			IsPrintBanner(true).                 // 显示启动Banner
			AutoRenew(true).
			Build(),
	)
}
func main() {
	// 创建Fiber插件
	plugin := safiberv3.NewPlugin(stputil.GetManager())
	app := fiber.New()
	api := app.Group("/api")
	// TokenInterceptor 拦截器：负责从 Header 提取 Token 并写入上下文
	api.Use(plugin.TokenInterceptor())
	// 简单用法
	api.Get("/user/profile", plugin.AuthMiddleware(), func(c fiber.Ctx) error {
		saCtx, ok := safiberv3.GetSaToken(c)
		if ok {
			log.Debug(saCtx.GetTokenValue())
		}
		return c.SendString("Hello, World!")
	})
	// 复杂用法
	api.Use(plugin.PathAuthMiddleware(&core.PathAuthConfig{
		Include: []string{"/**"},         // 默认拦截全局 /api 下的所有路由
		Exclude: []string{"/user/login"}, // 唯独放行登录接口
	}))
	log.Fatal(app.Listen(":3000"))
}
