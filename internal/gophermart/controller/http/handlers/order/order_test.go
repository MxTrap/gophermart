package order

import (
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
	"time"
)

func TestOrderHandler_SaveOrderHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockorderService(ctrl)
	h := &orderHandler{service: mockService}

	userID := int64(123)
	orderNumber := "12345"

	t.Run("successful order save", func(t *testing.T) {
		mockService.EXPECT().
			SaveOrder(gomock.Any(), entity.Order{UserID: userID, Number: orderNumber}).
			Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(orderNumber))
		req.Header.Set("Content-Type", "text/plain")
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.SaveOrderHandler(rr, req)

		assert.Equal(t, http.StatusAccepted, rr.Code)
	})

	t.Run("invalid content type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(orderNumber))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.SaveOrderHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("unauthorized - no user ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(orderNumber))
		req.Header.Set("Content-Type", "text/plain")
		rr := httptest.NewRecorder()

		h.SaveOrderHandler(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("empty body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(""))
		req.Header.Set("Content-Type", "text/plain")
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.SaveOrderHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("order already exists", func(t *testing.T) {
		mockService.EXPECT().
			SaveOrder(gomock.Any(), entity.Order{UserID: userID, Number: orderNumber}).
			Return(common.ErrOrderAlreadyExist)

		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(orderNumber))
		req.Header.Set("Content-Type", "text/plain")
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.SaveOrderHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("order registered by another user", func(t *testing.T) {
		mockService.EXPECT().
			SaveOrder(gomock.Any(), entity.Order{UserID: userID, Number: orderNumber}).
			Return(common.ErrOrderRegisteredByAnother)

		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(orderNumber))
		req.Header.Set("Content-Type", "text/plain")
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.SaveOrderHandler(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("invalid order number", func(t *testing.T) {
		mockService.EXPECT().
			SaveOrder(gomock.Any(), entity.Order{UserID: userID, Number: orderNumber}).
			Return(common.ErrInvalidOrderNumber)

		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(orderNumber))
		req.Header.Set("Content-Type", "text/plain")
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.SaveOrderHandler(rr, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockService.EXPECT().
			SaveOrder(gomock.Any(), entity.Order{UserID: userID, Number: orderNumber}).
			Return(errors.New("database error"))

		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(orderNumber))
		req.Header.Set("Content-Type", "text/plain")
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.SaveOrderHandler(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestOrderHandler_mapOrderToDTO(t *testing.T) {
	h := &orderHandler{}
	uploadedAt := time.Now().Truncate(time.Second)
	accrual := float32(100.50)
	order := entity.Order{
		Number:     "12345",
		Status:     entity.OrderProcessed,
		Accrual:    &accrual,
		UploadedAt: uploadedAt,
	}

	result := h.mapOrderToDTO(order)

	expected := orderDTO{
		Number:     order.Number,
		Status:     order.Status,
		Accrual:    order.Accrual,
		UploadedAt: uploadedAt.Format(time.RFC3339),
	}
	assert.Equal(t, expected, result)
}

func TestOrderHandler_GetAllHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockorderService(ctrl)
	h := &orderHandler{service: mockService}

	userID := int64(123)
	uploadedAt := time.Now().Truncate(time.Second)
	accrual := float32(100.50)
	orders := []entity.Order{
		{
			Number:     "12345",
			Status:     entity.OrderProcessed,
			Accrual:    &accrual,
			UploadedAt: uploadedAt,
		},
	}

	t.Run("successful get all orders", func(t *testing.T) {
		mockService.EXPECT().
			GetAll(gomock.Any(), userID).
			Return(orders, nil)

		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetAllHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response []orderDTO
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 1)
		assert.Equal(t, orders[0].Number, response[0].Number)
		assert.Equal(t, orders[0].Status, response[0].Status)
		assert.Equal(t, orders[0].Accrual, response[0].Accrual)
		assert.Equal(t, orders[0].UploadedAt.Format(time.RFC3339), response[0].UploadedAt)
	})

	t.Run("no orders", func(t *testing.T) {
		mockService.EXPECT().
			GetAll(gomock.Any(), userID).
			Return(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetAllHandler(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("unauthorized - no user ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		rr := httptest.NewRecorder()

		h.GetAllHandler(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockService.EXPECT().
			GetAll(gomock.Any(), userID).
			Return(nil, errors.New("database error"))

		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		ctx := context.WithValue(req.Context(), middlewares.UserIDKey("UserID"), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		h.GetAllHandler(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
