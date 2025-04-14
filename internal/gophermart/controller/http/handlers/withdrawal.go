package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/MxTrap/gophermart/internal/gophermart/controller/http/utils"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type withdrawer interface {
	GetAll(ctx context.Context, userId int64) ([]entity.Withdrawal, error)
}

type withdrawalHandler struct {
	svc withdrawer
}

func NewWithdrawalHandler(
	middleware authMiddleware,
	svc withdrawer,
) func(chi.Router) {
	h := &withdrawalHandler{svc: svc}

	return func(r chi.Router) {
		r.Route("/withdrawals", func(r chi.Router) {
			r.Use(middleware.Validate)
			r.Get("/", h.GetAll)
		})

	}
}

type withdrawalDTO struct {
	Order       string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func (h *withdrawalHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserID(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.svc.GetAll(r.Context(), userID)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if withdrawals == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var withdrawalsDto []withdrawalDTO
	for _, withdrawal := range withdrawals {
		withdrawalsDto = append(
			withdrawalsDto,
			withdrawalDTO{
				Order:       withdrawal.Order,
				Sum:         withdrawal.Sum,
				ProcessedAt: withdrawal.ProcessedAt.Format(time.RFC3339),
			},
		)
	}

	render.JSON(w, r, withdrawalsDto)
}
