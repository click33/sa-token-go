package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/sa-tokens/sa-token-go/examples/go-zero/go-zero-example/internal/svc"
	"github.com/sa-tokens/sa-token-go/examples/go-zero/go-zero-example/internal/types"
	sazero "github.com/sa-tokens/sa-token-go/integrations/go-zero"
)

type AuthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthLogic {
	return &AuthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuthLogic) Login(req *types.LoginRequest) (*types.LoginResponse, error) {
	device := req.Device
	if device == "" {
		device = "default"
	}
	token, err := sazero.Login(req.Username, device)
	if err != nil {
		return nil, err
	}
	return &types.LoginResponse{Token: token}, nil
}
