package app

import (
	httpapp "github.com/zhavkk/L0-test-service/internal/app/http"
	"github.com/zhavkk/L0-test-service/internal/config"
	"github.com/zhavkk/L0-test-service/internal/logger"
)

type App struct {
	httpApp *httpapp.HTTPApp
}

func NewApp(cfg *config.Config) (*App, error) {
	logger.Log.Info("Initializing application", "env", cfg.Env)

}
