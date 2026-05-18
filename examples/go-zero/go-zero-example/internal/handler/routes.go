package handler

import (
	"encoding/json"
	"net/http"

	"github.com/zeromicro/go-zero/rest"

	"github.com/click33/sa-token-go/examples/go-zero/go-zero-example/internal/svc"
	sazero "github.com/click33/sa-token-go/integrations/go-zero"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/login",
				Handler: LoginHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/public",
				Handler: publicHandler,
			},
		},
	)

	// 受保护路由：先做 Token 拦截，再做登录校验，便于业务直接读取解析后的 token
	authMiddlewares := []rest.Middleware{
		serverCtx.Plugin.TokenInterceptor(),
		serverCtx.Plugin.AuthMiddleware(),
	}

	server.AddRoutes(
		rest.WithMiddlewares(authMiddlewares,
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/api/token",
				Handler: tokenHandler,
			},
		),
	)

	server.AddRoutes(
		rest.WithMiddlewares(authMiddlewares,
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/api/user",
				Handler: UserInfoHandler(serverCtx),
			},
		),
	)

	server.AddRoutes(
		rest.WithMiddlewares(append(authMiddlewares, serverCtx.Plugin.PermissionRequired("admin:*")),
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/api/admin",
				Handler: adminHandler,
			},
		),
	)
}

func publicHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "public access",
	})
}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tokenFromCtx": sazero.GetTokenFromCtx(r),
	})
}

func adminHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "admin ok",
	})
}
