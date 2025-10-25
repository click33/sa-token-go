package main

import (
	"net/http"

	sagin "github.com/click33/sa-token-go/integrations/gf"
	"github.com/click33/sa-token-go/storage/memory"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func main() {
	// 初始化存储
	storage := memory.NewStorage()

	// 创建配置 (现在可以直接使用 sagin 包的函数)
	config := sagin.DefaultConfig()
	// 创建管理器 (现在可以直接使用 sagin 包的函数)
	manager := sagin.NewManager(storage, config)

	// 创建 Gin 插件
	plugin := sagin.NewPlugin(manager)
	s := g.Server()

	s.BindHandler("/", func(r *ghttp.Request) {
		r.Response.Writef(
			"Hello %s! Your Age is %d",
			r.Get("name", "unknown").String(),
			r.Get("age").Int(),
		)
	})
	// 公开路由
	s.BindHandler("/public", func(r *ghttp.Request) {
		r.Response.WriteStatusExit(
			http.StatusOK,
			g.Map{
				"message": "公开访问",
			},
		)
	})
	s.BindHandler("/login", plugin.LoginHandler)
	// 受保护路由
	protected := s.Group("/api").Middleware(plugin.AuthMiddleware())

	{
		protected.GET("/user", plugin.UserInfoHandler)
	}

	s.SetPort(8000)
	s.Run()
}
