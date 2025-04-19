package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"

	"github.com/go-chi/chi/v5"
)

type authService interface {
	Login(ctx context.Context, user entity.User) (entity.Token, error)
	RegisterNewUser(ctx context.Context, user entity.User) (entity.Token, error)
}

type handler struct {
	service authService
}

type TokenDto struct {
	Token string `json:"access_token"`
}

func NewAuthHandler(service authService) func(chi.Router) {
	h := &handler{service: service}
	return func(r chi.Router) {
		r.Post("/login", h.LoginHandler)
		r.Post("/register", h.RegisterHandler)
	}
}

func (h *handler) readUser(r *http.Request) (entity.User, error) {
	var user entity.User

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return user, err
	}

	if err := json.Unmarshal(body, &user); err != nil {
		return user, err
	}
	return user, nil
}

func (h *handler) sendTokens(w http.ResponseWriter, token entity.Token) {
	w.Header().Set("Authorization", string(token))
	w.WriteHeader(http.StatusOK)
}

func (h *handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.readUser(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tokens, err := h.service.Login(r.Context(), user)
	if err == nil {
		h.sendTokens(w, tokens)
		return
	}

	if errors.Is(err, common.ErrInvalidCredentials) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
}

func (h *handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.readUser(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tokens, err := h.service.RegisterNewUser(r.Context(), user)
	if err == nil {
		h.sendTokens(w, tokens)
		return
	}

	if errors.Is(err, common.ErrUserAlreadyExist) {
		w.WriteHeader(http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
}
