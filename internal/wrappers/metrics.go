package wrappers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	labelStatusCode = "code"
	labelMethod     = "method"
	labelHost       = "host"
	labelRoute      = "route"
)

type Metrics struct {
	ReqTotal      *prometheus.CounterVec
	ReqDurationMs *prometheus.HistogramVec
	RespSizeBytes *prometheus.SummaryVec
	defBuckets    []float64
}

func NewMetrics(namespace, subsystem string) *Metrics {
	m := Metrics{
		defBuckets: prometheus.DefBuckets,
	}

	m.ReqTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "requests_total",
		Subsystem: subsystem,
		Namespace: namespace,
		Help:      "The total number of requests received",
	}, []string{labelStatusCode, labelMethod, labelHost, labelRoute})

	m.ReqDurationMs = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "request_duration_ms",
		Subsystem: subsystem,
		Namespace: namespace,
		Help:      "Histogram of the request duration",
		Buckets:   m.defBuckets,
	}, []string{labelStatusCode, labelMethod, labelHost, labelRoute})

	m.RespSizeBytes = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:      "response_size_bytes",
		Subsystem: subsystem,
		Namespace: namespace,
		Help:      "Summary of response bytes sent",
	}, []string{labelStatusCode, labelMethod, labelHost, labelRoute})

	return &m
}

// WithMetrics instruments the HTTP handler with metrics about incoming requests and responses
func (m *Metrics) WithMetrics(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		crw := customResponseWriter{ResponseWriter: w}

		next.ServeHTTP(&crw, r)

		labelValues := []string{strconv.Itoa(crw.status), r.Method, r.Host, r.RequestURI}

		m.ReqTotal.WithLabelValues(labelValues...).Inc()
		m.ReqDurationMs.WithLabelValues(labelValues...).Observe(float64(time.Since(start).Milliseconds()))
		m.RespSizeBytes.WithLabelValues(labelValues...).Observe(float64(crw.size))
	})
}
