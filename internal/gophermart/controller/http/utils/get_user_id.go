package utils

import (
	"context"
	"errors"

	"github.com/MxTrap/gophermart/internal/gophermart/controller/http/middlewares"
)

func GetUserID(ctx context.Context) (int64, error) {
	userID, ok := ctx.Value(middlewares.UserIDKey("UserID")).(int64)
	if !ok {
		return 0, errors.New("unknown user id")
	}
	return userID, nil
}
