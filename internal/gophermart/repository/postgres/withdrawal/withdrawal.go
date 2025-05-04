package withdrawn

import (
	"context"
	storage "github.com/MxTrap/gophermart/internal/gophermart/repository"

	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WithdrawnRepo struct {
	db *pgxpool.Pool
}

const repoName = "postgres.WithdrawalRepo."

func NewWithdrawnRepo(db *pgxpool.Pool) *WithdrawnRepo {
	return &WithdrawnRepo{
		db: db,
	}
}

func (r *WithdrawnRepo) GetAll(ctx context.Context, userID int64) ([]entity.Withdrawal, error) {
	const op = repoName + "GetAll"
	rows, err := r.db.Query(ctx, selectStmt, userID)
	if err != nil {
		return nil, storage.NewRepositoryError(op, err)
	}
	defer rows.Close()

	withdrawals, err := pgx.CollectRows(rows, pgx.RowToStructByPos[entity.Withdrawal])
	if err != nil {
		return nil, storage.NewRepositoryError(op, err)
	}

	return withdrawals, nil
}

func (*WithdrawnRepo) Save(ctx context.Context, tx pgx.Tx, userID int64, withdrawn entity.Withdrawal) error {
	_, err := tx.Exec(ctx, insertStmt, userID, withdrawn.Order, withdrawn.Sum, withdrawn.ProcessedAt)
	if err != nil {
		return storage.NewRepositoryError(repoName+"Save", err)
	}
	return nil
}
