package balance

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BalanceRepository struct {
	db *pgxpool.Pool
}

func NewBalanceRepository(pool *pgxpool.Pool) *BalanceRepository {

	return &BalanceRepository{
		db: pool,
	}
}

func (r *BalanceRepository) Increase(ctx context.Context, tx pgx.Tx, userId int64, sum float32) error {
	_, err := tx.Exec(ctx, increaseBalanceStmt, sum, userId)
	return err
}

func (r *BalanceRepository) Withdraw(ctx context.Context, tx pgx.Tx, userId int64, sum float32) error {
	_, err := tx.Exec(ctx, withdrawalStmt, sum, userId)
	return err
}
