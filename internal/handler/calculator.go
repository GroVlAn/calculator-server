package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

const (
	calcEndpoint = "/calc"

	numKey = "num"
)

func (h *Handler) calculatorRoute(r chi.Router) {
	r.With(h.metrics).Post(calcEndpoint, h.calc)
}

func (h *Handler) calc(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	numQuery := query.Get(numKey)
	if len(numQuery) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("missing '%s' query parameter", numKey)))
		return
	}

	num, err := strconv.ParseInt(numQuery, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("'%s' must be an integer", numKey)))
		return
	}

	h.s.Calculate(num)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
