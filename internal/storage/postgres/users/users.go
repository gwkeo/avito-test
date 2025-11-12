package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Storage struct {
	db *pgx.Conn
}

func New(ctx context.Context, connectionString string) (*Storage, error) {
	db, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}
