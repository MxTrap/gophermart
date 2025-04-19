package middlewares

import (
	"errors"
	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthorizationMiddleware_Validate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockValidator := NewMocktokenValidator(ctrl)
	middleware := NewAuhtorizationMiddleware(mockValidator)

	// Создаем тестовый обработчик для проверки
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDKey("UserID")).(int64)
		if !ok {
			http.Error(w, "user ID not found in context", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("userID: " + string(rune(userID))))
	})

	t.Run("valid token", func(t *testing.T) {
		userID := int64(123)
		token := "valid-token"

		mockValidator.EXPECT().
			Parse(entity.Token(token)).
			Return(userID, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", token)
		rr := httptest.NewRecorder()

		handler := middleware.Validate(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

	})

	t.Run("missing authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handler := middleware.Validate(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), common.ErrInvalidCredentials.Error())
	})

	t.Run("invalid token", func(t *testing.T) {
		token := "invalid-token"
		parseErr := common.ErrInvalidToken

		mockValidator.EXPECT().
			Parse(entity.Token(token)).
			Return(int64(0), parseErr)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", token)
		rr := httptest.NewRecorder()

		handler := middleware.Validate(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), parseErr.Error())
	})

	t.Run("token validator error", func(t *testing.T) {
		token := "error-token"
		genericErr := errors.New("token parsing error")

		mockValidator.EXPECT().
			Parse(entity.Token(token)).
			Return(int64(0), genericErr)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", token)
		rr := httptest.NewRecorder()

		handler := middleware.Validate(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), genericErr.Error())
	})
}
