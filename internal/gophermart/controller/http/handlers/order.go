package routes

import (
	"context"
	"io"
	"net/http"

	"github.com/MxTrap/gophermart/internal/gophermart/controller/http/utils"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/go-chi/chi/v5"
)

type orderService interface {
	SaveOrder(ctx context.Context, order entity.Order) error
}
type authMiddleware interface {
	Validate(next http.Handler) http.Handler
}

type orderHandler struct {
	service orderService
}

func NewOrdersHandler(middleware authMiddleware, service orderService) func(chi.Router) {
	h := &orderHandler{
		service: service,
	}
	return func(r chi.Router) {
		r.Route("/orders", func(r chi.Router) {
			r.Use(middleware.Validate)
			r.Post("/", h.SaveOrderHandler)
		})

	}
}

func (h *orderHandler) SaveOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userID, err := utils.GetUserID(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(string(body)) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.service.SaveOrder(r.Context(), entity.Order{UserID: userID, Number: string(body)})
	w.WriteHeader(200)

}
