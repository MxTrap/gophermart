package withdrawal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/MxTrap/gophermart/internal/gophermart/mocks"
	"github.com/MxTrap/gophermart/logger"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestWithdrawalService_Withdraw(t *testing.T) {
	ctx := context.Background()
	log := logger.NewLogger()
	userID := int64(1)
	withdrawal := entity.Withdrawal{
		Order: "12345674",
		Sum:   100.0,
	}
	validWithdrawal := entity.Withdrawal{
		Order:       "12345674",
		Sum:         100.0,
		ProcessedAt: time.Now().UTC(),
	}

	tests := []struct {
		name        string
		userID      int64
		withdrawal  entity.Withdrawal
		setupMock   func(withdrawer *mocks.MockBalanceWithdrawalRepository)
		expectedErr error
	}{
		{
			name:       "Success",
			userID:     userID,
			withdrawal: withdrawal,
			setupMock: func(withdrawer *mocks.MockBalanceWithdrawalRepository) {
				withdrawer.EXPECT().
					Withdraw(ctx, userID, gomock.Any()).
					DoAndReturn(func(ctx context.Context, uid int64, w entity.Withdrawal) error {
						assert.Equal(t, validWithdrawal.Order, w.Order)
						assert.Equal(t, validWithdrawal.Sum, w.Sum)
						assert.WithinDuration(t, validWithdrawal.ProcessedAt, w.ProcessedAt, time.Second)
						return nil
					})
			},
			expectedErr: nil,
		},
		{
			name:       "Invalid order number",
			userID:     userID,
			withdrawal: entity.Withdrawal{Order: "", Sum: 100.0},
			setupMock: func(withdrawer *mocks.MockBalanceWithdrawalRepository) {

			},
			expectedErr: common.ErrInvalidOrderNumber,
		},
		{
			name:       "Withdraw error",
			userID:     userID,
			withdrawal: withdrawal,
			setupMock: func(withdrawer *mocks.MockBalanceWithdrawalRepository) {
				withdrawer.EXPECT().
					Withdraw(ctx, userID, gomock.Any()).
					Return(errors.New("insufficient balance"))
			},
			expectedErr: errors.New("insufficient balance"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			withdrawer := mocks.NewMockBalanceWithdrawalRepository(ctrl)
			getter := mocks.NewMockWithdrawalRepository(ctrl)

			tt.setupMock(withdrawer)

			s := NewWithdrawalService(log, withdrawer, getter)

			err := s.Withdraw(ctx, tt.userID, tt.withdrawal)

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected an error")
				assert.EqualError(t, err, tt.expectedErr.Error(), "error message mismatch")
			} else {
				assert.NoError(t, err, "unexpected error")
			}
		})
	}
}

func TestWithdrawalService_GetAll(t *testing.T) {
	ctx := context.Background()
	log := logger.NewLogger()
	userID := int64(1)

	tests := []struct {
		name                string
		setupMock           func(getter *mocks.MockWithdrawalRepository)
		expectedWithdrawals []entity.Withdrawal
		expectedErr         error
	}{
		{
			name: "Success",
			setupMock: func(getter *mocks.MockWithdrawalRepository) {
				getter.EXPECT().
					GetAll(ctx, userID).
					Return([]entity.Withdrawal{
						{Order: "12345", Sum: 100.0, ProcessedAt: time.Now().UTC()},
						{Order: "67890", Sum: 50.0, ProcessedAt: time.Now().UTC()},
					}, nil)
			},
			expectedWithdrawals: []entity.Withdrawal{
				{Order: "12345", Sum: 100.0, ProcessedAt: time.Now().UTC()},
				{Order: "67890", Sum: 50.0, ProcessedAt: time.Now().UTC()},
			},
			expectedErr: nil,
		},
		{
			name: "Repository error",
			setupMock: func(getter *mocks.MockWithdrawalRepository) {
				getter.EXPECT().
					GetAll(ctx, userID).
					Return(nil, errors.New("db error"))
			},
			expectedWithdrawals: nil,
			expectedErr:         errors.New("db error"),
		},
		{
			name: "Empty result",
			setupMock: func(getter *mocks.MockWithdrawalRepository) {
				getter.EXPECT().
					GetAll(ctx, userID).
					Return([]entity.Withdrawal{}, nil)
			},
			expectedWithdrawals: []entity.Withdrawal{},
			expectedErr:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			withdrawer := mocks.NewMockBalanceWithdrawalRepository(ctrl)
			getter := mocks.NewMockWithdrawalRepository(ctrl)

			tt.setupMock(getter)

			s := NewWithdrawalService(log, withdrawer, getter)

			withdrawals, err := s.GetAll(ctx, userID)

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected an error")
				assert.EqualError(t, err, tt.expectedErr.Error(), "error message mismatch")
				assert.Equal(t, tt.expectedWithdrawals, withdrawals, "withdrawals mismatch")
			} else {
				assert.NoError(t, err, "unexpected error")
				assert.Len(t, withdrawals, len(tt.expectedWithdrawals), "withdrawals length mismatch")
				for i, w := range withdrawals {
					assert.Equal(t, tt.expectedWithdrawals[i].Order, w.Order)
					assert.Equal(t, tt.expectedWithdrawals[i].Sum, w.Sum)
					assert.WithinDuration(t, tt.expectedWithdrawals[i].ProcessedAt, w.ProcessedAt, time.Second)
				}
			}
		})
	}
}
