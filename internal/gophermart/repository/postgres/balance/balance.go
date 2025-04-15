package balance

import (
	"context"
	storage "github.com/MxTrap/gophermart/internal/gophermart/repository"

	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BalanceRepository struct {
	db *pgxpool.Pool
}

const repoName = "postgres.BalanceRepo."

func NewBalanceRepository(pool *pgxpool.Pool) *BalanceRepository {
	return &BalanceRepository{
		db: pool,
	}
}

func (*BalanceRepository) Increase(ctx context.Context, tx *pgx.Tx, userID int64, sum float32) error {
	_, err := (*tx).Exec(ctx, increaseBalanceStmt, sum, userID)
	if err != nil {
		return storage.NewRepositoryError(repoName+"Increase", err)
	}
	return nil
}

func (*BalanceRepository) Withdraw(ctx context.Context, tx *pgx.Tx, userID int64, sum float32) error {
	_, err := (*tx).Exec(ctx, withdrawalStmt, sum, userID)
	if err != nil {
		return storage.NewRepositoryError(repoName+"Withdraw", err)
	}
	return nil
}

func (r *BalanceRepository) Get(ctx context.Context, userID int64) (entity.Balance, error) {
	const op = repoName + "Get"
	var balance entity.Balance
	row, err := r.db.Query(ctx, selectStmt, userID)
	if err != nil {
		return balance, storage.NewRepositoryError(op, err)
	}
	defer row.Close()

	balance, err = pgx.CollectOneRow(row, pgx.RowToStructByName[entity.Balance])
	if err != nil {
		return balance, storage.NewRepositoryError(op, err)
	}

	return balance, nil
}
