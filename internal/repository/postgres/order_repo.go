package postgres

import (
	"github.com/zhavkk/order-service/internal/models"
	"github.com/zhavkk/order-service/pkg/pgstorage"
)

type OrderRepository struct {
	storage *pgstorage.Storage
}

func NewOrderRepository(storage *pgstorage.Storage) *OrderRepository {
	return &OrderRepository{
		storage: storage,
	}
}

func (r *OrderRepository) GetOrderByID(orderID string) (*models.Order, error) {
	return nil, nil
}

func (r *OrderRepository) CreateOrder(order *models.Order) error {
	return nil
}
