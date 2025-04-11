package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type orderService interface {
	SaveOrder(ctx context.Context, order string) error
}

type orderHandler struct {
	service orderService
}

type authMiddleware interface {
	Validate(next http.Handler) http.Handler
}

func NewOrdersHandler(middleware authMiddleware) func(chi.Router) {
	h := &orderHandler{}
	return func(r chi.Router) {
		r.Route("/orders", func(r chi.Router) {
			r.Use(middleware.Validate)
			r.Post("/", h.SaveOrderHandler)
		})

	}
}

func (h *orderHandler) SaveOrderHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("UserId")
	if userId == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	fmt.Println(userId)
}
