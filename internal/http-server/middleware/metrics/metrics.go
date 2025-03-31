package middleware_metrics

import (
	"net/http"
	"time"

	metric "github.com/RozmiDan/url_shortener/internal/metrics"
	"github.com/go-chi/chi"
)

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(ww, r)

		duration := time.Since(start).Seconds()

		metric.HTTPRequestsTotal.WithLabelValues(
			chi.RouteContext(r.Context()).RoutePattern(),
			r.Method,
			http.StatusText(ww.status),
		).Inc()

		metric.HTTPRequestsDuration.WithLabelValues(
			chi.RouteContext(r.Context()).RoutePattern(),
			r.Method,
		).Observe(duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
