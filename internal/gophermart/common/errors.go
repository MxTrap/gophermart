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
	ErrUserNotFound         = errors.New("user not found")
	ErrUserAlreadyExist     = errors.New("user already exist")
)

var (
	ErrInvalidOrderNumber       = errors.New("invalid order number")
	ErrOrderAlreadyExist        = errors.New("order number already exist")
	ErrOrderRegisteredByAnother = errors.New("order registered by another user")
	ErrNonExistentOrder         = errors.New("order does not exist")
)
