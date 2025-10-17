package main

import (
	sagin "github.com/click33/sa-token-go/integrations/gin"
	"github.com/click33/sa-token-go/storage/memory"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"log"
)

func main() {
	// 加载配置
	viper.SetConfigFile("configs/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: No config file found, using defaults: %v", err)
	}

	// 初始化存储
	storage := memory.NewStorage()

	// 创建配置 (现在可以直接使用 sagin 包的函数)
	config := sagin.DefaultConfig()
	if viper.IsSet("token.timeout") {
		config.Timeout = viper.GetInt64("token.timeout")
	}
	if viper.IsSet("token.active_timeout") {
		config.ActiveTimeout = viper.GetInt64("token.active_timeout")
	}

	// 创建管理器 (现在可以直接使用 sagin 包的函数)
	manager := sagin.NewManager(storage, config)

	// 创建 Gin 插件
	plugin := sagin.NewPlugin(manager)

	// 设置路由
	r := gin.Default()

	// 公开路由
	r.POST("/login", plugin.LoginHandler)
	r.GET("/public", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "公开访问"})
	})

	// 受保护路由
	protected := r.Group("/api")
	protected.Use(plugin.AuthMiddleware())
	{
		protected.GET("/user", plugin.UserInfoHandler)
		protected.GET("/admin", plugin.AdminOnlyHandler)
	}

	// 启动服务器
	port := "8080"
	if viper.IsSet("server.port") {
		port = viper.GetString("server.port")
	}

	log.Printf("服务器启动在端口: %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
