package http

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Controller struct {
	router   chi.Router
	server   *http.Server
	host     string
	handlers map[string][]func(chi.Router)
}

func NewController(host string) *Controller {
	r := chi.NewRouter()

	return &Controller{
		router:   r,
		host:     host,
		handlers: make(map[string][]func(chi.Router)),
	}
}

func (c *Controller) RegisterMiddlewares(middlewares ...func(http.Handler) http.Handler) {
	for _, middleware := range middlewares {
		c.router.Use(middleware)
	}
}

func (c *Controller) AddHandler(path string, group ...func(chi.Router)) {
	if val, ok := c.handlers[path]; ok {
		val = append(val, group...)
		c.handlers[path] = val
		return
	}
	c.handlers[path] = group
}

func (c *Controller) registerHandlers() {
	c.router.Route("/api", func(r chi.Router) {
		for path, group := range c.handlers {
			r.Route(path, func(r chi.Router) {
				for _, handler := range group {
					handler(r)
				}
			})
		}
	})
}

func (c *Controller) Start() error {
	c.server = &http.Server{
		Addr:    c.host,
		Handler: c.router,
	}
	c.registerHandlers()
	return c.server.ListenAndServe()
}

func (c *Controller) Stop(ctx context.Context) error {
	return c.server.Shutdown(ctx)
}
