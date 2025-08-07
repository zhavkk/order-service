package postgres

import (
	"context"

	"github.com/zhavkk/order-service/internal/models"
	"github.com/zhavkk/order-service/pkg/pgstorage"
)

type PaymentRepository struct {
	storage *pgstorage.Storage
}

func NewPaymentRepository(storage *pgstorage.Storage) *PaymentRepository {
	return &PaymentRepository{
		storage: storage,
	}
}

func (r *PaymentRepository) CreatePayment(ctx context.Context, payment *models.Payment) error {

	query := `INSERT INTO payments (
        transaction, order_uid, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
    ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`

	tx, ok := pgstorage.GetTxFromContext(ctx)
	if !ok {
		return ErrNoTransaction
	}

	_, err := tx.Exec(ctx, query,
		payment.Transaction, payment.OrderID, payment.RequestID, payment.Currency, payment.Provider,
		payment.Amount, payment.PaymentDt, payment.Bank, payment.DeliveryCost, payment.GoodsTotal, payment.CustomFee,
	)
	return err
}

func (r *PaymentRepository) GetPaymentByTransaction(ctx context.Context, transaction string) (*models.Payment, error) {
	query := `SELECT transaction, order_uid, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
        FROM payments WHERE transaction = $1`
	row := r.storage.GetPool().QueryRow(ctx, query, transaction)
	var payment models.Payment
	err := row.Scan(
		&payment.Transaction, &payment.OrderID, &payment.RequestID, &payment.Currency, &payment.Provider,
		&payment.Amount, &payment.PaymentDt, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
	)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}
