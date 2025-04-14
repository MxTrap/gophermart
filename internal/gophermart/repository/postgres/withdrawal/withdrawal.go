package withdrawn

import (
	"context"

	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WithdrawnRepo struct {
	db *pgxpool.Pool
}

func NewWithdrawnRepo(db *pgxpool.Pool) *WithdrawnRepo {
	return &WithdrawnRepo{
		db: db,
	}
}

func (r *WithdrawnRepo) GetAll(ctx context.Context, userId int64) ([]entity.Withdrawal, error) {
	rows, err := r.db.Query(ctx, selectStmt, userId)

	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByPos[entity.Withdrawal])
}

func (*WithdrawnRepo) Save(ctx context.Context, tx *pgx.Tx, userId int64, withdrawn entity.Withdrawal) error {
	_, err := (*tx).Query(ctx, insertStmt, userId, withdrawn.Order, withdrawn.Sum, withdrawn.ProcessedAt)
	return err
}
