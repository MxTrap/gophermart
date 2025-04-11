package middlewares

import (
	"net/http"
	"strings"
)

type tokenValidator interface {
	Validate(token string) error
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
			w.WriteHeader(http.StatusForbidden)
			return
		}

		const tokenType = "Bearer"
		if strings.Index(authHeader, tokenType) == -1 {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if err := m.validator.Validate(authHeader[len(tokenType)+1:]); err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
