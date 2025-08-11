package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/zhavkk/order-service/internal/logger"
	"github.com/zhavkk/order-service/internal/models"
	"github.com/zhavkk/order-service/pkg/pgstorage"
	"github.com/zhavkk/order-service/pkg/utils"
)

type OrderRepository struct {
	storage    *pgstorage.Storage
	retryCount int
	backoff    time.Duration
}

func NewOrderRepository(storage *pgstorage.Storage, retryCount int, backoff time.Duration) *OrderRepository {
	return &OrderRepository{
		storage:    storage,
		retryCount: retryCount,
		backoff:    backoff,
	}
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, orderID string) (*models.Order, error) {
	var fo models.Order
	orderQ := `
        SELECT order_uid, track_number, entry, locale, internal_signature, customer_id,
               delivery_service, shardkey, sm_id, date_created, oof_shard
        FROM orders
    	WHERE order_uid = $1
    `
	err := r.storage.GetPool().
		QueryRow(ctx, orderQ, orderID).
		Scan(
			&fo.OrderUID, &fo.TrackNumber, &fo.Entry,
			&fo.Locale, &fo.InternalSignature, &fo.CustomerID,
			&fo.DeliveryService, &fo.ShardKey, &fo.SmID,
			&fo.DateCreated, &fo.OofShard,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	deliveryQ := `
        SELECT delivery_id, order_uid, name, phone, zip, city, address, region, email
          FROM delivery
         WHERE order_uid = $1
    `
	if err := r.storage.GetPool().
		QueryRow(ctx, deliveryQ, orderID).
		Scan(
			&fo.Delivery.ID, &fo.Delivery.OrderID, &fo.Delivery.Name,
			&fo.Delivery.Phone, &fo.Delivery.Zip, &fo.Delivery.City,
			&fo.Delivery.Address, &fo.Delivery.Region, &fo.Delivery.Email,
		); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	paymentQ := `
        SELECT transaction, order_uid, request_id, currency, provider, amount,
               payment_dt, bank, delivery_cost, goods_total, custom_fee
          FROM payments
         WHERE order_uid = $1
    `
	if err := r.storage.GetPool().
		QueryRow(ctx, paymentQ, orderID).
		Scan(
			&fo.Payment.Transaction, &fo.Payment.OrderID, &fo.Payment.RequestID,
			&fo.Payment.Currency, &fo.Payment.Provider, &fo.Payment.Amount,
			&fo.Payment.PaymentDt, &fo.Payment.Bank, &fo.Payment.DeliveryCost,
			&fo.Payment.GoodsTotal, &fo.Payment.CustomFee,
		); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	itemsQ := `
        SELECT item_id, order_uid, chrt_id, track_number, price, rid, name,
               sale, size, total_price, nm_id, brand, status
          FROM items
         WHERE order_uid = $1
    `
	rows, err := r.storage.GetPool().Query(ctx, itemsQ, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var it models.Item
		if err := rows.Scan(
			&it.ID, &it.OrderID, &it.ChrtID, &it.TrackNumber, &it.Price, &it.Rid,
			&it.Name, &it.Sale, &it.Size, &it.TotalPrice, &it.NmId, &it.Brand, &it.Status,
		); err != nil {
			return nil, err
		}
		fo.Items = append(fo.Items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &fo, nil
}

func (r *OrderRepository) GetRecentOrders(ctx context.Context, limit int) ([]*models.Order, error) {
	const op = "OrderRepository.GetRecentOrders"
	orderQuery := `
        SELECT order_uid
        FROM orders
        ORDER BY date_created DESC
        LIMIT $1
    `

	rows, err := r.storage.GetPool().Query(ctx, orderQuery, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orderUIDs []string
	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			return nil, err
		}
		orderUIDs = append(orderUIDs, orderUID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var orders []*models.Order
	for _, uid := range orderUIDs {
		fullOrder, err := r.GetOrderByID(ctx, uid)
		if err != nil {
			logger.Log.Error(op, "Failed to get order by ID", err)
			continue
		}
		orders = append(orders, fullOrder)
	}

	return orders, nil
}
func (r *OrderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	return utils.RetryWithBackoff(func() error {

		query := `
	INSERT INTO orders (
        order_uid, track_number, entry, locale, internal_signature, customer_id,
        delivery_service, shardkey, sm_id, date_created, oof_shard
    ) VALUES (
        $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11
    )
	ON CONFLICT (order_uid) DO NOTHING	
	`

		tx, ok := pgstorage.GetTxFromContext(ctx)
		if !ok {
			return ErrNoTransaction
		}
		_, err := tx.Exec(ctx, query,
			order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerID,
			order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard,
		)
		return err
	}, r.retryCount, r.backoff)
}
