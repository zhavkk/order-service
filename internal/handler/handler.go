package handler

import (
	"context"

	"github.com/zhavkk/order-service/internal/dto"
)

type OrderService interface {
	GetByID(ctx context.Context, req *dto.GetOrderByIDRequest) (*dto.GetOrderByIDResponse, error)
}
