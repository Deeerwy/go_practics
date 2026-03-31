package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"example.com/pz4-monitoring/internal/metrics"
)

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lrw, r)

		duration := time.Since(start).Seconds()
		path := r.URL.Path

		metrics.HttpRequestsTotal.WithLabelValues(r.Method, path).Inc()
		metrics.HttpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)

		if lrw.StatusCode() >= 400 {
			metrics.HttpErrorsTotal.WithLabelValues(
				r.Method,
				path,
				strconv.Itoa(lrw.StatusCode()),
			).Inc()
		}
	})
}
