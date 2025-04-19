package storage

import (
	"testing"

	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/stretchr/testify/assert"
)

func TestStorage_Push(t *testing.T) {
	tests := []struct {
		name          string
		orders        []entity.Order
		expectedQueue []entity.Order
	}{
		{
			name:          "Push single order",
			orders:        []entity.Order{{UserID: 1, Number: "123"}},
			expectedQueue: []entity.Order{{UserID: 1, Number: "123"}},
		},
		{
			name: "Push multiple orders",
			orders: []entity.Order{
				{UserID: 1, Number: "123"},
				{UserID: 2, Number: "456"},
			},
			expectedQueue: []entity.Order{
				{UserID: 1, Number: "123"},
				{UserID: 2, Number: "456"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем новый Storage
			s := NewStorageService()

			// Добавляем заказы
			for _, order := range tt.orders {
				s.Push(order)
			}

			// Проверяем содержимое очереди
			assert.Equal(t, tt.expectedQueue, *s.queue, "queue content mismatch")
		})
	}
}

func TestStorage_Get(t *testing.T) {
	tests := []struct {
		name           string
		initialQueue   []entity.Order
		elemCount      int
		expectedResult []entity.Order
		expectedQueue  []entity.Order
	}{
		{
			name:           "Test empty queue",
			initialQueue:   []entity.Order{},
			elemCount:      1,
			expectedResult: nil,
			expectedQueue:  []entity.Order{},
		},
		{
			name: "Test get all elements (exact count)",
			initialQueue: []entity.Order{
				{UserID: 1, Number: "123"},
				{UserID: 2, Number: "456"},
			},
			elemCount: 2,
			expectedResult: []entity.Order{
				{UserID: 1, Number: "123"},
				{UserID: 2, Number: "456"},
			},
			expectedQueue: []entity.Order{},
		},
		{
			name: "Test get all elements (less than requested)",
			initialQueue: []entity.Order{
				{UserID: 1, Number: "123"},
			},
			elemCount: 2,
			expectedResult: []entity.Order{
				{UserID: 1, Number: "123"},
			},
			expectedQueue: []entity.Order{},
		},
		{
			name: "Test get partial elements",
			initialQueue: []entity.Order{
				{UserID: 1, Number: "123"},
				{UserID: 2, Number: "456"},
				{UserID: 3, Number: "789"},
			},
			elemCount: 2,
			expectedResult: []entity.Order{
				{UserID: 1, Number: "123"},
				{UserID: 2, Number: "456"},
			},
			expectedQueue: []entity.Order{
				{UserID: 3, Number: "789"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStorageService()
			*s.queue = tt.initialQueue

			result := s.Get(tt.elemCount)

			assert.Equal(t, tt.expectedResult, result, "result mismatch")
			assert.Equal(t, tt.expectedQueue, *s.queue, "queue content mismatch")
		})
	}
}
