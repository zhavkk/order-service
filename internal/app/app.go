package app

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/redis/go-redis/v9"
	httpapp "github.com/zhavkk/order-service/internal/app/http"
	"github.com/zhavkk/order-service/internal/config"
	"github.com/zhavkk/order-service/internal/handler"
	"github.com/zhavkk/order-service/internal/logger"
	"github.com/zhavkk/order-service/internal/repository/postgres"
	"github.com/zhavkk/order-service/internal/service"
	rediscache "github.com/zhavkk/order-service/pkg/cache/redis"
	"github.com/zhavkk/order-service/pkg/pgstorage"
)

type App struct {
	httpApp *httpapp.HTTPApp
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	logger.Log.Info("Initializing application", "env", cfg.Env)

	txManager, err := pgstorage.NewTxManager(ctx, cfg)
	if err != nil {
		logger.Log.Error("Failed to create transaction manager", "error", err)
		return nil, err
	}

	postgresStorage, err := pgstorage.NewStorage(ctx, cfg)
	if err != nil {
		logger.Log.Error("Failed to connect to PostgreSQL", "error", err)
		return nil, err
	}
	redisClient := redis.NewClient(
		&redis.Options{
			Addr: cfg.Redis.Addr(),
		},
	)
	cache, err := rediscache.NewClient(redisClient, logger.Log)
	if err != nil {
		logger.Log.Error("Failed to create Redis cache client", "error", err)
		return nil, err
	}

	cacheTTL := cfg.Redis.TTL

	orderRepo := postgres.NewOrderRepository(postgresStorage)
	itemsRepo := postgres.NewItemRepository(postgresStorage)
	paymentRepo := postgres.NewPaymentRepository(postgresStorage)
	deliveryRepo := postgres.NewDeliveryRepository(postgresStorage)

	orderService := service.NewOrderService(orderRepo, deliveryRepo, paymentRepo, itemsRepo, txManager, cache, cacheTTL)

	handler := handler.NewHandler(orderService)
	router := httpapp.SetupRouter()

	httpApp := httpapp.New(cfg, router)

	handler.RegisterRoutes(router)
	addSystemRoutes(router)

	app := &App{
		httpApp: httpApp,
	}

	logger.Log.Info("Application initialized successfully", "env", cfg.Env)

	return app, nil
}

func (a *App) Run() error {
	logger.Log.Info("Starting application")
	return a.httpApp.Start()
}

func (a *App) Stop(ctx context.Context) error {
	logger.Log.Info("Stopping application")
	return a.httpApp.Stop(ctx)
}

func addSystemRoutes(router *chi.Mux) {
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"async-task-manager"}`))
	})

	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})
}
