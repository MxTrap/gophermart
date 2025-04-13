package utils

import (
	"context"
	"errors"
)

func GetUserID(ctx context.Context) (int64, error) {
	uid := ctx.Value("userID")
	userID, ok := uid.(int64)
	if !ok {
		return 0, errors.New("unknown user id")
	}
	return userID, nil
}
