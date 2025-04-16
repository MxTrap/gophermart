package services

import (
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
)

type Storage struct {
	queue *[]entity.Order
}

func NewStorageService() *Storage {
	return &Storage{
		queue: &[]entity.Order{},
	}
}

func (s *Storage) Push(el entity.Order) {
	*s.queue = append(*s.queue, el)
}

func (s *Storage) Get(elemCount int) []entity.Order {
	self := *s.queue
	var el []entity.Order
	if len(self) == 0 {
		return el
	}
	if len(self) > elemCount {
		el, *s.queue = self[:elemCount+1], self[elemCount+1:]
		return el
	}

	el, *s.queue = *s.queue, []entity.Order{}

	return el
}
