package service

import (
	"context"

	"github.com/zhavkk/order-service/internal/models"
)

type OrderRepository interface {
	GetOrderByID(ctx context.Context, orderID string) (*models.Order, error)
	CreateOrder(ctx context.Context, order *models.Order) error
}

type OrderService struct {
	orderRepo OrderRepository
}

func NewOrderService(
	orderRepo OrderRepository,
) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
	}
}
