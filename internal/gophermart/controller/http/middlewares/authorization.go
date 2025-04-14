package middlewares

import (
	"context"
	"net/http"

	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
)

type tokenValidator interface {
	Parse(token entity.Token) (int64, error)
}

type AuhtorizationMiddleware struct {
	validator tokenValidator
}

func NewAuhtorizationMiddleware(val tokenValidator) *AuhtorizationMiddleware {
	return &AuhtorizationMiddleware{
		validator: val,
	}
}

type UserIDKey string

func (m *AuhtorizationMiddleware) Validate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, common.ErrInvalidCredentials.Error(), http.StatusUnauthorized)
			return
		}

		userID, err := m.validator.Parse(entity.Token(authHeader))
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey("UserID"), userID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)

	})
}
