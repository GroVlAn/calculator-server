package prometheus_metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusMetrics struct {
	registry          *prometheus.Registry
	httpRequestsTotal *prometheus.CounterVec
	cCallDuration     prometheus.Histogram
	rustCallDuration  prometheus.Histogram
}

func New() *PrometheusMetrics {
	registry := prometheus.NewRegistry()

	httpRequestTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	cCallDuration := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "c_call_duration_seconds",
			Help:    "Duration of C functions calls.",
			Buckets: prometheus.DefBuckets,
		},
	)

	rustCallDuration := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "rust_call_duration_seconds",
			Help:    "Duration of Rust functions calls.",
			Buckets: prometheus.DefBuckets,
		},
	)

	registry.MustRegister(
		httpRequestTotal,
		cCallDuration,
		rustCallDuration,
	)

	return &PrometheusMetrics{
		registry:          registry,
		httpRequestsTotal: httpRequestTotal,
		cCallDuration:     cCallDuration,
		rustCallDuration:  rustCallDuration,
	}
}

func (pm *PrometheusMetrics) CMetric(fn func(a, b int64) int64, a, b int64) int64 {
	start := time.Now()

	res := fn(a, b)

	pm.cCallDuration.Observe(time.Since(start).Seconds())

	return res
}

func (pm *PrometheusMetrics) RustMetric(fn func(a, b int64) int64, a, b int64) int64 {
	start := time.Now()

	res := fn(a, b)

	pm.rustCallDuration.Observe(time.Since(start).Seconds())

	return res
}

func (pm *PrometheusMetrics) WriteTotal(
	method string,
	path string,
	status int,
) {
	pm.httpRequestsTotal.WithLabelValues(
		method,
		path,
		strconv.Itoa(status),
	).Inc()
}

func (pm *PrometheusMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(
		pm.registry,
		promhttp.HandlerOpts{},
	)
}
