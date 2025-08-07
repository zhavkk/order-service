package service

import (
	"context"
	"encoding/json"
	"fmt"

	"time"

	"github.com/zhavkk/order-service/internal/dto"
	"github.com/zhavkk/order-service/internal/logger"
	"github.com/zhavkk/order-service/internal/models"
	"github.com/zhavkk/order-service/pkg/cache"
	"github.com/zhavkk/order-service/pkg/pgstorage"
)

type OrderRepository interface {
	GetOrderByID(ctx context.Context, orderID string) (*models.Order, error)
	CreateOrder(ctx context.Context, order *models.Order) error
}

type DeliveryRepository interface {
	GetDeliveryByOrderID(ctx context.Context, orderID string) (*models.Delivery, error)
	CreateDelivery(ctx context.Context, delivery *models.Delivery) error
}

type PaymentRepository interface {
	GetPaymentByOrderID(ctx context.Context, orderID string) (*models.Payment, error)
	CreatePayment(ctx context.Context, payment *models.Payment) error
}

type ItemsRepository interface {
	GetItemsByOrderID(ctx context.Context, orderID string) ([]*models.Item, error)
	AddItems(ctx context.Context, orderID string, items []*models.Item) error
}

type OrderService struct {
	orderRepo    OrderRepository
	deliveryRepo DeliveryRepository
	paymentRepo  PaymentRepository
	itemsRepo    ItemsRepository
	txManager    pgstorage.TxManagerInterface
	cache        cache.Cache
	cacheTTL     time.Duration
}

func NewOrderService(
	orderRepo OrderRepository,
	deliveryRepo DeliveryRepository,
	paymentRepo PaymentRepository,
	itemsRepo ItemsRepository,
	txManager pgstorage.TxManagerInterface,
	cache cache.Cache,
	cacheTTL time.Duration,
) *OrderService {
	return &OrderService{
		orderRepo:    orderRepo,
		deliveryRepo: deliveryRepo,
		paymentRepo:  paymentRepo,
		itemsRepo:    itemsRepo,
		txManager:    txManager,
		cache:        cache,
		cacheTTL:     cacheTTL,
	}
}

func (s *OrderService) GetOrderByID(ctx context.Context,
	req *dto.GetOrderByIDRequest,
) (*dto.GetOrderByIDResponse, error) {

	const op = "OrderService.GetOrderByID"

	logger.Log.Info(op, "Fetching order by ID : ", req.OrderID)

	order, err := s.orderRepo.GetOrderByID(ctx, req.OrderID)
	if err != nil {
		logger.Log.Error(op, "Failed to get order", err)
		return nil, err
	}

	return &dto.GetOrderByIDResponse{
		Order: order,
	}, nil
}

func (s *OrderService) ProcessMessage(ctx context.Context, message []byte) error {
	const op = "OrderService.ProcessMessage"

	logger.Log.Info(op, "Processing message from Kafka", nil)

	var order models.Order
	if err := json.Unmarshal(message, &order); err != nil {
		logger.Log.Error(op, "Failed to unmarshal order", err)
		return nil
	}

	if order.OrderUID == "" || len(order.Items) == 0 {
		logger.Log.Warn("Invalid order data", "order", order)
		return nil
	}

	return s.ProcessOrder(ctx, &dto.ProcessOrderRequest{
		Order: &order,
	})
}

func (s *OrderService) ProcessOrder(ctx context.Context, req *dto.ProcessOrderRequest) error {
	const op = "OrderService.ProcessOrder"

	logger.Log.Info(op, "Processing order with ID : ", req.Order.OrderUID)
	return s.txManager.RunSerializable(ctx, func(ctx context.Context) error {
		if err := s.orderRepo.CreateOrder(ctx, req.Order); err != nil {
			logger.Log.Error(op, "Failed to create order", err)
			return err
		}
		items := make([]*models.Item, len(req.Order.Items))
		for i := range req.Order.Items {
			items[i] = &req.Order.Items[i]
		}

		if err := s.itemsRepo.AddItems(ctx, req.Order.OrderUID, items); err != nil {
			logger.Log.Error(op, "Failed to add items to order", err)
			return err
		}

		if err := s.deliveryRepo.CreateDelivery(ctx, &req.Order.Delivery); err != nil {
			logger.Log.Error(op, "Failed to create delivery", err)
			return err
		}

		if err := s.paymentRepo.CreatePayment(ctx, &req.Order.Payment); err != nil {
			logger.Log.Error(op, "Failed to create payment", err)
			return err
		}

		logger.Log.Info(op, "Order processed successfully, order_id: ", req.Order.OrderUID)

		if err := s.cache.Set(ctx, req.Order.OrderUID, req.Order, s.cacheTTL); err != nil {
			logger.Log.Error(op, "Failed to cache order", err)
			return err
		}

		return nil

	})
}

func (r *OrderService) GetByID(
	ctx context.Context,
	req *dto.GetOrderByIDRequest,
) (*dto.GetOrderByIDResponse, error) {
	const op = "OrderService.GetByID"
	logger.Log.Info(op, "Fetching order by ID: ", req.OrderID)

	cacheKey := fmt.Sprintf("order:%s", req.OrderID)
	var cachedOrder models.Order
	err := r.cache.Get(ctx, cacheKey, cachedOrder)

	if err == nil {
		logger.Log.Info(op, "Order found in cache with order_id: ", req.OrderID)
		return &dto.GetOrderByIDResponse{
			Order: &cachedOrder,
		}, nil
	}

	logger.Log.Warn(op, "Cache miss for order with order_id : ", req.OrderID)
	var order *models.Order
	order, err = r.orderRepo.GetOrderByID(ctx, req.OrderID)
	if err != nil {
		logger.Log.Error(op, "Failed to get order from repository", err)
		return nil, err
	}
	logger.Log.Info(op, "Order fetched from repository with order_id: ", req.OrderID)
	return &dto.GetOrderByIDResponse{
		Order: order,
	}, nil

}
