package httpx

import (
	"context"
	"net"
	"net/http"
	"time"

	"tech-ip-sem2/shared/middleware"
)

// NewClient creates an http.Client with sensible defaults and timeout.
func NewClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// WithRequestID attaches request-id header from context if present.
func WithRequestID(ctx context.Context, req *http.Request) {
	if req == nil {
		return
	}
	if id := middleware.FromContext(ctx); id != "" {
		req.Header.Set(middleware.RequestIDHeader, id)
	}
}

