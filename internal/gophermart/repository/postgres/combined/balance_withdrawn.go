package combined

import (
	"context"
	"errors"

	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type withdrawn interface {
	Save(ctx context.Context, tx *pgx.Tx, userID int64, withdrawn entity.Withdrawal) error
}

type balance interface {
	Withdraw(ctx context.Context, tx *pgx.Tx, userID int64, sum float32) error
}

type BalanceWithdrawnRepo struct {
	db             *pgxpool.Pool
	balanceRepo    balance
	withdrawalRepo withdrawn
}

func NewBalanceWithdrawnRepo(db *pgxpool.Pool, bRepo balance, wRepo withdrawn) *BalanceWithdrawnRepo {
	return &BalanceWithdrawnRepo{
		db:             db,
		balanceRepo:    bRepo,
		withdrawalRepo: wRepo,
	}
}

func (r *BalanceWithdrawnRepo) Withdraw(ctx context.Context, userID int64, withdrawal entity.Withdrawal) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	err = r.balanceRepo.Withdraw(ctx, &tx, userID, withdrawal.Sum)

	if err != nil {
		err := tx.Rollback(ctx)
		if err != nil {
			return err
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23514" {
			return common.ErrInsufficientBalance
		}
		return err
	}

	err = r.withdrawalRepo.Save(ctx, &tx, userID, withdrawal)
	if err != nil {
		err := tx.Rollback(ctx)
		if err != nil {
			return err
		}
		return err
	}

	return tx.Commit(ctx)
}
