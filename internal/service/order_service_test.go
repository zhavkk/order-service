package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zhavkk/order-service/internal/dto"
	"github.com/zhavkk/order-service/internal/logger"
	"github.com/zhavkk/order-service/internal/models"
	"github.com/zhavkk/order-service/internal/repository/mocks"
	"github.com/zhavkk/order-service/pkg/cache"
)

func TestOrderService_ProcessOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderRepo := mocks.NewMockOrderRepository(ctrl)
	mockDeliveryRepo := mocks.NewMockDeliveryRepository(ctrl)
	mockPaymentRepo := mocks.NewMockPaymentRepository(ctrl)
	mockItemsRepo := mocks.NewMockItemsRepository(ctrl)
	mockTxManager := mocks.NewMockTxManagerInterface(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	logger.Init("local")
	orderService := NewOrderService(
		mockOrderRepo,
		mockDeliveryRepo,
		mockPaymentRepo,
		mockItemsRepo,
		mockTxManager,
		mockCache,
		5*time.Minute,
	)

	randomOrder := generateRandomOrder()

	orderRequest := &dto.ProcessOrderRequest{
		Order: dto.OrderRequest{
			OrderUID:    randomOrder.OrderUID,
			TrackNumber: randomOrder.TrackNumber,
			Entry:       randomOrder.Entry,
			Delivery: dto.DeliveryDTO{
				Name:    randomOrder.Delivery.Name,
				Phone:   randomOrder.Delivery.Phone,
				Zip:     randomOrder.Delivery.Zip,
				City:    randomOrder.Delivery.City,
				Address: randomOrder.Delivery.Address,
				Region:  randomOrder.Delivery.Region,
				Email:   randomOrder.Delivery.Email,
			},
			Payment: dto.PaymentDTO{
				Transaction:  randomOrder.Payment.Transaction,
				Currency:     randomOrder.Payment.Currency,
				Provider:     randomOrder.Payment.Provider,
				Amount:       randomOrder.Payment.Amount,
				PaymentDt:    randomOrder.Payment.PaymentDt,
				Bank:         randomOrder.Payment.Bank,
				DeliveryCost: randomOrder.Payment.DeliveryCost,
				GoodsTotal:   randomOrder.Payment.GoodsTotal,
				CustomFee:    randomOrder.Payment.CustomFee,
			},
			Items: []dto.ItemDTO{
				{
					ChrtID:      randomOrder.Items[0].ChrtID,
					TrackNumber: randomOrder.Items[0].TrackNumber,
					Price:       randomOrder.Items[0].Price,
					Rid:         randomOrder.Items[0].Rid,
					Name:        randomOrder.Items[0].Name,
					Sale:        randomOrder.Items[0].Sale,
					Size:        randomOrder.Items[0].Size,
					TotalPrice:  randomOrder.Items[0].TotalPrice,
					NmId:        randomOrder.Items[0].NmId,
					Brand:       randomOrder.Items[0].Brand,
					Status:      randomOrder.Items[0].Status,
				},
			},
			Locale:            randomOrder.Locale,
			InternalSignature: randomOrder.InternalSignature,
			CustomerID:        randomOrder.CustomerID,
			DeliveryService:   randomOrder.DeliveryService,
			ShardKey:          randomOrder.ShardKey,
			SmID:              randomOrder.SmID,
			DateCreated:       randomOrder.DateCreated,
			OofShard:          randomOrder.OofShard,
		},
	}

	mockTxManager.EXPECT().RunSerializable(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(ctx context.Context) error) error {
			return fn(ctx)
		},
	)

	mockOrderRepo.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).Return(nil)
	mockItemsRepo.EXPECT().AddItems(gomock.Any(), randomOrder.OrderUID, gomock.Any()).Return(nil)
	mockDeliveryRepo.EXPECT().CreateDelivery(gomock.Any(), gomock.Any()).Return(nil)
	mockPaymentRepo.EXPECT().CreatePayment(gomock.Any(), gomock.Any()).Return(nil)
	mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	err := orderService.ProcessOrder(context.Background(), orderRequest)

	assert.NoError(t, err)
}

