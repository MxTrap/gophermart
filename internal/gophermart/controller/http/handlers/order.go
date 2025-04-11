package routes

import (
	"context"
	"net/http"

	"github.com/MxTrap/gophermart/internal/gophermart/entity"

	"github.com/go-chi/chi/v5"
)

type orderService interface {
	Login(ctx context.Context, user entity.User) (entity.Tokens, error)
	RegisterNewUser(ctx context.Context, user entity.User) (entity.Tokens, error)
}

type orderHandler struct {
	service authService
}

type authMiddleware interface {
	Validate(next http.Handler) http.Handler
}

func NewOrderHandler(service authService, middleware authMiddleware) func(chi.Router) {
	h := &orderHandler{service: service}
	return func(r chi.Router) {
		r.Use(middleware.Validate)
		r.Post("/order", h.OrderHandler)
	}
}

func (h *orderHandler) OrderHandler(w http.ResponseWriter, r *http.Request) {

}
