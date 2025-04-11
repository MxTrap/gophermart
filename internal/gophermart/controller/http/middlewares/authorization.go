package middlewares

import (
	"fmt"
	"net/http"
	"strings"

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

func (m *AuhtorizationMiddleware) Validate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		const tokenType = "Bearer"
		if strings.Index(authHeader, tokenType) == -1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userId, err := m.validator.Parse(entity.Token(authHeader[len(tokenType)+1:]))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		r.Header.Add("UserId", fmt.Sprintf("%d", userId))

		next.ServeHTTP(w, r)
	})
}
