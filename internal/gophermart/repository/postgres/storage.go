package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	Pool *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, conString string) (*Storage, error) {
	pool, err := pgxpool.New(ctx, conString)

	if err != nil {
		return nil, err
	}

	return &Storage{
		Pool: pool,
	}, nil
}

func (s Storage) Stop() {
	s.Pool.Close()
}
