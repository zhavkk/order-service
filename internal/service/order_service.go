package service

import (
	"context"
	"encoding/json"
	"fmt"

	"time"

	"github.com/go-playground/validator"
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
func (s *OrderService) ProcessMessage(ctx context.Context, message []byte) error {
	const op = "OrderService.ProcessMessage"
	logger.Log.Info(op, "Processing message from Kafka", nil)

	var in dto.OrderRequest
	if err := json.Unmarshal(message, &in); err != nil {
		logger.Log.Error(op, "Failed to unmarshal order", err)
		return nil
	}

	if err := validator.New().Struct(in); err != nil {
		logger.Log.Warn(op, "Invalid order DTO ", err)
		return nil
	}

	return s.ProcessOrder(ctx, &dto.ProcessOrderRequest{Order: in})
}

func (s *OrderService) ProcessOrder(ctx context.Context, req *dto.ProcessOrderRequest) error {
	const op = "OrderService.ProcessOrder"
	logger.Log.Info(op, "Processing order with ID:", req.Order.OrderUID)

	modelOrder := s.dtoToModel(req.Order)

	return s.txManager.RunSerializable(ctx, func(ctx context.Context) error {
		if err := s.orderRepo.CreateOrder(ctx, modelOrder); err != nil {
			logger.Log.Error(op, "Failed to create order", err)
			return err
		}

		items := make([]*models.Item, len(modelOrder.Items))
		for i := range modelOrder.Items {
			items[i] = &modelOrder.Items[i]
		}
		if err := s.itemsRepo.AddItems(ctx, modelOrder.OrderUID, items); err != nil {
			logger.Log.Error(op, "Failed to add items to order", err)
			return err
		}

		if err := s.deliveryRepo.CreateDelivery(ctx, &modelOrder.Delivery); err != nil {
			logger.Log.Error(op, "Failed to create delivery", err)
			return err
		}

		if err := s.paymentRepo.CreatePayment(ctx, &modelOrder.Payment); err != nil {
			logger.Log.Error(op, "Failed to create payment", err)
			return err
		}

		logger.Log.Info(op, "Order processed successfully, order_id:", modelOrder.OrderUID)

		cacheKey := fmt.Sprintf("order:%s", modelOrder.OrderUID)
		if err := s.cache.Set(ctx, cacheKey, modelOrder, s.cacheTTL); err != nil {
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
	logger.Log.Info(op, "Fetching order by ID:", req.OrderID)

	cacheKey := fmt.Sprintf("order:%s", req.OrderID)
	var cached models.Order
	if err := r.cache.Get(ctx, cacheKey, &cached); err == nil {
		logger.Log.Info(op, "Order found in cache with order_id: ", req.OrderID)
		return &dto.GetOrderByIDResponse{Order: r.modelToDTO(&cached)}, nil
	}

	logger.Log.Warn(op, "Cache miss order_id: ", req.OrderID)

	order, err := r.orderRepo.GetOrderByID(ctx, req.OrderID)
	if err != nil {
		logger.Log.Error(op, "Failed to get order from repository", err)
		return nil, err
	}

	_ = r.cache.Set(ctx, cacheKey, order, r.cacheTTL)

	return &dto.GetOrderByIDResponse{Order: r.modelToDTO(order)}, nil
}

func (s *OrderService) dtoToModel(in dto.OrderRequest) *models.Order {
	out := &models.Order{
		OrderUID:          in.OrderUID,
		TrackNumber:       in.TrackNumber,
		Entry:             in.Entry,
		Locale:            in.Locale,
		InternalSignature: in.InternalSignature,
		CustomerID:        in.CustomerID,
		DeliveryService:   in.DeliveryService,
		ShardKey:          in.ShardKey,
		SmID:              in.SmID,
		DateCreated:       in.DateCreated,
		OofShard:          in.OofShard,
		Delivery: models.Delivery{
			OrderID: in.OrderUID,
			Name:    in.Delivery.Name,
			Phone:   in.Delivery.Phone,
			Zip:     in.Delivery.Zip,
			City:    in.Delivery.City,
			Address: in.Delivery.Address,
			Region:  in.Delivery.Region,
			Email:   in.Delivery.Email,
		},
		Payment: models.Payment{
			Transaction:  in.Payment.Transaction,
			OrderID:      in.OrderUID,
			RequestID:    in.Payment.RequestID,
			Currency:     in.Payment.Currency,
			Provider:     in.Payment.Provider,
			Amount:       in.Payment.Amount,
			PaymentDt:    in.Payment.PaymentDt,
			Bank:         in.Payment.Bank,
			DeliveryCost: in.Payment.DeliveryCost,
			GoodsTotal:   in.Payment.GoodsTotal,
			CustomFee:    in.Payment.CustomFee,
		},
	}
	for _, it := range in.Items {
		out.Items = append(out.Items, models.Item{
			OrderID:     in.OrderUID,
			ChrtID:      it.ChrtID,
			TrackNumber: it.TrackNumber,
			Price:       it.Price,
			Rid:         it.Rid,
			Name:        it.Name,
			Sale:        it.Sale,
			Size:        it.Size,
			TotalPrice:  it.TotalPrice,
			NmId:        it.NmId,
			Brand:       it.Brand,
			Status:      it.Status,
		})
	}
	return out
}

func (s *OrderService) modelToDTO(in *models.Order) dto.OrderResponse {
	out := dto.OrderResponse{
		OrderUID:          in.OrderUID,
		TrackNumber:       in.TrackNumber,
		Entry:             in.Entry,
		Locale:            in.Locale,
		InternalSignature: in.InternalSignature,
		CustomerID:        in.CustomerID,
		DeliveryService:   in.DeliveryService,
		ShardKey:          in.ShardKey,
		SmID:              in.SmID,
		DateCreated:       in.DateCreated,
		OofShard:          in.OofShard,
		Delivery: dto.DeliveryDTO{
			Name:    in.Delivery.Name,
			Phone:   in.Delivery.Phone,
			Zip:     in.Delivery.Zip,
			City:    in.Delivery.City,
			Address: in.Delivery.Address,
			Region:  in.Delivery.Region,
			Email:   in.Delivery.Email,
		},
		Payment: dto.PaymentDTO{
			Transaction:  in.Payment.Transaction,
			RequestID:    in.Payment.RequestID,
			Currency:     in.Payment.Currency,
			Provider:     in.Payment.Provider,
			Amount:       in.Payment.Amount,
			PaymentDt:    in.Payment.PaymentDt,
			Bank:         in.Payment.Bank,
			DeliveryCost: in.Payment.DeliveryCost,
			GoodsTotal:   in.Payment.GoodsTotal,
			CustomFee:    in.Payment.CustomFee,
		},
	}
	for _, it := range in.Items {
		out.Items = append(out.Items, dto.ItemDTO{
			ChrtID:      it.ChrtID,
			TrackNumber: it.TrackNumber,
			Price:       it.Price,
			Rid:         it.Rid,
			Name:        it.Name,
			Sale:        it.Sale,
			Size:        it.Size,
			TotalPrice:  it.TotalPrice,
			NmId:        it.NmId,
			Brand:       it.Brand,
			Status:      it.Status,
		})
	}
	return out
}
