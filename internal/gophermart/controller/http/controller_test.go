package http

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewController(t *testing.T) {
	host := ":8080"
	ctrl := NewController(host)

	t.Run("correct initialization", func(t *testing.T) {
		assert.NotNil(t, ctrl)
		assert.Equal(t, host, ctrl.host)
		assert.NotNil(t, ctrl.router)
		assert.NotNil(t, ctrl.handlers)
		assert.Empty(t, ctrl.handlers)
		assert.Nil(t, ctrl.server)
	})
}

func TestController_RegisterMiddlewares(t *testing.T) {
	host := "localhost:8080"
	ctrl := NewController(host)

	called := 0
	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called++
			next.ServeHTTP(w, r)
		})
	}
	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called++
			next.ServeHTTP(w, r)
		})
	}

	ctrl.RegisterMiddlewares(middleware1, middleware2)
	ctrl.AddHandler("/", func(router chi.Router) {
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		})
	})

	go ctrl.Start()

	resp, err := http.Get("http://" + host)
	ctrl.Stop(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, 2, called, "both middlewares should be called")
}

func TestController_AddHandler(t *testing.T) {
	ctrl := NewController(":8080")

	path1 := "/test1"
	path2 := "/test2"
	handler1 := func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {})
	}
	handler2 := func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {})
	}

	t.Run("add new handler", func(t *testing.T) {
		ctrl.AddHandler(path1, handler1)
		assert.Len(t, ctrl.handlers[path1], 1)
		assert.Len(t, ctrl.handlers, 1)
	})

	t.Run("append handler to existing path", func(t *testing.T) {
		ctrl.AddHandler(path1, handler2)
		assert.Len(t, ctrl.handlers[path1], 2)
		assert.Len(t, ctrl.handlers, 1)
	})

	t.Run("add handler to new path", func(t *testing.T) {
		ctrl.AddHandler(path2, handler1)
		assert.Len(t, ctrl.handlers[path2], 1)
		assert.Len(t, ctrl.handlers, 2)
	})
}

func TestController_registerHandlers(t *testing.T) {
	ctrl := NewController(":8080")

	path := "/user"
	handler := func(r chi.Router) {
		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("user handler"))
		})
	}
	ctrl.AddHandler(path, handler)
	ctrl.registerHandlers()

	server := httptest.NewServer(ctrl.router)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/user/123")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := readResponseBody(resp)
	assert.Equal(t, "user handler", string(body))
}

func readResponseBody(resp *http.Response) []byte {
	var body []byte
	if resp.Body != nil {
		body, _ = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	}
	return body
}