func TestOrderService_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderRepo := mocks.NewMockOrderRepository(ctrl)
	// mockDeliveryRepo := mocks.NewMockDeliveryRepository(ctrl)

	// mockItemsRepo := mocks.NewMockItemsRepository(ctrl)
	// mockTxManager := mocks.NewMockTxManagerInterface(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	logger.Init("local")
	orderService := NewOrderService(
		mockOrderRepo,
		nil,
		nil,
		nil,
		nil,
		mockCache,
		5*time.Minute,
	)
	orderID := uuid.NewString()
	req := &dto.GetOrderByIDRequest{
		OrderID: orderID,
	}
	expectedOrder := &models.Order{
		OrderUID:    orderID,
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: models.Delivery{
			Name:    "Test Delivery",
			Phone:   "+1234567890",
			Zip:     "123456",
			City:    "Test City",
			Address: "Test Address",
			Region:  "Test Region",
			Email:   "test@gmail.com",
		},
		Payment: models.Payment{
			Transaction:  uuid.NewString(),
			Currency:     "RUB",
			Provider:     "Test Provider",
			Amount:       1000,
			PaymentDt:    time.Now().Unix(),
			Bank:         "Test Bank",
			DeliveryCost: 100,
			GoodsTotal:   900,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:      123456,
				TrackNumber: "WBILMTESTTRACK",
				Price:       1000,
				Rid:         "test-rid",
				Name:        "Test Item",
				Sale:        0,
				Size:        "M",
				TotalPrice:  1000,
				NmId:        2389212,
				Brand:       "Test Brand",
				Status:      1,
			},
		},
		Locale:            "ru",
		InternalSignature: "Test Signature",
		CustomerID:        "test_customer",
		DeliveryService:   "Test Delivery Service",
		ShardKey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}

	mockCache.EXPECT().Get(gomock.Any(), "order:"+orderID, gomock.Any()).Return(cache.ErrCacheMiss)
	mockOrderRepo.EXPECT().GetOrderByID(gomock.Any(), orderID).Return(expectedOrder, nil)
	mockCache.EXPECT().Set(gomock.Any(), "order:"+orderID, gomock.Any(), 5*time.Minute).Return(nil)

	result, err := orderService.GetByID(context.Background(), req)

	expectedItems := []dto.ItemDTO{
		{
			ChrtID:      expectedOrder.Items[0].ChrtID,
			TrackNumber: expectedOrder.Items[0].TrackNumber,
			Price:       expectedOrder.Items[0].Price,
			Rid:         expectedOrder.Items[0].Rid,
			Name:        expectedOrder.Items[0].Name,
			Sale:        expectedOrder.Items[0].Sale,
			Size:        expectedOrder.Items[0].Size,
			TotalPrice:  expectedOrder.Items[0].TotalPrice,
			NmId:        expectedOrder.Items[0].NmId,
			Brand:       expectedOrder.Items[0].Brand,
			Status:      expectedOrder.Items[0].Status,
		},
	}
	assert.NoError(t, err)
	assert.Equal(t, expectedOrder.OrderUID, result.Order.OrderUID)
	assert.Equal(t, expectedItems, result.Order.Items)
	assert.Equal(t, expectedOrder.Delivery.Name, result.Order.Delivery.Name)
	assert.Equal(t, expectedOrder.Payment.Transaction, result.Order.Payment.Transaction)
	assert.Equal(t, expectedOrder.Locale, result.Order.Locale)
	assert.Equal(t, expectedOrder.InternalSignature, result.Order.InternalSignature)
	assert.Equal(t, expectedOrder.CustomerID, result.Order.CustomerID)
	assert.Equal(t, expectedOrder.DeliveryService, result.Order.DeliveryService)
	assert.Equal(t, expectedOrder.ShardKey, result.Order.ShardKey)
	assert.Equal(t, expectedOrder.SmID, result.Order.SmID)
	assert.Equal(t, expectedOrder.DateCreated.Unix(), result.Order.DateCreated.Unix())
	assert.Equal(t, expectedOrder.OofShard, result.Order.OofShard)

}

func TestOrderService_WarmUpCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mocks.NewMockCache(ctrl)
	logger.Init("local")
	mockOrderRepo := mocks.NewMockOrderRepository(ctrl)
	orderService := NewOrderService(
		mockOrderRepo,
		nil,
		nil,
		nil,
		nil,
		mockCache,
		5*time.Minute,
	)

	order1 := models.Order{OrderUID: "order-1"}
	order2 := models.Order{OrderUID: "order-2"}
	order3 := models.Order{OrderUID: "order-3"}
	recentOrders := []*models.Order{&order1, &order2, &order3}

	mockOrderRepo.EXPECT().GetRecentOrders(gomock.Any(), 1000).Return(recentOrders, nil)
	mockCache.EXPECT().Set(gomock.Any(), "order:order-1", gomock.Any(), 5*time.Minute).Return(nil)
	mockCache.EXPECT().Set(gomock.Any(), "order:order-2", gomock.Any(), 5*time.Minute).Return(nil)
	mockCache.EXPECT().Set(gomock.Any(), "order:order-3", gomock.Any(), 5*time.Minute).Return(nil)

	err := orderService.WarmUpCache(context.Background())

	assert.NoError(t, err)
}
func generateRandomOrder() models.Order {
	return models.Order{
		OrderUID:    uuid.NewString(),
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: models.Delivery{
			Name:    "Test Delivery",
			Phone:   "+1234567890",
			Zip:     "123456",
			City:    "Test City",
			Address: "Test Address",
			Region:  "Test Region",
			Email:   "test@gmail.com",
		},
		Payment: models.Payment{
			Transaction:  uuid.NewString(),
			Currency:     "RUB",
			Provider:     "Test Provider",
			Amount:       1000,
			PaymentDt:    time.Now().Unix(),
			Bank:         "Test Bank",
			DeliveryCost: 100,
			GoodsTotal:   900,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:      123456,
				TrackNumber: "WBILMTESTTRACK",
				Price:       1000,
				Rid:         uuid.NewString(),
				Name:        "Test Item",
				Sale:        0,
				Size:        "M",
				TotalPrice:  1000,
				NmId:        2389212,
				Brand:       "Test Brand",
				Status:      1,
			},
		},
		Locale:            "ru",
		InternalSignature: "Test Signature",
		CustomerID:        "test_customer",
		DeliveryService:   "Test Delivery Service",
		ShardKey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}
}
