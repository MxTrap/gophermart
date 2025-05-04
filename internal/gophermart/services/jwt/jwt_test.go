package jwt

import (
	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestJwtService_GenerateAccessToken(t *testing.T) {
	secretKey := "test-secret-key"
	svc := NewJWTService(secretKey)
	user := entity.User{
		ID:    1,
		Login: "testuser",
	}
	ttl := 1 * time.Hour

	t.Run("successful token generation", func(t *testing.T) {
		token, err := svc.GenerateAccessToken(user, ttl)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Проверяем содержимое токена
		parsedToken, err := jwt.Parse(string(token), func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		assert.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		assert.True(t, ok)
		assert.Equal(t, float64(user.ID), claims["uid"])
		assert.Equal(t, user.Login, claims["login"])
		assert.InDelta(t, time.Now().Add(ttl).Unix(), claims["exp"], 5)
	})

}

func TestJwtService_Parse(t *testing.T) {
	secretKey := "test-secret-key"
	svc := NewJWTService(secretKey)
	user := entity.User{
		ID:    1,
		Login: "testuser",
	}
	ttl := 1 * time.Hour

	t.Run("successful parse", func(t *testing.T) {
		token, err := svc.GenerateAccessToken(user, ttl)
		assert.NoError(t, err)

		uid, err := svc.Parse(token)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, uid)
	})

	t.Run("expired token", func(t *testing.T) {
		expiredTTL := -1 * time.Hour
		token, err := svc.GenerateAccessToken(user, expiredTTL)
		assert.NoError(t, err)

		uid, err := svc.Parse(token)
		assert.ErrorIs(t, err, common.ErrTokenHasExpired)
		assert.Equal(t, int64(0), uid)
	})

	t.Run("invalid token", func(t *testing.T) {
		invalidToken := entity.Token("invalid.token.string")
		uid, err := svc.Parse(invalidToken)
		assert.Error(t, err)
		assert.NotEqual(t, common.ErrTokenHasExpired, err) // Ошибка не связана с истечением
		assert.Equal(t, int64(0), uid)
	})

	t.Run("wrong secret key", func(t *testing.T) {
		token, err := svc.GenerateAccessToken(user, ttl)
		assert.NoError(t, err)

		wrongSvc := NewJWTService("wrong-secret-key")
		uid, err := wrongSvc.Parse(token)
		assert.Error(t, err)
		assert.Equal(t, int64(0), uid)
	})

	t.Run("missing uid claim", func(t *testing.T) {
		// Создаем токен без uid
		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["login"] = user.Login
		claims["exp"] = time.Now().Add(ttl).Unix()
		signedString, err := token.SignedString([]byte(secretKey))
		assert.NoError(t, err)

		uid, err := svc.Parse(entity.Token(signedString))
		assert.ErrorIs(t, err, common.ErrInvalidToken)
		assert.Equal(t, int64(0), uid)
	})

	t.Run("invalid uid type", func(t *testing.T) {
		// Создаем токен с некорректным типом uid
		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["uid"] = "not-a-number"
		claims["login"] = user.Login
		claims["exp"] = time.Now().Add(ttl).Unix()
		signedString, err := token.SignedString([]byte(secretKey))
		assert.NoError(t, err)

		uid, err := svc.Parse(entity.Token(signedString))
		assert.ErrorIs(t, err, common.ErrInvalidToken)
		assert.Equal(t, int64(0), uid)
	})
}
