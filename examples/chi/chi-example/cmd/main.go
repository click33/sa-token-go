package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/click33/sa-token-go/core"
	sachi "github.com/click33/sa-token-go/integrations/chi"
	"github.com/click33/sa-token-go/storage/memory"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	// 创建Chi插件
	plugin := sachi.NewPlugin(manager)

	// 创建路由
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 公开路由
	r.Post("/login", plugin.LoginHandler)
	r.Get("/public", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "公开访问",
		})
	})

	// 受保护路由
	r.Group(func(r chi.Router) {
		r.Use(plugin.AuthMiddleware())
		r.Get("/api/user/info", func(w http.ResponseWriter, r *http.Request) {
			saCtx, _ := sachi.GetSaToken(r)
			loginID, _ := saCtx.GetLoginID()
			permissions, _ := manager.GetPermissions(loginID)
			roles, _ := manager.GetRoles(loginID)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    200,
				"message": "success",
				"data": map[string]interface{}{
					"loginId":     loginID,
					"permissions": permissions,
					"roles":       roles,
				},
			})
		})
	})

	// 启动服务器
	log.Println("服务器启动在端口: 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
