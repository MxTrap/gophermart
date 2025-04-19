package withdrawal

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/MxTrap/gophermart/internal/gophermart/controller/http/middlewares"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWithdrawalHandler_GetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockwithdrawer(ctrl)
	h := &withdrawalHandler{svc: mockService}

	userID := int64(123)
	processedAt := time.Now().Truncate(time.Second)
	withdrawals := []entity.Withdrawal{
		{
			Order:       "12345",
			Sum:         50.25,
			ProcessedAt: processedAt,
		},
	}

	t.Run("successful get all withdrawals", func(t *testing.T) {
		mockService.EXPECT().
			GetAll(gomock.Any(), userID).
			Return(withdrawals, nil)

		req := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetAll(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response []withdrawalDTO
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 1)
		assert.Equal(t, withdrawals[0].Order, response[0].Order)
		assert.Equal(t, withdrawals[0].Sum, response[0].Sum)
		assert.Equal(t, withdrawals[0].ProcessedAt.Format(time.RFC3339), response[0].ProcessedAt)
	})

	t.Run("no withdrawals", func(t *testing.T) {
		mockService.EXPECT().
			GetAll(gomock.Any(), userID).
			Return(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetAll(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("unauthorized - no user ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
		rr := httptest.NewRecorder()

		h.GetAll(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockService.EXPECT().
			GetAll(gomock.Any(), userID).
			Return(nil, errors.New("database error"))

		req := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetAll(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
