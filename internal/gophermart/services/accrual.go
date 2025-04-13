package services

import (
	"fmt"

	"github.com/MxTrap/gophermart/logger"
	"resty.dev/v3"
)

type AccrualService struct {
	log *logger.Logger
	url string
}

func NewAccrualService(log *logger.Logger, url string) *AccrualService {
	return &AccrualService{
		log: log,
		url: url,
	}
}

func (s *AccrualService) GetAccrual(number string) {
	r := resty.New()
	res, err := r.R().Get(fmt.Sprintf("%s/api/orders/%s", s.url, number))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res)
}
