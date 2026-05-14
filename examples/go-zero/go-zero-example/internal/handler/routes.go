package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"

	"github.com/click33/sa-token-go/examples/go-zero/go-zero-example/internal/svc"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	// 公开路由
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/api/login",
				Handler: LoginHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/logout",
				Handler: LogoutHandler(serverCtx),
			},
		},
	)

	// 受保护路由（需认证）
	server.AddRoutes(
		rest.WithMiddleware(serverCtx.Plugin.AuthMiddleware(),
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/api/user/info",
				Handler: UserInfoHandler(serverCtx),
			},
		),
	)
}
