package postgres

import (
	"context"
	"time"

	"github.com/zhavkk/order-service/internal/logger"
	"github.com/zhavkk/order-service/internal/models"
	"github.com/zhavkk/order-service/pkg/pgstorage"
	"github.com/zhavkk/order-service/pkg/utils"
)

type DeliveryRepository struct {
	storage    *pgstorage.Storage
	retryCount int
	backoff    time.Duration
}

func NewDeliveryRepository(storage *pgstorage.Storage, retryCount int, backoff time.Duration) *DeliveryRepository {
	return &DeliveryRepository{
		storage:    storage,
		retryCount: retryCount,
		backoff:    backoff,
	}
}

func (r *DeliveryRepository) GetDeliveryByOrderID(ctx context.Context, orderID string) (*models.Delivery, error) {
	query := `
        SELECT delivery_id, order_uid, name, phone, zip, city, address, region, email
        FROM delivery
        WHERE order_uid = $1
    `
	var delivery models.Delivery
	err := r.storage.GetPool().QueryRow(ctx, query, orderID).Scan(
		&delivery.ID,
		&delivery.OrderID,
		&delivery.Name,
		&delivery.Phone,
		&delivery.Zip,
		&delivery.City,
		&delivery.Address,
		&delivery.Region,
		&delivery.Email,
	)
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (r *DeliveryRepository) CreateDelivery(ctx context.Context, delivery *models.Delivery) error {
	const op = "DeliveryRepository.CreateDelivery"

	return utils.RetryWithBackoff(func() error {
		query := `INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
	 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

		tx, ok := pgstorage.GetTxFromContext(ctx)
		if !ok {
			logger.Log.Error(op, "No transaction found in context", nil)
			return ErrNoTransaction
		}

		_, err := tx.Exec(ctx, query, delivery.OrderID, delivery.Name, delivery.Phone, delivery.Zip,
			delivery.City, delivery.Address, delivery.Region, delivery.Email)
		if err != nil {
			logger.Log.Error(op, "Failed to create delivery", err)
			return err
		}

		logger.Log.Info(op, "Delivery created successfully, order_uid: ", delivery.OrderID)
		return nil
	}, r.retryCount, r.backoff)
}
