package combined

import (
	"context"

	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type withdrawn interface {
	Save(ctx context.Context, tx *pgx.Tx, userID int64, withdrawn entity.Withdrawal) error
}

type balance interface {
	Withdraw(ctx context.Context, tx *pgx.Tx, userID int64, sum float32) error
}

type BalanceWithdrawnRepo struct {
	db    *pgxpool.Pool
	bRepo balance
	wRepo withdrawn
}

func NewBalanceWithdrawnRepo(db *pgxpool.Pool, bRepo balance, wRepo withdrawn) *BalanceWithdrawnRepo {
	return &BalanceWithdrawnRepo{
		db:    db,
		bRepo: bRepo,
		wRepo: wRepo,
	}
}

func (r *BalanceWithdrawnRepo) Withdraw(ctx context.Context, userID int64, withdrawal entity.Withdrawal) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	err = r.bRepo.Withdraw(ctx, &tx, userID, withdrawal.Sum)
	if err != nil {
		tx.Rollback(ctx)
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23514" {
			return common.ErrInsufficientBalance
		}
		return err
	}

	err = r.wRepo.Save(ctx, &tx, userID, withdrawal)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
