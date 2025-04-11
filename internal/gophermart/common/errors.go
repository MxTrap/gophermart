package common

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

var (
	ErrInternalError = errors.New("internal error")
)

var (
	ErrInvalidToken    = errors.New("invalid_token")
	ErrTokenHasExpired = errors.New("token_has_expired")
)

var (
	ErrUsernameAlreadyExist = errors.New("user already exist")
	ErrEmailAlreadyExist    = errors.New("user email already exist")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrUserNotFound         = errors.New("user not found")
	ErrUserAlreadyExist     = errors.New("user already exist")
)
