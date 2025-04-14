package usecase

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/MxTrap/gophermart/logger"
)

type orderRepository interface {
	SaveOrder(ctx context.Context, order entity.Order) error
	FindOrder(ctx context.Context, number string) (entity.Order, error)
	UpdateOrder(ctx context.Context, order entity.Order) error
}

type accrualService interface {
	GetOrderAccrual(number string) (entity.Order, error)
}

type OrderService struct {
	queue   common.Queue[string]
	log     *logger.Logger
	service accrualService
	repo    orderRepository
}

func NewOrderService(log *logger.Logger, service accrualService, repo orderRepository) *OrderService {
	return &OrderService{
		queue:   common.Queue[string]{},
		log:     log,
		service: service,
		repo:    repo,
	}
}

func (*OrderService) checkOrderNumber(number string) bool {
	var sum int
	parity := len(number) % 2
	for i := range len(number) {
		digit, _ := strconv.Atoi(string(number[i]))
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0

}

func (*OrderService) isTerminalStatus(status string) bool {
	return status == entity.OrderInvalid || status == entity.OrderProcessed
}

func (s *OrderService) SaveOrder(ctx context.Context, order entity.Order) error {
	log := s.log.With("op:", "OrderService.SaveOrder")
	if !s.checkOrderNumber(order.Number) {
		return common.ErrInvalidOrderNumber
	}

	existingOrder, err := s.repo.FindOrder(ctx, order.Number)
	if err != nil {
		log.Error(err)
		return common.ErrInternalError
	}

	if existingOrder.Number != "" {
		if existingOrder.UserID != order.UserID {
			return common.ErrOrderRegisteredByAnother
		}
		return common.ErrOrderAlreadyExist
	}

	accrualOrder, err := s.service.GetOrderAccrual(order.Number)
	if err != nil {
		if !errors.Is(err, common.ErrNonExistentOrder) {
			log.Error(err)
			return common.ErrInternalError
		}
		order.Status = entity.OrderNew
	}

	if order != (entity.Order{}) {
		order.Status = accrualOrder.Status
		order.Accrual = accrualOrder.Accrual
	}

	order.UploadedAt = time.Now().UTC()

	err = s.repo.SaveOrder(ctx, order)
	if err != nil {
		log.Error(err)
		return common.ErrInternalError
	}

	if !s.isTerminalStatus(order.Status) {
		s.queue.Push(order.Number)
	}

	return nil
}

func (s *OrderService) worker(ctx context.Context, id int, job chan string) {
	log := s.log.With("worker#", id)
	const maxRettyCount = 3
outer:
	for number := range job {
		var order entity.Order
		var err error
		for i := range maxRettyCount {
			order, err = s.service.GetOrderAccrual(number)
			if s.isTerminalStatus(order.Status) {
				order.Status = number
				s.repo.UpdateOrder(ctx, order)
				break outer
			}
			time.Sleep(time.Duration(i*2) * time.Second)
		}
		if err != nil {
			log.Error(err)
		}
		s.queue.Push(order.Status)
	}
}

func (s *OrderService) UpdateOrder(number string) {
	s.queue.Push(number)
}

func (s *OrderService) Run(ctx context.Context) {
	const jobNum = 5
	jobs := make(chan string, jobNum)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				close(jobs)
				return
			default:
				if len(s.queue) >= 0 {
					jobs <- s.queue.Deque()
				}
			}
		}
	}(ctx)
	for i := range jobNum {
		go s.worker(ctx, i, jobs)
	}

}
