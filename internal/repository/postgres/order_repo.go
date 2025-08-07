package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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

func (r *OrderRepository) GetFullOrderByID(ctx context.Context, orderID string) (*models.FullOrder, error) {
	var fo models.FullOrder
	orderQ := `
        SELECT order_uid, track_number, entry, locale, internal_signature, customer_id,
               delivery_service, shardkey, sm_id, date_created, oof_shard
        FROM orders
    	WHERE order_uid = $1
    `
	err := r.storage.GetPool().
		QueryRow(ctx, orderQ, orderID).
		Scan(
			&fo.Order.OrderUID, &fo.Order.TrackNumber, &fo.Order.Entry,
			&fo.Order.Locale, &fo.Order.InternalSignature, &fo.Order.CustomerID,
			&fo.Order.DeliveryService, &fo.Order.ShardKey, &fo.Order.SmID,
			&fo.Order.DateCreated, &fo.Order.OofShard,
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

func (r *OrderRepository) GetByID(orderID string) (*models.Order, error) {
	query := `SELECT * FROM orders WHERE order_uid = $1`
	row := r.storage.GetPool().QueryRow(context.Background(), query, orderID)
	var order models.Order
	err := row.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.DeliveryService,
		&order.ShardKey, &order.SmID, &order.DateCreated, &order.OofShard)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) CreateOrder(order *models.Order) error {
	query := `INSERT INTO orders (order_uid, track_number, entry, delivery_service, shardkey, sm_id, date_created, oof_shard)
	 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	tx, ok := pgstorage.GetTxFromContext(context.Background())
	if !ok {
		return ErrNoTransaction
	}
	_, err := tx.Exec(context.Background(), query,
		order.OrderUID, order.TrackNumber, order.Entry, order.DeliveryService, order.ShardKey,
		order.SmID, order.DateCreated, order.OofShard,
	)
	if err != nil {
		return err
	}

	return nil
}
