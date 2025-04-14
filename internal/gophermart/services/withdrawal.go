package services

import (
	"context"
	"time"

	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/MxTrap/gophermart/internal/utils"
	"github.com/MxTrap/gophermart/logger"
)

type withdrawer interface {
	Withdraw(ctx context.Context, userID int64, withdrawal entity.Withdrawal) error
}

type getter interface {
	GetAll(ctx context.Context, userID int64) ([]entity.Withdrawal, error)
}

type WithdrawalService struct {
	log        *logger.Logger
	withdrawer withdrawer
	getter     getter
}

func NewWithdrawalService(log *logger.Logger, withdrawer withdrawer, getter getter) *WithdrawalService {
	return &WithdrawalService{
		log:        log,
		withdrawer: withdrawer,
		getter:     getter,
	}
}

func (s *WithdrawalService) Withdraw(ctx context.Context, userID int64, withdrawal entity.Withdrawal) error {
	log := s.log.With("op", "WithdrawalService.Withdraw")
	if !utils.IsOrderNumberValid(withdrawal.Order) {
		return common.ErrInvalidOrderNumber
	}

	withdrawal.ProcessedAt = time.Now().UTC()
	err := s.withdrawer.Withdraw(ctx, userID, withdrawal)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (s *WithdrawalService) GetAll(ctx context.Context, userID int64) ([]entity.Withdrawal, error) {
	log := s.log.With("op", "WithdrawalService.GetAll")
	var withdrawals []entity.Withdrawal

	withdrawals, err := s.getter.GetAll(ctx, userID)
	if err != nil {
		log.Error(err)
		return withdrawals, err
	}

	return withdrawals, nil
}
