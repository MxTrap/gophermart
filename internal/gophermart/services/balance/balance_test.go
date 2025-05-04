package balance

import (
	"context"
	"errors"
	"testing"

	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/MxTrap/gophermart/internal/gophermart/mocks"
	"github.com/MxTrap/gophermart/logger"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestBalanceService_Get(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)
	log := logger.NewLogger() // Заглушка для логгера

	tests := []struct {
		name            string
		setupMock       func(repo *mocks.MockBalanceRepository)
		expectedBalance entity.Balance
		expectedErr     error
	}{
		{
			name: "Test success get",
			setupMock: func(repo *mocks.MockBalanceRepository) {
				repo.EXPECT().
					Get(ctx, userID).
					Return(entity.Balance{Balance: 100.0, Withdrawn: 20.0}, nil)
			},
			expectedBalance: entity.Balance{Balance: 100.0, Withdrawn: 20.0},
			expectedErr:     nil,
		},
		{
			name: "test with repository error",
			setupMock: func(repo *mocks.MockBalanceRepository) {
				repo.EXPECT().
					Get(ctx, userID).
					Return(entity.Balance{}, errors.New("database error"))
			},
			expectedBalance: entity.Balance{},
			expectedErr:     errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockBalanceRepository(ctrl)

			tt.setupMock(repo)

			s := NewBalanceService(log, repo)

			balance, err := s.Get(ctx, userID)

			assert.Equal(t, tt.expectedBalance, balance, "balance mismatch")
			if tt.expectedErr != nil {
				assert.Error(t, err, "expected an error")
				assert.EqualError(t, err, tt.expectedErr.Error(), "error message mismatch")
			} else {
				assert.NoError(t, err, "unexpected error")
			}
		})
	}
}
