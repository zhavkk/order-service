package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/zhavkk/order-service/docs"
	"github.com/zhavkk/order-service/internal/app"
	"github.com/zhavkk/order-service/internal/config"
	"github.com/zhavkk/order-service/internal/logger"
)

// @title Order Service API
// @version 1.0
// @description API для управления заказами.
// @termsOfService http://example.com/terms/

// @contact.name API Support
// @contact.url http://example.com/support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
func main() {
	cfg := config.MustLoad("config/config.yml")

	logger.Init(cfg.Env)

	logger.Log.Info("Order Service")

	logger.Log.Info("Redis address from config", "addr", cfg.Redis.Addr())
	logger.Log.Info("Postgres DSN", "dsn", cfg.Postgres.DSN())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	app, err := app.NewApp(ctx, cfg)
	if err != nil {
		logger.Log.Error("Failed to initialize application", "error", err)
		return
	}

	go func() {
		if err := app.Run(); err != nil {
			logger.Log.Error("Application run error", "error", err)
			cancel()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Received shutdown signal, stopping application")

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.Stop(ctx); err != nil {
		logger.Log.Error("Failed to stop application gracefully", "error", err)
	}

	logger.Log.Info("Application stopped gracefully")

}
