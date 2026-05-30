// Package api is a typed HTTP client for the lumo-drive server. Each exported
// method maps to exactly one server endpoint.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client talks to a lumo-drive server.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// New returns a Client for the given base URL (e.g. http://localhost:6000) and
// optional bearer token.
func New(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: 60 * time.Second},
	}
}

// SetToken updates the bearer token used for authenticated requests.
func (c *Client) SetToken(token string) { c.token = token }

// APIError is a non-2xx response decoded from the server's {"error": ...} envelope.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("server returned %d", e.StatusCode)
	}
	return fmt.Sprintf("%s (status %d)", e.Message, e.StatusCode)
}

// doJSON performs an HTTP request with an optional JSON body and decodes a JSON
// response into out (which may be nil). It attaches the auth token when set.
func (c *Client) doJSON(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Message: decodeErrorMessage(data)}
	}

	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// decodeErrorMessage pulls the "error" field out of an error response body,
// falling back to the raw body.
func decodeErrorMessage(data []byte) string {
	var env struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(data, &env); err == nil && env.Error != "" {
		return env.Error
	}
	return strings.TrimSpace(string(data))
}

// Health probes the server's /health endpoint. Used to detect connectivity.
func (c *Client) Health(ctx context.Context) error {
	return c.doJSON(ctx, http.MethodGet, "/health", nil, nil)
}
