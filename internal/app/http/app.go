package httpapp

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/zhavkk/L0-test-service/internal/config"
	"github.com/zhavkk/L0-test-service/internal/logger"
)

type HTTPApp struct {
	httpServer *http.Server
	port       string
}

func New(cfg *config.Config, handler http.Handler) *HTTPApp {
	return &HTTPApp{
		httpServer: &http.Server{
			Addr:    cfg.HTTP.Port,
			Handler: handler,
		},
		port: cfg.HTTP.Port,
	}
}

func (a *HTTPApp) Start() error {
	logger.Log.Info("Starting HTTP server", "port", a.port)
	return a.httpServer.ListenAndServe()
}

func (a *HTTPApp) Stop(ctx context.Context) error {
	logger.Log.Info("Stopping HTTP server", "port", a.port)
	return a.httpServer.Shutdown(ctx)
}

func SetupRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))
	// then CORS and etc
	return r
}
