package postgres

import (
	"context"

	"github.com/zhavkk/order-service/internal/logger"
	"github.com/zhavkk/order-service/internal/models"
	"github.com/zhavkk/order-service/pkg/pgstorage"
)

type ItemRepository struct {
	storage *pgstorage.Storage
}

func NewItemRepository(storage *pgstorage.Storage) *ItemRepository {
	return &ItemRepository{
		storage: storage,
	}
}

func (r *ItemRepository) GetItemsByOrderID(ctx context.Context, orderID string) ([]*models.Item, error) {
	query := `SELECT * FROM items WHERE order_id = $1`
	rows, err := r.storage.GetPool().Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.Item
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ChrtID,
			&item.TrackNumber, &item.Price, &item.Rid,
			&item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmId,
			&item.Brand, &item.Status,
		); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ItemRepository) AddItemsToOrder(ctx context.Context, orderID string, items []*models.Item) error {
	const op = "ItemRepository.AddItemsToOrder"
	query := `INSERT INTO items (order_id, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
	 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	tx, ok := pgstorage.GetTxFromContext(ctx)
	if !ok {
		logger.Log.Error(op, "No transaction found in context", nil)
		return ErrNoTransaction
	}

	for _, item := range items {
		_, err := tx.Exec(ctx, query,
			orderID,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmId,
			item.Brand,
			item.Status,
		)
		if err != nil {
			return err
		}
		logger.Log.Info(op, "Item added successfully, order_id: ", orderID, "item_id", item.ID)
	}

	logger.Log.Info(op, "All items added successfully to order, order_id: ", orderID)

	return nil
}
