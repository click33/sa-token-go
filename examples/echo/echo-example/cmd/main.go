package main

import (
	"log"

	"github.com/click33/sa-token-go/core"
	saecho "github.com/click33/sa-token-go/integrations/echo"
	"github.com/click33/sa-token-go/storage/memory"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	// 创建Echo插件
	plugin := saecho.NewPlugin(manager)

	// 创建Echo实例
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 公开路由
	e.POST("/login", plugin.LoginHandler)
	e.GET("/public", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"message": "公开访问",
		})
	})

	// 受保护路由
	api := e.Group("/api")
	api.Use(plugin.AuthMiddleware())
	{
		api.GET("/user/info", func(c echo.Context) error {
			saCtx, _ := saecho.GetSaToken(c)
			loginID, _ := saCtx.GetLoginID()
			permissions, _ := manager.GetPermissions(loginID)
			roles, _ := manager.GetRoles(loginID)

			return c.JSON(200, map[string]interface{}{
				"code":    200,
				"message": "success",
				"data": map[string]interface{}{
					"loginId":     loginID,
					"permissions": permissions,
					"roles":       roles,
				},
			})
		})
	}

	// 启动服务器
	log.Println("服务器启动在端口: 8080")
	if err := e.Start(":8080"); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
