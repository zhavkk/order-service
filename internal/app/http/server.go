package httpapp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/zhavkk/order-service/internal/config"
	"github.com/zhavkk/order-service/internal/logger"
	metricsmw "github.com/zhavkk/order-service/internal/middleware"
)

type HTTPApp struct {
	httpServer *http.Server
	port       int
}

func New(cfg *config.Config, handler http.Handler) *HTTPApp {
	return &HTTPApp{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.HTTP.Port),
			Handler:      handler,
			ReadTimeout:  cfg.HTTP.ReadTimeout,
			WriteTimeout: cfg.HTTP.WriteTimeout,
			IdleTimeout:  cfg.HTTP.IdleTimeout,
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

	r.Use(metricsmw.MetricsMiddleware)
	return r
}
