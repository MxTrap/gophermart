package order

import (
	"context"
	"time"

	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/MxTrap/gophermart/internal/utils"
	"github.com/MxTrap/gophermart/logger"
)

type storageService interface {
	Push(order entity.Order)
}

type orderRepository interface {
	Save(ctx context.Context, order entity.Order) error
	Find(ctx context.Context, number string) (entity.Order, error)
	GetAll(ctx context.Context, userID int64) ([]entity.Order, error)
}

type OrderService struct {
	log       *logger.Logger
	service   storageService
	orderRepo orderRepository
}

func NewOrderService(
	log *logger.Logger,
	service storageService,
	orderRepo orderRepository,
) *OrderService {
	return &OrderService{
		log:       log,
		service:   service,
		orderRepo: orderRepo,
	}
}

func (s *OrderService) SaveOrder(ctx context.Context, order entity.Order) error {
	log := s.log.With("op", "OrderService.SaveOrder")
	if !utils.IsOrderNumberValid(order.Number) {
		return common.ErrInvalidOrderNumber
	}

	existingOrder, err := s.orderRepo.Find(ctx, order.Number)
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

	order.Status = entity.OrderNew
	order.UploadedAt = time.Now().UTC()

	err = s.orderRepo.Save(ctx, order)

	if err != nil {
		log.Error(err)
		return common.ErrInternalError
	}

	s.service.Push(order)

	return nil
}

func (s *OrderService) GetAll(ctx context.Context, userID int64) ([]entity.Order, error) {
	log := s.log.With("op", "OrderService.GetAll")
	orders, err := s.orderRepo.GetAll(ctx, userID)
	if err != nil {
		log.Error(err)
		return orders, err
	}
	return orders, nil
}
