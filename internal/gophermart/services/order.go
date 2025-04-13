package services

import (
	"strconv"

	"github.com/MxTrap/gophermart/internal/gophermart/entity"
)

type OrderService struct {
}

func NewOrderService() *OrderService {
	return &OrderService{}
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

func (s *OrderService) SaveOrder(order entity.Order) error {
	if !s.checkOrderNumber(order.Number) {

	}

	return nil
}
