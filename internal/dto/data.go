package dto

import "time"

type GetOrderByIDRequest struct {
	OrderID string `json:"order_id" validate:"required"`
}

type GetOrderByIDResponse struct {
	Order OrderResponse `json:"order"`
}

type ProcessOrderRequest struct {
	Order OrderRequest `json:"order" validate:"required"`
}

type OrderRequest struct {
	OrderUID          string      `json:"order_uid" validate:"required"`
	TrackNumber       string      `json:"track_number" validate:"required"`
	Entry             string      `json:"entry" validate:"required"`
	Delivery          DeliveryDTO `json:"delivery" validate:"required,dive"`
	Payment           PaymentDTO  `json:"payment" validate:"required,dive"`
	Items             []ItemDTO   `json:"items" validate:"required,min=1,dive"`
	Locale            string      `json:"locale" validate:"required"`
	InternalSignature string      `json:"internal_signature"`
	CustomerID        string      `json:"customer_id" validate:"required"`
	DeliveryService   string      `json:"delivery_service" validate:"required"`
	ShardKey          string      `json:"shardkey" validate:"required"`
	SmID              int         `json:"sm_id" validate:"gte=0"`
	DateCreated       time.Time   `json:"date_created" validate:"required"`
	OofShard          string      `json:"oof_shard" validate:"required"`
}

type OrderResponse struct {
	OrderUID          string      `json:"order_uid"`
	TrackNumber       string      `json:"track_number"`
	Entry             string      `json:"entry"`
	Delivery          DeliveryDTO `json:"delivery"`
	Payment           PaymentDTO  `json:"payment"`
	Items             []ItemDTO   `json:"items"`
	Locale            string      `json:"locale"`
	InternalSignature string      `json:"internal_signature"`
	CustomerID        string      `json:"customer_id"`
	DeliveryService   string      `json:"delivery_service"`
	ShardKey          string      `json:"shardkey"`
	SmID              int         `json:"sm_id"`
	DateCreated       time.Time   `json:"date_created"`
	OofShard          string      `json:"oof_shard"`
}

type DeliveryDTO struct {
	Name    string `json:"name" validate:"required"`
	Phone   string `json:"phone" validate:"required"`
	Zip     string `json:"zip" validate:"required"`
	City    string `json:"city" validate:"required"`
	Address string `json:"address" validate:"required"`
	Region  string `json:"region" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
}

type PaymentDTO struct {
	Transaction  string `json:"transaction" validate:"required"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency" validate:"required,oneof=USD EUR RUB"`
	Provider     string `json:"provider" validate:"required"`
	Amount       int    `json:"amount" validate:"gte=0"`
	PaymentDt    int64  `json:"payment_dt" validate:"gt=0"`
	Bank         string `json:"bank" validate:"required"`
	DeliveryCost int    `json:"delivery_cost" validate:"gte=0"`
	GoodsTotal   int    `json:"goods_total" validate:"gte=0"`
	CustomFee    int    `json:"custom_fee" validate:"gte=0"`
}

type ItemDTO struct {
	ChrtID      int64  `json:"chrt_id" validate:"required"`
	TrackNumber string `json:"track_number" validate:"required"`
	Price       int    `json:"price" validate:"gte=0"`
	Rid         string `json:"rid" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Sale        int    `json:"sale" validate:"gte=0"`
	Size        string `json:"size" validate:"required"`
	TotalPrice  int    `json:"total_price" validate:"gte=0"`
	NmId        int64  `json:"nm_id" validate:"required"`
	Brand       string `json:"brand" validate:"required"`
	Status      int    `json:"status" validate:"required"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}
