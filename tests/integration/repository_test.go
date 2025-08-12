package integration

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/zhavkk/order-service/internal/config"
	"github.com/zhavkk/order-service/internal/logger"
	"github.com/zhavkk/order-service/internal/models"
	"github.com/zhavkk/order-service/internal/repository/postgres"
	"github.com/zhavkk/order-service/pkg/pgstorage"
)

type RepositorySuite struct {
	suite.Suite
	ctx          context.Context
	pgContainer  testcontainers.Container
	storage      *pgstorage.Storage
	txManager    *pgstorage.TxManager
	orderRepo    *postgres.OrderRepository
	deliveryRepo *postgres.DeliveryRepository
	paymentRepo  *postgres.PaymentRepository
	itemRepo     *postgres.ItemRepository
}

func (s *RepositorySuite) SetupSuite() {
	var err error
	ctx, pgContainer, storage, cfg := setupPostgresContainer(s.T())
	s.pgContainer = pgContainer
	s.ctx = ctx
	s.storage = storage
	logger.Init("local")
	s.txManager, err = pgstorage.NewTxManager(ctx, cfg)
	require.NoError(s.T(), err)

	s.orderRepo = postgres.NewOrderRepository(storage, cfg.Postgres.Retries, cfg.Postgres.Backoff)
	s.deliveryRepo = postgres.NewDeliveryRepository(storage, cfg.Postgres.Retries, cfg.Postgres.Backoff)
	s.paymentRepo = postgres.NewPaymentRepository(storage, cfg.Postgres.Retries, cfg.Postgres.Backoff)
	s.itemRepo = postgres.NewItemRepository(storage, cfg.Postgres.Retries, cfg.Postgres.Backoff)

	schemaBytes, err := os.ReadFile("../../migrations/20250803160056_init.sql")
	require.NoError(s.T(), err)
	schema := string(schemaBytes)
	schema = strings.Split(schema, "-- +goose Down")[0]
	schema = strings.ReplaceAll(schema, "-- +goose Up", "")
	schema = strings.ReplaceAll(schema, "-- +goose StatementBegin", "")
	schema = strings.ReplaceAll(schema, "-- +goose StatementEnd", "")

	_, err = s.storage.GetPool().Exec(ctx, schema)
	require.NoError(s.T(), err, "Failed to apply migration schema")
}

func (s *RepositorySuite) SetupTest() {
	_, err := s.storage.GetPool().Exec(s.ctx, "TRUNCATE TABLE orders, delivery, payments, items RESTART IDENTITY CASCADE")
	require.NoError(s.T(), err)
}

func (s *RepositorySuite) TearDownSuite() {
	err := s.storage.Close()
	require.NoError(s.T(), err)
	err = s.pgContainer.Terminate(s.ctx)
	require.NoError(s.T(), err)
}

func (s *RepositorySuite) TestCreateAndGetOrder() {
	order := generateTestOrder()

	err := s.txManager.RunSerializable(s.ctx, func(txCtx context.Context) error {
		if err := s.orderRepo.CreateOrder(txCtx, &order); err != nil {
			return err
		}
		order.Delivery.OrderID = order.OrderUID
		if err := s.deliveryRepo.CreateDelivery(txCtx, &order.Delivery); err != nil {
			return err
		}
		order.Payment.OrderID = order.OrderUID
		if err := s.paymentRepo.CreatePayment(txCtx, &order.Payment); err != nil {
			return err
		}
		if err := s.itemRepo.AddItems(txCtx, order.OrderUID, itemsToPointers(order.Items)); err != nil {
			return err
		}
		return nil
	})
	s.Require().NoError(err, "Failed to create full order in transaction")

	retrievedOrder, err := s.orderRepo.GetOrderByID(s.ctx, order.OrderUID)
	s.Require().NoError(err, "Failed to get order by ID")
	s.Require().NotNil(retrievedOrder)

	opts := []cmp.Option{
		cmpopts.EquateApproxTime(time.Second),
		cmpopts.IgnoreFields(models.Order{}, "DateCreated"),
		cmpopts.IgnoreFields(models.Delivery{}, "ID"),
		cmpopts.IgnoreFields(models.Item{}, "ID", "OrderID"),
	}

	if diff := cmp.Diff(&order, retrievedOrder, opts...); diff != "" {
		s.T().Errorf("retrieved order mismatch (-want +got):\n%s", diff)
	}
}

