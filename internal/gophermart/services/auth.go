package services

import (
	"context"
	"errors"
	"time"

	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/MxTrap/gophermart/logger"

	"golang.org/x/crypto/bcrypt"
)

type userFinder interface {
	FindUserByUsername(ctx context.Context, username string) (entity.User, error)
	FindUserByID(ctx context.Context, userID int64) (entity.User, error)
}

type userSaver interface {
	SaveUser(ctx context.Context, user entity.User) (uid int64, err error)
}

type userRepo interface {
	userSaver
	userFinder
}

type jwtService interface {
	GenerateAccessToken(user entity.User, ttl time.Duration) (entity.Token, error)
}

type AuthService struct {
	log      *logger.Logger
	userRepo userRepo
	jwtSvc   jwtService
	tokenTTL time.Duration
}

func NewAuthService(
	logger *logger.Logger,
	userRepo userRepo,
	jwtSvc jwtService,
	tokenTTL time.Duration,
) *AuthService {
	return &AuthService{
		log:      logger,
		userRepo: userRepo,
		jwtSvc:   jwtSvc,
		tokenTTL: tokenTTL,
	}
}

func (s *AuthService) RegisterNewUser(ctx context.Context, user entity.User) (entity.Token, error) {
	log := s.log.With("op", "AuthService.RegisterNewUser", "login", user.Login)
	var token entity.Token

	existingUser, err := s.userRepo.FindUserByUsername(ctx, user.Login)
	if err != nil && !errors.Is(err, common.ErrUserNotFound) {
		log.Error("failed to find existing user", err)
		return token, common.ErrInternalError
	}
	if existingUser != (entity.User{}) {
		return token, common.ErrUserAlreadyExist
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	if err != nil {
		log.Error("failed to generate password hash: ", err)

		return token, common.ErrInternalError
	}

	user.Password = string(passHash)
	id, err := s.userRepo.SaveUser(ctx, user)

	if err != nil {
		log.Error("failed to save user", err)
		return token, common.ErrInternalError
	}
	user.ID = id
	token, err = s.jwtSvc.GenerateAccessToken(user, s.tokenTTL)
	if err != nil {
		log.Error("failed to generate tokens", err)
		return token, err
	}

	return token, nil
}

func (s *AuthService) Login(ctx context.Context, user entity.User) (entity.Token, error) {
	log := s.log.With("op", "AuthService.Login", "login", user.Login)

	var token entity.Token

	existingUser, err := s.userRepo.FindUserByUsername(ctx, user.Login)

	if err != nil {
		if errors.Is(err, common.ErrUserNotFound) {
			log.Info("user not found", err)
			return token, common.ErrInvalidCredentials
		}

		log.Error(err)
		return token, common.ErrInternalError
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password)); err != nil {
		log.Info("invalid password", err)

		return token, common.ErrInvalidCredentials
	}

	token, err = s.jwtSvc.GenerateAccessToken(existingUser, s.tokenTTL)
	if err != nil {
		log.Error("failed to generate tokens", err)
		return token, common.ErrInternalError
	}

	return token, nil
}
