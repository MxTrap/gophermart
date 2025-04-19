package order

import (
	"context"
	"errors"
	storage "github.com/MxTrap/gophermart/internal/gophermart/repository"

	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

const repoName = "postgres.OrderRepo."

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		db: pool,
	}
}

func (r *OrderRepository) Save(ctx context.Context, order entity.Order) error {
	_, err := r.db.Exec(
		ctx,
		insertStmt,
		order.UserID,
		order.Number,
		order.Status,
		order.Accrual,
		order.UploadedAt,
	)
	if err != nil {
		return storage.NewRepositoryError(repoName+"Save", err)
	}

	return nil
}

func (r *OrderRepository) Find(ctx context.Context, number string) (entity.Order, error) {
	const op = repoName + "Find"
	var order entity.Order
	row, err := r.db.Query(ctx, selectByNumber, number)
	if err != nil {
		return order, storage.NewRepositoryError(op, err)
	}
	defer row.Close()
	order, err = pgx.CollectOneRow(row, pgx.RowToStructByName[entity.Order])
	if err == nil || errors.Is(err, pgx.ErrNoRows) {
		return order, nil
	}

	return order, storage.NewRepositoryError(op, err)
}

func (r *OrderRepository) Update(ctx context.Context, tx pgx.Tx, order entity.Order) error {
	_, err := tx.Exec(ctx, updateStmt, order.Status, order.Accrual, order.Number)
	if err != nil {
		return storage.NewRepositoryError(repoName+"Update", err)
	}
	return nil
}

func (r *OrderRepository) GetAll(ctx context.Context, userID int64) ([]entity.Order, error) {
	const op = repoName + "GetAll"
	var orders []entity.Order
	rows, err := r.db.Query(ctx, selectAllStmt, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return orders, storage.NewRepositoryError(op, err)
	}
	defer rows.Close()

	orders, err = pgx.CollectRows(rows, pgx.RowToStructByName[entity.Order])
	if err != nil {
		return orders, storage.NewRepositoryError(op, err)
	}
	return orders, nil
}
