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

func (s *Storage) Get() *entity.Order {
	self := *s.queue
	var el *entity.Order
	if len(self) > 0 {
		el, *s.queue = &self[0], self[1:]
	}

	return el
}
