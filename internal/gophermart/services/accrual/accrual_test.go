package accrual

import (
	"encoding/json"
	"fmt"
	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/MxTrap/gophermart/logger"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccrualService_GetOrderAccrual(t *testing.T) {
	log := logger.NewLogger()
	orderNumber := "12345"
	baseURL := "/api/orders/"

	t.Run("successful response", func(t *testing.T) {
		accrualValue := float32(100.50)
		dto := accrualDto{
			Order:   orderNumber,
			Status:  "PROCESSED",
			Accrual: &accrualValue,
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, fmt.Sprintf("%s%s", baseURL, orderNumber), r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(dto)
		}))
		defer server.Close()

		svc := NewAccrualService(log, server.URL)
		order, err := svc.GetOrderAccrual(orderNumber)
		fmt.Println(order, err)
		assert.NoError(t, err)
		assert.Equal(t, entity.Order{
			Status:  "PROCESSED",
			Accrual: &accrualValue,
		}, order)
	})

	t.Run("no content response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		svc := NewAccrualService(log, server.URL)
		order, err := svc.GetOrderAccrual(orderNumber)
		assert.ErrorIs(t, err, common.ErrNonExistentOrder)
		assert.Equal(t, entity.Order{}, order)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		svc := NewAccrualService(log, server.URL)
		order, err := svc.GetOrderAccrual(orderNumber)
		assert.Error(t, err)
		assert.Equal(t, entity.Order{}, order)
	})

	t.Run("network error", func(t *testing.T) {
		svc := NewAccrualService(log, "http://invalid-url")
		order, err := svc.GetOrderAccrual(orderNumber)
		assert.Error(t, err)
		assert.Equal(t, entity.Order{}, order)
	})
}

func TestAccrualService_mapDtoToOrder(t *testing.T) {
	svc := &AccrualService{}
	accrualValue := float32(50.25)

	tests := []struct {
		name     string
		dto      accrualDto
		expected entity.Order
	}{
		{
			name: "registered status",
			dto: accrualDto{
				Order:   "123",
				Status:  "REGISTERED",
				Accrual: &accrualValue,
			},
			expected: entity.Order{
				Status:  entity.OrderNew,
				Accrual: &accrualValue,
			},
		},
		{
			name: "processed status",
			dto: accrualDto{
				Order:   "123",
				Status:  "PROCESSED",
				Accrual: &accrualValue,
			},
			expected: entity.Order{
				Status:  "PROCESSED",
				Accrual: &accrualValue,
			},
		},
		{
			name: "no accrual",
			dto: accrualDto{
				Order:  "123",
				Status: "INVALID",
			},
			expected: entity.Order{
				Status:  "INVALID",
				Accrual: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.mapDtoToOrder(tt.dto)
			assert.Equal(t, tt.expected, result)
		})
	}
}
