package service

import (
	"context"

	"github.com/zhavkk/order-service/internal/models"
	"github.com/zhavkk/order-service/pkg/cache"
)

type OrderRepository interface {
	GetFullOrderByID(ctx context.Context, orderID string) (*models.Order, error)
	GetByID(ctx context.Context, orderID string) (*models.Order, error)
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
	AddItemsToOrder(ctx context.Context, orderID string, items []*models.Item) error
}

type OrderService struct {
	orderRepo    OrderRepository
	deliveryRepo DeliveryRepository
	paymentRepo  PaymentRepository
	itemsRepo    ItemsRepository

	cache cache.Cache
}

func NewOrderService(
	orderRepo OrderRepository,
	deliveryRepo DeliveryRepository,
	paymentRepo PaymentRepository,
	itemsRepo ItemsRepository,
	cache cache.Cache,
) *OrderService {
	return &OrderService{
		orderRepo:    orderRepo,
		deliveryRepo: deliveryRepo,
		paymentRepo:  paymentRepo,
		itemsRepo:    itemsRepo,
		cache:        cache,
	}
}
