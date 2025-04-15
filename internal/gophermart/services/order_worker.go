package services

import (
	"context"
	"time"

	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/MxTrap/gophermart/logger"
)

type accrualService interface {
	GetOrderAccrual(number string) (entity.Order, error)
}

type storage interface {
	Push(el entity.Order)
	Get() *entity.Order
}

type orderBalanceRepo interface {
	UpdateOrderBalance(ctx context.Context, order entity.Order) error
}

type OrderWorkerService struct {
	log     *logger.Logger
	svc     accrualService
	storage storage
	repo    orderBalanceRepo
}

func NewOrderWorkerService(
	log *logger.Logger,
	svc accrualService,
	storage storage,
	repo orderBalanceRepo,
) *OrderWorkerService {
	return &OrderWorkerService{
		log:     log,
		svc:     svc,
		storage: storage,
		repo:    repo,
	}
}

func (*OrderWorkerService) isTerminalStatus(status string) bool {
	return status == entity.OrderInvalid || status == entity.OrderProcessed
}

func (s *OrderWorkerService) worker(ctx context.Context, id int, job chan entity.Order) {
	log := s.log.With("worker#", id)
	const maxRetryCount = 3
	const retryPause = 5
outer:
	for order := range job {
		var accrualOrder entity.Order
		var err error
		for i := range maxRetryCount {
			accrualOrder, err = s.svc.GetOrderAccrual(order.Number)
			if err == nil && accrualOrder.Status != order.Status {
				accrualOrder.UserID = order.UserID
				accrualOrder.Number = order.Number
				err := s.repo.UpdateOrderBalance(ctx, accrualOrder)
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

func (s *OrderWorkerService) Run(ctx context.Context) {
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
}
