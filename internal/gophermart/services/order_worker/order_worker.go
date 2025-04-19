package order_worker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/MxTrap/gophermart/logger"
)

type accrualService interface {
	GetOrderAccrual(number string) (entity.Order, error)
}

type storage interface {
	Push(el entity.Order)
	Get(elemCount int) []entity.Order
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

type result struct {
	order entity.Order
	err   error
}

func (*OrderWorkerService) isTerminalStatus(status string) bool {
	return status == entity.OrderInvalid || status == entity.OrderProcessed
}

func (s *OrderWorkerService) save(ctx context.Context, ch chan result) {
	go func() {
		for res := range ch {
			select {
			case <-ctx.Done():
				return
			default:
				if res.err != nil {
					s.log.Error("failed to save ch: %v", res.err)
				}
				err := s.repo.UpdateOrderBalance(ctx, res.order)
				if err != nil || !s.isTerminalStatus(res.order.Status) {
					s.storage.Push(res.order)
				}
			}
		}
	}()
}

func (s *OrderWorkerService) update(ctx context.Context, inputChan chan entity.Order) chan result {
	resultCh := make(chan result)

	go func() {
		defer close(resultCh)
		for order := range inputChan {
			accrualOrder, err := s.svc.GetOrderAccrual(order.Number)
			if accrualOrder.Status == order.Status {
				err = errors.New("order has already been processed")
			}
			if err == nil {
				accrualOrder.UserID = order.UserID
				accrualOrder.Number = order.Number
			}

			select {
			case <-ctx.Done():
				return
			case resultCh <- result{accrualOrder, err}:
			}
		}

	}()

	return resultCh
}

func (*OrderWorkerService) generate(ctx context.Context, input []entity.Order) chan entity.Order {
	inputCh := make(chan entity.Order)

	go func() {
		defer close(inputCh)

		for _, data := range input {
			select {
			case <-ctx.Done():
				return
			case inputCh <- data:
			}
		}
	}()

	return inputCh
}

func (s *OrderWorkerService) fanOut(ctx context.Context, input chan entity.Order) []chan result {
	numWorkers := 5

	channels := make([]chan result, numWorkers)

	for i := 0; i < numWorkers; i++ {
		resultCh := s.update(ctx, input)
		channels[i] = resultCh
	}

	return channels
}

func (*OrderWorkerService) fanIn(ctx context.Context, results []chan result) chan result {
	finalCh := make(chan result)

	var wg sync.WaitGroup

	for _, ch := range results {
		chClosure := ch

		wg.Add(1)

		go func() {
			defer wg.Done()
			for data := range chClosure {
				select {
				case <-ctx.Done():
					return
				case finalCh <- data:
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	return finalCh

}

func (s *OrderWorkerService) Run(ctx context.Context) {
	const jobNum = 5
	const updateDelay = time.Second * 5

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				orders := s.storage.Get(jobNum)
				inputCh := s.generate(ctx, orders)
				channels := s.fanOut(ctx, inputCh)
				resultCh := s.fanIn(ctx, channels)
				s.save(ctx, resultCh)
			}

			time.Sleep(updateDelay)
		}
	}(ctx)
}
