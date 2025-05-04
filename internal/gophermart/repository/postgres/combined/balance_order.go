package combined

import (
	"context"
	"errors"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/jackc/pgx/v5"
)

type orderRepo interface {
	Update(ctx context.Context, tx pgx.Tx, order entity.Order) error
}

type balanceRepo interface {
	Increase(ctx context.Context, tx pgx.Tx, userID int64, accrual float32) error
}

type db interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type OrderBalanceRepo struct {
	db
	orderRepo
	balanceRepo
}

func NewOrderBalanceRepo(pool db, order orderRepo, balance balanceRepo) *OrderBalanceRepo {
	return &OrderBalanceRepo{
		db:          pool,
		orderRepo:   order,
		balanceRepo: balance,
	}
}

func (r *OrderBalanceRepo) UpdateOrderBalance(ctx context.Context, order entity.Order) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	err = r.orderRepo.Update(ctx, tx, order)
	if err != nil {
		rErr := tx.Rollback(ctx)
		if rErr != nil {
			err = errors.Join(rErr, err)
		}
		return err
	}

	if order.Accrual != nil {
		err = r.balanceRepo.Increase(ctx, tx, order.UserID, *order.Accrual)
		if err != nil {
			rErr := tx.Rollback(ctx)
			if rErr != nil {
				err = errors.Join(rErr, err)
			}
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
