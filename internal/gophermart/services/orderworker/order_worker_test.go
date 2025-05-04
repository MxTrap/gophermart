package orderworker

import (
	"context"
	"errors"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/MxTrap/gophermart/logger"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestOrderWorkerService_isTerminalStatus(t *testing.T) {
	svc := &OrderWorkerService{}
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"invalid status", entity.OrderInvalid, true},
		{"processed status", entity.OrderProcessed, true},
		{"new status", entity.OrderNew, false},
		{"processing status", entity.OrderProcessing, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.isTerminalStatus(tt.status)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOrderWorkerService_generate(t *testing.T) {
	svc := &OrderWorkerService{}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	orders := []entity.Order{
		{Number: "123", UserID: 1, Status: entity.OrderNew},
		{Number: "456", UserID: 2, Status: entity.OrderNew},
	}

	inputCh := svc.generate(ctx, orders)

	var received []entity.Order
	done := make(chan struct{})
	go func() {
		for order := range inputCh {
			received = append(received, order)
		}
		close(done)
	}()

	select {
	case <-done:
		assert.Equal(t, orders, received)
	case <-ctx.Done():
		t.Fatal("timeout waiting for channel")
	}

	// Тест с отменой контекста
	ctx, cancel = context.WithCancel(context.Background())
	cancel()
	inputCh = svc.generate(ctx, orders)
	received = nil
	done = make(chan struct{})
	go func() {
		for order := range inputCh {
			received = append(received, order)
		}
		close(done)
	}()

	select {
	case <-done:
		assert.Empty(t, received)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for channel close")
	}
}

func TestOrderWorkerService_update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAccrualService := NewMockaccrualService(ctrl)
	svc := &OrderWorkerService{svc: mockAccrualService}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	inputCh := make(chan entity.Order, 1)
	order := entity.Order{Number: "123", UserID: 1, Status: entity.OrderNew}
	var accrualTestNum float32 = 100.0
	accrualOrder := entity.Order{Number: "123", Status: entity.OrderProcessed, Accrual: &accrualTestNum}

	t.Run("successful update", func(t *testing.T) {
		inputCh <- order
		close(inputCh)

		mockAccrualService.EXPECT().
			GetOrderAccrual(order.Number).
			Return(accrualOrder, nil)

		resultCh := svc.update(ctx, inputCh)
		results := collectResults(resultCh)
		assert.Len(t, results, 1)
		assert.Equal(t, entity.Order{Number: order.Number, UserID: order.UserID, Status: accrualOrder.Status, Accrual: accrualOrder.Accrual}, results[0].order)
		assert.NoError(t, results[0].err)
	})

	t.Run("same status error", func(t *testing.T) {
		inputCh = make(chan entity.Order, 1)
		sameStatusOrder := entity.Order{Number: "123", UserID: 1, Status: entity.OrderProcessed}
		inputCh <- sameStatusOrder
		close(inputCh)

		mockAccrualService.EXPECT().
			GetOrderAccrual(sameStatusOrder.Number).
			Return(entity.Order{Status: sameStatusOrder.Status}, nil)

		resultCh := svc.update(ctx, inputCh)
		results := collectResults(resultCh)
		assert.Len(t, results, 1)
		assert.Error(t, results[0].err)
		assert.Equal(t, "order has already been processed", results[0].err.Error())
	})

	t.Run("accrual service error", func(t *testing.T) {
		inputCh = make(chan entity.Order, 1)
		inputCh <- order
		close(inputCh)

		accrualErr := errors.New("accrual error")
		mockAccrualService.EXPECT().
			GetOrderAccrual(order.Number).
			Return(entity.Order{}, accrualErr)

		resultCh := svc.update(ctx, inputCh)
		results := collectResults(resultCh)
		assert.Len(t, results, 1)
		assert.ErrorIs(t, results[0].err, accrualErr)
	})
}

