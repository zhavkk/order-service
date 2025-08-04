package app

import (
	httpapp "github.com/zhavkk/order-service/internal/app/http"
	"github.com/zhavkk/order-service/internal/config"
	"github.com/zhavkk/order-service/internal/logger"
)

type App struct {
	httpApp *httpapp.HTTPApp
}

func NewApp(cfg *config.Config) (*App, error) {
	logger.Log.Info("Initializing application", "env", cfg.Env)

	return nil, nil
}
