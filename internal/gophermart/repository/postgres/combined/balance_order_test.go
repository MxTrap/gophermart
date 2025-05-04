package combined

import (
	"context"
	"errors"
	"testing"

	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/MxTrap/gophermart/internal/gophermart/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func TestOrderBalanceRepo_UpdateOrderBalance(t *testing.T) {
	ctx := context.Background()
	order := entity.Order{
		Number:  "12345",
		UserID:  1,
		Status:  "PROCESSED",
		Accrual: float32Ptr(100.0),
	}
	orderNoAccrual := entity.Order{
		Number: "67890",
		UserID: 1,
		Status: "PROCESSED",
	}

	tests := []struct {
		name        string
		order       entity.Order
		setupMocks  func(db *mocks.MockDBPool, orderRepo *mocks.MockOrderRepository, balanceRepo *mocks.MockBalanceRepository)
		expectedErr error
	}{
		{
			name:  "Success with accrual",
			order: order,
			setupMocks: func(db *mocks.MockDBPool, orderRepo *mocks.MockOrderRepository, balanceRepo *mocks.MockBalanceRepository) {
				tx := &mocks.MockTx{}
				db.EXPECT().
					BeginTx(ctx, pgx.TxOptions{}).
					Return(tx, nil)
				orderRepo.EXPECT().
					Update(ctx, tx, order).
					Return(nil)
				balanceRepo.EXPECT().
					Increase(ctx, tx, order.UserID, *order.Accrual).
					Return(nil)
				tx.CommitFn = func(ctx context.Context) error { return nil }
			},
			expectedErr: nil,
		},
		{
			name:  "Success without accrual",
			order: orderNoAccrual,
			setupMocks: func(db *mocks.MockDBPool, orderRepo *mocks.MockOrderRepository, balanceRepo *mocks.MockBalanceRepository) {
				tx := &mocks.MockTx{}
				db.EXPECT().
					BeginTx(ctx, pgx.TxOptions{}).
					Return(tx, nil)
				orderRepo.EXPECT().
					Update(ctx, tx, orderNoAccrual).
					Return(nil)
				tx.CommitFn = func(ctx context.Context) error { return nil }
			},
			expectedErr: nil,
		},
		{
			name:  "BeginTx error",
			order: order,
			setupMocks: func(db *mocks.MockDBPool, orderRepo *mocks.MockOrderRepository, balanceRepo *mocks.MockBalanceRepository) {
				db.EXPECT().
					BeginTx(ctx, pgx.TxOptions{}).
					Return(nil, errors.New("begin tx error"))
			},
			expectedErr: errors.New("begin tx error"),
		},
		{
			name:  "Update order error",
			order: order,
			setupMocks: func(db *mocks.MockDBPool, orderRepo *mocks.MockOrderRepository, balanceRepo *mocks.MockBalanceRepository) {
				tx := &mocks.MockTx{}
				db.EXPECT().
					BeginTx(ctx, pgx.TxOptions{}).
					Return(tx, nil)
				orderRepo.EXPECT().
					Update(ctx, tx, order).
					Return(errors.New("update error"))
				tx.RollbackFn = func(ctx context.Context) error { return nil }
			},
			expectedErr: errors.New("update error"),
		},
		{
			name:  "Increase balance error",
			order: order,
			setupMocks: func(db *mocks.MockDBPool, orderRepo *mocks.MockOrderRepository, balanceRepo *mocks.MockBalanceRepository) {
				tx := &mocks.MockTx{}
				db.EXPECT().
					BeginTx(ctx, pgx.TxOptions{}).
					Return(tx, nil)
				orderRepo.EXPECT().
					Update(ctx, tx, order).
					Return(nil)
				balanceRepo.EXPECT().
					Increase(ctx, tx, order.UserID, *order.Accrual).
					Return(errors.New("increase error"))
				tx.RollbackFn = func(ctx context.Context) error { return nil }
			},
			expectedErr: errors.New("increase error"),
		},
		{
			name:  "Commit error",
			order: order,
			setupMocks: func(db *mocks.MockDBPool, orderRepo *mocks.MockOrderRepository, balanceRepo *mocks.MockBalanceRepository) {
				tx := &mocks.MockTx{}
				db.EXPECT().
					BeginTx(ctx, pgx.TxOptions{}).
					Return(tx, nil)
				orderRepo.EXPECT().
					Update(ctx, tx, order).
					Return(nil)
				balanceRepo.EXPECT().
					Increase(ctx, tx, order.UserID, *order.Accrual).
					Return(nil)
				tx.CommitFn = func(ctx context.Context) error { return errors.New("commit error") }
			},
			expectedErr: errors.New("commit error"),
		},
		{
			name:  "Rollback error",
			order: order,
			setupMocks: func(db *mocks.MockDBPool, orderRepo *mocks.MockOrderRepository, balanceRepo *mocks.MockBalanceRepository) {
				tx := &mocks.MockTx{}
				db.EXPECT().
					BeginTx(ctx, pgx.TxOptions{}).
					Return(tx, nil)
				orderRepo.EXPECT().
					Update(ctx, tx, order).
					Return(errors.New("update error"))
				tx.RollbackFn = func(ctx context.Context) error { return errors.New("rollback error") }
			},
			expectedErr: errors.New("rollback error\nupdate error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mocks.NewMockDBPool(ctrl)
			orderRepo := mocks.NewMockOrderRepository(ctrl)
			balanceRepo := mocks.NewMockBalanceRepository(ctrl)

			tt.setupMocks(db, orderRepo, balanceRepo)

			r := NewOrderBalanceRepo(db, orderRepo, balanceRepo)

			err := r.UpdateOrderBalance(ctx, tt.order)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Вспомогательная функция для создания указателя на float32
func float32Ptr(f float32) *float32 {
	return &f
}
