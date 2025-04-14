package balance

import (
	"context"

	"github.com/MxTrap/gophermart/internal/gophermart/entity"
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

func (*BalanceRepository) Increase(ctx context.Context, tx *pgx.Tx, userID int64, sum float32) error {
	_, err := (*tx).Exec(ctx, increaseBalanceStmt, sum, userID)
	return err
}

func (*BalanceRepository) Withdraw(ctx context.Context, tx *pgx.Tx, userID int64, sum float32) error {
	_, err := (*tx).Exec(ctx, withdrawalStmt, sum, userID)
	return err
}

func (r *BalanceRepository) Get(ctx context.Context, userID int64) (entity.Balance, error) {
	row, err := r.db.Query(ctx, selectStmt, userID)

	if err != nil {
		return entity.Balance{}, err
	}

	return pgx.CollectOneRow(row, pgx.RowToStructByName[entity.Balance])
}
