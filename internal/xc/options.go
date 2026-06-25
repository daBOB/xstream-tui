// Package xc — functional options for configuring a Client.
package xc

import (
	"io"
	"net/http"
	"time"
)

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithDebug enables raw response logging.
func WithDebug(enabled bool) ClientOption {
	return func(c *Client) {
		c.debug = enabled
	}
}

// WithDebugWriter sets a custom writer for debug output.
func WithDebugWriter(w io.Writer) ClientOption {
	return func(c *Client) {
		c.debugWriter = w
	}
}

// WithRetry configures retry behavior for transient failures.
func WithRetry(config RetryConfig) ClientOption {
	return func(c *Client) {
		c.retryConfig = config
	}
}

// WithNoRetry disables retry logic.
func WithNoRetry() ClientOption {
	return func(c *Client) {
		c.retryConfig.MaxRetries = 0
	}
}
