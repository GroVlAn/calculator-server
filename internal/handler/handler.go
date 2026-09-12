package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

const (
	metricEndpoint = "/metrics"
)

type service interface {
	Calculate(value int64)
}

type prometheus interface {
	Handler() http.Handler
	WriteTotal(
		method string,
		path string,
		status int,
	)
}

type Deps struct {
	BasePath       string
	DefaultTimeout time.Duration
}

type Handler struct {
	l  zerolog.Logger
	s  service
	pr prometheus
	Deps
}

func New(
	l zerolog.Logger,
	s service,
	pr prometheus,
	deps Deps,
) *Handler {
	return &Handler{
		l:    l,
		s:    s,
		pr:   pr,
		Deps: deps,
	}
}

func (h *Handler) Handler() *chi.Mux {
	r := chi.NewRouter()

	r.Route(h.BasePath, func(r chi.Router) {
		r.Handle(metricEndpoint, h.pr.Handler())
		h.calculatorRoute(r)
	})

	return r
}
