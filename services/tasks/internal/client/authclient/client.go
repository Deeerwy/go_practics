package authclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"tech-ip-sem2/shared/httpx"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

type Client struct {
	baseURL string
	http    *http.Client
}

type verifyResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error"`
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		http:    httpx.NewClient(timeout),
	}
}

// Verify calls Auth /v1/auth/verify with the provided Authorization header.
func (c *Client) Verify(ctx context.Context, authorization string) error {
	if authorization == "" {
		return ErrUnauthorized
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/auth/verify", nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", authorization)
	httpx.WithRequestID(ctx, req)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("auth request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var vr verifyResponse
		_ = json.NewDecoder(resp.Body).Decode(&vr)
		if !vr.Valid {
			return ErrUnauthorized
		}
		return nil
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	default:
		if resp.StatusCode >= 500 {
			return fmt.Errorf("auth server error: %d", resp.StatusCode)
		}
		return fmt.Errorf("unexpected status from auth: %d", resp.StatusCode)
	}
}

