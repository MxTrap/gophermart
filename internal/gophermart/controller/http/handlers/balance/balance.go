package balance

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/controller/http/utils"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type balanceService interface {
	Get(ctx context.Context, userID int64) (entity.Balance, error)
}

type withdrawalService interface {
	Withdraw(ctx context.Context, userID int64, withdrawal entity.Withdrawal) error
}

type authMiddleware interface {
	Validate(next http.Handler) http.Handler
}

type balanceHandler struct {
	balanceSvc    balanceService
	withdrawalSvc withdrawalService
}

func NewBalanceHandler(
	middleware authMiddleware,
	balanceSvc balanceService,
	withdrawalSvc withdrawalService,
) func(chi.Router) {
	h := &balanceHandler{
		balanceSvc:    balanceSvc,
		withdrawalSvc: withdrawalSvc,
	}
	return func(r chi.Router) {
		r.Route("/balance", func(r chi.Router) {
			r.Use(middleware.Validate)
			r.Post("/withdraw", h.Withdraw)
			r.Get("/", h.GetBalance)
		})

	}
}

func (h *balanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserID(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	balance, err := h.balanceSvc.Get(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	balanceDTO := struct {
		Current   float32 `json:"Balance"`
		Withdrawn float32 `json:"withdrawn"`
	}{Current: balance.Balance, Withdrawn: balance.Withdrawn}

	render.JSON(w, r, balanceDTO)
}

func (h *balanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserID(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var withdrawal entity.Withdrawal

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(body, &withdrawal); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.withdrawalSvc.Withdraw(r.Context(), userID, withdrawal)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	if errors.Is(err, common.ErrInsufficientBalance) {
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}

	if errors.Is(err, common.ErrInvalidOrderNumber) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
}
