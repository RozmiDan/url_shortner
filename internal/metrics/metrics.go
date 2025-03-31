package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Количество HTTP-запросов.",
		},
		[]string{"path", "method", "status"},
	)

	HTTPRequestsDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Длительность HTTP-запросов в секундах.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)
)

func RegisterMetrics() {
	prometheus.MustRegister(HTTPRequestsTotal, HTTPRequestsDuration)
}
