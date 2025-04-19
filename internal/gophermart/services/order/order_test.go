package order

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

func TestOrderService_SaveOrder(t *testing.T) {
	ctx := context.Background()
	log := logger.NewLogger()
	order := entity.Order{
		Number: "12345674",
		UserID: 1,
	}
	validOrder := entity.Order{
		Number:     "12345674",
		UserID:     1,
		Status:     entity.OrderNew,
		UploadedAt: time.Now().UTC(),
	}

	tests := []struct {
		name        string
		order       entity.Order
		setupMocks  func(repo *mocks.MockOrderRepository, storage *mocks.MockStorageService)
		expectedErr error
	}{
		{
			name:  "Success",
			order: order,
			setupMocks: func(repo *mocks.MockOrderRepository, storage *mocks.MockStorageService) {
				repo.EXPECT().
					Find(ctx, order.Number).
					Return(entity.Order{}, nil)
				repo.EXPECT().
					Save(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, o entity.Order) error {
						assert.Equal(t, validOrder.Number, o.Number)
						assert.Equal(t, validOrder.UserID, o.UserID)
						assert.Equal(t, validOrder.Status, o.Status)
						assert.WithinDuration(t, validOrder.UploadedAt, o.UploadedAt, time.Second)
						return nil
					})
				storage.EXPECT().
					Push(gomock.Any()).
					Do(func(o entity.Order) {
						assert.Equal(t, validOrder.Number, o.Number)
						assert.Equal(t, validOrder.UserID, o.UserID)
						assert.Equal(t, validOrder.Status, o.Status)
						assert.WithinDuration(t, validOrder.UploadedAt, o.UploadedAt, time.Second)
					})
			},
			expectedErr: nil,
		},
		{
			name:  "Invalid order number",
			order: entity.Order{Number: "", UserID: 1},
			setupMocks: func(repo *mocks.MockOrderRepository, storage *mocks.MockStorageService) {
				// Никаких вызовов, так как валидация не проходит
			},
			expectedErr: common.ErrInvalidOrderNumber,
		},
		{
			name:  "Order already exists",
			order: order,
			setupMocks: func(repo *mocks.MockOrderRepository, storage *mocks.MockStorageService) {
				repo.EXPECT().
					Find(ctx, order.Number).
					Return(entity.Order{Number: order.Number, UserID: order.UserID}, nil)
			},
			expectedErr: common.ErrOrderAlreadyExist,
		},
		{
			name:  "Order registered by another user",
			order: order,
			setupMocks: func(repo *mocks.MockOrderRepository, storage *mocks.MockStorageService) {
				repo.EXPECT().
					Find(ctx, order.Number).
					Return(entity.Order{Number: order.Number, UserID: 2}, nil)
			},
			expectedErr: common.ErrOrderRegisteredByAnother,
		},
		{
			name:  "Find error",
			order: order,
			setupMocks: func(repo *mocks.MockOrderRepository, storage *mocks.MockStorageService) {
				repo.EXPECT().
					Find(ctx, order.Number).
					Return(entity.Order{}, errors.New("db error"))
			},
			expectedErr: common.ErrInternalError,
		},
		{
			name:  "Save error",
			order: order,
			setupMocks: func(repo *mocks.MockOrderRepository, storage *mocks.MockStorageService) {
				repo.EXPECT().
					Find(ctx, order.Number).
					Return(entity.Order{}, nil)
				repo.EXPECT().
					Save(ctx, gomock.Any()).
					Return(errors.New("save error"))
			},
			expectedErr: common.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockOrderRepository(ctrl)
			storage := mocks.NewMockStorageService(ctrl)

			tt.setupMocks(repo, storage)

			s := NewOrderService(log, storage, repo)

			err := s.SaveOrder(ctx, tt.order)

			if tt.expectedErr != nil {
				assert.Error(t, err, "expected an error")
				assert.EqualError(t, err, tt.expectedErr.Error(), "error message mismatch")
			} else {
				assert.NoError(t, err, "unexpected error")
			}
		})
	}
}

func TestOrderService_GetAll(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)
	log := logger.NewLogger()

	tests := []struct {
		name           string
		setupMock      func(repo *mocks.MockOrderRepository)
		expectedOrders []entity.Order
		expectedErr    error
	}{
		{
			name: "Success",
			setupMock: func(repo *mocks.MockOrderRepository) {
				repo.EXPECT().
					GetAll(ctx, userID).
					Return([]entity.Order{
						{Number: "12345", UserID: userID, Status: entity.OrderNew},
						{Number: "67890", UserID: userID, Status: "PROCESSED"},
					}, nil)
			},
			expectedOrders: []entity.Order{
				{Number: "12345", UserID: userID, Status: entity.OrderNew},
				{Number: "67890", UserID: userID, Status: "PROCESSED"},
			},
			expectedErr: nil,
		},
		{
			name: "Repository error",
			setupMock: func(repo *mocks.MockOrderRepository) {
				repo.EXPECT().
					GetAll(ctx, userID).
					Return(nil, errors.New("db error"))
			},
			expectedOrders: nil,
			expectedErr:    errors.New("db error"),
		},
		{
			name: "Empty result",
			setupMock: func(repo *mocks.MockOrderRepository) {
				repo.EXPECT().
					GetAll(ctx, userID).
					Return([]entity.Order{}, nil)
			},
			expectedOrders: []entity.Order{},
			expectedErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockOrderRepository(ctrl)
			storage := mocks.NewMockStorageService(ctrl)

			tt.setupMock(repo)

			s := NewOrderService(log, storage, repo)

			orders, err := s.GetAll(ctx, userID)

			assert.Equal(t, tt.expectedOrders, orders, "orders mismatch")
			if tt.expectedErr != nil {
				assert.Error(t, err, "expected an error")
				assert.EqualError(t, err, tt.expectedErr.Error(), "error message mismatch")
			} else {
				assert.NoError(t, err, "unexpected error")
			}
		})
	}
}
