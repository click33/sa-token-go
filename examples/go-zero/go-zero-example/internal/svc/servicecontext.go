package svc

import (
	"github.com/click33/sa-token-go/core"
	"github.com/click33/sa-token-go/examples/go-zero/go-zero-example/internal/config"
	sazero "github.com/click33/sa-token-go/integrations/go-zero"
	"github.com/click33/sa-token-go/storage/memory"
)

type ServiceContext struct {
	Config  config.Config
	Manager *core.Manager
	Plugin  *sazero.Plugin
}

func NewServiceContext(c config.Config) *ServiceContext {
	storage := memory.NewStorage()

	cfg := core.DefaultConfig()
	cfg.TokenName = c.TokenName
	cfg.Timeout = c.TokenTimeout

	manager := core.NewManager(storage, cfg)
	sazero.SetManager(manager)

	plugin := sazero.NewPlugin(manager)

	return &ServiceContext{
		Config:  c,
		Manager: manager,
		Plugin:  plugin,
	}
}
