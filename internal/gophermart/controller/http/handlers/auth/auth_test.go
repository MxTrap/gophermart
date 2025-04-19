package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewAuthHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockauthService(ctrl)
	handlerFunc := NewAuthHandler(mockService)

	t.Run("handler registration", func(t *testing.T) {
		router := chi.NewRouter()
		handlerFunc(router)

		// Проверяем, что маршруты зарегистрированы
		routes := router.Routes()
		assert.NotEmpty(t, routes)
		var loginFound, registerFound bool
		for _, route := range routes {
			if route.Pattern == "/login" {
				loginFound = true
			}
			if route.Pattern == "/register" {
				registerFound = true
			}
		}
		assert.True(t, loginFound, "POST /login route should be registered")
		assert.True(t, registerFound, "POST /register route should be registered")
	})
}

func TestHandler_readUser(t *testing.T) {
	h := &handler{}

	t.Run("valid user JSON", func(t *testing.T) {
		user := entity.User{Login: "testuser", Password: "password123"}
		body, _ := json.Marshal(user)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

		result, err := h.readUser(req)
		assert.NoError(t, err)
		assert.Equal(t, user, result)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("invalid json"))
		result, err := h.readUser(req)
		assert.Error(t, err)
		assert.Equal(t, entity.User{}, result)
	})

	t.Run("empty body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		result, err := h.readUser(req)
		assert.Error(t, err)
		assert.Equal(t, entity.User{}, result)
	})
}

func TestHandler_sendTokens(t *testing.T) {
	h := &handler{}
	token := entity.Token("jwt-token")

	rr := httptest.NewRecorder()
	h.sendTokens(rr, token)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, string(token), rr.Header().Get("Authorization"))
	assert.Empty(t, rr.Body.String())
}

func TestHandler_LoginHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockauthService(ctrl)
	h := &handler{service: mockService}

	user := entity.User{Login: "testuser", Password: "password123"}
	token := entity.Token("jwt-token")
	body, _ := json.Marshal(user)

	t.Run("successful login", func(t *testing.T) {
		mockService.EXPECT().
			Login(gomock.Any(), user).
			Return(token, nil)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		h.LoginHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, string(token), rr.Header().Get("Authorization"))
	})

	t.Run("invalid credentials", func(t *testing.T) {
		mockService.EXPECT().
			Login(gomock.Any(), user).
			Return(entity.Token(""), common.ErrInvalidCredentials)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		h.LoginHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockService.EXPECT().
			Login(gomock.Any(), user).
			Return(entity.Token(""), errors.New("server error"))

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		h.LoginHandler(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("invalid json"))
		rr := httptest.NewRecorder()

		h.LoginHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestHandler_RegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockauthService(ctrl)
	h := &handler{service: mockService}

	user := entity.User{Login: "testuser", Password: "password123"}
	token := entity.Token("jwt-token")
	body, _ := json.Marshal(user)

	t.Run("successful registration", func(t *testing.T) {
		mockService.EXPECT().
			RegisterNewUser(gomock.Any(), user).
			Return(token, nil)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		h.RegisterHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, string(token), rr.Header().Get("Authorization"))
	})

	t.Run("user already exists", func(t *testing.T) {
		mockService.EXPECT().
			RegisterNewUser(gomock.Any(), user).
			Return(entity.Token(""), common.ErrUserAlreadyExist)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		h.RegisterHandler(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockService.EXPECT().
			RegisterNewUser(gomock.Any(), user).
			Return(entity.Token(""), errors.New("server error"))

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		h.RegisterHandler(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader("invalid json"))
		rr := httptest.NewRecorder()

		h.RegisterHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
