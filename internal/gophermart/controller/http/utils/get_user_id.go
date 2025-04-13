package utils

import (
	"context"
	"errors"
)

func GetUserId(ctx context.Context) (int64, error) {
	uid := ctx.Value("UserId")
	userId, ok := uid.(int64)
	if !ok {
		return 0, errors.New("unknown user id")
	}
	return userId, nil
}
