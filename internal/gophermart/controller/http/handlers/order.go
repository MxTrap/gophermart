package routes

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/controller/http/utils"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/go-chi/chi/v5"
)

type orderService interface {
	SaveOrder(ctx context.Context, order entity.Order) error
	GetAll(ctx context.Context, userID int64) ([]entity.Order, error)
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
			r.Get("/", h.GetAllHandler)
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
	order := entity.Order{UserID: userID, Number: string(body)}

	err = h.service.SaveOrder(r.Context(), order)

	if err == nil {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	if errors.Is(err, common.ErrOrderAlreadyExist) {
		w.WriteHeader(http.StatusOK)
		return
	}

	if errors.Is(err, common.ErrOrderRegisteredByAnother) {
		w.WriteHeader(http.StatusConflict)
		return
	}

	if errors.Is(err, common.ErrInvalidOrderNumber) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
}

type orderDTO struct {
	Number     string   `json:"number"`
	Status     string   `json:"status"`
	Accrual    *float32 `json:"accrual,omitempty"`
	UploadedAt string   `json:"uploaded_at"`
}

func (*orderHandler) mapOrderToDTO(o entity.Order) orderDTO {
	return orderDTO{
		Number:     o.Number,
		Status:     o.Status,
		Accrual:    o.Accrual,
		UploadedAt: o.UploadedAt.Format(time.RFC3339),
	}
}

func (h *orderHandler) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserID(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	orders, err := h.service.GetAll(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if orders == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	ordersDto := make([]orderDTO, 0, len(orders))
	for _, order := range orders {
		ordersDto = append(ordersDto, h.mapOrderToDTO(order))
	}

	body, err := json.Marshal(ordersDto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
