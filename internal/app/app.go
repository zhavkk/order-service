package app

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/zhavkk/order-service/internal/app/consumer"
	httpapp "github.com/zhavkk/order-service/internal/app/http"
	"github.com/zhavkk/order-service/internal/config"
	"github.com/zhavkk/order-service/internal/handler"
	"github.com/zhavkk/order-service/internal/logger"
	"github.com/zhavkk/order-service/internal/repository/postgres"
	"github.com/zhavkk/order-service/internal/service"
	rediscache "github.com/zhavkk/order-service/pkg/cache/redis"
	kafkapkg "github.com/zhavkk/order-service/pkg/kafka/consumer"
	prometheusmetrics "github.com/zhavkk/order-service/pkg/metrics/prometheus"
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
			DB:   cfg.Redis.Db,
		},
	)
	cache, err := rediscache.NewClient(redisClient, logger.Log)
	if err != nil {
		logger.Log.Error("Failed to create Redis cache client", "error", err)
		return nil, err
	}

	cacheTTL := cfg.Redis.TTL

	retriesDB := cfg.Postgres.Retries
	backoffDB := cfg.Postgres.Backoff

	orderRepo := postgres.NewOrderRepository(postgresStorage, retriesDB, backoffDB)
	itemsRepo := postgres.NewItemRepository(postgresStorage, retriesDB, backoffDB)
	paymentRepo := postgres.NewPaymentRepository(postgresStorage, retriesDB, backoffDB)
	deliveryRepo := postgres.NewDeliveryRepository(postgresStorage, retriesDB, backoffDB)

	orderService := service.NewOrderService(orderRepo, deliveryRepo, paymentRepo, itemsRepo, txManager, cache, cacheTTL)

	go func() {
		warmUpCTX, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err := orderService.WarmUpCache(warmUpCTX); err != nil {
			logger.Log.Error("Failed to warm up cache", "error", err)
		}

		logger.Log.Info("Cache warmed up successfully")
	}()

	handler := handler.NewHandler(orderService)
	router := httpapp.SetupRouter()

	httpApp := httpapp.New(cfg, router)

	handler.RegisterRoutes(router)

	prometheusmetrics.Init()

	addSystemRoutes(router)

	app := &App{
		httpApp: httpApp,
	}

	saramaCfg, err := kafkapkg.NewSaramaConfig(cfg)
	if err != nil {
		logger.Log.Error("Failed to create Sarama config", "error", err)
		return nil, err
	}

	retriesKafka := cfg.Kafka.Retries
	backoffKafka := cfg.Kafka.Backoff

	kafkaConsumer, err := consumer.NewKafkaConsumer(
		cfg.Kafka.Brokers, cfg.Kafka.OrderTopic,
		func(msg []byte) error { return orderService.ProcessMessage(ctx, msg) },
		saramaCfg, cfg.Kafka.GroupID, retriesKafka, backoffKafka,
	)

	if err != nil {
		return nil, err
	}

	go func() {
		if err := kafkaConsumer.Consume(ctx); err != nil {
			logger.Log.Error("Kafka consumer stopped", "error", err)
		}
	}()

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

	router.Handle("/metrics", prometheusmetrics.Handler())
	router.Get("/swagger/*", httpSwagger.WrapHandler)

	fileServer := http.FileServer(http.Dir("./internal/web"))
	router.Handle("/*", fileServer)
}
