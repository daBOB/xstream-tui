// Package xc provides an HTTP client for the Xtream Codes API.
//
// The Xtream Codes API is used by IPTV providers to serve live TV, VOD,
// and series content. This package handles the JSON type inconsistencies
// common across different providers (IDs as int or string).
//
// # Security Note
//
// The XC API specification requires credentials in URL paths and query
// parameters. This is an API limitation, not a design choice. Ensure:
//   - Never log full URLs containing credentials
//   - Use HTTPS when the server supports it
//   - Store credentials securely (env vars, not in code)
package xc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// maxErrorBodySize limits error response body reads to prevent memory exhaustion.
const maxErrorBodySize = 4096

// ErrInvalidConfig is returned when client configuration is invalid.
var ErrInvalidConfig = errors.New("invalid client configuration")

// Client is an HTTP client for the Xtream Codes API.
type Client struct {
	baseURL     string
	username    string
	password    string
	httpClient  *http.Client
	retryConfig RetryConfig
	debug       bool      // Enable raw response logging
	debugWriter io.Writer // Writer for debug output (defaults to os.Stderr)
}

// NewClient creates a new XC API client.
// Host should include protocol (http:// or https://).
// Returns error if host, username, or password are empty.
func NewClient(host string, username, password string, opts ...ClientOption) (*Client, error) {
	// Validate required parameters
	host = strings.TrimSpace(host)
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	if host == "" {
		return nil, fmt.Errorf("%w: host cannot be empty", ErrInvalidConfig)
	}
	if username == "" {
		return nil, fmt.Errorf("%w: username cannot be empty", ErrInvalidConfig)
	}
	if password == "" {
		return nil, fmt.Errorf("%w: password cannot be empty", ErrInvalidConfig)
	}

	// Validate host is a valid URL
	if _, err := url.Parse(host); err != nil {
		return nil, fmt.Errorf("%w: invalid host URL: %v", ErrInvalidConfig, err)
	}

	c := &Client{
		baseURL:  strings.TrimSuffix(host, "/"),
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		retryConfig: DefaultRetryConfig(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// buildURL constructs an API URL with authentication and action parameters.
func (c *Client) buildURL(action string, params map[string]string) (string, error) {
	u, err := url.Parse(c.baseURL + "/player_api.php")
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	q := u.Query()
	q.Set("username", c.username)
	q.Set("password", c.password)
	if action != "" {
		q.Set("action", action)
	}
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// get performs a GET request with retry logic and decodes JSON response.
func (c *Client) get(ctx context.Context, action string, params map[string]string, result any) error {
	reqURL, err := c.buildURL(action, params)
	if err != nil {
		return fmt.Errorf("xc %s: %w", actionName(action), err)
	}

	var lastErr error
	maxAttempts := c.retryConfig.MaxRetries + 1

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Check context before each attempt
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("xc %s: %w", actionName(action), err)
		}

		// Wait before retry (skip on first attempt)
		if attempt > 0 {
			delay := calculateBackoff(attempt-1, c.retryConfig)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return fmt.Errorf("xc %s: %w", actionName(action), ctx.Err())
			}
		}

		lastErr = c.doRequest(ctx, reqURL, action, result)
		if lastErr == nil {
			return nil
		}

		// Don't retry if error is not retryable
		if !isRetryable(lastErr) {
			return lastErr
		}
	}

	return fmt.Errorf("xc %s: max retries exceeded: %w", actionName(action), lastErr)
}

// doRequest performs a single HTTP request attempt.
func (c *Client) doRequest(ctx context.Context, reqURL, action string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("xc %s: create request: %w", actionName(action), err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Wrap network errors to indicate retryability
		if isRetryable(err) {
			return &retryableError{err: err, retryable: true}
		}
		return fmt.Errorf("xc %s: http request: %w", actionName(action), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
		err := fmt.Errorf("xc %s: status %d: %s", actionName(action), resp.StatusCode, string(body))

		// Check if this status code is retryable
		if isRetryableStatus(resp.StatusCode) {
			return &retryableError{err: err, statusCode: resp.StatusCode, retryable: true}
		}

		// Auth errors should never be retried
		if isAuthError(resp.StatusCode) {
			return &retryableError{err: err, statusCode: resp.StatusCode, retryable: false}
		}

		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("xc %s: read response: %w", actionName(action), err)
	}

	if c.debug {
		c.logDebug(action, body)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("xc %s: decode response: %w", actionName(action), err)
	}

	return nil
}

// actionName returns a display name for the action (or "auth" if empty).
func actionName(action string) string {
	if action == "" {
		return "auth"
	}
	return action
}

// logDebug writes raw response to the debug writer with credentials redacted.
func (c *Client) logDebug(action string, body []byte) {
	w := c.debugWriter
	if w == nil {
		w = os.Stderr
	}

	// Format action name for display
	if action == "" {
		action = "auth"
	}

	// Redact password from debug output to prevent credential leakage
	output := string(body)
	if c.password != "" {
		output = strings.ReplaceAll(output, c.password, "***REDACTED***")
	}

	fmt.Fprintf(w, "\n=== DEBUG: XC API Response [%s] ===\n", action)
	fmt.Fprintf(w, "%s\n", output)
	fmt.Fprintf(w, "=== END DEBUG ===\n\n")
}
