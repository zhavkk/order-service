package postgres

import "errors"

var (
	ErrNoTransaction = errors.New("no transaction found")
)
