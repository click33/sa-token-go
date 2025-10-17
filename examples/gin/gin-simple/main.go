package main

import (
	"log"

	sagin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/click33/sa-token-go/storage/memory"
	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化存储 | Initialize storage
	storage := memory.NewStorage()

	// 创建配置 (只需要 sagin 包!) | Create config (only need sagin package!)
	config := sagin.DefaultConfig()
	config.TokenName = "token"
	config.Timeout = 7200
	config.IsPrint = true

	// 创建管理器 | Create manager
	manager := sagin.NewManager(storage, config)

	// 设置全局管理器 | Set global manager
	sagin.SetManager(manager)

	// 创建 Gin 插件 | Create Gin plugin
	plugin := sagin.NewPlugin(manager)

	// 创建路由 | Create router
	r := gin.Default()

	// 登录接口 | Login endpoint
	r.POST("/login", func(c *gin.Context) {
		userID := c.PostForm("user_id")
		if userID == "" {
			c.JSON(400, gin.H{"error": "user_id is required"})
			return
		}

		// 使用 sagin 包的全局函数登录 | Use sagin package global function to login
		token, err := sagin.Login(userID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"message": "登录成功",
			"token":   token,
		})
	})

	// 登出接口 | Logout endpoint
	r.POST("/logout", func(c *gin.Context) {
		token := c.GetHeader("token")
		if token == "" {
			c.JSON(400, gin.H{"error": "token is required"})
			return
		}

		// 使用 sagin 包的全局函数登出 | Use sagin package global function to logout
		if err := sagin.LogoutByToken(token); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "登出成功"})
	})

	// 检查登录状态 | Check login status
	r.GET("/check", func(c *gin.Context) {
		token := c.GetHeader("token")
		if token == "" {
			c.JSON(400, gin.H{"error": "token is required"})
			return
		}

		// 使用 sagin 包的全局函数检查登录 | Use sagin package global function to check login
		isLogin := sagin.IsLogin(token)
		if !isLogin {
			c.JSON(401, gin.H{"error": "未登录"})
			return
		}

		// 获取登录ID | Get login ID
		loginID, _ := sagin.GetLoginID(token)

		c.JSON(200, gin.H{
			"message":  "已登录",
			"login_id": loginID,
		})
	})

	// 受保护的路由组 | Protected route group
	protected := r.Group("/api")
	protected.Use(plugin.AuthMiddleware())
	{
		// 用户信息 | User info
		protected.GET("/user", func(c *gin.Context) {
			token := c.GetHeader("token")
			loginID, _ := sagin.GetLoginID(token)

			c.JSON(200, gin.H{
				"user_id": loginID,
				"name":    "User " + loginID,
			})
		})

		// 踢人下线 | Kickout user
		protected.POST("/kickout/:user_id", func(c *gin.Context) {
			userID := c.Param("user_id")

			// 使用 sagin 包的全局函数踢人 | Use sagin package global function to kickout
			if err := sagin.Kickout(userID); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, gin.H{"message": "踢人成功"})
		})
	}

	// 启动服务器 | Start server
	log.Println("服务器启动在端口: 8080")
	log.Println("示例: curl -X POST http://localhost:8080/login -d 'user_id=1000'")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
