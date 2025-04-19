package balance

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/controller/http/middlewares"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBalanceHandler_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBalanceSvc := NewMockbalanceService(ctrl)
	mockWithdrawalSvc := NewMockwithdrawalService(ctrl)
	h := &balanceHandler{balanceSvc: mockBalanceSvc, withdrawalSvc: mockWithdrawalSvc}

	userID := int64(123)
	balance := entity.Balance{
		Current:   100.50,
		Withdrawn: 50.25,
	}

	t.Run("successful balance retrieval", func(t *testing.T) {
		mockBalanceSvc.EXPECT().
			Get(gomock.Any(), userID).
			Return(balance, nil)

		req := httptest.NewRequest(http.MethodGet, "/balance", nil)
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetBalance(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response struct {
			Current   float32 `json:"Current"`
			Withdrawn float32 `json:"withdrawn"`
		}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, balance.Current, response.Current)
		assert.Equal(t, balance.Withdrawn, response.Withdrawn)
	})

	t.Run("unauthorized - no user ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/balance", nil)
		rr := httptest.NewRecorder()

		h.GetBalance(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockBalanceSvc.EXPECT().
			Get(gomock.Any(), userID).
			Return(entity.Balance{}, errors.New("database error"))

		req := httptest.NewRequest(http.MethodGet, "/balance", nil)
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetBalance(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestBalanceHandler_Withdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBalanceSvc := NewMockbalanceService(ctrl)
	mockWithdrawalSvc := NewMockwithdrawalService(ctrl)
	h := &balanceHandler{balanceSvc: mockBalanceSvc, withdrawalSvc: mockWithdrawalSvc}

	userID := int64(123)
	withdrawal := entity.Withdrawal{
		Order: "12345",
		Sum:   50.25,
	}
	body, _ := json.Marshal(withdrawal)

	t.Run("successful withdrawal", func(t *testing.T) {
		mockWithdrawalSvc.EXPECT().
			Withdraw(gomock.Any(), userID, withdrawal).
			Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/balance/withdraw", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Withdraw(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("unauthorized - no user ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/balance/withdraw", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		h.Withdraw(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/balance/withdraw", strings.NewReader("invalid json"))
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Withdraw(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("insufficient balance", func(t *testing.T) {
		mockWithdrawalSvc.EXPECT().
			Withdraw(gomock.Any(), userID, withdrawal).
			Return(common.ErrInsufficientBalance)

		req := httptest.NewRequest(http.MethodPost, "/balance/withdraw", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Withdraw(rr, req)

		assert.Equal(t, http.StatusPaymentRequired, rr.Code)
	})

	t.Run("invalid order number", func(t *testing.T) {
		mockWithdrawalSvc.EXPECT().
			Withdraw(gomock.Any(), userID, withdrawal).
			Return(common.ErrInvalidOrderNumber)

		req := httptest.NewRequest(http.MethodPost, "/balance/withdraw", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Withdraw(rr, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockWithdrawalSvc.EXPECT().
			Withdraw(gomock.Any(), userID, withdrawal).
			Return(errors.New("database error"))

		req := httptest.NewRequest(http.MethodPost, "/balance/withdraw", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.Withdraw(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
