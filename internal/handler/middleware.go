package handler

import (
	"net/http"
)

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (srw *statusResponseWriter) WriteHeader(statusCode int) {
	srw.statusCode = statusCode
	srw.ResponseWriter.WriteHeader(statusCode)
}

func (h *Handler) metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &statusResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		h.pr.WriteTotal(
			r.Method,
			r.URL.Path,
			rw.statusCode,
		)
	})
}
