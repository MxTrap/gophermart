package services

import (
	"context"
	"strconv"

	"github.com/MxTrap/gophermart/internal/gophermart/entity"
)

type accrualService interface {
	GetAccrual(number string)
}

type OrderService struct {
	service accrualService
}

func NewOrderService(service accrualService) *OrderService {
	return &OrderService{
		service: service,
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

func (s *OrderService) SaveOrder(ctx context.Context, order entity.Order) error {
	if !s.checkOrderNumber(order.Number) {

	}

	s.service.GetAccrual(order.Number)

	return nil
}
