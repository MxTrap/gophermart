package services

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/MxTrap/gophermart/logger"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/errgroup"
)

type tokenDeleter interface {
	DeleteToken(ctx context.Context, token string) error
	ClearTokens(ctx context.Context, userID int64) error
}

type tokenRepo interface {
	GetTokens(ctx context.Context, userID int64) ([]entity.RefreshToken, error)
	SaveToken(ctx context.Context, token entity.RefreshToken) error
	tokenDeleter
}

type userFinder interface {
	FindUserByUsername(ctx context.Context, username string) (entity.User, error)
	FindUserById(ctx context.Context, userID int64) (entity.User, error)
}

type userSaver interface {
	SaveUser(ctx context.Context, user entity.User) (uid int64, err error)
}

type userRepo interface {
	userSaver
	userFinder
}

type jwtService interface {
	GenerateAccessToken(user entity.User, ttl time.Duration) (string, error)
	GenerateRefreshToken(userID int64) string
	GetUserIDFromRefreshToken(token entity.Token) (int64, error)
}

type TokenConfig struct {
	MaxCount int8
	TTL      struct {
		Access  time.Duration
		Refresh time.Duration
	}
}

type AuthService struct {
	log         *logger.Logger
	userRepo    userRepo
	tokenRepo   tokenRepo
	jwtService  jwtService
	tokenConfig TokenConfig
}

func NewAuthService(
	logger *logger.Logger,
	userRepo userRepo,
	tokenRepo tokenRepo,
	jwtService jwtService,
	tokenConfig TokenConfig,
) *AuthService {
	return &AuthService{
		log:         logger,
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		jwtService:  jwtService,
		tokenConfig: tokenConfig,
	}
}

func (s AuthService) generateTokens(user entity.User) (entity.Tokens, error) {
	tokenPair := entity.Tokens{}
	type res struct {
		token string
		err   error
	}
	accessToken := make(chan res)
	refreshToken := make(chan string)

	go func() {
		token, err := s.jwtService.GenerateAccessToken(user, s.tokenConfig.TTL.Access)
		accessToken <- res{token, err}
	}()

	go func() {
		refreshToken <- s.jwtService.GenerateRefreshToken(user.Id)
	}()

	for range 2 {
		select {
		case t := <-accessToken:
			if t.err != nil {
				return tokenPair, t.err
			}
			tokenPair.AccessToken = entity.Token(t.token)
		case t := <-refreshToken:
			tokenPair.RefreshToken = entity.Token(t)
		}
	}

	return tokenPair, nil
}

func (s AuthService) saveRefreshToken(ctx context.Context, refreshToken entity.Token, userID int64) error {
	var token entity.RefreshToken
	token.UserID = userID
	token.RefreshToken = refreshToken
	token.ExpirationTime = time.Now().Add(s.tokenConfig.TTL.Refresh).Unix()

	return s.tokenRepo.SaveToken(ctx, token)
}

func (s *AuthService) RegisterNewUser(ctx context.Context, user entity.User) (entity.Tokens, error) {
	log := s.log.With("op", "AuthService.RegisterNewUser", "login", user.Login)
	var tokens entity.Tokens

	existingUser, err := s.userRepo.FindUserByUsername(ctx, user.Login)
	if err != nil && !errors.Is(err, common.ErrUserNotFound) {
		log.Error("failed to find existing user", err)
		return tokens, common.ErrInternalError
	}
	if existingUser != (entity.User{}) {
		return tokens, common.ErrUserAlreadyExist
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	if err != nil {
		log.Error("failed to generate password hash: ", err)

		return tokens, common.ErrInternalError
	}

	user.Password = string(passHash)
	id, err := s.userRepo.SaveUser(ctx, user)

	if err != nil {
		log.Error("failed to save user", err)
		return tokens, common.ErrInternalError
	}
	user.Id = id
	tokens, err = s.generateTokens(user)
	if err != nil {
		log.Error("failed to generate tokens", err)
		return tokens, err
	}

	err = s.saveRefreshToken(ctx, tokens.RefreshToken, id)
	if err != nil {
		log.Error("failed to save token")
	}

	return tokens, nil
}

func (s *AuthService) Login(ctx context.Context, user entity.User) (entity.Tokens, error) {
	log := s.log.With("op", "AuthService.Login", "login", user.Login)

	var tokens entity.Tokens

	existingUser, err := s.userRepo.FindUserByUsername(ctx, user.Login)
	log.Info("Try to login with", existingUser.Id, existingUser.Login)
	if err != nil {
		if errors.Is(err, common.ErrUserNotFound) {
			log.Info("user not found", err)
			return tokens, common.ErrInvalidCredentials
		}

		log.Error("failed to find user", err)
		return tokens, common.ErrInternalError
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password)); err != nil {
		log.Info("invalid credentials", err)

		return tokens, common.ErrInvalidCredentials
	}

	tokens, err = s.generateTokens(existingUser)
	if err != nil {
		log.Error("failed to generate tokens", err)
		return tokens, common.ErrInternalError
	}

	err = s.saveRefreshToken(ctx, tokens.RefreshToken, existingUser.Id)
	if err != nil {
		log.Error("failed to save refresh token", err)
		return tokens, common.ErrInternalError
	}

	return tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	err := s.tokenRepo.DeleteToken(ctx, token)
	if err != nil {
		s.log.Error("failed to delete token", err)
		return common.ErrInternalError
	}
	return nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken entity.Token) (entity.Tokens, error) {
	log := s.log.With("op", "AuthService.Refresh", "refresh token", refreshToken)
	log.Info("try to refresh")

	errs, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var tokens entity.Tokens

	userID, err := s.jwtService.GetUserIDFromRefreshToken(refreshToken)

	savedTokens, err := s.tokenRepo.GetTokens(ctx, userID)
	if err != nil {
		log.Error("failed to get tokens", err)
		if errors.Is(err, common.ErrUserNotFound) {
			return tokens, common.ErrInvalidCredentials
		}

		return tokens, common.ErrInternalError
	}

	if len(savedTokens) > int(s.tokenConfig.MaxCount) {
		errs.Go(func() error {
			return s.tokenRepo.ClearTokens(ctx, userID)
		})
	}

	tokenIdx := slices.IndexFunc(savedTokens, func(token entity.RefreshToken) bool {
		return token.RefreshToken == refreshToken
	})

	if tokenIdx == -1 {
		return tokens, common.ErrInvalidToken
	}

	savedToken := savedTokens[tokenIdx]

	if savedToken.ExpirationTime < time.Now().Unix() {
		log.Error(
			fmt.Sprintf("%s: has expired %d", savedToken.RefreshToken, savedToken.ExpirationTime-time.Now().Unix()))
		return tokens, common.ErrTokenHasExpired
	}

	user, err := s.userRepo.FindUserById(ctx, userID)

	if err != nil {
		return tokens, err
	}
	tokens, err = s.generateTokens(user)

	errs.Go(func() error {
		return s.tokenRepo.DeleteToken(ctx, string(savedToken.RefreshToken))
	})

	errs.Go(func() error {
		return s.saveRefreshToken(ctx, refreshToken, user.Id)
	})

	if err = errs.Wait(); err != nil {
		log.Error("failed to update token", err)
		return entity.Tokens{}, common.ErrInternalError
	}

	return tokens, err
}
