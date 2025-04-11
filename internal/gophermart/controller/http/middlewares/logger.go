package middlewares

import "net/http"

type logger interface {
	Info(data ...any)
}

func LoggerMiddleware(log logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Info(r.RequestURI)
			h.ServeHTTP(w, r)
		})
	}
}
