package dto

import "github.com/zhavkk/order-service/internal/models"

type GetOrderByIDRequest struct {
	OrderID string `json:"order_id"`
}

type GetOrderByIDResponse struct {
	Order *models.Order `json:"order"`
}
