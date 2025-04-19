package utils

import (
	"context"
	"github.com/MxTrap/gophermart/internal/gophermart/controller/http/middlewares"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetUserID(t *testing.T) {
	userID := int64(123)
	userIDKey := middlewares.UserIDKey("UserID")

	t.Run("successful user ID extraction", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), userIDKey, userID)
		result, err := GetUserID(ctx)
		assert.NoError(t, err)
		assert.Equal(t, userID, result)
	})

	t.Run("missing user ID in context", func(t *testing.T) {
		ctx := context.Background()
		result, err := GetUserID(ctx)
		assert.Error(t, err)
		assert.Equal(t, int64(0), result)
		assert.Equal(t, "unknown user id", err.Error())
	})

	t.Run("invalid user ID type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), userIDKey, "not-an-int64")
		result, err := GetUserID(ctx)
		assert.Error(t, err)
		assert.Equal(t, int64(0), result)
		assert.Equal(t, "unknown user id", err.Error())
	})

	t.Run("different context key", func(t *testing.T) {
		wrongKey := middlewares.UserIDKey("WrongKey")
		ctx := context.WithValue(context.Background(), wrongKey, userID)
		result, err := GetUserID(ctx)
		assert.Error(t, err)
		assert.Equal(t, int64(0), result)
		assert.Equal(t, "unknown user id", err.Error())
	})
}
