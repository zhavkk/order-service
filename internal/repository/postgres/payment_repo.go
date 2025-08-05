package postgres

import "github.com/zhavkk/order-service/pkg/pgstorage"

type PaymentRepository struct {
	storage *pgstorage.Storage
}
