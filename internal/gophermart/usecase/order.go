package usecase

import (
	"context"
	"time"

	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/MxTrap/gophermart/logger"
)

type accrualService interface {
	GetOrderAccrual(number string) (entity.Order, error)
}

type storageService interface {
	Push(el entity.Order)
	Get() *entity.Order
}

type orderBalanceRepo interface {
	UpdateOrderBalance(ctx context.Context, order entity.Order) error
}

type OrderUsecase struct {
	log     *logger.Logger
	service accrualService
	storage storageService
	orderBalanceRepo
}

func NewOrderUsecase(
	log *logger.Logger,
	accrualSvc accrualService,
	storage storageService,
	repo orderBalanceRepo,
) *OrderUsecase {
	return &OrderUsecase{
		log:              log,
		service:          accrualSvc,
		storage:          storage,
		orderBalanceRepo: repo,
	}
}

func (*OrderUsecase) isTerminalStatus(status string) bool {
	return status == entity.OrderInvalid || status == entity.OrderProcessed
}

func (s *OrderUsecase) worker(ctx context.Context, id int, job chan entity.Order) {
	log := s.log.With("worker#", id)
	const maxRetryCount = 3
	const retryPause = 5
outer:
	for order := range job {
		var accrualOrder entity.Order
		var err error
		for i := range maxRetryCount {
			accrualOrder, err = s.service.GetOrderAccrual(order.Number)
			if accrualOrder.Status != "" && accrualOrder.Status != order.Status {
				accrualOrder.UserID = order.UserID
				accrualOrder.Number = order.Number
				err := s.orderBalanceRepo.UpdateOrderBalance(ctx, accrualOrder)
				if err != nil {
					log.Error(err)
				}
				if s.isTerminalStatus(accrualOrder.Status) {
					break outer
				}
			}
			time.Sleep(time.Duration(i*retryPause) * time.Second)
		}
		if err != nil {
			log.Error(err)
		}
		s.storage.Push(order)
	}
}

func (s *OrderUsecase) Run(ctx context.Context) error {
	const jobNum = 5
	jobs := make(chan entity.Order, jobNum)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				close(jobs)
				return
			default:
				order := s.storage.Get()
				if order != nil {
					jobs <- *order
				}
			}
		}
	}(ctx)
	for i := range jobNum {
		go s.worker(ctx, i, jobs)
	}

	return nil
}
