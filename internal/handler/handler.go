package handler

import (
	"context"

	"github.com/zhavkk/order-service/internal/dto"
)

type OrderService interface {
	GetByID(ctx context.Context, req *dto.GetOrderByIDRequest) (*dto.GetOrderByIDResponse, error)
	ProcessMessage(ctx context.Context, message []byte) error
	ProcessOrder(ctx context.Context, req *dto.ProcessOrderRequest) error
}

type Handler struct {
	orderService OrderService
}

func NewHandler(orderService OrderService) *Handler {
	return &Handler{
		orderService: orderService,
	}
}

func (h *Handler) GetOrderByID(ctx context.Context,
	req *dto.GetOrderByIDRequest,
) (*dto.GetOrderByIDResponse, error) {

	return h.orderService.GetByID(ctx, req)
}
