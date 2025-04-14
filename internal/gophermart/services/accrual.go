package services

import (
	"fmt"
	"net/http"

	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/MxTrap/gophermart/logger"
	"resty.dev/v3"
)

type AccrualService struct {
	log *logger.Logger
	url string
}

func NewWorkerService(log *logger.Logger, url string) *AccrualService {
	return &AccrualService{
		log: log,
		url: url,
	}
}

type accrualDto struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float32 `json:"accrual,omitempty"`
}

func (s *AccrualService) GetOrderAccrual(number string) (entity.Order, error) {
	var order accrualDto
	res, err := resty.New().
		R().
		SetResult(&order).
		Get(fmt.Sprintf("%s/api/orders/%s", s.url, number))
	fmt.Println(res.Result())

	if err != nil {
		return entity.Order{}, err
	}

	if res.StatusCode() == http.StatusNoContent {
		return entity.Order{}, common.ErrNonExistentOrder
	}

	return s.mapDtoToOrder(order), nil
}

func (*AccrualService) mapDtoToOrder(dto accrualDto) entity.Order {
	status := dto.Status
	if status == "REGISTERED" {
		status = entity.OrderNew
	}
	return entity.Order{
		Status:  status,
		Accrual: dto.Accrual,
	}
}