func TestOrderWorkerService_fanOut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAccrualService := NewMockaccrualService(ctrl)
	svc := &OrderWorkerService{
		svc: mockAccrualService,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	inputCh := make(chan entity.Order, 2)
	inputCh <- entity.Order{Number: "123", UserID: 1}
	inputCh <- entity.Order{Number: "456", UserID: 2}
	close(inputCh)

	mockAccrualService.EXPECT().
		GetOrderAccrual("123").
		Return(entity.Order{Number: "123", UserID: 1}, nil)

	mockAccrualService.EXPECT().
		GetOrderAccrual("456").
		Return(entity.Order{Number: "456", UserID: 2}, nil)

	channels := svc.fanOut(ctx, inputCh)
	assert.Len(t, channels, 5) // numWorkers = 5

	var received []result
	var wg sync.WaitGroup
	for _, ch := range channels {
		wg.Add(1)
		go func(ch chan result) {
			defer wg.Done()
			for res := range ch {
				received = append(received, res)
			}
		}(ch)
	}
	wg.Wait()

	assert.Len(t, received, 2) // Должны получить оба заказа
}

func TestOrderWorkerService_fanIn(t *testing.T) {
	svc := &OrderWorkerService{}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	inputChans := make([]chan result, 2)
	inputChans[0] = make(chan result, 1)
	inputChans[1] = make(chan result, 1)

	results := []result{
		{order: entity.Order{Number: "123"}, err: nil},
		{order: entity.Order{Number: "456"}, err: errors.New("error")},
	}

	inputChans[0] <- results[0]
	inputChans[1] <- results[1]
	close(inputChans[0])
	close(inputChans[1])

	finalCh := svc.fanIn(ctx, inputChans)
	received := collectResults(finalCh)

	assert.Len(t, received, 2)
	assert.Contains(t, received, results[0])
	assert.Contains(t, received, results[1])
}

func TestOrderWorkerService_save(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockstorage(ctrl)
	mockRepo := NewMockorderBalanceRepo(ctrl)
	log := logger.NewLogger()
	svc := &OrderWorkerService{log: log, storage: mockStorage, repo: mockRepo}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	resultCh := make(chan result, 2)
	resultCh <- result{order: entity.Order{Number: "123", Status: entity.OrderNew}, err: nil}
	resultCh <- result{order: entity.Order{Number: "456", Status: entity.OrderProcessed}, err: errors.New("accrual error")}
	close(resultCh)

	t.Run("save orders", func(t *testing.T) {
		mockRepo.EXPECT().
			UpdateOrderBalance(ctx, gomock.Any()).
			Times(2).
			Return(nil)

		mockStorage.EXPECT().
			Push(gomock.Any()).
			Times(1) // Только для не терминального статуса (OrderNew)

		svc.save(ctx, resultCh)
		time.Sleep(100 * time.Millisecond) // Даем время горутине обработать
	})

	t.Run("update error", func(t *testing.T) {
		resultCh = make(chan result, 1)
		resultCh <- result{order: entity.Order{Number: "123", Status: entity.OrderNew}, err: nil}
		close(resultCh)

		updateErr := errors.New("update error")
		mockRepo.EXPECT().
			UpdateOrderBalance(ctx, gomock.Any()).
			Return(updateErr)

		mockStorage.EXPECT().
			Push(gomock.Any()).
			Times(1)

		svc.save(ctx, resultCh)
		time.Sleep(100 * time.Millisecond)
	})
}

func TestOrderWorkerService_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockstorage(ctrl)
	mockAccrualService := NewMockaccrualService(ctrl)
	mockRepo := NewMockorderBalanceRepo(ctrl)
	log := logger.NewLogger()
	svc := &OrderWorkerService{log: log, svc: mockAccrualService, storage: mockStorage, repo: mockRepo}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	orders := []entity.Order{
		{Number: "123", UserID: 1, Status: entity.OrderNew},
	}
	var accrualTestNum float32 = 100
	accrualOrder := entity.Order{Number: "123", Status: entity.OrderProcessed, Accrual: &accrualTestNum}

	mockStorage.EXPECT().
		Get(5).
		Return(orders).
		Times(1)

	mockAccrualService.EXPECT().
		GetOrderAccrual(orders[0].Number).
		Return(accrualOrder, nil)

	mockRepo.EXPECT().
		UpdateOrderBalance(ctx, gomock.Any()).
		Return(nil)

	svc.Run(ctx)
	time.Sleep(150 * time.Millisecond) // Даем время для одного цикла
}

// Вспомогательная функция для сбора результатов из канала
func collectResults(ch chan result) []result {
	var results []result
	for res := range ch {
		results = append(results, res)
	}
	return results
}
