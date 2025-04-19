package auth

import (
	"context"
	"errors"
	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/MxTrap/gophermart/logger"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)

func TestAuthService_RegisterNewUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := NewMockuserRepo(ctrl)
	mockJwtService := NewMockjwtService(ctrl)
	log := logger.NewLogger()
	tokenTTL := 24 * time.Hour
	ctx := context.Background()

	authService := NewAuthService(log, mockUserRepo, mockJwtService, tokenTTL)

	user := entity.User{
		Login:    "testuser",
		Password: "password123",
	}
	userID := int64(1)
	token := entity.Token("jwt-token")

	t.Run("successful registration", func(t *testing.T) {
		mockUserRepo.EXPECT().
			FindUserByUsername(ctx, user.Login).
			Return(entity.User{}, common.ErrUserNotFound)

		mockUserRepo.EXPECT().
			SaveUser(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, u entity.User) (int64, error) {
				assert.NotEqual(t, user.Password, u.Password) // Пароль должен быть хеширован
				return userID, nil
			})

		mockJwtService.EXPECT().
			GenerateAccessToken(gomock.Any(), tokenTTL).
			DoAndReturn(func(u entity.User, _ time.Duration) (entity.Token, error) {
				assert.Equal(t, userID, u.ID)
				assert.Equal(t, user.Login, u.Login)
				return token, nil
			})

		resultToken, err := authService.RegisterNewUser(ctx, user)
		assert.NoError(t, err)
		assert.Equal(t, token, resultToken)
	})

	t.Run("user already exists", func(t *testing.T) {
		existingUser := entity.User{Login: user.Login, ID: userID}
		mockUserRepo.EXPECT().
			FindUserByUsername(ctx, user.Login).
			Return(existingUser, nil)

		resultToken, err := authService.RegisterNewUser(ctx, user)
		assert.ErrorIs(t, err, common.ErrUserAlreadyExist)
		assert.Equal(t, entity.Token(""), resultToken)
	})

	t.Run("find user error", func(t *testing.T) {
		dbErr := errors.New("database error")
		mockUserRepo.EXPECT().
			FindUserByUsername(ctx, user.Login).
			Return(entity.User{}, dbErr)

		resultToken, err := authService.RegisterNewUser(ctx, user)
		assert.ErrorIs(t, err, common.ErrInternalError)
		assert.Equal(t, entity.Token(""), resultToken)
	})

	t.Run("save user error", func(t *testing.T) {
		mockUserRepo.EXPECT().
			FindUserByUsername(ctx, user.Login).
			Return(entity.User{}, common.ErrUserNotFound)

		dbErr := errors.New("save error")
		mockUserRepo.EXPECT().
			SaveUser(ctx, gomock.Any()).
			Return(int64(0), dbErr)

		resultToken, err := authService.RegisterNewUser(ctx, user)
		assert.ErrorIs(t, err, common.ErrInternalError)
		assert.Equal(t, entity.Token(""), resultToken)
	})

	t.Run("generate token error", func(t *testing.T) {
		mockUserRepo.EXPECT().
			FindUserByUsername(ctx, user.Login).
			Return(entity.User{}, common.ErrUserNotFound)

		mockUserRepo.EXPECT().
			SaveUser(ctx, gomock.Any()).
			Return(userID, nil)

		tokenErr := errors.New("token generation error")
		mockJwtService.EXPECT().
			GenerateAccessToken(gomock.Any(), tokenTTL).
			Return(entity.Token(""), tokenErr)

		resultToken, err := authService.RegisterNewUser(ctx, user)
		assert.ErrorIs(t, err, tokenErr)
		assert.Equal(t, entity.Token(""), resultToken)
	})
}

func TestAuthService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := NewMockuserRepo(ctrl)
	mockJwtService := NewMockjwtService(ctrl)
	log := logger.NewLogger()
	tokenTTL := 24 * time.Hour
	ctx := context.Background()

	authService := NewAuthService(log, mockUserRepo, mockJwtService, tokenTTL)

	user := entity.User{
		Login:    "testuser",
		Password: "password123",
	}
	userID := int64(1)
	token := entity.Token("jwt-token")

	// Подготовка хешированного пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	existingUser := entity.User{
		ID:       userID,
		Login:    user.Login,
		Password: string(hashedPassword),
	}

	t.Run("successful login", func(t *testing.T) {
		mockUserRepo.EXPECT().
			FindUserByUsername(ctx, user.Login).
			Return(existingUser, nil)

		mockJwtService.EXPECT().
			GenerateAccessToken(existingUser, tokenTTL).
			Return(token, nil)

		resultToken, err := authService.Login(ctx, user)
		assert.NoError(t, err)
		assert.Equal(t, token, resultToken)
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepo.EXPECT().
			FindUserByUsername(ctx, user.Login).
			Return(entity.User{}, common.ErrUserNotFound)

		resultToken, err := authService.Login(ctx, user)
		assert.ErrorIs(t, err, common.ErrInvalidCredentials)
		assert.Equal(t, entity.Token(""), resultToken)
	})

	t.Run("invalid password", func(t *testing.T) {
		mockUserRepo.EXPECT().
			FindUserByUsername(ctx, user.Login).
			Return(existingUser, nil)

		invalidUser := entity.User{
			Login:    user.Login,
			Password: "wrongpassword",
		}

		resultToken, err := authService.Login(ctx, invalidUser)
		assert.ErrorIs(t, err, common.ErrInvalidCredentials)
		assert.Equal(t, entity.Token(""), resultToken)
	})

	t.Run("database error", func(t *testing.T) {
		dbErr := errors.New("database error")
		mockUserRepo.EXPECT().
			FindUserByUsername(ctx, user.Login).
			Return(entity.User{}, dbErr)

		resultToken, err := authService.Login(ctx, user)
		assert.ErrorIs(t, err, common.ErrInternalError)
		assert.Equal(t, entity.Token(""), resultToken)
	})

	t.Run("generate token error", func(t *testing.T) {
		mockUserRepo.EXPECT().
			FindUserByUsername(ctx, user.Login).
			Return(existingUser, nil)

		tokenErr := errors.New("token generation error")
		mockJwtService.EXPECT().
			GenerateAccessToken(existingUser, tokenTTL).
			Return(entity.Token(""), tokenErr)

		resultToken, err := authService.Login(ctx, user)
		assert.ErrorIs(t, err, common.ErrInternalError)
		assert.Equal(t, entity.Token(""), resultToken)
	})
}