func (s *RepositorySuite) TestGetOrderByID_NotFound() {
	_, err := s.orderRepo.GetOrderByID(s.ctx, uuid.NewString())
	s.Require().Error(err)
	s.Assert().ErrorIs(err, postgres.ErrOrderNotFound)
}

func (s *RepositorySuite) TestGetRecentOrders() {
	for i := 0; i < 3; i++ {
		order := generateTestOrder()
		order.DateCreated = time.Now().Add(time.Duration(i) * time.Minute)
		err := s.txManager.RunSerializable(s.ctx, func(txCtx context.Context) error {
			return s.orderRepo.CreateOrder(txCtx, &order)
		})
		s.Require().NoError(err)
	}

	recentOrders, err := s.orderRepo.GetRecentOrders(s.ctx, 2)
	s.Require().NoError(err)
	s.Require().Len(recentOrders, 2)

	s.Assert().True(recentOrders[0].DateCreated.After(recentOrders[1].DateCreated))
}

func (s *RepositorySuite) TestIndividualComponentRepos() {
	order := generateTestOrder()

	err := s.txManager.RunSerializable(s.ctx, func(txCtx context.Context) error {
		return s.orderRepo.CreateOrder(txCtx, &order)
	})
	s.Require().NoError(err)

	order.Delivery.OrderID = order.OrderUID
	err = s.txManager.RunSerializable(s.ctx, func(txCtx context.Context) error {
		return s.deliveryRepo.CreateDelivery(txCtx, &order.Delivery)
	})
	s.Require().NoError(err)
	retrievedDelivery, err := s.deliveryRepo.GetDeliveryByOrderID(s.ctx, order.OrderUID)
	s.Require().NoError(err)
	s.Assert().Equal(order.Delivery.Name, retrievedDelivery.Name)

	items := itemsToPointers(order.Items)
	err = s.txManager.RunSerializable(s.ctx, func(txCtx context.Context) error {
		return s.itemRepo.AddItems(txCtx, order.OrderUID, items)
	})
	s.Require().NoError(err)
	retrievedItems, err := s.itemRepo.GetItemsByOrderID(s.ctx, order.OrderUID)
	s.Require().NoError(err)
	s.Require().Len(retrievedItems, 1)
	s.Assert().Equal(items[0].ChrtID, retrievedItems[0].ChrtID)
}

func TestStorageSuite(t *testing.T) {
	suite.Run(t, new(RepositorySuite))
}

func setupPostgresContainer(t *testing.T) (context.Context, testcontainers.Container, *pgstorage.Storage, *config.Config) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "order_service_test",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)
	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	cfg := &config.Config{
		Postgres: config.PostgresConfig{
			Host:     host,
			Port:     port.Port(),
			Username: "postgres",
			Password: "postgres",
			Database: "order_service_test",
			SSLMode:  "disable",
		},
	}
	cfg.Postgres.Retries = 3
	cfg.Postgres.Backoff = time.Millisecond * 200

	storage, err := pgstorage.NewStorage(ctx, cfg)
	require.NoError(t, err)

	return ctx, pgContainer, storage, cfg
}

func generateTestOrder() models.Order {
	orderUID := uuid.NewString()
	return models.Order{
		OrderUID:    orderUID,
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: models.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Haifa",
			Email:   "test@gmail.com",
		},
		Payment: models.Payment{
			Transaction:  orderUID,
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    time.Now().Unix(),
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmId:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		ShardKey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}
}

func itemsToPointers(items []models.Item) []*models.Item {
	pointers := make([]*models.Item, len(items))
	for i := range items {
		pointers[i] = &items[i]
	}
	return pointers
}
