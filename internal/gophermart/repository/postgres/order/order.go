package order

import (
	"context"

	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		db: pool,
	}
}

func (r *OrderRepository) SaveOrder(ctx context.Context, order entity.Order) error {
	_, err := r.db.Exec(
		ctx,
		insertStmt,
		order.UserID,
		order.Number,
		order.Status,
		order.Status,
		order.UploadedAt,
	)

	return err
}

func (r *OrderRepository) FindOrder(ctx context.Context, number string) (entity.Order, error) {
	var order entity.Order
	row, err := r.db.Query(ctx, selectByNumber, number)
	if err != nil {
		return order, nil
	}

	order, err = pgx.CollectOneRow(row, pgx.RowToStructByName[entity.Order])
	if err != nil {
		return order, nil
	}

	return order, nil
}

func (*OrderRepository) UpdateOrder(ctx context.Context, order entity.Order) error {
	return nil
}
