// Package pgstorage отвечает за подключение к базе данных и предоставление пула соединений.
package pgstorage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zhavkk/order-service/internal/config"
)

type Storage struct {
	db *pgxpool.Pool
}

var (
	ErrFailedToConnectToDB = errors.New("failed to connect to db")
	ErrDBNotConnected      = errors.New("db is not connected")
)

func NewStorage(ctx context.Context, cfg *config.Config) (*Storage, error) {
	dsn := cfg.Postgres.DSN()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, ErrFailedToConnectToDB
	}

	return &Storage{db: pool}, nil
}

func (s *Storage) Close() error {
	if s.db == nil {
		return ErrDBNotConnected
	}

	s.db.Close()
	return nil
}

func (s *Storage) GetPool() *pgxpool.Pool {
	return s.db
}
