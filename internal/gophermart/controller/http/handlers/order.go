package routes

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/MxTrap/gophermart/internal/gophermart/controller/http/utils"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/go-chi/chi/v5"
)

type orderService interface {
	SaveOrder(ctx context.Context, order entity.Order) error
}

type orderHandler struct {
	service orderService
}

type authMiddleware interface {
	Validate(next http.Handler) http.Handler
}

func NewOrdersHandler(middleware authMiddleware, service orderService) func(chi.Router) {
	h := &orderHandler{}
	return func(r chi.Router) {
		r.Route("/orders", func(r chi.Router) {
			r.Use(middleware.Validate)
			r.Post("/", h.SaveOrderHandler)
		})

	}
}

func (h *orderHandler) SaveOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserID(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	body, err := io.ReadAll(r.Body)

	fmt.Print(string(body))

	h.service.SaveOrder(r.Context(), entity.Order{UserID: userID})

}
