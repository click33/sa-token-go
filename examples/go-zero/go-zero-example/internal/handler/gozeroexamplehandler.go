package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/sa-tokens/sa-token-go/examples/go-zero/go-zero-example/internal/logic"
	"github.com/sa-tokens/sa-token-go/examples/go-zero/go-zero-example/internal/svc"
	"github.com/sa-tokens/sa-token-go/examples/go-zero/go-zero-example/internal/types"
)

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewAuthLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func LogoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return svcCtx.Plugin.LogoutHandler
}

func UserInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return svcCtx.Plugin.UserInfoHandler
}
