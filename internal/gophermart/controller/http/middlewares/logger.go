package middlewares

import (
	"bytes"
	"net/http"

	"go.uber.org/zap"
)

type logger interface {
	With(args ...any) *zap.SugaredLogger
	Info(data ...any)
	Error(args ...any)
}

type responseWriter struct {
	http.ResponseWriter
	Code int
	Body *bytes.Buffer
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.Code = statusCode
}

func (w *responseWriter) Write(data []byte) (int, error) {
	return w.Body.Write(data)
}

func (w *responseWriter) Flush() {
	if w.Code != 0 {
		w.ResponseWriter.WriteHeader(w.Code)
	}
	if w.Body.Len() > 0 {
		w.ResponseWriter.Write(w.Body.Bytes())
	}
}

func LoggerMiddleware(log logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := log.With(r.RequestURI, r.Method)
			rw := &responseWriter{ResponseWriter: w, Body: new(bytes.Buffer)}
			h.ServeHTTP(rw, r)
			if rw.Code >= 400 && rw.Body.Len() > 0 {
				log.Error(rw.Body.String())
				rw.Body.Reset()
			}
			rw.Flush()
		})
	}
}
