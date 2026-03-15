package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logging writes basic access logs with method, path, status and latency.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)
		log.Printf("method=%s path=%s status=%d duration=%s request_id=%s",
			r.Method, r.URL.Path, ww.status, time.Since(start), FromContext(r.Context()))
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

