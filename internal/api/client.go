package api

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aaronfaby/kalshi-perp-cli/internal/auth"
)

// Client is an authenticated Kalshi margin REST client.
type Client struct {
	BaseURL    string // includes /trade-api/v2
	HTTPClient *http.Client
	KeyID      string
	PrivateKey *rsa.PrivateKey
	Verbose    bool
	Logf       func(format string, args ...any)
}

// New creates a Client. baseURL should be like https://host/trade-api/v2 (no trailing slash).
func New(baseURL, keyID string, privateKey *rsa.PrivateKey, timeout time.Duration) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		KeyID:      keyID,
		PrivateKey: privateKey,
	}
}

// APIError wraps HTTP status and optional ErrorResponse body.
type APIError struct {
	StatusCode int
	Body       ErrorResponse
	RawBody    string
}

func (e *APIError) Error() string {
	if e.Body.Message != "" {
		return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body.Error())
	}
	if e.RawBody != "" {
		return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.RawBody)
	}
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}

// Do performs a signed request. path is relative to BaseURL (e.g. /margin/orders).
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return err
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	// Sign full URL path without query.
	signPath := u.Path
	headers, err := auth.Headers(c.KeyID, c.PrivateKey, method, signPath)
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if c.Verbose && c.Logf != nil {
		c.Logf("%s %s", method, u.Redacted())
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if c.Verbose && c.Logf != nil {
		c.Logf("-> %d (%d bytes)", resp.StatusCode, len(raw))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Body: parseErrorBody(raw), RawBody: string(raw)}
	}

	if out == nil || len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode response: %w (body: %s)", err, truncate(string(raw), 200))
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// parseErrorBody accepts either a flat ErrorResponse or {"error": ErrorResponse}.
func parseErrorBody(raw []byte) ErrorResponse {
	var flat ErrorResponse
	if err := json.Unmarshal(raw, &flat); err == nil && (flat.Code != "" || flat.Message != "") {
		return flat
	}
	var wrapped struct {
		Error ErrorResponse `json:"error"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil {
		return wrapped.Error
	}
	return flat
}

// Get is a convenience for GET without body.
func (c *Client) Get(ctx context.Context, path string, query url.Values, out any) error {
	return c.Do(ctx, http.MethodGet, path, query, nil, out)
}

// Post JSON body.
func (c *Client) Post(ctx context.Context, path string, body any, out any) error {
	return c.Do(ctx, http.MethodPost, path, nil, body, out)
}

// Put JSON body.
func (c *Client) Put(ctx context.Context, path string, body any, out any) error {
	return c.Do(ctx, http.MethodPut, path, nil, body, out)
}

// Delete request.
func (c *Client) Delete(ctx context.Context, path string, out any) error {
	return c.Do(ctx, http.MethodDelete, path, nil, nil, out)
}

// Query helpers.
func Q() url.Values { return url.Values{} }

func SetInt(q url.Values, key string, v *int) {
	if v != nil {
		q.Set(key, fmt.Sprintf("%d", *v))
	}
}

func SetInt64(q url.Values, key string, v *int64) {
	if v != nil {
		q.Set(key, fmt.Sprintf("%d", *v))
	}
}

func SetStr(q url.Values, key, v string) {
	if v != "" {
		q.Set(key, v)
	}
}
