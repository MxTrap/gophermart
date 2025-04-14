package services

import (
	"context"

	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/MxTrap/gophermart/logger"
)

type balanceRepo interface {
	Get(ctx context.Context, userId int64) (entity.Balance, error)
}

type BalanceService struct {
	log  *logger.Logger
	repo balanceRepo
}

func NewBalanceService(log *logger.Logger, repo balanceRepo) *BalanceService {
	return &BalanceService{
		log:  log,
		repo: repo,
	}
}

func (s *BalanceService) Get(ctx context.Context, userId int64) (entity.Balance, error) {
	log := s.log.With("op", "BalanceService.Get")

	balance, err := s.repo.Get(ctx, userId)

	if err != nil {
		log.Error(err)
		return balance, err
	}

	return balance, nil
}
